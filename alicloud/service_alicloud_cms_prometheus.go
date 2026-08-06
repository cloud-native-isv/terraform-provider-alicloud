package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// CreateCmsPrometheusInstance creates a CMS 2.0 Prometheus instance.
func (s *CmsService) CreateCmsPrometheusInstance(instance *cmsapi.CmsPrometheusInstance) (*cmsapi.CmsPrometheusInstance, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreatePrometheusInstance(instance)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsPrometheusInstance reads a CMS 2.0 Prometheus instance by ID.
func (s *CmsService) GetCmsPrometheusInstance(instanceID string) (*cmsapi.CmsPrometheusInstance, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetPrometheusInstance(instanceID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpdateCmsPrometheusInstance updates a CMS 2.0 Prometheus instance.
func (s *CmsService) UpdateCmsPrometheusInstance(instanceID string, instance *cmsapi.CmsPrometheusInstance) (*cmsapi.CmsPrometheusInstance, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.UpdatePrometheusInstance(instanceID, instance)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsPrometheusInstance removes a CMS 2.0 Prometheus instance.
// Deleting an instance that is already gone is treated as success.
func (s *CmsService) DeleteCmsPrometheusInstance(instanceID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeletePrometheusInstance(instanceID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}

// ListCmsPrometheusViews lists CMS 2.0 Prometheus views.
func (s *CmsService) ListCmsPrometheusViews(query *cmsapi.CmsListQuery) ([]cmsapi.CmsPrometheusView, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListPrometheusViews(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsPrometheusVirtualInstances lists CMS 2.0 Prometheus virtual
// instances.
func (s *CmsService) ListCmsPrometheusVirtualInstances(query *cmsapi.CmsListQuery) ([]cmsapi.CmsPrometheusVirtualInstance, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListPrometheusVirtualInstances(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsPrometheusUserSetting reads the account-level CMS 2.0 Prometheus
// user setting.
func (s *CmsService) GetCmsPrometheusUserSetting() (*cmsapi.CmsPrometheusUserSetting, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetPrometheusUserSetting()
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpdateCmsPrometheusUserSetting updates the account-level CMS 2.0
// Prometheus user setting.
func (s *CmsService) UpdateCmsPrometheusUserSetting(setting *cmsapi.CmsPrometheusUserSetting) (*cmsapi.CmsPrometheusUserSetting, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.UpdatePrometheusUserSetting(setting)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}
