package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunTablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
)

type OtsService struct {
	client        *connectivity.AliyunClient
	tablestoreAPI *aliyunTablestoreAPI.TablestoreAPI
}

// NewOtsService creates a new OtsService using cws-lib-go implementation
func NewOtsService(client *connectivity.AliyunClient) (*OtsService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go TablestoreAPI
	tablestoreAPI, err := aliyunTablestoreAPI.NewTablestoreAPI(&aliyunCommonAPI.ConnectionConfig{
		Credentials: credentials,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go TablestoreAPI: %w", err)
	}

	return &OtsService{
		client:        client,
		tablestoreAPI: tablestoreAPI,
	}, nil
}

// GetAPI returns the TablestoreAPI instance for direct API access
func (service *OtsService) GetAPI() *aliyunTablestoreAPI.TablestoreAPI {
	// add some customize logic for this API object
	return service.tablestoreAPI
}
