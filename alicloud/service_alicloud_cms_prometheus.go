package alicloud

func (s *CmsService) ListCmsPrometheusViews() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsPrometheusViews")
}

func (s *CmsService) GetCmsPrometheusInstance(instanceID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsPrometheusInstance")
}

func (s *CmsService) ListCmsPrometheusVirtualInstances() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsPrometheusVirtualInstances")
}

func (s *CmsService) UpdateCmsPrometheusUserSetting(setting map[string]any) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("UpdateCmsPrometheusUserSetting")
}
