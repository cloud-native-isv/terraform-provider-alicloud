package alicloud

func (s *CmsService) ListCmsAddons() ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsAddons")
}

func (s *CmsService) ListCmsAddonReleases(addonName string) ([]map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("ListCmsAddonReleases")
}

func (s *CmsService) GetCmsAddonRelease(addonName, version string) (map[string]any, error) {
	return nil, cmsServiceCapabilityBlocked("GetCmsAddonRelease")
}
