package alicloud

func (s *CmsService) ListCmsIntegrationPolicies() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsIntegrationPolicies")
}

func (s *CmsService) GetCmsIntegrationPolicy(policyID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsIntegrationPolicy")
}
