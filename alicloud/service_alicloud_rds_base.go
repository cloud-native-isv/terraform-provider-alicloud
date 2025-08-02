package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunRdsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/rds"
)

// RdsService provides RDS related operations
type RdsService struct {
	client *connectivity.AliyunClient
	rdsAPI *aliyunRdsAPI.RDSAPI
}

// NewRdsService creates a new RdsService using cws-lib-go implementation
func NewRdsService(client *connectivity.AliyunClient) (*RdsService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go RDSAPI
	rdsAPI, err := aliyunRdsAPI.NewRDSAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go RDSAPI: %w", err)
	}

	return &RdsService{
		client: client,
		rdsAPI: rdsAPI,
	}, nil
}

// GetAPI returns the RDSAPI instance for direct API access
func (service *RdsService) GetAPI() *aliyunRdsAPI.RDSAPI {
	// add some customize logic for this API object
	return service.rdsAPI
}

//	_______________                      _______________                       _______________
//	|              | ______param______\  |              |  _____request_____\  |              |
//	|   Business   |                     |    Service   |                      |    SDK/API   |
//	|              | __________________  |              |  __________________  |              |
//	|______________| \    (obj, err)     |______________|  \ (status, cont)    |______________|
//	                    |                                    |
//	                    |A. {instance, nil}                  |a. {200, content}
//	                    |B. {nil, error}                     |b. {200, nil}
//	               					  |c. {4xx, nil}
//
// The API return 200 for resource not found.
// When getInstance is empty, then throw InstanceNotfound error.
// That the business layer only need to check error.
var DBInstanceStatusCatcher = Catcher{"OperationDenied.DBInstanceStatus", 60, 5}
