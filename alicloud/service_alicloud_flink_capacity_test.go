package alicloud

import (
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestBuildFlinkCapacityTreeFromObjects(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		ResourceId:          "sc-test",
		ChargeType:          "PRE",
		ResourceSpec:        &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		HaResourceSpec:      &flink.ResourceSpec{Cpu: 3, MemoryGB: 12},
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
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
	if got.Workspace != (flinkcapacity.WorkspaceCapacity{FixedCU: 4, CrossZoneFixedCU: 6, Limit: 18}) {
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
