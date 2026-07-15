package flinkworkspace

import (
	"reflect"
	"testing"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestWorkspaceCreateTokenIsStableForIdentity(t *testing.T) {
	workspace := &aliyunFlinkAPI.Workspace{
		Name:   "workspace",
		Region: "cn-beijing",
		VpcId:  "vpc-a",
	}
	first := WorkspaceCreateToken(workspace)
	workspace.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: 16, MemoryGB: 64}
	second := WorkspaceCreateToken(workspace)
	if first == "" || first != second {
		t.Fatalf("tokens = %q and %q, want the same non-empty identity token", first, second)
	}
	workspace.Name = "other"
	if got := WorkspaceCreateToken(workspace); got == first {
		t.Fatalf("renamed workspace token = %q, want a different token", got)
	}
}

func TestHAConfigUsesDescribeInstanceFields(t *testing.T) {
	workspace := &aliyunFlinkAPI.Workspace{
		Ha:           true,
		HaZoneId:     "cn-beijing-g",
		HaVSwitchIds: []string{"vsw-standby"},
		HaResourceSpec: &aliyunFlinkAPI.ResourceSpec{
			Cpu:      2,
			MemoryGB: 8,
		},
	}

	config, ok := HAConfigWithFallback(workspace, "")
	if !ok {
		t.Fatal("HAConfig reported that HA is disabled")
	}
	if got := config["zone_id"]; got != "cn-beijing-g" {
		t.Fatalf("zone_id = %#v, want %q", got, "cn-beijing-g")
	}
	if got, want := config["vswitch_ids"], []string{"vsw-standby"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("vswitch_ids = %#v, want %#v", got, want)
	}
	resource, ok := config["resource"].(map[string]interface{})
	if !ok {
		t.Fatalf("resource = %#v, want a map", config["resource"])
	}
	if resource["cpu"] != 2 || resource["memory"] != 8 {
		t.Fatalf("resource = %#v, want cpu=2 memory=8", resource)
	}
}

func TestHAConfigUsesCreateFields(t *testing.T) {
	workspace := &aliyunFlinkAPI.Workspace{
		HighAvailability: &aliyunFlinkAPI.HighAvailability{
			Enabled:    true,
			ZoneId:     "cn-beijing-g",
			VSwitchIds: []string{"vsw-standby"},
			ResourceSpec: &aliyunFlinkAPI.ResourceSpec{
				Cpu:      2,
				MemoryGB: 8,
			},
		},
	}

	config, ok := HAConfigWithFallback(workspace, "")
	if !ok {
		t.Fatal("HAConfig reported that HA is disabled")
	}
	if got := config["zone_id"]; got != "cn-beijing-g" {
		t.Fatalf("zone_id = %#v, want %q", got, "cn-beijing-g")
	}
}

func TestZoneReadFallsBackToVSwitchInfoThenState(t *testing.T) {
	workspace := &aliyunFlinkAPI.Workspace{
		Ha:            true,
		VSwitchInfo:   []aliyunFlinkAPI.VSwitchInfo{{ZoneId: "cn-beijing-f"}},
		HaVSwitchInfo: []aliyunFlinkAPI.VSwitchInfo{{ZoneId: "cn-beijing-g"}},
	}
	if got := PrimaryZoneID(workspace, "state-primary"); got != "cn-beijing-f" {
		t.Fatalf("primary zone = %q", got)
	}
	config, ok := HAConfigWithFallback(workspace, "state-standby")
	if !ok || config["zone_id"] != "cn-beijing-g" {
		t.Fatalf("HA config = %#v, ok=%v", config, ok)
	}

	workspace.VSwitchInfo = nil
	workspace.HaVSwitchInfo = nil
	if got := PrimaryZoneID(workspace, "state-primary"); got != "state-primary" {
		t.Fatalf("primary state fallback = %q", got)
	}
	config, _ = HAConfigWithFallback(workspace, "state-standby")
	if config["zone_id"] != "state-standby" {
		t.Fatalf("HA state fallback = %#v", config)
	}
}

func TestHAConfigRejectsDisabledHA(t *testing.T) {
	for _, workspace := range []*aliyunFlinkAPI.Workspace{
		nil,
		{},
		{HighAvailability: &aliyunFlinkAPI.HighAvailability{}},
	} {
		if config, ok := HAConfigWithFallback(workspace, ""); ok || config != nil {
			t.Fatalf("HAConfigWithFallback(%#v) = (%#v, %t), want (nil, false)", workspace, config, ok)
		}
	}
}
