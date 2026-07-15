package alicloud

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkWorkspace() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAliCloudFlinkWorkspaceCreate,
		Read:          resourceAliCloudFlinkWorkspaceRead,
		Update:        resourceAliCloudFlinkWorkspaceUpdate,
		Delete:        resourceAliCloudFlinkWorkspaceDelete,
		CustomizeDiff: flinkWorkspaceCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Flink instance.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the resource group.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Deprecated:  "zone_id is derived from vswitch_ids; keep it only for compatibility with existing configurations.",
				Description: "The zone ID where the Flink instance is located.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VPC ID of the Flink instance.",
			},
			"vswitch_ids": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The IDs of the vSwitches for the Flink instance.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of the security group.",
			},
			"architecture_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "X86",
				ValidateFunc: validation.StringInSlice([]string{"X86", "ARM"}, false),
				Description:  "The architecture type of the Flink instance.",
			},
			"auto_renew": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether the instance automatically renews.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POST",
				ValidateFunc: validation.StringInSlice([]string{"POST", "PRE"}, false),
				Description:  "The billing method of the instance.",
			},
			"duration": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     1,
				Description: "The subscription duration.",
			},
			"pricing_cycle": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "Month",
				Description: "The billing cycle for Subscription instances.",
			},
			"extra": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Additional configuration for the instance.",
			},
			"ha": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:             schema.TypeList,
							Optional:         true,
							ForceNew:         true,
							MaxItems:         1,
							DiffSuppressFunc: suppressFlinkLegacyCapacityDiff,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										ForceNew:    true,
										Description: "CPU specifications for HA resources.",
									},
									"memory": {
										Type:        schema.TypeInt,
										Required:    true,
										ForceNew:    true,
										Description: "Memory specifications for HA resources.",
									},
								},
							},
							Description: "HA resource specifications.",
							Deprecated:  "Use alicloud_flink_capacity_coordinator for new capacity configurations.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The IDs of the vSwitches for high availability.",
						},
						"zone_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
							Deprecated:  "ha.zone_id is derived from ha.vswitch_ids; keep it only for compatibility.",
							Description: "The zone ID for high availability.",
						},
					},
				},
				Description: "High availability configuration.",
			},
			"monitor_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ARMS",
				ValidateFunc: validation.StringInSlice([]string{"ARMS", "TAIHAO"}, true),
				Description:  "The monitoring type of the instance.",
			},
			"promotion_code": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The promotion code.",
			},
			"use_promotion_code": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Whether to use promotion code.",
			},
			"storage": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oss_bucket": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The OSS bucket name for the Flink instance.",
						},
					},
				},
				Description: "Storage configuration of oss bucket for the Flink instance.",
			},
			"resource": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressFlinkLegacyCapacityDiff,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
							Description: "CPU units in millicores.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
							Description: "Memory in MB.",
						},
					},
				},
				Description: "Resource specifications for the Flink instance.",
				Deprecated:  "Use bootstrap_capacity plus alicloud_flink_capacity_coordinator for new configurations.",
			},
			"capacity_management": flinkCapacityManagementSchema(),
			"bootstrap_capacity":  flinkBootstrapCapacitySchema(),
			"observed_capacity":   flinkObservedCapacitySchema(true),
			"resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource ID of the Flink workspace instance.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create workspace request using cws-lib-go types
	workspaceRequest := &aliyunFlinkAPI.Workspace{
		Name:            d.Get("name").(string),
		ResourceGroupId: d.Get("resource_group_id").(string),
		VpcId:           d.Get("vpc_id").(string),
		Region:          client.RegionId,
	}

	// Handle vswitch_ids
	if vswitchIds := d.Get("vswitch_ids").([]interface{}); len(vswitchIds) > 0 {
		workspaceRequest.VSwitchIds = make([]string, len(vswitchIds))
		for i, v := range vswitchIds {
			workspaceRequest.VSwitchIds[i] = v.(string)
		}
	}
	haMap, hasHA := flinkFirstBlock(d.Get("ha"))
	var haVSwitchIDs []string
	if hasHA {
		haVSwitchIDs = flinkStringList(haMap["vswitch_ids"])
	}
	legacyPrimaryZoneID, _ := d.Get("zone_id").(string)
	legacyStandbyZoneID := ""
	if hasHA {
		legacyStandbyZoneID, _ = haMap["zone_id"].(string)
	}
	topology, err := resolveFlinkWorkspaceTopology(client, workspaceRequest.VpcId, legacyPrimaryZoneID, legacyStandbyZoneID, workspaceRequest.VSwitchIds, haVSwitchIDs)
	if err != nil {
		return WrapError(err)
	}
	workspaceRequest.ZoneId = topology.PrimaryZoneID

	// Handle security_group_id
	if sgId, ok := d.GetOk("security_group_id"); ok {
		if workspaceRequest.SecurityGroupInfo == nil {
			workspaceRequest.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{}
		}
		workspaceRequest.SecurityGroupInfo.SecurityGroupId = sgId.(string)
	}

	// Handle architecture_type
	if archType, ok := d.GetOk("architecture_type"); ok {
		workspaceRequest.ArchitectureType = archType.(string)
	}

	// Handle charge_type
	if chargeType, ok := d.GetOk("charge_type"); ok {
		workspaceRequest.ChargeType = chargeType.(string)
	}

	// Handle the create-only capacity source selected by capacity_management.
	if flinkCapacityManagementValue(d.Get("capacity_management")) == CapacityManagedByCoordinator {
		fixedCU, crossZoneFixedCU := expandFlinkBootstrapCapacity(d.Get("bootstrap_capacity"))
		workspaceRequest.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(fixedCU), MemoryGB: float64(fixedCU * 4)}
		if hasHA {
			workspaceRequest.HighAvailability = &aliyunFlinkAPI.HighAvailability{
				Enabled:      true,
				ZoneId:       topology.StandbyZoneID,
				VSwitchIds:   haVSwitchIDs,
				ResourceSpec: &aliyunFlinkAPI.ResourceSpec{Cpu: float64(crossZoneFixedCU), MemoryGB: float64(crossZoneFixedCU * 4)},
			}
		}
	} else if resourceList := d.Get("resource").([]interface{}); len(resourceList) > 0 {
		resourceMap := resourceList[0].(map[string]interface{})
		workspaceRequest.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(resourceMap["cpu"].(int)), MemoryGB: float64(resourceMap["memory"].(int))}
	}

	// Handle storage configuration
	if storageList := d.Get("storage").([]interface{}); len(storageList) > 0 {
		storageMap := storageList[0].(map[string]interface{})
		workspaceRequest.Storage = &aliyunFlinkAPI.Storage{
			Oss: &aliyunFlinkAPI.OSSStorage{
				Bucket: storageMap["oss_bucket"].(string),
			},
		}
	}

	// Handle HA configuration
	if hasHA && flinkCapacityManagementValue(d.Get("capacity_management")) == CapacityManagedByResource {
		// Set high availability flag
		workspaceRequest.HighAvailability = &aliyunFlinkAPI.HighAvailability{
			Enabled: true,
			ZoneId:  topology.StandbyZoneID,
		}

		// Handle HA vswitch IDs
		workspaceRequest.HighAvailability.VSwitchIds = haVSwitchIDs

		// Handle HA resource specs
		if resourceList := haMap["resource"].([]interface{}); len(resourceList) > 0 {
			resourceMap := resourceList[0].(map[string]interface{})
			workspaceRequest.HighAvailability.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{
				Cpu:      float64(resourceMap["cpu"].(int)),
				MemoryGB: float64(resourceMap["memory"].(int)),
			}
		}
	}

	// Create exactly once. Paid CreateInstance has no client token; recovery is
	// performed using a stable provider-owned tag before and after the call.
	autoRenew := d.Get("auto_renew").(bool)
	duration := int32(d.Get("duration").(int))
	usePromotionCode := d.Get("use_promotion_code").(bool)
	createOptions := flinkworkspace.CreateOptions{
		AutoRenew:        &autoRenew,
		Duration:         &duration,
		PricingCycle:     d.Get("pricing_cycle").(string),
		Extra:            d.Get("extra").(string),
		MonitorType:      d.Get("monitor_type").(string),
		PromotionCode:    d.Get("promotion_code").(string),
		UsePromotionCode: &usePromotionCode,
	}
	workspace, err := createFlinkWorkspace(flinkService, workspaceRequest, createOptions, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		var pending *pendingFlinkWorkspaceCreateError
		if errors.As(err, &pending) {
			// helper/schema persists the non-empty ID in the errored create state.
			// Keeping a provider-owned marker prevents a later apply from sending a
			// second paid purchase while the first outcome is still unknown.
			d.SetId(pendingFlinkWorkspaceCreateID(pending.token))
		}
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_workspace", "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	if workspace == nil || workspace.Id == "" {
		return WrapError(Error("Failed to get instance ID from workspace"))
	}

	d.SetId(workspace.Id)

	// Wait for the instance to be in running state using service layer function
	if err := flinkService.WaitForWorkspaceStarting(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkWorkspaceRead(d, meta)
}

type flinkWorkspaceCreateService interface {
	CreateInstance(*aliyunFlinkAPI.Workspace, flinkworkspace.CreateOptions) (*aliyunFlinkAPI.Workspace, error)
	ListInstances() ([]aliyunFlinkAPI.Workspace, error)
}

const pendingFlinkWorkspaceCreateIDPrefix = "terraform-pending-create:"

type pendingFlinkWorkspaceCreateError struct {
	token string
	cause error
}

func (e *pendingFlinkWorkspaceCreateError) Error() string {
	return fmt.Sprintf("CreateInstance outcome is ambiguous and no uniquely tagged workspace became visible before the create timeout; Terraform retained a pending-create marker and will refuse another purchase until the account is checked and any matching workspace is imported: %v", e.cause)
}

func (e *pendingFlinkWorkspaceCreateError) Unwrap() error { return e.cause }

type flinkWorkspaceCreateDiscoveryError struct{ cause error }

func (e *flinkWorkspaceCreateDiscoveryError) Error() string { return e.cause.Error() }
func (e *flinkWorkspaceCreateDiscoveryError) Unwrap() error { return e.cause }

func pendingFlinkWorkspaceCreateID(token string) string {
	if token == "" {
		return ""
	}
	return pendingFlinkWorkspaceCreateIDPrefix + token
}

func flinkWorkspaceCreateTokenFromPendingID(id string) (string, bool) {
	if !strings.HasPrefix(id, pendingFlinkWorkspaceCreateIDPrefix) {
		return "", false
	}
	token := strings.TrimPrefix(id, pendingFlinkWorkspaceCreateIDPrefix)
	return token, token != ""
}

func createFlinkWorkspace(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, options flinkworkspace.CreateOptions, timeout time.Duration) (*aliyunFlinkAPI.Workspace, error) {
	if request == nil {
		return nil, fmt.Errorf("Flink workspace create request is nil")
	}
	token := flinkworkspace.WorkspaceCreateToken(request)
	if token == "" {
		return nil, fmt.Errorf("cannot derive Flink workspace creation token")
	}
	request.Tags = appendFlinkWorkspaceCreateToken(request.Tags, token)

	recovered, err := findFlinkWorkspaceByCreateToken(service, request, token)
	if err != nil {
		return nil, fmt.Errorf("check for a previously accepted Flink workspace purchase: %w", err)
	}
	if recovered != nil {
		return recovered, nil
	}

	workspace, createErr := service.CreateInstance(request, options)
	if createErr == nil && workspace != nil && workspace.Id != "" {
		return workspace, nil
	}
	if createErr == nil {
		createErr = &ambiguousFlinkWorkspaceCreateError{cause: fmt.Errorf("CreateInstance returned no workspace ID")}
	}

	if isAmbiguousFlinkWorkspaceCreateError(createErr) {
		// Read-only discovery errors are retryable here: the paid request may
		// already have succeeded, so a transient DescribeInstances failure must
		// not turn into a second purchase on the next apply.
		recovered, recoveryErr := waitForFlinkWorkspaceCreateRecovery(service, request, token, timeout)
		if recoveryErr == nil {
			return recovered, nil
		}
		return nil, &pendingFlinkWorkspaceCreateError{token: token, cause: fmt.Errorf("CreateInstance: %w; recovery: %v", createErr, recoveryErr)}
	}

	// A definitive service rejection should not normally have purchased
	// anything, but perform one final authoritative lookup before returning it.
	recovered, recoveryErr := findFlinkWorkspaceByCreateToken(service, request, token)
	if recoveryErr != nil {
		return nil, fmt.Errorf("CreateInstance failed: %w; recovery by provider creation tag also failed: %v", createErr, recoveryErr)
	}
	if recovered != nil {
		return recovered, nil
	}
	return nil, createErr
}

func appendFlinkWorkspaceCreateToken(tags []aliyunFlinkAPI.Tag, token string) []aliyunFlinkAPI.Tag {
	result := make([]aliyunFlinkAPI.Tag, 0, len(tags)+1)
	for _, tag := range tags {
		if tag.Key != flinkworkspace.CreateTokenTagKey {
			result = append(result, tag)
		}
	}
	return append(result, aliyunFlinkAPI.Tag{Key: flinkworkspace.CreateTokenTagKey, Value: token})
}

func waitForFlinkWorkspaceCreateRecovery(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, token string, timeout time.Duration) (*aliyunFlinkAPI.Workspace, error) {
	var recovered *aliyunFlinkAPI.Workspace
	err := resource.Retry(timeout, func() *resource.RetryError {
		workspace, err := findFlinkWorkspaceByCreateToken(service, request, token)
		if err != nil {
			var discoveryErr *flinkWorkspaceCreateDiscoveryError
			if errors.As(err, &discoveryErr) || NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if workspace == nil {
			return resource.RetryableError(fmt.Errorf("tagged Flink workspace is not visible yet"))
		}
		recovered = workspace
		return nil
	})
	return recovered, err
}

func findFlinkWorkspaceByCreateToken(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, token string) (*aliyunFlinkAPI.Workspace, error) {
	workspaces, err := service.ListInstances()
	if err != nil {
		return nil, &flinkWorkspaceCreateDiscoveryError{cause: err}
	}
	matches := make([]aliyunFlinkAPI.Workspace, 0, 1)
	for _, workspace := range workspaces {
		if !flinkWorkspaceHasTag(workspace.Tags, flinkworkspace.CreateTokenTagKey, token) {
			continue
		}
		matches = append(matches, workspace)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple Flink workspaces carry the same provider creation token; refusing another purchase")
	}
	match := matches[0]
	if match.Id == "" {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has no instance ID")
	}
	if match.Name != "" && match.Name != request.Name {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has name %q, expected %q", match.Name, request.Name)
	}
	if match.Region != "" && match.Region != request.Region {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has region %q, expected %q", match.Region, request.Region)
	}
	if err := validateRecoveredFlinkWorkspace(match, request); err != nil {
		return nil, err
	}
	return &match, nil
}

func validateRecoveredFlinkWorkspace(match aliyunFlinkAPI.Workspace, request *aliyunFlinkAPI.Workspace) error {
	for _, field := range []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "VPC", actual: match.VpcId, expected: request.VpcId},
		{name: "charge type", actual: match.ChargeType, expected: request.ChargeType},
		{name: "architecture type", actual: match.ArchitectureType, expected: request.ArchitectureType},
		{name: "resource group", actual: match.ResourceGroupId, expected: request.ResourceGroupId},
	} {
		if field.actual != "" && field.expected != "" && field.actual != field.expected {
			return fmt.Errorf("the workspace carrying the provider creation token has %s %q, expected %q", field.name, field.actual, field.expected)
		}
	}
	if len(match.VSwitchIds) > 0 && !sameFlinkStringSet(match.VSwitchIds, request.VSwitchIds) {
		return fmt.Errorf("the workspace carrying the provider creation token has different primary vSwitch IDs")
	}
	requestHA := request.HighAvailability != nil && request.HighAvailability.Enabled
	if match.Ha && !requestHA {
		return fmt.Errorf("the workspace carrying the provider creation token is HA but the request is non-HA")
	}
	if match.Status == "RUNNING" && requestHA && !match.Ha {
		return fmt.Errorf("the workspace carrying the provider creation token is non-HA but the request is HA")
	}
	if len(match.HaVSwitchIds) > 0 {
		if !requestHA || !sameFlinkStringSet(match.HaVSwitchIds, request.HighAvailability.VSwitchIds) {
			return fmt.Errorf("the workspace carrying the provider creation token has different HA vSwitch IDs")
		}
	}
	if match.Storage != nil && match.Storage.Oss != nil && request.Storage != nil && request.Storage.Oss != nil && match.Storage.Oss.Bucket != "" && match.Storage.Oss.Bucket != request.Storage.Oss.Bucket {
		return fmt.Errorf("the workspace carrying the provider creation token has OSS bucket %q, expected %q", match.Storage.Oss.Bucket, request.Storage.Oss.Bucket)
	}
	return nil
}

func sameFlinkStringSet(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	counts := make(map[string]int, len(actual))
	for _, value := range actual {
		counts[value]++
	}
	for _, value := range expected {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	return true
}

func flinkWorkspaceHasTag(tags []aliyunFlinkAPI.Tag, key, value string) bool {
	for _, tag := range tags {
		if tag.Key == key && tag.Value == value {
			return true
		}
	}
	return false
}

func checkPendingFlinkWorkspaceCreate(service flinkWorkspaceCreateService, pendingID, name, region string) error {
	token, ok := flinkWorkspaceCreateTokenFromPendingID(pendingID)
	if !ok {
		return nil
	}
	request := &aliyunFlinkAPI.Workspace{Name: name, Region: region}
	workspace, err := findFlinkWorkspaceByCreateToken(service, request, token)
	if err != nil {
		return fmt.Errorf("resolve pending Flink workspace purchase: %w", err)
	}
	if workspace == nil {
		return fmt.Errorf("Flink workspace purchase outcome is still unresolved; the pending-create marker is retained and another purchase is blocked")
	}
	return fmt.Errorf("Flink workspace purchase became visible as instance %q; the pending-create marker is retained to prevent automatic replacement: remove the pending state and import that instance before applying again", workspace.Id)
}

func resourceAliCloudFlinkWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}
	if _, pending := flinkWorkspaceCreateTokenFromPendingID(d.Id()); pending {
		return WrapError(checkPendingFlinkWorkspaceCreate(flinkService, d.Id(), d.Get("name").(string), client.RegionId))
	}

	workspace, err := flinkService.DescribeFlinkWorkspace(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_workspace DescribeFlinkWorkspace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes from workspace
	d.Set("name", workspace.Name)
	d.Set("resource_group_id", workspace.ResourceGroupId)

	configuredPrimaryZoneID, _ := d.Get("zone_id").(string)
	if zoneID := flinkworkspace.PrimaryZoneID(workspace, configuredPrimaryZoneID); zoneID != "" {
		d.Set("zone_id", zoneID)
	}

	d.Set("vpc_id", workspace.VpcId)

	// Handle vswitch_ids with proper type conversion
	if workspace.VSwitchIds != nil && len(workspace.VSwitchIds) > 0 {
		d.Set("vswitch_ids", workspace.VSwitchIds)
	} else {
		// Preserve configured vswitch_ids if API doesn't return them
		if configuredVSwitchIds := d.Get("vswitch_ids").([]interface{}); len(configuredVSwitchIds) > 0 {
			// Keep the configured values to prevent plan changes
		}
	}

	// Handle security group from SecurityGroupInfo structure
	if workspace.SecurityGroupInfo != nil && workspace.SecurityGroupInfo.SecurityGroupId != "" {
		d.Set("security_group_id", workspace.SecurityGroupInfo.SecurityGroupId)
	}

	// Set fields that are returned by the API with fallback to configured values
	if workspace.ArchitectureType != "" {
		d.Set("architecture_type", workspace.ArchitectureType)
	}
	if workspace.ChargeType != "" {
		d.Set("charge_type", workspace.ChargeType)
	}
	if workspace.MonitorType != "" {
		d.Set("monitor_type", workspace.MonitorType)
	}

	// Always set resource_id if available
	if workspace.ResourceId != "" {
		d.Set("resource_id", workspace.ResourceId)
	}

	capacityManagement := flinkCapacityManagementValue(d.Get("capacity_management"))
	d.Set("observed_capacity", flattenFlinkWorkspaceObservedCapacity(workspace))

	// Set legacy capacity intent only while this resource owns capacity.
	if capacityManagement == CapacityManagedByResource && workspace.ResourceSpec != nil {
		resourceConfig := map[string]interface{}{
			"cpu":    int(workspace.ResourceSpec.Cpu),
			"memory": int(workspace.ResourceSpec.MemoryGB),
		}
		d.Set("resource", []interface{}{resourceConfig})
	} else if capacityManagement == CapacityManagedByCoordinator {
		d.Set("resource", nil)
	}

	// Set storage configuration
	if workspace.Storage != nil && workspace.Storage.Oss != nil {
		storageConfig := map[string]interface{}{
			"oss_bucket": workspace.Storage.Oss.Bucket,
		}
		d.Set("storage", []interface{}{storageConfig})
	}

	// Set HA configuration. DescribeInstances returns the flat Ha* fields,
	// while the create path uses HighAvailability.
	configuredStandbyZoneID := ""
	if configuredHA, ok := flinkFirstBlock(d.Get("ha")); ok {
		configuredStandbyZoneID, _ = configuredHA["zone_id"].(string)
	}
	if haConfig, ok := flinkworkspace.HAConfigWithFallback(workspace, configuredStandbyZoneID); ok {
		if capacityManagement == CapacityManagedByCoordinator {
			delete(haConfig, "resource")
		}
		if haResource, ok := haConfig["resource"]; ok {
			haConfig["resource"] = []interface{}{haResource}
		}
		d.Set("ha", []interface{}{haConfig})
	} else {
		// Set empty HA config if not enabled
		d.Set("ha", []interface{}{})
	}

	// Handle input-only fields that are not returned by the API
	// These fields are used only during creation and are not returned by DescribeInstances
	// We need to preserve their configured values to avoid terraform plan showing changes

	// For auto_renew: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("auto_renew"); !ok {
		// Only set default if not already configured
		d.Set("auto_renew", true)
	}

	// For duration: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("duration"); !ok {
		// Only set default if not already configured
		d.Set("duration", 1)
	}

	// For pricing_cycle: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("pricing_cycle"); !ok {
		// Only set default if not already configured
		d.Set("pricing_cycle", "Month")
	}

	// For extra, promotion_code, use_promotion_code: these are input-only fields
	// They don't need to be set here as they are only used during creation

	return nil
}

func resourceAliCloudFlinkWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	// Capacity changes are either replacement-only legacy changes or are owned
	// by alicloud_flink_capacity_coordinator. The workspace resource currently
	// has no other in-place mutable fields.
	return resourceAliCloudFlinkWorkspaceRead(d, meta)
}

func resourceAliCloudFlinkWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}
	if _, pending := flinkWorkspaceCreateTokenFromPendingID(d.Id()); pending {
		return WrapError(checkPendingFlinkWorkspaceCreate(flinkService, d.Id(), d.Get("name").(string), client.RegionId))
	}

	if err := deleteFlinkWorkspace(flinkService, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}

	return nil
}

type flinkWorkspaceDeleteService interface {
	DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error)
	DeleteInstance(string) error
	RefundInstance(string) error
	WaitForWorkspaceDeleting(string, time.Duration) error
}

func deleteFlinkWorkspace(service flinkWorkspaceDeleteService, instanceID string, timeout time.Duration) error {
	workspace, err := service.DescribeFlinkWorkspace(instanceID)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return fmt.Errorf("read Flink workspace %q before deletion: %w", instanceID, err)
	}
	if workspace == nil {
		return fmt.Errorf("read Flink workspace %q before deletion returned nil", instanceID)
	}
	switch workspace.ChargeType {
	case "POST":
		err = service.DeleteInstance(instanceID)
	case "PRE":
		err = service.RefundInstance(instanceID)
	default:
		return fmt.Errorf("cannot delete Flink workspace %q with unknown charge type %q", instanceID, workspace.ChargeType)
	}
	if err != nil {
		_, verifyErr := service.DescribeFlinkWorkspace(instanceID)
		if NotFoundError(verifyErr) {
			return nil
		}
		if verifyErr != nil {
			return fmt.Errorf("delete Flink workspace %q failed: %w; verifying whether the workspace still exists also failed: %v", instanceID, err, verifyErr)
		}
		return err
	}
	if err := service.WaitForWorkspaceDeleting(instanceID, timeout); err != nil {
		return fmt.Errorf("wait for Flink workspace %q deletion: %w", instanceID, err)
	}
	return nil
}

func resolveFlinkWorkspaceTopology(client *connectivity.AliyunClient, vpcID, legacyPrimaryZoneID, legacyStandbyZoneID string, primaryIDs, standbyIDs []string) (flinkworkspace.VSwitchTopology, error) {
	vpcService := VpcService{client}
	describe := func(ids []string) ([]flinkworkspace.VSwitch, error) {
		result := make([]flinkworkspace.VSwitch, 0, len(ids))
		for _, id := range ids {
			vSwitch, err := vpcService.DescribeVSwitch(id)
			if err != nil {
				return nil, err
			}
			result = append(result, flinkworkspace.VSwitch{
				ID:       vSwitch.VSwitchId,
				RegionID: client.RegionId,
				VPCID:    vSwitch.VpcId,
				ZoneID:   vSwitch.ZoneId,
			})
		}
		return result, nil
	}
	primary, err := describe(primaryIDs)
	if err != nil {
		return flinkworkspace.VSwitchTopology{}, err
	}
	standby, err := describe(standbyIDs)
	if err != nil {
		return flinkworkspace.VSwitchTopology{}, err
	}
	return flinkworkspace.ValidateVSwitchTopology(client.RegionId, vpcID, legacyPrimaryZoneID, legacyStandbyZoneID, primary, standby)
}

func flinkStringList(value interface{}) []string {
	items, _ := value.([]interface{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok && text != "" {
			result = append(result, text)
		}
	}
	return result
}
