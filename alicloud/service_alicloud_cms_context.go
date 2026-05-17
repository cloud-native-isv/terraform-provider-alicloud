package alicloud

func (s *CmsService) ListCmsContexts(storeName string) ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsContexts")
}

func (s *CmsService) GetCmsContext(storeName, contextID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsContext")
}
