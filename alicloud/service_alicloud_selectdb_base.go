package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSelectDBAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
)

type SelectDBService struct {
	client      *connectivity.AliyunClient
	selectdbAPI *aliyunSelectDBAPI.SelectDBAPI
}

// NewSelectDBService creates a new SelectDBService instance with initialized clients
func NewSelectDBService(client *connectivity.AliyunClient) (*SelectDBService, error) {
	service := &SelectDBService{
		client: client,
	}

	// Initialize the SelectDB API client lazily
	_, err := service.GetSelectDBAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SelectDB API client: %w", err)
	}

	return service, nil
}
