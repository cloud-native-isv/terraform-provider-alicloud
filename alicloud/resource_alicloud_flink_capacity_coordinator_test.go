package alicloud

import (
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestFlinkCapacityCoordinatorSchema(t *testing.T) {
	resource := resourceAliCloudFlinkCapacityCoordinator()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatal(err)
	}
	if !resource.Schema["workspace_instance_id"].Required || !resource.Schema["workspace_instance_id"].ForceNew {
		t.Fatal("workspace_instance_id must be required and ForceNew")
	}
	if !resource.Schema["workspace_resource_id"].Computed {
		t.Fatal("workspace_resource_id must be computed")
	}
	if got := *resource.Timeouts.Create; got != 60*time.Minute {
		t.Fatalf("create timeout = %s", got)
	}
	if got := *resource.Timeouts.Update; got != 60*time.Minute {
		t.Fatalf("update timeout = %s", got)
	}
}

func TestFlinkCapacityCoordinatorRegistered(t *testing.T) {
	provider := Provider().(*schema.Provider)
	if provider.ResourcesMap["alicloud_flink_capacity_coordinator"] == nil {
		t.Fatal("alicloud_flink_capacity_coordinator is not registered")
	}
}

func TestExpandFlinkCoordinatorDesiredCapacity(t *testing.T) {
	tree, err := expandFlinkCoordinatorDesired(coordinatorWorkspaceConfig())
	if err != nil {
		t.Fatal(err)
	}
	tree.ChargeType = "PRE"
	resolved, err := flinkcapacity.Resolve(tree)
	if err != nil {
		t.Fatal(err)
	}
	if got := resolved.Workspace.FixedCU.Float64(); got != 4 {
		t.Fatalf("workspace fixed CU = %v", got)
	}
	if got := resolved.Workspace.Limit.Float64(); got != 8 {
		t.Fatalf("workspace limit = %v", got)
	}
	if got := resolved.Namespaces[0].Capacity; got == nil || got.Fixed.Float64() != 4 || got.Limit.Float64() != 8 {
		t.Fatalf("namespace remainder = %#v", got)
	}
	if got := resolved.Namespaces[0].Queues[0].Capacity; got == nil || got.Fixed.Float64() != 4 || got.Limit.Float64() != 8 {
		t.Fatalf("queue remainder = %#v", got)
	}
}

func TestExpandFlinkCoordinatorDesiredRejectsInvalidCapacity(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]interface{})
		wantErr string
	}{
		{
			name: "limit aliases conflict",
			mutate: func(workspace map[string]interface{}) {
				capacity := firstTestBlock(workspace["capacity"])
				capacity["elastic_cu_limit"] = 4.0
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "workspace must use integer CU",
			mutate: func(workspace map[string]interface{}) {
				firstTestBlock(workspace["capacity"])["fixed_cu"] = 1.5
			},
			wantErr: "integer CU",
		},
		{
			name: "namespace must use integer CU",
			mutate: func(workspace map[string]interface{}) {
				namespace := firstTestBlock(workspace["namespace"])
				namespace["capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 1.5, "max_cu_limit": 2.0}}
			},
			wantErr: "integer CU",
		},
		{
			name: "queue only allows half CU",
			mutate: func(workspace map[string]interface{}) {
				queue := firstTestBlock(firstTestBlock(workspace["namespace"])["queue"])
				queue["capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 0.1, "max_cu_limit": 1.0}}
			},
			wantErr: "multiple of 0.5",
		},
		{
			name: "duplicate namespace",
			mutate: func(workspace map[string]interface{}) {
				namespace := firstTestBlock(workspace["namespace"])
				workspace["namespace"] = []interface{}{namespace, cloneTestMap(namespace)}
			},
			wantErr: "duplicate namespace",
		},
		{
			name: "two remainder queues",
			mutate: func(workspace map[string]interface{}) {
				namespace := firstTestBlock(workspace["namespace"])
				namespace["queue"] = []interface{}{
					map[string]interface{}{"name": "q1"},
					map[string]interface{}{"name": "q2"},
				}
			},
			wantErr: "at most one",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			workspace := coordinatorWorkspaceConfig()
			tc.mutate(firstTestBlock(workspace))
			_, err := expandFlinkCoordinatorDesired(workspace)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestFlinkCapacityCoordinatorCustomizeDiffRejectsAliases(t *testing.T) {
	config := map[string]interface{}{
		"workspace_instance_id": "f-test",
		"workspace":             coordinatorWorkspaceConfig(),
	}
	capacity := firstTestBlock(firstTestBlock(config["workspace"])["capacity"])
	capacity["elastic_cu_limit"] = 4.0
	_, err := resourceAliCloudFlinkCapacityCoordinator().Diff(nil, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("Diff() error = %v", err)
	}
}

func TestFlinkCapacityCoordinatorCustomizeDiffAcceptsValidConfig(t *testing.T) {
	config := map[string]interface{}{
		"workspace_instance_id": "f-test",
		"workspace":             coordinatorWorkspaceConfig(),
	}
	if _, err := resourceAliCloudFlinkCapacityCoordinator().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatal(err)
	}
}

func TestExpandFlinkCoordinatorWorkspaceElasticIsInAdditionToAllFixedCU(t *testing.T) {
	workspace := coordinatorWorkspaceConfig()
	capacity := firstTestBlock(firstTestBlock(workspace)["capacity"])
	capacity["fixed_cu"] = 0.0
	delete(capacity, "max_cu_limit")
	capacity["elastic_cu_limit"] = 4.0
	capacity["ha"] = []interface{}{map[string]interface{}{"cross_zone_fixed_cu": 4.0}}

	tree, err := expandFlinkCoordinatorDesired(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if got := tree.Workspace.Limit.Float64(); got != 8 {
		t.Fatalf("workspace limit = %v, want 8", got)
	}
	if got := tree.Workspace.AsCapacity().Elastic().Float64(); got != 4 {
		t.Fatalf("workspace elastic CU = %v, want 4", got)
	}
}

func TestFlattenFlinkCoordinatorImportUsesExplicitElasticAlias(t *testing.T) {
	actual := coordinatorActualTree()
	flat := flattenFlinkCoordinatorImport(actual)
	workspace := firstTestBlock(flat)
	capacity := firstTestBlock(workspace["capacity"])
	if capacity["fixed_cu"] != 4.0 || capacity["elastic_cu_limit"] != 4.0 {
		t.Fatalf("workspace capacity = %#v", capacity)
	}
	if _, exists := capacity["max_cu_limit"]; exists {
		t.Fatalf("import unexpectedly populated max_cu_limit: %#v", capacity)
	}
	namespace := firstTestBlock(workspace["namespace"])
	if _, exists := namespace["capacity"]; !exists {
		t.Fatalf("import must expand namespace capacity: %#v", namespace)
	}
	queue := firstTestBlock(namespace["queue"])
	if _, exists := queue["capacity"]; !exists {
		t.Fatalf("import must expand queue capacity: %#v", queue)
	}
}

func TestMergeFlinkCoordinatorObservedPreservesIntentShape(t *testing.T) {
	configured := coordinatorWorkspaceConfig()
	merged, err := mergeFlinkCoordinatorObserved(configured, coordinatorActualTree())
	if err != nil {
		t.Fatal(err)
	}
	workspace := firstTestBlock(merged)
	capacity := firstTestBlock(workspace["capacity"])
	if _, exists := capacity["max_cu_limit"]; !exists {
		t.Fatalf("configured max_cu_limit alias was lost: %#v", capacity)
	}
	if _, exists := capacity["elastic_cu_limit"]; exists {
		t.Fatalf("merge introduced elastic_cu_limit: %#v", capacity)
	}
	namespace := firstTestBlock(workspace["namespace"])
	if _, exists := namespace["capacity"]; exists {
		t.Fatalf("remainder namespace gained explicit capacity: %#v", namespace)
	}
	queue := firstTestBlock(namespace["queue"])
	if _, exists := queue["capacity"]; exists {
		t.Fatalf("remainder queue gained explicit capacity: %#v", queue)
	}
	if len(firstTestList(workspace["observed_capacity"])) != 1 || len(firstTestList(namespace["observed_capacity"])) != 1 || len(firstTestList(queue["observed_capacity"])) != 1 {
		t.Fatalf("observed capacities were not populated: %#v", merged)
	}
}

func TestFlinkCapacityCoordinatorDeleteIsNoOp(t *testing.T) {
	resource := resourceAliCloudFlinkCapacityCoordinator()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"workspace_instance_id": "f-test",
		"workspace":             coordinatorWorkspaceConfig(),
	})
	data.SetId("f-test")
	if err := resource.Delete(data, nil); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "" {
		t.Fatalf("coordinator ID was not cleared: %q", data.Id())
	}
}

func coordinatorWorkspaceConfig() []interface{} {
	return []interface{}{map[string]interface{}{
		"capacity": []interface{}{map[string]interface{}{
			"fixed_cu":     4.0,
			"max_cu_limit": 8.0,
		}},
		"namespace": []interface{}{map[string]interface{}{
			"name": "default",
			"queue": []interface{}{map[string]interface{}{
				"name": "default-queue",
			}},
		}},
	}}
}

func coordinatorActualTree() flinkcapacity.Tree {
	workspaceCapacity := flinkcapacity.Capacity{Fixed: 8, Limit: 16}
	queueCapacity := workspaceCapacity
	return flinkcapacity.Tree{
		ChargeType: "PRE",
		Workspace: flinkcapacity.WorkspaceCapacity{
			FixedCU: 8,
			Limit:   16,
		},
		Namespaces: []flinkcapacity.Namespace{{
			Name:     "default",
			Capacity: &workspaceCapacity,
			Queues: []flinkcapacity.Queue{{
				Name:     "default-queue",
				Capacity: &queueCapacity,
			}},
		}},
	}
}

func firstTestBlock(value interface{}) map[string]interface{} {
	items := firstTestList(value)
	if len(items) == 0 {
		return nil
	}
	return items[0].(map[string]interface{})
}

func firstTestList(value interface{}) []interface{} {
	items, _ := value.([]interface{})
	return items
}

func cloneTestMap(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
