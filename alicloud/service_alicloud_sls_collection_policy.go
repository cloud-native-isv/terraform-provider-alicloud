package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Collection Policy CRUD Operations
// ===========================================

// CreateSlsCollectionPolicy creates a new SLS collection policy
func (s *SlsService) CreateSlsCollectionPolicy(policy *aliyunSlsAPI.CollectionPolicy) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpsertCollectionPolicy(policy)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, policy.PolicyName, "CreateSlsCollectionPolicy", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateSlsCollectionPolicy updates an existing SLS collection policy
func (s *SlsService) UpdateSlsCollectionPolicy(policy *aliyunSlsAPI.CollectionPolicy) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpsertCollectionPolicy(policy)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, policy.PolicyName, "UpdateSlsCollectionPolicy", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DescribeSlsCollectionPolicy retrieves a collection policy by name
func (s *SlsService) DescribeSlsCollectionPolicy(policyName string) (*aliyunSlsAPI.CollectionPolicy, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	policy, err := s.aliyunSlsAPI.GetCollectionPolicy(policyName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(Error(GetNotFoundMessage("SlsCollectionPolicy", policyName)), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, policyName, "DescribeSlsCollectionPolicy", AlibabaCloudSdkGoERROR)
	}

	return policy, nil
}

// DeleteSlsCollectionPolicy deletes a collection policy
func (s *SlsService) DeleteSlsCollectionPolicy(policyName string, dataCode string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteCollectionPolicy(policyName, dataCode)
	if err != nil {
		if IsExpectedErrors(err, []string{"ProjectNotExist", "PolicyNotExist", "ResourceNotExist"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, policyName, "DeleteSlsCollectionPolicy", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ListSlsCollectionPolicies lists collection policies with filters
func (s *SlsService) ListSlsCollectionPolicies(policyName, instanceId, centralProject, dataCode, productCode string) ([]*aliyunSlsAPI.CollectionPolicy, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	policies, err := s.aliyunSlsAPI.ListCollectionPolicies(policyName, instanceId, centralProject, dataCode, productCode)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "collection_policies", "ListSlsCollectionPolicies", AlibabaCloudSdkGoERROR)
	}

	return policies, nil
}

// SlsCollectionPolicyStateRefreshFunc returns a StateRefreshFunc for collection policy state monitoring
func (s *SlsService) SlsCollectionPolicyStateRefreshFunc(policyName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		policy, err := s.DescribeSlsCollectionPolicy(policyName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Check if policy is in a failed state
		for _, failState := range failStates {
			if strings.Contains(fmt.Sprint(policy), failState) {
				return policy, failState, WrapError(Error(FailedToReachTargetStatus, failState))
			}
		}

		// Determine current status based on enabled state
		currentStatus := "Available"
		if !policy.Enabled {
			currentStatus = "Disabled"
		}

		return policy, currentStatus, nil
	}
}

// CheckSlsCollectionPolicyExists checks if a collection policy exists
func (s *SlsService) CheckSlsCollectionPolicyExists(policyName string) (bool, error) {
	_, err := s.DescribeSlsCollectionPolicy(policyName)
	if err != nil {
		if IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// WaitForSlsCollectionPolicy waits for a collection policy to reach the target state
func (s *SlsService) WaitForSlsCollectionPolicy(policyName string, status Status, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		policy, err := s.DescribeSlsCollectionPolicy(policyName)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if policy != nil {
			switch status {
			case Available:
				if policy.Enabled {
					return nil
				}
			case Deleted:
				return nil // Policy was successfully deleted
			default:
				if !policy.Enabled {
					return nil
				}
			}
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, policyName, GetFunc(1), timeout, "", policyName, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort)
	}
}

// Helper functions for converting between Terraform and SLS API types

// ConvertToSlsCollectionPolicy converts Terraform configuration to SLS CollectionPolicy
func ConvertToSlsCollectionPolicy(terraformConfig map[string]interface{}) *aliyunSlsAPI.CollectionPolicy {
	policy := &aliyunSlsAPI.CollectionPolicy{}

	if v, ok := terraformConfig["policy_name"].(string); ok {
		policy.PolicyName = v
	}
	if v, ok := terraformConfig["data_code"].(string); ok {
		policy.DataCode = v
	}
	if v, ok := terraformConfig["product_code"].(string); ok {
		policy.ProductCode = v
	}
	if v, ok := terraformConfig["enabled"].(bool); ok {
		policy.Enabled = v
	}
	if v, ok := terraformConfig["centralize_enabled"].(bool); ok {
		policy.CentralizeEnabled = v
	}

	// Convert policy configuration
	if policyConfigInterface, ok := terraformConfig["policy_config"]; ok {
		if policyConfigList, ok := policyConfigInterface.([]interface{}); ok && len(policyConfigList) > 0 {
			if policyConfigMap, ok := policyConfigList[0].(map[string]interface{}); ok {
				policy.PolicyConfig = &aliyunSlsAPI.CollectionPolicyPolicyConfig{}
				if v, ok := policyConfigMap["resource_mode"].(string); ok {
					policy.PolicyConfig.ResourceMode = v
				}
				if v, ok := policyConfigMap["regions"].([]interface{}); ok {
					regions := make([]string, len(v))
					for i, region := range v {
						regions[i] = region.(string)
					}
					policy.PolicyConfig.Regions = regions
				}
				if v, ok := policyConfigMap["instance_ids"].([]interface{}); ok {
					instanceIds := make([]string, len(v))
					for i, id := range v {
						instanceIds[i] = id.(string)
					}
					policy.PolicyConfig.InstanceIds = instanceIds
				}
				if v, ok := policyConfigMap["resource_tags"].(map[string]interface{}); ok {
					policy.PolicyConfig.ResourceTags = v
				}
			}
		}
	}

	// Convert data configuration
	if dataConfigInterface, ok := terraformConfig["data_config"]; ok {
		if dataConfigList, ok := dataConfigInterface.([]interface{}); ok && len(dataConfigList) > 0 {
			if dataConfigMap, ok := dataConfigList[0].(map[string]interface{}); ok {
				policy.DataConfig = &aliyunSlsAPI.CollectionPolicyDataConfig{}
				if v, ok := dataConfigMap["data_project"].(string); ok {
					policy.DataConfig.DataProject = v
				}
				if v, ok := dataConfigMap["data_region"].(string); ok {
					policy.DataConfig.DataRegion = v
				}
			}
		}
	}

	// Convert centralize configuration
	if centralizeConfigInterface, ok := terraformConfig["centralize_config"]; ok {
		if centralizeConfigList, ok := centralizeConfigInterface.([]interface{}); ok && len(centralizeConfigList) > 0 {
			if centralizeConfigMap, ok := centralizeConfigList[0].(map[string]interface{}); ok {
				policy.CentralizeConfig = &aliyunSlsAPI.CollectionPolicyCentralizeConfig{}
				if v, ok := centralizeConfigMap["dest_project"].(string); ok {
					policy.CentralizeConfig.DestProject = v
				}
				if v, ok := centralizeConfigMap["dest_logstore"].(string); ok {
					policy.CentralizeConfig.DestLogstore = v
				}
				if v, ok := centralizeConfigMap["dest_region"].(string); ok {
					policy.CentralizeConfig.DestRegion = v
				}
				if v, ok := centralizeConfigMap["dest_ttl"].(int); ok {
					policy.CentralizeConfig.DestTTL = int32(v)
				}
			}
		}
	}

	// Convert resource directory
	if resourceDirInterface, ok := terraformConfig["resource_directory"]; ok {
		if resourceDirList, ok := resourceDirInterface.([]interface{}); ok && len(resourceDirList) > 0 {
			if resourceDirMap, ok := resourceDirList[0].(map[string]interface{}); ok {
				policy.ResourceDirectory = &aliyunSlsAPI.CollectionPolicyResourceDirectory{}
				if v, ok := resourceDirMap["account_group_type"].(string); ok {
					policy.ResourceDirectory.AccountGroupType = v
				}
				if v, ok := resourceDirMap["members"].([]interface{}); ok {
					members := make([]string, len(v))
					for i, member := range v {
						members[i] = member.(string)
					}
					policy.ResourceDirectory.Members = members
				}
			}
		}
	}

	return policy
}

// ConvertCollectionPolicyToMap converts CollectionPolicy struct to map for Terraform compatibility
func ConvertCollectionPolicyToMap(policy *aliyunSlsAPI.CollectionPolicy) map[string]interface{} {
	if policy == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["policy_name"] = policy.PolicyName
	result["policy_uid"] = policy.PolicyUid
	result["data_code"] = policy.DataCode
	result["product_code"] = policy.ProductCode
	result["enabled"] = policy.Enabled
	result["centralize_enabled"] = policy.CentralizeEnabled
	result["internal_policy"] = policy.InternalPolicy
	result["create_time"] = policy.CreateTime
	result["last_modify_time"] = policy.LastModifyTime

	// Handle policy configuration
	if policy.PolicyConfig != nil {
		policyConfig := make(map[string]interface{})
		policyConfig["resource_mode"] = policy.PolicyConfig.ResourceMode
		policyConfig["regions"] = policy.PolicyConfig.Regions
		policyConfig["instance_ids"] = policy.PolicyConfig.InstanceIds
		policyConfig["resource_tags"] = policy.PolicyConfig.ResourceTags
		result["policy_config"] = []interface{}{policyConfig}
	}

	// Handle data configuration
	if policy.DataConfig != nil {
		dataConfig := make(map[string]interface{})
		dataConfig["data_project"] = policy.DataConfig.DataProject
		dataConfig["data_region"] = policy.DataConfig.DataRegion
		result["data_config"] = []interface{}{dataConfig}
	}

	// Handle centralize configuration
	if policy.CentralizeConfig != nil {
		centralizeConfig := make(map[string]interface{})
		centralizeConfig["dest_project"] = policy.CentralizeConfig.DestProject
		centralizeConfig["dest_logstore"] = policy.CentralizeConfig.DestLogstore
		centralizeConfig["dest_region"] = policy.CentralizeConfig.DestRegion
		centralizeConfig["dest_ttl"] = policy.CentralizeConfig.DestTTL
		result["centralize_config"] = []interface{}{centralizeConfig}
	}

	// Handle resource directory
	if policy.ResourceDirectory != nil {
		resourceDir := make(map[string]interface{})
		resourceDir["account_group_type"] = policy.ResourceDirectory.AccountGroupType
		resourceDir["members"] = policy.ResourceDirectory.Members
		result["resource_directory"] = []interface{}{resourceDir}
	}

	return result
}
