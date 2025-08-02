package alicloud

import (
	"strings"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsScheduledSQL retrieves a scheduled SQL job by ID
func (s *SlsService) DescribeSlsScheduledSQL(id string) (*aliyunSlsAPI.ScheduledSQL, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapErrorf(Error("invalid scheduled SQL ID format"), DefaultErrorMsg, "alicloud_sls_scheduled_sql", "DescribeSlsScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	projectName := parts[0]
	scheduledSQLName := parts[1]

	// Get scheduled SQL from API
	scheduledSQL, err := s.GetAPI().GetScheduledSQL(projectName, scheduledSQLName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "GetScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	if scheduledSQL == nil {
		return nil, WrapErrorf(Error("scheduled SQL not found"), NotFoundMsg, AlibabaCloudSdkGoERROR)
	}

	return scheduledSQL, nil
}

// CreateSlsScheduledSQL creates a new scheduled SQL job
func (s *SlsService) CreateSlsScheduledSQL(projectName string, scheduledSQL *aliyunSlsAPI.ScheduledSQL) error {
	// Create scheduled SQL
	err := s.GetAPI().CreateScheduledSQL(projectName, scheduledSQL)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "CreateScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsScheduledSQL updates an existing scheduled SQL job
func (s *SlsService) UpdateSlsScheduledSQL(projectName string, scheduledSQLName string, scheduledSQL *aliyunSlsAPI.ScheduledSQL) error {
	// Update scheduled SQL
	err := s.GetAPI().UpdateScheduledSQL(projectName, scheduledSQLName, scheduledSQL)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "UpdateScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsScheduledSQL deletes a scheduled SQL job
func (s *SlsService) DeleteSlsScheduledSQL(projectName string, scheduledSQLName string) error {
	// Delete scheduled SQL
	err := s.GetAPI().DeleteScheduledSQL(projectName, scheduledSQLName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "DeleteScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// EnableSlsScheduledSQL enables a scheduled SQL job
func (s *SlsService) EnableSlsScheduledSQL(projectName string, scheduledSQLName string) error {
	// Enable scheduled SQL
	err := s.GetAPI().EnableScheduledSQL(projectName, scheduledSQLName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "EnableScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DisableSlsScheduledSQL disables a scheduled SQL job
func (s *SlsService) DisableSlsScheduledSQL(projectName string, scheduledSQLName string) error {
	// Disable scheduled SQL
	err := s.GetAPI().DisableScheduledSQL(projectName, scheduledSQLName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "DisableScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsScheduledSQLs lists all scheduled SQL jobs in a project
func (s *SlsService) ListSlsScheduledSQLs(projectName string, scheduledSQLName string, logstore string) ([]*aliyunSlsAPI.ScheduledSQL, error) {
	// List scheduled SQLs
	scheduledSQLs, err := s.GetAPI().ListScheduledSQLs(projectName, scheduledSQLName, logstore)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "ListScheduledSQLs", AlibabaCloudSdkGoERROR)
	}

	return scheduledSQLs, nil
}

// SlsScheduledSQLStateRefreshFunc returns a StateRefreshFunc for monitoring scheduled SQL status
func (s *SlsService) SlsScheduledSQLStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Parse the ID to get project and scheduled SQL name
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return nil, "", WrapErrorf(Error("invalid scheduled SQL ID format"), DefaultErrorMsg, "alicloud_sls_scheduled_sql", "SlsScheduledSQLStateRefreshFunc", AlibabaCloudSdkGoERROR)
		}

		projectName := parts[0]
		scheduledSQLName := parts[1]

		// Get scheduled SQL from API
		scheduledSQL, err := s.GetAPI().GetScheduledSQL(projectName, scheduledSQLName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "GetScheduledSQL", AlibabaCloudSdkGoERROR)
		}

		if scheduledSQL == nil {
			return nil, "", nil
		}

		// Convert to map format
		object := convertScheduledSQLToMap(scheduledSQL)

		// Extract the state from the specified field
		var state string
		if field == "$.state" || field == "$.status" {
			if scheduledSQL.Status != "" {
				state = scheduledSQL.Status
			} else {
				state = "UNKNOWN"
			}
		} else {
			state = "COMPLETE" // Default state for other operations
		}

		// Check if current state is in fail states
		for _, failState := range failStates {
			if state == failState {
				return object, state, WrapErrorf(Error("scheduled SQL is in failed state"), DefaultErrorMsg, "alicloud_sls_scheduled_sql", "SlsScheduledSQLStateRefreshFunc", AlibabaCloudSdkGoERROR)
			}
		}

		return object, state, nil
	}
}

// convertScheduledSQLToMap converts ScheduledSQL struct to map format expected by Terraform
func convertScheduledSQLToMap(scheduledSQL *aliyunSlsAPI.ScheduledSQL) map[string]interface{} {
	result := make(map[string]interface{})

	if scheduledSQL == nil {
		return result
	}

	// Basic fields
	if scheduledSQL.Name != "" {
		result["name"] = scheduledSQL.Name
	}
	if scheduledSQL.DisplayName != "" {
		result["displayName"] = scheduledSQL.DisplayName
	}
	if scheduledSQL.Description != "" {
		result["description"] = scheduledSQL.Description
	}
	if scheduledSQL.Status != "" {
		result["status"] = scheduledSQL.Status
	}
	if scheduledSQL.ScheduleId != "" {
		result["scheduleId"] = scheduledSQL.ScheduleId
	}
	if scheduledSQL.CreateTime > 0 {
		result["createTime"] = scheduledSQL.CreateTime
	}
	if scheduledSQL.LastModifyTime > 0 {
		result["lastModifyTime"] = scheduledSQL.LastModifyTime
	}

	// Schedule configuration
	if scheduledSQL.Schedule != nil {
		scheduleMap := make(map[string]interface{})
		if scheduledSQL.Schedule.Type != "" {
			scheduleMap["type"] = scheduledSQL.Schedule.Type
		}
		if scheduledSQL.Schedule.Interval != "" {
			scheduleMap["interval"] = scheduledSQL.Schedule.Interval
		}
		if scheduledSQL.Schedule.CronExpression != "" {
			scheduleMap["cronExpression"] = scheduledSQL.Schedule.CronExpression
		}
		if scheduledSQL.Schedule.TimeZone != "" {
			scheduleMap["timeZone"] = scheduledSQL.Schedule.TimeZone
		}
		if scheduledSQL.Schedule.Delay > 0 {
			scheduleMap["delay"] = scheduledSQL.Schedule.Delay
		}
		scheduleMap["runImmediately"] = scheduledSQL.Schedule.RunImmediately
		result["schedule"] = scheduleMap
	}

	// Configuration
	if scheduledSQL.Configuration != nil {
		configMap := make(map[string]interface{})
		config := scheduledSQL.Configuration

		if config.DataFormat != "" {
			configMap["dataFormat"] = config.DataFormat
		}
		if config.DestEndpoint != "" {
			configMap["destEndpoint"] = config.DestEndpoint
		}
		if config.DestLogstore != "" {
			configMap["destLogstore"] = config.DestLogstore
		}
		if config.DestProject != "" {
			configMap["destProject"] = config.DestProject
		}
		if config.DestRoleArn != "" {
			configMap["destRoleArn"] = config.DestRoleArn
		}
		if config.FromTime > 0 {
			configMap["fromTime"] = config.FromTime
		}
		if config.FromTimeExpr != "" {
			configMap["fromTimeExpr"] = config.FromTimeExpr
		}
		if config.MaxRetries > 0 {
			configMap["maxRetries"] = config.MaxRetries
		}
		if config.MaxRunTimeInSeconds > 0 {
			configMap["maxRunTimeInSeconds"] = config.MaxRunTimeInSeconds
		}
		if config.Parameters != nil {
			configMap["parameters"] = config.Parameters
		}
		if config.ResourcePool != "" {
			configMap["resourcePool"] = config.ResourcePool
		}
		if config.RoleArn != "" {
			configMap["roleArn"] = config.RoleArn
		}
		if config.Script != "" {
			configMap["script"] = config.Script
		}
		if config.SourceLogstore != "" {
			configMap["sourceLogstore"] = config.SourceLogstore
		}
		if config.SqlType != "" {
			configMap["sqlType"] = config.SqlType
		}
		if config.ToTime > 0 {
			configMap["toTime"] = config.ToTime
		}
		if config.ToTimeExpr != "" {
			configMap["toTimeExpr"] = config.ToTimeExpr
		}

		result["configuration"] = configMap
	}

	return result
}
