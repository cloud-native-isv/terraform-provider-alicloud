package alicloud

func (s *CmsService) ListCmsMemoryStores() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsMemoryStores")
}

func (s *CmsService) GetCmsMemoryStore(storeName string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsMemoryStore")
}
