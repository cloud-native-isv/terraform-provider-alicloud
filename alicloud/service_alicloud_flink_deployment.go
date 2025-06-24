package alicloud

import (
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Deployment methods
func (s *FlinkService) GetDeployment(id string) (*aliyunFlinkAPI.Deployment, error) {
	// Parse deployment ID to extract namespace and deployment ID
	// Format: namespace:deploymentId
	namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}
	return s.aliyunFlinkAPI.GetDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(namespaceName *string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	deployment.Namespace = *namespaceName
	return s.aliyunFlinkAPI.CreateDeployment(deployment)
}

func (s *FlinkService) UpdateDeployment(deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	return s.aliyunFlinkAPI.UpdateDeployment(deployment)
}

func (s *FlinkService) DeleteDeployment(namespaceName, deploymentId string) error {
	return s.aliyunFlinkAPI.DeleteDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) ListDeployments(namespaceName string) ([]aliyunFlinkAPI.Deployment, error) {
	return s.aliyunFlinkAPI.ListDeployments(namespaceName)
}

func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := s.GetDeployment(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
		}

		return deployment, deployment.Status, nil
	}
}

// Deployment Draft methods
func (s *FlinkService) CreateDeploymentDraft(workspaceID string, namespaceName string, draft *aliyunFlinkAPI.DeploymentDraft) (*aliyunFlinkAPI.DeploymentDraft, error) {
	// Call the underlying API with the proper signature
	result, err := s.aliyunFlinkAPI.CreateDeploymentDraft(workspaceID, namespaceName, draft)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_draft", "CreateDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

func (s *FlinkService) GetDeploymentDraft(workspaceId string, namespaceName string, draftId string) (*aliyunFlinkAPI.DeploymentDraft, error) {
	return s.aliyunFlinkAPI.GetDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) UpdateDeploymentDraft(workspaceId string, namespaceName string, draft *aliyunFlinkAPI.DeploymentDraft) (*aliyunFlinkAPI.DeploymentDraft, error) {
	return s.aliyunFlinkAPI.UpdateDeploymentDraft(workspaceId, namespaceName, draft)
}

func (s *FlinkService) DeleteDeploymentDraft(workspaceId string, namespaceName string, draftId string) error {
	return s.aliyunFlinkAPI.DeleteDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) ListDeploymentDrafts(workspaceId, namespaceName string) ([]aliyunFlinkAPI.DeploymentDraft, error) {
	return s.aliyunFlinkAPI.ListDeploymentDrafts(workspaceId, namespaceName)
}

// FlinkDeploymentDraftStateRefreshFunc provides state refresh for deployment drafts
func (s *FlinkService) FlinkDeploymentDraftStateRefreshFunc(workspaceId string, namespaceName string, draftId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		draft, err := s.GetDeploymentDraft(workspaceId, namespaceName, draftId)
		if err != nil {
			if NotFoundError(err) {
				// Draft not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For deployment drafts, if we can get it successfully, it means it's available
		status := "Available"
		if draft.Status != "" {
			status = draft.Status
		}

		for _, failState := range failStates {
			// Check if draft is in a failed state (if any fail states are defined)
			if status == failState {
				return draft, status, WrapError(Error(FailedToReachTargetStatus, status))
			}
		}

		return draft, status, nil
	}
}
