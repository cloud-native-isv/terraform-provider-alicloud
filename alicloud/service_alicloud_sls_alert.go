package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateSlsAlert creates a new SLS alert using the CWS-Lib-Go API
func (s *SlsService) CreateSlsAlert(projectName string, alert *aliyunSlsAPI.Alert) error {
	if projectName == "" {
		return WrapError(fmt.Errorf("project name is required"))
	}
	if alert == nil {
		return WrapError(fmt.Errorf("alert cannot be nil"))
	}
	slsAPI := s.GetAPI()

	err := slsAPI.CreateAlert(projectName, alert)
	if err == nil {
		addDebugJson("CreateSlsAlert", alert)
	} else {
		return WrapError(fmt.Errorf("failed to create alert %s: %w", alert.Name, err))
	}

	return nil
}

// GetSlsAlert retrieves an SLS alert by name using the CWS-Lib-Go API
func (s *SlsService) GetSlsAlert(projectName string, alertName string) (*aliyunSlsAPI.Alert, error) {
	if projectName == "" {
		return nil, WrapError(fmt.Errorf("project name is required"))
	}
	if alertName == "" {
		return nil, WrapError(fmt.Errorf("alert name is required"))
	}

	slsAPI := s.GetAPI()

	// Get the alert
	alert, err := slsAPI.GetAlert(projectName, alertName)
	if err != nil {
		return nil, WrapError(fmt.Errorf("failed to get alert %s: %w", alertName, err))
	}

	return alert, nil
}

// UpdateSlsAlert updates an existing SLS alert using the CWS-Lib-Go API
func (s *SlsService) UpdateSlsAlert(projectName string, alertName string, alert *aliyunSlsAPI.Alert) error {
	if projectName == "" {
		return WrapError(fmt.Errorf("project name is required"))
	}
	if alertName == "" {
		return WrapError(fmt.Errorf("alert name is required"))
	}
	if alert == nil {
		return WrapError(fmt.Errorf("alert cannot be nil"))
	}
	slsAPI := s.GetAPI()

	// Update the alert
	err := slsAPI.UpdateAlert(projectName, alertName, alert)
	if err != nil {
		return WrapError(fmt.Errorf("failed to update alert %s: %w", alertName, err))
	}

	return nil
}

// DeleteSlsAlert deletes an SLS alert using the CWS-Lib-Go API
func (s *SlsService) DeleteSlsAlert(projectName string, alertName string) error {
	if projectName == "" {
		return WrapError(fmt.Errorf("project name is required"))
	}
	if alertName == "" {
		return WrapError(fmt.Errorf("alert name is required"))
	}
	slsAPI := s.GetAPI()

	// Delete the alert
	err := slsAPI.DeleteAlert(projectName, alertName)
	if err != nil {
		return WrapError(fmt.Errorf("failed to delete alert %s: %w", alertName, err))
	}

	return nil
}

// ListSlsAlerts lists all alerts in a project using the CWS-Lib-Go API
func (s *SlsService) ListSlsAlerts(projectName string, pagination *common.PaginationRequest) ([]*aliyunSlsAPI.Alert, error) {
	if projectName == "" {
		return nil, WrapError(fmt.Errorf("project name is required"))
	}

	slsAPI := s.GetAPI()

	// List alerts
	alerts, err := slsAPI.ListAlerts(projectName, pagination)
	if err != nil {
		return nil, WrapError(fmt.Errorf("failed to list alerts: %w", err))
	}

	return alerts, nil
}

// EnableSlsAlert enables an SLS alert using the CWS-Lib-Go API
func (s *SlsService) EnableSlsAlert(projectName string, alertName string) error {
	if projectName == "" {
		return WrapError(fmt.Errorf("project name is required"))
	}
	if alertName == "" {
		return WrapError(fmt.Errorf("alert name is required"))
	}
	slsAPI := s.GetAPI()

	// Enable the alert
	err := slsAPI.EnableAlert(projectName, alertName)
	if err != nil {
		return WrapError(fmt.Errorf("failed to enable alert %s: %w", alertName, err))
	}

	return nil
}

// DisableSlsAlert disables an SLS alert using the CWS-Lib-Go API
func (s *SlsService) DisableSlsAlert(projectName string, alertName string) error {
	if projectName == "" {
		return WrapError(fmt.Errorf("project name is required"))
	}
	if alertName == "" {
		return WrapError(fmt.Errorf("alert name is required"))
	}
	slsAPI := s.GetAPI()

	// Disable the alert
	err := slsAPI.DisableAlert(projectName, alertName)
	if err != nil {
		return WrapError(fmt.Errorf("failed to disable alert %s: %w", alertName, err))
	}

	return nil
}

// WaitForSlsAlert waits for an SLS alert to reach a desired state
func (s *SlsService) WaitForSlsAlert(projectName string, alertName string, status string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating", "Updating"},
		Target:     []string{status},
		Refresh:    s.SlsAlertStateRefreshFunc(projectName, alertName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, alertName)
	}

	return nil
}

// SlsAlertStateRefreshFunc returns a StateRefreshFunc for SLS alert status
func (s *SlsService) SlsAlertStateRefreshFunc(projectName string, alertName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		alert, err := s.GetSlsAlert(projectName, alertName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Handle pointer type for Status field
		status := ""
		if alert.Status != nil {
			status = *alert.Status
		}

		return alert, status, nil
	}
}

// DescribeSlsAlert describes an SLS alert by constructing ID from project and alert name
func (s *SlsService) DescribeSlsAlert(id string) (map[string]interface{}, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid SLS alert ID format: %s, expected format: project_name:alert_name", id))
	}

	projectName := parts[0]
	alertName := parts[1]

	alert, err := s.GetSlsAlert(projectName, alertName)
	if err != nil {
		return nil, WrapError(err)
	}

	// Convert alert to map for compatibility with existing resource implementation
	// Handle pointer types and removed fields properly
	result := map[string]interface{}{}

	if alert.Name != nil {
		result["name"] = *alert.Name
	}
	if alert.DisplayName != nil {
		result["displayName"] = *alert.DisplayName
	}
	if alert.Description != nil {
		result["description"] = *alert.Description
	}
	if alert.Status != nil {
		result["status"] = *alert.Status
	}
	if alert.CreateTime != nil {
		result["createTime"] = *alert.CreateTime
	}
	if alert.LastModifiedTime != nil {
		result["lastModifyTime"] = *alert.LastModifiedTime
	}

	// Add configuration if present
	if alert.Configuration != nil {
		configMap := convertAlertConfigurationToMap(alert.Configuration)
		result["configuration"] = configMap
	}

	// Add schedule if present
	if alert.Schedule != nil {
		scheduleMap := convertScheduleToMap(alert.Schedule)
		result["schedule"] = scheduleMap
	}

	return result, nil
}

// Helper function to convert AlertConfig to map
func convertAlertConfigurationToMap(config *aliyunSlsAPI.AlertConfig) map[string]interface{} {
	configMap := make(map[string]interface{})

	if config.AutoAnnotation != nil {
		configMap["autoAnnotation"] = *config.AutoAnnotation
	}
	if config.Dashboard != nil {
		configMap["dashboard"] = *config.Dashboard
	}
	if config.MuteUntil != nil {
		configMap["muteUntil"] = *config.MuteUntil
	}
	if config.NoDataFire != nil {
		configMap["noDataFire"] = *config.NoDataFire
	}
	if config.NoDataSeverity != nil {
		configMap["noDataSeverity"] = *config.NoDataSeverity
	}
	if config.SendResolved != nil {
		configMap["sendResolved"] = *config.SendResolved
	}

	// Convert query list
	if config.QueryList != nil {
		queryListMaps := make([]map[string]interface{}, 0, len(config.QueryList))
		for _, query := range config.QueryList {
			queryMap := map[string]interface{}{
				"chartTitle":   query.ChartTitle,
				"project":      query.Project,
				"store":        query.Store,
				"storeType":    query.StoreType,
				"query":        query.Query,
				"start":        query.Start,
				"end":          query.End,
				"timeSpanType": query.TimeSpanType,
				"region":       query.Region,
				"roleArn":      query.RoleArn,
				"dashboardId":  query.DashboardId,
			}
			queryListMaps = append(queryListMaps, queryMap)
		}
		configMap["queryList"] = queryListMaps
	}

	// Convert severity configurations
	if config.SeverityConfigurations != nil {
		severityConfigsMaps := make([]map[string]interface{}, 0, len(config.SeverityConfigurations))
		for _, severityConfig := range config.SeverityConfigurations {
			severityMap := map[string]interface{}{
				"severity": severityConfig.Severity,
				"evalCondition": map[string]interface{}{
					"condition":      severityConfig.EvalCondition.Condition,
					"countCondition": severityConfig.EvalCondition.CountCondition,
				},
			}
			severityConfigsMaps = append(severityConfigsMaps, severityMap)
		}
		configMap["severityConfigurations"] = severityConfigsMaps
	}

	// Convert labels
	if config.Labels != nil {
		labelsMaps := make([]map[string]interface{}, 0, len(config.Labels))
		for _, label := range config.Labels {
			labelMap := map[string]interface{}{}
			if label.Key != nil {
				labelMap["key"] = *label.Key
			}
			if label.Value != nil {
				labelMap["value"] = *label.Value
			}
			labelsMaps = append(labelsMaps, labelMap)
		}
		configMap["labels"] = labelsMaps
	}

	// Convert annotations
	if config.Annotations != nil {
		annotationsMaps := make([]map[string]interface{}, 0, len(config.Annotations))
		for _, annotation := range config.Annotations {
			annotationMap := map[string]interface{}{}
			if annotation.Key != nil {
				annotationMap["key"] = *annotation.Key
			}
			if annotation.Value != nil {
				annotationMap["value"] = *annotation.Value
			}
			annotationsMaps = append(annotationsMaps, annotationMap)
		}
		configMap["annotations"] = annotationsMaps
	}

	// Convert join configurations
	if config.JoinConfigurations != nil {
		joinConfigsMaps := make([]map[string]interface{}, 0, len(config.JoinConfigurations))
		for _, joinConfig := range config.JoinConfigurations {
			joinMap := map[string]interface{}{
				"type":      joinConfig.Type,
				"condition": joinConfig.Condition,
			}
			joinConfigsMaps = append(joinConfigsMaps, joinMap)
		}
		configMap["joinConfigurations"] = joinConfigsMaps
	}

	// Convert group configuration
	if config.GroupConfiguration != nil {
		groupMap := map[string]interface{}{
			"type":   config.GroupConfiguration.Type,
			"fields": config.GroupConfiguration.Fields,
		}
		configMap["groupConfiguration"] = groupMap
	}

	// Convert policy configuration
	if config.PolicyConfiguration != nil {
		policyMap := map[string]interface{}{
			"alertPolicyId":  config.PolicyConfiguration.AlertPolicyId,
			"actionPolicyId": config.PolicyConfiguration.ActionPolicyId,
			"repeatInterval": config.PolicyConfiguration.RepeatInterval,
		}
		configMap["policyConfiguration"] = policyMap
	}

	// Convert template configuration
	if config.TemplateConfiguration != nil {
		templateMap := map[string]interface{}{
			"id":          config.TemplateConfiguration.Id,
			"type":        config.TemplateConfiguration.Type,
			"lang":        config.TemplateConfiguration.Lang,
			"version":     config.TemplateConfiguration.Version,
			"tokens":      config.TemplateConfiguration.Tokens,
			"annotations": config.TemplateConfiguration.Annotations,
		}
		configMap["templateConfiguration"] = templateMap
	}

	// Convert sink configurations
	if config.SinkAlerthub != nil {
		configMap["sinkAlerthub"] = map[string]interface{}{
			"enabled": config.SinkAlerthub.Enabled,
		}
	}

	if config.SinkCms != nil {
		configMap["sinkCms"] = map[string]interface{}{
			"enabled": config.SinkCms.Enabled,
		}
	}

	if config.SinkEventStore != nil {
		sinkEventStoreMap := map[string]interface{}{}
		if config.SinkEventStore.Enabled != nil {
			sinkEventStoreMap["enabled"] = *config.SinkEventStore.Enabled
		}
		if config.SinkEventStore.Endpoint != nil {
			sinkEventStoreMap["endpoint"] = *config.SinkEventStore.Endpoint
		}
		if config.SinkEventStore.Project != nil {
			sinkEventStoreMap["project"] = *config.SinkEventStore.Project
		}
		if config.SinkEventStore.EventStore != nil {
			sinkEventStoreMap["eventStore"] = *config.SinkEventStore.EventStore
		}
		if config.SinkEventStore.RoleArn != nil {
			sinkEventStoreMap["roleArn"] = *config.SinkEventStore.RoleArn
		}
		configMap["sinkEventStore"] = sinkEventStoreMap
	}

	// Convert tags
	if config.Tags != nil {
		tagsList := make([]string, 0, len(config.Tags))
		for _, tag := range config.Tags {
			if tag != nil {
				tagsList = append(tagsList, *tag)
			}
		}
		configMap["tags"] = tagsList
	}

	return configMap
}

// Helper function to convert Schedule to map
func convertScheduleToMap(schedule *aliyunSlsAPI.Schedule) map[string]interface{} {
	scheduleMap := map[string]interface{}{
		"type":           schedule.Type,
		"interval":       schedule.Interval,
		"cronExpression": schedule.CronExpression,
		"dayOfWeek":      schedule.DayOfWeek,
		"hour":           schedule.Hour,
		"delay":          schedule.Delay,
		"timeZone":       schedule.TimeZone,
		"runImmediately": schedule.RunImmediately,
	}

	// Handle legacy field name
	if schedule.Cron != "" {
		scheduleMap["cronExpression"] = schedule.Cron
	}

	return scheduleMap
}
