package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsLogtailAttachment retrieves logtail attachment details
func (s *SlsService) DescribeSlsLogtailAttachment(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	configName := parts[1]
	machineGroupName := parts[2]

	// Use cws-lib-go API method to get attachment
	attachment, err := s.aliyunSlsAPI.GetAttachment(projectName, configName, machineGroupName)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetAttachment", AlibabaCloudSdkGoERROR)
	}

	// Convert LogtailAttachment to map for compatibility with existing Terraform schema
	result := make(map[string]interface{})
	result["projectName"] = attachment.Project
	result["configName"] = attachment.ConfigName
	result["machineGroupName"] = attachment.MachineGroup
	result["attached"] = true
	result["status"] = attachment.Status
	result["createTime"] = attachment.CreateTime
	result["lastModifyTime"] = attachment.LastModifyTime

	return result, nil
}

// CreateSlsLogtailAttachment creates a new logtail attachment
func (s *SlsService) CreateSlsLogtailAttachment(projectName string, configName string, machineGroup string) (*slsAPI.LogtailAttachment, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.ApplyConfigToMachineGroup(projectName, configName, machineGroup)
}

// UpdateSlsLogtailAttachment updates an existing logtail attachment
func (s *SlsService) UpdateSlsLogtailAttachment(projectName string, configName string, machineGroup string) (*slsAPI.LogtailAttachment, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method (internally handles remove and re-apply)
	return s.aliyunSlsAPI.UpdateAttachment(projectName, configName, machineGroup)
}

// DeleteSlsLogtailAttachment deletes a logtail attachment
func (s *SlsService) DeleteSlsLogtailAttachment(projectName string, configName string, machineGroup string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.RemoveConfigFromMachineGroup(projectName, configName, machineGroup)
}

// SlsLogtailAttachmentStateRefreshFunc returns a StateRefreshFunc for logtail attachment status monitoring
func (s *SlsService) SlsLogtailAttachmentStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogtailAttachment(id)
		if err != nil {
			if IsNotFoundError(err) {
				// When resource is not found during deletion, this is the expected success state
				return nil, "deleted", nil
			}
			return nil, "", WrapError(err)
		}

		// Handle the case when object is found
		if object == nil {
			// Object is nil but no error - treat as deleted
			return nil, "deleted", nil
		}

		v, err := jsonpath.Get(field, object)
		if err != nil {
			// If we can't get the field, try to use a default status
			if status, ok := object["status"]; ok {
				currentStatus := fmt.Sprint(status)
				for _, failState := range failStates {
					if currentStatus == failState {
						return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
					}
				}
				return object, currentStatus, nil
			}
			// If no status field exists, assume it's active
			return object, "active", nil
		}

		currentStatus := fmt.Sprint(v)

		// Handle empty status - default to active if object exists
		if currentStatus == "" || currentStatus == "<nil>" {
			currentStatus = "active"
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}

		return object, currentStatus, nil
	}
}

// ValidateSlsLogtailAttachment validates logtail attachment parameters
func (s *SlsService) ValidateSlsLogtailAttachment(projectName string, configName string, machineGroup string) error {
	if projectName == "" {
		return fmt.Errorf("project name is required")
	}

	if configName == "" {
		return fmt.Errorf("config name is required")
	}

	if machineGroup == "" {
		return fmt.Errorf("machine group is required")
	}

	return nil
}

// BuildSlsLogtailAttachmentFromMap creates a LogtailAttachment from Terraform resource data map
func (s *SlsService) BuildSlsLogtailAttachmentFromMap(d map[string]interface{}) (*slsAPI.LogtailAttachment, error) {
	attachment := &slsAPI.LogtailAttachment{}

	if v, ok := d["project"]; ok {
		attachment.Project = v.(string)
	}

	if v, ok := d["config_name"]; ok {
		attachment.ConfigName = v.(string)
	}

	if v, ok := d["machine_group"]; ok {
		attachment.MachineGroup = v.(string)
	}

	if v, ok := d["status"]; ok {
		attachment.Status = v.(string)
	}

	return attachment, nil
}

// ConvertSlsLogtailAttachmentToMap converts LogtailAttachment to map for Terraform schema compatibility
func (s *SlsService) ConvertSlsLogtailAttachmentToMap(attachment *slsAPI.LogtailAttachment) map[string]interface{} {
	result := make(map[string]interface{})

	result["project"] = attachment.Project
	result["config_name"] = attachment.ConfigName
	result["machine_group"] = attachment.MachineGroup
	result["status"] = attachment.Status
	result["create_time"] = attachment.CreateTime
	result["last_modify_time"] = attachment.LastModifyTime
	result["attached"] = true

	// For backward compatibility
	result["projectName"] = attachment.Project
	result["configName"] = attachment.ConfigName
	result["machineGroupName"] = attachment.MachineGroup

	return result
}

// SlsLogtailAttachmentExists checks if a logtail attachment exists
func (s *SlsService) SlsLogtailAttachmentExists(projectName string, configName string, machineGroup string) (bool, error) {
	if s.aliyunSlsAPI == nil {
		return false, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.AttachmentExists(projectName, configName, machineGroup)
}

// GetSlsLogtailConfigAttachments gets all attachments for a specific logtail config
func (s *SlsService) GetSlsLogtailConfigAttachments(projectName string, configName string) ([]*slsAPI.LogtailAttachment, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.GetAppliedMachineGroups(projectName, configName)
}

// GetSlsMachineGroupAttachments gets all attachments for a specific machine group
func (s *SlsService) GetSlsMachineGroupAttachments(projectName string, machineGroup string) ([]*slsAPI.LogtailAttachment, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.GetAppliedConfigs(projectName, machineGroup)
}

// BatchCreateSlsLogtailAttachments creates multiple logtail attachments for a config
func (s *SlsService) BatchCreateSlsLogtailAttachments(projectName string, configName string, machineGroups []string) ([]*slsAPI.LogtailAttachment, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	var attachments []*slsAPI.LogtailAttachment
	var errors []error

	for _, machineGroup := range machineGroups {
		attachment, err := s.aliyunSlsAPI.ApplyConfigToMachineGroup(projectName, configName, machineGroup)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to attach config %s to machine group %s: %w", configName, machineGroup, err))
			continue
		}
		attachments = append(attachments, attachment)
	}

	if len(errors) > 0 {
		// Return partial success with error details
		var errorMsgs []string
		for _, err := range errors {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return attachments, fmt.Errorf("batch create partially failed: %s", strings.Join(errorMsgs, "; "))
	}

	return attachments, nil
}

// BatchDeleteSlsLogtailAttachments deletes multiple logtail attachments for a config
func (s *SlsService) BatchDeleteSlsLogtailAttachments(projectName string, configName string, machineGroups []string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	var errors []error

	for _, machineGroup := range machineGroups {
		err := s.aliyunSlsAPI.RemoveConfigFromMachineGroup(projectName, configName, machineGroup)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to remove config %s from machine group %s: %w", configName, machineGroup, err))
		}
	}

	if len(errors) > 0 {
		var errorMsgs []string
		for _, err := range errors {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return fmt.Errorf("batch delete failed: %s", strings.Join(errorMsgs, "; "))
	}

	return nil
}

// SyncSlsLogtailAttachments synchronizes logtail attachments to match desired state
func (s *SlsService) SyncSlsLogtailAttachments(projectName string, configName string, desiredMachineGroups []string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Get current attachments
	currentAttachments, err := s.aliyunSlsAPI.GetAppliedMachineGroups(projectName, configName)
	if err != nil {
		return fmt.Errorf("failed to get current attachments: %w", err)
	}

	// Build current machine groups set
	currentMachineGroups := make(map[string]bool)
	for _, attachment := range currentAttachments {
		currentMachineGroups[attachment.MachineGroup] = true
	}

	// Build desired machine groups set
	desiredMachineGroupsSet := make(map[string]bool)
	for _, mg := range desiredMachineGroups {
		desiredMachineGroupsSet[mg] = true
	}

	// Find machine groups to add
	var toAdd []string
	for mg := range desiredMachineGroupsSet {
		if !currentMachineGroups[mg] {
			toAdd = append(toAdd, mg)
		}
	}

	// Find machine groups to remove
	var toRemove []string
	for mg := range currentMachineGroups {
		if !desiredMachineGroupsSet[mg] {
			toRemove = append(toRemove, mg)
		}
	}

	// Execute changes
	if len(toAdd) > 0 {
		_, err := s.BatchCreateSlsLogtailAttachments(projectName, configName, toAdd)
		if err != nil {
			return fmt.Errorf("failed to add attachments: %w", err)
		}
	}

	if len(toRemove) > 0 {
		err := s.BatchDeleteSlsLogtailAttachments(projectName, configName, toRemove)
		if err != nil {
			return fmt.Errorf("failed to remove attachments: %w", err)
		}
	}

	return nil
}
