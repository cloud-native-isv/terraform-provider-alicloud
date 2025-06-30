package alicloud

import (
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Member methods
func (s *FlinkService) CreateMember(workspaceId string, namespaceName string, member *flinkAPI.Member) (*flinkAPI.Member, error) {
	// Set workspace and namespace
	member.WorkspaceId = workspaceId
	member.NamespaceName = namespaceName

	// Call the underlying API
	return s.flinkAPI.CreateMember(member)
}

func (s *FlinkService) GetMember(workspaceId string, namespaceName string, memberId string) (*flinkAPI.Member, error) {
	// Call the underlying API
	return s.flinkAPI.GetMember(workspaceId, namespaceName, memberId)
}

func (s *FlinkService) UpdateMember(workspaceId string, namespaceName string, member *flinkAPI.Member) (*flinkAPI.Member, error) {
	// Set workspace and namespace
	member.WorkspaceId = workspaceId
	member.NamespaceName = namespaceName

	// Call the underlying API
	return s.flinkAPI.UpdateMember(member)
}

func (s *FlinkService) DeleteMember(workspaceId string, namespaceName string, memberId string) error {
	// Call the underlying API
	return s.flinkAPI.DeleteMember(workspaceId, namespaceName, memberId)
}

func (s *FlinkService) ListMembers(workspaceId, namespaceName string) ([]flinkAPI.Member, error) {
	return s.flinkAPI.ListMembers(workspaceId, namespaceName)
}

// FlinkMemberStateRefreshFunc provides state refresh for members
func (s *FlinkService) FlinkMemberStateRefreshFunc(workspaceId string, namespaceName string, memberId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		member, err := s.GetMember(workspaceId, namespaceName, memberId)
		if err != nil {
			if NotFoundError(err) {
				// Member not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For members, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if member is in a failed state (if any fail states are defined)
			if member.Role == failState {
				return member, member.Role, WrapError(Error(FailedToReachTargetStatus, member.Role))
			}
		}

		return member, "Available", nil
	}
}
