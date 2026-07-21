package alicloud

import (
	"fmt"

	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func flinkInitialCapacitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"fixed_cu": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"cross_zone_fixed_cu": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
	chargeType := d.Get("charge_type").(string)
	hasResource := flinkListBlockConfigured(d.Get("resource"))
	initialCapacity := d.Get("initial_capacity")
	hasInitialCapacity := flinkListBlockConfigured(initialCapacity)
	ha, hasHA := flinkFirstBlock(d.Get("ha"))
	hasHAResource := hasHA && flinkListBlockConfigured(ha["resource"])

	if d.Id() != "" && d.HasChange("charge_type") {
		oldValue, newValue := d.GetChange("charge_type")
		return fmt.Errorf("charge_type cannot be changed in place from %q to %q; create a separate Flink workspace for billing migrations", oldValue, newValue)
	}
	if d.Id() != "" && d.HasChange("initial_capacity") {
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
		return nil
	}

	if !hasResource {
		return fmt.Errorf("resource is required when initial_capacity is not configured")
	}
	if hasHA && !hasHAResource {
		return fmt.Errorf("ha.resource is required when initial_capacity is not configured")
	}
	return nil
}

func validateInitialCapacity(value interface{}, workspaceHA bool) error {
	capacity, ok := flinkFirstBlock(value)
	if !ok {
		return fmt.Errorf("initial_capacity is required")
	}
	fixed, _ := capacity["fixed_cu"].(int)
	crossZone, _ := capacity["cross_zone_fixed_cu"].(int)
	if workspaceHA {
		if crossZone <= 0 {
			return fmt.Errorf("initial_capacity.cross_zone_fixed_cu must be greater than zero for an HA workspace")
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
