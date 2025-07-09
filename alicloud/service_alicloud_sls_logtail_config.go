package alicloud

import (
	"fmt"
	"strings"

	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *SlsService) DescribeSlsLogtailConfig(id string) (*slsAPI.LogtailConfig, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err := WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return nil, err
	}

	projectName := parts[0]
	configName := parts[2]

	config, err := s.aliyunSlsAPI.GetLogtailConfig(projectName, configName)
	if err != nil {
		if strings.Contains(err.Error(), "ConfigNotExist") || strings.Contains(err.Error(), "config not found") {
			return nil, WrapErrorf(NotFoundErr("LogtailConfig", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	return config, nil
}

func (s *SlsService) CreateSlsLogtailConfig(projectName string, config *slsAPI.LogtailConfig) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.CreateLogtailConfig(projectName, config)
	if err == nil {
		addDebugJson("CreateSlsLogtailConfig", config)
	}
	return err
}

func (s *SlsService) UpdateSlsLogtailConfig(projectName string, configName string, config *slsAPI.LogtailConfig) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateLogtailConfig(projectName, configName, config)
	if err == nil {
		addDebugJson("UpdateSlsLogtailConfig", config)
	}
	return err
}

func (s *SlsService) DeleteSlsLogtailConfig(projectName string, configName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteLogtailConfig(projectName, configName)
	if err == nil {
		addDebugJson("DeleteSlsLogtailConfig", fmt.Sprintf("Config %s deleted successfully", configName))
	}
	return err
}

func (s *SlsService) ListSlsLogtailConfigs(projectName string, configNameFilter string) ([]*slsAPI.LogtailConfig, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	return s.aliyunSlsAPI.ListLogtailConfigs(projectName, configNameFilter)
}

func (s *SlsService) LogtailConfigStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogtailConfig(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string
		switch field {
		case "configName":
			currentStatus = object.ConfigName
		case "inputType":
			currentStatus = object.InputType
		case "outputType":
			currentStatus = object.OutputType
		case "createTime":
			currentStatus = fmt.Sprint(object.CreateTime)
		case "lastModifyTime":
			currentStatus = fmt.Sprint(object.LastModifyTime)
		default:
			currentStatus = "Available"
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

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
