package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// PutCmsWorkspace creates or replaces a CMS 2.0 workspace. The CMS
// workspace API is idempotent Put semantics, so create and update share
// this single action.
func (s *CmsService) PutCmsWorkspace(workspace *cmsapi.CmsWorkspace) (*cmsapi.CmsWorkspace, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.PutWorkspace(workspace)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsWorkspace reads a CMS 2.0 workspace by its identifier.
func (s *CmsService) GetCmsWorkspace(workspaceID string) (*cmsapi.CmsWorkspace, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetWorkspace(workspaceID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsWorkspaces lists CMS 2.0 workspaces.
func (s *CmsService) ListCmsWorkspaces(query *cmsapi.CmsListQuery) ([]cmsapi.CmsWorkspace, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListWorkspaces(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsWorkspace removes a CMS 2.0 workspace. Deleting a workspace that
// is already gone is treated as success.
func (s *CmsService) DeleteCmsWorkspace(workspaceID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteWorkspace(workspaceID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}
