package alicloud

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// flinkMaxCUBeforeInt32MemoryOverflow keeps MemoryGB=CU*4 representable by
// the FOAS int32 request fields.
const flinkMaxCUBeforeInt32MemoryOverflow = 2147483647 / 4

var flinkWorkspacePurchaseInputFields = []string{
	"auto_renew", "duration", "pricing_cycle", "extra", "promotion_code", "use_promotion_code",
}

func validateFlinkCUFitsInt32Memory(value interface{}, key string) ([]string, []error) {
	cu, ok := value.(int)
	if !ok {
		return nil, []error{fmt.Errorf("%s must be an integer CU", key)}
	}
	if cu < 0 || cu > flinkMaxCUBeforeInt32MemoryOverflow {
		return nil, []error{fmt.Errorf("%s must be between 0 and %d so MemoryGB=CU*4 fits int32", key, flinkMaxCUBeforeInt32MemoryOverflow)}
	}
	return nil, nil
}

func flinkInitialCapacitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"fixed_cu": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateFlinkCUFitsInt32Memory,
			},
			"cross_zone_fixed_cu": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateFlinkCUFitsInt32Memory,
			},
		}},
		Description: "Create-only fixed capacity used to purchase a PRE workspace.",
	}
}

func flinkObservedCapacitySchema(includeHA bool) *schema.Schema {
	fields := map[string]*schema.Schema{
		"fixed_cu": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"elastic_cu_limit": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"max_cu_limit": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"used_cu": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
	}
	if includeHA {
		fields["ha"] = &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"cross_zone_fixed_cu": {
					Type:     schema.TypeFloat,
					Computed: true,
				},
			}},
		}
	}
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        &schema.Resource{Schema: fields},
		Description: "The capacity observed from Flink.",
	}
}

func flinkWorkspaceCustomizeDiff(d *schema.ResourceDiff, _ interface{}) error {
	protocolContext := flinkWorkspaceProtocolExisting
	if d.Id() == "" {
		protocolContext = flinkWorkspaceProtocolFreshCreate
	}
	if err := validateFlinkWorkspaceProtocol(d.Id(), d, protocolContext); err != nil {
		return err
	}
	chargeType := d.Get("charge_type").(string)
	purchaseState, _ := d.Get("purchase_options_state").(string)
	capacityMode, _ := d.Get("capacity_intent_mode").(string)
	hasResource := flinkListBlockConfigured(d.Get("resource"))
	initialCapacity := d.Get("initial_capacity")
	hasInitialCapacity := flinkListBlockConfigured(initialCapacity)
	ha, hasHA := flinkFirstBlock(d.Get("ha"))
	hasHAResource := hasHA && flinkListBlockConfigured(ha["resource"])

	if purchaseState == flinkWorkspacePurchaseMigratedInitial {
		return fmt.Errorf("schema-v0 initial Flink workspace %q requires a successful authoritative Read before any plan can proceed; paid identity is retained", d.Id())
	}
	if purchaseState == flinkWorkspacePurchaseLegacyUnclassified {
		for _, field := range flinkWorkspacePurchaseInputFields {
			if d.HasChange(field) {
				return fmt.Errorf("cannot change %s because this schema-v0 Flink workspace state has unclassified purchase provenance; back up state and explicitly re-import the workspace", field)
			}
		}
	}
	if purchaseState == flinkWorkspacePurchaseManaged && d.Id() != "" {
		for _, field := range flinkWorkspacePurchaseInputFields {
			if d.HasChange(field) {
				if err := d.ForceNew(field); err != nil {
					return fmt.Errorf("preserve managed Flink workspace %s replacement semantics: %w", field, err)
				}
			}
		}
	}
	if purchaseState == flinkWorkspacePurchaseImportedUnknown && capacityMode == flinkWorkspaceCapacityLegacy && hasInitialCapacity {
		return fmt.Errorf("a legacy imported Flink workspace cannot adopt initial_capacity implicitly; re-import it with the explicit |initial mode")
	}
	if d.Id() != "" && d.HasChange("charge_type") {
		oldValue, newValue := d.GetChange("charge_type")
		return fmt.Errorf("charge_type cannot be changed in place from %q to %q; create a separate Flink workspace for billing migrations", oldValue, newValue)
	}
	if d.Id() != "" && d.HasChange("initial_capacity") && capacityMode != flinkWorkspaceCapacityInitialAdoptionPending {
		return fmt.Errorf("initial_capacity cannot be changed after the Flink workspace is created")
	}
	if chargeType == "POST" && hasHA {
		return fmt.Errorf("POST Flink workspaces do not support high availability; HA requires PRE billing")
	}

	if hasInitialCapacity {
		if chargeType != "PRE" {
			return fmt.Errorf("initial_capacity requires PRE billing")
		}
		if hasResource {
			return fmt.Errorf("resource is mutually exclusive with initial_capacity")
		}
		if hasHAResource {
			return fmt.Errorf("ha.resource is mutually exclusive with initial_capacity")
		}
		if err := validateInitialCapacity(initialCapacity, hasHA); err != nil {
			return err
		}
	} else {
		if !hasResource {
			return fmt.Errorf("resource is required when initial_capacity is not configured")
		}
		if hasHA && !hasHAResource {
			return fmt.Errorf("ha.resource is required when initial_capacity is not configured")
		}
	}
	if err := planFlinkWorkspaceAdoption(d); err != nil {
		return err
	}
	return validateFlinkWorkspaceProtocol(d.Id(), d, protocolContext)
}

func planFlinkWorkspaceAdoption(d *schema.ResourceDiff) error {
	purchaseState, _ := d.Get("purchase_options_state").(string)
	capacityMode, _ := d.Get("capacity_intent_mode").(string)
	recoveryPending := purchaseState == flinkWorkspacePurchaseRecoveryPending
	initialPending := capacityMode == flinkWorkspaceCapacityInitialAdoptionPending
	if !recoveryPending && !initialPending {
		return nil
	}
	if err := rejectFlinkWorkspaceAdoptionNonProtocolDiff(d); err != nil {
		return err
	}
	intentMode := flinkworkspace.CapacityIntentLegacy
	if initialPending {
		intentMode = flinkworkspace.CapacityIntentInitial
	}
	if err := validateFlinkWorkspaceAdoptionPlanCapacity(d, intentMode); err != nil {
		return err
	}
	if recoveryPending {
		cloudFingerprint, _ := d.Get("create_intent_fingerprint").(string)
		if !flinkWorkspaceRecoveryTokenPattern.MatchString(cloudFingerprint) {
			return fmt.Errorf("verified Flink workspace recovery has no valid cloud intent fingerprint in state; run import Read again")
		}
		request, options, err := flinkWorkspaceCreateIntentFromGetter(d, intentMode)
		if err != nil {
			return err
		}
		configuredFingerprint := flinkworkspace.WorkspaceCreateIntentFingerprint(request, options, intentMode)
		if configuredFingerprint != cloudFingerprint {
			return fmt.Errorf("verified Flink workspace recovery intent fingerprint mismatch: configuration does not match the paid CreateInstance intent")
		}
		if err := d.SetNew("purchase_options_state", flinkWorkspacePurchaseManaged); err != nil {
			return fmt.Errorf("plan Flink workspace purchase provenance adoption: %w", err)
		}
	}
	if initialPending {
		if err := d.SetNew("capacity_intent_mode", flinkWorkspaceCapacityInitial); err != nil {
			return fmt.Errorf("plan Flink workspace initial capacity adoption: %w", err)
		}
	}
	return nil
}

var flinkWorkspaceAdoptionAllowedDiffRoots = map[string]struct{}{
	"auto_renew":                {},
	"capacity_intent_mode":      {},
	"create_intent_fingerprint": {},
	"duration":                  {},
	"extra":                     {},
	"identity_visibility_state": {},
	"initial_capacity":          {},
	"observed_capacity":         {},
	"pricing_cycle":             {},
	"promotion_code":            {},
	"purchase_options_state":    {},
	"resource_id":               {},
	"terraform_create_token":    {},
	"use_promotion_code":        {},
}

func rejectFlinkWorkspaceAdoptionNonProtocolDiff(d *schema.ResourceDiff) error {
	keys := d.GetChangedKeysPrefix("")
	sort.Strings(keys)
	for _, key := range keys {
		root := strings.SplitN(key, ".", 2)[0]
		if _, allowed := flinkWorkspaceAdoptionAllowedDiffRoots[root]; allowed {
			continue
		}
		if !d.NewValueKnown(root) {
			return fmt.Errorf("Flink workspace state-only adoption cannot verify unknown %s configuration; restore the refreshed observable configuration and plan again", root)
		}
		if root == "ha" {
			return fmt.Errorf("Flink workspace state-only adoption refuses HA observable configuration diff %s; restore the refreshed observable configuration and plan again", key)
		}
		return fmt.Errorf("Flink workspace state-only adoption refuses non-protocol diff %s; restore the refreshed observable configuration and plan again", key)
	}
	return nil
}

func validateFlinkWorkspaceAdoptionPlanCapacity(d *schema.ResourceDiff, intentMode string) error {
	if chargeType, _ := d.Get("charge_type").(string); chargeType != "PRE" {
		return fmt.Errorf("Flink workspace adoption requires PRE billing, got %q", chargeType)
	}
	_, expectedHA := flinkFirstBlock(d.Get("ha"))
	oldHA, _ := d.GetChange("ha")
	_, actualHA := flinkFirstBlock(oldHA)
	if actualHA != expectedHA {
		return fmt.Errorf("Flink workspace adoption HA topology mismatch: cloud HA=%t configuration HA=%t", actualHA, expectedHA)
	}
	if intentMode == flinkworkspace.CapacityIntentLegacy {
		oldPrimary, _ := d.GetChange("resource")
		if !sameFlinkWorkspaceLegacyResource(oldPrimary, d.Get("resource")) {
			return fmt.Errorf("Flink workspace adoption primary legacy capacity mismatch")
		}
		if expectedHA {
			oldHAMap, _ := flinkFirstBlock(oldHA)
			newHAMap, _ := flinkFirstBlock(d.Get("ha"))
			if !sameFlinkWorkspaceLegacyResource(oldHAMap["resource"], newHAMap["resource"]) {
				return fmt.Errorf("Flink workspace adoption HA legacy capacity mismatch")
			}
		}
		return nil
	}
	observed, ok := flinkFirstBlock(d.Get("observed_capacity"))
	if !ok {
		return fmt.Errorf("Flink workspace adoption cannot verify observed capacity")
	}
	fixed, fixedOK := numberAsFloat(observed["fixed_cu"])
	cross := 0.0
	if observedHA, ok := flinkFirstBlock(observed["ha"]); ok {
		cross, _ = numberAsFloat(observedHA["cross_zone_fixed_cu"])
	}
	if !fixedOK {
		return fmt.Errorf("Flink workspace adoption cannot verify observed fixed capacity")
	}
	wantFixed, wantCross := expandFlinkInitialCapacity(d.Get("initial_capacity"))
	if fixed != float64(wantFixed) || cross != float64(wantCross) {
		return fmt.Errorf("Flink workspace adoption capacity mismatch: cloud fixed/cross-zone CU %.6g/%.6g configuration %d/%d", fixed, cross, wantFixed, wantCross)
	}
	return nil
}

func sameFlinkWorkspaceLegacyResource(actual, expected interface{}) bool {
	actualMap, actualOK := flinkFirstBlock(actual)
	expectedMap, expectedOK := flinkFirstBlock(expected)
	if actualOK != expectedOK || !actualOK {
		return actualOK == expectedOK
	}
	return actualMap["cpu"] == expectedMap["cpu"] && actualMap["memory"] == expectedMap["memory"]
}

type flinkWorkspaceValueGetter interface {
	Get(string) interface{}
}

func flinkWorkspaceCreateIntentFromGetter(getter flinkWorkspaceValueGetter, intentMode string) (*flink.Workspace, flinkworkspace.CreateOptions, error) {
	workspace := &flink.Workspace{ChargeType: getter.Get("charge_type").(string)}
	ha, hasHA := flinkFirstBlock(getter.Get("ha"))
	if intentMode == flinkworkspace.CapacityIntentInitial {
		fixed, cross := expandFlinkInitialCapacity(getter.Get("initial_capacity"))
		workspace.ResourceSpec = &flink.ResourceSpec{Cpu: float64(fixed), MemoryGB: float64(fixed * 4)}
		if hasHA {
			workspace.HighAvailability = &flink.HighAvailability{
				Enabled:      true,
				ResourceSpec: &flink.ResourceSpec{Cpu: float64(cross), MemoryGB: float64(cross * 4)},
			}
		}
	} else if intentMode == flinkworkspace.CapacityIntentLegacy {
		resource, ok := flinkFirstBlock(getter.Get("resource"))
		if !ok {
			return nil, flinkworkspace.CreateOptions{}, fmt.Errorf("legacy Flink workspace recovery requires resource")
		}
		workspace.ResourceSpec = &flink.ResourceSpec{Cpu: float64(resource["cpu"].(int)), MemoryGB: float64(resource["memory"].(int))}
		if hasHA {
			haResource, ok := flinkFirstBlock(ha["resource"])
			if !ok {
				return nil, flinkworkspace.CreateOptions{}, fmt.Errorf("legacy HA Flink workspace recovery requires ha.resource")
			}
			workspace.HighAvailability = &flink.HighAvailability{
				Enabled: true,
				ResourceSpec: &flink.ResourceSpec{
					Cpu: float64(haResource["cpu"].(int)), MemoryGB: float64(haResource["memory"].(int)),
				},
			}
		}
	} else {
		return nil, flinkworkspace.CreateOptions{}, fmt.Errorf("invalid Flink workspace intent mode %q", intentMode)
	}
	autoRenew := getter.Get("auto_renew").(bool)
	duration := int32(getter.Get("duration").(int))
	usePromotionCode := getter.Get("use_promotion_code").(bool)
	return workspace, flinkworkspace.CreateOptions{
		AutoRenew:        &autoRenew,
		Duration:         &duration,
		PricingCycle:     getter.Get("pricing_cycle").(string),
		Extra:            getter.Get("extra").(string),
		PromotionCode:    getter.Get("promotion_code").(string),
		UsePromotionCode: &usePromotionCode,
	}, nil
}

func validateInitialCapacity(value interface{}, workspaceHA bool) error {
	capacity, ok := flinkFirstBlock(value)
	if !ok {
		return fmt.Errorf("initial_capacity is required")
	}
	fixed, _ := capacity["fixed_cu"].(int)
	crossZone, _ := capacity["cross_zone_fixed_cu"].(int)
	if _, errors := validateFlinkCUFitsInt32Memory(fixed, "initial_capacity.fixed_cu"); len(errors) > 0 {
		return errors[0]
	}
	if _, errors := validateFlinkCUFitsInt32Memory(crossZone, "initial_capacity.cross_zone_fixed_cu"); len(errors) > 0 {
		return errors[0]
	}
	if workspaceHA {
		if crossZone <= 0 {
			return fmt.Errorf("initial_capacity.cross_zone_fixed_cu must be greater than zero for an HA workspace")
		}
		if fixed != 0 {
			return fmt.Errorf("initial_capacity.fixed_cu must be zero for an HA workspace; HA initial_capacity must use a pure cross-zone pool")
		}
	} else if crossZone != 0 {
		return fmt.Errorf("initial_capacity.cross_zone_fixed_cu must be zero for a non-HA workspace")
	}
	if fixed+crossZone <= 0 {
		return fmt.Errorf("initial_capacity fixed CU total must be greater than zero")
	}
	return nil
}

func flinkListBlockConfigured(value interface{}) bool {
	_, ok := flinkFirstBlock(value)
	return ok
}

func flinkFirstBlock(value interface{}) (map[string]interface{}, bool) {
	items, ok := value.([]interface{})
	if !ok || len(items) == 0 || items[0] == nil {
		return nil, false
	}
	block, ok := items[0].(map[string]interface{})
	return block, ok
}

func expandFlinkInitialCapacity(value interface{}) (fixedCU, crossZoneFixedCU int) {
	capacity, ok := flinkFirstBlock(value)
	if !ok {
		return 0, 0
	}
	fixedCU, _ = capacity["fixed_cu"].(int)
	crossZoneFixedCU, _ = capacity["cross_zone_fixed_cu"].(int)
	return fixedCU, crossZoneFixedCU
}

func flattenFlinkWorkspaceObservedCapacity(workspace *flink.Workspace) []interface{} {
	if workspace == nil {
		return []interface{}{}
	}
	fixedCU := flinkSpecCPU(workspace.ResourceSpec)
	crossZoneFixedCU := flinkSpecCPU(workspace.HaResourceSpec)
	elasticCU := flinkSpecCPU(workspace.ElasticResourceSpec)
	if workspace.ChargeType == "POST" {
		elasticCU = fixedCU
		fixedCU = 0
		crossZoneFixedCU = 0
	}
	usedCU := 0.0
	if workspace.ClusterUsedResources != nil {
		usedCU = workspace.ClusterUsedResources.UsedResource
	}
	return flattenFlinkObservedCapacity(fixedCU, crossZoneFixedCU, fixedCU+crossZoneFixedCU+elasticCU, usedCU, true)
}

func flattenFlinkNamespaceObservedCapacity(namespace *flink.Namespace) []interface{} {
	if namespace == nil {
		return []interface{}{}
	}
	fixedCU := flinkSpecCPU(namespace.GuaranteedResourceSpec)
	elasticCU := flinkSpecCPU(namespace.ElasticResourceSpec)
	if namespace.GuaranteedResourceSpec == nil && namespace.ElasticResourceSpec == nil {
		fixedCU = flinkSpecCPU(namespace.ResourceSpec)
	}
	usedCU := 0.0
	if namespace.ResourceUsed != nil {
		usedCU = namespace.ResourceUsed.Cu
	}
	return flattenFlinkObservedCapacity(fixedCU, 0, fixedCU+elasticCU, usedCU, false)
}

func flattenFlinkQueueObservedCapacity(quota *flink.ResourceQuota) []interface{} {
	if quota == nil {
		return []interface{}{}
	}
	fixedCU := flinkSpecCPU(quota.Request)
	limitCU := flinkSpecCPU(quota.Limit)
	usedCU := flinkSpecCPU(quota.Used)
	return flattenFlinkObservedCapacity(fixedCU, 0, limitCU, usedCU, false)
}

func flattenFlinkObservedCapacity(fixedCU, crossZoneFixedCU, limitCU, usedCU float64, includeHA bool) []interface{} {
	value := map[string]interface{}{
		"fixed_cu":         fixedCU,
		"elastic_cu_limit": limitCU - fixedCU - crossZoneFixedCU,
		"max_cu_limit":     limitCU,
		"used_cu":          usedCU,
	}
	if includeHA {
		value["ha"] = []interface{}{map[string]interface{}{"cross_zone_fixed_cu": crossZoneFixedCU}}
	}
	return []interface{}{value}
}

func flinkSpecCPU(spec *flink.ResourceSpec) float64 {
	if spec == nil {
		return 0
	}
	return spec.Cpu
}
