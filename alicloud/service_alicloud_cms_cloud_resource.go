package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// CreateCmsCloudResource registers a cloud resource with CMS 2.0.
func (s *CmsService) CreateCmsCloudResource(resource *cmsapi.CmsCloudResource) (*cmsapi.CmsCloudResource, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreateCloudResource(resource)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsCloudResource reads a CMS 2.0 cloud resource by ID.
func (s *CmsService) GetCmsCloudResource(resourceID string) (*cmsapi.CmsCloudResource, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetCloudResource(resourceID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsCloudResources lists CMS 2.0 cloud resources.
func (s *CmsService) ListCmsCloudResources(query *cmsapi.CmsListQuery) ([]cmsapi.CmsCloudResource, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListCloudResources(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsCloudResource unregisters a CMS 2.0 cloud resource. Deleting a
// resource that is already gone is treated as success.
func (s *CmsService) DeleteCmsCloudResource(resourceID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteCloudResource(resourceID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}
