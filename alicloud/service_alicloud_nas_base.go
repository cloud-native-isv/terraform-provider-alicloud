package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
)

type NasService struct {
	client       *connectivity.AliyunClient
	aliyunNasAPI *aliyunNasAPI.NasAPI
}

// NewNasService creates a new NasService instance with initialized clients
func NewNasService(client *connectivity.AliyunClient) (*NasService, error) {
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	nasAPI, err := aliyunNasAPI.NewNasAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create NAS API client: %w", err)
	}

	return &NasService{
		client:       client,
		aliyunNasAPI: nasAPI,
	}, nil
}

func (s *NasService) getNasAPI() (*aliyunNasAPI.NasAPI, error) {
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	nasAPI, err := aliyunNasAPI.NewNasAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create NAS API client: %w", err)
	}

	return nasAPI, nil
}
