package alicloud

import (
	"errors"
	"strings"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestValidateFlinkCapacityWorkspaceReady(t *testing.T) {
	ready := &flink.Workspace{
		Id:         "f-test",
		Status:     "RUNNING",
		OrderState: "NORMAL",
		ResourceId: "sc-test",
	}
	if err := validateFlinkCapacityWorkspaceReady(ready); err != nil {
		t.Fatal(err)
	}

	for name, mutate := range map[string]func(*flink.Workspace){
		"creating":            func(workspace *flink.Workspace) { workspace.Status = "CREATING" },
		"order pending":       func(workspace *flink.Workspace) { workspace.OrderState = "PROCESSING" },
		"resource id missing": func(workspace *flink.Workspace) { workspace.ResourceId = "" },
		"elastic id missing": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
		},
		"elastic spec missing": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticInstanceId = "f-elastic"
		},
	} {
		t.Run(name, func(t *testing.T) {
			workspace := *ready
			mutate(&workspace)
			err := validateFlinkCapacityWorkspaceReady(&workspace)
			var retryable interface{ Retryable() bool }
			if err == nil || !errors.As(err, &retryable) || !retryable.Retryable() {
				t.Fatalf("error = %v, want retryable", err)
			}
		})
	}
}

func TestValidateFlinkCapacityWorkspaceRejectsContradictoryElasticState(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		Status:              "RUNNING",
		OrderState:          "NORMAL",
		ResourceId:          "sc-test",
		Elastic:             false,
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
	}
	err := validateFlinkCapacityWorkspaceReady(workspace)
	if err == nil || !strings.Contains(err.Error(), "Elastic=false") {
		t.Fatalf("error = %v", err)
	}
	var retryable interface{ Retryable() bool }
	if errors.As(err, &retryable) && retryable.Retryable() {
		t.Fatalf("contradictory state must be terminal: %v", err)
	}
}

func TestValidateFlinkCapacityNamespaceReady(t *testing.T) {
	for _, status := range []string{"SUCCESS", "Available"} {
		if err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: status}); err != nil {
			t.Fatalf("status %q: %v", status, err)
		}
	}
	for _, status := range []string{"CREATING", "MODIFYING"} {
		err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: status})
		var retryable interface{ Retryable() bool }
		if err == nil || !errors.As(err, &retryable) || !retryable.Retryable() {
			t.Fatalf("status %q error = %v, want retryable", status, err)
		}
	}
	if err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: "FAILED"}); err == nil {
		t.Fatal("FAILED namespace status was accepted")
	}
}

func TestBuildFlinkCapacityTreeFromObjects(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		ResourceId:          "sc-test",
		ChargeType:          "PRE",
		ResourceSpec:        &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		HaResourceSpec:      &flink.ResourceSpec{Cpu: 3, MemoryGB: 12},
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		ClusterUsedResources: &flink.WorkspaceUsedResources{
			UsedResource: 1.5,
		},
	}
	namespaces := []flink.Namespace{{
		Name:                   "default",
		GuaranteedResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		ElasticResourceSpec:    &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		ResourceUsed:           &flink.ResourceUsed{Cu: 1},
	}}
	targets := map[string][]flink.DeploymentTarget{
		"default": {{
			Name: "q",
			Quota: &flink.ResourceQuota{
				Request: &flink.ResourceSpec{Cpu: 1.5, MemoryGB: 6},
				Limit:   &flink.ResourceSpec{Cpu: 3, MemoryGB: 12},
				Used:    &flink.ResourceSpec{Cpu: 0.5, MemoryGB: 2},
			},
		}},
	}

	got, err := buildFlinkCapacityTree(workspace, namespaces, targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.Workspace != (flinkcapacity.WorkspaceCapacity{FixedCU: 4, CrossZoneFixedCU: 6, Limit: 18, Used: 3}) {
		t.Fatalf("workspace capacity = %#v", got.Workspace)
	}
	if got.Namespaces[0].Capacity == nil || *got.Namespaces[0].Capacity != (flinkcapacity.Capacity{Fixed: 8, Limit: 12}) || got.Namespaces[0].Used != 2 {
		t.Fatalf("namespace = %#v", got.Namespaces[0])
	}
	queue := got.Namespaces[0].Queues[0]
	if queue.Capacity == nil || *queue.Capacity != (flinkcapacity.Capacity{Fixed: 3, Limit: 6}) || queue.Used != 1 {
		t.Fatalf("queue = %#v", queue)
	}
}

func TestBuildFlinkCapacityTreePostpaid(t *testing.T) {
	workspace := &flink.Workspace{ChargeType: "POST", ResourceSpec: &flink.ResourceSpec{Cpu: 8, MemoryGB: 32}}
	namespaces := []flink.Namespace{{Name: "default", GuaranteedResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16}, ElasticResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16}}}
	targets := map[string][]flink.DeploymentTarget{"default": {{Name: "q", Quota: &flink.ResourceQuota{Request: &flink.ResourceSpec{Cpu: 4}, Limit: &flink.ResourceSpec{Cpu: 8}, Used: &flink.ResourceSpec{}}}}}

	got, err := buildFlinkCapacityTree(workspace, namespaces, targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.Workspace.TotalFixed() != 0 || got.Workspace.Limit != 16 {
		t.Fatalf("workspace capacity = %#v", got.Workspace)
	}
}

func TestFlinkResourceSpecForCU(t *testing.T) {
	got := flinkResourceSpecForCU(3)
	if got.Cpu != 1.5 || got.MemoryGB != 6 {
		t.Fatalf("resource spec = %#v", got)
	}
}
