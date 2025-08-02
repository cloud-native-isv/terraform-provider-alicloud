package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

type ArmsService struct {
	client  *connectivity.AliyunClient
	armsAPI *aliyunArmsAPI.ArmsAPI
}

// NewArmsService creates a new ArmsService using cws-lib-go implementation
func NewArmsService(client *connectivity.AliyunClient) (*ArmsService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go ArmsAPI
	armsAPI, err := aliyunArmsAPI.NewARMSClientWithCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go ArmsAPI: %w", err)
	}

	return &ArmsService{
		client:  client,
		armsAPI: armsAPI,
	}, nil
}

// GetAPI returns the ArmsAPI instance for direct API access
func (service *ArmsService) GetAPI() *aliyunArmsAPI.ArmsAPI {
	// add some customize logic for this API object
	return service.armsAPI
}
