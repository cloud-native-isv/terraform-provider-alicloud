package alicloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestFlinkWorkspaceCapacityManagementDefaultsToResource(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspace().Schema, workspaceConfig("RESOURCE", true, false))
	if got := data.Get("capacity_management"); got != CapacityManagedByResource {
		t.Fatalf("raw capacity_management = %#v", got)
	}
	if got := flinkCapacityManagementValue(data.Get("capacity_management")); got != "RESOURCE" {
		t.Fatalf("capacity_management = %#v", got)
	}
}

func TestFlinkLegacyStateDoesNotDiffOnCapacityManagementDefault(t *testing.T) {
	config := workspaceConfig("RESOURCE", true, false)
	delete(config, "capacity_management")
	state := workspaceState("RESOURCE", "PRE", true)
	delete(state.Attributes, "capacity_management")

	diff, err := resourceAliCloudFlinkWorkspace().Diff(state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertNoDiffPrefix(t, diff, "capacity_management")
}

func TestFlinkChildLegacyStateDoesNotDiffOnCapacityManagementDefault(t *testing.T) {
	for name, test := range map[string]struct {
		resource *schema.Resource
		config   map[string]interface{}
		state    *terraform.InstanceState
	}{
		"namespace": {
			resource: resourceAliCloudFlinkNamespace(),
			config: map[string]interface{}{
				"workspace_id":   "f-test",
				"namespace_name": "default",
				"guaranteed_resource_spec": []interface{}{map[string]interface{}{
					"cpu": 2, "memory_gb": 8,
				}},
			},
			state: &terraform.InstanceState{ID: "f-test:default", Attributes: map[string]string{
				"workspace_id":                         "f-test",
				"namespace_name":                       "default",
				"guaranteed_resource_spec.#":           "1",
				"guaranteed_resource_spec.0.cpu":       "2",
				"guaranteed_resource_spec.0.memory_gb": "8",
				"elastic_resource_spec.#":              "0",
				"ha":                                   "false",
				"status":                               "Available",
			}},
		},
		"deployment target": {
			resource: resourceAliCloudFlinkDeploymentTarget(),
			config: map[string]interface{}{
				"workspace_id":   "f-test",
				"namespace_name": "default",
				"name":           "q",
				"quota": []interface{}{map[string]interface{}{
					"limit": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
				}},
			},
			state: &terraform.InstanceState{ID: "f-test:default:q", Attributes: map[string]string{
				"workspace_id":              "f-test",
				"namespace_name":            "default",
				"name":                      "q",
				"quota.#":                   "1",
				"quota.0.request.#":         "0",
				"quota.0.limit.#":           "1",
				"quota.0.limit.0.cpu":       "2",
				"quota.0.limit.0.memory_gb": "8",
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			diff, err := test.resource.Diff(test.state, terraform.NewResourceConfigRaw(test.config), nil)
			if err != nil {
				t.Fatal(err)
			}
			assertNoDiffPrefix(t, diff, "capacity_management")
		})
	}
}

func TestFlinkLifecycleResourceSchemasValidate(t *testing.T) {
	for name, resource := range map[string]*schema.Resource{
		"workspace":         resourceAliCloudFlinkWorkspace(),
		"namespace":         resourceAliCloudFlinkNamespace(),
		"deployment_target": resourceAliCloudFlinkDeploymentTarget(),
	} {
		t.Run(name, func(t *testing.T) {
			if err := resource.InternalValidate(nil, true); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFlinkDeploymentTargetQuotaSupportsV2Request(t *testing.T) {
	resource := resourceAliCloudFlinkDeploymentTarget()
	quota := resource.Schema["quota"].Elem.(*schema.Resource)
	request := quota.Schema["request"]
	if request == nil || !request.Optional || !request.Computed {
		t.Fatalf("quota.request schema = %#v", request)
	}

	configured := []interface{}{map[string]interface{}{
		"request": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
		"limit":   []interface{}{map[string]interface{}{"cpu": 4.0, "memory_gb": 16.0}},
	}}
	got := expandResourceQuota(configured)
	if got.Request == nil || got.Request.Cpu != 2 || got.Limit == nil || got.Limit.Cpu != 4 {
		t.Fatalf("expanded quota = %#v", got)
	}
	flat := firstTestBlock(flattenResourceQuota(got))
	if !flinkListBlockConfigured(flat["request"]) || !flinkListBlockConfigured(flat["limit"]) {
		t.Fatalf("flattened quota = %#v", flat)
	}
}

func TestFlinkDeploymentTargetQuotaAllowsZeroFixedRequest(t *testing.T) {
	config := map[string]interface{}{
		"workspace_id":   "f-test",
		"namespace_name": "default",
		"name":           "elastic-only",
		"quota": []interface{}{map[string]interface{}{
			"request": []interface{}{map[string]interface{}{"cpu": 0.0, "memory_gb": 0.0}},
			"limit":   []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
		}},
	}
	if _, err := resourceAliCloudFlinkDeploymentTarget().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatal(err)
	}
	quota := expandResourceQuota(config["quota"].([]interface{}))
	if quota.Request == nil || quota.Request.Cpu != 0 || quota.Request.MemoryGB != 0 {
		t.Fatalf("expanded quota = %#v", quota)
	}
}

func TestFlinkWorkspaceCapacityManagementValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{name: "resource mode requires legacy resource", config: workspaceConfig("RESOURCE", false, false), wantErr: "resource"},
		{name: "coordinator rejects legacy resource", config: workspaceConfig("COORDINATOR", true, true), wantErr: "legacy"},
		{name: "resource rejects bootstrap", config: workspaceConfig("RESOURCE", true, true), wantErr: "bootstrap_capacity"},
		{name: "new pre coordinator requires bootstrap", config: workspaceConfig("COORDINATOR", false, false), wantErr: "bootstrap_capacity"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(tc.config), nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Diff() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestFlinkWorkspacePostHARejected(t *testing.T) {
	config := workspaceConfig("RESOURCE", true, false)
	config["charge_type"] = "POST"
	config["ha"] = []interface{}{map[string]interface{}{
		"zone_id":     "cn-test-b",
		"vswitch_ids": []interface{}{"vsw-b"},
		"resource":    []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
	}}
	_, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "POST") {
		t.Fatalf("Diff() error = %v", err)
	}
}

func TestFlinkWorkspaceChargeTypeChangeReturnsError(t *testing.T) {
	config := workspaceConfig("RESOURCE", true, false)
	config["charge_type"] = "POST"
	state := &terraform.InstanceState{ID: "f-test", Attributes: map[string]string{
		"charge_type":          "PRE",
		"capacity_management":  "RESOURCE",
		"resource.#":           "1",
		"resource.0.cpu":       "2",
		"resource.0.memory":    "8",
		"storage.#":            "1",
		"storage.0.oss_bucket": "bucket",
		"vswitch_ids.#":        "1",
		"vswitch_ids.0":        "vsw-a",
		"zone_id":              "cn-test-a",
		"vpc_id":               "vpc-1",
		"name":                 "workspace",
		"resource_group_id":    "rg-1",
		"architecture_type":    "X86",
		"monitor_type":         "ARMS",
		"auto_renew":           "true",
		"duration":             "1",
		"pricing_cycle":        "Month",
	}}
	_, err := resourceAliCloudFlinkWorkspace().Diff(state, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "charge_type") {
		t.Fatalf("Diff() error = %v", err)
	}
}

func TestFlinkWorkspaceLegacyCapacityChangeStillRequiresNew(t *testing.T) {
	config := workspaceConfig("RESOURCE", true, false)
	config["resource"] = []interface{}{map[string]interface{}{"cpu": 4, "memory": 16}}
	diff, err := resourceAliCloudFlinkWorkspace().Diff(workspaceState("RESOURCE", "PRE", true), terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !diffRequiresNew(diff) {
		t.Fatalf("legacy capacity diff does not require replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceModeSwitchDoesNotRequireNew(t *testing.T) {
	config := workspaceConfig("COORDINATOR", false, false)
	diff, err := resourceAliCloudFlinkWorkspace().Diff(workspaceState("RESOURCE", "PRE", true), terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("capacity ownership switch unexpectedly requires replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceModeSwitchBackDoesNotRequireNew(t *testing.T) {
	config := workspaceConfig("RESOURCE", true, false)
	diff, err := resourceAliCloudFlinkWorkspace().Diff(workspaceState("COORDINATOR", "PRE", false), terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("capacity ownership switch unexpectedly requires replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceBootstrapCapacityValidation(t *testing.T) {
	config := workspaceConfig("COORDINATOR", false, true)
	config["bootstrap_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 0}}
	if _, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("non-HA zero bootstrap error = %v", err)
	}

	config = workspaceConfig("COORDINATOR", false, false)
	config["ha"] = []interface{}{map[string]interface{}{
		"vswitch_ids": []interface{}{"vsw-b"},
	}}
	config["bootstrap_capacity"] = []interface{}{map[string]interface{}{
		"fixed_cu": 0,
		"ha": []interface{}{map[string]interface{}{
			"cross_zone_fixed_cu": 2,
		}},
	}}
	if _, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatalf("pure HA bootstrap rejected: %v", err)
	}
}

func TestFlinkChildCapacityManagementRejectsLegacyCapacity(t *testing.T) {
	namespaceConfig := map[string]interface{}{
		"workspace_id":        "f-test",
		"namespace_name":      "default",
		"capacity_management": "COORDINATOR",
		"elastic_resource_spec": []interface{}{map[string]interface{}{
			"cpu": 2, "memory_gb": 8,
		}},
	}
	if _, err := resourceAliCloudFlinkNamespace().Diff(nil, terraform.NewResourceConfigRaw(namespaceConfig), nil); err == nil || !strings.Contains(err.Error(), "legacy") {
		t.Fatalf("namespace Diff() error = %v", err)
	}

	targetConfig := map[string]interface{}{
		"workspace_id":        "f-test",
		"namespace_name":      "default",
		"name":                "q",
		"capacity_management": "COORDINATOR",
		"quota": []interface{}{map[string]interface{}{
			"limit": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
		}},
	}
	if _, err := resourceAliCloudFlinkDeploymentTarget().Diff(nil, terraform.NewResourceConfigRaw(targetConfig), nil); err == nil || !strings.Contains(err.Error(), "legacy") {
		t.Fatalf("deployment target Diff() error = %v", err)
	}
}

func TestFlinkChildCapacityManagementRequiresStagedCreation(t *testing.T) {
	for name, resourceAndConfig := range map[string]struct {
		resource *schema.Resource
		config   map[string]interface{}
	}{
		"namespace": {
			resource: resourceAliCloudFlinkNamespace(),
			config: map[string]interface{}{
				"workspace_id":        "f-test",
				"namespace_name":      "custom",
				"capacity_management": "COORDINATOR",
			},
		},
		"deployment target": {
			resource: resourceAliCloudFlinkDeploymentTarget(),
			config: map[string]interface{}{
				"workspace_id":        "f-test",
				"namespace_name":      "default",
				"name":                "custom-queue",
				"capacity_management": "COORDINATOR",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := resourceAndConfig.resource.Diff(nil, terraform.NewResourceConfigRaw(resourceAndConfig.config), nil)
			if err == nil || !strings.Contains(err.Error(), "created with capacity_management RESOURCE") {
				t.Fatalf("Diff() error = %v", err)
			}
		})
	}
}

func TestFlinkChildModeSwitchBackSuppressesCapacityWriteDiff(t *testing.T) {
	namespaceConfig := map[string]interface{}{
		"workspace_id":        "f-test",
		"namespace_name":      "default",
		"capacity_management": "RESOURCE",
		"guaranteed_resource_spec": []interface{}{map[string]interface{}{
			"cpu": 2, "memory_gb": 8,
		}},
	}
	namespaceState := &terraform.InstanceState{ID: "f-test:default", Attributes: map[string]string{
		"workspace_id":               "f-test",
		"namespace_name":             "default",
		"capacity_management":        "COORDINATOR",
		"guaranteed_resource_spec.#": "0",
		"elastic_resource_spec.#":    "0",
		"ha":                         "false",
		"status":                     "Available",
	}}
	diff, err := resourceAliCloudFlinkNamespace().Diff(namespaceState, terraform.NewResourceConfigRaw(namespaceConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertNoDiffPrefix(t, diff, "guaranteed_resource_spec")

	targetConfig := map[string]interface{}{
		"workspace_id":        "f-test",
		"namespace_name":      "default",
		"name":                "q",
		"capacity_management": "RESOURCE",
		"quota": []interface{}{map[string]interface{}{
			"limit": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
		}},
	}
	targetState := &terraform.InstanceState{ID: "f-test:default:q", Attributes: map[string]string{
		"workspace_id":        "f-test",
		"namespace_name":      "default",
		"name":                "q",
		"capacity_management": "COORDINATOR",
		"quota.#":             "0",
	}}
	diff, err = resourceAliCloudFlinkDeploymentTarget().Diff(targetState, terraform.NewResourceConfigRaw(targetConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertNoDiffPrefix(t, diff, "quota")
}

func workspaceConfig(mode string, legacyResource, bootstrap bool) map[string]interface{} {
	config := map[string]interface{}{
		"name":                "workspace",
		"resource_group_id":   "rg-1",
		"zone_id":             "cn-test-a",
		"vpc_id":              "vpc-1",
		"vswitch_ids":         []interface{}{"vsw-a"},
		"charge_type":         "PRE",
		"capacity_management": mode,
		"storage": []interface{}{map[string]interface{}{
			"oss_bucket": "bucket",
		}},
	}
	if legacyResource {
		config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
	}
	if bootstrap {
		config["bootstrap_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 2}}
	}
	return config
}

func workspaceState(mode, chargeType string, legacyResource bool) *terraform.InstanceState {
	attributes := map[string]string{
		"capacity_management":  mode,
		"charge_type":          chargeType,
		"storage.#":            "1",
		"storage.0.oss_bucket": "bucket",
		"vswitch_ids.#":        "1",
		"vswitch_ids.0":        "vsw-a",
		"zone_id":              "cn-test-a",
		"vpc_id":               "vpc-1",
		"name":                 "workspace",
		"resource_group_id":    "rg-1",
		"architecture_type":    "X86",
		"monitor_type":         "ARMS",
		"auto_renew":           "true",
		"duration":             "1",
		"pricing_cycle":        "Month",
		"ha.#":                 "0",
	}
	if legacyResource {
		attributes["resource.#"] = "1"
		attributes["resource.0.cpu"] = "2"
		attributes["resource.0.memory"] = "8"
	} else {
		attributes["resource.#"] = "0"
	}
	return &terraform.InstanceState{ID: "f-test", Attributes: attributes}
}

func diffRequiresNew(diff *terraform.InstanceDiff) bool {
	if diff == nil {
		return false
	}
	for _, attribute := range diff.Attributes {
		if attribute != nil && attribute.RequiresNew {
			return true
		}
	}
	return false
}

func assertNoDiffPrefix(t *testing.T, diff *terraform.InstanceDiff, prefix string) {
	t.Helper()
	for key := range diff.Attributes {
		if strings.HasPrefix(key, prefix) {
			t.Fatalf("unexpected %s diff: %#v", prefix, diff.Attributes)
		}
	}
}
