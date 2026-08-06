package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// CreateCmsEntityStore creates a CMS 2.0 entity store.
func (s *CmsService) CreateCmsEntityStore(store *cmsapi.CmsEntityStore) (*cmsapi.CmsEntityStore, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreateEntityStore(store)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsEntityStore reads a CMS 2.0 entity store by ID.
func (s *CmsService) GetCmsEntityStore(storeID string) (*cmsapi.CmsEntityStore, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetEntityStore(storeID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsEntityStores lists CMS 2.0 entity stores.
func (s *CmsService) ListCmsEntityStores(query *cmsapi.CmsListQuery) ([]cmsapi.CmsEntityStore, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListEntityStores(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsEntityStore removes a CMS 2.0 entity store. Deleting a store
// that is already gone is treated as success.
func (s *CmsService) DeleteCmsEntityStore(storeID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteEntityStore(storeID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}
