package alicloud

import (
	"fmt"
	"strings"

	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *FlinkService) CreateDeploymentDraft(workspaceID string, namespaceName string, draft *flinkAPI.DeploymentDraft) (*flinkAPI.DeploymentDraft, error) {
	result, err := s.flinkAPI.CreateDeploymentDraft(workspaceID, namespaceName, draft)
	if err == nil && result != nil {
		addDebugJson("CreateDeploymentDraft", result)
	}
	return result, err
}

func (s *FlinkService) GetDeploymentDraft(workspaceId string, namespaceName string, draftId string) (*flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.GetDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) UpdateDeploymentDraft(workspaceId string, namespaceName string, draft *flinkAPI.DeploymentDraft) (*flinkAPI.DeploymentDraft, error) {
	result, err := s.flinkAPI.UpdateDeploymentDraft(workspaceId, namespaceName, draft)
	if err == nil && result != nil {
		addDebugJson("UpdateDeploymentDraft", result)
	}
	return result, err
}

func (s *FlinkService) DeleteDeploymentDraft(workspaceId string, namespaceName string, draftId string) error {
	err := s.flinkAPI.DeleteDeploymentDraft(workspaceId, namespaceName, draftId)
	if err == nil {
		addDebugJson("DeleteDeploymentDraft", fmt.Sprintf("Draft %s deleted successfully", draftId))
	}
	return err
}

func (s *FlinkService) ListDeploymentDrafts(workspaceId, namespaceName string) ([]flinkAPI.DeploymentDraft, error) {
	return s.flinkAPI.ListDeploymentDrafts(workspaceId, namespaceName)
}

func (s *FlinkService) FlinkDeploymentDraftStateRefreshFunc(workspaceId string, namespaceName string, draftId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		draft, err := s.GetDeploymentDraft(workspaceId, namespaceName, draftId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "NotFound", nil
			}
			return nil, "", WrapError(err)
		}

		if draft == nil {
			return nil, "NotFound", nil
		}

		status := "Available"

		for _, failState := range failStates {
			if status == failState {
				return draft, status, WrapErrorf(err, DefaultErrorMsg, draftId, "GetDeploymentDraft", AlibabaCloudSdkGoERROR)
			}
		}

		return draft, status, nil
	}
}

func parseDeploymentDraftId(id string) (workspaceId, namespaceName, draftId string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid deployment draft ID format, expected workspaceId:namespaceName:draftId, got: %s", id))
		return
	}

	workspaceId = parts[0]
	namespaceName = parts[1]
	draftId = parts[2]

	if workspaceId == "" {
		err = WrapError(fmt.Errorf("workspaceId cannot be empty in deployment draft ID: %s", id))
		return
	}

	if namespaceName == "" {
		err = WrapError(fmt.Errorf("namespaceName cannot be empty in deployment draft ID: %s", id))
		return
	}

	if draftId == "" {
		err = WrapError(fmt.Errorf("draftId cannot be empty in deployment draft ID: %s", id))
		return
	}

	return
}

func formatDeploymentDraftId(workspaceId, namespaceName, draftId string) string {
	return fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, draftId)
}
