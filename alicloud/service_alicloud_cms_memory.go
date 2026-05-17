package alicloud

func (s *CmsService) ListCmsMemories(storeName string) ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsMemories")
}

func (s *CmsService) GetCmsMemory(storeName, memoryID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsMemory")
}
