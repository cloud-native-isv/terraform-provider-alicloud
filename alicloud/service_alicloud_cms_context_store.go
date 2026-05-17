package alicloud

func (s *CmsService) ListCmsContextStores() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsContextStores")
}

func (s *CmsService) GetCmsContextStore(storeName string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsContextStore")
}
