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

// Deployment methods
func (s *FlinkService) GetDeployment(id string) (*aliyunFlinkAPI.Deployment, error) {
	// Parse deployment ID to extract workspace, namespace and deployment ID
	// Format: workspaceId:namespace:deploymentId
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}
	return s.flinkAPI.GetDeployment(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	// Parse deployment ID to extract workspace and namespace
	// Format: workspaceId:namespace:deploymentId (deploymentId part will be generated)
	workspaceId, namespaceName, _, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	return s.flinkAPI.CreateDeployment(deployment)
}

func (s *FlinkService) UpdateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	// Parse deployment ID to extract workspace, namespace and deployment ID
	// Format: workspaceId:namespace:deploymentId
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	deployment.DeploymentId = deploymentId
	return s.flinkAPI.UpdateDeployment(deployment)
}

func (s *FlinkService) DeleteDeployment(id string) error {
	// Parse deployment ID to extract workspace, namespace and deployment ID
	// Format: workspaceId:namespace:deploymentId
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return err
	}
	return s.flinkAPI.DeleteDeployment(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) ListDeployments(id string) ([]aliyunFlinkAPI.Deployment, error) {
	// Parse namespace ID to extract workspace ID and namespace name
	// Format: workspaceId:namespace (for listing deployments in a namespace)
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid namespace ID format for listing deployments, expected workspaceId:namespace, got %s", id)
	}
	workspaceId := parts[0]
	namespaceName := parts[1]
	return s.flinkAPI.ListDeployments(workspaceId, namespaceName)
}

func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := s.GetDeployment(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "NotFound", nil
			}
			return nil, "FAILED", WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
		}

		// If deployment is nil, it means the resource doesn't exist
		if deployment == nil {
			return nil, "NotFound", nil
		}

		// Check for fail states
		for _, failState := range failStates {
			if deployment.Status == failState {
				return deployment, deployment.Status, WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
			}
		}

		return deployment, deployment.Status, nil
	}
}
