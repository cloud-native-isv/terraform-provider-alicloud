package alicloud

import (
	"time"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Workspace methods
func (s *FlinkService) DescribeFlinkWorkspace(id string) (*aliyunFlinkAPI.Workspace, error) {
	return s.GetAPI().GetWorkspace(id)
}

func (s *FlinkService) CreateInstance(workspace *aliyunFlinkAPI.Workspace) (*aliyunFlinkAPI.Workspace, error) {
	return s.GetAPI().CreateWorkspace(workspace)
}

func (s *FlinkService) DeleteInstance(id string) error {
	return s.GetAPI().DeleteWorkspace(id)
}

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspace, err := s.GetAPI().GetWorkspace(id)
		if err != nil {
			// Handle the case where workspace is temporarily not found after creation
			// This is common with cloud resources that have async creation processes
			if IsNotFoundError(err) { // Use generic IsNotFoundError instead of specific error code
				// Return empty state to indicate the resource is still being created
				return nil, aliyunFlinkAPI.FlinkWorkspaceStatusCreating.String(), nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetWorkspace", AlibabaCloudSdkGoERROR)
		}
		return workspace, workspace.Status, nil
	}
}

// WaitForWorkspaceStarting waits for a Flink workspace to reach running state after creation
func (s *FlinkService) WaitForWorkspaceStarting(id string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending:    aliyunFlinkAPI.FlinkWorkspaceStatusesToStrings([]aliyunFlinkAPI.FlinkWorkspaceStatus{aliyunFlinkAPI.FlinkWorkspaceStatusCreating}),
		Target:     aliyunFlinkAPI.FlinkWorkspaceStatusesToStrings([]aliyunFlinkAPI.FlinkWorkspaceStatus{aliyunFlinkAPI.FlinkWorkspaceStatusRunning}),
		Refresh:    s.FlinkWorkspaceStateRefreshFunc(id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

// WaitForWorkspaceDeleting waits for a Flink workspace to be completely deleted
func (s *FlinkService) WaitForWorkspaceDeleting(id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: aliyunFlinkAPI.FlinkWorkspaceStatusesToStrings([]aliyunFlinkAPI.FlinkWorkspaceStatus{aliyunFlinkAPI.FlinkWorkspaceStatusDeleting}),
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			// Check if the workspace still exists
			workspace, err := s.DescribeFlinkWorkspace(id)
			if err != nil {
				if IsNotFoundError(err) {
					// Resource is gone, which is what we want
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			// If we can still get the workspace, it's still being deleted
			return workspace, workspace.Status, nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return err
}

// Instance/Workspace methods (aliases for workspace methods)
func (s *FlinkService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	return s.GetAPI().ListWorkspaces()
}
