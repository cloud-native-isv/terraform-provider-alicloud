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

	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return object, WrapError(err)
	}

	target, err := flinkService.flinkAPI.GetDeploymentTarget(workspaceId, namespaceName, targetName)
	if err != nil {
		return object, WrapError(err)
	}

	if target == nil {
		return object, WrapErrorf(Error(GetNotFoundMessage("FlinkDeploymentTarget", id)), NotFoundMsg, ProviderERROR)
	}

	return *target, nil
}

func (s *FlinkDeploymentTargetService) DescribeFlinkDeploymentTargets(workspaceId, namespaceName string) (objects []flink.DeploymentTarget, err error) {
	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return objects, WrapError(err)
	}

	targets, err := flinkService.flinkAPI.ListDeploymentTargets(workspaceId, namespaceName)
	if err != nil {
		return objects, WrapError(err)
	}

	return targets, nil
}

func (s *FlinkDeploymentTargetService) CreateFlinkDeploymentTarget(workspaceId, namespaceName string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return nil, WrapError(err)
	}

	result, err := flinkService.flinkAPI.CreateDeploymentTarget(workspaceId, namespaceName, target)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

func (s *FlinkDeploymentTargetService) UpdateFlinkDeploymentTarget(workspaceId, namespaceName string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return nil, WrapError(err)
	}

	result, err := flinkService.flinkAPI.UpdateDeploymentTarget(workspaceId, namespaceName, target)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

func (s *FlinkDeploymentTargetService) DeleteFlinkDeploymentTarget(workspaceId, namespaceName, targetName string) error {
	flinkService, err := NewFlinkService(s.client)
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.flinkAPI.DeleteDeploymentTarget(workspaceId, namespaceName, targetName)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *FlinkDeploymentTargetService) DeploymentTargetStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFlinkDeploymentTarget(id)
		if err != nil {
			if IsNotFoundError(err) {
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
