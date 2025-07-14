package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

type ArmsService struct {
	client  *connectivity.AliyunClient
	armsAPI *aliyunArmsAPI.ArmsAPI
}

// NewArmsService creates a new ArmsService instance
func NewArmsService(client *connectivity.AliyunClient) *ArmsService {
	// Initialize the ARMS API client
	armsAPI, err := aliyunArmsAPI.NewARMSClientWithCredentials(&common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	})
	if err != nil {
		// Log error but continue with nil API (will fall back to direct RPC calls)
		addDebug("NewArmsService", "Failed to initialize ARMS API client", err)
	}

	return &ArmsService{
		client:  client,
		armsAPI: armsAPI,
	}
}

// Note: All specific functionality has been moved to dedicated service files:
// - service_alicloud_arms_alert.go: Alert contacts, contact groups, robots, schedules
// - service_alicloud_arms_rule.go: Alert rules and notification policies
// - service_alicloud_arms_dispatch.go: Dispatch rules
// - service_alicloud_arms_prometheus.go: Prometheus monitoring and alert rules
// - service_alicloud_arms_integration.go: Integration management
// - service_alicloud_arms_environment.go: Environment, features, monitors
// - service_alicloud_arms_grafana.go: Grafana workspace management
// - service_alicloud_arms_synthetic.go: Synthetic tasks and silence policies
