package alicloud

import (
	"testing"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

func TestFlinkDeploymentTargetUsesV2OnlyWithRequest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target *flink.DeploymentTarget
		want   bool
	}{
		{name: "no quota", target: &flink.DeploymentTarget{}, want: false},
		{name: "legacy limit only", target: &flink.DeploymentTarget{Quota: &flink.ResourceQuota{Limit: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}}}, want: false},
		{name: "request and limit", target: &flink.DeploymentTarget{Quota: &flink.ResourceQuota{Request: &flink.ResourceSpec{Cpu: 1, MemoryGB: 4}, Limit: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}}}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := flinkDeploymentTargetUsesV2(tc.target); got != tc.want {
				t.Fatalf("flinkDeploymentTargetUsesV2() = %t, want %t", got, tc.want)
			}
		})
	}
}
