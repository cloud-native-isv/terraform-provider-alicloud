package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Cluster Management Operations

// CreateSelectDBCluster creates a new SelectDB cluster
func (s *SelectDBService) CreateSelectDBCluster(instanceId, clusterClass, cacheSize string, options ...selectdb.ClusterCreateOption) (*selectdb.Cluster, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterClass == "" {
		return nil, WrapError(fmt.Errorf("cluster class cannot be empty"))
	}
	if cacheSize == "" {
		return nil, WrapError(fmt.Errorf("cache size cannot be empty"))
	}

	result, err := s.GetAPI().CreateCluster(instanceId, clusterClass, cacheSize, options...)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBCluster retrieves information about a SelectDB cluster
func (s *SelectDBService) DescribeSelectDBCluster(instanceId, clusterId string) (*selectdb.Cluster, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	// Since there's no direct GetCluster API, we use the config API to check cluster existence
	config, err := s.GetAPI().GetClusterConfig(clusterId, instanceId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Convert config result to cluster info
	cluster := &selectdb.Cluster{
		ClusterId:  clusterId,
		InstanceId: instanceId,
		Config:     config,
	}

	return cluster, nil
}

// ModifySelectDBCluster modifies a SelectDB cluster
func (s *SelectDBService) ModifySelectDBCluster(instanceId, clusterId string, options ...selectdb.ModifyClusterOption) (*selectdb.Cluster, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	result, err := s.GetAPI().ModifyCluster(instanceId, clusterId, options...)
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

	err := s.GetAPI().DeleteCluster(instanceId, clusterId, regionId...)
	if err != nil {
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

	err := s.GetAPI().RestartCluster(instanceId, clusterId, parallelOperation, regionId...)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Cluster Configuration Operations

// DescribeSelectDBClusterConfig retrieves cluster configuration
func (s *SelectDBService) DescribeSelectDBClusterConfig(clusterId, instanceId string, configKey ...string) (*selectdb.ClusterConfig, error) {
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	config, err := s.GetAPI().GetClusterConfig(clusterId, instanceId, configKey...)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return config, nil
}

// DescribeSelectDBClusterConfigChangeLogs retrieves cluster configuration change logs
func (s *SelectDBService) DescribeSelectDBClusterConfigChangeLogs(clusterId, instanceId string) (*selectdb.ClusterConfigChangeLog, error) {
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	logs, err := s.GetAPI().GetClusterConfigChangeLogs(clusterId, instanceId)
	if err != nil {
		return nil, WrapError(err)
	}

	return logs, nil
}

// BE Cluster Operations

// ModifySelectDBBEClusterAttribute modifies BE cluster attributes
func (s *SelectDBService) ModifySelectDBBEClusterAttribute(clusterId, instanceId, attributeType, attributeValue string) error {
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if attributeType == "" {
		return WrapError(fmt.Errorf("attribute type cannot be empty"))
	}
	if attributeValue == "" {
		return WrapError(fmt.Errorf("attribute value cannot be empty"))
	}

	err := s.GetAPI().ModifyBEClusterAttribute(clusterId, instanceId, attributeType, attributeValue)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// State Management and Refresh Functions

// SelectDBClusterStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB cluster
func (s *SelectDBService) SelectDBClusterStateRefreshFunc(instanceId, clusterId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := s.DescribeSelectDBCluster(instanceId, clusterId)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string
		if cluster != nil {
			currentStatus = cluster.Status
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return cluster, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return cluster, currentStatus, nil
	}
}

// WaitForSelectDBCluster waits for SelectDB cluster to reach expected status
func (s *SelectDBService) WaitForSelectDBCluster(instanceId, clusterId string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		cluster, err := s.DescribeSelectDBCluster(instanceId, clusterId)
		if err != nil {
			return WrapError(err)
		}

		if cluster.Status == string(status) {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceId, GetFunc(1), timeout, cluster.Status, string(status), ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper functions for converting between Terraform schema and API types

// ConvertClusterToMap converts API cluster to Terraform map
func ConvertClusterToMap(cluster *selectdb.Cluster) map[string]interface{} {
	if cluster == nil {
		return nil
	}

	result := map[string]interface{}{
		"cluster_id":           cluster.ClusterId,
		"instance_id":          cluster.InstanceId,
		"cluster_name":         cluster.ClusterName,
		"status":               cluster.Status,
		"cluster_class":        cluster.ClusterClass,
		"charge_type":          cluster.ChargeType,
		"cluster_binding":      cluster.ClusterBinding,
		"cpu_cores":            cluster.CpuCores,
		"memory":               cluster.Memory,
		"cache_storage_size":   cluster.CacheStorageSizeGB,
		"cache_storage_type":   cluster.CacheStorageType,
		"performance_level":    cluster.PerformanceLevel,
		"scaling_rules_enable": cluster.ScalingRulesEnable,
		"vswitch_id":           cluster.VSwitchId,
		"zone_id":              cluster.ZoneId,
		"sub_domain":           cluster.SubDomain,
		"created_time":         cluster.CreatedTime,
		"modified_time":        cluster.ModifiedTime,
		"start_time":           cluster.StartTime,
	}

	// Convert configuration parameters
	if cluster.Config != nil && len(cluster.Config.Params) > 0 {
		params := make([]map[string]interface{}, 0)
		for _, param := range cluster.Config.Params {
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
