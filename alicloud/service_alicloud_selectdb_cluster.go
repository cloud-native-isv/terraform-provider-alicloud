package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Cluster Management Operations

// CreateSelectDBCluster creates a new SelectDB cluster
func (s *SelectDBService) CreateSelectDBCluster(options *selectdb.CreateClusterOptions) (*selectdb.CreateClusterResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("create cluster options cannot be nil"))
	}

	result, err := s.api.CreateCluster(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBCluster retrieves information about a SelectDB cluster
func (s *SelectDBService) DescribeSelectDBCluster(instanceId, clusterId string) (*selectdb.DBCluster, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	// Since there's no direct GetCluster API, we use the config API to check cluster existence
	config, err := s.api.GetClusterConfig(&selectdb.ClusterConfigQuery{
		DBInstanceId: instanceId,
		DBClusterId:  clusterId,
	})
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Convert config result to cluster info
	cluster := &selectdb.DBCluster{
		DBClusterId:  config.DbClusterId,
		DBInstanceId: config.DbInstanceId,
		// Other fields would need to be populated from additional API calls if needed
	}

	return cluster, nil
}

// ModifySelectDBCluster modifies a SelectDB cluster
func (s *SelectDBService) ModifySelectDBCluster(options *selectdb.ModifyClusterOptions) (*selectdb.ModifyClusterResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("modify cluster options cannot be nil"))
	}

	result, err := s.api.ModifyCluster(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteSelectDBCluster deletes a SelectDB cluster
func (s *SelectDBService) DeleteSelectDBCluster(instanceId, clusterId string, regionId ...string) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	err := s.api.DeleteCluster(instanceId, clusterId, regionId...)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil // Cluster already deleted
		}
		return WrapError(err)
	}

	return nil
}

// RestartSelectDBCluster restarts a SelectDB cluster
func (s *SelectDBService) RestartSelectDBCluster(instanceId, clusterId string, parallelOperation bool, regionId ...string) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	err := s.api.RestartCluster(instanceId, clusterId, parallelOperation, regionId...)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Cluster Binding Operations

// CreateSelectDBClusterBinding creates a cluster binding
func (s *SelectDBService) CreateSelectDBClusterBinding(options *selectdb.ClusterBindingOptions) (*selectdb.ClusterBindingResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("cluster binding options cannot be nil"))
	}

	result, err := s.api.CreateClusterBinding(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteSelectDBClusterBinding deletes a cluster binding
func (s *SelectDBService) DeleteSelectDBClusterBinding(options *selectdb.ClusterBindingOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("cluster binding options cannot be nil"))
	}

	err := s.api.DeleteClusterBinding(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Cluster Configuration Operations

// DescribeSelectDBClusterConfig retrieves cluster configuration
func (s *SelectDBService) DescribeSelectDBClusterConfig(id *selectdb.ClusterConfigQuery) (*selectdb.ClusterConfig, error) {
	if query == nil {
		return nil, WrapError(fmt.Errorf("cluster config query cannot be nil"))
	}

	config, err := s.api.GetClusterConfig(id)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return config, nil
}

// ModifySelectDBClusterConfig modifies cluster configuration
func (s *SelectDBService) ModifySelectDBClusterConfig(modification *selectdb.ClusterConfigModification) (*selectdb.ClusterConfig, error) {
	if modification == nil {
		return nil, WrapError(fmt.Errorf("cluster config modification cannot be nil"))
	}

	result, err := s.api.ModifyClusterConfig(modification)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBClusterConfigChangeLogs retrieves cluster configuration change logs
func (s *SelectDBService) DescribeSelectDBClusterConfigChangeLogs(query *selectdb.ClusterConfigChangeLogsQuery) (*selectdb.ClusterConfigChangeLog, error) {
	if query == nil {
		return nil, WrapError(fmt.Errorf("cluster config change logs query cannot be nil"))
	}

	logs, err := s.api.GetClusterConfigChangeLogs(query)
	if err != nil {
		return nil, WrapError(err)
	}

	return logs, nil
}

// Service Linked Role Operations

// CheckSelectDBServiceLinkedRole checks if service linked role exists
func (s *SelectDBService) CheckSelectDBServiceLinkedRole(options *selectdb.ServiceLinkedRoleOptions) (*selectdb.ServiceLinkedRoleResult, error) {
	result, err := s.api.CheckServiceLinkedRole(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// CreateSelectDBServiceLinkedRole creates service linked role for SelectDB
func (s *SelectDBService) CreateSelectDBServiceLinkedRole(options *selectdb.ServiceLinkedRoleOptions) error {
	err := s.api.CreateServiceLinkedRoleForSelectDB(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// BE Cluster Operations

// StartSelectDBBECluster starts a BE cluster
func (s *SelectDBService) StartSelectDBBECluster(options *selectdb.BEClusterOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("BE cluster options cannot be nil"))
	}

	err := s.api.StartBECluster(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// StopSelectDBBECluster stops a BE cluster
func (s *SelectDBService) StopSelectDBBECluster(options *selectdb.BEClusterOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("BE cluster options cannot be nil"))
	}

	err := s.api.StopBECluster(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// ModifySelectDBBEClusterAttribute modifies BE cluster attributes
func (s *SelectDBService) ModifySelectDBBEClusterAttribute(modification *selectdb.BEClusterAttributeModification) error {
	if modification == nil {
		return WrapError(fmt.Errorf("BE cluster attribute modification cannot be nil"))
	}

	err := s.api.ModifyBEClusterAttribute(modification)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Price Inquiry Operations

// GetCreateSelectDBBEClusterInquiry gets pricing information for creating BE cluster
func (s *SelectDBService) GetCreateSelectDBBEClusterInquiry(options *selectdb.BEClusterInquiryOptions) (*selectdb.InquiryResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("BE cluster inquiry options cannot be nil"))
	}

	result, err := s.api.GetCreateBEClusterInquiry(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// GetModifySelectDBBEClusterInquiry gets pricing information for modifying BE cluster
func (s *SelectDBService) GetModifySelectDBBEClusterInquiry(options *selectdb.BEClusterInquiryOptions) (*selectdb.InquiryResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("BE cluster inquiry options cannot be nil"))
	}

	result, err := s.api.GetModifyBEClusterInquiry(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// State Management and Refresh Functions

// SelectDBClusterStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB cluster
func (s *SelectDBService) SelectDBClusterStateRefreshFunc(instanceId, clusterId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := s.DescribeSelectDBCluster(instanceId, clusterId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", WrapErrorf(Error(GetNotFoundMessage("SelectDB Cluster", clusterId)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if cluster.Status == failState {
				return cluster, cluster.Status, WrapError(Error(FailedToReachTargetStatus, cluster.Status))
			}
		}

		return cluster, cluster.Status, nil
	}
}

// WaitForSelectDBCluster waits for SelectDB cluster to reach expected status
func (s *SelectDBService) WaitForSelectDBCluster(instanceId, clusterId string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		cluster, err := s.DescribeSelectDBCluster(instanceId, clusterId)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if cluster != nil && cluster.Status == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, clusterId, GetFunc(1), timeout, cluster.Status, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper functions for converting between Terraform schema and API types

// ConvertToCreateClusterOptions converts schema data to API create cluster options
func ConvertToCreateClusterOptions(d *schema.ResourceData) *selectdb.CreateClusterOptions {
	options := &selectdb.CreateClusterOptions{}

	if v, ok := d.GetOk("db_instance_id"); ok {
		options.DBInstanceId = v.(string)
	}
	if v, ok := d.GetOk("db_cluster_description"); ok {
		options.DBClusterDescription = v.(string)
	}
	if v, ok := d.GetOk("db_cluster_class"); ok {
		options.DBClusterClass = v.(string)
	}
	if v, ok := d.GetOk("engine"); ok {
		options.Engine = v.(string)
	}
	if v, ok := d.GetOk("engine_version"); ok {
		options.EngineVersion = v.(string)
	}
	if v, ok := d.GetOk("charge_type"); ok {
		options.ChargeType = v.(string)
	}
	if v, ok := d.GetOk("period"); ok {
		options.Period = v.(string)
	}
	if v, ok := d.GetOk("used_time"); ok {
		options.UsedTime = v.(string)
	}
	if v, ok := d.GetOk("cache_size"); ok {
		options.CacheSize = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}
	if v, ok := d.GetOk("zone_id"); ok {
		options.ZoneId = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		options.VpcId = v.(string)
	}
	if v, ok := d.GetOk("vswitch_id"); ok {
		options.VSwitchId = v.(string)
	}

	return options
}

// ConvertToModifyClusterOptions converts schema data to API modify cluster options
func ConvertToModifyClusterOptions(d *schema.ResourceData, instanceId, clusterId string) *selectdb.ModifyClusterOptions {
	options := &selectdb.ModifyClusterOptions{
		DBInstanceId: instanceId,
		DBClusterId:  clusterId,
	}

	if v, ok := d.GetOk("db_cluster_class"); ok {
		options.DBClusterClass = v.(string)
	}
	if v, ok := d.GetOk("cache_size"); ok {
		options.CacheSize = v.(string)
	}
	if v, ok := d.GetOk("engine"); ok {
		options.Engine = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}

// ConvertClusterToMap converts API cluster to Terraform map
func ConvertClusterToMap(cluster *selectdb.DBCluster) map[string]interface{} {
	if cluster == nil {
		return nil
	}

	result := map[string]interface{}{
		"db_cluster_id":    cluster.DBClusterId,
		"db_instance_id":   cluster.DBInstanceId,
		"db_cluster_class": cluster.DBClusterClass,
		"description":      cluster.Description,
		"engine":           cluster.Engine,
		"engine_version":   cluster.EngineVersion,
		"status":           cluster.Status,
		"charge_type":      cluster.ChargeType,
		"cache_size":       cluster.CacheSize,
		"create_time":      cluster.CreateTime,
		"modify_time":      cluster.ModifyTime,
		"order_id":         cluster.OrderId,
	}

	return result
}

// ConvertToClusterConfigQuery converts schema data to API cluster config query
func ConvertToClusterConfigQuery(d *schema.ResourceData, instanceId, clusterId string) *selectdb.ClusterConfigQuery {
	query := &selectdb.ClusterConfigQuery{
		DBInstanceId: instanceId,
		DBClusterId:  clusterId,
	}

	if v, ok := d.GetOk("config_key"); ok {
		query.ConfigKey = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		query.RegionId = v.(string)
	}

	return query
}

// ConvertToClusterConfigModification converts schema data to API cluster config modification
func ConvertToClusterConfigModification(d *schema.ResourceData, instanceId, clusterId string) *selectdb.ClusterConfigModification {
	modification := &selectdb.ClusterConfigModification{
		DBInstanceId: instanceId,
		DBClusterId:  clusterId,
	}

	if v, ok := d.GetOk("parameters"); ok {
		modification.Parameters = v.(string)
	}
	if v, ok := d.GetOk("config_key"); ok {
		modification.ConfigKey = v.(string)
	}
	if v, ok := d.GetOk("switch_time_mode"); ok {
		modification.SwitchTimeMode = v.(string)
	}
	if v, ok := d.GetOk("parallel_operation"); ok {
		modification.ParallelOperation = v.(bool)
	}
	if v, ok := d.GetOk("region_id"); ok {
		modification.RegionId = v.(string)
	}

	return modification
}

// ConvertClusterConfigToMap converts API cluster config to Terraform map
func ConvertClusterConfigToMap(config *selectdb.ClusterConfig) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"db_cluster_id":    config.DbClusterId,
		"db_instance_id":   config.DbInstanceId,
		"db_instance_name": config.DbInstanceName,
		"task_id":          config.TaskId,
	}

	// Convert parameters
	if len(config.Params) > 0 {
		params := make([]map[string]interface{}, 0, len(config.Params))
		for _, param := range config.Params {
			p := map[string]interface{}{
				"name":               param.Name,
				"value":              param.Value,
				"default_value":      param.DefaultValue,
				"comment":            param.Comment,
				"is_dynamic":         param.IsDynamic,
				"is_user_modifiable": param.IsUserModifiable,
				"optional":           param.Optional,
				"param_category":     param.ParamCategory,
			}
			params = append(params, p)
		}
		result["params"] = params
	}

	return result
}
