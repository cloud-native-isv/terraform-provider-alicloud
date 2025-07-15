package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsAlertRule describes ARMS alert rule
func (s *ArmsService) DescribeArmsAlertRule(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		alertId, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			alertRule, err := s.armsAPI.GetAlertRule(alertId)
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"AlertId":               alertRule.AlertId,
					"AlertName":             alertRule.AlertName,
					"AlertCheckType":        alertRule.AlertCheckType,
					"AlertGroup":            alertRule.AlertGroup,
					"AlertStatus":           alertRule.AlertStatus,
					"AlertType":             alertRule.AlertType,
					"Level":                 alertRule.Level,
					"Severity":              alertRule.Level, // Map Level to Severity for backward compatibility
					"Message":               alertRule.Message,
					"Duration":              alertRule.Duration,
					"PromQL":                alertRule.PromQL,
					"ClusterId":             alertRule.ClusterId,
					"MetricsType":           alertRule.MetricsType,
					"NotifyStrategy":        alertRule.NotifyStrategy,
					"AutoAddNewApplication": alertRule.AutoAddNewApplication,
					"RegionId":              alertRule.RegionId,
					"UserId":                alertRule.UserId,
					"CreatedTime":           alertRule.CreatedTime,
					"UpdatedTime":           alertRule.UpdatedTime,
					"CreateTime":            alertRule.CreatedTime, // Map CreatedTime to CreateTime for backward compatibility
					"Extend":                alertRule.Extend,
					"Pids":                  alertRule.Pids,
					"Annotations":           alertRule.Annotations,
					"Labels":                alertRule.Labels,
					"Tags":                  alertRule.Tags,
					"Filters":               alertRule.Filters,
					"AlertRuleContent":      alertRule.AlertRuleContent,
					"State":                 alertRule.AlertStatus, // Map AlertStatus to State for consistency
					"Describe":              alertRule.Message,     // Map Message to Describe for backward compatibility
					"Owner":                 "",                    // Default empty, can be extracted from Extend if needed
					"Handler":               "",                    // Default empty, can be extracted from Extend if needed
					"Solution":              "",                    // Default empty, can be extracted from Extend if needed
					"DispatchRuleId":        0,                     // Default 0, will be set if available
					"DispatchRuleName":      "",                    // Default empty, will be set if available
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListAlerts"
	client := s.client

	request := map[string]interface{}{
		"Page":     1,
		"Size":     1,
		"RegionId": s.client.RegionId,
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.PageBean.ListAlerts", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.ListAlerts", response)
	}

	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}

	// Find the alert with matching ID
	alertIdInt, _ := strconv.ParseInt(id, 10, 64)
	for _, alert := range v.([]interface{}) {
		alertMap := alert.(map[string]interface{})
		if alertId, ok := alertMap["AlertId"]; ok {
			if int64(alertId.(float64)) == alertIdInt {
				// Convert RPC response to standardized format matching AlertRule struct
				standardizedAlert := make(map[string]interface{})

				// Core fields from AlertRule struct
				if alertId, ok := alertMap["AlertId"]; ok {
					standardizedAlert["AlertId"] = int64(alertId.(float64))
				}
				if alertName, ok := alertMap["AlertName"]; ok && alertName != nil {
					standardizedAlert["AlertName"] = alertName.(string)
				}
				if alertCheckType, ok := alertMap["AlertCheckType"]; ok && alertCheckType != nil {
					standardizedAlert["AlertCheckType"] = alertCheckType.(string)
				}
				if alertGroup, ok := alertMap["AlertGroup"]; ok && alertGroup != nil {
					if alertGroupFloat, ok := alertGroup.(float64); ok {
						standardizedAlert["AlertGroup"] = int64(alertGroupFloat)
					}
				}
				if alertStatus, ok := alertMap["AlertStatus"]; ok && alertStatus != nil {
					standardizedAlert["AlertStatus"] = alertStatus.(string)
					standardizedAlert["State"] = alertStatus.(string) // Map to State for consistency
				}
				if alertType, ok := alertMap["AlertType"]; ok && alertType != nil {
					standardizedAlert["AlertType"] = alertType.(string)
				}
				if level, ok := alertMap["Level"]; ok && level != nil {
					standardizedAlert["Level"] = level.(string)
					standardizedAlert["Severity"] = level.(string) // Map to Severity for backward compatibility
				}
				if message, ok := alertMap["Message"]; ok && message != nil {
					standardizedAlert["Message"] = message.(string)
					standardizedAlert["Describe"] = message.(string) // Map to Describe for backward compatibility
				}
				if duration, ok := alertMap["Duration"]; ok && duration != nil {
					standardizedAlert["Duration"] = duration.(string)
				}
				if promQL, ok := alertMap["PromQL"]; ok && promQL != nil {
					standardizedAlert["PromQL"] = promQL.(string)
				}
				if clusterId, ok := alertMap["ClusterId"]; ok && clusterId != nil {
					standardizedAlert["ClusterId"] = clusterId.(string)
				}
				if metricsType, ok := alertMap["MetricsType"]; ok && metricsType != nil {
					standardizedAlert["MetricsType"] = metricsType.(string)
				}
				if notifyStrategy, ok := alertMap["NotifyStrategy"]; ok && notifyStrategy != nil {
					standardizedAlert["NotifyStrategy"] = notifyStrategy.(string)
				}
				if autoAddNewApp, ok := alertMap["AutoAddNewApplication"]; ok && autoAddNewApp != nil {
					if autoAddBool, ok := autoAddNewApp.(bool); ok {
						standardizedAlert["AutoAddNewApplication"] = autoAddBool
					}
				}
				if regionId, ok := alertMap["RegionId"]; ok && regionId != nil {
					standardizedAlert["RegionId"] = regionId.(string)
				}
				if userId, ok := alertMap["UserId"]; ok && userId != nil {
					standardizedAlert["UserId"] = userId.(string)
				}
				if createdTime, ok := alertMap["CreatedTime"]; ok && createdTime != nil {
					if createdTimeFloat, ok := createdTime.(float64); ok {
						standardizedAlert["CreatedTime"] = int64(createdTimeFloat)
						standardizedAlert["CreateTime"] = int64(createdTimeFloat) // Map to CreateTime for backward compatibility
					}
				}
				if updatedTime, ok := alertMap["UpdatedTime"]; ok && updatedTime != nil {
					if updatedTimeFloat, ok := updatedTime.(float64); ok {
						standardizedAlert["UpdatedTime"] = int64(updatedTimeFloat)
					}
				}
				if extend, ok := alertMap["Extend"]; ok && extend != nil {
					standardizedAlert["Extend"] = extend.(string)
				}
				if pids, ok := alertMap["Pids"]; ok && pids != nil {
					standardizedAlert["Pids"] = pids
				}
				if annotations, ok := alertMap["Annotations"]; ok && annotations != nil {
					standardizedAlert["Annotations"] = annotations
				}
				if labels, ok := alertMap["Labels"]; ok && labels != nil {
					standardizedAlert["Labels"] = labels
				}
				if tags, ok := alertMap["Tags"]; ok && tags != nil {
					standardizedAlert["Tags"] = tags
				}
				if filters, ok := alertMap["Filters"]; ok && filters != nil {
					standardizedAlert["Filters"] = filters
				}
				if alertRuleContent, ok := alertMap["AlertRuleContent"]; ok && alertRuleContent != nil {
					standardizedAlert["AlertRuleContent"] = alertRuleContent
				}

				// Legacy/backward compatibility fields - set defaults if not available
				if _, ok := standardizedAlert["Owner"]; !ok {
					if owner, ok := alertMap["Owner"]; ok && owner != nil {
						standardizedAlert["Owner"] = owner.(string)
					} else {
						standardizedAlert["Owner"] = ""
					}
				}
				if _, ok := standardizedAlert["Handler"]; !ok {
					if handler, ok := alertMap["Handler"]; ok && handler != nil {
						standardizedAlert["Handler"] = handler.(string)
					} else {
						standardizedAlert["Handler"] = ""
					}
				}
				if _, ok := standardizedAlert["Solution"]; !ok {
					if solution, ok := alertMap["Solution"]; ok && solution != nil {
						standardizedAlert["Solution"] = solution.(string)
					} else {
						standardizedAlert["Solution"] = ""
					}
				}
				if _, ok := standardizedAlert["DispatchRuleId"]; !ok {
					if dispatchRuleId, ok := alertMap["DispatchRuleId"]; ok && dispatchRuleId != nil {
						if dispatchRuleIdFloat, ok := dispatchRuleId.(float64); ok {
							standardizedAlert["DispatchRuleId"] = dispatchRuleIdFloat
						}
					} else {
						standardizedAlert["DispatchRuleId"] = float64(0)
					}
				}
				if _, ok := standardizedAlert["DispatchRuleName"]; !ok {
					if dispatchRuleName, ok := alertMap["DispatchRuleName"]; ok && dispatchRuleName != nil {
						standardizedAlert["DispatchRuleName"] = dispatchRuleName.(string)
					} else {
						standardizedAlert["DispatchRuleName"] = ""
					}
				}

				return standardizedAlert, nil
			}
		}
	}

	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}

// CreateArmsAlertRule creates a new ARMS alert rule
func (s *ArmsService) CreateArmsAlertRule(alertName, severity, description, integrationType, clusterId string, rule map[string]interface{}) (int64, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert parameters to AlertRule struct
		alertRule := &aliyunArmsAPI.AlertRule{
			AlertName: alertName,
			Level:     severity,
			Message:   description,
			AlertType: "PROMETHEUS_MONITORING_ALERT_RULE",
			ClusterId: clusterId,
		}

		// Set rule parameters from map
		if rule != nil {
			if promql, ok := rule["promql"].(string); ok && promql != "" {
				alertRule.PromQL = promql
			} else if expression, ok := rule["expression"].(string); ok && expression != "" {
				alertRule.PromQL = expression
			}

			if duration, ok := rule["duration"].(int); ok && duration > 0 {
				alertRule.Duration = fmt.Sprintf("%d", duration)
			} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
				alertRule.Duration = fmt.Sprintf("%.0f", durationFloat)
			}

			if message, ok := rule["message"].(string); ok && message != "" {
				alertRule.Message = message
			}

			if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
				alertRule.AlertCheckType = checkType
			} else {
				alertRule.AlertCheckType = "CUSTOM"
			}

			if alertGroup, ok := rule["alert_group"].(int); ok {
				alertRule.AlertGroup = int64(alertGroup)
			} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
				alertRule.AlertGroup = int64(alertGroupFloat)
			} else {
				alertRule.AlertGroup = -1
			}

			// Set labels if provided in rule
			if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
				for key, value := range labels {
					alertRule.Labels = append(alertRule.Labels, aliyunArmsAPI.AlertRuleLabel{
						Key:   key,
						Value: fmt.Sprintf("%v", value),
					})
				}
			}
		}

		// Set default values for required fields if not specified
		if alertRule.AlertCheckType == "" {
			alertRule.AlertCheckType = "CUSTOM"
		}
		if alertRule.AlertGroup == 0 {
			alertRule.AlertGroup = -1
		}

		createdRule, err := s.armsAPI.CreateAlertRule(alertRule)
		if err == nil && createdRule != nil {
			return createdRule.AlertId, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "CreateOrUpdateAlertRule"
	client := s.client

	request := map[string]interface{}{
		"AlertName": alertName,
		"Level":     severity,
		"AlertType": "PROMETHEUS_MONITORING_ALERT_RULE",
		"RegionId":  s.client.RegionId,
	}

	// Set description if provided
	if description != "" {
		request["Message"] = description
	}

	// Set cluster ID if provided
	if clusterId != "" {
		request["ClusterId"] = clusterId
	}

	// Set PromQL expression and other rule parameters if provided
	if rule != nil {
		if promql, ok := rule["promql"].(string); ok && promql != "" {
			request["PromQL"] = promql
		} else if expression, ok := rule["expression"].(string); ok && expression != "" {
			request["PromQL"] = expression
		}

		// Set duration if provided in rule
		if duration, ok := rule["duration"].(int); ok && duration > 0 {
			request["Duration"] = duration
		} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
			request["Duration"] = int(durationFloat)
		}

		// Set message if provided in rule
		if message, ok := rule["message"].(string); ok && message != "" {
			request["Message"] = message
		}

		// Set alert check type if provided
		if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
			request["AlertCheckType"] = checkType
		} else {
			request["AlertCheckType"] = "CUSTOM" // Default for custom PromQL
		}

		// Set alert group if provided
		if alertGroup, ok := rule["alert_group"].(int); ok {
			request["AlertGroup"] = alertGroup
		} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
			request["AlertGroup"] = int(alertGroupFloat)
		} else {
			request["AlertGroup"] = -1 // Default for custom PromQL
		}

		// Set labels if provided in rule
		if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			labelsMaps := make([]map[string]interface{}, 0)
			for key, value := range labels {
				labelsMaps = append(labelsMaps, map[string]interface{}{
					"name":  key,
					"value": fmt.Sprintf("%v", value),
				})
			}
			if labelString, err := convertArrayObjectToJsonString(labelsMaps); err == nil {
				request["Labels"] = labelString
			}
		}
	}

	// Set default values for required fields if not specified
	if _, ok := request["AlertCheckType"]; !ok {
		request["AlertCheckType"] = "CUSTOM"
	}
	if _, ok := request["AlertGroup"]; !ok {
		request["AlertGroup"] = -1
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return 0, WrapErrorf(err, DefaultErrorMsg, "CreateArmsAlertRule", action, AlibabaCloudSdkGoERROR)
	}

	if alertRule, ok := response["AlertRule"].(map[string]interface{}); ok {
		if alertId, ok := alertRule["AlertId"]; ok {
			if alertIdFloat, ok := alertId.(float64); ok {
				return int64(alertIdFloat), nil
			}
		}
	}

	return 0, WrapError(fmt.Errorf("AlertId not found in response"))
}

// UpdateArmsAlertRule updates an existing ARMS alert rule
func (s *ArmsService) UpdateArmsAlertRule(alertId int64, alertName, severity, description, integrationType, clusterId string, rule map[string]interface{}) error {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		_, err := s.armsAPI.UpdateAlertRule(&aliyunArmsAPI.AlertRule{
			AlertId:   alertId,
			AlertName: alertName,
			Level:     severity,
			Message:   description,
			AlertType: "PROMETHEUS_MONITORING_ALERT_RULE",
			ClusterId: clusterId,
		})
		if err == nil {
			return nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "CreateOrUpdateAlertRule"
	client := s.client

	request := map[string]interface{}{
		"AlertId":   alertId,
		"AlertName": alertName,
		"Level":     severity,
		"AlertType": "PROMETHEUS_MONITORING_ALERT_RULE",
		"RegionId":  s.client.RegionId,
	}

	// Set description if provided
	if description != "" {
		request["Message"] = description
	}

	// Set cluster ID if provided
	if clusterId != "" {
		request["ClusterId"] = clusterId
	}

	// Set PromQL expression and other rule parameters if provided
	if rule != nil {
		if promql, ok := rule["promql"].(string); ok && promql != "" {
			request["PromQL"] = promql
		} else if expression, ok := rule["expression"].(string); ok && expression != "" {
			request["PromQL"] = expression
		}

		// Set duration if provided in rule
		if duration, ok := rule["duration"].(int); ok && duration > 0 {
			request["Duration"] = duration
		} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
			request["Duration"] = int(durationFloat)
		}

		// Set message if provided in rule
		if message, ok := rule["message"].(string); ok && message != "" {
			request["Message"] = message
		}

		// Set alert check type if provided
		if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
			request["AlertCheckType"] = checkType
		} else {
			request["AlertCheckType"] = "CUSTOM" // Default for custom PromQL
		}

		// Set alert group if provided
		if alertGroup, ok := rule["alert_group"].(int); ok {
			request["AlertGroup"] = alertGroup
		} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
			request["AlertGroup"] = int(alertGroupFloat)
		} else {
			request["AlertGroup"] = -1 // Default for custom PromQL
		}

		// Set labels if provided in rule
		if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			labelsMaps := make([]map[string]interface{}, 0)
			for key, value := range labels {
				labelsMaps = append(labelsMaps, map[string]interface{}{
					"name":  key,
					"value": fmt.Sprintf("%v", value),
				})
			}
			if labelString, err := convertArrayObjectToJsonString(labelsMaps); err == nil {
				request["Labels"] = labelString
			}
		}
	}

	// Set default values for required fields if not specified
	if _, ok := request["AlertCheckType"]; !ok {
		request["AlertCheckType"] = "CUSTOM"
	}
	if _, ok := request["AlertGroup"]; !ok {
		request["AlertGroup"] = -1
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "UpdateArmsAlertRule", action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DescribeArmsAlertNotificationPolicy describes ARMS alert notification policy
func (s *ArmsService) DescribeArmsAlertNotificationPolicy(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		policy, err := s.armsAPI.GetAlertNotificationPolicy(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"Id":                 policy.Id,
				"Name":               policy.Name,
				"SendRecoverMessage": policy.SendRecoverMessage,
				"RepeatInterval":     policy.RepeatInterval,
				"EscalationPolicyId": policy.EscalationPolicyId,
				"GroupRule":          policy.GroupRule,
				"MatchingRules":      policy.MatchingRules,
				"NotifyRule":         policy.NotifyRule,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "ListNotificationPolicies"
	request := map[string]interface{}{
		"Page":     1,
		"Size":     PageSizeXLarge,
		"IsDetail": true,
		"RegionId": s.client.RegionId,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.PageBean.NotificationPolicies", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.NotificationPolicies", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}
	for _, v := range v.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["Id"]) == id {
			return v.(map[string]interface{}), nil
		}
	}
	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}
