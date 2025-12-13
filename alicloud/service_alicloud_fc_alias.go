package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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

// BuildCreateAliasInputFromSchema builds Alias from Terraform schema data
func (s *FCService) BuildCreateAliasInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Alias {
	alias := &aliyunFCAPI.Alias{}

	if v, ok := d.GetOk("alias_name"); ok {
		alias.AliasName = tea.String(v.(string))
	}

	if v, ok := d.GetOk("version_id"); ok {
		alias.VersionId = tea.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		alias.Description = tea.String(v.(string))
	}

	// Add additional version weight
	if v, ok := d.GetOk("additional_version_weight"); ok {
		if weightMap, ok := v.(map[string]interface{}); ok {
			alias.AdditionalVersionWeight = make(map[string]*float32)
			for k, val := range weightMap {
				if weight, ok := val.(float64); ok {
					// Convert float64 to float32
					weight32 := float32(weight)
					alias.AdditionalVersionWeight[k] = tea.Float32(weight32)
				}
			}
		}
	}

	return alias
}

// BuildUpdateAliasInputFromSchema builds Alias for update from Terraform schema data
func (s *FCService) BuildUpdateAliasInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Alias {
	alias := &aliyunFCAPI.Alias{}

	if d.HasChange("version_id") {
		if v, ok := d.GetOk("version_id"); ok {
			alias.VersionId = tea.String(v.(string))
		}
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			alias.Description = tea.String(v.(string))
		}
	}

	if d.HasChange("additional_version_weight") {
		if v, ok := d.GetOk("additional_version_weight"); ok {
			if weightMap, ok := v.(map[string]interface{}); ok {
				alias.AdditionalVersionWeight = make(map[string]*float32)
				for k, val := range weightMap {
					if weight, ok := val.(float64); ok {
						// Convert float64 to float32
						weight32 := float32(weight)
						alias.AdditionalVersionWeight[k] = tea.Float32(weight32)
					}
				}
			}
		}
	}

	return alias
}

// SetSchemaFromAlias sets terraform schema data from Alias
func (s *FCService) SetSchemaFromAlias(d *schema.ResourceData, alias *aliyunFCAPI.Alias) error {
	if alias == nil {
		return fmt.Errorf("alias cannot be nil")
	}

	if alias.AliasName != nil {
		d.Set("alias_name", *alias.AliasName)
	}

	if alias.VersionId != nil {
		d.Set("version_id", *alias.VersionId)
	}

	if alias.Description != nil {
		d.Set("description", *alias.Description)
	}

	if alias.CreatedTime != nil {
		d.Set("create_time", *alias.CreatedTime)
	}

	if alias.LastModifiedTime != nil {
		d.Set("last_modified_time", *alias.LastModifiedTime)
	}

	// Set additional version weight
	if alias.AdditionalVersionWeight != nil {
		weightMap := make(map[string]interface{})
		for k, v := range alias.AdditionalVersionWeight {
			if v != nil {
				weightMap[k] = *v
			}
		}
		d.Set("additional_version_weight", weightMap)
	}

	return nil
}

// AliasStateRefreshFunc returns a StateRefreshFunc to wait for alias status changes
func (s *FCService) AliasStateRefreshFunc(functionName, aliasName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCAlias(functionName, aliasName)
		if err != nil {
			if NotFoundError(err) {
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
