package alicloud

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkCapacityCoordinator() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAliCloudFlinkCapacityCoordinatorCreate,
		Read:          resourceAliCloudFlinkCapacityCoordinatorRead,
		Update:        resourceAliCloudFlinkCapacityCoordinatorUpdate,
		Delete:        resourceAliCloudFlinkCapacityCoordinatorDelete,
		CustomizeDiff: flinkCapacityCoordinatorCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "Identifies the workspace exclusively owned by this coordinator. Exactly one coordinator across Terraform states and provider processes may manage a workspace; provider locking serializes only within one process.",
			},
			"workspace_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace": flinkCoordinatorWorkspaceSchema(),
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},
	}
}

func flinkCoordinatorWorkspaceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"capacity":          flinkCoordinatorCapacitySchema(true),
			"observed_capacity": flinkObservedCapacitySchema(true),
			"namespace": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"capacity":          flinkCoordinatorCapacitySchema(false),
					"observed_capacity": flinkObservedCapacitySchema(false),
					"queue": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"name": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"capacity":          flinkCoordinatorCapacitySchema(false),
							"observed_capacity": flinkObservedCapacitySchema(false),
						}},
					},
				}},
			},
		}},
	}
}

func flinkCoordinatorCapacitySchema(includeHA bool) *schema.Schema {
	fields := map[string]*schema.Schema{
		"fixed_cu": {
			Type:         schema.TypeFloat,
			Optional:     true,
			Default:      0.0,
			ValidateFunc: validation.FloatAtLeast(0),
		},
		"elastic_cu_limit": {
			Type:         schema.TypeFloat,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.FloatAtLeast(0),
		},
		"max_cu_limit": {
			Type:         schema.TypeFloat,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.FloatAtLeast(0),
		},
	}
	if includeHA {
		fields["ha"] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"cross_zone_fixed_cu": {
					Type:         schema.TypeFloat,
					Optional:     true,
					Default:      0.0,
					ValidateFunc: validation.FloatAtLeast(0),
				},
			}},
		}
	}
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: !includeHA,
		Required: includeHA,
		MaxItems: 1,
		Elem:     &schema.Resource{Schema: fields},
	}
}

func flinkCapacityCoordinatorCustomizeDiff(d *schema.ResourceDiff, _ interface{}) error {
	if err := validateFlinkCoordinatorAliasPresence(d); err != nil {
		return err
	}
	if _, err := expandFlinkCoordinatorDesired(d.Get("workspace")); err != nil {
		return err
	}
	if d.Id() == "" {
		return nil
	}
	workspace, ok := flinkFirstBlock(d.Get("workspace"))
	if !ok {
		return nil
	}
	observed, observedOK := flinkFirstBlock(workspace["observed_capacity"])
	if !observedOK {
		return nil
	}
	observedElastic, _ := numberAsFloat(observed["elastic_cu_limit"])
	if observedElastic <= 0 {
		return nil
	}
	desired, err := expandFlinkCoordinatorDesired(d.Get("workspace"))
	if err != nil {
		return err
	}
	if desired.Workspace.AsCapacity().Elastic() == 0 {
		return fmt.Errorf("cannot reduce workspace elastic CU to zero through a supported public API; first reduce namespace and queue elastic quotas to zero, disable workspace elastic billing in the console, refresh state, and plan again")
	}
	return nil
}

func validateFlinkCoordinatorAliasPresence(d *schema.ResourceDiff) error {
	check := func(path string, includeHA bool) error {
		_, hasElastic := d.GetOkExists(path + ".elastic_cu_limit")
		maxValue, hasMax := d.GetOkExists(path + ".max_cu_limit")
		if hasElastic && hasMax {
			return fmt.Errorf("%s.elastic_cu_limit and %s.max_cu_limit are mutually exclusive", path, path)
		}
		if !hasMax {
			return nil
		}
		max, ok := numberAsFloat(maxValue)
		if !ok {
			return nil
		}
		fixed, _ := numberAsFloat(d.Get(path + ".fixed_cu"))
		if includeHA {
			crossZone, _ := numberAsFloat(d.Get(path + ".ha.0.cross_zone_fixed_cu"))
			fixed += crossZone
		}
		if max < fixed {
			return fmt.Errorf("%s.max_cu_limit %v must be greater than or equal to fixed CU %v", path, max, fixed)
		}
		return nil
	}

	workspace, ok := flinkFirstBlock(d.Get("workspace"))
	if !ok {
		return nil
	}
	if flinkListBlockConfigured(workspace["capacity"]) {
		if err := check("workspace.0.capacity.0", true); err != nil {
			return err
		}
	}
	namespaces, _ := workspace["namespace"].([]interface{})
	for namespaceIndex, rawNamespace := range namespaces {
		namespace, ok := rawNamespace.(map[string]interface{})
		if !ok {
			continue
		}
		namespacePath := fmt.Sprintf("workspace.0.namespace.%d", namespaceIndex)
		if flinkListBlockConfigured(namespace["capacity"]) {
			if err := check(namespacePath+".capacity.0", false); err != nil {
				return err
			}
		}
		queues, _ := namespace["queue"].([]interface{})
		for queueIndex, rawQueue := range queues {
			queue, ok := rawQueue.(map[string]interface{})
			if !ok || !flinkListBlockConfigured(queue["capacity"]) {
				continue
			}
			if err := check(fmt.Sprintf("%s.queue.%d.capacity.0", namespacePath, queueIndex), false); err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceAliCloudFlinkCapacityCoordinatorCreate(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("workspace_instance_id").(string))
	return resourceAliCloudFlinkCapacityCoordinatorReconcile(d, meta, schema.TimeoutCreate)
}

func resourceAliCloudFlinkCapacityCoordinatorUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAliCloudFlinkCapacityCoordinatorReconcile(d, meta, schema.TimeoutUpdate)
}

func resourceAliCloudFlinkCapacityCoordinatorReconcile(d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	instanceID := d.Get("workspace_instance_id").(string)
	return withFlinkCapacityCoordinatorLock(instanceID, func() error {
		return resourceAliCloudFlinkCapacityCoordinatorReconcileUnlocked(d, meta, timeoutKey)
	})
}

func withFlinkCapacityCoordinatorLock(instanceID string, reconcile func() error) error {
	lockKey := fmt.Sprintf("flink-capacity-%s", instanceID)
	alicloudMutexKV.Lock(lockKey)
	defer alicloudMutexKV.Unlock(lockKey)
	return reconcile()
}

func resourceAliCloudFlinkCapacityCoordinatorReconcileUnlocked(d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFlinkCapacityService(client)
	if err != nil {
		return WrapError(err)
	}
	desired, err := expandFlinkCoordinatorDesired(d.Get("workspace"))
	if err != nil {
		return WrapError(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(timeoutKey))
	defer cancel()
	actual, reconcileErr := (flinkcapacity.Reconciler{API: service}).Reconcile(ctx, d.Id(), desired)
	if reconcileErr != nil {
		_ = setFlinkCoordinatorReconcileState(d, actual)
		return WrapError(reconcileErr)
	}
	return WrapError(setFlinkCoordinatorReconcileState(d, actual))
}

func setFlinkCoordinatorReconcileState(d *schema.ResourceData, actual flinkcapacity.Tree) error {
	if actual.WorkspaceResourceID == "" {
		return fmt.Errorf("workspace %q does not expose a ResourceId", d.Id())
	}
	merged, err := mergeFlinkCoordinatorObserved(d.Get("workspace"), actual)
	if err != nil {
		return err
	}
	if err := d.Set("workspace", merged); err != nil {
		return err
	}
	return d.Set("workspace_resource_id", actual.WorkspaceResourceID)
}

func resourceAliCloudFlinkCapacityCoordinatorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFlinkCapacityService(client)
	if err != nil {
		return WrapError(err)
	}
	ctx := context.Background()
	actual, err := service.ReadTree(ctx, d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	_ = d.Set("workspace_instance_id", d.Id())
	if err := d.Set("workspace_resource_id", actual.WorkspaceResourceID); err != nil {
		return WrapError(err)
	}

	if !flinkListBlockConfigured(d.Get("workspace")) {
		if err := d.Set("workspace", flattenFlinkCoordinatorImport(actual)); err != nil {
			return WrapError(err)
		}
		return nil
	}
	merged, err := mergeFlinkCoordinatorObserved(d.Get("workspace"), actual)
	if err != nil {
		return WrapError(err)
	}
	if err := d.Set("workspace", merged); err != nil {
		return WrapError(err)
	}
	return nil
}

func resourceAliCloudFlinkCapacityCoordinatorDelete(d *schema.ResourceData, _ interface{}) error {
	d.SetId("")
	return nil
}

func expandFlinkCoordinatorDesired(value interface{}) (flinkcapacity.Tree, error) {
	workspace, ok := flinkFirstBlock(value)
	if !ok {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace block is required")
	}
	capacity, ok := flinkFirstBlock(workspace["capacity"])
	if !ok {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace.capacity block is required")
	}
	workspaceCapacity, err := expandFlinkWorkspaceCapacity(capacity)
	if err != nil {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace capacity: %w", err)
	}
	tree := flinkcapacity.Tree{Workspace: workspaceCapacity}

	namespaceItems, _ := workspace["namespace"].([]interface{})
	if len(namespaceItems) == 0 {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace must declare at least one namespace")
	}
	seenNamespaces := make(map[string]struct{}, len(namespaceItems))
	namespaceRemainders := 0
	for _, rawNamespace := range namespaceItems {
		namespaceMap, ok := rawNamespace.(map[string]interface{})
		if !ok || namespaceMap == nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace block must not be empty")
		}
		name, _ := namespaceMap["name"].(string)
		if name == "" {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace name must not be empty")
		}
		if _, exists := seenNamespaces[name]; exists {
			return flinkcapacity.Tree{}, fmt.Errorf("duplicate namespace %q", name)
		}
		seenNamespaces[name] = struct{}{}

		namespaceCapacity, err := expandFlinkCapacity(namespaceMap["capacity"], true)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q capacity: %w", name, err)
		}
		if namespaceCapacity == nil {
			namespaceRemainders++
		}
		queueItems, _ := namespaceMap["queue"].([]interface{})
		if len(queueItems) == 0 {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q must declare at least one queue", name)
		}
		domainNamespace := flinkcapacity.Namespace{Name: name, Capacity: namespaceCapacity}
		seenQueues := make(map[string]struct{}, len(queueItems))
		queueRemainders := 0
		for _, rawQueue := range queueItems {
			queueMap, ok := rawQueue.(map[string]interface{})
			if !ok || queueMap == nil {
				return flinkcapacity.Tree{}, fmt.Errorf("namespace %q queue block must not be empty", name)
			}
			queueName, _ := queueMap["name"].(string)
			if queueName == "" {
				return flinkcapacity.Tree{}, fmt.Errorf("namespace %q queue name must not be empty", name)
			}
			if _, exists := seenQueues[queueName]; exists {
				return flinkcapacity.Tree{}, fmt.Errorf("duplicate queue %q/%q", name, queueName)
			}
			seenQueues[queueName] = struct{}{}
			queueCapacity, err := expandFlinkCapacity(queueMap["capacity"], false)
			if err != nil {
				return flinkcapacity.Tree{}, fmt.Errorf("queue %q/%q capacity: %w", name, queueName, err)
			}
			if queueCapacity == nil {
				queueRemainders++
			}
			domainNamespace.Queues = append(domainNamespace.Queues, flinkcapacity.Queue{Name: queueName, Capacity: queueCapacity})
		}
		if queueRemainders > 1 {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q may have at most one queue without an explicit capacity", name)
		}
		tree.Namespaces = append(tree.Namespaces, domainNamespace)
	}
	if namespaceRemainders > 1 {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace may have at most one namespace without an explicit capacity")
	}
	return tree, nil
}

func expandFlinkWorkspaceCapacity(block map[string]interface{}) (flinkcapacity.WorkspaceCapacity, error) {
	fixedValue, _ := numberAsFloat(block["fixed_cu"])
	fixed, err := parseFlinkCoordinatorCU(fixedValue, true)
	if err != nil {
		return flinkcapacity.WorkspaceCapacity{}, fmt.Errorf("fixed_cu: %w", err)
	}
	crossZone := flinkcapacity.CU(0)
	_, hasHA := flinkFirstBlock(block["ha"])
	if ha, ok := flinkFirstBlock(block["ha"]); ok {
		value, _ := numberAsFloat(ha["cross_zone_fixed_cu"])
		crossZone, err = parseFlinkCoordinatorCU(value, true)
		if err != nil {
			return flinkcapacity.WorkspaceCapacity{}, fmt.Errorf("ha.cross_zone_fixed_cu: %w", err)
		}
	}
	totalFixed := fixed + crossZone
	capacity, err := newFlinkCoordinatorCapacity(totalFixed, block, true)
	if err != nil {
		return flinkcapacity.WorkspaceCapacity{}, err
	}
	return flinkcapacity.WorkspaceCapacity{HA: hasHA, FixedCU: fixed, CrossZoneFixedCU: crossZone, Limit: capacity.Limit}, nil
}

func expandFlinkCapacity(value interface{}, integerOnly bool) (*flinkcapacity.Capacity, error) {
	block, ok := flinkFirstBlock(value)
	if !ok {
		return nil, nil
	}
	capacity, err := expandFlinkCapacityMap(block, integerOnly)
	if err != nil {
		return nil, err
	}
	return &capacity, nil
}

func expandFlinkCapacityMap(block map[string]interface{}, integerOnly bool) (flinkcapacity.Capacity, error) {
	fixedValue, _ := numberAsFloat(block["fixed_cu"])
	fixed, err := parseFlinkCoordinatorCU(fixedValue, integerOnly)
	if err != nil {
		return flinkcapacity.Capacity{}, fmt.Errorf("fixed_cu: %w", err)
	}

	return newFlinkCoordinatorCapacity(fixed, block, integerOnly)
}

func newFlinkCoordinatorCapacity(fixed flinkcapacity.CU, block map[string]interface{}, integerOnly bool) (flinkcapacity.Capacity, error) {
	elasticValue, hasElastic := optionalNumber(block, "elastic_cu_limit")
	maxValue, hasMax := optionalNumber(block, "max_cu_limit")
	if hasElastic && hasMax {
		return flinkcapacity.Capacity{}, fmt.Errorf("elastic_cu_limit and max_cu_limit are mutually exclusive")
	}
	var elastic, max *flinkcapacity.CU
	if hasElastic {
		parsed, parseErr := parseFlinkCoordinatorCU(elasticValue, integerOnly)
		if parseErr != nil {
			return flinkcapacity.Capacity{}, fmt.Errorf("elastic_cu_limit: %w", parseErr)
		}
		elastic = &parsed
	}
	if hasMax {
		parsed, parseErr := parseFlinkCoordinatorCU(maxValue, integerOnly)
		if parseErr != nil {
			return flinkcapacity.Capacity{}, fmt.Errorf("max_cu_limit: %w", parseErr)
		}
		max = &parsed
	}
	return flinkcapacity.NewCapacity(fixed, elastic, max)
}

func parseFlinkCoordinatorCU(value float64, integerOnly bool) (flinkcapacity.CU, error) {
	parsed, err := flinkcapacity.ParseCU(value)
	if err != nil {
		return 0, err
	}
	if integerOnly && parsed%2 != 0 {
		return 0, fmt.Errorf("must use integer CU, got %v", value)
	}
	return parsed, nil
}

func optionalNumber(block map[string]interface{}, key string) (float64, bool) {
	value, exists := block[key]
	if !exists || value == nil {
		return 0, false
	}
	number, ok := numberAsFloat(value)
	// Plugin SDK v1 materializes omitted nested optional numbers as zero in the
	// decoded block. CustomizeDiff preserves presence separately and validates
	// conflicts; at apply time zero has the same capacity effect as omission.
	return number, ok && number != 0
}

func numberAsFloat(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func flattenFlinkCoordinatorImport(actual flinkcapacity.Tree) []interface{} {
	namespaces := append([]flinkcapacity.Namespace(nil), actual.Namespaces...)
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Name < namespaces[j].Name })
	workspace := map[string]interface{}{
		"capacity":          flattenFlinkCoordinatorWorkspaceCapacity(actual.Workspace),
		"observed_capacity": flattenFlinkCoordinatorWorkspaceObserved(actual.Workspace),
	}
	for _, namespace := range namespaces {
		queues := append([]flinkcapacity.Queue(nil), namespace.Queues...)
		sort.Slice(queues, func(i, j int) bool { return queues[i].Name < queues[j].Name })
		namespaceMap := map[string]interface{}{
			"name":              namespace.Name,
			"capacity":          flattenFlinkCoordinatorCapacity(*namespace.Capacity),
			"observed_capacity": flattenFlinkCoordinatorObserved(*namespace.Capacity, namespace.Used),
		}
		for _, queue := range queues {
			namespaceMap["queue"] = appendInterface(namespaceMap["queue"], map[string]interface{}{
				"name":              queue.Name,
				"capacity":          flattenFlinkCoordinatorCapacity(*queue.Capacity),
				"observed_capacity": flattenFlinkCoordinatorObserved(*queue.Capacity, queue.Used),
			})
		}
		workspace["namespace"] = appendInterface(workspace["namespace"], namespaceMap)
	}
	return []interface{}{workspace}
}

func mergeFlinkCoordinatorObserved(configured interface{}, actual flinkcapacity.Tree) ([]interface{}, error) {
	workspace, ok := flinkFirstBlock(configured)
	if !ok {
		return nil, fmt.Errorf("workspace block is required")
	}
	merged := cloneFlinkValue(workspace).(map[string]interface{})
	merged["observed_capacity"] = flattenFlinkCoordinatorWorkspaceObserved(actual.Workspace)

	actualNamespaces := make(map[string]flinkcapacity.Namespace, len(actual.Namespaces))
	for _, namespace := range actual.Namespaces {
		actualNamespaces[namespace.Name] = namespace
	}
	namespaceItems, _ := merged["namespace"].([]interface{})
	seenNamespaces := make(map[string]struct{}, len(namespaceItems))
	for _, rawNamespace := range namespaceItems {
		namespaceMap := rawNamespace.(map[string]interface{})
		name, _ := namespaceMap["name"].(string)
		actualNamespace, exists := actualNamespaces[name]
		if !exists {
			return nil, fmt.Errorf("declared namespace %q does not exist in the workspace", name)
		}
		seenNamespaces[name] = struct{}{}
		namespaceMap["observed_capacity"] = flattenFlinkCoordinatorObserved(*actualNamespace.Capacity, actualNamespace.Used)

		actualQueues := make(map[string]flinkcapacity.Queue, len(actualNamespace.Queues))
		for _, queue := range actualNamespace.Queues {
			actualQueues[queue.Name] = queue
		}
		queueItems, _ := namespaceMap["queue"].([]interface{})
		seenQueues := make(map[string]struct{}, len(queueItems))
		for _, rawQueue := range queueItems {
			queueMap := rawQueue.(map[string]interface{})
			queueName, _ := queueMap["name"].(string)
			actualQueue, exists := actualQueues[queueName]
			if !exists {
				return nil, fmt.Errorf("declared queue %q/%q does not exist in the workspace", name, queueName)
			}
			seenQueues[queueName] = struct{}{}
			queueMap["observed_capacity"] = flattenFlinkCoordinatorObserved(*actualQueue.Capacity, actualQueue.Used)
		}
		for queueName := range actualQueues {
			if _, exists := seenQueues[queueName]; !exists {
				return nil, fmt.Errorf("cloud queue %q/%q is undeclared", name, queueName)
			}
		}
	}
	for name := range actualNamespaces {
		if _, exists := seenNamespaces[name]; !exists {
			return nil, fmt.Errorf("cloud namespace %q is undeclared", name)
		}
	}
	return []interface{}{merged}, nil
}

func flattenFlinkCoordinatorWorkspaceCapacity(capacity flinkcapacity.WorkspaceCapacity) []interface{} {
	result := firstMap(flattenFlinkCoordinatorCapacity(capacity.AsCapacity()))
	result["fixed_cu"] = capacity.FixedCU.Float64()
	if capacity.HA {
		result["ha"] = []interface{}{map[string]interface{}{"cross_zone_fixed_cu": capacity.CrossZoneFixedCU.Float64()}}
	}
	return []interface{}{result}
}

func flattenFlinkCoordinatorCapacity(capacity flinkcapacity.Capacity) []interface{} {
	return []interface{}{map[string]interface{}{
		"fixed_cu":         capacity.Fixed.Float64(),
		"elastic_cu_limit": capacity.Elastic().Float64(),
	}}
}

func flattenFlinkCoordinatorWorkspaceObserved(capacity flinkcapacity.WorkspaceCapacity) []interface{} {
	return flattenFlinkObservedCapacity(
		capacity.FixedCU.Float64(),
		capacity.CrossZoneFixedCU.Float64(),
		capacity.Limit.Float64(),
		capacity.Used,
		true,
	)
}

func flattenFlinkCoordinatorObserved(capacity flinkcapacity.Capacity, used float64) []interface{} {
	return flattenFlinkObservedCapacity(capacity.Fixed.Float64(), 0, capacity.Limit.Float64(), used, false)
}

func firstMap(value []interface{}) map[string]interface{} {
	return value[0].(map[string]interface{})
}

func appendInterface(value interface{}, item interface{}) []interface{} {
	items, _ := value.([]interface{})
	return append(items, item)
}

func cloneFlinkValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			result[key] = cloneFlinkValue(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for i, item := range typed {
			result[i] = cloneFlinkValue(item)
		}
		return result
	default:
		return typed
	}
}
