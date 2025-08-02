package alicloud

import (
	"fmt"
	"strings"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BuildDeploymentFolderId constructs the composite resource ID for deployment folder
// Format: workspace_id:namespace:folder_id
func (s *FlinkService) BuildDeploymentFolderId(workspaceId, namespace, folderId string) string {
	return fmt.Sprintf("%s:%s:%s", workspaceId, namespace, folderId)
}

// ParseDeploymentFolderId parses the composite resource ID for deployment folder
// Returns workspaceId, namespace, folderId and error
func (s *FlinkService) ParseDeploymentFolderId(id string) (workspaceId, namespace, folderId string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid resource id format: expected workspace_id:namespace:folder_id, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// CreateDeploymentFolder creates a new deployment folder
func (s *FlinkService) CreateDeploymentFolder(workspaceId, namespace, folderName, parentId string) (*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().CreateDeploymentFolder(workspaceId, namespace, folderName, parentId)
}

// GetDeploymentRootFolder retrieves the root deployment folder for a namespace
func (s *FlinkService) GetDeploymentRootFolder(workspaceId, namespace string) (*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().GetDeploymentRootFolder(workspaceId, namespace)
}

// GetDeploymentFolder retrieves a deployment folder by ID
func (s *FlinkService) GetDeploymentFolder(workspaceId, namespace, folderId string) (*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().GetDeploymentFolder(workspaceId, namespace, folderId)
}

// UpdateDeploymentFolder updates a deployment folder's properties
func (s *FlinkService) UpdateDeploymentFolder(workspaceId, namespace, folderId, newName string) (*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().UpdateDeploymentFolder(workspaceId, namespace, folderId, newName)
}

// DeleteDeploymentFolder deletes a deployment folder
func (s *FlinkService) DeleteDeploymentFolder(workspaceId, namespace, folderId string) error {
	return s.GetAPI().DeleteDeploymentFolder(workspaceId, namespace, folderId)
}

// ListDeploymentFolders lists all deployment folders in a namespace
func (s *FlinkService) ListDeploymentFolders(workspaceId, namespace string) ([]*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().ListDeploymentFolders(workspaceId, namespace)
}

// GetDeploymentFoldersByParent retrieves all folders under a specific parent folder
func (s *FlinkService) GetDeploymentFoldersByParent(workspaceId, namespace, parentId string) ([]*aliyunFlinkAPI.DeploymentFolder, error) {
	return s.GetAPI().GetDeploymentFoldersByParent(workspaceId, namespace, parentId)
}

// FindDeploymentFolderByName finds a deployment folder by name within a specific parent folder
func (s *FlinkService) FindDeploymentFolderByName(workspaceId, namespace, folderName, parentId string) (*aliyunFlinkAPI.DeploymentFolder, error) {
	// Get folders only within the specific parent folder for better performance
	folders, err := s.GetDeploymentFoldersByParent(workspaceId, namespace, parentId)
	if err != nil {
		return nil, WrapError(err)
	}

	// Filter folders by name within the parent folder
	for _, folder := range folders {
		if folder.Name == folderName {
			return folder, nil
		}
	}

	// If no folder found, return nil without error (not found is a valid state)
	return nil, nil
}

// DeploymentFolderStateRefreshFunc returns a StateRefreshFunc for deployment folder state management
func (s *FlinkService) DeploymentFolderStateRefreshFunc(workspaceId, namespace, folderId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.GetDeploymentFolder(workspaceId, namespace, folderId)
		if err != nil {
			if IsNotFoundError(err) {
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
