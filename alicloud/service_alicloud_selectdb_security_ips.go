package alicloud

import (
	"fmt"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Security IP List Management Operations for SelectDB Instance

// DescribeSelectDBSecurityIPList retrieves security IP list for a SelectDB instance
func (s *SelectDBService) DescribeSelectDBSecurityIPList(dbInstanceId string, regionId string, resourceOwnerId int64) ([]selectdb.SecurityIPGroup, error) {
	if dbInstanceId == "" {
		return nil, WrapError(fmt.Errorf("database instance ID cannot be empty"))
	}

	result, err := s.GetAPI().GetSecurityIPList(dbInstanceId, regionId, resourceOwnerId)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return result, nil
}

// ModifySelectDBSecurityIPList modifies security IP list for a SelectDB instance
func (s *SelectDBService) ModifySelectDBSecurityIPList(dbInstanceId string, securityIPList string, groupName string, modifyMode string, regionId string, resourceOwnerId int64) (*selectdb.SecurityIPGroup, error) {
	if dbInstanceId == "" {
		return nil, WrapError(fmt.Errorf("database instance ID cannot be empty"))
	}
	if securityIPList == "" {
		return nil, WrapError(fmt.Errorf("security IP list cannot be empty"))
	}

	result, err := s.GetAPI().ModifySecurityIPList(dbInstanceId, securityIPList, groupName, modifyMode, regionId, resourceOwnerId)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBSecurityIPGroup retrieves a specific security IP group
func (s *SelectDBService) DescribeSelectDBSecurityIPGroup(instanceId, groupName string) (*selectdb.SecurityIPGroup, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if groupName == "" {
		return nil, WrapError(fmt.Errorf("group name cannot be empty"))
	}

	regionId := s.GetRegionId()
	result, err := s.GetAPI().GetSecurityIPList(instanceId, regionId, 0)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Find the specific group by name
	for _, group := range result {
		if group.GroupName == groupName {
			return &group, nil
		}
	}

	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Security IP Group", groupName)), NotFoundMsg, ProviderERROR)
}

// SelectDBSecurityIPListStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB security IP list
func (s *SelectDBService) SelectDBSecurityIPListStateRefreshFunc(instanceId, groupName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		group, err := s.DescribeSelectDBSecurityIPGroup(instanceId, groupName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", WrapErrorf(Error(GetNotFoundMessage("SelectDB Security IP Group", groupName)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		// Security IP groups don't have explicit status, so we use a simple Available/NotAvailable
		status := "Available"
		if len(group.SecurityIPList) == 0 {
			status = "NotAvailable"
		}

		for _, failState := range failStates {
			if status == failState {
				return group, status, WrapError(Error(FailedToReachTargetStatus, status))
			}
		}

		return group, status, nil
	}
}
