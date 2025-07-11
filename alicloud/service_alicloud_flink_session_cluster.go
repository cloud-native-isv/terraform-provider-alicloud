package alicloud

import (
	"fmt"
	"strings"
	"time"

	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateSessionCluster creates a new session cluster
func (s *FlinkService) CreateSessionCluster(workspaceId string, namespaceName string, sessionCluster *flinkAPI.SessionCluster) (*flinkAPI.SessionCluster, error) {
	return s.flinkAPI.CreateSessionCluster(workspaceId, namespaceName, sessionCluster)
}

// DescribeSessionCluster retrieves a session cluster by name
func (s *FlinkService) DescribeSessionCluster(id string) (*flinkAPI.SessionCluster, error) {
	workspaceId, namespaceName, sessionClusterName, err := parseSessionClusterId(id)
	if err != nil {
		return nil, err
	}

	return s.flinkAPI.GetSessionCluster(workspaceId, namespaceName, sessionClusterName)
}

// UpdateSessionCluster updates an existing session cluster
func (s *FlinkService) UpdateSessionCluster(workspaceId string, namespaceName string, sessionClusterName string, sessionCluster *flinkAPI.SessionCluster) (*flinkAPI.SessionCluster, error) {
	return s.flinkAPI.UpdateSessionCluster(workspaceId, namespaceName, sessionClusterName, sessionCluster)
}

// DeleteSessionCluster deletes a session cluster
func (s *FlinkService) DeleteSessionCluster(workspaceId string, namespaceName string, sessionClusterName string) (*flinkAPI.SessionCluster, error) {
	return s.flinkAPI.DeleteSessionCluster(workspaceId, namespaceName, sessionClusterName)
}

// ListSessionClusters lists all session clusters in a namespace
func (s *FlinkService) ListSessionClusters(workspaceId string, namespaceName string) ([]*flinkAPI.SessionCluster, error) {
	return s.flinkAPI.ListSessionClusters(workspaceId, namespaceName)
}

// StartSessionCluster starts a session cluster
func (s *FlinkService) StartSessionCluster(workspaceId string, namespaceName string, sessionClusterName string) error {
	return s.flinkAPI.StartSessionCluster(workspaceId, namespaceName, sessionClusterName)
}

// StopSessionCluster stops a session cluster
func (s *FlinkService) StopSessionCluster(workspaceId string, namespaceName string, sessionClusterName string) error {
	return s.flinkAPI.StopSessionCluster(workspaceId, namespaceName, sessionClusterName)
}

// SessionClusterStateRefreshFunc returns a StateRefreshFunc that tracks session cluster status
func (s *FlinkService) SessionClusterStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSessionCluster(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Status != nil && object.Status.CurrentSessionClusterStatus == failState {
				return object, object.Status.CurrentSessionClusterStatus, WrapError(Error(FailedToReachTargetStatus, object.Status.CurrentSessionClusterStatus))
			}
		}

		status := "Unknown"
		if object.Status != nil {
			status = object.Status.CurrentSessionClusterStatus
		}
		return object, status, nil
	}
}

// WaitForSessionClusterStopping waits for a session cluster to stop
func (s *FlinkService) WaitForSessionClusterStopping(id string, timeout time.Duration) error {
	stopPendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
		flinkAPI.FlinkSessionClusterStatusStopping,
	})
	stopExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStopped,
		flinkAPI.FlinkSessionClusterStatusFailed,
	})
	stopFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		stopPendingStatus,
		stopExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, stopFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterStarting waits for a session cluster to start
func (s *FlinkService) WaitForSessionClusterStarting(id string, timeout time.Duration) error {
	startPendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStopped,
		flinkAPI.FlinkSessionClusterStatusStarting,
	})
	startExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
	})
	startFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		startPendingStatus,
		startExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, startFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterDeleting waits for a session cluster to be deleted
func (s *FlinkService) WaitForSessionClusterDeleting(id string, timeout time.Duration) error {
	deletePendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStopped,
		flinkAPI.FlinkSessionClusterStatusFailed,
	})
	deleteExpectStatus := []string{} // Empty means waiting for resource to be gone
	deleteFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		deletePendingStatus,
		deleteExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, deleteFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterCreating waits for a session cluster to be created and ready
func (s *FlinkService) WaitForSessionClusterCreating(id string, timeout time.Duration) error {
	createPendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStarting,
		flinkAPI.FlinkSessionClusterStatusUpdating,
	})
	createExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
		flinkAPI.FlinkSessionClusterStatusStopped,
	})
	createFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		createPendingStatus,
		createExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, createFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterUpdating waits for a session cluster to finish updating
func (s *FlinkService) WaitForSessionClusterUpdating(id string, timeout time.Duration) error {
	updatePendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusUpdating,
	})
	updateExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
		flinkAPI.FlinkSessionClusterStatusStopped,
	})
	updateFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		updatePendingStatus,
		updateExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, updateFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterStopped waits for a session cluster to be stopped
func (s *FlinkService) WaitForSessionClusterStopped(id string, timeout time.Duration) error {
	stopPendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
		flinkAPI.FlinkSessionClusterStatusStopping,
	})
	stopExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStopped,
	})
	stopFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		stopPendingStatus,
		stopExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, stopFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// WaitForSessionClusterRunning waits for a session cluster to be running
func (s *FlinkService) WaitForSessionClusterRunning(id string, timeout time.Duration) error {
	startPendingStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusStopped,
		flinkAPI.FlinkSessionClusterStatusStarting,
	})
	startExpectStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusRunning,
	})
	startFailedStatus := flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
		flinkAPI.FlinkSessionClusterStatusFailed,
	})

	stateConf := BuildStateConf(
		startPendingStatus,
		startExpectStatus,
		timeout,
		30*time.Second,
		s.SessionClusterStateRefreshFunc(id, startFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// Helper function to parse session cluster ID
func parseSessionClusterId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid session cluster ID format, expected workspaceId:namespaceName:sessionClusterName, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// Helper function to format session cluster ID
func formatSessionClusterId(workspaceId string, namespaceName string, sessionClusterName string) string {
	return fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, sessionClusterName)
}
