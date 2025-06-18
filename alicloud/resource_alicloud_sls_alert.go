package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudSlsAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudSlsAlertCreate,
		Read:   resourceAlicloudSlsAlertRead,
		Update: resourceAlicloudSlsAlertUpdate,
		Delete: resourceAlicloudSlsAlertDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
				Default:  "",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ENABLED",
				ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
			},

			// Configuration block
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "2.0",
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "default",
						},
						"auto_annotation": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"dashboard": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mute_until": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"no_data_fire": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"no_data_severity": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      6,
							ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
						},
						"send_resolved": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},

						// Condition Configuration
						"condition_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"condition": {
										Type:     schema.TypeString,
										Required: true,
									},
									"count_condition": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},

						// Query List
						"query_list": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"chart_title": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
									"project": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"store": {
										Type:     schema.TypeString,
										Required: true,
									},
									"store_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "log",
										ValidateFunc: validation.StringInSlice([]string{"log", "metric"}, false),
									},
									"query": {
										Type:     schema.TypeString,
										Required: true,
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
										Default:  "Relative",
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
										Default:      "auto",
										ValidateFunc: validation.StringInSlice([]string{"auto", "enable", "disable"}, false),
									},
								},
							},
						},

						// Severity Configurations
						"severity_configurations": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"severity": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
									},
									"eval_condition": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"condition": {
													Type:     schema.TypeString,
													Required: true,
												},
												"count_condition": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},

						// Labels
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

						// Annotations
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

						// Join Configurations
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

						// Group Configuration
						"group_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "no_group",
									},
									"fields": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},

						// Policy Configuration
						"policy_configuration": {
							Type:     schema.TypeList,
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

						// Template Configuration
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
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"sys", "user"}, false),
									},
									"lang": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "cn",
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "1.0",
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

						// Sink Configurations
						"sink_alerthub": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},

						"sink_cms": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},

						"sink_event_store": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"endpoint": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"project": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"event_store": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},

						// Tags
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			// Schedule block
			"schedule": {
				Type:     schema.TypeList,
				Required: true,
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
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 6),
						},
						"hour": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"delay": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"run_immediately": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "+0800",
						},
					},
				},
			},
		},
	}
}

func resourceAlicloudSlsAlertCreate(d *schema.ResourceData, meta interface{}) error {
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

	return resourceAlicloudSlsAlertRead(d, meta)
}

func resourceAlicloudSlsAlertRead(d *schema.ResourceData, meta interface{}) error {
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

func resourceAlicloudSlsAlertUpdate(d *schema.ResourceData, meta interface{}) error {
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

	return resourceAlicloudSlsAlertRead(d, meta)
}

func resourceAlicloudSlsAlertDelete(d *schema.ResourceData, meta interface{}) error {
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
	status := d.Get("status").(string)

	alert := &aliyunSlsAPI.Alert{
		Name:        &alertName,
		DisplayName: &displayName,
		Description: &description,
		Status:      &status,
	}

	// Build configuration from configuration block
	if configList, ok := d.GetOk("configuration"); ok {
		configListData := configList.([]interface{})
		if len(configListData) > 0 {
			configData := configListData[0].(map[string]interface{})

			config := &aliyunSlsAPI.AlertConfig{}

			// Set basic configuration fields
			if version, ok := configData["version"].(string); ok && version != "" {
				config.Version = &version
			} else {
				config.Version = tea.String("2.0")
			}

			if configType, ok := configData["type"].(string); ok && configType != "" {
				config.Type = &configType
			} else {
				config.Type = tea.String("default")
			}

			if autoAnnotation, ok := configData["auto_annotation"].(bool); ok {
				config.AutoAnnotation = &autoAnnotation
			}

			if dashboard, ok := configData["dashboard"].(string); ok && dashboard != "" {
				config.Dashboard = &dashboard
			}

			if muteUntil, ok := configData["mute_until"].(int); ok {
				muteUntilInt64 := int64(muteUntil)
				config.MuteUntil = &muteUntilInt64
			}

			if noDataFire, ok := configData["no_data_fire"].(bool); ok {
				config.NoDataFire = &noDataFire
			}

			if noDataSeverity, ok := configData["no_data_severity"].(int); ok {
				noDataSeverityInt32 := int32(noDataSeverity)
				config.NoDataSeverity = &noDataSeverityInt32
			}

			if sendResolved, ok := configData["send_resolved"].(bool); ok {
				config.SendResolved = &sendResolved
			}

			if threshold, ok := configData["threshold"].(int); ok {
				thresholdInt32 := int32(threshold)
				config.Threshold = &thresholdInt32
			}

			// Build condition configuration
			if conditionConfigList, ok := configData["condition_configuration"].([]interface{}); ok && len(conditionConfigList) > 0 {
				conditionConfigData := conditionConfigList[0].(map[string]interface{})
				config.ConditionConfiguration = &aliyunSlsAPI.ConditionConfiguration{
					Condition:      conditionConfigData["condition"].(string),
					CountCondition: conditionConfigData["count_condition"].(string),
				}
			} else {
				// Provide default condition configuration if not set
				defaultCondition := "__count__ > 0"
				config.ConditionConfiguration = &aliyunSlsAPI.ConditionConfiguration{
					Condition:      defaultCondition,
					CountCondition: defaultCondition,
				}
			}

			// Build query list
			if queryList, ok := configData["query_list"].([]interface{}); ok {
				config.QueryList = make([]*aliyunSlsAPI.AlertQuery, 0, len(queryList))

				for _, queryRaw := range queryList {
					queryMap := queryRaw.(map[string]interface{})
					query := &aliyunSlsAPI.AlertQuery{
						Query: queryMap["query"].(string),
						Start: queryMap["start"].(string),
						End:   queryMap["end"].(string),
					}

					if chartTitle, ok := queryMap["chart_title"].(string); ok && chartTitle != "" {
						query.ChartTitle = chartTitle
					}
					if project, ok := queryMap["project"].(string); ok && project != "" {
						query.Project = project
					}
					if store, ok := queryMap["store"].(string); ok && store != "" {
						query.Store = store
					}
					if storeType, ok := queryMap["store_type"].(string); ok && storeType != "" {
						query.StoreType = storeType
					}
					if timeSpanType, ok := queryMap["time_span_type"].(string); ok && timeSpanType != "" {
						query.TimeSpanType = timeSpanType
					}
					if region, ok := queryMap["region"].(string); ok && region != "" {
						query.Region = region
					}
					if roleArn, ok := queryMap["role_arn"].(string); ok && roleArn != "" {
						query.RoleArn = roleArn
					}
					if dashboardId, ok := queryMap["dashboard_id"].(string); ok && dashboardId != "" {
						query.DashboardId = dashboardId
					}
					if powerSqlMode, ok := queryMap["power_sql_mode"].(string); ok && powerSqlMode != "" {
						query.PowerSqlMode = aliyunSlsAPI.PowerSqlMode(powerSqlMode)
					}

					config.QueryList = append(config.QueryList, query)
				}
			}

			// Build severity configurations
			if severityConfigsList, ok := configData["severity_configurations"].([]interface{}); ok {
				config.SeverityConfigurations = make([]*aliyunSlsAPI.SeverityConfiguration, 0, len(severityConfigsList))

				for _, severityConfigRaw := range severityConfigsList {
					severityConfigMap := severityConfigRaw.(map[string]interface{})
					severityConfig := &aliyunSlsAPI.SeverityConfiguration{
						Severity: aliyunSlsAPI.Severity(severityConfigMap["severity"].(int)),
					}

					if evalConditionList, ok := severityConfigMap["eval_condition"].([]interface{}); ok && len(evalConditionList) > 0 {
						evalConditionData := evalConditionList[0].(map[string]interface{})
						severityConfig.EvalCondition = aliyunSlsAPI.ConditionConfiguration{
							Condition:      evalConditionData["condition"].(string),
							CountCondition: evalConditionData["count_condition"].(string),
						}
					} else {
						// Provide default condition if not set
						defaultCondition := "__count__ > 0"
						severityConfig.EvalCondition = aliyunSlsAPI.ConditionConfiguration{
							Condition:      defaultCondition,
							CountCondition: defaultCondition,
						}
					}

					config.SeverityConfigurations = append(config.SeverityConfigurations, severityConfig)
				}
			}

			// Build labels
			if labelsList, ok := configData["labels"].([]interface{}); ok {
				config.Labels = make([]*aliyunSlsAPI.AlertTag, 0, len(labelsList))

				for _, labelRaw := range labelsList {
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
			if annotationsList, ok := configData["annotations"].([]interface{}); ok {
				config.Annotations = make([]*aliyunSlsAPI.AlertTag, 0, len(annotationsList))

				for _, annotationRaw := range annotationsList {
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
			if joinConfigsList, ok := configData["join_configurations"].([]interface{}); ok {
				config.JoinConfigurations = make([]*aliyunSlsAPI.JoinConfiguration, 0, len(joinConfigsList))

				for _, joinConfigRaw := range joinConfigsList {
					joinConfigMap := joinConfigRaw.(map[string]interface{})
					config.JoinConfigurations = append(config.JoinConfigurations, &aliyunSlsAPI.JoinConfiguration{
						Type:      joinConfigMap["type"].(string),
						Condition: joinConfigMap["condition"].(string),
					})
				}
			}

			// Build group configuration
			if groupConfigList, ok := configData["group_configuration"].([]interface{}); ok && len(groupConfigList) > 0 {
				groupConfigMap := groupConfigList[0].(map[string]interface{})
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

			// Build policy configuration
			if policyConfigList, ok := configData["policy_configuration"].([]interface{}); ok && len(policyConfigList) > 0 {
				policyConfigMap := policyConfigList[0].(map[string]interface{})
				config.PolicyConfiguration = &aliyunSlsAPI.PolicyConfiguration{
					AlertPolicyId:  policyConfigMap["alert_policy_id"].(string),
					RepeatInterval: policyConfigMap["repeat_interval"].(string),
				}

				if actionPolicyId, ok := policyConfigMap["action_policy_id"].(string); ok && actionPolicyId != "" {
					config.PolicyConfiguration.ActionPolicyId = actionPolicyId
				}
			}

			// Build template configuration
			if templateConfigList, ok := configData["template_configuration"].([]interface{}); ok && len(templateConfigList) > 0 {
				templateConfigMap := templateConfigList[0].(map[string]interface{})
				config.TemplateConfiguration = &aliyunSlsAPI.TemplateConfiguration{
					Id:   templateConfigMap["id"].(string),
					Type: templateConfigMap["type"].(string),
				}

				if lang, ok := templateConfigMap["lang"].(string); ok && lang != "" {
					config.TemplateConfiguration.Lang = lang
				}

				if version, ok := templateConfigMap["version"].(string); ok && version != "" {
					config.TemplateConfiguration.Version = version
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

			// Build sink configurations
			if sinkAlerthubList, ok := configData["sink_alerthub"].([]interface{}); ok && len(sinkAlerthubList) > 0 {
				sinkAlerthubMap := sinkAlerthubList[0].(map[string]interface{})
				enabled := sinkAlerthubMap["enabled"].(bool)
				config.SinkAlerthub = &aliyunSlsAPI.SinkAlerthubConfiguration{
					Enabled: &enabled,
				}
			}

			if sinkCmsList, ok := configData["sink_cms"].([]interface{}); ok && len(sinkCmsList) > 0 {
				sinkCmsMap := sinkCmsList[0].(map[string]interface{})
				enabled := sinkCmsMap["enabled"].(bool)
				config.SinkCms = &aliyunSlsAPI.SinkCmsConfiguration{
					Enabled: &enabled,
				}
			}

			if sinkEventStoreList, ok := configData["sink_event_store"].([]interface{}); ok && len(sinkEventStoreList) > 0 {
				sinkEventStoreMap := sinkEventStoreList[0].(map[string]interface{})
				sinkEventStore := &aliyunSlsAPI.SinkEventStoreConfiguration{
					Enabled: tea.Bool(sinkEventStoreMap["enabled"].(bool)),
				}

				if endpoint, ok := sinkEventStoreMap["endpoint"].(string); ok && endpoint != "" {
					sinkEventStore.Endpoint = &endpoint
				}
				if project, ok := sinkEventStoreMap["project"].(string); ok && project != "" {
					sinkEventStore.Project = &project
				}
				if eventStore, ok := sinkEventStoreMap["event_store"].(string); ok && eventStore != "" {
					sinkEventStore.EventStore = &eventStore
				}
				if roleArn, ok := sinkEventStoreMap["role_arn"].(string); ok && roleArn != "" {
					sinkEventStore.RoleArn = &roleArn
				}

				config.SinkEventStore = sinkEventStore
			}

			// Build tags
			if tags, ok := configData["tags"].(*schema.Set); ok {
				tagsList := make([]*string, 0, tags.Len())
				for _, tag := range tags.List() {
					tagStr := tag.(string)
					tagsList = append(tagsList, &tagStr)
				}
				config.Tags = tagsList
			}

			alert.Configuration = config
		}
	}

	// Build schedule from schedule block
	if scheduleList, ok := d.GetOk("schedule"); ok {
		scheduleListData := scheduleList.([]interface{})
		if len(scheduleListData) > 0 {
			scheduleData := scheduleListData[0].(map[string]interface{})
			schedule := &aliyunSlsAPI.Schedule{
				Type: scheduleData["type"].(string),
			}

			if interval, ok := scheduleData["interval"].(string); ok && interval != "" {
				schedule.Interval = interval
			}
			if cronExpression, ok := scheduleData["cron_expression"].(string); ok && cronExpression != "" {
				schedule.CronExpression = cronExpression
			}
			if dayOfWeek, ok := scheduleData["day_of_week"].(int); ok {
				schedule.DayOfWeek = int32(dayOfWeek)
			}
			if hour, ok := scheduleData["hour"].(int); ok {
				schedule.Hour = int32(hour)
			}
			if delay, ok := scheduleData["delay"].(int); ok {
				schedule.Delay = int32(delay)
			}
			if timeZone, ok := scheduleData["time_zone"].(string); ok && timeZone != "" {
				schedule.TimeZone = timeZone
			}
			if runImmediately, ok := scheduleData["run_immediately"].(bool); ok {
				schedule.RunImmediately = runImmediately
			}

			alert.Schedule = schedule
		}
	}

	return alert, nil
}
