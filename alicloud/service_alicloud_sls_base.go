package alicloud

import (
	"fmt"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

type SlsService struct {
	client       *connectivity.AliyunClient
	aliyunSlsAPI *aliyunSlsAPI.SlsAPI
}

// NewSlsService creates a new SlsService instance with initialized clients
func NewSlsService(client *connectivity.AliyunClient) (*SlsService, error) {
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	slsAPI, err := aliyunSlsAPI.NewSlsAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS API client: %w", err)
	}

	return &SlsService{
		client:       client,
		aliyunSlsAPI: slsAPI,
	}, nil
}


// getSlsAPI creates and returns an SLS API client using CWS-Lib-Go
func (s *SlsService) getSlsAPI() (*aliyunSlsAPI.SlsAPI, error) {
	// Create credentials from the AliyunClient
	credentials := &common.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}

	// Create SLS API client
	slsAPI, err := aliyunSlsAPI.NewSlsAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS API client: %w", err)
	}

	return slsAPI, nil
}

// Only set dataRedundancyType for regions that support it
var SupportsDataRedundancyRegions = []string{
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-southeast-1",
	// "ap-southeast-3",
	"ap-southeast-5",
	"ap-southeast-6",
	"ap-southeast-7",
	"cn-beijing",
	"cn-chengdu",
	// "cn-fuzhou",
	"cn-guangzhou",
	"cn-hangzhou",
	// "cn-heyuan",
	// "cn-hongkong",
	"cn-huhehaote",
	"cn-nanjing",
	"cn-qingdao",
	// "cn-shanghai-finance-1",
	"cn-shanghai",
	"cn-shenzhen",
	// "cn-wulanchabu-acdr-1",
	"cn-wulanchabu",
	"cn-zhangjiakou",
	// "eu-central-1",
	// "eu-west-1",
	// "us-east-1",
	// "us-west-1",
}
