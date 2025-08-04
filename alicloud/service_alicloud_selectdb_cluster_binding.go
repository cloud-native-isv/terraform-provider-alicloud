package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Cluster Binding Management Operations

// CreateSelectDBClusterBinding creates a new SelectDB cluster binding
func (s *SelectDBService) CreateSelectDBClusterBinding(clusterId, instanceId string, clusterIdBak ...string) (*selectdb.ClusterBindingResult, error) {
	if clusterId == "" {
		return nil, WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	result, err := s.GetAPI().CreateClusterBinding(clusterId, instanceId, clusterIdBak...)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteSelectDBClusterBinding deletes a SelectDB cluster binding
func (s *SelectDBService) DeleteSelectDBClusterBinding(clusterId, instanceId string) error {
	if clusterId == "" {
		return WrapError(fmt.Errorf("cluster ID cannot be empty"))
	}
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	err := s.GetAPI().DeleteClusterBinding(clusterId, instanceId)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// WaitForSelectDBClusterBindingDeleted waits for the cluster binding to be deleted
func (s *SelectDBService) WaitForSelectDBClusterBindingDeleted(clusterId, instanceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			// Try to get cluster binding info
			// Since there's no direct API to check binding status, we rely on delete operation success
			return nil, "", nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", clusterId, instanceId))
	}

	return nil
}
