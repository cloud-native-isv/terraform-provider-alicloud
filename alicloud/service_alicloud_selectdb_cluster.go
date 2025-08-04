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
		Status:     selectdb.ClusterStatusActivation, // Default status if we can get config
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

// WaitForSelectDBClusterCreated waits for SelectDB cluster to be created and active
func (s *SelectDBService) WaitForSelectDBClusterCreated(instanceId, clusterId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			selectdb.ClusterStatusCreating,
			selectdb.ClusterStatusOrderPreparing,
		},
		Target: []string{selectdb.ClusterStatusActivation},
		Refresh: s.SelectDBClusterStateRefreshFunc(instanceId, clusterId, []string{
			"FAILED", "ERROR", "EXCEPTION",
		}),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
}

// WaitForSelectDBClusterUpdated waits for SelectDB cluster update operations to complete
func (s *SelectDBService) WaitForSelectDBClusterUpdated(instanceId, clusterId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			selectdb.ClusterStatusResourceChanging,
			selectdb.ClusterStatusReadonlyResourceChanging,
			selectdb.ClusterStatusOrderPreparing,
		},
		Target: []string{selectdb.ClusterStatusActivation},
		Refresh: s.SelectDBClusterStateRefreshFunc(instanceId, clusterId, []string{
			"FAILED", "ERROR", "EXCEPTION",
		}),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
}

// WaitForSelectDBClusterDeleted waits for SelectDB cluster to be deleted
func (s *SelectDBService) WaitForSelectDBClusterDeleted(instanceId, clusterId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			selectdb.ClusterStatusActivation,
			selectdb.ClusterStatusDeleting,
			selectdb.ClusterStatusResourceChanging,
		},
		Target: []string{},
		Refresh: func() (interface{}, string, error) {
			cluster, err := s.DescribeSelectDBCluster(instanceId, clusterId)
			if err != nil {
				if IsNotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}

			// Check for failed states
			failStates := []string{"FAILED", "ERROR", "EXCEPTION"}
			for _, failState := range failStates {
				if cluster.Status == failState {
					return cluster, cluster.Status, WrapError(Error(FailedToReachTargetStatus, cluster.Status))
				}
			}

			return cluster, cluster.Status, nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
}
