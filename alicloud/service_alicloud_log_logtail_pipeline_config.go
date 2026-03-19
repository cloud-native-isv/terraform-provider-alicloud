package alicloud

import (
	"encoding/json"
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

func NormalizeLogtailConfigJson(jsonStr string) (string, error) {
	return normalizeJsonString(jsonStr)
}

func ValidateLogtailConfigJsonObject(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}
	var i interface{}
	if err := json.Unmarshal([]byte(jsonStr), &i); err != nil {
		return err
	}
	if _, ok := i.(map[string]interface{}); !ok {
		return fmt.Errorf("expected JSON object")
	}
	return nil
}

func validateLogtailConfigJsonObjectString(v interface{}, k string) (ws []string, errors []error) {
	jsonStr, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("%s must be a JSON string", k))
		return
	}

	if err := ValidateLogtailConfigJsonObject(jsonStr); err != nil {
		errors = append(errors, fmt.Errorf("%s: %s", k, err.Error()))
	}
	return
}

func (c *SlsLogtailPipelineConfigPlugin) ToMap() (map[string]interface{}, error) {
	m := make(map[string]interface{})

	// Initialize with empty map if config json is empty, otherwise parse it
	if c.ConfigJson != "" {
		if err := json.Unmarshal([]byte(c.ConfigJson), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config_json for plugin type %s: %v", c.Type, err)
		}
	}

	// Ensure type is set and overrides any existing type in json
	delete(m, "type")
	delete(m, "Type")
	m["Type"] = c.Type
	return m, nil
}

func (c *SlsLogtailPipelineConfig) ToLibConfig() (*slsAPI.LogtailPipelineConfig, error) {
	if c == nil {
		return nil, nil
	}
	libConfig := &slsAPI.LogtailPipelineConfig{
		ConfigName: c.Name, // Project is passed separately in API
		LogSample:  c.LogSample,
	}

	// inputs
	if len(c.Inputs) > 0 {
		inputs := make([]map[string]interface{}, len(c.Inputs))
		for i, v := range c.Inputs {
			m, err := v.ToMap()
			if err != nil {
				return nil, err
			}
			inputs[i] = m
		}
		libConfig.Inputs = inputs
	}

	// processors
	if len(c.Processors) > 0 {
		processors := make([]map[string]interface{}, len(c.Processors))
		for i, v := range c.Processors {
			m, err := v.ToMap()
			if err != nil {
				return nil, err
			}
			processors[i] = m
		}
		libConfig.Processors = processors
	}

	// flushers
	if len(c.Flushers) > 0 {
		flushers := make([]map[string]interface{}, len(c.Flushers))
		for i, v := range c.Flushers {
			m, err := v.ToMap()
			if err != nil {
				return nil, err
			}
			flushers[i] = m
		}
		libConfig.Flushers = flushers
	}

	// aggregators
	if len(c.Aggregators) > 0 {
		aggregators := make([]map[string]interface{}, len(c.Aggregators))
		for i, v := range c.Aggregators {
			m, err := v.ToMap()
			if err != nil {
				return nil, err
			}
			aggregators[i] = m
		}
		libConfig.Aggregators = aggregators
	}

	// global
	if c.GlobalJson != "" {
		var global map[string]interface{}
		if err := json.Unmarshal([]byte(c.GlobalJson), &global); err != nil {
			return nil, fmt.Errorf("failed to unmarshal global_json: %v", err)
		}
		libConfig.Global = global
	}

	// task
	if c.TaskJson != "" {
		var task map[string]interface{}
		if err := json.Unmarshal([]byte(c.TaskJson), &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task_json: %v", err)
		}
		libConfig.Task = task
	}

	return libConfig, nil
}

func FromLibConfigPlugin(m map[string]interface{}) SlsLogtailPipelineConfigPlugin {
	p := SlsLogtailPipelineConfigPlugin{}
	if t, ok := m["type"].(string); ok {
		p.Type = t
	} else if t, ok := m["Type"].(string); ok {
		p.Type = t
	}

	// Clone map to avoid modifying original, and remove type from JSON to be cleaner?
	// The schema says: type is one field, config_json is another.
	// We can leave type in config_json or remove it.
	// If we leave it, the user sees type in both places.
	// Usually for clarity, if 'type' is a separate attribute, we might want to strip it from config_json if it exists,
	// BUT, if the user provided it in JSON, we should probably keep it consistent.
	// However, if we write back to state, we need to be stable.
	// Let's decide to KEEP it or REMOVE it based on ensuring 'normalizeJsonString' works well.
	// If we keep it, it's just redundant data.
	// Let's REMOVE 'type' from the JSON map before marshaling to config_json,
	// because 'type' is already exposed as a first-class attribute.

	// Copy map
	cMap := make(map[string]interface{})
	for k, v := range m {
		if k != "type" && k != "Type" {
			cMap[k] = v
		}
	}

	if len(cMap) > 0 {
		bytes, _ := json.Marshal(cMap)
		p.ConfigJson, _ = normalizeJsonString(string(bytes))
	}

	return p
}

func FromLibConfig(libConfig *slsAPI.LogtailPipelineConfig, project string) *SlsLogtailPipelineConfig {
	if libConfig == nil {
		return nil
	}

	c := &SlsLogtailPipelineConfig{
		Project:        project,
		Name:           libConfig.ConfigName,
		LogSample:      libConfig.LogSample,
		CreateTime:     libConfig.CreateTime,
		LastModifyTime: libConfig.LastModifyTime,
	}

	if len(libConfig.Inputs) > 0 {
		c.Inputs = make([]SlsLogtailPipelineConfigPlugin, len(libConfig.Inputs))
		for i, v := range libConfig.Inputs {
			c.Inputs[i] = FromLibConfigPlugin(v)
		}
	}

	if len(libConfig.Processors) > 0 {
		c.Processors = make([]SlsLogtailPipelineConfigPlugin, len(libConfig.Processors))
		for i, v := range libConfig.Processors {
			c.Processors[i] = FromLibConfigPlugin(v)
		}
	}

	if len(libConfig.Flushers) > 0 {
		c.Flushers = make([]SlsLogtailPipelineConfigPlugin, len(libConfig.Flushers))
		for i, v := range libConfig.Flushers {
			c.Flushers[i] = FromLibConfigPlugin(v)
		}
	}

	if len(libConfig.Aggregators) > 0 {
		c.Aggregators = make([]SlsLogtailPipelineConfigPlugin, len(libConfig.Aggregators))
		for i, v := range libConfig.Aggregators {
			c.Aggregators[i] = FromLibConfigPlugin(v)
		}
	}

	if libConfig.Global != nil {
		bytes, _ := json.Marshal(libConfig.Global)
		c.GlobalJson, _ = normalizeJsonString(string(bytes))
	}

	if libConfig.Task != nil {
		bytes, _ := json.Marshal(libConfig.Task)
		c.TaskJson, _ = normalizeJsonString(string(bytes))
	}

	return c
}

// SlsLogtailPipelineConfig represents the domain model for Logtail Pipeline Config
type SlsLogtailPipelineConfig struct {
	Project        string
	Name           string
	Inputs         []SlsLogtailPipelineConfigPlugin
	Processors     []SlsLogtailPipelineConfigPlugin
	Flushers       []SlsLogtailPipelineConfigPlugin
	Aggregators    []SlsLogtailPipelineConfigPlugin
	GlobalJson     string
	TaskJson       string
	LogSample      string
	CreateTime     int64
	LastModifyTime int64
}

type SlsLogtailPipelineConfigPlugin struct {
	Type       string
	ConfigJson string
}
