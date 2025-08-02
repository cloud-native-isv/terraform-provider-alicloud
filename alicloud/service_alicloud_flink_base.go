package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type FlinkService struct {
	client   *connectivity.AliyunClient
	flinkAPI *aliyunFlinkAPI.FlinkAPI
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
	flinkAPI, err := aliyunFlinkAPI.NewFlinkAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go FlinkAPI: %w", err)
	}

	return &FlinkService{
		client:   client,
		flinkAPI: flinkAPI,
	}, nil
}

// GetAPI returns the FlinkAPI instance for direct API access
func (service *FlinkService) GetAPI() *aliyunFlinkAPI.FlinkAPI {
	// add some customize logic for this API object
	return service.flinkAPI
}

// Zone methods
func (s *FlinkService) DescribeSupportedZones() ([]*aliyunFlinkAPI.ZoneInfo, error) {
	return s.flinkAPI.ListSupportedZones(s.client.RegionId)
}

// Engine methods
func (s *FlinkService) ListEngines(workspaceId string) ([]*aliyunFlinkAPI.FlinkEngine, error) {
	return s.flinkAPI.ListEngines(workspaceId)
}
