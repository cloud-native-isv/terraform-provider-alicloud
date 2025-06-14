package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

type SlsService struct {
	client       *connectivity.AliyunClient
	aliyunSlsAPI *aliyunSlsAPI.SlsAPI
}

// NewSlsService creates a new SlsService instance with initialized clients
func NewSlsService(client *connectivity.AliyunClient) (*SlsService, error) {
	credentials := &aliyunSlsAPI.SlsCredentials{
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

// NewSlsServiceV2 creates a new SlsService instance (alias for NewSlsService for backward compatibility)
func NewSlsServiceV2(client *connectivity.AliyunClient) (*SlsService, error) {
	return NewSlsService(client)
}
