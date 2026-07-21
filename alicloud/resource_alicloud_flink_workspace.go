package alicloud

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/zclconf/go-cty/cty"
)

const (
	flinkWorkspacePurchaseManaged            = "MANAGED"
	flinkWorkspacePurchaseImportedUnknown    = "IMPORTED_UNKNOWN"
	flinkWorkspacePurchaseRecoveryPending    = "RECOVERY_PENDING"
	flinkWorkspacePurchaseLegacyUnclassified = "LEGACY_UNCLASSIFIED"
	flinkWorkspacePurchaseMigratedInitial    = "MIGRATED_INITIAL_PENDING"

	flinkWorkspaceCapacityLegacy                 = "LEGACY"
	flinkWorkspaceCapacityInitial                = "INITIAL"
	flinkWorkspaceCapacityInitialAdoptionPending = "INITIAL_ADOPTION_PENDING"

	flinkWorkspaceIdentityAwaitingFirstRead = "AWAITING_FIRST_READ"
	flinkWorkspaceIdentityMigratedFirstRead = "MIGRATED_AWAITING_FIRST_READ"
	flinkWorkspaceIdentityStable            = "STABLE"
	flinkWorkspaceProtocolUnavailable       = "UNAVAILABLE"
)

var flinkWorkspaceRecoveryTokenPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func resourceAliCloudFlinkWorkspace() *schema.Resource {
	workspaceResource := &schema.Resource{
		Description:   "Flink workspace. Import accepts <instance-id>, <instance-id>|legacy, <instance-id>|initial, <instance-id>|recover=<64-hex-create-token>|legacy, or <instance-id>|recover=<64-hex-create-token>|initial. API-unobservable purchase options on an imported unknown workspace remain unmanaged. Verified recovery uses one state-only apply that revalidates cloud identity, intent, topology, and capacity without modifying the workspace.",
		Create:        resourceAliCloudFlinkWorkspaceCreate,
		Read:          resourceAliCloudFlinkWorkspaceRead,
		Update:        resourceAliCloudFlinkWorkspaceUpdate,
		Delete:        resourceAliCloudFlinkWorkspaceDelete,
		CustomizeDiff: flinkWorkspaceCustomizeDiff,
		Importer:      &schema.ResourceImporter{State: importAliCloudFlinkWorkspace},
		Schema: map[string]*schema.Schema{
			"identity_visibility_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal provider marker that retains a paid workspace identity until its first authoritative successful Read.",
			},
			"purchase_options_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal provider provenance for API-unobservable purchase options; imported unknown values are not managed by Terraform.",
			},
			"capacity_intent_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal provider marker preserving legacy versus initial capacity intent across import and recovery.",
			},
			"terraform_create_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal provider identity token used only for verified paid-create recovery.",
			},
			"create_intent_fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Internal provider fingerprint used only to verify paid-create intent during recovery.",
			},
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
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "Whether the instance automatically renews. For a generically imported workspace this API-unobservable value remains unmanaged.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POST",
				ValidateFunc: validation.StringInSlice([]string{"POST", "PRE"}, false),
				Description:  "The billing method of the instance.",
			},
			"duration": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          1,
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "The subscription duration. For a generically imported workspace this API-unobservable value remains unmanaged.",
			},
			"pricing_cycle": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "Month",
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "The billing cycle for Subscription instances. For a generically imported workspace this API-unobservable value remains unmanaged.",
			},
			"extra": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "Additional creation configuration. For a generically imported workspace this API-unobservable value remains unmanaged.",
			},
			"ha": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
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
							Deprecated:  "Use initial_capacity for new workspace capacity configurations.",
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
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "The promotion code. For a generically imported workspace this API-unobservable value remains unmanaged.",
			},
			"use_promotion_code": {
				Type:             schema.TypeBool,
				Optional:         true,
				DiffSuppressFunc: suppressUnmanagedFlinkWorkspacePurchaseOptionDiff,
				Description:      "Whether to use promotion code. For a generically imported workspace this API-unobservable value remains unmanaged.",
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
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
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
				Deprecated:  "Use initial_capacity for new workspace capacity configurations.",
			},
			"initial_capacity":  flinkInitialCapacitySchema(),
			"observed_capacity": flinkObservedCapacitySchema(true),
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
	workspaceResource.SchemaVersion = 1
	workspaceResource.StateUpgraders = []schema.StateUpgrader{{
		Version: 0,
		Type:    flinkWorkspaceV0StateType(workspaceResource.Schema),
		Upgrade: upgradeFlinkWorkspaceStateV0,
	}}
	return workspaceResource
}

type flinkWorkspaceImportSpec struct {
	instanceID    string
	purchaseState string
	capacityMode  string
	createToken   string
}

func importAliCloudFlinkWorkspace(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	spec, err := parseFlinkWorkspaceImportID(d.Id())
	if err != nil {
		return nil, err
	}
	d.SetId(spec.instanceID)
	createToken := spec.createToken
	if createToken == "" {
		createToken = flinkWorkspaceProtocolUnavailable
	}
	for key, value := range map[string]string{
		"identity_visibility_state": flinkWorkspaceIdentityStable,
		"purchase_options_state":    spec.purchaseState,
		"capacity_intent_mode":      spec.capacityMode,
		"terraform_create_token":    createToken,
		"create_intent_fingerprint": flinkWorkspaceProtocolUnavailable,
	} {
		if key == "identity_visibility_state" && spec.purchaseState == flinkWorkspacePurchaseRecoveryPending {
			value = flinkWorkspaceIdentityAwaitingFirstRead
		}
		if err := d.Set(key, value); err != nil {
			return nil, fmt.Errorf("set Flink workspace import %s: %w", key, err)
		}
	}
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolExisting); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func parseFlinkWorkspaceImportID(importID string) (flinkWorkspaceImportSpec, error) {
	segments := strings.Split(importID, "|")
	if len(segments) == 0 || segments[0] == "" {
		return flinkWorkspaceImportSpec{}, fmt.Errorf("Flink workspace import ID must start with a non-empty instance ID")
	}
	spec := flinkWorkspaceImportSpec{
		instanceID:    segments[0],
		purchaseState: flinkWorkspacePurchaseImportedUnknown,
		capacityMode:  flinkWorkspaceCapacityLegacy,
	}
	switch len(segments) {
	case 1:
		return spec, nil
	case 2:
		switch segments[1] {
		case "legacy":
			return spec, nil
		case "initial":
			spec.capacityMode = flinkWorkspaceCapacityInitialAdoptionPending
			return spec, nil
		default:
			return flinkWorkspaceImportSpec{}, fmt.Errorf("invalid Flink workspace import mode %q; expected legacy or initial", segments[1])
		}
	case 3:
		if !strings.HasPrefix(segments[1], "recover=") {
			return flinkWorkspaceImportSpec{}, fmt.Errorf("invalid Flink workspace recovery segment %q", segments[1])
		}
		token := strings.TrimPrefix(segments[1], "recover=")
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) {
			return flinkWorkspaceImportSpec{}, fmt.Errorf("Flink workspace recovery token must be exactly 64 lowercase hexadecimal characters")
		}
		spec.purchaseState = flinkWorkspacePurchaseRecoveryPending
		spec.createToken = token
		switch segments[2] {
		case "legacy":
			return spec, nil
		case "initial":
			spec.capacityMode = flinkWorkspaceCapacityInitialAdoptionPending
			return spec, nil
		default:
			return flinkWorkspaceImportSpec{}, fmt.Errorf("verified Flink workspace recovery requires an explicit legacy or initial mode")
		}
	default:
		return flinkWorkspaceImportSpec{}, fmt.Errorf("invalid Flink workspace import ID %q: unknown or duplicate segments", importID)
	}
}

func flinkWorkspaceV0StateType(current map[string]*schema.Schema) cty.Type {
	legacy := make(map[string]*schema.Schema, len(current)-5)
	for name, field := range current {
		switch name {
		case "identity_visibility_state", "purchase_options_state", "capacity_intent_mode", "terraform_create_token", "create_intent_fingerprint":
			continue
		default:
			legacy[name] = field
		}
	}
	return (&schema.Resource{Schema: legacy, Read: schema.Noop}).CoreConfigSchema().ImpliedType()
}

func upgradeFlinkWorkspaceStateV0(raw map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if raw == nil {
		raw = map[string]interface{}{}
	}
	id, _ := raw["id"].(string)
	pending := strings.HasPrefix(id, pendingFlinkWorkspaceCreateIDPrefix)
	hasInitial := flinkWorkspaceRawStateHasBlock(raw["initial_capacity"])
	if pending {
		raw["purchase_options_state"] = flinkWorkspacePurchaseManaged
	} else if hasInitial {
		raw["purchase_options_state"] = flinkWorkspacePurchaseMigratedInitial
	} else {
		raw["purchase_options_state"] = flinkWorkspacePurchaseLegacyUnclassified
	}
	if hasInitial {
		raw["capacity_intent_mode"] = flinkWorkspaceCapacityInitial
	} else {
		raw["capacity_intent_mode"] = flinkWorkspaceCapacityLegacy
	}
	if pending {
		raw["identity_visibility_state"] = flinkWorkspaceIdentityAwaitingFirstRead
	} else if hasInitial {
		raw["identity_visibility_state"] = flinkWorkspaceIdentityMigratedFirstRead
	} else {
		raw["identity_visibility_state"] = flinkWorkspaceIdentityStable
	}
	raw["terraform_create_token"] = flinkWorkspaceProtocolUnavailable
	raw["create_intent_fingerprint"] = flinkWorkspaceProtocolUnavailable
	if pending {
		token := strings.TrimPrefix(id, pendingFlinkWorkspaceCreateIDPrefix)
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) {
			return nil, fmt.Errorf("upgrade Flink workspace pending-create protocol: invalid create token in state ID")
		}
		raw["terraform_create_token"] = token
	} else if hasInitial {
		name, _ := raw["name"].(string)
		client, ok := meta.(*connectivity.AliyunClient)
		if !ok || client == nil || client.RegionId == "" || name == "" {
			return nil, fmt.Errorf("upgrade schema-v0 initial Flink workspace %q: provider Region and persisted name are required to preserve paid identity until its first authoritative Read", id)
		}
		token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: client.RegionId, Name: name})
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) {
			return nil, fmt.Errorf("upgrade schema-v0 initial Flink workspace %q: cannot derive create token", id)
		}
		raw["terraform_create_token"] = token
	}
	return raw, nil
}

func flinkWorkspaceRawStateHasBlock(value interface{}) bool {
	items, ok := value.([]interface{})
	return ok && len(items) > 0 && items[0] != nil
}

func suppressUnmanagedFlinkWorkspacePurchaseOptionDiff(_ string, _, _ string, d *schema.ResourceData) bool {
	if d.Id() == "" {
		return false
	}
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolExisting); err != nil {
		// DiffSuppressFunc cannot return an error. Suppress paid-input diffs for
		// an invalid tuple so CustomizeDiff can reject it without first
		// manufacturing ForceNew or a silent state-only purchase update.
		return true
	}
	switch state, _ := d.Get("purchase_options_state").(string); state {
	case flinkWorkspacePurchaseImportedUnknown:
		return true
	default:
		return false
	}
}

func resourceAliCloudFlinkWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolFreshCreate); err != nil {
		return err
	}
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
	usesInitialCapacity := flinkListBlockConfigured(d.Get("initial_capacity"))
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

	// initial_capacity is serialized only for the initial PRE purchase.
	if usesInitialCapacity {
		fixedCU, crossZoneFixedCU := expandFlinkInitialCapacity(d.Get("initial_capacity"))
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
	if hasHA && !usesInitialCapacity {
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
	capacityMode := flinkworkspace.CapacityIntentLegacy
	if usesInitialCapacity {
		capacityMode = flinkworkspace.CapacityIntentInitial
	}
	if err := setFlinkWorkspaceCreateProtocolState(d, workspaceRequest, createOptions, capacityMode); err != nil {
		return WrapError(err)
	}
	workspace, err := createFlinkWorkspaceWithIntent(flinkService, workspaceRequest, createOptions, capacityMode, d.Timeout(schema.TimeoutCreate))
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

	return completeFlinkWorkspaceCreate(
		d,
		meta,
		flinkService,
		workspace.Id,
		usesInitialCapacity,
		d.Timeout(schema.TimeoutCreate),
		resourceAliCloudFlinkWorkspaceRead,
	)
}

type flinkWorkspacePostCreateService interface {
	WaitForWorkspaceStarting(string, time.Duration) error
}

type flinkWorkspaceReadFunc func(*schema.ResourceData, interface{}) error

func completeFlinkWorkspaceCreate(
	d *schema.ResourceData,
	_ interface{},
	_ flinkWorkspacePostCreateService,
	workspaceID string,
	_ bool,
	_ time.Duration,
	_ flinkWorkspaceReadFunc,
) error {
	d.SetId(workspaceID)
	// CreateInstance has already returned a paid identity. Returning any
	// readiness or observer error from this Create action would make SDK v1
	// Core taint that identity and replace it on the next apply. Readiness and
	// authoritative state synchronization therefore belong to later refreshes.
	return nil
}

func waitForFlinkWorkspaceReadinessBeforeFirstRead(d *schema.ResourceData, service flinkWorkspacePostCreateService, timeout time.Duration) error {
	tuple := flinkWorkspaceProtocolTupleFromGetter(d)
	if tuple.purchase != flinkWorkspacePurchaseManaged ||
		tuple.capacity != flinkWorkspaceCapacityLegacy ||
		tuple.visibility != flinkWorkspaceIdentityAwaitingFirstRead {
		return nil
	}
	if err := service.WaitForWorkspaceStarting(d.Id(), timeout); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return nil
}

type flinkWorkspaceCreateService interface {
	CreateInstance(*aliyunFlinkAPI.Workspace, flinkworkspace.CreateOptions) (*aliyunFlinkAPI.Workspace, error)
	flinkWorkspaceListService
}

type flinkWorkspaceListService interface {
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

func setFlinkWorkspaceCreateProtocolState(d *schema.ResourceData, request *aliyunFlinkAPI.Workspace, options flinkworkspace.CreateOptions, capacityMode string) error {
	token := flinkworkspace.WorkspaceCreateToken(request)
	fingerprint := flinkworkspace.WorkspaceCreateIntentFingerprint(request, options, capacityMode)
	if token == "" || fingerprint == "" {
		return fmt.Errorf("cannot derive Flink workspace create protocol identity and intent")
	}
	stateCapacityMode := flinkWorkspaceCapacityLegacy
	switch capacityMode {
	case flinkworkspace.CapacityIntentLegacy:
	case flinkworkspace.CapacityIntentInitial:
		stateCapacityMode = flinkWorkspaceCapacityInitial
	default:
		return fmt.Errorf("invalid Flink workspace create capacity intent mode %q", capacityMode)
	}
	for key, value := range map[string]string{
		"identity_visibility_state": flinkWorkspaceIdentityAwaitingFirstRead,
		"purchase_options_state":    flinkWorkspacePurchaseManaged,
		"capacity_intent_mode":      stateCapacityMode,
		"terraform_create_token":    token,
		"create_intent_fingerprint": fingerprint,
	} {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("set Flink workspace create protocol %s: %w", key, err)
		}
	}
	return validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolPreparedCreate)
}

func createFlinkWorkspace(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, options flinkworkspace.CreateOptions, timeout time.Duration) (*aliyunFlinkAPI.Workspace, error) {
	return createFlinkWorkspaceWithIntent(service, request, options, flinkworkspace.CapacityIntentLegacy, timeout)
}

func createFlinkWorkspaceWithIntent(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, options flinkworkspace.CreateOptions, capacityMode string, timeout time.Duration) (*aliyunFlinkAPI.Workspace, error) {
	if request == nil {
		return nil, fmt.Errorf("Flink workspace create request is nil")
	}
	if capacityMode != flinkworkspace.CapacityIntentLegacy && capacityMode != flinkworkspace.CapacityIntentInitial {
		return nil, fmt.Errorf("invalid Flink workspace create capacity intent mode %q", capacityMode)
	}
	token := flinkworkspace.WorkspaceCreateToken(request)
	if token == "" {
		return nil, fmt.Errorf("cannot derive Flink workspace creation token")
	}
	fingerprint := flinkworkspace.WorkspaceCreateIntentFingerprint(request, options, capacityMode)
	if fingerprint == "" {
		return nil, fmt.Errorf("cannot derive Flink workspace creation intent fingerprint")
	}
	request.Tags = appendFlinkWorkspaceProviderTag(request.Tags, flinkworkspace.CreateTokenTagKey, token)
	request.Tags = appendFlinkWorkspaceProviderTag(request.Tags, flinkworkspace.CreateIntentTagKey, fingerprint)

	recovered, err := findFlinkWorkspaceByCreateIntent(service, request, token, fingerprint)
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
		recovered, recoveryErr := waitForFlinkWorkspaceCreateRecovery(service, request, token, fingerprint, timeout)
		if recoveryErr == nil {
			return recovered, nil
		}
		return nil, &pendingFlinkWorkspaceCreateError{token: token, cause: fmt.Errorf("CreateInstance: %w; recovery: %v", createErr, recoveryErr)}
	}

	// A definitive service rejection should not normally have purchased
	// anything, but perform one final authoritative lookup before returning it.
	recovered, recoveryErr := findFlinkWorkspaceByCreateIntent(service, request, token, fingerprint)
	if recoveryErr != nil {
		return nil, fmt.Errorf("CreateInstance failed: %w; recovery by provider creation tag also failed: %v", createErr, recoveryErr)
	}
	if recovered != nil {
		return recovered, nil
	}
	return nil, createErr
}

func appendFlinkWorkspaceProviderTag(tags []aliyunFlinkAPI.Tag, key, value string) []aliyunFlinkAPI.Tag {
	result := make([]aliyunFlinkAPI.Tag, 0, len(tags)+1)
	for _, tag := range tags {
		if tag.Key != key {
			result = append(result, tag)
		}
	}
	return append(result, aliyunFlinkAPI.Tag{Key: key, Value: value})
}

func waitForFlinkWorkspaceCreateRecovery(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, token, fingerprint string, timeout time.Duration) (*aliyunFlinkAPI.Workspace, error) {
	var recovered *aliyunFlinkAPI.Workspace
	err := resource.Retry(timeout, func() *resource.RetryError {
		workspace, err := findFlinkWorkspaceByCreateIntent(service, request, token, fingerprint)
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

func findFlinkWorkspaceByCreateIntent(service flinkWorkspaceCreateService, request *aliyunFlinkAPI.Workspace, token, fingerprint string) (*aliyunFlinkAPI.Workspace, error) {
	workspace, err := findFlinkWorkspaceByCreateToken(service, request, token)
	if err != nil || workspace == nil {
		return workspace, err
	}
	cloudFingerprint, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateIntentTagKey)
	if err != nil {
		return nil, fmt.Errorf("verify recovered Flink workspace create intent: %w", err)
	}
	if cloudFingerprint != fingerprint {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has intent fingerprint %q, expected %q", cloudFingerprint, fingerprint)
	}
	return workspace, nil
}

func findFlinkWorkspaceByCreateToken(service flinkWorkspaceListService, request *aliyunFlinkAPI.Workspace, token string) (*aliyunFlinkAPI.Workspace, error) {
	match, err := findUniqueFlinkWorkspaceByCreateToken(service, token)
	if err != nil || match == nil {
		return match, err
	}
	if match.Name != "" && match.Name != request.Name {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has name %q, expected %q", match.Name, request.Name)
	}
	if match.Region != "" && match.Region != request.Region {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has region %q, expected %q", match.Region, request.Region)
	}
	if err := validateRecoveredFlinkWorkspace(*match, request); err != nil {
		return nil, err
	}
	return match, nil
}

func findUniqueFlinkWorkspaceByCreateToken(service flinkWorkspaceListService, token string) (*aliyunFlinkAPI.Workspace, error) {
	workspaces, err := service.ListInstances()
	if err != nil {
		return nil, &flinkWorkspaceCreateDiscoveryError{cause: err}
	}
	matches := make([]aliyunFlinkAPI.Workspace, 0, 1)
	for _, workspace := range workspaces {
		workspaceToken, found, err := optionalUniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateTokenTagKey)
		if err != nil {
			return nil, fmt.Errorf("verify Flink workspace create-token uniqueness: %w", err)
		}
		if !found || workspaceToken != token {
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
	return &match, nil
}

func findFlinkWorkspaceIdentityByCreateToken(service flinkWorkspaceListService, name, token string) (*aliyunFlinkAPI.Workspace, error) {
	match, err := findUniqueFlinkWorkspaceByCreateToken(service, token)
	if err != nil || match == nil {
		return match, err
	}
	if match.Name != "" && match.Name != name {
		return nil, fmt.Errorf("the workspace carrying the provider creation token has name %q, expected %q", match.Name, name)
	}
	return match, nil
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
	if len(match.HaVSwitchIds) > 0 {
		if !requestHA || !sameFlinkStringSet(match.HaVSwitchIds, request.HighAvailability.VSwitchIds) {
			return fmt.Errorf("the workspace carrying the provider creation token has different HA vSwitch IDs")
		}
	}
	if match.Storage != nil && match.Storage.Oss != nil && request.Storage != nil && request.Storage.Oss != nil && match.Storage.Oss.Bucket != "" && match.Storage.Oss.Bucket != request.Storage.Oss.Bucket {
		return fmt.Errorf("the workspace carrying the provider creation token has OSS bucket %q, expected %q", match.Storage.Oss.Bucket, request.Storage.Oss.Bucket)
	}
	if request.ResourceSpec != nil {
		actualHA := match.Ha || (match.HighAvailability != nil && match.HighAvailability.Enabled)
		if actualHA != requestHA {
			return fmt.Errorf("the workspace carrying the provider creation token has HA=%t, expected %t", actualHA, requestHA)
		}
		if !sameFlinkWorkspaceResourceSpec(match.ResourceSpec, request.ResourceSpec) {
			return fmt.Errorf("the workspace carrying the provider creation token has different primary capacity")
		}
		if requestHA {
			actualHAResource := match.HaResourceSpec
			if actualHAResource == nil && match.HighAvailability != nil {
				actualHAResource = match.HighAvailability.ResourceSpec
			}
			if !sameFlinkWorkspaceResourceSpec(actualHAResource, request.HighAvailability.ResourceSpec) {
				return fmt.Errorf("the workspace carrying the provider creation token has different HA capacity")
			}
		}
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
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolExisting); err != nil {
		return err
	}
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}
	if _, pending := flinkWorkspaceCreateTokenFromPendingID(d.Id()); pending {
		return WrapError(checkPendingFlinkWorkspaceCreate(flinkService, d.Id(), d.Get("name").(string), client.RegionId))
	}
	if err := waitForFlinkWorkspaceReadinessBeforeFirstRead(d, flinkService, d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}
	return readFlinkWorkspaceWithService(d, flinkService)
}

type flinkWorkspaceDescribeService interface {
	DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error)
}

type flinkWorkspaceReadService interface {
	flinkWorkspaceDescribeService
	flinkWorkspaceListService
}

func readFlinkWorkspaceWithService(d *schema.ResourceData, service flinkWorkspaceReadService) error {
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolExisting); err != nil {
		return err
	}
	requestedID := d.Id()
	workspace, err := service.DescribeFlinkWorkspace(requestedID)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			visibility, _ := d.Get("identity_visibility_state").(string)
			if visibility == flinkWorkspaceIdentityAwaitingFirstRead || visibility == flinkWorkspaceIdentityMigratedFirstRead {
				return retainFlinkWorkspaceIdentityAfterNotFound(d, service, err)
			}
			if visibility != flinkWorkspaceIdentityStable {
				return fmt.Errorf("Flink workspace %q has invalid identity_visibility_state %q; retaining state after NotFound", d.Id(), visibility)
			}
			log.Printf("[DEBUG] Resource alicloud_flink_workspace DescribeFlinkWorkspace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	if err := validateFlinkWorkspaceDescribeIdentity(requestedID, workspace); err != nil {
		return err
	}
	return applyFlinkWorkspaceReadState(d, workspace)
}

func validateFlinkWorkspaceDescribeIdentity(requestedID string, workspace *aliyunFlinkAPI.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("DescribeFlinkWorkspace(%q) returned nil", requestedID)
	}
	if workspace.Id == "" {
		return fmt.Errorf("DescribeFlinkWorkspace(%q) returned a workspace with an empty ID; refusing state or cloud mutation", requestedID)
	}
	if workspace.Id != requestedID {
		return fmt.Errorf("DescribeFlinkWorkspace(%q) returned workspace %q; refusing state or cloud mutation for a mismatched identity", requestedID, workspace.Id)
	}
	return nil
}

func retainFlinkWorkspaceIdentityAfterNotFound(d *schema.ResourceData, service flinkWorkspaceListService, describeErr error) error {
	token, _ := d.Get("terraform_create_token").(string)
	if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) {
		return fmt.Errorf("Flink workspace %q is awaiting its first authoritative Read but has no valid create token; paid identity is retained: %w", d.Id(), describeErr)
	}
	workspace, err := findFlinkWorkspaceIdentityByCreateToken(service, d.Get("name").(string), token)
	if err != nil {
		return fmt.Errorf("Flink workspace %q first authoritative Read returned NotFound and token discovery failed; paid identity is retained: %w", d.Id(), err)
	}
	if workspace == nil {
		return fmt.Errorf("Flink workspace %q is not yet visible by ID or create token; paid identity is retained and another purchase is blocked: %w", d.Id(), describeErr)
	}
	if workspace.Id != d.Id() {
		return fmt.Errorf("Flink workspace create token became visible on instance %q while state retains %q; refusing automatic identity transfer", workspace.Id, d.Id())
	}
	return fmt.Errorf("Flink workspace %q is visible by create token but not yet by ID; paid identity is retained until an authoritative ID Read succeeds: %w", d.Id(), describeErr)
}

func applyFlinkWorkspaceReadState(d *schema.ResourceData, workspace *aliyunFlinkAPI.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("cannot refresh Flink workspace state from nil workspace")
	}
	visibility, _ := d.Get("identity_visibility_state").(string)
	if visibility != flinkWorkspaceIdentityAwaitingFirstRead && visibility != flinkWorkspaceIdentityMigratedFirstRead && visibility != flinkWorkspaceIdentityStable {
		return fmt.Errorf("Flink workspace %q has invalid identity_visibility_state %q", d.Id(), visibility)
	}
	purchaseState, _ := d.Get("purchase_options_state").(string)
	migratedInitial := purchaseState == flinkWorkspacePurchaseMigratedInitial
	if purchaseState == flinkWorkspacePurchaseRecoveryPending {
		token := d.Get("terraform_create_token").(string)
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) {
			return fmt.Errorf("verified recovery for Flink workspace %q has an invalid Terraform create token", d.Id())
		}
		cloudToken, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateTokenTagKey)
		if err != nil {
			return err
		}
		if cloudToken != token {
			return fmt.Errorf("verified recovery for Flink workspace %q found create token %q, expected %q", d.Id(), cloudToken, token)
		}
		fingerprint, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateIntentTagKey)
		if err != nil {
			return err
		}
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(fingerprint) {
			return fmt.Errorf("verified recovery for Flink workspace %q requires a valid %s tag", d.Id(), flinkworkspace.CreateIntentTagKey)
		}
		if err := d.Set("create_intent_fingerprint", fingerprint); err != nil {
			return fmt.Errorf("set recovered Flink workspace intent fingerprint: %w", err)
		}
	} else if migratedInitial {
		if visibility != flinkWorkspaceIdentityMigratedFirstRead {
			return fmt.Errorf("schema-v0 initial Flink workspace %q has incompatible identity_visibility_state %q", d.Id(), visibility)
		}
		stateToken, _ := d.Get("terraform_create_token").(string)
		cloudToken, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateTokenTagKey)
		if err != nil || !flinkWorkspaceRecoveryTokenPattern.MatchString(stateToken) || cloudToken != stateToken {
			return fmt.Errorf("schema-v0 initial Flink workspace %q first authoritative Read cannot verify its create token", d.Id())
		}
	} else if visibility == flinkWorkspaceIdentityAwaitingFirstRead {
		if purchaseState != flinkWorkspacePurchaseManaged {
			return fmt.Errorf("Flink workspace %q awaiting its first authoritative Read has incompatible purchase_options_state %q", d.Id(), purchaseState)
		}
		stateToken, _ := d.Get("terraform_create_token").(string)
		cloudToken, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateTokenTagKey)
		if err != nil || !flinkWorkspaceRecoveryTokenPattern.MatchString(stateToken) || cloudToken != stateToken {
			return fmt.Errorf("Flink workspace %q first authoritative Read cannot verify its create token", d.Id())
		}
		stateFingerprint, _ := d.Get("create_intent_fingerprint").(string)
		cloudFingerprint, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateIntentTagKey)
		if err != nil || !flinkWorkspaceRecoveryTokenPattern.MatchString(stateFingerprint) || cloudFingerprint != stateFingerprint {
			return fmt.Errorf("Flink workspace %q first authoritative Read cannot verify its create intent fingerprint", d.Id())
		}
	}

	// Set attributes from workspace
	if err := d.Set("name", workspace.Name); err != nil {
		return err
	}
	if err := d.Set("resource_group_id", workspace.ResourceGroupId); err != nil {
		return err
	}

	configuredPrimaryZoneID, _ := d.Get("zone_id").(string)
	if zoneID := flinkworkspace.PrimaryZoneID(workspace, configuredPrimaryZoneID); zoneID != "" {
		if err := d.Set("zone_id", zoneID); err != nil {
			return err
		}
	}

	if err := d.Set("vpc_id", workspace.VpcId); err != nil {
		return err
	}

	// Handle vswitch_ids with proper type conversion
	if workspace.VSwitchIds != nil && len(workspace.VSwitchIds) > 0 {
		if err := d.Set("vswitch_ids", workspace.VSwitchIds); err != nil {
			return err
		}
	}

	// Handle security group from SecurityGroupInfo structure
	if workspace.SecurityGroupInfo != nil && workspace.SecurityGroupInfo.SecurityGroupId != "" {
		if err := d.Set("security_group_id", workspace.SecurityGroupInfo.SecurityGroupId); err != nil {
			return err
		}
	}

	// Set fields that are returned by the API with fallback to configured values
	if workspace.ArchitectureType != "" {
		if err := d.Set("architecture_type", workspace.ArchitectureType); err != nil {
			return err
		}
	}
	if workspace.ChargeType != "" {
		if err := d.Set("charge_type", workspace.ChargeType); err != nil {
			return err
		}
	}
	if workspace.MonitorType != "" {
		if err := d.Set("monitor_type", workspace.MonitorType); err != nil {
			return err
		}
	}

	// Always set resource_id if available
	if workspace.ResourceId != "" {
		if err := d.Set("resource_id", workspace.ResourceId); err != nil {
			return err
		}
	}

	capacityMode := d.Get("capacity_intent_mode").(string)
	if capacityMode == "" {
		return fmt.Errorf("Flink workspace %q has no capacity_intent_mode; refresh with the schema-v1 state upgrader or explicitly re-import it", d.Id())
	}
	if err := d.Set("observed_capacity", flattenFlinkWorkspaceObservedCapacity(workspace)); err != nil {
		return err
	}

	// initial_capacity is input-only. Only the persisted LEGACY marker permits
	// API-observable runtime capacity to materialize as deprecated configuration.
	if capacityMode == flinkWorkspaceCapacityLegacy && workspace.ResourceSpec != nil {
		resourceConfig := map[string]interface{}{
			"cpu":    int(workspace.ResourceSpec.Cpu),
			"memory": int(workspace.ResourceSpec.MemoryGB),
		}
		if err := d.Set("resource", []interface{}{resourceConfig}); err != nil {
			return err
		}
	} else if capacityMode == flinkWorkspaceCapacityInitial || capacityMode == flinkWorkspaceCapacityInitialAdoptionPending {
		if err := d.Set("resource", nil); err != nil {
			return err
		}
	} else if capacityMode != flinkWorkspaceCapacityLegacy {
		return fmt.Errorf("Flink workspace %q has invalid capacity_intent_mode %q", d.Id(), capacityMode)
	}

	// Set storage configuration
	if workspace.Storage != nil && workspace.Storage.Oss != nil {
		storageConfig := map[string]interface{}{
			"oss_bucket": workspace.Storage.Oss.Bucket,
		}
		if err := d.Set("storage", []interface{}{storageConfig}); err != nil {
			return err
		}
	}

	// Set HA configuration. DescribeInstances returns the flat Ha* fields,
	// while the create path uses HighAvailability.
	configuredStandbyZoneID := ""
	if configuredHA, ok := flinkFirstBlock(d.Get("ha")); ok {
		configuredStandbyZoneID, _ = configuredHA["zone_id"].(string)
	}
	if haConfig, ok := flinkworkspace.HAConfigWithFallback(workspace, configuredStandbyZoneID); ok {
		if capacityMode != flinkWorkspaceCapacityLegacy {
			delete(haConfig, "resource")
		}
		if haResource, ok := haConfig["resource"]; ok {
			haConfig["resource"] = []interface{}{haResource}
		}
		if err := d.Set("ha", []interface{}{haConfig}); err != nil {
			return err
		}
	} else {
		if err := d.Set("ha", []interface{}{}); err != nil {
			return err
		}
	}
	if migratedInitial {
		oldTuple := flinkWorkspaceProtocolTupleFromGetter(d)
		for key, value := range map[string]string{
			"purchase_options_state":    flinkWorkspacePurchaseManaged,
			"identity_visibility_state": flinkWorkspaceIdentityStable,
			"terraform_create_token":    flinkWorkspaceProtocolUnavailable,
		} {
			if err := d.Set(key, value); err != nil {
				return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, fmt.Errorf("stabilize migrated schema-v0 Flink workspace identity: %w", err))
			}
		}
	} else if visibility == flinkWorkspaceIdentityAwaitingFirstRead {
		if err := d.Set("identity_visibility_state", flinkWorkspaceIdentityStable); err != nil {
			return fmt.Errorf("stabilize Flink workspace identity visibility: %w", err)
		}
	}
	return validateFlinkWorkspaceProtocol(d.Id(), d, flinkWorkspaceProtocolExisting)
}

func uniqueFlinkWorkspaceTagValue(tags []aliyunFlinkAPI.Tag, key string) (string, error) {
	value, found, err := optionalUniqueFlinkWorkspaceTagValue(tags, key)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("Flink workspace is missing required %s tag", key)
	}
	return value, nil
}

func optionalUniqueFlinkWorkspaceTagValue(tags []aliyunFlinkAPI.Tag, key string) (string, bool, error) {
	value := ""
	found := false
	for _, tag := range tags {
		if tag.Key != key {
			continue
		}
		if found {
			return "", false, fmt.Errorf("Flink workspace has duplicate %s tags", key)
		}
		found = true
		value = tag.Value
	}
	return value, found, nil
}

func resourceAliCloudFlinkWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	oldTuple, err := prepareFlinkWorkspaceProtocolUpdate(d)
	if err != nil {
		return err
	}
	client, ok := meta.(*connectivity.AliyunClient)
	if !ok || client == nil {
		return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, fmt.Errorf("Flink workspace Update has no configured Aliyun client"))
	}
	service, err := NewFlinkService(client)
	if err != nil {
		return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, WrapError(err))
	}
	return updateFlinkWorkspaceWithPreparedProtocol(d, service, oldTuple)
}

func updateFlinkWorkspaceWithService(d *schema.ResourceData, service flinkWorkspaceDescribeService) error {
	oldTuple, err := prepareFlinkWorkspaceProtocolUpdate(d)
	if err != nil {
		return err
	}
	return updateFlinkWorkspaceWithPreparedProtocol(d, service, oldTuple)
}

func prepareFlinkWorkspaceProtocolUpdate(d *schema.ResourceData) (flinkWorkspaceProtocolTuple, error) {
	newTuple := flinkWorkspaceProtocolTupleFromGetter(d)
	oldTuple := flinkWorkspaceProtocolTupleFromResourceDataChange(d, true)
	if oldTuple == (flinkWorkspaceProtocolTuple{}) {
		oldTuple = newTuple
	}
	if err := validateFlinkWorkspaceProtocolTuple(d.Id(), oldTuple, flinkWorkspaceProtocolExisting); err != nil {
		return oldTuple, err
	}
	if err := validateFlinkWorkspaceProtocolTuple(d.Id(), newTuple, flinkWorkspaceProtocolExisting); err != nil {
		return oldTuple, rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, err)
	}
	if err := validateFlinkWorkspaceProtocolTransition(d.Id(), oldTuple, newTuple); err != nil {
		return oldTuple, rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, err)
	}
	return oldTuple, nil
}

func updateFlinkWorkspaceWithPreparedProtocol(d *schema.ResourceData, service flinkWorkspaceDescribeService, oldTuple flinkWorkspaceProtocolTuple) error {
	recoveryPending := oldTuple.purchase == flinkWorkspacePurchaseRecoveryPending
	initialPending := oldTuple.capacity == flinkWorkspaceCapacityInitialAdoptionPending
	if !recoveryPending && !initialPending {
		readService, ok := service.(flinkWorkspaceReadService)
		if !ok {
			return fmt.Errorf("normal Flink workspace Update requires read-only Describe/List service")
		}
		return readFlinkWorkspaceWithService(d, readService)
	}

	requestedID := d.Id()
	workspace, err := service.DescribeFlinkWorkspace(requestedID)
	if err != nil {
		return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, WrapError(err))
	}
	if err := validateFlinkWorkspaceDescribeIdentity(requestedID, workspace); err != nil {
		return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, fmt.Errorf("validate Flink workspace adoption identity: %w", err))
	}
	intentMode := flinkworkspace.CapacityIntentLegacy
	if initialPending {
		intentMode = flinkworkspace.CapacityIntentInitial
	}
	if err := validateFlinkWorkspaceAdoptionWorkspace(d, workspace, intentMode, recoveryPending); err != nil {
		return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, err)
	}
	if recoveryPending {
		if err := d.Set("purchase_options_state", flinkWorkspacePurchaseManaged); err != nil {
			return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, err)
		}
	}
	if initialPending {
		if err := d.Set("capacity_intent_mode", flinkWorkspaceCapacityInitial); err != nil {
			return rollbackFlinkWorkspaceProtocolTuple(d, oldTuple, err)
		}
	}
	return applyFlinkWorkspaceReadState(d, workspace)
}

func rollbackFlinkWorkspaceProtocolTuple(d *schema.ResourceData, tuple flinkWorkspaceProtocolTuple, cause error) error {
	var stateErr error
	for key, value := range map[string]string{
		"purchase_options_state":    tuple.purchase,
		"capacity_intent_mode":      tuple.capacity,
		"identity_visibility_state": tuple.visibility,
		"terraform_create_token":    tuple.token,
		"create_intent_fingerprint": tuple.fingerprint,
	} {
		if err := d.Set(key, value); err != nil && stateErr == nil {
			stateErr = err
		}
	}
	if stateErr != nil {
		return fmt.Errorf("%v; also failed to retain pending Flink workspace adoption state: %w", cause, stateErr)
	}
	return cause
}

func validateFlinkWorkspaceAdoptionWorkspace(d *schema.ResourceData, workspace *aliyunFlinkAPI.Workspace, intentMode string, recoveryPending bool) error {
	if workspace.ChargeType != "PRE" {
		return fmt.Errorf("Flink workspace adoption requires PRE billing, cloud returned %q", workspace.ChargeType)
	}
	if err := validateFlinkWorkspaceAdoptionObservableState(d, workspace); err != nil {
		return err
	}
	_, expectedHA := flinkFirstBlock(d.Get("ha"))
	actualHA := workspace.Ha || (workspace.HighAvailability != nil && workspace.HighAvailability.Enabled)
	if actualHA != expectedHA {
		return fmt.Errorf("Flink workspace adoption HA topology mismatch: cloud HA=%t configuration HA=%t", actualHA, expectedHA)
	}
	request, _, err := flinkWorkspaceCreateIntentFromGetter(d, intentMode)
	if err != nil {
		return err
	}
	if !sameFlinkWorkspaceResourceSpec(workspace.ResourceSpec, request.ResourceSpec) {
		return fmt.Errorf("Flink workspace adoption primary capacity mismatch")
	}
	if expectedHA {
		actualHAResource := workspace.HaResourceSpec
		if actualHAResource == nil && workspace.HighAvailability != nil {
			actualHAResource = workspace.HighAvailability.ResourceSpec
		}
		if request.HighAvailability == nil || !sameFlinkWorkspaceResourceSpec(actualHAResource, request.HighAvailability.ResourceSpec) {
			return fmt.Errorf("Flink workspace adoption HA capacity mismatch")
		}
	}
	if recoveryPending {
		token := d.Get("terraform_create_token").(string)
		cloudToken, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateTokenTagKey)
		if err != nil {
			return err
		}
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(token) || cloudToken != token {
			return fmt.Errorf("verified Flink workspace recovery create token mismatch")
		}
		cloudFingerprint, err := uniqueFlinkWorkspaceTagValue(workspace.Tags, flinkworkspace.CreateIntentTagKey)
		if err != nil {
			return err
		}
		plannedFingerprint := d.Get("create_intent_fingerprint").(string)
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(plannedFingerprint) || cloudFingerprint != plannedFingerprint {
			return fmt.Errorf("verified Flink workspace recovery intent fingerprint changed between plan and apply: cloud=%s planned=%s", cloudFingerprint, plannedFingerprint)
		}
	}
	return nil
}

func validateFlinkWorkspaceAdoptionObservableState(d *schema.ResourceData, workspace *aliyunFlinkAPI.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("Flink workspace adoption cannot verify observable state from a nil workspace")
	}
	securityGroupID := ""
	if workspace.SecurityGroupInfo != nil {
		securityGroupID = workspace.SecurityGroupInfo.SecurityGroupId
	}
	storageBucket := ""
	if workspace.Storage != nil && workspace.Storage.Oss != nil {
		storageBucket = workspace.Storage.Oss.Bucket
	}
	expectedStorage, _ := flinkFirstBlock(d.Get("storage"))
	expectedStorageBucket, _ := expectedStorage["oss_bucket"].(string)
	for _, field := range []struct {
		name     string
		actual   string
		expected string
	}{
		{name: "name", actual: workspace.Name, expected: d.Get("name").(string)},
		{name: "resource group", actual: workspace.ResourceGroupId, expected: d.Get("resource_group_id").(string)},
		{name: "primary zone", actual: flinkworkspace.PrimaryZoneID(workspace, ""), expected: d.Get("zone_id").(string)},
		{name: "VPC", actual: workspace.VpcId, expected: d.Get("vpc_id").(string)},
		{name: "security group", actual: securityGroupID, expected: d.Get("security_group_id").(string)},
		{name: "architecture", actual: workspace.ArchitectureType, expected: d.Get("architecture_type").(string)},
		{name: "monitor type", actual: workspace.MonitorType, expected: d.Get("monitor_type").(string)},
		{name: "storage bucket", actual: storageBucket, expected: expectedStorageBucket},
	} {
		if field.actual != field.expected {
			return fmt.Errorf("Flink workspace adoption observable %s mismatch", field.name)
		}
	}
	if !sameFlinkStringSet(workspace.VSwitchIds, flinkStringList(d.Get("vswitch_ids"))) {
		return fmt.Errorf("Flink workspace adoption observable primary vSwitch mismatch")
	}
	expectedHA, hasExpectedHA := flinkFirstBlock(d.Get("ha"))
	actualHA, hasActualHA := flinkworkspace.HAConfigWithFallback(workspace, "")
	if hasActualHA != hasExpectedHA {
		return fmt.Errorf("Flink workspace adoption observable HA topology mismatch")
	}
	if hasExpectedHA {
		actualHAVSwitches := flinkStringValues(actualHA["vswitch_ids"])
		expectedHAVSwitches := flinkStringValues(expectedHA["vswitch_ids"])
		if !sameFlinkStringSet(actualHAVSwitches, expectedHAVSwitches) {
			return fmt.Errorf("Flink workspace adoption observable HA vSwitch mismatch")
		}
		actualHAZone, _ := actualHA["zone_id"].(string)
		expectedHAZone, _ := expectedHA["zone_id"].(string)
		if actualHAZone != expectedHAZone {
			return fmt.Errorf("Flink workspace adoption observable HA zone mismatch")
		}
	}
	return nil
}

func flinkStringValues(value interface{}) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []interface{}:
		return flinkStringList(items)
	default:
		return nil
	}
}

func sameFlinkWorkspaceResourceSpec(actual, expected *aliyunFlinkAPI.ResourceSpec) bool {
	if actual == nil || expected == nil {
		return actual == nil && expected == nil
	}
	return actual.Cpu == expected.Cpu && actual.MemoryGB == expected.MemoryGB
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
	if err := validateFlinkWorkspaceDescribeIdentity(instanceID, workspace); err != nil {
		return fmt.Errorf("read Flink workspace %q before deletion: %w", instanceID, err)
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
