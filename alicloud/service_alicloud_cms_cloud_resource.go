package alicloud

func (s *CmsService) ListCmsCloudResources() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsCloudResources")
}

func (s *CmsService) GetCmsCloudResource(resourceID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsCloudResource")
}
