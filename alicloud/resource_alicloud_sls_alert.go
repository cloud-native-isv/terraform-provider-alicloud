package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudLogAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogAlertCreate,
		Read:   resourceAlicloudLogAlertRead,
		Update: resourceAlicloudLogAlertUpdate,
		Delete: resourceAlicloudLogAlertDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alert_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alert_displayname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alert_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"condition": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use eval_condition in severity_configurations",
			},
			"dashboard": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use dashboardId in query_list",
			},
			"mute_until": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"throttling": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use repeat_interval in policy_configuration",
			},
			"notify_threshold": {
				Type:       schema.TypeInt,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use threshold",
			},
			"threshold": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"no_data_fire": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"no_data_severity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
			},
			"send_resolved": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_annotation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"query_list": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"chart_title": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"store": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"store_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"logstore": {
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Deprecated from 1.161.0+, use store",
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
						"time_span_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Custom",
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dashboard_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"power_sql_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "enable", "disable"}, false),
						},
					},
				},
			},

			"notification_list": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use policy_configuration for notification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"sms",
								"dingtalk",
								"email",
								"messageCenter"},
								false),
						},
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mobile_list": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"email_list": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"annotations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"severity_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"severity": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
						},
						"eval_condition": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"join_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"cross_join", "inner_join", "left_join", "right_join", "full_join", "left_exclude", "right_exclude", "concat", "no_join"}, false),
						},
						"condition": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"policy_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert_policy_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"action_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"repeat_interval": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"group_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"fields": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"template_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"lang": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"sys", "user"}, false),
						},
						"tokens": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"schedule_interval": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Field 'schedule_interval' has been deprecated from provider version 1.176.0. New field 'schedule' instead.",
				ConflictsWith: []string{"schedule"},
			},
			"schedule_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Field 'schedule_type' has been deprecated from provider version 1.176.0. New field 'schedule' instead.",
				ConflictsWith: []string{"schedule"},
			},
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"FixedRate", "Cron", "Hourly", "Daily", "Weekly"}, false),
						},
						"interval": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cron_expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"day_of_week": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"delay": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"run_immediately": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{"schedule_type", "schedule_interval"},
			},
		},
	}
}

func resourceAlicloudLogAlertCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Get("project_name").(string)
	alertName := d.Get("alert_name").(string)

	// Build alert object from schema
	alert, err := buildSlsAlertFromSchema(d)
	if err != nil {
		return WrapError(fmt.Errorf("failed to build alert from schema: %w", err))
	}

	// Create the alert
	err = slsService.CreateSlsAlert(projectName, alert)
	if err != nil {
		return WrapError(err)
	}

	// Set resource ID as project_name:alert_name
	d.SetId(fmt.Sprintf("%s:%s", projectName, alertName))

	// Wait for alert to be created successfully
	stateConf := BuildStateConf([]string{}, []string{"ENABLED", "DISABLED"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsAlertStateRefreshFunc(projectName, alertName))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudLogAlertRead(d, meta)
}

func resourceAlicloudLogAlertRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid SLS alert ID format: %s, expected format: project_name:alert_name", d.Id()))
	}

	projectName := parts[0]
	alertName := parts[1]

	// Get alert details
	alertMap, err := slsService.DescribeSlsAlert(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic attributes
	d.Set("project_name", projectName)
	d.Set("alert_name", alertName)
	d.Set("alert_displayname", alertMap["displayName"])
	d.Set("alert_description", alertMap["description"])
	d.Set("status", alertMap["status"])

	// Set configuration if present
	if configuration, ok := alertMap["configuration"]; ok && configuration != nil {
		configMap := configuration.(map[string]interface{})
		configurationMaps := []map[string]interface{}{configMap}
		d.Set("configuration", configurationMaps)
	}

	// Set schedule if present
	if schedule, ok := alertMap["schedule"]; ok && schedule != nil {
		scheduleMap := schedule.(map[string]interface{})
		scheduleMaps := []map[string]interface{}{scheduleMap}
		d.Set("schedule", scheduleMaps)
	}

	return nil
}

func resourceAlicloudLogAlertUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid SLS alert ID format: %s, expected format: project_name:alert_name", d.Id()))
	}

	projectName := parts[0]
	alertName := parts[1]

	// Build updated alert object from schema
	alert, err := buildSlsAlertFromSchema(d)
	if err != nil {
		return WrapError(fmt.Errorf("failed to build alert from schema: %w", err))
	}

	// Update the alert
	err = slsService.UpdateSlsAlert(projectName, alertName, alert)
	if err != nil {
		return WrapError(err)
	}

	// Wait for alert to be updated successfully
	stateConf := BuildStateConf([]string{}, []string{"ENABLED", "DISABLED"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.SlsAlertStateRefreshFunc(projectName, alertName))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudLogAlertRead(d, meta)
}

func resourceAlicloudLogAlertDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid SLS alert ID format: %s, expected format: project_name:alert_name", d.Id()))
	}

	projectName := parts[0]
	alertName := parts[1]

	// Delete the alert
	err = slsService.DeleteSlsAlert(projectName, alertName)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// buildSlsAlertFromSchema builds an Alert object from Terraform schema data
func buildSlsAlertFromSchema(d *schema.ResourceData) (*aliyunSlsAPI.Alert, error) {
	// Use pointer types to match the new Alert struct definition
	alertName := d.Get("alert_name").(string)
	displayName := d.Get("alert_displayname").(string)
	description := d.Get("alert_description").(string)
	status := "ENABLED" // Default status

	alert := &aliyunSlsAPI.Alert{
		Name:        &alertName,
		DisplayName: &displayName,
		Description: &description,
		Status:      &status,
	}

	// Build configuration
	config := &aliyunSlsAPI.AlertConfig{}

	// This was missing in the original function and caused the SLS API error
	if v, ok := d.GetOk("condition"); ok && v.(string) != "" {
		// Handle legacy condition field (deprecated but still used)
		condition := v.(string)
		config.ConditionConfiguration = &aliyunSlsAPI.ConditionConfiguration{
			Condition:      condition,
			CountCondition: condition, // Use same condition for both fields
		}
	} else {
		// If no explicit condition is set, we need to build one from severity_configurations
		// This ensures ConditionConfiguration is always set
		defaultCondition := "__count__ > 0"
		config.ConditionConfiguration = &aliyunSlsAPI.ConditionConfiguration{
			Condition:      defaultCondition,
			CountCondition: defaultCondition,
		}
	}

	// Set basic configuration fields
	if v, ok := d.GetOk("auto_annotation"); ok {
		autoAnnotation := v.(bool)
		config.AutoAnnotation = &autoAnnotation
	}

	if v, ok := d.GetOk("dashboard"); ok {
		dashboard := v.(string)
		config.Dashboard = &dashboard
	}

	if v, ok := d.GetOk("mute_until"); ok {
		muteUntil := int64(v.(int))
		config.MuteUntil = &muteUntil
	}

	if v, ok := d.GetOk("no_data_fire"); ok {
		noDataFire := v.(bool)
		config.NoDataFire = &noDataFire
	}

	if v, ok := d.GetOk("no_data_severity"); ok {
		noDataSeverity := int32(v.(int))
		config.NoDataSeverity = &noDataSeverity
	}

	if v, ok := d.GetOk("send_resolved"); ok {
		sendResolved := v.(bool)
		config.SendResolved = &sendResolved
	}

	// Build query list
	if v, ok := d.GetOk("query_list"); ok {
		queryListRaw := v.([]interface{})
		config.QueryList = make([]*aliyunSlsAPI.AlertQuery, 0, len(queryListRaw))

		for _, queryRaw := range queryListRaw {
			queryMap := queryRaw.(map[string]interface{})
			query := &aliyunSlsAPI.AlertQuery{
				Query: queryMap["query"].(string),
				Start: queryMap["start"].(string),
				End:   queryMap["end"].(string),
			}

			if chartTitle, ok := queryMap["chart_title"].(string); ok {
				query.ChartTitle = chartTitle
			}
			if project, ok := queryMap["project"].(string); ok {
				query.Project = project
			}
			if store, ok := queryMap["store"].(string); ok {
				query.Store = store
			}
			if storeType, ok := queryMap["store_type"].(string); ok {
				query.StoreType = storeType
			}
			if timeSpanType, ok := queryMap["time_span_type"].(string); ok {
				query.TimeSpanType = timeSpanType
			}
			if region, ok := queryMap["region"].(string); ok {
				query.Region = region
			}
			if roleArn, ok := queryMap["role_arn"].(string); ok {
				query.RoleArn = roleArn
			}
			if dashboardId, ok := queryMap["dashboard_id"].(string); ok {
				query.DashboardId = dashboardId
			}

			config.QueryList = append(config.QueryList, query)
		}
	}

	// Build severity configurations
	if v, ok := d.GetOk("severity_configurations"); ok {
		severityConfigsRaw := v.([]interface{})
		config.SeverityConfigurations = make([]*aliyunSlsAPI.SeverityConfiguration, 0, len(severityConfigsRaw))

		for _, severityConfigRaw := range severityConfigsRaw {
			severityConfigMap := severityConfigRaw.(map[string]interface{})
			severityConfig := &aliyunSlsAPI.SeverityConfiguration{
				Severity: aliyunSlsAPI.Severity(severityConfigMap["severity"].(int)),
			}

			if evalCondition, ok := severityConfigMap["eval_condition"].(map[string]interface{}); ok {
				// Handle eval_condition with backward compatibility
				condition := ""
				countCondition := ""

				if conditionVal, exists := evalCondition["condition"]; exists {
					if condStr, isString := conditionVal.(string); isString && condStr != "" {
						condition = condStr
					} else {
						// Provide default condition if empty or invalid
						condition = "__count__ > 0"
					}
				} else {
					condition = "__count__ > 0"
				}

				if countConditionVal, exists := evalCondition["count_condition"]; exists {
					if countCondStr, isString := countConditionVal.(string); isString && countCondStr != "" {
						countCondition = countCondStr
					} else {
						// Provide default count condition if empty or invalid
						countCondition = "__count__ > 0"
					}
				} else {
					countCondition = "__count__ > 0"
				}

				severityConfig.EvalCondition = aliyunSlsAPI.ConditionConfiguration{
					Condition:      condition,
					CountCondition: countCondition,
				}
			}

			config.SeverityConfigurations = append(config.SeverityConfigurations, severityConfig)
		}
	}

	// Build labels
	if v, ok := d.GetOk("labels"); ok {
		labelsRaw := v.([]interface{})
		config.Labels = make([]*aliyunSlsAPI.AlertTag, 0, len(labelsRaw))

		for _, labelRaw := range labelsRaw {
			labelMap := labelRaw.(map[string]interface{})
			key := labelMap["key"].(string)
			value := labelMap["value"].(string)
			config.Labels = append(config.Labels, &aliyunSlsAPI.AlertTag{
				Key:   &key,
				Value: &value,
			})
		}
	}

	// Build annotations
	if v, ok := d.GetOk("annotations"); ok {
		annotationsRaw := v.([]interface{})
		config.Annotations = make([]*aliyunSlsAPI.AlertTag, 0, len(annotationsRaw))

		for _, annotationRaw := range annotationsRaw {
			annotationMap := annotationRaw.(map[string]interface{})
			key := annotationMap["key"].(string)
			value := annotationMap["value"].(string)
			config.Annotations = append(config.Annotations, &aliyunSlsAPI.AlertTag{
				Key:   &key,
				Value: &value,
			})
		}
	}

	// Build join configurations
	if v, ok := d.GetOk("join_configurations"); ok {
		joinConfigsRaw := v.([]interface{})
		config.JoinConfigurations = make([]*aliyunSlsAPI.JoinConfiguration, 0, len(joinConfigsRaw))

		for _, joinConfigRaw := range joinConfigsRaw {
			joinConfigMap := joinConfigRaw.(map[string]interface{})
			config.JoinConfigurations = append(config.JoinConfigurations, &aliyunSlsAPI.JoinConfiguration{
				Type:      joinConfigMap["type"].(string),
				Condition: joinConfigMap["condition"].(string),
			})
		}
	}

	// Build group configuration
	if v, ok := d.GetOk("group_configuration"); ok {
		groupConfigSet := v.(*schema.Set)
		if groupConfigSet.Len() > 0 {
			groupConfigMap := groupConfigSet.List()[0].(map[string]interface{})
			groupConfig := &aliyunSlsAPI.GroupConfiguration{
				Type: groupConfigMap["type"].(string),
			}

			if fields, ok := groupConfigMap["fields"].(*schema.Set); ok {
				fieldsList := make([]string, 0, fields.Len())
				for _, field := range fields.List() {
					fieldsList = append(fieldsList, field.(string))
				}
				groupConfig.Fields = fieldsList
			}

			config.GroupConfiguration = groupConfig
		}
	}

	// Build policy configuration
	if v, ok := d.GetOk("policy_configuration"); ok {
		policyConfigSet := v.(*schema.Set)
		if policyConfigSet.Len() > 0 {
			policyConfigMap := policyConfigSet.List()[0].(map[string]interface{})
			config.PolicyConfiguration = &aliyunSlsAPI.PolicyConfiguration{
				AlertPolicyId:  policyConfigMap["alert_policy_id"].(string),
				RepeatInterval: policyConfigMap["repeat_interval"].(string),
			}

			if actionPolicyId, ok := policyConfigMap["action_policy_id"].(string); ok {
				config.PolicyConfiguration.ActionPolicyId = actionPolicyId
			}
		}
	}

	// Build template configuration
	if v, ok := d.GetOk("template_configuration"); ok {
		templateConfigList := v.([]interface{})
		if len(templateConfigList) > 0 {
			templateConfigMap := templateConfigList[0].(map[string]interface{})
			config.TemplateConfiguration = &aliyunSlsAPI.TemplateConfiguration{
				Id:   templateConfigMap["id"].(string),
				Type: templateConfigMap["type"].(string),
			}

			if lang, ok := templateConfigMap["lang"].(string); ok {
				config.TemplateConfiguration.Lang = lang
			}

			if tokens, ok := templateConfigMap["tokens"].(map[string]interface{}); ok {
				tokenMap := make(map[string]string)
				for k, v := range tokens {
					tokenMap[k] = v.(string)
				}
				config.TemplateConfiguration.Tokens = tokenMap
			}

			if annotations, ok := templateConfigMap["annotations"].(map[string]interface{}); ok {
				annotationMap := make(map[string]string)
				for k, v := range annotations {
					annotationMap[k] = v.(string)
				}
				config.TemplateConfiguration.Annotations = annotationMap
			}
		}
	}

	alert.Configuration = config

	// Build schedule
	if v, ok := d.GetOk("schedule"); ok {
		scheduleSet := v.(*schema.Set)
		if scheduleSet.Len() > 0 {
			scheduleMap := scheduleSet.List()[0].(map[string]interface{})
			schedule := &aliyunSlsAPI.Schedule{
				Type: scheduleMap["type"].(string),
			}

			if interval, ok := scheduleMap["interval"].(string); ok {
				schedule.Interval = interval
			}
			if cronExpression, ok := scheduleMap["cron_expression"].(string); ok {
				schedule.CronExpression = cronExpression
			}
			if dayOfWeek, ok := scheduleMap["day_of_week"].(int); ok {
				schedule.DayOfWeek = int32(dayOfWeek)
			}
			if hour, ok := scheduleMap["hour"].(int); ok {
				schedule.Hour = int32(hour)
			}
			if delay, ok := scheduleMap["delay"].(int); ok {
				schedule.Delay = int32(delay)
			}
			if timeZone, ok := scheduleMap["time_zone"].(string); ok {
				schedule.TimeZone = timeZone
			}
			if runImmediately, ok := scheduleMap["run_immediately"].(bool); ok {
				schedule.RunImmediately = runImmediately
			}

			alert.Schedule = schedule
		}
	}

	return alert, nil
}
