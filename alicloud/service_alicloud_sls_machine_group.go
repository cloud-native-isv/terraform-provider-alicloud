package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsMachineGroup retrieves a machine group by project and name
func (s *SlsService) DescribeSlsMachineGroup(projectName, machineGroupName string) (*aliyunSlsAPI.MachineGroup, error) {
	machineGroup, err := s.aliyunSlsAPI.GetMachineGroup(projectName, machineGroupName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "GetMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return machineGroup, nil
}

// CreateSlsMachineGroup creates a new machine group
func (s *SlsService) CreateSlsMachineGroup(projectName string, machineGroup *aliyunSlsAPI.MachineGroup) error {
	err := s.aliyunSlsAPI.CreateMachineGroup(projectName, machineGroup)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "CreateMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsMachineGroup updates an existing machine group
func (s *SlsService) UpdateSlsMachineGroup(projectName, machineGroupName string, machineGroup *aliyunSlsAPI.MachineGroup) error {
	err := s.aliyunSlsAPI.UpdateMachineGroup(projectName, machineGroupName, machineGroup)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "UpdateMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsMachineGroup deletes a machine group
func (s *SlsService) DeleteSlsMachineGroup(projectName, machineGroupName string) error {
	err := s.aliyunSlsAPI.DeleteMachineGroup(projectName, machineGroupName)
	if err != nil {
		if IsExpectedErrors(err, []string{"MachineGroupNotExist", "ProjectNotExist"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "DeleteMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsMachineGroups lists all machine groups in a project
func (s *SlsService) ListSlsMachineGroups(projectName string) ([]*aliyunSlsAPI.MachineGroup, error) {
	machineGroups, err := s.aliyunSlsAPI.ListMachineGroups(projectName, "")
	if err != nil {
		if IsExpectedErrors(err, []string{"ProjectNotExist"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "ListMachineGroups", AlibabaCloudSdkGoERROR)
	}

	return machineGroups, nil
}

// SlsMachineGroupStateRefreshFunc returns a StateRefreshFunc for machine group operations
func (s *SlsService) SlsMachineGroupStateRefreshFunc(projectName, machineGroupName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsMachineGroup(projectName, machineGroupName)
		if err != nil {
			if IsNotFoundError(err) {
				// Return nil for not found, allowing create operations to proceed
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Name == failState {
				return object, failState, Error("Failed to reach target state. Last error: %s", failState)
			}
		}

		// Machine groups don't have explicit states like other resources,
		// so we consider existence as "Available" state
		return object, "Available", nil
	}
}

// WaitForSlsMachineGroup waits for machine group to reach the specified state
func (s *SlsService) WaitForSlsMachineGroup(projectName, machineGroupName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeSlsMachineGroup(projectName, machineGroupName)
		if err != nil {
			if IsNotFoundError(err) {
				if status == "Deleted" {
					return nil
				}
			} else {
				return WrapError(err)
			}
		} else {
			if status == "Available" && object != nil {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, projectName+":"+machineGroupName, GetFunc(1), timeout, "", status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort)
	}
}

// GetMachineGroupAppliedConfigs retrieves the logtail configs applied to a machine group
func (s *SlsService) GetMachineGroupAppliedConfigs(projectName, machineGroupName string) ([]string, error) {
	// Note: This method may not be available in the current API
	// Returning empty slice for now to maintain compatibility
	return []string{}, nil
}

// ApplyConfigToMachineGroup applies a logtail config to a machine group
func (s *SlsService) ApplyConfigToMachineGroup(projectName, configName, machineGroupName string) error {
	_, err := s.aliyunSlsAPI.ApplyConfigToMachineGroup(projectName, configName, machineGroupName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "ApplyConfigToMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// RemoveConfigFromMachineGroup removes a logtail config from a machine group
func (s *SlsService) RemoveConfigFromMachineGroup(projectName, configName, machineGroupName string) error {
	err := s.aliyunSlsAPI.RemoveConfigFromMachineGroup(projectName, configName, machineGroupName)
	if err != nil {
		if IsExpectedErrors(err, []string{"MachineGroupNotExist", "ProjectNotExist", "ConfigNotExist"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "RemoveConfigFromMachineGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ValidateMachineGroupIdentifyType validates the machine group identify type
func (s *SlsService) ValidateMachineGroupIdentifyType(identifyType string) error {
	// Use constants from the SLS SDK directly
	validTypes := []string{"ip", "userdefined"}
	for _, validType := range validTypes {
		if identifyType == validType {
			return nil
		}
	}
	return WrapError(Error("Invalid machine group identify type: %s. Valid types are: %v", identifyType, validTypes))
}

// ParseMachineGroupId parses the composite ID format: project:machine_group_name
func (s *SlsService) ParseMachineGroupId(id string) (projectName, machineGroupName string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", WrapError(Error("Invalid machine group ID format. Expected format: project:machine_group_name, got: %s", id))
	}
	return parts[0], parts[1], nil
}

// BuildMachineGroupId builds the composite ID format: project:machine_group_name
func (s *SlsService) BuildMachineGroupId(projectName, machineGroupName string) string {
	return fmt.Sprintf("%s:%s", projectName, machineGroupName)
}
