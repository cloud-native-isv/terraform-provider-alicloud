package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Cluster Management Operations

// CreateSelectDBCluster creates a new SelectDB cluster
func (s *SelectDBService) CreateSelectDBCluster(cluster *selectdb.Cluster) (*selectdb.Cluster, error) {
	if cluster == nil {
		return nil, WrapError(fmt.Errorf("cluster cannot be nil"))
	}
	if cluster.InstanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if cluster.ClusterClass == "" {
		return nil, WrapError(fmt.Errorf("cluster class cannot be empty"))
	}
	if cluster.CacheSize <= 0 {
		return nil, WrapError(fmt.Errorf("cache size must be greater than 0"))
	}

	result, err := s.GetAPI().CreateCluster(cluster)
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

	// Get instance information to find the cluster in DBClusterList
	instance, err := s.GetAPI().GetInstance(instanceId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Search for the cluster in the instance's DBClusterList
	for _, cluster := range instance.DBClusterList {
		if cluster.ClusterId == clusterId {
			return &cluster, nil
		}
	}

	// Cluster not found in the instance
	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Cluster", clusterId)), NotFoundMsg, ProviderERROR)
}

// ModifySelectDBCluster modifies a SelectDB cluster
func (s *SelectDBService) ModifySelectDBCluster(cluster *selectdb.Cluster) (*selectdb.Cluster, error) {
	if cluster == nil {
		return nil, WrapError(fmt.Errorf("cluster cannot be nil"))
	}
	if cluster.InstanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if cluster.ClusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	result, err := s.GetAPI().ModifyCluster(cluster)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteSelectDBCluster deletes a SelectDB cluster
func (s *SelectDBService) DeleteSelectDBCluster(instanceId, clusterId string) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	err := s.GetAPI().DeleteCluster(instanceId, clusterId, s.GetRegionId())
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// RestartSelectDBCluster restarts a SelectDB cluster
func (s *SelectDBService) RestartSelectDBCluster(instanceId, clusterId string, parallelOperation bool) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}

	err := s.GetAPI().RestartCluster(instanceId, clusterId, parallelOperation, s.GetRegionId())
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Cluster Configuration Operations

// DescribeSelectDBClusterConfig retrieves cluster configuration parameters
func (s *SelectDBService) DescribeSelectDBClusterConfig(clusterId, instanceId string, configKey ...string) ([]selectdb.ClusterConfigParam, error) {
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	params, err := s.GetAPI().GetClusterConfig(clusterId, instanceId, configKey...)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return params, nil
}

// UpdateSelectDBClusterConfig updates cluster configuration parameters
func (s *SelectDBService) UpdateSelectDBClusterConfig(clusterId, instanceId string, params []selectdb.ClusterConfigParam) error {
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if len(params) == 0 {
		return nil // Nothing to update
	}

	// Convert parameters to JSON string format expected by the API
	var paramUpdates []string
	for _, param := range params {
		if param.Name != "" && param.Value != "" {
			paramUpdates = append(paramUpdates, fmt.Sprintf(`"%s":"%s"`, param.Name, param.Value))
		}
	}

	if len(paramUpdates) == 0 {
		return nil // No valid parameters to update
	}

	parametersJSON := "{" + fmt.Sprintf("%s", paramUpdates[0])
	for i := 1; i < len(paramUpdates); i++ {
		parametersJSON += "," + paramUpdates[i]
	}
	parametersJSON += "}"

	err := s.GetAPI().ModifyClusterConfig(clusterId, instanceId, parametersJSON)
	if err != nil {
		return WrapError(err)
	}

	return nil
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
			selectdb.ClusterStatusResourcePreparing,
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
	if err != nil {
		return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
	}
	return nil
}

// WaitForSelectDBClusterUpdated waits for SelectDB cluster update operations to complete
func (s *SelectDBService) WaitForSelectDBClusterUpdated(instanceId, clusterId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			selectdb.ClusterStatusResourceChanging,
			selectdb.ClusterStatusReadonlyResourceChanging,
			selectdb.ClusterStatusClassChanging,
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
	if err != nil {
		return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
	}
	return nil
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
	if err != nil {
		return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", instanceId, clusterId))
	}
	return nil
}
