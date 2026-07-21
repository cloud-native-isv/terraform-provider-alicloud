package alicloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestFlinkWorkspaceInitialCapacitySchemaIsIntegerCreateOnly(t *testing.T) {
	workspace := resourceAliCloudFlinkWorkspace()
	initial := workspace.Schema["initial_capacity"]
	if initial == nil || initial.Type != schema.TypeList || !initial.Optional || initial.ForceNew || initial.MaxItems != 1 {
		t.Fatalf("initial_capacity schema = %#v", initial)
	}
	fields := initial.Elem.(*schema.Resource).Schema
	for _, name := range []string{"fixed_cu", "cross_zone_fixed_cu"} {
		field := fields[name]
		if field == nil || field.Type != schema.TypeInt || !field.Required || field.ForceNew {
			t.Fatalf("initial_capacity.%s schema = %#v", name, field)
		}
	}
}

func TestFlinkWorkspaceInitialCapacityRejectsFreshPostBeforeCreate(t *testing.T) {
	config := workspaceConfig("POST", true)
	if _, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err == nil || !strings.Contains(err.Error(), "PRE") {
		t.Fatalf("fresh POST initial_capacity Diff() error = %v", err)
	}
}

func TestFlinkWorkspaceInitialCapacityRejectsLegacyCapacityAndInvalidTopology(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{
			name: "legacy primary resource",
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", true)
				config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
				return config
			}(),
			wantErr: "resource",
		},
		{
			name: "legacy ha resource",
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", true)
				config["ha"] = []interface{}{map[string]interface{}{
					"vswitch_ids": []interface{}{"vsw-b"},
					"resource":    []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
				}}
				return config
			}(),
			wantErr: "ha.resource",
		},
		{
			name: "single zone cross zone capacity",
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", true)
				config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 0, "cross_zone_fixed_cu": 2}}
				return config
			}(),
			wantErr: "cross_zone_fixed_cu",
		},
		{
			name: "ha requires cross zone capacity",
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", true)
				config["ha"] = []interface{}{map[string]interface{}{"vswitch_ids": []interface{}{"vsw-b"}}}
				config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 2, "cross_zone_fixed_cu": 0}}
				return config
			}(),
			wantErr: "cross_zone_fixed_cu",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(tc.config), nil); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Diff() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestFlinkWorkspaceInitialCapacityAllowsPreTopology(t *testing.T) {
	for name, config := range map[string]map[string]interface{}{
		"single zone": workspaceConfig("PRE", true),
		"ha": func() map[string]interface{} {
			config := workspaceConfig("PRE", true)
			config["ha"] = []interface{}{map[string]interface{}{"vswitch_ids": []interface{}{"vsw-b"}}}
			config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 0, "cross_zone_fixed_cu": 2}}
			return config
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resourceAliCloudFlinkWorkspace().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFlinkWorkspaceInitialCapacityAndChargeTypeEditsRejectExistingWorkspace(t *testing.T) {
	for name, test := range map[string]struct {
		state  *terraform.InstanceState
		config map[string]interface{}
	}{
		"add initial capacity": {
			state:  legacyWorkspaceState(),
			config: workspaceConfig("PRE", true),
		},
		"modify initial capacity": {
			state: initialWorkspaceState(false),
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", true)
				config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 4, "cross_zone_fixed_cu": 0}}
				return config
			}(),
		},
		"charge type": {
			state: initialWorkspaceState(false),
			config: func() map[string]interface{} {
				config := workspaceConfig("POST", false)
				delete(config, "initial_capacity")
				config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
				return config
			}(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			diff, err := resourceAliCloudFlinkWorkspace().Diff(test.state, terraform.NewResourceConfigRaw(test.config), nil)
			if err == nil {
				t.Fatal("expected existing workspace change to be rejected")
			}
			if diffRequiresNew(diff) {
				t.Fatalf("existing workspace edit planned replacement: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceInitialCapacityRemovalRejectsLegacyReplacement(t *testing.T) {
	for name, test := range map[string]struct {
		state  *terraform.InstanceState
		config map[string]interface{}
	}{
		"single zone": {
			state: initialWorkspaceState(false),
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", false)
				config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
				return config
			}(),
		},
		"ha": {
			state: initialWorkspaceState(true),
			config: func() map[string]interface{} {
				config := workspaceConfig("PRE", false)
				config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
				config["ha"] = []interface{}{map[string]interface{}{
					"vswitch_ids": []interface{}{"vsw-b"},
					"resource":    []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
				}}
				return config
			}(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			diff, err := resourceAliCloudFlinkWorkspace().Diff(test.state, terraform.NewResourceConfigRaw(test.config), nil)
			if err == nil || !strings.Contains(err.Error(), "initial_capacity") {
				t.Fatalf("Diff() error = %v, want initial_capacity removal rejection", err)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("initial_capacity removal planned replacement: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceLegacyCapacityStillRequiresNew(t *testing.T) {
	config := workspaceConfig("PRE", false)
	delete(config, "initial_capacity")
	config["resource"] = []interface{}{map[string]interface{}{"cpu": 4, "memory": 16}}
	diff, err := resourceAliCloudFlinkWorkspace().Diff(legacyWorkspaceState(), terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !diffRequiresNew(diff) {
		t.Fatalf("legacy capacity diff does not require replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceAndChildrenRemoveUnpublishedCapacityHandshakes(t *testing.T) {
	for name, resource := range map[string]*schema.Resource{
		"workspace":         resourceAliCloudFlinkWorkspace(),
		"namespace":         resourceAliCloudFlinkNamespace(),
		"deployment target": resourceAliCloudFlinkDeploymentTarget(),
	} {
		t.Run(name, func(t *testing.T) {
			if _, exists := resource.Schema["capacity_management"]; exists {
				t.Fatalf("%s still exposes capacity_management", name)
			}
			if _, exists := resource.Schema["bootstrap_capacity"]; exists {
				t.Fatalf("%s still exposes bootstrap_capacity", name)
			}
			if resource.Importer == nil || resource.Importer.State == nil {
				t.Fatalf("%s lost its importer", name)
			}
			if err := resource.InternalValidate(nil, true); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFlinkNamespaceAndDeploymentTargetLegacyNoDiff(t *testing.T) {
	tests := map[string]struct {
		resource *schema.Resource
		config   map[string]interface{}
		state    *terraform.InstanceState
	}{
		"namespace": {
			resource: resourceAliCloudFlinkNamespace(),
			config: map[string]interface{}{
				"workspace_id": "f-test", "namespace_name": "default",
				"guaranteed_resource_spec": []interface{}{map[string]interface{}{"cpu": 2, "memory_gb": 8}},
			},
			state: &terraform.InstanceState{ID: "f-test:default", Attributes: map[string]string{
				"workspace_id": "f-test", "namespace_name": "default", "guaranteed_resource_spec.#": "1",
				"guaranteed_resource_spec.0.cpu": "2", "guaranteed_resource_spec.0.memory_gb": "8",
				"elastic_resource_spec.#": "0", "ha": "false", "status": "Available", "capacity_management": "RESOURCE",
			}},
		},
		"deployment target": {
			resource: resourceAliCloudFlinkDeploymentTarget(),
			config: map[string]interface{}{
				"workspace_id": "f-test", "namespace_name": "default", "name": "q",
				"quota": []interface{}{map[string]interface{}{"limit": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}}}},
			},
			state: &terraform.InstanceState{ID: "f-test:default:q", Attributes: map[string]string{
				"workspace_id": "f-test", "namespace_name": "default", "name": "q", "quota.#": "1",
				"quota.0.request.#": "0", "quota.0.limit.#": "1", "quota.0.limit.0.cpu": "2", "quota.0.limit.0.memory_gb": "8", "capacity_management": "RESOURCE",
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			diff, err := tc.resource.Diff(tc.state, terraform.NewResourceConfigRaw(tc.config), nil)
			if err != nil {
				t.Fatal(err)
			}
			assertNoDiffPrefix(t, diff, "capacity_management")
			switch name {
			case "namespace":
				assertNoDiffPrefix(t, diff, "elastic_resource_spec")
				assertNoDiffPrefix(t, diff, "guaranteed_resource_spec")
			case "deployment target":
				assertNoDiffPrefix(t, diff, "quota")
			}
		})
	}
}

func TestFlinkDeploymentTargetQuotaSupportsV2Request(t *testing.T) {
	configured := []interface{}{map[string]interface{}{
		"request": []interface{}{map[string]interface{}{"cpu": 2.0, "memory_gb": 8.0}},
		"limit":   []interface{}{map[string]interface{}{"cpu": 4.0, "memory_gb": 16.0}},
	}}
	got := expandResourceQuota(configured)
	if got.Request == nil || got.Request.Cpu != 2 || got.Limit == nil || got.Limit.Cpu != 4 {
		t.Fatalf("expanded quota = %#v", got)
	}
}

func workspaceConfig(chargeType string, initial bool) map[string]interface{} {
	config := map[string]interface{}{
		"name": "workspace", "resource_group_id": "rg-1", "zone_id": "cn-test-a", "vpc_id": "vpc-1",
		"vswitch_ids": []interface{}{"vsw-a"}, "charge_type": chargeType,
		"storage": []interface{}{map[string]interface{}{"oss_bucket": "bucket"}},
	}
	if initial {
		config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 2, "cross_zone_fixed_cu": 0}}
	}
	return config
}

func legacyWorkspaceState() *terraform.InstanceState {
	return &terraform.InstanceState{ID: "f-test", Attributes: map[string]string{
		"charge_type": "PRE", "resource.#": "1", "resource.0.cpu": "2", "resource.0.memory": "8",
		"storage.#": "1", "storage.0.oss_bucket": "bucket", "vswitch_ids.#": "1", "vswitch_ids.0": "vsw-a",
		"zone_id": "cn-test-a", "vpc_id": "vpc-1", "name": "workspace", "resource_group_id": "rg-1",
		"architecture_type": "X86", "monitor_type": "ARMS", "auto_renew": "true", "duration": "1", "pricing_cycle": "Month", "ha.#": "0", "capacity_management": "RESOURCE",
	}}
}

func initialWorkspaceState(ha bool) *terraform.InstanceState {
	state := legacyWorkspaceState()
	state.Attributes["resource.#"] = "0"
	delete(state.Attributes, "resource.0.cpu")
	delete(state.Attributes, "resource.0.memory")
	state.Attributes["initial_capacity.#"] = "1"
	state.Attributes["initial_capacity.0.fixed_cu"] = "2"
	state.Attributes["initial_capacity.0.cross_zone_fixed_cu"] = "0"
	if ha {
		state.Attributes["initial_capacity.0.fixed_cu"] = "0"
		state.Attributes["initial_capacity.0.cross_zone_fixed_cu"] = "2"
		state.Attributes["ha.#"] = "1"
		state.Attributes["ha.0.vswitch_ids.#"] = "1"
		state.Attributes["ha.0.vswitch_ids.0"] = "vsw-b"
		state.Attributes["ha.0.resource.#"] = "0"
	}
	return state
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
	if diff == nil {
		return
	}
	for key := range diff.Attributes {
		if strings.HasPrefix(key, prefix) {
			t.Fatalf("unexpected %s diff: %#v", prefix, diff.Attributes)
		}
	}
}
