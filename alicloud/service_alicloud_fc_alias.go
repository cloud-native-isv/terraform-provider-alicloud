package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Alias methods for FCService

// EncodeAliasResourceId encodes function name and alias name into a resource ID string
// Format: functionName:aliasName
func EncodeAliasResourceId(functionName, aliasName string) string {
	return fmt.Sprintf("%s:%s", functionName, aliasName)
}

// DecodeAliasResourceId decodes alias resource ID string to function name and alias name
func DecodeAliasResourceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid alias resource ID format, expected functionName:aliasName, got %s", id)
	}
	return parts[0], parts[1], nil
}

// DescribeFCAlias retrieves alias information by function name and alias name
func (s *FCService) DescribeFCAlias(functionName, aliasName string) (*aliyunFCAPI.Alias, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if aliasName == "" {
		return nil, fmt.Errorf("alias name cannot be empty")
	}
	return s.GetAPI().GetAlias(functionName, aliasName)
}

// ListFCAliases lists all aliases for a function with optional filters
func (s *FCService) ListFCAliases(functionName string, options *aliyunFCAPI.AliasQueryOptions) ([]*aliyunFCAPI.Alias, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	result, err := s.GetAPI().ListAliases(functionName, options)
	if err != nil {
		return nil, err
	}
	return result.Aliases, nil
}

// CreateFCAlias creates a new FC alias
func (s *FCService) CreateFCAlias(functionName string, alias *aliyunFCAPI.Alias) (*aliyunFCAPI.Alias, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if alias == nil {
		return nil, fmt.Errorf("alias cannot be nil")
	}
	return s.GetAPI().CreateAlias(functionName, alias)
}

// UpdateFCAlias updates an existing FC alias
func (s *FCService) UpdateFCAlias(functionName, aliasName string, alias *aliyunFCAPI.Alias) (*aliyunFCAPI.Alias, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if aliasName == "" {
		return nil, fmt.Errorf("alias name cannot be empty")
	}
	if alias == nil {
		return nil, fmt.Errorf("alias cannot be nil")
	}
	return s.GetAPI().UpdateAlias(functionName, aliasName, alias)
}

// DeleteFCAlias deletes an FC alias
func (s *FCService) DeleteFCAlias(functionName, aliasName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if aliasName == "" {
		return fmt.Errorf("alias name cannot be empty")
	}
	return s.GetAPI().DeleteAlias(functionName, aliasName)
}

// AliasStateRefreshFunc returns a StateRefreshFunc to wait for alias status changes
func (s *FCService) AliasStateRefreshFunc(functionName, aliasName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCAlias(functionName, aliasName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 aliases are typically Active when created
		// FC v3 aliases don't typically have a status field, so we assume they're active

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForAliasCreating waits for alias creation to complete
func (s *FCService) WaitForAliasCreating(functionName, aliasName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.AliasStateRefreshFunc(functionName, aliasName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, aliasName))
}

// WaitForAliasDeleting waits for alias deletion to complete
func (s *FCService) WaitForAliasDeleting(functionName, aliasName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.AliasStateRefreshFunc(functionName, aliasName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, aliasName))
}

// WaitForAliasUpdating waits for alias update to complete
func (s *FCService) WaitForAliasUpdating(functionName, aliasName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.AliasStateRefreshFunc(functionName, aliasName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, aliasName))
}
