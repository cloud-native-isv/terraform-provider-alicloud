package alicloud

func (s *CmsService) ListCmsAlertRules() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsAlertRules")
}

func (s *CmsService) ManageCmsAlertRules(rule map[string]any) ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ManageCmsAlertRules")
}

func (s *CmsService) ListCmsAlertWebhooks() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsAlertWebhooks")
}
