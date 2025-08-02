package alicloud

import (
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Connector methods
func (s *FlinkService) ListCustomConnectors(workspaceId string, namespaceName string) ([]*flinkAPI.Connector, error) {
	return s.GetAPI().ListConnectors(workspaceId, namespaceName)
}

func (s *FlinkService) RegisterCustomConnector(workspaceId string, namespaceName string, connector *flinkAPI.Connector) (*flinkAPI.Connector, error) {
	return s.GetAPI().RegisterConnector(workspaceId, namespaceName, connector)
}

func (s *FlinkService) GetConnector(workspaceId string, namespaceName string, connectorName string) (*flinkAPI.Connector, error) {
	return s.GetAPI().GetConnector(workspaceId, namespaceName, connectorName)
}

func (s *FlinkService) DeleteCustomConnector(workspaceId string, namespaceName string, connectorName string) error {
	return s.GetAPI().DeleteConnector(workspaceId, namespaceName, connectorName)
}

func (s *FlinkService) FlinkConnectorStateRefreshFunc(workspaceId string, namespaceName string, connectorName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		connector, err := s.GetConnector(workspaceId, namespaceName, connectorName)
		if err != nil {
			if IsNotFoundError(err) {
				// Connector not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For connectors, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if connector is in a failed state (if any fail states are defined)
			if connector.Type == failState {
				return connector, connector.Type, WrapError(Error(FailedToReachTargetStatus, connector.Type))
			}
		}

		return connector, "Available", nil
	}
}
