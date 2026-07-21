package flinkworkspace

import (
	"reflect"
	"strings"
	"testing"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestWorkspaceCreateIntentFingerprintCoversEveryPurchaseInput(t *testing.T) {
	falseValue := false
	trueValue := true
	one := int32(1)
	three := int32(3)
	baseWorkspace := func() *aliyunFlinkAPI.Workspace {
		return &aliyunFlinkAPI.Workspace{
			ChargeType:       "PRE",
			ResourceSpec:     &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8},
			HighAvailability: &aliyunFlinkAPI.HighAvailability{Enabled: false},
		}
	}
	baseOptions := func() CreateOptions {
		return CreateOptions{
			AutoRenew:        &trueValue,
			Duration:         &one,
			PricingCycle:     "Month",
			Extra:            "extra",
			PromotionCode:    "promotion",
			UsePromotionCode: &falseValue,
		}
	}
	baseFingerprint := WorkspaceCreateIntentFingerprint(baseWorkspace(), baseOptions(), CapacityIntentLegacy)
	if len(baseFingerprint) != 64 {
		t.Fatalf("fingerprint = %q, want 64 hex characters", baseFingerprint)
	}

	tests := map[string]func(*aliyunFlinkAPI.Workspace, *CreateOptions) string{
		"charge type": func(workspace *aliyunFlinkAPI.Workspace, _ *CreateOptions) string {
			workspace.ChargeType = "POST"
			return CapacityIntentLegacy
		},
		"ha enabled": func(workspace *aliyunFlinkAPI.Workspace, _ *CreateOptions) string {
			workspace.HighAvailability = &aliyunFlinkAPI.HighAvailability{Enabled: true, ResourceSpec: &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8}}
			return CapacityIntentLegacy
		},
		"capacity mode": func(*aliyunFlinkAPI.Workspace, *CreateOptions) string { return CapacityIntentInitial },
		"primary cpu": func(workspace *aliyunFlinkAPI.Workspace, _ *CreateOptions) string {
			workspace.ResourceSpec.Cpu = 4
			return CapacityIntentLegacy
		},
		"primary memory": func(workspace *aliyunFlinkAPI.Workspace, _ *CreateOptions) string {
			workspace.ResourceSpec.MemoryGB = 16
			return CapacityIntentLegacy
		},
		"auto renew": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.AutoRenew = &falseValue
			return CapacityIntentLegacy
		},
		"duration": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.Duration = &three
			return CapacityIntentLegacy
		},
		"pricing cycle": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.PricingCycle = "Year"
			return CapacityIntentLegacy
		},
		"extra": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.Extra = "different"
			return CapacityIntentLegacy
		},
		"promotion code": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.PromotionCode = "different"
			return CapacityIntentLegacy
		},
		"use promotion code": func(_ *aliyunFlinkAPI.Workspace, options *CreateOptions) string {
			options.UsePromotionCode = &trueValue
			return CapacityIntentLegacy
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			workspace := baseWorkspace()
			options := baseOptions()
			mode := mutate(workspace, &options)
			if got := WorkspaceCreateIntentFingerprint(workspace, options, mode); got == baseFingerprint {
				t.Fatalf("mutating %s did not change fingerprint %q", name, got)
			}
		})
	}
}

func TestWorkspaceCreateIntentFingerprintUsesUnambiguousEncodingAndHidesInputs(t *testing.T) {
	trueValue := true
	one := int32(1)
	workspace := &aliyunFlinkAPI.Workspace{ChargeType: "PRE", ResourceSpec: &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8}}
	first := CreateOptions{AutoRenew: &trueValue, Duration: &one, PricingCycle: "Month", Extra: "a\x00b", PromotionCode: "c", UsePromotionCode: &trueValue}
	second := CreateOptions{AutoRenew: &trueValue, Duration: &one, PricingCycle: "Month", Extra: "a", PromotionCode: "b\x00c", UsePromotionCode: &trueValue}
	firstFingerprint := WorkspaceCreateIntentFingerprint(workspace, first, CapacityIntentLegacy)
	secondFingerprint := WorkspaceCreateIntentFingerprint(workspace, second, CapacityIntentLegacy)
	if firstFingerprint == secondFingerprint {
		t.Fatalf("length-ambiguous inputs collided at %q", firstFingerprint)
	}
	for _, plaintext := range []string{"a\x00b", "promotion-secret"} {
		options := first
		options.PromotionCode = plaintext
		fingerprint := WorkspaceCreateIntentFingerprint(workspace, options, CapacityIntentLegacy)
		if strings.Contains(fingerprint, plaintext) {
			t.Fatalf("fingerprint %q leaked plaintext %q", fingerprint, plaintext)
		}
	}
}

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
