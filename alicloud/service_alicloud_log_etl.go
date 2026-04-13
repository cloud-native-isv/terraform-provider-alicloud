package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsETL retrieves an ETL task by project and ETL name
func (s *SlsService) DescribeSlsETL(projectName, etlName string) (*aliyunSlsAPI.ETL, error) {
	slsAPI := s.GetAPI()

	etl, err := slsAPI.GetETL(projectName, etlName)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(NotFoundErr("SlsETL", fmt.Sprintf("%s:%s", projectName, etlName)), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", projectName, etlName), "GetETL", AlibabaCloudSdkGoERROR)
	}

	return etl, nil
}

// CreateSlsETL creates a new ETL task in the specified project
func (s *SlsService) CreateSlsETL(projectName string, etl *aliyunSlsAPI.ETL) error {
	slsAPI := s.GetAPI()

	err := slsAPI.CreateETL(projectName, etl)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, etl.Name, "CreateETL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsETL updates an existing ETL task in the specified project
func (s *SlsService) UpdateSlsETL(projectName, etlName string, etl *aliyunSlsAPI.ETL) error {
	slsAPI := s.GetAPI()

	err := slsAPI.UpdateETL(projectName, etlName, etl)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, etlName, "UpdateETL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsETL deletes an ETL task from the specified project
func (s *SlsService) DeleteSlsETL(projectName, etlName string) error {
	slsAPI := s.GetAPI()

	err := slsAPI.DeleteETL(projectName, etlName)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, etlName, "DeleteETL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StartSlsETL starts an ETL task
func (s *SlsService) StartSlsETL(projectName, etlName string) error {
	slsAPI := s.GetAPI()

	err := slsAPI.StartETL(projectName, etlName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, etlName, "StartETL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StopSlsETL stops an ETL task
func (s *SlsService) StopSlsETL(projectName, etlName string) error {
	slsAPI := s.GetAPI()

	err := slsAPI.StopETL(projectName, etlName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, etlName, "StopETL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsETLs lists ETL tasks in the specified project with optional filtering
func (s *SlsService) ListSlsETLs(projectName, etlName, logstore string) ([]*aliyunSlsAPI.ETL, error) {
	slsAPI := s.GetAPI()

	etls, err := slsAPI.ListETLs(projectName, etlName, logstore)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "ListETLs", AlibabaCloudSdkGoERROR)
	}

	return etls, nil
}

// SlsETLExists checks if an ETL task exists in the specified project
func (s *SlsService) SlsETLExists(projectName, etlName string) (bool, error) {
	_, err := s.DescribeSlsETL(projectName, etlName)
	if err != nil {
		if NotFoundError(err) {
			return false, nil
		}
		return false, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", projectName, etlName), "DescribeSlsETL", AlibabaCloudSdkGoERROR)
	}

	return true, nil
}

// SlsETLStateRefreshFunc returns a StateRefreshFunc that monitors ETL state changes
func (s *SlsService) SlsETLStateRefreshFunc(projectName, etlName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		etl, err := s.DescribeSlsETL(projectName, etlName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if etl.Status == failState {
				return etl, failState, WrapError(Error(FailedToReachTargetStatus, failState))
			}
		}

		// Map ETL status to Terraform state
		state := "UNKNOWN"
		if etl.Status != "" {
			state = etl.Status
		}

		return etl, state, nil
	}
}

// WaitForSlsETL waits for an ETL task to reach the target state
func (s *SlsService) WaitForSlsETL(id string, status Status, timeout time.Duration) error {
	// Parse the ID to get project name and ETL name
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return WrapErrorf(Error("Invalid SLS ETL ID format"), IdMsg, id)
	}
	projectName, etlName := parts[0], parts[1]

	// Handle different target states
	var targets []string
	var pending []string
	var failStates []string

	switch status {
	case Deleted:
		// For deletion, we expect the resource to not be found
		targets = []string{}
		pending = []string{"RUNNING", "STOPPED", "STARTING", "STOPPING", "UNKNOWN"}
		failStates = []string{}
	case Running:
		targets = []string{"RUNNING"}
		pending = []string{"STOPPED", "STARTING", "STOPPING", "UNKNOWN"}
		failStates = []string{"Failed"}
	case Stopped:
		targets = []string{"STOPPED"}
		pending = []string{"RUNNING", "STOPPING", "STARTING", "UNKNOWN"}
		failStates = []string{"Failed"}
	case Available:
		targets = []string{"RUNNING", "STOPPED"}
		pending = []string{"STARTING", "STOPPING", "UNKNOWN"}
		failStates = []string{"Failed"}
	default:
		targets = []string{string(status)}
		pending = []string{"UNKNOWN"}
		failStates = []string{"Failed"}
	}

	stateConf := BuildStateConf(pending, targets, timeout, 5*time.Second, s.SlsETLStateRefreshFunc(projectName, etlName, failStates))
	stateConf.Delay = 5 * time.Second
	stateConf.MinTimeout = 3 * time.Second

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// Helper functions for converting between Terraform and SLS API types

// ConvertToSlsETLConfiguration converts Terraform ETL configuration to SLS ETL Configuration
func ConvertToSlsETLConfiguration(terraformConfig map[string]interface{}) *aliyunSlsAPI.ETLConfiguration {
	config := &aliyunSlsAPI.ETLConfiguration{}

	if script, ok := terraformConfig["script"].(string); ok {
		config.Script = script
	}

	if version, ok := terraformConfig["version"].(string); ok {
		config.Version = version
	}

	if logstore, ok := terraformConfig["logstore"].(string); ok {
		config.Logstore = logstore
	}

	// Convert parameters
	if parameters, ok := terraformConfig["parameters"].(map[string]interface{}); ok {
		paramStrings := make([]string, 0, len(parameters))
		for k, v := range parameters {
			if strVal, ok := v.(string); ok {
				paramStrings = append(paramStrings, fmt.Sprintf("%s=%s", k, strVal))
			}
		}
		config.Parameters = paramStrings
	}

	// Convert from time
	if fromTime, ok := terraformConfig["from_time"].(int); ok {
		config.FromTime = int64(fromTime)
	}

	// Convert to time
	if toTime, ok := terraformConfig["to_time"].(int); ok {
		config.ToTime = int64(toTime)
	}

	// Convert role arn
	if roleArn, ok := terraformConfig["role_arn"].(string); ok {
		config.RoleArn = roleArn
	}

	if sinks, ok := terraformConfig["sinks"].([]interface{}); ok {
		result := make([]aliyunSlsAPI.ETLSink, 0, len(sinks))
		for _, sinkRaw := range sinks {
			sinkMap, sinkOK := sinkRaw.(map[string]interface{})
			if !sinkOK {
				continue
			}
			sink := aliyunSlsAPI.ETLSink{}
			if name, ok := sinkMap["name"].(string); ok {
				sink.Name = name
			}
			if sinkType, ok := sinkMap["type"].(string); ok {
				sink.Type = sinkType
			}
			if project, ok := sinkMap["project"].(string); ok {
				sink.Project = project
			}
			if logstore, ok := sinkMap["logstore"].(string); ok {
				sink.Logstore = logstore
			}
			if roleArn, ok := sinkMap["role_arn"].(string); ok {
				sink.RoleArn = roleArn
			}
			if description, ok := sinkMap["description"].(string); ok {
				sink.Description = description
			}
			result = append(result, sink)
		}
		config.Sinks = result
	}

	return config
}

// ConvertFromSlsETLConfiguration converts SLS ETL Configuration to Terraform ETL configuration
func ConvertFromSlsETLConfiguration(config *aliyunSlsAPI.ETLConfiguration) map[string]interface{} {
	result := make(map[string]interface{})

	if config.Script != "" {
		result["script"] = config.Script
	}

	if config.Version != "" {
		result["version"] = config.Version
	}

	if config.Logstore != "" {
		result["logstore"] = config.Logstore
	}

	if config.Parameters != nil && len(config.Parameters) > 0 {
		result["parameters"] = convertETLParametersSliceToTerraformMap(config.Parameters)
	}

	if config.FromTime > 0 {
		result["from_time"] = int(config.FromTime)
	}

	if config.ToTime > 0 {
		result["to_time"] = int(config.ToTime)
	}

	if config.RoleArn != "" {
		result["role_arn"] = config.RoleArn
	}

	if len(config.Sinks) > 0 {
		sinks := make([]map[string]interface{}, 0, len(config.Sinks))
		for _, sink := range config.Sinks {
			sinks = append(sinks, map[string]interface{}{
				"name":        sink.Name,
				"type":        sink.Type,
				"project":     sink.Project,
				"logstore":    sink.Logstore,
				"role_arn":    sink.RoleArn,
				"description": sink.Description,
			})
		}
		result["sinks"] = sinks
	}

	return result
}

// ConvertToSlsETL converts Terraform ETL configuration to SLS ETL
func ConvertToSlsETL(d map[string]interface{}) *aliyunSlsAPI.ETL {
	etl := &aliyunSlsAPI.ETL{}

	if name, ok := d["name"].(string); ok {
		etl.Name = name
	}

	if displayName, ok := d["display_name"].(string); ok {
		etl.DisplayName = displayName
	}

	if description, ok := d["description"].(string); ok {
		etl.Description = description
	}

	if status, ok := d["status"].(string); ok {
		etl.Status = status
	}

	if createTime, ok := d["create_time"].(int); ok {
		etl.CreateTime = int64(createTime)
	}

	if scheduleData, ok := d["schedule"].([]interface{}); ok && len(scheduleData) > 0 {
		if scheduleMap, ok := scheduleData[0].(map[string]interface{}); ok {
			schedule := &aliyunSlsAPI.ETLSchedule{}
			if scheduleType, ok := scheduleMap["type"].(string); ok {
				schedule.Type = scheduleType
			}
			if interval, ok := scheduleMap["interval"].(string); ok {
				schedule.Interval = interval
			}
			etl.Schedule = schedule
		}
	}

	// Convert configuration
	if configData, ok := d["configuration"].([]interface{}); ok && len(configData) > 0 {
		if configMap, ok := configData[0].(map[string]interface{}); ok {
			etl.Configuration = ConvertToSlsETLConfiguration(configMap)
		}
	}

	return etl
}

// ConvertFromSlsETL converts SLS ETL to Terraform ETL configuration
func ConvertFromSlsETL(etl *aliyunSlsAPI.ETL) map[string]interface{} {
	result := make(map[string]interface{})

	if etl.Name != "" {
		result["name"] = etl.Name
	}

	if etl.DisplayName != "" {
		result["display_name"] = etl.DisplayName
	}

	if etl.Description != "" {
		result["description"] = etl.Description
	}

	if etl.Status != "" {
		result["status"] = etl.Status
	}

	if etl.CreateTime != 0 {
		result["create_time"] = int(etl.CreateTime)
	}

	if etl.Schedule != nil {
		result["schedule"] = []interface{}{map[string]interface{}{
			"type":     etl.Schedule.Type,
			"interval": etl.Schedule.Interval,
		}}
	}

	// Convert configuration
	if etl.Configuration != nil {
		configMap := ConvertFromSlsETLConfiguration(etl.Configuration)
		result["configuration"] = []interface{}{configMap}
	}

	return result
}
