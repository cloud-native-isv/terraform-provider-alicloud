package flinkworkspace

import (
	"strings"
	"testing"
)

func TestValidateVSwitchTopology(t *testing.T) {
	primary := []VSwitch{
		{ID: "vsw-a1", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-a"},
		{ID: "vsw-a2", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-a"},
	}
	standby := []VSwitch{
		{ID: "vsw-b1", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-b"},
		{ID: "vsw-b2", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-b"},
	}

	got, err := ValidateVSwitchTopology("cn-test", "vpc-1", "cn-test-a", "cn-test-b", primary, standby)
	if err != nil {
		t.Fatal(err)
	}
	if got.PrimaryZoneID != "cn-test-a" || got.StandbyZoneID != "cn-test-b" {
		t.Fatalf("topology = %#v", got)
	}
}

func TestValidateVSwitchTopologyRejectsInvalidNetworks(t *testing.T) {
	basePrimary := []VSwitch{{ID: "vsw-a", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-a"}}
	baseStandby := []VSwitch{{ID: "vsw-b", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-b"}}
	tests := []struct {
		name       string
		region     string
		vpc        string
		legacyZone string
		legacyHA   string
		primary    []VSwitch
		standby    []VSwitch
		want       string
	}{
		{name: "empty primary", region: "cn-test", vpc: "vpc-1", standby: baseStandby, want: "primary"},
		{name: "primary spans zones", region: "cn-test", vpc: "vpc-1", primary: appendCopy(basePrimary, VSwitch{ID: "vsw-c", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-c"}), standby: baseStandby, want: "same zone"},
		{name: "standby spans zones", region: "cn-test", vpc: "vpc-1", primary: basePrimary, standby: appendCopy(baseStandby, VSwitch{ID: "vsw-c", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-c"}), want: "same zone"},
		{name: "zones equal", region: "cn-test", vpc: "vpc-1", primary: basePrimary, standby: []VSwitch{{ID: "vsw-a2", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-a"}}, want: "different zones"},
		{name: "wrong vpc", region: "cn-test", vpc: "vpc-1", primary: basePrimary, standby: []VSwitch{{ID: "vsw-b", RegionID: "cn-test", VPCID: "vpc-2", ZoneID: "cn-test-b"}}, want: "VPC"},
		{name: "wrong region", region: "cn-test", vpc: "vpc-1", primary: basePrimary, standby: []VSwitch{{ID: "vsw-b", RegionID: "cn-other", VPCID: "vpc-1", ZoneID: "cn-test-b"}}, want: "region"},
		{name: "legacy primary mismatch", region: "cn-test", vpc: "vpc-1", legacyZone: "cn-test-c", primary: basePrimary, standby: baseStandby, want: "legacy primary"},
		{name: "legacy standby mismatch", region: "cn-test", vpc: "vpc-1", legacyHA: "cn-test-c", primary: basePrimary, standby: baseStandby, want: "legacy standby"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateVSwitchTopology(tc.region, tc.vpc, tc.legacyZone, tc.legacyHA, tc.primary, tc.standby)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestValidateVSwitchTopologyAllowsSingleZoneWithoutStandby(t *testing.T) {
	got, err := ValidateVSwitchTopology("cn-test", "vpc-1", "", "", []VSwitch{{ID: "vsw-a", RegionID: "cn-test", VPCID: "vpc-1", ZoneID: "cn-test-a"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.PrimaryZoneID != "cn-test-a" || got.StandbyZoneID != "" {
		t.Fatalf("topology = %#v", got)
	}
}

func appendCopy(input []VSwitch, values ...VSwitch) []VSwitch {
	result := append([]VSwitch(nil), input...)
	return append(result, values...)
}
