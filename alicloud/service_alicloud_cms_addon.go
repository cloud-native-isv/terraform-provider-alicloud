package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// ListCmsAddons lists the CMS 2.0 addon catalog entries.
func (s *CmsService) ListCmsAddons(query *cmsapi.CmsListQuery) ([]cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListAddons(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsAddon reads a single CMS 2.0 addon catalog entry by name.
func (s *CmsService) GetCmsAddon(addonName string) (*cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetAddon(addonName)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// CreateCmsAddonRelease publishes a release for a CMS 2.0 addon.
func (s *CmsService) CreateCmsAddonRelease(addon *cmsapi.CmsAddons) (*cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreateAddonRelease(addon)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsAddonRelease reads one release of a CMS 2.0 addon.
func (s *CmsService) GetCmsAddonRelease(addonName, version string) (*cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetAddonRelease(addonName, version)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpdateCmsAddonRelease updates a release of a CMS 2.0 addon.
func (s *CmsService) UpdateCmsAddonRelease(addon *cmsapi.CmsAddons) (*cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.UpdateAddonRelease(addon)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsAddonRelease removes a release of a CMS 2.0 addon. Deleting a
// release that is already gone is treated as success.
func (s *CmsService) DeleteCmsAddonRelease(addonName, version string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteAddonRelease(addonName, version); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}

// ListCmsAddonReleases lists the releases of a CMS 2.0 addon.
func (s *CmsService) ListCmsAddonReleases(addonName string, query *cmsapi.CmsListQuery) ([]cmsapi.CmsAddons, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListAddonReleases(addonName, query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}
