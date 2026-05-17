package alicloud

func (s *CmsService) ListCmsEntityStores() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsEntityStores")
}

func (s *CmsService) GetCmsEntityStore(storeName string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsEntityStore")
}
