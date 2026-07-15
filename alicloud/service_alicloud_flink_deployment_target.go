package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type FlinkDeploymentTargetService struct {
	client *connectivity.AliyunClient
}

func NewFlinkDeploymentTargetService(client *connectivity.AliyunClient) *FlinkDeploymentTargetService {
	return &FlinkDeploymentTargetService{
		client: client,
	}
}

func (s *FlinkDeploymentTargetService) DescribeFlinkDeploymentTarget(id string) (object flink.DeploymentTarget, err error) {
	workspaceId, namespaceName, targetName, err := parseDeploymentTargetId(id)
	if err != nil {
		return object, WrapError(err)
	}

	api, resourceID, err := s.apiAndWorkspaceResourceID(workspaceId)
	if err != nil {
		return object, WrapError(err)
	}

	target, err := api.GetDeploymentTarget(resourceID, namespaceName, targetName)
	if err != nil {
		return object, WrapError(err)
	}

	if target == nil {
		return object, WrapErrorf(Error(GetNotFoundMessage("FlinkDeploymentTarget", id)), NotFoundMsg, ProviderERROR)
	}

	return *target, nil
}

func (s *FlinkDeploymentTargetService) DescribeFlinkDeploymentTargets(workspaceId, namespaceName string) (objects []flink.DeploymentTarget, err error) {
	api, resourceID, err := s.apiAndWorkspaceResourceID(workspaceId)
	if err != nil {
		return objects, WrapError(err)
	}

	targets, err := api.ListDeploymentTargets(resourceID, namespaceName)
	if err != nil {
		return objects, WrapError(err)
	}

	return targets, nil
}

func (s *FlinkDeploymentTargetService) CreateFlinkDeploymentTarget(workspaceId, namespaceName string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	api, resourceID, err := s.apiAndWorkspaceResourceID(workspaceId)
	if err != nil {
		return nil, WrapError(err)
	}

	var result *flink.DeploymentTarget
	if flinkDeploymentTargetUsesV2(target) {
		result, err = api.CreateDeploymentTargetV2(resourceID, namespaceName, target)
	} else {
		result, err = api.CreateDeploymentTarget(resourceID, namespaceName, target)
	}
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

func (s *FlinkDeploymentTargetService) UpdateFlinkDeploymentTarget(workspaceId, namespaceName string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	api, resourceID, err := s.apiAndWorkspaceResourceID(workspaceId)
	if err != nil {
		return nil, WrapError(err)
	}

	var result *flink.DeploymentTarget
	if flinkDeploymentTargetUsesV2(target) {
		result, err = api.UpdateDeploymentTargetV2(resourceID, namespaceName, target)
	} else {
		result, err = api.UpdateDeploymentTarget(resourceID, namespaceName, target)
	}
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

func (s *FlinkDeploymentTargetService) DeleteFlinkDeploymentTarget(workspaceId, namespaceName, targetName string) error {
	api, resourceID, err := s.apiAndWorkspaceResourceID(workspaceId)
	if err != nil {
		return WrapError(err)
	}

	err = api.DeleteDeploymentTarget(resourceID, namespaceName, targetName)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *FlinkDeploymentTargetService) apiAndWorkspaceResourceID(instanceID string) (*flink.FlinkAPI, string, error) {
	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return nil, "", err
	}
	workspace, err := flinkService.GetAPI().GetWorkspace(instanceID)
	if err != nil {
		return nil, "", err
	}
	if workspace == nil || workspace.ResourceId == "" {
		return nil, "", fmt.Errorf("Flink workspace %q does not expose a ResourceId", instanceID)
	}
	return flinkService.GetAPI(), workspace.ResourceId, nil
}

func flinkDeploymentTargetUsesV2(target *flink.DeploymentTarget) bool {
	return target != nil && target.Quota != nil && target.Quota.Request != nil
}

func (s *FlinkDeploymentTargetService) DeploymentTargetStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFlinkDeploymentTarget(id)
		if err != nil {
			if NotFoundError(err) {
				// Return nil, empty state when resource is deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Name == failState {
				return object, object.Name, WrapError(Error(FailedToReachTargetStatus, object.Name))
			}
		}

		return object, "Available", nil
	}
}

// Helper functions for ID parsing and formatting
func formatDeploymentTargetId(workspaceId, namespaceName, targetName string) string {
	return fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, targetName)
}

func parseDeploymentTargetId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", WrapError(Error("Invalid deployment target ID format. Expected format: workspaceId:namespaceName:targetName"))
	}
	return parts[0], parts[1], parts[2], nil
}
