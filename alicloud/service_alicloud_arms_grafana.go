package alicloud

import (
	"time"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsGrafanaWorkspace describes ARMS Grafana workspace using CWS-Lib-Go API
func (s *ArmsService) DescribeArmsGrafanaWorkspace(id string) (*aliyunArmsAPI.GrafanaWorkspaceDetail, error) {
	if id == "" {
		return nil, WrapError(Error("GrafanaWorkspaceId cannot be empty"))
	}

	workspace, err := s.armsAPI.GetGrafanaWorkspace(id)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}

	return workspace, nil
}

// ListGrafanaWorkspace lists all ARMS Grafana workspaces using CWS-Lib-Go API
func (s *ArmsService) ListGrafanaWorkspace(resourceGroupId, aliyunLang string) ([]*aliyunArmsAPI.GrafanaWorkspace, error) {
	var allWorkspaces []*aliyunArmsAPI.GrafanaWorkspace
	page := 1
	pageSize := 50

	for {
		workspaces, err := s.armsAPI.ListGrafanaWorkspaces(page, pageSize, resourceGroupId, aliyunLang)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, "all", "ListGrafanaWorkspaces", AlibabaCloudSdkGoERROR)
		}

		if len(workspaces) == 0 {
			break
		}

		allWorkspaces = append(allWorkspaces, workspaces...)

		// If we got fewer results than requested, we've reached the end
		if len(workspaces) < pageSize {
			break
		}

		page++

		// Safety check to prevent infinite loops (limit to 10000 resources)
		if len(allWorkspaces) >= 10000 {
			break
		}
	}

	return allWorkspaces, nil
}

// CreateGrafanaWorkspace creates a new ARMS Grafana workspace
func (s *ArmsService) CreateGrafanaWorkspace(grafanaWorkspaceName, grafanaWorkspaceEdition string, description, resourceGroupId, grafanaVersion, password, aliyunLang string, tags []aliyunArmsAPI.GrafanaWorkspaceTag) (*aliyunArmsAPI.GrafanaWorkspace, error) {
	workspace, err := s.armsAPI.CreateGrafanaWorkspace(grafanaWorkspaceName, grafanaWorkspaceEdition, description, resourceGroupId, grafanaVersion, password, aliyunLang, tags)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_grafana_workspace", "CreateGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}
	return workspace, nil
}

// UpdateGrafanaWorkspace updates an existing ARMS Grafana workspace
func (s *ArmsService) UpdateGrafanaWorkspace(grafanaWorkspaceId string, grafanaWorkspaceName, description, aliyunLang string) (*aliyunArmsAPI.GrafanaWorkspaceDetail, error) {
	workspace, err := s.armsAPI.UpdateGrafanaWorkspace(grafanaWorkspaceId, grafanaWorkspaceName, description, aliyunLang)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, grafanaWorkspaceId, "UpdateGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}
	return workspace, nil
}

// DeleteGrafanaWorkspace deletes an ARMS Grafana workspace
func (s *ArmsService) DeleteGrafanaWorkspace(grafanaWorkspaceId string) error {
	err := s.armsAPI.DeleteGrafanaWorkspace(grafanaWorkspaceId)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, grafanaWorkspaceId, "DeleteGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ArmsGrafanaWorkspaceStateRefreshFunc returns state refresh function for ARMS Grafana workspace
func (s *ArmsService) ArmsGrafanaWorkspaceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspace, err := s.DescribeArmsGrafanaWorkspace(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Safely handle pointer dereference for Status
		currentStatus := ""
		if workspace.Status != nil {
			currentStatus = *workspace.Status
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return workspace, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}

		return workspace, currentStatus, nil
	}
}

// WaitForArmsGrafanaWorkspaceCreated waits for ARMS Grafana workspace to be created
func (s *ArmsService) WaitForArmsGrafanaWorkspaceCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"Starting", "Creating"}, []string{"Running"}, timeout, 5*time.Second, s.ArmsGrafanaWorkspaceStateRefreshFunc(id, []string{"CreateFailed", "Failed"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsGrafanaWorkspaceDeleted waits for ARMS Grafana workspace to be deleted
func (s *ArmsService) WaitForArmsGrafanaWorkspaceDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"Running", "Stopping", "Deleting"}, []string{}, timeout, 5*time.Second, s.ArmsGrafanaWorkspaceStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
