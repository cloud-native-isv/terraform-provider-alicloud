package alicloud

import (
	"fmt"
	"strings"

	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *SlsService) DescribeSlsLogtailPipelineConfig(id string) (*slsAPI.LogtailPipelineConfig, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err := WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return nil, err
	}

	projectName := parts[0]
	// parts[1] is generally "config"
	configName := parts[2]

	config, err := s.GetAPI().GetLogtailPipelineConfig(projectName, configName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetLogtailPipelineConfig", AlibabaCloudSdkGoERROR)
	}

	return config, nil
}

func (s *SlsService) CreateSlsLogtailPipelineConfig(projectName string, config *slsAPI.LogtailPipelineConfig) error {
	err := s.GetAPI().CreateLogtailPipelineConfig(projectName, config)
	if err == nil {
		addDebugJson("CreateSlsLogtailPipelineConfig", config)
	}
	return err
}

func (s *SlsService) UpdateSlsLogtailPipelineConfig(projectName string, config *slsAPI.LogtailPipelineConfig) error {
	err := s.GetAPI().UpdateLogtailPipelineConfig(projectName, config)
	if err == nil {
		addDebugJson("UpdateSlsLogtailPipelineConfig", config)
	}
	return err
}

func (s *SlsService) DeleteSlsLogtailPipelineConfig(projectName string, configName string) error {
	err := s.GetAPI().DeleteLogtailPipelineConfig(projectName, configName)
	if err == nil {
		addDebugJson("DeleteSlsLogtailPipelineConfig", fmt.Sprintf("PipelineConfig %s deleted successfully", configName))
	}
	return err
}

func (s *SlsService) LogtailPipelineConfigStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeSlsLogtailPipelineConfig(id)

		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", err
		}

		return obj, "active", nil
	}
}
