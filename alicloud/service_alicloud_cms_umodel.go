package alicloud

func (s *CmsService) ListCmsUmodels() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsUmodels")
}

func (s *CmsService) GetCmsUmodel(modelID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsUmodel")
}
