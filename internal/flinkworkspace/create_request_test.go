package flinkworkspace

import (
	"encoding/json"
	"reflect"
	"testing"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestBuildCreateInstanceBodyIncludesHAFields(t *testing.T) {
	workspace := &aliyunFlinkAPI.Workspace{
		Name:             "probe-ha-workspace",
		Region:           "cn-beijing",
		ChargeType:       "PRE",
		ArchitectureType: "X86",
		VpcId:            "vpc-primary",
		VSwitchIds:       []string{"vsw-primary"},
		ResourceGroupId:  "rg-probe",
		ResourceSpec: &aliyunFlinkAPI.ResourceSpec{
			Cpu:      0,
			MemoryGB: 0,
		},
		Storage: &aliyunFlinkAPI.Storage{
			Oss: &aliyunFlinkAPI.OSSStorage{Bucket: "probe-bucket"},
		},
		HighAvailability: &aliyunFlinkAPI.HighAvailability{
			Enabled:    true,
			VSwitchIds: []string{"vsw-standby"},
			ResourceSpec: &aliyunFlinkAPI.ResourceSpec{
				Cpu:      2,
				MemoryGB: 8,
			},
		},
	}

	options := CreateOptions{
		AutoRenew:        true,
		Duration:         1,
		PricingCycle:     "Month",
		MonitorType:      "ARMS",
		Extra:            "probe-extra",
		PromotionCode:    "probe-promotion",
		UsePromotionCode: true,
	}
	body, err := BuildCreateInstanceBody(workspace, options)
	if err != nil {
		t.Fatalf("BuildCreateInstanceBody returned an error: %v", err)
	}
	for key, want := range map[string]interface{}{
		"InstanceName":     "probe-ha-workspace",
		"Region":           "cn-beijing",
		"ChargeType":       "PRE",
		"ArchitectureType": "X86",
		"VpcId":            "vpc-primary",
		"ResourceGroupId":  "rg-probe",
		"AutoRenew":        true,
		"Duration":         1,
		"PricingCycle":     "Month",
		"MonitorType":      "ARMS",
		"Extra":            "probe-extra",
		"PromotionCode":    "probe-promotion",
		"UsePromotionCode": true,
	} {
		if got := body[key]; got != want {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}
	}

	var primaryVSwitchIDs []string
	decodeJSONField(t, body, "VSwitchIds", &primaryVSwitchIDs)
	if want := []string{"vsw-primary"}; !reflect.DeepEqual(primaryVSwitchIDs, want) {
		t.Fatalf("VSwitchIds = %#v, want %#v", primaryVSwitchIDs, want)
	}

	var primaryResourceSpec struct {
		Cpu      int32 `json:"Cpu"`
		MemoryGB int32 `json:"MemoryGB"`
	}
	decodeJSONField(t, body, "ResourceSpec", &primaryResourceSpec)
	if primaryResourceSpec.Cpu != 0 || primaryResourceSpec.MemoryGB != 0 {
		t.Fatalf("ResourceSpec = %#v, want Cpu=0 MemoryGB=0", primaryResourceSpec)
	}

	var storage struct {
		Oss struct {
			Bucket string `json:"Bucket"`
		} `json:"Oss"`
	}
	decodeJSONField(t, body, "Storage", &storage)
	if storage.Oss.Bucket != "probe-bucket" {
		t.Fatalf("Storage.Oss.Bucket = %q, want %q", storage.Oss.Bucket, "probe-bucket")
	}

	if got := body["Ha"]; got != true {
		t.Fatalf("Ha = %#v, want true", got)
	}

	var haVSwitchIDs []string
	decodeJSONField(t, body, "HaVSwitchIds", &haVSwitchIDs)
	if want := []string{"vsw-standby"}; !reflect.DeepEqual(haVSwitchIDs, want) {
		t.Fatalf("HaVSwitchIds = %#v, want %#v", haVSwitchIDs, want)
	}

	var haResourceSpec struct {
		Cpu      int32 `json:"Cpu"`
		MemoryGB int32 `json:"MemoryGB"`
	}
	decodeJSONField(t, body, "HaResourceSpec", &haResourceSpec)
	if haResourceSpec.Cpu != 2 || haResourceSpec.MemoryGB != 8 {
		t.Fatalf("HaResourceSpec = %#v, want Cpu=2 MemoryGB=8", haResourceSpec)
	}
}

func TestBuildCreateInstanceBodyRejectsNilWorkspace(t *testing.T) {
	if _, err := BuildCreateInstanceBody(nil, CreateOptions{}); err == nil {
		t.Fatal("BuildCreateInstanceBody returned no error")
	}
}

func TestInstanceIDFromCreateResponse(t *testing.T) {
	response := map[string]interface{}{
		"OrderInfo": map[string]interface{}{
			"InstanceId": "f-cn-probe",
		},
	}

	instanceID, err := InstanceIDFromCreateResponse(response)
	if err != nil {
		t.Fatalf("InstanceIDFromCreateResponse returned an error: %v", err)
	}
	if instanceID != "f-cn-probe" {
		t.Fatalf("instance ID = %q, want %q", instanceID, "f-cn-probe")
	}
}

func TestInstanceIDFromCreateResponseRejectsMalformedResponses(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
	}{
		{name: "missing order info", response: map[string]interface{}{}},
		{name: "missing instance ID", response: map[string]interface{}{
			"OrderInfo": map[string]interface{}{},
		}},
		{name: "empty instance ID", response: map[string]interface{}{
			"OrderInfo": map[string]interface{}{"InstanceId": ""},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := InstanceIDFromCreateResponse(test.response); err == nil {
				t.Fatal("InstanceIDFromCreateResponse returned no error")
			}
		})
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

	config, ok := HAConfig(workspace)
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

	config, ok := HAConfig(workspace)
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
		if config, ok := HAConfig(workspace); ok || config != nil {
			t.Fatalf("HAConfig(%#v) = (%#v, %t), want (nil, false)", workspace, config, ok)
		}
	}
}

func decodeJSONField(t *testing.T, body map[string]interface{}, key string, target interface{}) {
	t.Helper()

	raw, ok := body[key].(string)
	if !ok {
		t.Fatalf("%s = %#v, want a JSON string", key, body[key])
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
}
