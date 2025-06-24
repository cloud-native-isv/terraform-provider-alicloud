package alicloud

import (
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Workspace methods
func (s *FlinkService) DescribeFlinkWorkspace(id string) (*aliyunFlinkAPI.Workspace, error) {
	return s.aliyunFlinkAPI.GetWorkspace(id)
}

func (s *FlinkService) CreateInstance(workspace *aliyunFlinkAPI.Workspace) (*aliyunFlinkAPI.Workspace, error) {
	return s.aliyunFlinkAPI.CreateWorkspace(workspace)
}

func (s *FlinkService) DeleteInstance(id string) error {
	return s.aliyunFlinkAPI.DeleteWorkspace(id)
}

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspace, err := s.aliyunFlinkAPI.GetWorkspace(id)
		if err != nil {
			// Handle the case where workspace is temporarily not found after creation
			// This is common with cloud resources that have async creation processes
			if IsExpectedErrors(err, []string{"903021"}) { // not exist yet
				// Return empty state to indicate the resource is still being created
				return nil, "CREATING", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetWorkspace", AlibabaCloudSdkGoERROR)
		}
		return workspace, workspace.Status, nil
	}
}

// Instance/Workspace methods (aliases for workspace methods)
func (s *FlinkService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	return s.aliyunFlinkAPI.ListWorkspaces()
}
