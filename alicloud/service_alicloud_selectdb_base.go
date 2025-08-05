package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
)

type SelectDBService struct {
	client      *connectivity.AliyunClient
	selectdbAPI *selectdb.SelectDBAPI
}

// NewSelectDBService creates a new SelectDBService using cws-lib-go implementation
func NewSelectDBService(client *connectivity.AliyunClient) (*SelectDBService, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go SelectDBAPI
	selectdbAPI, err := selectdb.NewSelectDBAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go SelectDBAPI: %w", err)
	}

	return &SelectDBService{
		client:      client,
		selectdbAPI: selectdbAPI,
	}, nil
}

// GetAPI returns the SelectDBAPI instance for direct API access
func (service *SelectDBService) GetAPI() *selectdb.SelectDBAPI {
	// add some customize logic for this API object
	return service.selectdbAPI
}

// GetRegionId returns the region ID from the client
func (service *SelectDBService) GetRegionId() string {
	return service.client.RegionId
}

// EncodeSelectDBClusterId encodes instance ID and cluster ID into a single ID string
// Format: instanceId:clusterId
func (s *SelectDBService) EncodeSelectDBClusterId(instanceId, clusterId string) string {
	return fmt.Sprintf("%s:%s", instanceId, clusterId)
}

// DecodeSelectDBClusterId parses the composite ID into instanceId and clusterId
func (s *SelectDBService) DecodeSelectDBClusterId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid SelectDB cluster ID format, expected instanceId:clusterId, got %s", id)
	}
	return parts[0], parts[1], nil
}

// Instance Management Operations

// DescribeSelectDBInstance retrieves information about a SelectDB instance
func (s *SelectDBService) DescribeSelectDBInstance(instanceId string) (*selectdb.Instance, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	instance, err := s.GetAPI().GetInstance(instanceId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return instance, nil
}

// DescribeSelectDBInstances retrieves list of SelectDB instances
func (s *SelectDBService) DescribeSelectDBInstances(pageNumber, pageSize int32) ([]selectdb.Instance, error) {
	instances, err := s.GetAPI().ListInstances(s.client.RegionId, pageNumber, pageSize)
	if err != nil {
		return nil, WrapError(err)
	}

	return instances, nil
}

// DescribeSelectDBInstanceClasses retrieves list of available SelectDB instance classes
func (s *SelectDBService) DescribeSelectDBInstanceClasses() ([]selectdb.InstanceClass, error) {
	instanceClasses, err := s.GetAPI().ListInstanceClass(s.client.RegionId)
	if err != nil {
		return nil, WrapError(err)
	}

	return instanceClasses, nil
}
