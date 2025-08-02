package alicloud

import (
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Variable methods
func (s *FlinkService) GetVariable(workspaceId string, namespaceName string, variableName string) (*flinkAPI.Variable, error) {
	// Call the underlying API to get variable
	return s.GetAPI().GetVariable(workspaceId, namespaceName, variableName)
}

func (s *FlinkService) CreateVariable(workspaceId string, namespaceName string, variable *flinkAPI.Variable) (*flinkAPI.Variable, error) {
	// Call underlying API
	return s.GetAPI().CreateVariable(workspaceId, namespaceName, variable)
}

func (s *FlinkService) UpdateVariable(workspaceId string, namespaceName string, variable *flinkAPI.Variable) (*flinkAPI.Variable, error) {
	// Call underlying API
	return s.GetAPI().UpdateVariable(workspaceId, namespaceName, variable)
}

func (s *FlinkService) DeleteVariable(workspaceId string, namespaceName string, variableName string) error {
	// Call the underlying API
	return s.GetAPI().DeleteVariable(workspaceId, namespaceName, variableName)
}

func (s *FlinkService) ListVariables(workspaceId string, namespaceName string) ([]flinkAPI.Variable, error) {
	return s.GetAPI().ListVariables(workspaceId, namespaceName)
}

func (s *FlinkService) FlinkVariableStateRefreshFunc(workspaceId string, namespaceName string, variableName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		variable, err := s.GetVariable(workspaceId, namespaceName, variableName)
		if err != nil {
			if IsNotFoundError(err) {
				// Variable not found, still being created
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For variables, if we can get it successfully, it means it's ready
		for _, failState := range failStates {
			// Check if variable is in a failed state (if any fail states are defined)
			if variable.Kind == failState {
				return variable, variable.Kind, WrapError(Error(FailedToReachTargetStatus, variable.Kind))
			}
		}

		return variable, "Available", nil
	}
}
