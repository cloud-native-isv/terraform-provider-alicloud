package alicloud

import (
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Deployment Draft methods
func (s *FlinkService) CreateDeploymentDraft(workspaceID string, namespaceName string, draft *flinkAPI.DeploymentDraft) (*flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.CreateDeploymentDraft(workspaceID, namespaceName, draft)
}

func (s *FlinkService) GetDeploymentDraft(workspaceId string, namespaceName string, draftId string) (*flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.GetDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) UpdateDeploymentDraft(workspaceId string, namespaceName string, draft *flinkAPI.DeploymentDraft) (*flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.UpdateDeploymentDraft(workspaceId, namespaceName, draft)
}

func (s *FlinkService) DeleteDeploymentDraft(workspaceId string, namespaceName string, draftId string) error {
	return s.flinkAPI.DeleteDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) ListDeploymentDrafts(workspaceId, namespaceName string) ([]flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.ListDeploymentDrafts(workspaceId, namespaceName)
}

// FlinkDeploymentDraftStateRefreshFunc provides state refresh for deployment drafts
func (s *FlinkService) FlinkDeploymentDraftStateRefreshFunc(workspaceId string, namespaceName string, draftId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		draft, err := s.GetDeploymentDraft(workspaceId, namespaceName, draftId)
		if err != nil {
			if NotFoundError(err) {
				// Draft not found - this is expected during deletion
				return nil, "Deleted", nil
			}
			return nil, "", WrapError(err)
		}

		// If draft is nil, it means the resource doesn't exist
		if draft == nil {
			return nil, "Deleted", nil
		}

		// For deployment drafts, if we can get it successfully, it means it's available
		// Since Status field is removed, we use a simple available state
		status := "Available"

		// Check if draft has required fields to determine if it's properly created
		if draft.Name == "" {
			status = "Creating"
		}

		// Check for fail states
		for _, failState := range failStates {
			if status == failState {
				return draft, status, WrapErrorf(err, DefaultErrorMsg, draftId, "GetDeploymentDraft", AlibabaCloudSdkGoERROR)
			}
		}

		return draft, status, nil
	}
}
