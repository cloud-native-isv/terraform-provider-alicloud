package alicloud

import (
	"fmt"
	"strings"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func parseDeploymentId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid deployment ID format, expected workspaceId:namespace:deploymentId, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func (s *FlinkService) GetDeployment(id string) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}
	return s.GetAPI().GetDeployment(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, _, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	result, err := s.GetAPI().CreateDeployment(deployment)
	if err == nil && result != nil {
		addDebugJson("CreateDeployment", result)
	}
	return result, err
}

func (s *FlinkService) UpdateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	deployment.DeploymentId = deploymentId
	result, err := s.GetAPI().UpdateDeployment(deployment)
	if err == nil && result != nil {
		addDebugJson("UpdateDeployment", result)
	}
	return result, err
}

func (s *FlinkService) DeleteDeployment(id string) error {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return err
	}
	err = s.GetAPI().DeleteDeployment(workspaceId, namespaceName, deploymentId)
	if err == nil {
		addDebugJson("DeleteDeployment", fmt.Sprintf("Deployment %s deleted successfully", deploymentId))
	}
	return err
}

func (s *FlinkService) ListDeployments(id string) ([]aliyunFlinkAPI.Deployment, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid namespace ID format for listing deployments, expected workspaceId:namespace, got %s", id)
	}
	workspaceId := parts[0]
	namespaceName := parts[1]
	return s.GetAPI().ListDeployments(workspaceId, namespaceName)
}

func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := s.GetDeployment(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "NotFound", nil
			}
			return nil, "FAILED", WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
		}

		if deployment == nil {
			return nil, "NotFound", nil
		}

		for _, failState := range failStates {
			if deployment.Status == failState {
				return deployment, deployment.Status, WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
			}
		}

		return deployment, deployment.Status, nil
	}
}
