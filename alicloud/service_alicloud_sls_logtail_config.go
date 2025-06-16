package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsLogtailConfig retrieves logtail configuration details
func (s *SlsService) DescribeSlsLogtailConfig(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	configName := parts[2]

	// Use cws-lib-go API method
	config, err := s.aliyunSlsAPI.GetLogtailConfig(projectName, configName)
	if err != nil {
		if strings.Contains(err.Error(), "ConfigNotExist") || strings.Contains(err.Error(), "config not found") {
			return object, WrapErrorf(NotFoundErr("LogtailConfig", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	// Convert LogtailConfig to map for compatibility with existing Terraform schema
	result := make(map[string]interface{})
	result["configName"] = config.ConfigName
	result["inputType"] = config.InputType
	result["inputDetail"] = config.InputDetail
	result["outputType"] = config.OutputType
	result["outputDetail"] = config.OutputDetail
	result["logSample"] = config.LogSample
	result["createTime"] = config.CreateTime
	result["lastModifyTime"] = config.LastModifyTime

	// Extract logstore name from output detail for backward compatibility
	if config.OutputDetail != nil {
		result["logstoreName"] = config.OutputDetail.LogstoreName
	}

	return result, nil
}

// CreateSlsLogtailConfig creates a new logtail configuration
func (s *SlsService) CreateSlsLogtailConfig(projectName string, config *slsAPI.LogtailConfig) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.CreateLogtailConfig(projectName, config)
}

// UpdateSlsLogtailConfig updates an existing logtail configuration
func (s *SlsService) UpdateSlsLogtailConfig(projectName string, configName string, config *slsAPI.LogtailConfig) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.UpdateLogtailConfig(projectName, configName, config)
}

// DeleteSlsLogtailConfig deletes a logtail configuration
func (s *SlsService) DeleteSlsLogtailConfig(projectName string, configName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method
	return s.aliyunSlsAPI.DeleteLogtailConfig(projectName, configName)
}

// ListSlsLogtailConfigs lists logtail configurations in a project
func (s *SlsService) ListSlsLogtailConfigs(projectName string, configNameFilter string) ([]*slsAPI.LogtailConfig, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Use cws-lib-go API method with pagination handled internally
	return s.aliyunSlsAPI.ListLogtailConfigs(projectName, configNameFilter)
}

// SlsLogtailConfigStateRefreshFunc returns a StateRefreshFunc for logtail config status monitoring
func (s *SlsService) SlsLogtailConfigStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogtailConfig(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		if err != nil {
			return nil, "", WrapError(err)
		}
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// ValidateSlsLogtailConfig validates logtail configuration parameters
func (s *SlsService) ValidateSlsLogtailConfig(config *slsAPI.LogtailConfig) error {
	if config == nil {
		return fmt.Errorf("logtail config cannot be nil")
	}

	if config.ConfigName == "" {
		return fmt.Errorf("config name is required")
	}

	if config.InputType == "" {
		return fmt.Errorf("input type is required")
	}

	if config.OutputType == "" {
		return fmt.Errorf("output type is required")
	}

	if config.InputDetail == nil {
		return fmt.Errorf("input detail is required")
	}

	if config.OutputDetail == nil {
		return fmt.Errorf("output detail is required")
	}

	if config.OutputDetail.Endpoint == "" {
		return fmt.Errorf("output endpoint is required")
	}

	if config.OutputDetail.LogstoreName == "" {
		return fmt.Errorf("output logstore name is required")
	}

	return nil
}

// BuildSlsLogtailConfigFromMap creates a LogtailConfig from Terraform resource data map
func (s *SlsService) BuildSlsLogtailConfigFromMap(d map[string]interface{}) (*slsAPI.LogtailConfig, error) {
	config := &slsAPI.LogtailConfig{}

	if v, ok := d["config_name"]; ok {
		config.ConfigName = v.(string)
	}

	if v, ok := d["input_type"]; ok {
		config.InputType = v.(string)
	}

	if v, ok := d["input_detail"]; ok {
		if inputDetail, ok := v.(map[string]interface{}); ok {
			config.InputDetail = inputDetail
		}
	}

	if v, ok := d["output_type"]; ok {
		config.OutputType = v.(string)
	}

	if v, ok := d["output_detail"]; ok {
		if outputDetailMap, ok := v.(map[string]interface{}); ok {
			outputDetail := &slsAPI.LogtailConfigOutputDetail{}

			if endpoint, exists := outputDetailMap["endpoint"]; exists {
				outputDetail.Endpoint = endpoint.(string)
			}

			if logstore, exists := outputDetailMap["logstore_name"]; exists {
				outputDetail.LogstoreName = logstore.(string)
			}

			if region, exists := outputDetailMap["region"]; exists {
				outputDetail.Region = region.(string)
			}

			if telemetryType, exists := outputDetailMap["telemetry_type"]; exists {
				outputDetail.TelemetryType = telemetryType.(string)
			}

			config.OutputDetail = outputDetail
		}
	}

	if v, ok := d["log_sample"]; ok {
		config.LogSample = v.(string)
	}

	return config, nil
}

// ConvertSlsLogtailConfigToMap converts LogtailConfig to map for Terraform schema compatibility
func (s *SlsService) ConvertSlsLogtailConfigToMap(config *slsAPI.LogtailConfig) map[string]interface{} {
	result := make(map[string]interface{})

	result["config_name"] = config.ConfigName
	result["input_type"] = config.InputType
	result["input_detail"] = config.InputDetail
	result["output_type"] = config.OutputType
	result["log_sample"] = config.LogSample
	result["create_time"] = config.CreateTime
	result["last_modify_time"] = config.LastModifyTime

	if config.OutputDetail != nil {
		outputDetail := map[string]interface{}{
			"endpoint":       config.OutputDetail.Endpoint,
			"logstore_name":  config.OutputDetail.LogstoreName,
			"region":         config.OutputDetail.Region,
			"telemetry_type": config.OutputDetail.TelemetryType,
		}
		result["output_detail"] = outputDetail

		// For backward compatibility
		result["logstore_name"] = config.OutputDetail.LogstoreName
	}

	return result
}

// GetSlsLogtailConfigMachineGroups gets machine groups attached to a logtail config
func (s *SlsService) GetSlsLogtailConfigMachineGroups(projectName string, configName string) ([]string, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Get applied machine groups for the config
	attachments, err := s.aliyunSlsAPI.GetAppliedMachineGroups(projectName, configName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, configName, "GetAppliedMachineGroups", AlibabaCloudSdkGoERROR)
	}

	var machineGroups []string
	for _, attachment := range attachments {
		machineGroups = append(machineGroups, attachment.MachineGroup)
	}

	return machineGroups, nil
}

// SlsLogtailConfigExists checks if a logtail config exists
func (s *SlsService) SlsLogtailConfigExists(projectName string, configName string) (bool, error) {
	_, err := s.DescribeSlsLogtailConfig(fmt.Sprintf("%s:config:%s", projectName, configName))
	if err != nil {
		if NotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
