package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Elastic Rule Management Operations

// CreateSelectDBElasticRule creates a new SelectDB elastic scaling rule
func (s *SelectDBService) CreateSelectDBElasticRule(options *selectdb.CreateElasticRuleOptions) (*selectdb.CreateElasticRuleResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("create elastic rule options cannot be nil"))
	}

	result, err := s.selectdbAPI.CreateElasticRule(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBElasticRule retrieves information about a SelectDB elastic rule
func (s *SelectDBService) DescribeSelectDBElasticRule(dbClusterId, dbInstanceId, product string, ruleId int64) (*selectdb.ElasticRule, error) {
	if dbClusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if dbInstanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if product == "" {
		return nil, WrapError(fmt.Errorf("product cannot be empty"))
	}

	options := &selectdb.ListElasticRulesOptions{
		DbClusterId:  dbClusterId,
		DbInstanceId: dbInstanceId,
		Product:      product,
		RuleId:       ruleId,
	}

	result, err := s.selectdbAPI.ListElasticRules(options)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Find the specific rule by ID
	for _, rule := range result.Rules {
		if rule.RuleId == ruleId {
			return &rule, nil
		}
	}

	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Elastic Rule", fmt.Sprintf("%d", ruleId))), NotFoundMsg, ProviderERROR)
}

// DescribeSelectDBElasticRules lists SelectDB elastic scaling rules
func (s *SelectDBService) DescribeSelectDBElasticRules(options *selectdb.ListElasticRulesOptions) (*selectdb.ListElasticRulesResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("list elastic rules options cannot be nil"))
	}

	result, err := s.selectdbAPI.ListElasticRules(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// ModifySelectDBElasticRule modifies a SelectDB elastic scaling rule
func (s *SelectDBService) ModifySelectDBElasticRule(options *selectdb.ModifyElasticRuleOptions) (*selectdb.ModifyElasticRuleResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("modify elastic rule options cannot be nil"))
	}

	result, err := s.selectdbAPI.ModifyElasticRule(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteSelectDBElasticRule deletes a SelectDB elastic scaling rule
func (s *SelectDBService) DeleteSelectDBElasticRule(dbClusterId, dbInstanceId, product string, ruleId int64, regionId ...string) error {
	if dbClusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if dbInstanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if product == "" {
		return WrapError(fmt.Errorf("product cannot be empty"))
	}
	if ruleId <= 0 {
		return WrapError(fmt.Errorf("rule ID must be positive"))
	}

	err := s.selectdbAPI.DeleteElasticRule(dbClusterId, dbInstanceId, product, ruleId, regionId...)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil // Rule already deleted
		}
		return WrapError(err)
	}

	return nil
}

// EnableSelectDBElasticRule enables or disables a SelectDB elastic scaling rule
func (s *SelectDBService) EnableSelectDBElasticRule(options *selectdb.EnableDisableElasticRuleOptions) (*selectdb.EnableDisableElasticRuleResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("enable/disable elastic rule options cannot be nil"))
	}

	result, err := s.selectdbAPI.EnableDisableElasticRule(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// State Management and Refresh Functions

// SelectDBElasticRuleStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB elastic rule
func (s *SelectDBService) SelectDBElasticRuleStateRefreshFunc(dbClusterId, dbInstanceId, product string, ruleId int64, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rule, err := s.DescribeSelectDBElasticRule(dbClusterId, dbInstanceId, product, ruleId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", WrapErrorf(Error(GetNotFoundMessage("SelectDB Elastic Rule", fmt.Sprintf("%d", ruleId))), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if rule.Status == failState {
				return rule, rule.Status, WrapError(Error(FailedToReachTargetStatus, rule.Status))
			}
		}

		return rule, rule.Status, nil
	}
}

// WaitForSelectDBElasticRule waits for SelectDB elastic rule to reach expected status
func (s *SelectDBService) WaitForSelectDBElasticRule(dbClusterId, dbInstanceId, product string, ruleId int64, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		rule, err := s.DescribeSelectDBElasticRule(dbClusterId, dbInstanceId, product, ruleId)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if rule != nil && rule.Status == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, fmt.Sprintf("%d", ruleId), GetFunc(1), timeout, rule.Status, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper functions for converting between Terraform schema and API types

// ConvertToCreateElasticRuleOptions converts schema data to API create elastic rule options
func ConvertToCreateElasticRuleOptions(d *schema.ResourceData) *selectdb.CreateElasticRuleOptions {
	options := &selectdb.CreateElasticRuleOptions{}

	if v, ok := d.GetOk("db_cluster_id"); ok {
		options.DbClusterId = v.(string)
	}
	if v, ok := d.GetOk("db_instance_id"); ok {
		options.DbInstanceId = v.(string)
	}
	if v, ok := d.GetOk("product"); ok {
		options.Product = v.(string)
	}
	if v, ok := d.GetOk("rule_name"); ok {
		options.RuleName = v.(string)
	}
	if v, ok := d.GetOk("cluster_class"); ok {
		options.ClusterClass = v.(string)
	}
	if v, ok := d.GetOk("elastic_rule_start_time"); ok {
		options.ElasticRuleStartTime = v.(string)
	}
	if v, ok := d.GetOk("execution_period"); ok {
		options.ExecutionPeriod = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}

// ConvertToModifyElasticRuleOptions converts schema data to API modify elastic rule options
func ConvertToModifyElasticRuleOptions(d *schema.ResourceData, ruleId int64) *selectdb.ModifyElasticRuleOptions {
	options := &selectdb.ModifyElasticRuleOptions{
		RuleId: ruleId,
	}

	if v, ok := d.GetOk("db_cluster_id"); ok {
		options.DbClusterId = v.(string)
	}
	if v, ok := d.GetOk("db_instance_id"); ok {
		options.DbInstanceId = v.(string)
	}
	if v, ok := d.GetOk("product"); ok {
		options.Product = v.(string)
	}
	if v, ok := d.GetOk("rule_name"); ok {
		options.RuleName = v.(string)
	}
	if v, ok := d.GetOk("cluster_class"); ok {
		options.ClusterClass = v.(string)
	}
	if v, ok := d.GetOk("elastic_rule_start_time"); ok {
		options.ElasticRuleStartTime = v.(string)
	}
	if v, ok := d.GetOk("execution_period"); ok {
		options.ExecutionPeriod = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}

// ConvertElasticRuleToMap converts API elastic rule to Terraform map
func ConvertElasticRuleToMap(rule *selectdb.ElasticRule) map[string]interface{} {
	if rule == nil {
		return nil
	}

	result := map[string]interface{}{
		"rule_id":                 rule.RuleId,
		"db_cluster_id":           rule.DbClusterId,
		"db_instance_id":          rule.DbInstanceId,
		"product":                 rule.Product,
		"rule_name":               rule.RuleName,
		"cluster_class":           rule.ClusterClass,
		"pre_cluster_class":       rule.PreClusterClass,
		"elastic_rule_start_time": rule.ElasticRuleStartTime,
		"execution_period":        rule.ExecutionPeriod,
		"status":                  rule.Status,
		"enabled":                 rule.Enabled,
	}

	return result
}

// ConvertToEnableDisableElasticRuleOptions converts schema data to API enable/disable elastic rule options
func ConvertToEnableDisableElasticRuleOptions(d *schema.ResourceData, ruleId int64, enable bool) *selectdb.EnableDisableElasticRuleOptions {
	options := &selectdb.EnableDisableElasticRuleOptions{
		RuleId: ruleId,
		Enable: enable,
	}

	if v, ok := d.GetOk("db_cluster_id"); ok {
		options.DbClusterId = v.(string)
	}
	if v, ok := d.GetOk("db_instance_id"); ok {
		options.DbInstanceId = v.(string)
	}
	if v, ok := d.GetOk("product"); ok {
		options.Product = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}

// ConvertToListElasticRulesOptions converts schema data to API list elastic rules options
func ConvertToListElasticRulesOptions(d *schema.ResourceData) *selectdb.ListElasticRulesOptions {
	options := &selectdb.ListElasticRulesOptions{}

	if v, ok := d.GetOk("db_cluster_id"); ok {
		options.DbClusterId = v.(string)
	}
	if v, ok := d.GetOk("db_instance_id"); ok {
		options.DbInstanceId = v.(string)
	}
	if v, ok := d.GetOk("product"); ok {
		options.Product = v.(string)
	}
	if v, ok := d.GetOk("rule_id"); ok {
		options.RuleId = int64(v.(int))
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}
