package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
)

type NasService struct {
	client *connectivity.AliyunClient
	nasAPI *aliyunNasAPI.NasAPI
}

// NewNasService creates a new NasService using cws-lib-go implementation
func NewNasService(client *connectivity.AliyunClient) (*NasService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go NasAPI
	nasAPI, err := aliyunNasAPI.NewNasAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go NasAPI: %w", err)
	}

	return &NasService{
		client: client,
		nasAPI: nasAPI,
	}, nil
}

// GetAPI returns the NasAPI instance for direct API access
func (service *NasService) GetAPI() *aliyunNasAPI.NasAPI {
	// add some customize logic for this API object
	return service.nasAPI
}
