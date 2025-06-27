package alicloud

import (
	"fmt"

	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateDeploymentFolder creates a new deployment folder
func (s *FlinkService) CreateDeploymentFolder(workspaceId, namespace, folderName, parentId string) (*flinkAPI.DeploymentFolder, error) {
	return s.flinkAPI.CreateDeploymentFolder(workspaceId, namespace, folderName, parentId)
}

// GetDeploymentFolder retrieves a deployment folder by ID
func (s *FlinkService) GetDeploymentFolder(workspaceId, namespace, folderId string) (*flinkAPI.DeploymentFolder, error) {
	return s.flinkAPI.GetDeploymentFolder(workspaceId, namespace, folderId)
}

// UpdateDeploymentFolder updates a deployment folder's properties
func (s *FlinkService) UpdateDeploymentFolder(workspaceId, namespace, folderId, newName string) (*flinkAPI.DeploymentFolder, error) {
	return s.flinkAPI.UpdateDeploymentFolder(workspaceId, namespace, folderId, newName)
}

// DeleteDeploymentFolder deletes a deployment folder
func (s *FlinkService) DeleteDeploymentFolder(workspaceId, namespace, folderId string) error {
	return s.flinkAPI.DeleteDeploymentFolder(workspaceId, namespace, folderId)
}

// ListDeploymentFolders lists all deployment folders in a namespace
func (s *FlinkService) ListDeploymentFolders(workspaceId, namespace string) ([]*flinkAPI.DeploymentFolder, error) {
	return s.flinkAPI.ListDeploymentFolders(workspaceId, namespace)
}

// GetDeploymentFoldersByParent retrieves all folders under a specific parent folder
func (s *FlinkService) GetDeploymentFoldersByParent(workspaceId, namespace, parentId string) ([]*flinkAPI.DeploymentFolder, error) {
	return s.flinkAPI.GetDeploymentFoldersByParent(workspaceId, namespace, parentId)
}

// DeploymentFolderStateRefreshFunc returns a StateRefreshFunc for deployment folder state management
func (s *FlinkService) DeploymentFolderStateRefreshFunc(workspaceId, namespace, folderId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.GetDeploymentFolder(workspaceId, namespace, folderId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Name == failState {
				return object, object.Name, WrapError(fmt.Errorf("deployment folder is in %s state", failState))
			}
		}

		return object, "Available", nil
	}
}
