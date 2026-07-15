package alicloud

import (
	"fmt"

	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	CapacityManagedByResource    = "RESOURCE"
	CapacityManagedByCoordinator = "COORDINATOR"
)

func flinkCapacityManagementSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Default:      CapacityManagedByResource,
		ValidateFunc: validation.StringInSlice([]string{CapacityManagedByResource, CapacityManagedByCoordinator}, false),
		Description:  "Selects whether this lifecycle resource or alicloud_flink_capacity_coordinator owns capacity updates.",
	}
}

func flinkBootstrapCapacitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"fixed_cu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"ha": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"cross_zone_fixed_cu": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(1),
					},
				}},
			},
		}},
		Description: "Create-only fixed capacity used to purchase a PRE workspace before the coordinator takes ownership.",
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
		Description: "The capacity observed from Flink. It is informational when capacity_management is COORDINATOR.",
	}
}

func flinkWorkspaceCustomizeDiff(d *schema.ResourceDiff, _ interface{}) error {
	mode := d.Get("capacity_management").(string)
	chargeType := d.Get("charge_type").(string)
	hasResource := flinkListBlockConfigured(d.Get("resource"))
	hasBootstrap := flinkListBlockConfigured(d.Get("bootstrap_capacity"))
	ha, hasHA := flinkFirstBlock(d.Get("ha"))
	hasHAResource := hasHA && flinkListBlockConfigured(ha["resource"])

	if d.Id() != "" && d.HasChange("charge_type") {
		oldValue, newValue := d.GetChange("charge_type")
		return fmt.Errorf("charge_type cannot be changed in place from %q to %q; create a separate Flink workspace for billing migrations", oldValue, newValue)
	}
	if chargeType == "POST" && hasHA {
		return fmt.Errorf("POST Flink workspaces do not support high availability; HA requires PRE billing")
	}

	switch mode {
	case CapacityManagedByResource:
		if !hasResource {
			return fmt.Errorf("resource is required when capacity_management is RESOURCE")
		}
		if hasBootstrap {
			return fmt.Errorf("bootstrap_capacity is only valid when capacity_management is COORDINATOR")
		}
		if hasHA && !hasHAResource {
			return fmt.Errorf("ha.resource is required when capacity_management is RESOURCE")
		}
	case CapacityManagedByCoordinator:
		oldMode, _ := d.GetChange("capacity_management")
		switchingFromResource := d.Id() != "" && d.HasChange("capacity_management") && oldMode == CapacityManagedByResource
		if (hasResource || hasHAResource) && !switchingFromResource {
			return fmt.Errorf("legacy resource and ha.resource blocks are forbidden when capacity_management is COORDINATOR")
		}
		if d.Id() == "" {
			if chargeType != "PRE" {
				return fmt.Errorf("new COORDINATOR workspaces require PRE billing and bootstrap_capacity")
			}
			if !hasBootstrap {
				return fmt.Errorf("bootstrap_capacity is required for a new PRE workspace when capacity_management is COORDINATOR")
			}
			if err := validateBootstrapCapacity(d.Get("bootstrap_capacity"), hasHA); err != nil {
				return err
			}
		} else if d.HasChange("bootstrap_capacity") {
			// bootstrap_capacity is an input to CreateInstance only. Ignore edits
			// after creation instead of replacing or updating the workspace.
			if err := d.Clear("bootstrap_capacity"); err != nil {
				return err
			}
		}
	}
	return nil
}

func flinkChildCapacityCustomizeDiff(legacyFields ...string) schema.CustomizeDiffFunc {
	return func(d *schema.ResourceDiff, _ interface{}) error {
		if d.Get("capacity_management").(string) != CapacityManagedByCoordinator {
			return nil
		}
		for _, field := range legacyFields {
			if flinkListBlockConfigured(d.Get(field)) {
				return fmt.Errorf("legacy capacity field %q is forbidden when capacity_management is COORDINATOR", field)
			}
		}
		if d.Id() == "" {
			return fmt.Errorf("new Flink child resources must be created with capacity_management RESOURCE, then switched to COORDINATOR in a later apply")
		}
		return nil
	}
}

func validateBootstrapCapacity(value interface{}, workspaceHA bool) error {
	bootstrap, ok := flinkFirstBlock(value)
	if !ok {
		return fmt.Errorf("bootstrap_capacity is required")
	}
	fixed, _ := bootstrap["fixed_cu"].(int)
	ha, hasBootstrapHA := flinkFirstBlock(bootstrap["ha"])
	crossZone := 0
	if hasBootstrapHA {
		crossZone, _ = ha["cross_zone_fixed_cu"].(int)
	}
	if workspaceHA {
		if !hasBootstrapHA || crossZone <= 0 {
			return fmt.Errorf("bootstrap_capacity.ha.cross_zone_fixed_cu must be greater than zero for an HA workspace")
		}
	} else if hasBootstrapHA {
		return fmt.Errorf("bootstrap_capacity.ha requires a workspace ha block")
	}
	if fixed+crossZone <= 0 {
		return fmt.Errorf("bootstrap_capacity fixed CU total must be greater than zero")
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

func suppressFlinkLegacyCapacityDiff(_ string, _ string, _ string, d *schema.ResourceData) bool {
	return d.Get("capacity_management").(string) == CapacityManagedByCoordinator
}

func expandFlinkBootstrapCapacity(value interface{}) (fixedCU, crossZoneFixedCU int) {
	bootstrap, ok := flinkFirstBlock(value)
	if !ok {
		return 0, 0
	}
	fixedCU, _ = bootstrap["fixed_cu"].(int)
	if ha, ok := flinkFirstBlock(bootstrap["ha"]); ok {
		crossZoneFixedCU, _ = ha["cross_zone_fixed_cu"].(int)
	}
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
