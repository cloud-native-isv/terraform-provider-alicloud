package alicloud

import (
	"github.com/alibabacloud-go/tea/tea"
	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func expandSlsAlert(d *schema.ResourceData) *slsAPI.Alert {
	alertName := d.Get("alert_name").(string)
	alertDisplayName := d.Get("alert_displayname").(string)
	alertDescription := d.Get("alert_description").(string)

	alert := &slsAPI.Alert{
		Name:        tea.String(alertName),
		DisplayName: tea.String(alertDisplayName),
		Description: tea.String(alertDescription),
		Status:      tea.String("Enabled"),
	}

	if v, ok := d.GetOk("schedule"); ok {
		scheduleList := v.(*schema.Set).List()
		if len(scheduleList) > 0 {
			scheduleMap := scheduleList[0].(map[string]interface{})
			schedule := &slsAPI.Schedule{
				Type:           scheduleMap["type"].(string),
				Interval:       scheduleMap["interval"].(string),
				CronExpression: scheduleMap["cron_expression"].(string),
				DayOfWeek:      int32(scheduleMap["day_of_week"].(int)),
				Hour:           int32(scheduleMap["hour"].(int)),
				Delay:          int32(scheduleMap["delay"].(int)),
				RunImmediately: scheduleMap["run_immediately"].(bool),
				TimeZone:       scheduleMap["time_zone"].(string),
			}
			alert.Schedule = schedule
		}
	} else {
		scheduleType, okType := d.GetOk("schedule_type")
		scheduleInterval, okInterval := d.GetOk("schedule_interval")
		if okType && okInterval {
			alert.Schedule = &slsAPI.Schedule{
				Type:     scheduleType.(string),
				Interval: scheduleInterval.(string),
			}
		}
	}

	// Expand Configuration for V2
	if _, ok := d.GetOk("version"); ok {
		alert.Configuration = expandSlsAlertConfiguration(d)
	}

	return alert
}

func expandSlsAlertConfiguration(d *schema.ResourceData) *slsAPI.AlertConfig {
	config := &slsAPI.AlertConfig{}

	version := d.Get("version").(string)
	if version == "" {
		version = "2.0"
	}
	config.Version = tea.String(version)

	if v, ok := d.GetOk("type"); ok {
		config.Type = tea.String(v.(string))
	} else {
		config.Type = tea.String("default")
	}

	if v, ok := d.GetOk("threshold"); ok {
		config.Threshold = tea.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("dashboard"); ok {
		config.Dashboard = tea.String(v.(string))
	}

	if v, ok := d.GetOk("no_data_fire"); ok {
		config.NoDataFire = tea.Bool(v.(bool))
	}

	if v, ok := d.GetOk("no_data_severity"); ok {
		config.NoDataSeverity = tea.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("send_resolved"); ok {
		config.SendResolved = tea.Bool(v.(bool))
	}

	if v, ok := d.GetOk("mute_until"); ok {
		config.MuteUntil = tea.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("auto_annotation"); ok {
		config.AutoAnnotation = tea.Bool(v.(bool))
	}

	// Labels
	if v, ok := d.GetOk("labels"); ok {
		var labels []*slsAPI.AlertTag
		for _, e := range v.([]interface{}) {
			m := e.(map[string]interface{})
			labels = append(labels, &slsAPI.AlertTag{
				Key:   tea.String(m["key"].(string)),
				Value: tea.String(m["value"].(string)),
			})
		}
		config.Labels = labels
	}

	// Annotations
	if v, ok := d.GetOk("annotations"); ok {
		var annotations []*slsAPI.AlertTag
		for _, e := range v.([]interface{}) {
			m := e.(map[string]interface{})
			annotations = append(annotations, &slsAPI.AlertTag{
				Key:   tea.String(m["key"].(string)),
				Value: tea.String(m["value"].(string)),
			})
		}
		config.Annotations = annotations
	}

	// SeverityConfigurations
	if v, ok := d.GetOk("severity_configurations"); ok {
		var severeConfigs []*slsAPI.SeverityConfiguration
		for _, e := range v.([]interface{}) {
			m := e.(map[string]interface{})
			severity := m["severity"].(int)

			evalMap := m["eval_condition"].(map[string]interface{})
			cond := evalMap["condition"].(string)
			countCond := evalMap["count_condition"].(string)

			severeConfigs = append(severeConfigs, &slsAPI.SeverityConfiguration{
				Severity: slsAPI.Severity(severity),
				EvalCondition: slsAPI.ConditionConfiguration{
					Condition:      cond,
					CountCondition: countCond,
				},
			})
		}
		config.SeverityConfigurations = severeConfigs
	}

	// JoinConfigurations
	if v, ok := d.GetOk("join_configurations"); ok {
		var joinConfigs []*slsAPI.JoinConfiguration
		for _, e := range v.([]interface{}) {
			m := e.(map[string]interface{})
			joinConfigs = append(joinConfigs, &slsAPI.JoinConfiguration{
				Type:      m["type"].(string),
				Condition: m["condition"].(string),
			})
		}
		config.JoinConfigurations = joinConfigs
	}

	// GroupConfiguration
	if v, ok := d.GetOk("group_configuration"); ok {
		list := v.(*schema.Set).List()
		if len(list) > 0 {
			m := list[0].(map[string]interface{})

			var fields []string
			if fv, ok := m["fields"]; ok {
				for _, f := range fv.(*schema.Set).List() {
					fields = append(fields, f.(string))
				}
			}

			config.GroupConfiguration = &slsAPI.GroupConfiguration{
				Type:   m["type"].(string),
				Fields: fields,
			}
		}
	}

	// PolicyConfiguration
	if v, ok := d.GetOk("policy_configuration"); ok {
		list := v.(*schema.Set).List()
		if len(list) > 0 {
			m := list[0].(map[string]interface{})
			actPolId := ""
			if val, ok := m["action_policy_id"]; ok {
				actPolId = val.(string)
			}

			config.PolicyConfiguration = &slsAPI.PolicyConfiguration{
				AlertPolicyId:  m["alert_policy_id"].(string),
				ActionPolicyId: actPolId,
				RepeatInterval: m["repeat_interval"].(string),
			}
		}
	}

	// TemplateConfiguration
	if v, ok := d.GetOk("template_configuration"); ok {
		list := v.([]interface{})
		if len(list) > 0 {
			m := list[0].(map[string]interface{})

			tmpl := &slsAPI.TemplateConfiguration{
				Id:      m["id"].(string),
				Type:    m["type"].(string),
				Version: "1",
			}
			if val, ok := m["lang"].(string); ok && val != "" {
				tmpl.Lang = val
			}

			if tokensMap, ok := m["tokens"].(map[string]interface{}); ok {
				tokens := make(map[string]string)
				for k, val := range tokensMap {
					tokens[k] = val.(string)
				}
				tmpl.Tokens = tokens
			}

			if annotMap, ok := m["annotations"].(map[string]interface{}); ok {
				annots := make(map[string]string)
				for k, val := range annotMap {
					annots[k] = val.(string)
				}
				tmpl.Annotations = annots
			}
			config.TemplateConfiguration = tmpl
		}
	}

	// QueryList
	if v, ok := d.GetOk("query_list"); ok {
		var queryList []*slsAPI.AlertQuery
		for _, e := range v.([]interface{}) {
			m := e.(map[string]interface{})

			q := &slsAPI.AlertQuery{
				Query:        m["query"].(string),
				Start:        m["start"].(string),
				End:          m["end"].(string),
				Store:        m["store"].(string),
				StoreType:    m["store_type"].(string),
				TimeSpanType: m["time_span_type"].(string),
			}

			if val, ok := m["chart_title"].(string); ok {
				q.ChartTitle = val
			}
			if val, ok := m["project"].(string); ok {
				q.Project = val
			}
			if val, ok := m["region"].(string); ok {
				q.Region = val
			}
			if val, ok := m["role_arn"].(string); ok {
				q.RoleArn = val
			}
			if val, ok := m["dashboard_id"].(string); ok {
				q.DashboardId = val
			}
			if val, ok := m["power_sql_mode"].(string); ok {
				q.PowerSqlMode = slsAPI.PowerSqlMode(val)
			}

			queryList = append(queryList, q)
		}
		config.QueryList = queryList
	}

	return config
}
