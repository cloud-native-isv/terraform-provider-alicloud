package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type FlinkService struct {
	client         *connectivity.AliyunClient
	aliyunFlinkAPI *aliyunFlinkAPI.FlinkAPI
}

// NewFlinkService creates a new FlinkService using cws-lib-go implementation
func NewFlinkService(client *connectivity.AliyunClient) (*FlinkService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go FlinkAPI
	aliyunFlinkAPI, err := aliyunFlinkAPI.NewFlinkAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go FlinkAPI: %w", err)
	}

	return &FlinkService{
		client:         client,
		aliyunFlinkAPI: aliyunFlinkAPI,
	}, nil
}

// Zone methods
func (s *FlinkService) DescribeSupportedZones() ([]*aliyunFlinkAPI.ZoneInfo, error) {
	return s.aliyunFlinkAPI.ListSupportedZones(s.client.RegionId)
}

// Engine methods
func (s *FlinkService) ListEngines(workspaceId string) ([]*aliyunFlinkAPI.FlinkEngine, error) {
	return s.aliyunFlinkAPI.ListEngines(workspaceId)
}

// Helper functions for parsing IDs
func parseDeploymentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid deployment ID format, expected namespace:deploymentId, got %s", id)
	}
	return parts[0], parts[1], nil
}

func parseJobId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid job ID format, expected namespace:jobId, got %s", id)
	}
	return parts[0], parts[1], nil
}
