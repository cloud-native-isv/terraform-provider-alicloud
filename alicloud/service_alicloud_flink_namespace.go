package alicloud

import (
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Namespace methods
func (s *FlinkService) ListNamespaces(workspaceId string) ([]aliyunFlinkAPI.Namespace, error) {
	return s.GetAPI().ListNamespaces(workspaceId)
}

func (s *FlinkService) CreateNamespace(workspaceId string, namespace *aliyunFlinkAPI.Namespace) (*aliyunFlinkAPI.Namespace, error) {
	// Create namespace using the flinkAPI directly
	result, err := s.GetAPI().CreateNamespace(workspaceId, namespace)
	return result, err
}

func (s *FlinkService) GetNamespace(workspaceId, namespaceName string) (*aliyunFlinkAPI.Namespace, error) {
	// Fetch the namespace using the API
	return s.GetAPI().GetNamespace(workspaceId, namespaceName)
}

func (s *FlinkService) DeleteNamespace(workspaceId string, namespaceName string) error {
	// Delete the namespace using flinkAPI
	return s.GetAPI().DeleteNamespace(workspaceId, namespaceName)
}

func (s *FlinkService) UpdateNamespace(workspaceId string, namespace *aliyunFlinkAPI.Namespace) (*aliyunFlinkAPI.Namespace, error) {
	// Update namespace using the flinkAPI directly
	result, err := s.GetAPI().UpdateNamespace(workspaceId, namespace)
	return result, err
}

// FlinkNamespaceStateRefreshFunc provides state refresh for namespaces
func (s *FlinkService) FlinkNamespaceStateRefreshFunc(workspaceId string, namespaceName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		namespace, err := s.GetNamespace(workspaceId, namespaceName)
		if err != nil {
			if NotFoundError(err) {
				// Namespace not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For namespaces, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if namespace is in a failed state (if any fail states are defined)
			if namespace.Status == failState {
				return namespace, namespace.Status, WrapError(Error(FailedToReachTargetStatus, namespace.Status))
			}
		}

		return namespace, "Available", nil
	}
}
