package alicloud

func (s *CmsService) ListCmsPipelines() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsPipelines")
}

func (s *CmsService) GetCmsPipeline(pipelineID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsPipeline")
}
