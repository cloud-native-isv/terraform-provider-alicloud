package alicloud

import (
	"fmt"
	"strings"

	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeNasAccessGroup gets NAS access group information
func (s *NasService) DescribeNasAccessGroup(id string) (accessGroup *aliyunNasAPI.AccessGroup, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	accessGroupName := parts[0]
	fileSystemType := parts[1]

	// Use getNasAPI() method to get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	// List access groups and find the specific one
	accessGroups, err := nasAPI.ListAccessGroups()
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil, WrapErrorf(NotFoundErr("AccessGroup", id), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "ListAccessGroups", AlibabaCloudSdkGoERROR)
	}

	// Find the specific access group by name and file system type
	for _, ag := range accessGroups {
		if ag.AccessGroupName == accessGroupName && ag.FileSystemType == fileSystemType {
			return &ag, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("AccessGroup", id), NotFoundMsg, ProviderERROR)
}

// NasAccessGroupStateRefreshFunc returns a StateRefreshFunc for NAS access group status
func (s *NasService) NasAccessGroupStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accessGroup, err := s.DescribeNasAccessGroup(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Get the field value using direct field access
		var currentStatus string
		switch field {
		case "$.AccessGroupName":
			currentStatus = accessGroup.AccessGroupName
		case "$.AccessGroupType":
			currentStatus = accessGroup.AccessGroupType
		case "$.Description":
			currentStatus = accessGroup.Description
		case "$.CreateTime":
			currentStatus = accessGroup.CreateTime
		case "$.ModifyTime":
			currentStatus = accessGroup.ModifyTime
		case "$.FileSystemType":
			currentStatus = accessGroup.FileSystemType
		case "$.RuleCount":
			currentStatus = fmt.Sprintf("%d", accessGroup.RuleCount)
		case "$.MountTargetCount":
			currentStatus = fmt.Sprintf("%d", accessGroup.MountTargetCount)
		default:
			currentStatus = fmt.Sprintf("%v", accessGroup)
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return accessGroup, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return accessGroup, currentStatus, nil
	}
}

// CreateNasAccessGroup creates a new NAS access group
func (s *NasService) CreateNasAccessGroup(request *aliyunNasAPI.AccessGroup) (*aliyunNasAPI.AccessGroup, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_group", "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	accessGroup, err := nasAPI.CreateAccessGroup(request)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_group", "CreateAccessGroup", AlibabaCloudSdkGoERROR)
	}

	return accessGroup, nil
}

// UpdateNasAccessGroup updates an existing NAS access group
func (s *NasService) UpdateNasAccessGroup(accessGroupName string, request *aliyunNasAPI.AccessGroup) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessGroupName, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.ModifyAccessGroup(accessGroupName, request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessGroupName, "ModifyAccessGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNasAccessGroup deletes a NAS access group
func (s *NasService) DeleteNasAccessGroup(accessGroupName string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessGroupName, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DeleteAccessGroup(accessGroupName)
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, accessGroupName, "DeleteAccessGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// CreateNasAccessPoint creates a new NAS access point
func (s *NasService) CreateNasAccessPoint(fileSystemId, accessPointName, accessGroup, rootPath string, enabledRam bool, vpcId, vSwitchId string, ownerUid, ownerGid int64, permission string, posixUser *aliyunNasAPI.PosixUser) (*aliyunNasAPI.AccessPoint, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_point", "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	accessPoint, err := nasAPI.CreateAccessPoint(fileSystemId, accessPointName, accessGroup, rootPath, enabledRam, vpcId, vSwitchId, ownerUid, ownerGid, permission, posixUser)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_point", "CreateAccessPoint", AlibabaCloudSdkGoERROR)
	}

	return accessPoint, nil
}

// UpdateNasAccessPoint updates an existing NAS access point
func (s *NasService) UpdateNasAccessPoint(fileSystemId, accessPointId, accessPointName, accessGroup string, enabledRam bool) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessPointId, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.ModifyAccessPoint(fileSystemId, accessPointId, accessPointName, accessGroup, enabledRam)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessPointId, "ModifyAccessPoint", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNasAccessPoint deletes a NAS access point
func (s *NasService) DeleteNasAccessPoint(fileSystemId, accessPointId string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessPointId, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DeleteAccessPoint(fileSystemId, accessPointId)
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, accessPointId, "DeleteAccessPoint", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// NasAccessPointStateRefreshFunc returns a StateRefreshFunc for NAS access point status
func (s *NasService) NasAccessPointStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return nil, "", WrapError(fmt.Errorf("invalid resource ID format"))
		}
		fileSystemId := parts[0]
		accessPointId := parts[1]

		accessPoint, err := s.DescribeNasAccessPoint(fileSystemId, accessPointId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentStatus := accessPoint.Status
		for _, failState := range failStates {
			if currentStatus == failState {
				return accessPoint, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return accessPoint, currentStatus, nil
	}
}

// DescribeNasAccessPoint gets NAS access point information
func (s *NasService) DescribeNasAccessPoint(fileSystemId, accessPointId string) (accessPoint *aliyunNasAPI.AccessPoint, err error) {
	// Use getNasAPI() method to get NAS API client
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, accessPointId, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	// Get the specific access point
	accessPoint, err = nasAPI.GetAccessPoint(fileSystemId, accessPointId)
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil, WrapErrorf(NotFoundErr("AccessPoint", accessPointId), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, accessPointId, "GetAccessPoint", AlibabaCloudSdkGoERROR)
	}

	return accessPoint, nil
}
