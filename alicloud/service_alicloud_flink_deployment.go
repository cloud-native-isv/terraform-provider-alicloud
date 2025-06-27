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
	return s.flinkAPI.GetDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(namespaceName *string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	deployment.Namespace = *namespaceName
	return s.flinkAPI.CreateDeployment(deployment)
}

func (s *FlinkService) UpdateDeployment(deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	return s.flinkAPI.UpdateDeployment(deployment)
}

func (s *FlinkService) DeleteDeployment(namespaceName, deploymentId string) error {
	return s.flinkAPI.DeleteDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) ListDeployments(namespaceName string) ([]aliyunFlinkAPI.Deployment, error) {
	return s.flinkAPI.ListDeployments(namespaceName)
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
