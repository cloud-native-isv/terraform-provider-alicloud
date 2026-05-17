package alicloud

func (s *CmsService) ListCmsAggTaskGroups() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsAggTaskGroups")
}

func (s *CmsService) GetCmsAggTaskGroup(groupID string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsAggTaskGroup")
}

func (s *CmsService) UpdateCmsAggTaskGroupStatus(groupID, status string) error {
	return cmsServiceCapabilityBlocked("UpdateCmsAggTaskGroupStatus")
}
