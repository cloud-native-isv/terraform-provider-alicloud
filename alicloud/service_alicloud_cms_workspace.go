package alicloud

func (s *CmsService) ListCmsWorkspaces() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsWorkspaces")
}

func (s *CmsService) GetCmsWorkspace(workspaceID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsWorkspace")
}

func (s *CmsService) PutCmsWorkspace(workspace map[string]any) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("PutCmsWorkspace")
}

func (s *CmsService) DeleteCmsWorkspace(workspaceID string) error {
	return cmsServiceCapabilityBlocked("DeleteCmsWorkspace")
}
