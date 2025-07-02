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

// WaitForSessionCluster waits for a session cluster to reach target status
func (s *FlinkService) WaitForSessionCluster(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeSessionCluster(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		currentStatus := "Unknown"
		if object != nil && object.Status != nil {
			currentStatus = object.Status.CurrentSessionClusterStatus
		}

		if currentStatus == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, currentStatus, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort)
	}
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
