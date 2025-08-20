package alicloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Custom Domain methods for FCService

// EncodeCustomDomainId encodes domain name into an ID string
func EncodeCustomDomainId(domainName string) string {
	return domainName
}

// DecodeCustomDomainId decodes custom domain ID string to domain name
func DecodeCustomDomainId(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid custom domain ID format, cannot be empty")
	}
	return id, nil
}

// DescribeFCCustomDomain retrieves custom domain information by domain name
func (s *FCService) DescribeFCCustomDomain(domainName string) (*aliyunFCAPI.CustomDomain, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}
	return s.GetAPI().GetCustomDomain(domainName)
}

// ListFCCustomDomains lists all custom domains with optional filters
func (s *FCService) ListFCCustomDomains(prefix *string, limit *int32, nextToken *string) ([]*aliyunFCAPI.CustomDomain, error) {
	var prefixStr, nextTokenStr string
	var limitInt int32

	if prefix != nil {
		prefixStr = *prefix
	}
	if nextToken != nil {
		nextTokenStr = *nextToken
	}
	if limit != nil {
		limitInt = *limit
	}

	domains, _, err := s.GetAPI().ListCustomDomains(prefixStr, nextTokenStr, limitInt)
	return domains, err
}

// CreateFCCustomDomain creates a new FC custom domain
func (s *FCService) CreateFCCustomDomain(domain *aliyunFCAPI.CustomDomain) (*aliyunFCAPI.CustomDomain, error) {
	if domain == nil {
		return nil, fmt.Errorf("domain cannot be nil")
	}
	return s.GetAPI().CreateCustomDomain(domain)
}

// UpdateFCCustomDomain updates an existing FC custom domain
func (s *FCService) UpdateFCCustomDomain(domainName string, domain *aliyunFCAPI.CustomDomain) (*aliyunFCAPI.CustomDomain, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}
	if domain == nil {
		return nil, fmt.Errorf("domain cannot be nil")
	}
	return s.GetAPI().UpdateCustomDomain(domainName, domain)
}

// DeleteFCCustomDomain deletes an FC custom domain
func (s *FCService) DeleteFCCustomDomain(domainName string) error {
	if domainName == "" {
		return fmt.Errorf("domain name cannot be empty")
	}
	return s.GetAPI().DeleteCustomDomain(domainName)
}

// CustomDomainStateRefreshFunc returns a StateRefreshFunc to wait for custom domain status changes
func (s *FCService) CustomDomainStateRefreshFunc(domainName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCCustomDomain(domainName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 custom domains are typically Active when created
		if object.DomainName != nil && *object.DomainName != "" {
			// Custom domains exist, so they are active
			currentState = "Active"
		}

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForCustomDomainCreating waits for custom domain creation to complete
func (s *FCService) WaitForCustomDomainCreating(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}

// WaitForCustomDomainDeleting waits for custom domain deletion to complete
func (s *FCService) WaitForCustomDomainDeleting(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}

// WaitForCustomDomainUpdating waits for custom domain update to complete
func (s *FCService) WaitForCustomDomainUpdating(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}
