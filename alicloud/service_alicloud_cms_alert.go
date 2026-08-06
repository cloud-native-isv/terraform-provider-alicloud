package alicloud

import (
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// ListCmsAlertRules queries the CMS 2.0 alert rules.
func (s *CmsService) ListCmsAlertRules(query *cmsapi.CmsListQuery) ([]cmsapi.CmsAlertRule, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.QueryAlertRules(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ManageCmsAlertRules creates or adjusts CMS 2.0 alert rules through the
// batch ManageAlertRules action.
func (s *CmsService) ManageCmsAlertRules(rule *cmsapi.CmsAlertRule) ([]cmsapi.CmsAlertRule, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ManageAlertRules(rule)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsAlertWebhooks lists the CMS 2.0 alert webhooks.
func (s *CmsService) ListCmsAlertWebhooks(query *cmsapi.CmsListQuery) ([]cmsapi.CmsAlertWebhook, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListAlertWebhooks(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// CreateCmsNotifyPolicy creates a CMS 2.0 notification policy inside the
// given workspace (goal C4: alert notifications converge on NotifyPolicy).
func (s *CmsService) CreateCmsNotifyPolicy(policy *cmsapi.CmsNotifyPolicy) (*cmsapi.CmsNotifyPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.CreateNotifyPolicy(policy)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// GetCmsNotifyPolicy reads one CMS 2.0 notification policy.
func (s *CmsService) GetCmsNotifyPolicy(workspaceID, policyID string) (*cmsapi.CmsNotifyPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.GetNotifyPolicy(workspaceID, policyID)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// ListCmsNotifyPolicies lists the CMS 2.0 notification policies of a
// workspace.
func (s *CmsService) ListCmsNotifyPolicies(query *cmsapi.CmsNotifyPolicyQuery) ([]cmsapi.CmsNotifyPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.ListNotifyPolicies(query)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpdateCmsNotifyPolicy updates a CMS 2.0 notification policy.
func (s *CmsService) UpdateCmsNotifyPolicy(policy *cmsapi.CmsNotifyPolicy) (*cmsapi.CmsNotifyPolicy, error) {
	api, err := s.cmsAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	result, err := api.UpdateNotifyPolicy(policy)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// DeleteCmsNotifyPolicy removes a CMS 2.0 notification policy. Deleting a
// policy that is already gone is treated as success.
func (s *CmsService) DeleteCmsNotifyPolicy(workspaceID, policyID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DeleteNotifyPolicy(workspaceID, policyID); err != nil {
		if commonapi.IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}
	return nil
}

// EnableCmsNotifyPolicy enables a CMS 2.0 notification policy.
func (s *CmsService) EnableCmsNotifyPolicy(workspaceID, policyID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.EnableNotifyPolicy(workspaceID, policyID); err != nil {
		return WrapError(err)
	}
	return nil
}

// DisableCmsNotifyPolicy disables a CMS 2.0 notification policy.
func (s *CmsService) DisableCmsNotifyPolicy(workspaceID, policyID string) error {
	api, err := s.cmsAPI()
	if err != nil {
		return WrapError(err)
	}
	if _, err := api.DisableNotifyPolicy(workspaceID, policyID); err != nil {
		return WrapError(err)
	}
	return nil
}
