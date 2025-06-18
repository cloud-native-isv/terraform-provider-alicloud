package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// OSS Ingestion Service Methods

// DescribeSlsOSSIngestion - Get OSS ingestion job configuration
func (s *SlsService) DescribeSlsOSSIngestion(id string) (*aliyunSlsAPI.Ingestion, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err := WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return nil, err
	}

	projectName := parts[0]
	ingestionName := parts[1]

	ingestion, err := s.aliyunSlsAPI.GetOSSIngestion(projectName, ingestionName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") || strings.Contains(err.Error(), "not found") {
			return nil, WrapErrorf(NotFoundErr("OSSIngestion", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return ingestion, nil
}

// CreateSlsOSSIngestion creates a new OSS ingestion
func (s *SlsService) CreateSlsOSSIngestion(project string, ingestion *aliyunSlsAPI.Ingestion) (*aliyunSlsAPI.Ingestion, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	result, err := s.aliyunSlsAPI.CreateOSSIngestion(project, ingestion)
	if err != nil {
		ingestionName := "unknown"
		if ingestion != nil {
			ingestionName = ingestion.ScheduledJob.BaseJob.Name
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, ingestionName, "CreateOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

// UpdateSlsOSSIngestion updates an existing OSS ingestion
func (s *SlsService) UpdateSlsOSSIngestion(project, ingestionName string, ingestion *aliyunSlsAPI.Ingestion) (*aliyunSlsAPI.Ingestion, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	result, err := s.aliyunSlsAPI.UpdateOSSIngestion(project, ingestionName, ingestion)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, ingestionName, "UpdateOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

// DeleteSlsOSSIngestion deletes an OSS ingestion
func (s *SlsService) DeleteSlsOSSIngestion(project, ingestionName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteOSSIngestion(project, ingestionName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ingestionName, "DeleteOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StartSlsOSSIngestion starts an OSS ingestion
func (s *SlsService) StartSlsOSSIngestion(project, ingestionName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.StartOSSIngestion(project, ingestionName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ingestionName, "StartOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StopSlsOSSIngestion stops an OSS ingestion
func (s *SlsService) StopSlsOSSIngestion(project, ingestionName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.StopOSSIngestion(project, ingestionName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ingestionName, "StopOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsOSSIngestions lists OSS ingestions in a project
func (s *SlsService) ListSlsOSSIngestions(project, logstore string, offset, size int32) ([]*aliyunSlsAPI.Ingestion, int32, error) {
	if s.aliyunSlsAPI == nil {
		return nil, 0, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	ingestions, total, err := s.aliyunSlsAPI.ListOSSIngestions(project, logstore, offset, size)
	if err != nil {
		return nil, 0, WrapErrorf(err, DefaultErrorMsg, project, "ListOSSIngestions", AlibabaCloudSdkGoERROR)
	}

	return ingestions, total, nil
}

// SlsOSSIngestionStateRefreshFunc returns a StateRefreshFunc for OSS ingestion status monitoring
func (s *SlsService) SlsOSSIngestionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ingestion, err := s.DescribeSlsOSSIngestion(id)
		if err != nil {
			if NotFoundError(err) {
				return ingestion, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string

		// Extract status from ingestion object based on field
		switch field {
		case "status":
			currentStatus = ingestion.ScheduledJob.BaseJob.Status
		case "name":
			currentStatus = ingestion.ScheduledJob.BaseJob.Name
		case "displayName":
			currentStatus = ingestion.ScheduledJob.BaseJob.DisplayName
		case "description":
			currentStatus = ingestion.ScheduledJob.BaseJob.Description
		default:
			// For other fields, convert to map and use jsonpath
			ingestionMap := make(map[string]interface{})
			ingestionMap["name"] = ingestion.ScheduledJob.BaseJob.Name
			ingestionMap["displayName"] = ingestion.ScheduledJob.BaseJob.DisplayName
			ingestionMap["description"] = ingestion.ScheduledJob.BaseJob.Description
			ingestionMap["status"] = ingestion.ScheduledJob.BaseJob.Status

			if ingestion.IngestionConfiguration != nil {
				ingestionMap["configuration"] = convertIngestionConfiguration(ingestion.IngestionConfiguration)
			}
			if ingestion.ScheduledJob.Schedule != nil {
				ingestionMap["schedule"] = convertIngestionSchedule(ingestion.ScheduledJob.Schedule)
			}

			v, jsonErr := jsonpath.Get(field, ingestionMap)
			if jsonErr != nil {
				return nil, "", WrapErrorf(jsonErr, FailedGetAttributeMsg, id, field)
			}
			currentStatus = fmt.Sprint(v)
		}

		// Handle special field prefix for existence check
		if strings.HasPrefix(field, "#") {
			if ingestion != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return ingestion, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return ingestion, currentStatus, nil
	}
}

// WaitForSlsOSSIngestionStatus waits for OSS ingestion to reach target status
func (s *SlsService) WaitForSlsOSSIngestionStatus(project, ingestionName, targetStatus string, timeout time.Duration) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"STARTING", "STOPPING", "CREATING", "UPDATING"},
		Target:     []string{targetStatus},
		Refresh:    s.SlsOSSIngestionStateRefreshFunc(fmt.Sprintf("%s:%s", project, ingestionName), "status", []string{"FAILED", "ERROR"}),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", project, ingestionName))
}

// Helper functions for data conversion

// convertIngestionConfiguration converts our IngestionConfiguration to map
func convertIngestionConfiguration(config *aliyunSlsAPI.IngestionConfiguration) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["logstore"] = config.LogStore

	if config.Source != nil {
		source := make(map[string]interface{})

		if bucket, ok := config.Source["bucket"].(string); ok {
			source["bucket"] = bucket
		}
		if endpoint, ok := config.Source["endpoint"].(string); ok {
			source["endpoint"] = endpoint
		}
		if roleARN, ok := config.Source["roleARN"].(string); ok {
			source["role_arn"] = roleARN
		}
		if prefix, ok := config.Source["prefix"].(string); ok {
			source["prefix"] = prefix
		}
		if pattern, ok := config.Source["pattern"].(string); ok {
			source["pattern"] = pattern
		}
		if encoding, ok := config.Source["encoding"].(string); ok {
			source["encoding"] = encoding
		}
		if compressionCodec, ok := config.Source["compressionCodec"].(string); ok {
			source["compression_codec"] = compressionCodec
		}
		if format, ok := config.Source["format"].(map[string]interface{}); ok {
			source["format"] = format
		}
		if timeField, ok := config.Source["timeField"].(string); ok {
			source["time_field"] = timeField
		}
		if timeFormat, ok := config.Source["timeFormat"].(string); ok {
			source["time_format"] = timeFormat
		}
		if timePattern, ok := config.Source["timePattern"].(string); ok {
			source["time_pattern"] = timePattern
		}
		if timeZone, ok := config.Source["timeZone"].(string); ok {
			source["time_zone"] = timeZone
		}
		if interval, ok := config.Source["interval"].(string); ok {
			source["interval"] = interval
		}
		if startTime, ok := config.Source["startTime"].(int64); ok {
			source["start_time"] = startTime
		}
		if endTime, ok := config.Source["endTime"].(int64); ok {
			source["end_time"] = endTime
		}
		if useMetaIndex, ok := config.Source["useMetaIndex"].(bool); ok {
			source["use_meta_index"] = useMetaIndex
		}
		if restoreObjectEnabled, ok := config.Source["restoreObjectEnabled"].(bool); ok {
			source["restore_object_enabled"] = restoreObjectEnabled
		}

		result["source"] = source
	}

	return result
}

// convertIngestionSchedule converts our Schedule to map
func convertIngestionSchedule(schedule *aliyunSlsAPI.Schedule) map[string]interface{} {
	if schedule == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["type"] = schedule.Type

	if schedule.Interval != "" {
		result["interval"] = schedule.Interval
	}
	result["run_immediately"] = schedule.RunImmediately

	if schedule.Delay != 0 {
		result["delay"] = schedule.Delay
	}
	if schedule.TimeZone != "" {
		result["time_zone"] = schedule.TimeZone
	}
	if schedule.Cron != "" {
		result["cron"] = schedule.Cron
	}
	if schedule.DayOfWeek != 0 {
		result["day_of_week"] = schedule.DayOfWeek
	}
	if schedule.Hour != 0 {
		result["hour"] = schedule.Hour
	}

	return result
}

// buildOSSIngestionFromMap builds OSS ingestion from Terraform map
func buildOSSIngestionFromMap(configMap map[string]interface{}, scheduleMap map[string]interface{}, name, displayName, description string) *aliyunSlsAPI.Ingestion {
	ingestion := &aliyunSlsAPI.Ingestion{
		ScheduledJob: aliyunSlsAPI.ScheduledJob{
			BaseJob: aliyunSlsAPI.BaseJob{
				Name:        name,
				DisplayName: displayName,
				Description: description,
				Type:        aliyunSlsAPI.INGESTION_JOB,
			},
		},
	}

	// Build configuration
	if configMap != nil {
		config := &aliyunSlsAPI.IngestionConfiguration{
			Version: "2.0",
			Source:  make(map[string]interface{}),
		}

		if logstore, ok := configMap["logstore"].(string); ok {
			config.LogStore = logstore
		}

		if sourceMap, ok := configMap["source"].(map[string]interface{}); ok {
			for key, value := range sourceMap {
				// Convert key names for internal use
				switch key {
				case "role_arn":
					config.Source["roleARN"] = value
				case "compression_codec":
					config.Source["compressionCodec"] = value
				case "time_field":
					config.Source["timeField"] = value
				case "time_format":
					config.Source["timeFormat"] = value
				case "time_pattern":
					config.Source["timePattern"] = value
				case "time_zone":
					config.Source["timeZone"] = value
				case "start_time":
					config.Source["startTime"] = value
				case "end_time":
					config.Source["endTime"] = value
				case "use_meta_index":
					config.Source["useMetaIndex"] = value
				case "restore_object_enabled":
					config.Source["restoreObjectEnabled"] = value
				default:
					config.Source[key] = value
				}
			}
		}

		ingestion.IngestionConfiguration = config
	}

	// Build schedule
	if scheduleMap != nil {
		schedule := &aliyunSlsAPI.Schedule{}

		if scheduleType, ok := scheduleMap["type"].(string); ok {
			schedule.Type = scheduleType
		}
		if interval, ok := scheduleMap["interval"].(string); ok {
			schedule.Interval = interval
		}
		if runImmediately, ok := scheduleMap["run_immediately"].(bool); ok {
			schedule.RunImmediately = runImmediately
		}
		if delay, ok := scheduleMap["delay"].(int32); ok {
			schedule.Delay = delay
		}
		if timeZone, ok := scheduleMap["time_zone"].(string); ok {
			schedule.TimeZone = timeZone
		}
		if cron, ok := scheduleMap["cron"].(string); ok {
			schedule.Cron = cron
		}
		if dayOfWeek, ok := scheduleMap["day_of_week"].(int32); ok {
			schedule.DayOfWeek = dayOfWeek
		}
		if hour, ok := scheduleMap["hour"].(int32); ok {
			schedule.Hour = hour
		}

		ingestion.ScheduledJob.Schedule = schedule
	}

	return ingestion
}

// DescribeSlsIngestion - Get ingestion job configuration
func (s *SlsService) DescribeSlsIngestion(id string) (*aliyunSlsAPI.Ingestion, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err := WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return nil, err
	}

	projectName := parts[0]
	ingestionName := parts[2]

	ingestion, err := s.aliyunSlsAPI.GetOSSIngestion(projectName, ingestionName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return nil, WrapErrorf(NotFoundErr("Ingestion", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetIngestion", AlibabaCloudSdkGoERROR)
	}

	return ingestion, nil
}

func (s *SlsService) SlsIngestionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ingestion, err := s.DescribeSlsIngestion(id)
		if err != nil {
			if NotFoundError(err) {
				return ingestion, "", nil
			}
			return nil, "", WrapError(err)
		}

		var currentStatus string

		// Extract status from ingestion object based on field
		switch field {
		case "status":
			currentStatus = ingestion.ScheduledJob.BaseJob.Status
		case "name":
			currentStatus = ingestion.ScheduledJob.BaseJob.Name
		case "displayName":
			currentStatus = ingestion.ScheduledJob.BaseJob.DisplayName
		case "description":
			currentStatus = ingestion.ScheduledJob.BaseJob.Description
		default:
			// For other fields, convert to map and use jsonpath
			ingestionMap := make(map[string]interface{})
			ingestionMap["name"] = ingestion.ScheduledJob.BaseJob.Name
			ingestionMap["displayName"] = ingestion.ScheduledJob.BaseJob.DisplayName
			ingestionMap["description"] = ingestion.ScheduledJob.BaseJob.Description
			ingestionMap["status"] = ingestion.ScheduledJob.BaseJob.Status

			if ingestion.IngestionConfiguration != nil {
				ingestionMap["configuration"] = convertIngestionConfiguration(ingestion.IngestionConfiguration)
			}
			if ingestion.ScheduledJob.Schedule != nil {
				ingestionMap["schedule"] = convertIngestionSchedule(ingestion.ScheduledJob.Schedule)
			}

			v, jsonErr := jsonpath.Get(field, ingestionMap)
			if jsonErr != nil {
				return nil, "", WrapErrorf(jsonErr, FailedGetAttributeMsg, id, field)
			}
			currentStatus = fmt.Sprint(v)
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return ingestion, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return ingestion, currentStatus, nil
	}
}
