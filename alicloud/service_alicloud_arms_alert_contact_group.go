package alicloud

import (
	"strconv"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
)

// =============================================================================
// Alert Contact Group Functions
// =============================================================================

// DescribeArmsAlertContactGroup describes ARMS alert contact group by group ID
func (s *ArmsService) DescribeArmsAlertContactGroup(contactGroupId string) (*aliyunArmsAPI.AlertContactGroup, error) {
	id, err := strconv.ParseInt(contactGroupId, 10, 64)
	if err != nil {
		return nil, WrapErrorf(err, "Invalid contact group ID: %s", contactGroupId)
	}

	// Use API to get contact groups with details and find the specific one
	groups, err := s.armsAPI.ListAlertContactGroups(1, 100)
	if err != nil {
		return nil, WrapError(err)
	}

	for _, group := range groups {
		if group.ContactGroupId == id {
			return group, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("ARMS Alert Contact Group", contactGroupId), NotFoundMsg, AlibabaCloudSdkGoERROR)
}

// DescribeArmsAlertContactGroups describes ARMS alert contact groups with filters
func (s *ArmsService) DescribeArmsAlertContactGroups() ([]*aliyunArmsAPI.AlertContactGroup, error) {
	return s.armsAPI.ListAllAlertContactGroups()
}

// DescribeArmsAlertContactGroupsByIds describes ARMS alert contact groups by group IDs
func (s *ArmsService) DescribeArmsAlertContactGroupsByIds(contactGroupIds []string, isDetail bool) ([]*aliyunArmsAPI.AlertContactGroup, error) {
	var result []*aliyunArmsAPI.AlertContactGroup

	for _, contactGroupId := range contactGroupIds {
		group, err := s.DescribeArmsAlertContactGroup(contactGroupId)
		if err != nil {
			if NotFoundError(err) {
				continue // Skip not found groups
			}
			return nil, err
		}
		result = append(result, group)
	}

	return result, nil
}
