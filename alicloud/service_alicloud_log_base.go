package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

type SlsService struct {
	client *connectivity.AliyunClient
	slsAPI *aliyunSlsAPI.SlsAPI
}

// NewSlsService creates a new SlsService using cws-lib-go implementation
func NewSlsService(client *connectivity.AliyunClient) (*SlsService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go SlsAPI
	slsAPI, err := aliyunSlsAPI.NewSlsAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go SlsAPI: %w", err)
	}

	return &SlsService{
		client: client,
		slsAPI: slsAPI,
	}, nil
}

// GetAPI returns the SlsAPI instance for direct API access
func (service *SlsService) GetAPI() *aliyunSlsAPI.SlsAPI {
	// add some customize logic for this API object
	return service.slsAPI
}
