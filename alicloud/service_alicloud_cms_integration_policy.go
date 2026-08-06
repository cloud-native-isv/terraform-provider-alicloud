package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// CreateCmsIntegrationPolicy creates a CMS 2.0 integration policy.
func (s *CmsService) CreateCmsIntegrationPolicy(policy *cmsapi.CmsIntegrationPolicy) (*cmsapi.CmsIntegrationPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreateIntegrationPolicy(policy)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsIntegrationPolicy reads a CMS 2.0 integration policy by ID.
func (s *CmsService) GetCmsIntegrationPolicy(policyID string) (*cmsapi.CmsIntegrationPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetIntegrationPolicy(policyID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpdateCmsIntegrationPolicy updates a CMS 2.0 integration policy.
func (s *CmsService) UpdateCmsIntegrationPolicy(policyID string, policy *cmsapi.CmsIntegrationPolicy) (*cmsapi.CmsIntegrationPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.UpdateIntegrationPolicy(policyID, policy)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsIntegrationPolicy removes a CMS 2.0 integration policy. Deleting
// a policy that is already gone is treated as success.
func (s *CmsService) DeleteCmsIntegrationPolicy(policyID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteIntegrationPolicy(policyID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}

// ListCmsIntegrationPolicies lists CMS 2.0 integration policies.
func (s *CmsService) ListCmsIntegrationPolicies(query *cmsapi.CmsListQuery) ([]cmsapi.CmsIntegrationPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListIntegrationPolicies(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}
