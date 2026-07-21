package alicloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// resourceAliCloudFlinkWorkspaceCapacityAllocation owns the mutable capacity
// tree attached to an existing Workspace. Its ID deliberately is the Workspace
// ID: deleting this Terraform attachment must not release any cloud resource.
func resourceAliCloudFlinkWorkspaceCapacityAllocation() *schema.Resource {
	namespace := &schema.Resource{Schema: map[string]*schema.Schema{
		"name": {
			Type: schema.TypeString, Required: true, ValidateFunc: validation.StringIsNotEmpty,
		},
		"fixed_cu": {
			Type: schema.TypeInt, Required: true, ValidateFunc: validateFlinkCUFitsInt32Memory,
		},
		"max_cu_limit": {
			Type: schema.TypeInt, Required: true, ValidateFunc: validateFlinkCUFitsInt32Memory,
		},
	}}
	return &schema.Resource{
		Create:        resourceAliCloudFlinkWorkspaceCapacityAllocationCreate,
		Read:          resourceAliCloudFlinkWorkspaceCapacityAllocationRead,
		Update:        resourceAliCloudFlinkWorkspaceCapacityAllocationUpdate,
		Delete:        resourceAliCloudFlinkWorkspaceCapacityAllocationDelete,
		CustomizeDiff: flinkWorkspaceCapacityAllocationCustomizeDiff,
		Importer: &schema.ResourceImporter{State: func(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
			if d.Id() == "" {
				return nil, fmt.Errorf("workspace_instance_id import ID must not be empty")
			}
			if err := d.Set("workspace_instance_id", d.Id()); err != nil {
				return nil, err
			}
			return []*schema.ResourceData{d}, nil
		}},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"workspace_instance_id": {
				Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.StringIsNotEmpty,
				Description: "Identifies the Workspace exclusively managed by this allocation. Provider locking serializes callbacks only inside one provider process; configurations must ensure one allocation owner across states.",
			},
			"fixed_cu": {
				Type: schema.TypeInt, Required: true, ValidateFunc: validateFlinkCUFitsInt32Memory,
			},
			"cross_zone_fixed_cu": {
				Type: schema.TypeInt, Required: true, ValidateFunc: validateFlinkCUFitsInt32Memory,
			},
			"max_cu_limit": {
				Type: schema.TypeInt, Required: true, ValidateFunc: validateFlinkCUFitsInt32Memory,
			},
			"namespace": {
				Type: schema.TypeSet, Required: true, MinItems: 1, Elem: namespace, Set: schema.HashResource(namespace),
			},
			"implicit_topology_hash": {
				Type: schema.TypeString, Computed: true,
			},
			"observed_capacity_tree": flinkWorkspaceCapacityAllocationObservedTreeSchema(),
		},
	}
}

func flinkWorkspaceCapacityAllocationObservedTreeSchema() *schema.Schema {
	queue := &schema.Resource{Schema: map[string]*schema.Schema{
		"name": {Type: schema.TypeString, Computed: true}, "fixed_cu": {Type: schema.TypeFloat, Computed: true},
		"max_cu_limit": {Type: schema.TypeFloat, Computed: true}, "used_cu": {Type: schema.TypeFloat, Computed: true},
	}}
	namespace := &schema.Resource{Schema: map[string]*schema.Schema{
		"name": {Type: schema.TypeString, Computed: true}, "cross_zone": {Type: schema.TypeBool, Computed: true},
		"fixed_cu": {Type: schema.TypeFloat, Computed: true}, "max_cu_limit": {Type: schema.TypeFloat, Computed: true},
		"used_cu": {Type: schema.TypeFloat, Computed: true}, "queue": {Type: schema.TypeList, Computed: true, Elem: queue},
	}}
	return &schema.Schema{Type: schema.TypeList, Computed: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
		"fixed_cu": {Type: schema.TypeFloat, Computed: true}, "cross_zone_fixed_cu": {Type: schema.TypeFloat, Computed: true},
		"max_cu_limit": {Type: schema.TypeFloat, Computed: true}, "used_cu": {Type: schema.TypeFloat, Computed: true},
		"namespace": {Type: schema.TypeList, Computed: true, Elem: namespace},
	}}}
}

func flinkWorkspaceCapacityAllocationCustomizeDiff(d *schema.ResourceDiff, _ interface{}) error {
	if !flinkWorkspaceCapacityAllocationDesiredKnown(d) {
		return d.SetNewComputed("implicit_topology_hash")
	}
	desired, err := expandFlinkWorkspaceCapacityAllocationDesired(map[string]interface{}{
		"fixed_cu": d.Get("fixed_cu"), "cross_zone_fixed_cu": d.Get("cross_zone_fixed_cu"),
		"max_cu_limit": d.Get("max_cu_limit"), "namespace": d.Get("namespace"),
	})
	if err != nil {
		return err
	}
	if d.Id() != "" && desired.Workspace.AsCapacity().Elastic() == 0 && flinkWorkspaceCapacityAllocationObservedElastic(d.Get("observed_capacity_tree")) > 0 {
		return fmt.Errorf("cannot reduce workspace elastic CU to zero through a supported public API; disable elastic billing manually, refresh state, and plan again")
	}
	hash, err := flinkWorkspaceCapacityAllocationTopologyHash(desired)
	if err != nil {
		return err
	}
	return d.SetNew("implicit_topology_hash", hash)
}

func flinkWorkspaceCapacityAllocationDesiredKnown(d *schema.ResourceDiff) bool {
	for _, key := range []string{"workspace_instance_id", "fixed_cu", "cross_zone_fixed_cu", "max_cu_limit", "namespace"} {
		if !d.NewValueKnown(key) {
			return false
		}
	}
	for _, key := range d.GetChangedKeysPrefix("namespace") {
		if !d.NewValueKnown(key) {
			return false
		}
	}
	return true
}

func flinkWorkspaceCapacityAllocationObservedElastic(value interface{}) float64 {
	workspace, ok := flinkFirstBlock(value)
	if !ok {
		return 0
	}
	fixed, _ := numberAsFloat(workspace["fixed_cu"])
	cross, _ := numberAsFloat(workspace["cross_zone_fixed_cu"])
	limit, _ := numberAsFloat(workspace["max_cu_limit"])
	return limit - fixed - cross
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("workspace_instance_id").(string))
	return resourceAliCloudFlinkWorkspaceCapacityAllocationReconcile(d, meta, schema.TimeoutCreate)
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAliCloudFlinkWorkspaceCapacityAllocationReconcile(d, meta, schema.TimeoutUpdate)
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationReconcile(d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	instanceID := d.Get("workspace_instance_id").(string)
	return withFlinkWorkspaceCapacityAllocationLock(instanceID, func() error {
		return resourceAliCloudFlinkWorkspaceCapacityAllocationReconcileUnlocked(d, meta, timeoutKey)
	})
}

func withFlinkWorkspaceCapacityAllocationLock(instanceID string, reconcile func() error) error {
	lockKey := fmt.Sprintf("flink-capacity-allocation-%s", instanceID)
	alicloudMutexKV.Lock(lockKey)
	defer alicloudMutexKV.Unlock(lockKey)
	return reconcile()
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationReconcileUnlocked(d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	service, err := newFlinkWorkspaceCapacityAllocationService(meta)
	if err != nil {
		return WrapError(err)
	}
	desired, err := expandFlinkWorkspaceCapacityAllocationDesiredValues(d.Get("namespace"), d.Get("fixed_cu"), d.Get("cross_zone_fixed_cu"), d.Get("max_cu_limit"))
	if err != nil {
		return WrapError(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(timeoutKey))
	defer cancel()
	actual, reconcileErr := (flinkcapacity.Reconciler{API: service}).ReconcileAuthoritative(ctx, d.Id(), desired)
	if reconcileErr != nil {
		if stateErr := setFlinkWorkspaceCapacityAllocationState(d, actual); stateErr != nil {
			return WrapError(fmt.Errorf("%w; failed to persist last actual capacity state: %v", reconcileErr, stateErr))
		}
		return WrapError(reconcileErr)
	}
	return WrapError(setFlinkWorkspaceCapacityAllocationState(d, actual))
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationRead(d *schema.ResourceData, meta interface{}) error {
	service, err := newFlinkWorkspaceCapacityAllocationService(meta)
	if err != nil {
		return WrapError(err)
	}
	actual, workspaceAuthoritativelyAbsent, err := readFlinkWorkspaceCapacityAllocationTree(service, d.Id())
	if workspaceAuthoritativelyAbsent {
		d.SetId("")
		return nil
	}
	if err != nil {
		return WrapError(err)
	}
	if err := validateFlinkWorkspaceCapacityAllocationActual(actual); err != nil {
		return WrapError(err)
	}
	if err := d.Set("workspace_instance_id", d.Id()); err != nil {
		return WrapError(err)
	}
	return WrapError(setFlinkWorkspaceCapacityAllocationState(d, actual))
}

func readFlinkWorkspaceCapacityAllocationTree(service flinkcapacity.API, instanceID string) (flinkcapacity.Tree, bool, error) {
	if resultAPI, ok := service.(flinkCapacityReadResultAPI); ok {
		result, err := resultAPI.readTreeResult(context.Background(), instanceID)
		return result.tree, result.workspaceAuthoritativelyAbsent, err
	}
	tree, err := service.ReadTree(context.Background(), instanceID)
	return tree, false, err
}

func resourceAliCloudFlinkWorkspaceCapacityAllocationDelete(d *schema.ResourceData, _ interface{}) error {
	d.SetId("")
	return nil
}

var newFlinkWorkspaceCapacityAllocationService = func(meta interface{}) (flinkcapacity.API, error) {
	client, ok := meta.(*connectivity.AliyunClient)
	if !ok || client == nil {
		return nil, fmt.Errorf("Flink capacity allocation requires an Aliyun client")
	}
	return NewFlinkCapacityService(client)
}

func expandFlinkWorkspaceCapacityAllocationDesired(config map[string]interface{}) (flinkcapacity.Tree, error) {
	return expandFlinkWorkspaceCapacityAllocationDesiredValues(config["namespace"], config["fixed_cu"], config["cross_zone_fixed_cu"], config["max_cu_limit"])
}

func expandFlinkWorkspaceCapacityAllocationDesiredValues(namespacesValue, fixedValue, crossZoneValue, maxValue interface{}) (flinkcapacity.Tree, error) {
	fixed, err := flinkWorkspaceCapacityAllocationIntegerCU(fixedValue, "fixed_cu")
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	crossZone, err := flinkWorkspaceCapacityAllocationIntegerCU(crossZoneValue, "cross_zone_fixed_cu")
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	limit, err := flinkWorkspaceCapacityAllocationIntegerCU(maxValue, "max_cu_limit")
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	ha := crossZone > 0
	if fixed > 0 && crossZone > 0 {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace must use one pure fixed CU pool, not both single-zone and cross-zone")
	}
	if !ha && fixed <= 0 {
		return flinkcapacity.Tree{}, fmt.Errorf("single-zone workspace fixed_cu must be greater than zero")
	}
	if ha && fixed != 0 {
		return flinkcapacity.Tree{}, fmt.Errorf("cross-zone workspace fixed_cu must be zero")
	}
	if limit < fixed+crossZone {
		return flinkcapacity.Tree{}, fmt.Errorf("max_cu_limit %v must be greater than or equal to fixed CU %v", limit.Float64(), (fixed + crossZone).Float64())
	}

	items := flinkWorkspaceCapacityAllocationNamespaceItems(namespacesValue)
	if len(items) == 0 {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace must declare at least one namespace")
	}
	tree := flinkcapacity.Tree{ChargeType: "PRE", Workspace: flinkcapacity.WorkspaceCapacity{HA: ha, FixedCU: fixed, CrossZoneFixedCU: crossZone, Limit: limit}}
	seen := make(map[string]struct{}, len(items))
	for _, raw := range items {
		namespace, ok := raw.(map[string]interface{})
		if !ok || namespace == nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace block must not be empty")
		}
		name, _ := namespace["name"].(string)
		if name == "" {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace name must not be empty")
		}
		if _, exists := seen[name]; exists {
			return flinkcapacity.Tree{}, fmt.Errorf("duplicate namespace %q", name)
		}
		seen[name] = struct{}{}
		namespaceFixed, err := flinkWorkspaceCapacityAllocationIntegerCU(namespace["fixed_cu"], fmt.Sprintf("namespace %q fixed_cu", name))
		if err != nil {
			return flinkcapacity.Tree{}, err
		}
		namespaceLimit, err := flinkWorkspaceCapacityAllocationIntegerCU(namespace["max_cu_limit"], fmt.Sprintf("namespace %q max_cu_limit", name))
		if err != nil {
			return flinkcapacity.Tree{}, err
		}
		if namespaceFixed < 2 {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q fixed CU must be at least 1", name)
		}
		if namespaceLimit < namespaceFixed {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q max_cu_limit must be greater than or equal to fixed CU", name)
		}
		capacity := flinkcapacity.Capacity{Fixed: namespaceFixed, Limit: namespaceLimit}
		tree.Namespaces = append(tree.Namespaces, flinkcapacity.Namespace{Name: name, CrossZone: ha, Capacity: &capacity, Queues: []flinkcapacity.Queue{{Name: "default-queue", Capacity: &capacity}}})
	}
	if err := flinkcapacity.ValidateDesired(tree); err != nil {
		return flinkcapacity.Tree{}, err
	}
	return tree, nil
}

func flinkWorkspaceCapacityAllocationNamespaceItems(value interface{}) []interface{} {
	switch typed := value.(type) {
	case *schema.Set:
		return typed.List()
	case []interface{}:
		return typed
	default:
		return nil
	}
}

func flinkWorkspaceCapacityAllocationIntegerCU(value interface{}, name string) (flinkcapacity.CU, error) {
	number, ok := numberAsFloat(value)
	if !ok {
		return 0, fmt.Errorf("%s must be an integer CU", name)
	}
	if number > float64(flinkMaxCUBeforeInt32MemoryOverflow) {
		return 0, fmt.Errorf("%s must not exceed %d so MemoryGB=CU*4 fits int32", name, flinkMaxCUBeforeInt32MemoryOverflow)
	}
	parsed, err := flinkcapacity.ParseCU(number)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	if parsed%2 != 0 {
		return 0, fmt.Errorf("%s must use integer CU, got %v", name, number)
	}
	return parsed, nil
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

func appendInterface(value interface{}, item interface{}) []interface{} {
	items, _ := value.([]interface{})
	return append(items, item)
}

func setFlinkWorkspaceCapacityAllocationState(d *schema.ResourceData, actual flinkcapacity.Tree) error {
	if err := validateFlinkWorkspaceCapacityAllocationActual(actual); err != nil {
		return err
	}
	if err := d.Set("fixed_cu", int(actual.Workspace.FixedCU/2)); err != nil {
		return err
	}
	if err := d.Set("cross_zone_fixed_cu", int(actual.Workspace.CrossZoneFixedCU/2)); err != nil {
		return err
	}
	if err := d.Set("max_cu_limit", int(actual.Workspace.Limit/2)); err != nil {
		return err
	}
	if err := d.Set("namespace", flattenFlinkWorkspaceCapacityAllocationNamespaces(actual)); err != nil {
		return err
	}
	hash, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		return err
	}
	if err := d.Set("implicit_topology_hash", hash); err != nil {
		return err
	}
	return d.Set("observed_capacity_tree", flattenFlinkWorkspaceCapacityAllocationObservedTree(actual))
}

func validateFlinkWorkspaceCapacityAllocationActual(actual flinkcapacity.Tree) error {
	if actual.ChargeType != "PRE" {
		return fmt.Errorf("Flink workspace capacity allocation supports only PRE workspaces, got %q", actual.ChargeType)
	}
	return nil
}

func flattenFlinkWorkspaceCapacityAllocationNamespaces(tree flinkcapacity.Tree) []interface{} {
	namespaces := append([]flinkcapacity.Namespace(nil), tree.Namespaces...)
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Name < namespaces[j].Name })
	result := make([]interface{}, 0, len(namespaces))
	for _, namespace := range namespaces {
		if namespace.Capacity == nil {
			continue
		}
		result = append(result, map[string]interface{}{"name": namespace.Name, "fixed_cu": int(namespace.Capacity.Fixed / 2), "max_cu_limit": int(namespace.Capacity.Limit / 2)})
	}
	return result
}

type flinkWorkspaceCapacityAllocationHashNamespace struct {
	Name      string                                      `json:"name"`
	CrossZone bool                                        `json:"cross_zone"`
	Queues    []flinkWorkspaceCapacityAllocationHashQueue `json:"queues"`
}
type flinkWorkspaceCapacityAllocationHashQueue struct {
	Name    string           `json:"name"`
	Request flinkcapacity.CU `json:"request_half_cu"`
	Limit   flinkcapacity.CU `json:"limit_half_cu"`
}

func flinkWorkspaceCapacityAllocationTopologyHash(tree flinkcapacity.Tree) (string, error) {
	namespaces := append([]flinkcapacity.Namespace(nil), tree.Namespaces...)
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Name < namespaces[j].Name })
	canonical := make([]flinkWorkspaceCapacityAllocationHashNamespace, 0, len(namespaces))
	for _, namespace := range namespaces {
		entry := flinkWorkspaceCapacityAllocationHashNamespace{Name: namespace.Name, CrossZone: namespace.CrossZone}
		queues := append([]flinkcapacity.Queue(nil), namespace.Queues...)
		sort.Slice(queues, func(i, j int) bool { return queues[i].Name < queues[j].Name })
		for _, queue := range queues {
			if queue.Capacity == nil {
				return "", fmt.Errorf("queue %q/%q capacity is missing", namespace.Name, queue.Name)
			}
			entry.Queues = append(entry.Queues, flinkWorkspaceCapacityAllocationHashQueue{Name: queue.Name, Request: queue.Capacity.Fixed, Limit: queue.Capacity.Limit})
		}
		canonical = append(canonical, entry)
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "v1:" + hex.EncodeToString(digest[:]), nil
}

func flattenFlinkWorkspaceCapacityAllocationObservedTree(tree flinkcapacity.Tree) []interface{} {
	namespaces := append([]flinkcapacity.Namespace(nil), tree.Namespaces...)
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Name < namespaces[j].Name })
	workspace := map[string]interface{}{"fixed_cu": tree.Workspace.FixedCU.Float64(), "cross_zone_fixed_cu": tree.Workspace.CrossZoneFixedCU.Float64(), "max_cu_limit": tree.Workspace.Limit.Float64(), "used_cu": tree.Workspace.Used}
	for _, namespace := range namespaces {
		if namespace.Capacity == nil {
			continue
		}
		entry := map[string]interface{}{"name": namespace.Name, "cross_zone": namespace.CrossZone, "fixed_cu": namespace.Capacity.Fixed.Float64(), "max_cu_limit": namespace.Capacity.Limit.Float64(), "used_cu": namespace.Used}
		queues := append([]flinkcapacity.Queue(nil), namespace.Queues...)
		sort.Slice(queues, func(i, j int) bool { return queues[i].Name < queues[j].Name })
		for _, queue := range queues {
			if queue.Capacity == nil {
				continue
			}
			entry["queue"] = appendInterface(entry["queue"], map[string]interface{}{"name": queue.Name, "fixed_cu": queue.Capacity.Fixed.Float64(), "max_cu_limit": queue.Capacity.Limit.Float64(), "used_cu": queue.Used})
		}
		workspace["namespace"] = appendInterface(workspace["namespace"], entry)
	}
	return []interface{}{workspace}
}
