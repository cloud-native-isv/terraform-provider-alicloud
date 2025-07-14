package alicloud


import (
	"fmt"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
  "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
)

type SelectDBService struct {
	client       *connectivity.AliyunClient
	selectdbAPI *selectdb.selectDBAPI
}

// NewSlsService creates a new SlsService instance with initialized clients
func NewSelectDBService(client *connectivity.AliyunClient) (*SelectDBService, error) {
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

// SelectDBService is the service struct for SelectDB operations
type SelectDBService struct {
	api    *selectdb.SelectDBService
}

// NewSelectDBService creates a new SelectDB service instance
func (s *SelectDBService) NewSelectDBService(client *connectivity.AliyunClient) (*SelectDBService, error) {
	selectdbClient, err := client.NewSelectDBClient()
	if err != nil {
		return nil, WrapError(err)
	}

	return &SelectDBService{
		client: client,
		api:    selectdb.NewSelectDBService(selectdbClient),
	}, nil
}