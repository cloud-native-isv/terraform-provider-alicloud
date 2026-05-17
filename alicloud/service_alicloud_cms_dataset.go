package alicloud

func (s *CmsService) ListCmsDatasets() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsDatasets")
}

func (s *CmsService) GetCmsDataset(datasetID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsDataset")
}

func (s *CmsService) ExecuteCmsDatasetQuery(datasetID string, query map[string]any) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ExecuteCmsDatasetQuery")
}
