package alicloud

import (
	"regexp"
	"strconv"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	armsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsAlertRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertRulesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of alert rule IDs to filter results",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
				Description:  "A regex string to filter results by alert rule name",
			},
			"names": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of alert rule names",
			},
			"alert_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the alert rule to filter results",
			},
			"alert_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Alert type to filter results (e.g., PROMETHEUS_MONITORING_ALERT_RULE)",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Cluster ID to filter Prometheus alert rules",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Alert rule status to filter results (RUNNING, STOPPED)",
			},
			"level": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Alert level to filter results (P1, P2, P3, P4)",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform apply`)",
			},
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of ARMS alert rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the alert rule",
						},
						"alert_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the alert rule",
						},
						"alert_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the alert rule",
						},
						"alert_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alert type",
						},
						"alert_check_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alert check type",
						},
						"level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alert level",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alert message template",
						},
						"duration": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Duration threshold",
						},
						"promql": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "PromQL expression for Prometheus alerts",
						},
						"cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cluster ID for Prometheus alerts",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alert rule status",
						},
						"auto_add_new_application": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Auto add new applications",
						},
						"pids": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "Process IDs for application monitoring",
						},
						"annotations": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Alert rule annotations",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Annotation key",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Annotation value",
									},
								},
							},
						},
						"labels": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Alert rule labels",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Label key",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Label value",
									},
								},
							},
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Alert rule tags",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Tag key",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Tag value",
									},
								},
							},
						},
						"notify_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Notification type",
						},
						"dispatch_rule_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Dispatch rule ID",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Region",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource group ID",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time",
						},
						"update_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Update time",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertRulesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Build filter parameters from input
	alertName := ""
	if v, ok := d.GetOk("alert_name"); ok {
		alertName = v.(string)
	}

	alertType := ""
	if v, ok := d.GetOk("alert_type"); ok {
		alertType = v.(string)
	}

	clusterId := ""
	if v, ok := d.GetOk("cluster_id"); ok {
		clusterId = v.(string)
	}

	status := ""
	if v, ok := d.GetOk("status"); ok {
		status = v.(string)
	}

	level := ""
	if v, ok := d.GetOk("level"); ok {
		level = v.(string)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		nameRegex = r
	}

	var ids []string
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			ids = append(ids, vv.(string))
		}
	}

	// Create ArmsService and list alert rules
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all alert rules by paginating through all pages
	var allRules []*armsAPI.AlertRule
	page := DefaultStartPage
	pageSize := DefaultPageSize

	for {
		// Get rules from current page (using simplified API)
		rules, totalCount, err := service.ListArmsAlertRules(page, pageSize, alertName, alertType, clusterId, status, ids, nameRegex)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_rules", "ListArmsAlertRules", AlibabaCloudSdkGoERROR)
		}

		allRules = append(allRules, rules...)

		// Check if we've reached the end
		if len(rules) == 0 || int64(len(rules)) < pageSize {
			break
		}

		// If we have processed all available rules, break
		if page*pageSize >= totalCount {
			break
		}

		page++

		// Safety check to prevent infinite loops
		if page > MaxSafePages {
			break
		}
	}

	// Apply level filter if specified (client-side filtering)
	var filteredRules []*armsAPI.AlertRule
	for _, rule := range allRules {
		if level != "" && rule.Level != level {
			continue
		}
		filteredRules = append(filteredRules, rule)
	}

	// Build output data using strong typing and helper functions
	ruleIds := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, rule := range filteredRules {
		alertIdStr := strconv.FormatInt(rule.AlertId, 10)

		// Convert AlertRule to terraform schema format using helper function
		mapping := convertAlertRuleToTerraformSchema(rule)

		ruleIds = append(ruleIds, alertIdStr)
		names = append(names, rule.AlertName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ruleIds))
	if err := d.Set("ids", ruleIds); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("rules", s); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}

// =============================================================================
// Helper Functions for AlertRule Type Conversion
// =============================================================================

// convertAlertRuleToTerraformSchema converts API AlertRule to Terraform schema format
func convertAlertRuleToTerraformSchema(rule *armsAPI.AlertRule) map[string]interface{} {
	if rule == nil {
		return nil
	}

	alertIdStr := strconv.FormatInt(rule.AlertId, 10)
	dispatchRuleIdStr := strconv.FormatInt(rule.DispatchRuleId, 10)

	// Convert annotations using helper function
	annotations := convertAnnotationsToTerraformFormat(rule.Annotations)

	// Convert labels using helper function
	labels := convertLabelsToTerraformFormat(rule.Labels)

	// Convert tags using helper function
	tags := convertTagsToTerraformFormat(rule.Tags)

	// Create rule mapping with comprehensive fields from AlertRule struct
	mapping := map[string]interface{}{
		"id":                       alertIdStr,
		"alert_id":                 alertIdStr,
		"alert_name":               rule.AlertName,
		"alert_type":               rule.AlertType,
		"alert_check_type":         rule.AlertCheckType,
		"level":                    rule.Level,
		"message":                  rule.Message,
		"duration":                 rule.Duration,
		"promql":                   rule.PromQL,
		"cluster_id":               rule.ClusterId,
		"status":                   rule.Status,
		"auto_add_new_application": rule.AutoAddNewApplication,
		"pids":                     rule.Pids,
		"annotations":              annotations,
		"labels":                   labels,
		"tags":                     tags,
		"notify_type":              rule.NotifyType,
		"dispatch_rule_id":         dispatchRuleIdStr,
		"region":                   rule.Region,
		"resource_group_id":        rule.ResourceGroupId,
		"create_time":              rule.CreateTime,
		"update_time":              rule.UpdateTime,
	}

	return mapping
}

// convertAnnotationsToTerraformFormat converts AlertRuleAnnotation slice to Terraform format
func convertAnnotationsToTerraformFormat(annotations []armsAPI.AlertRuleAnnotation) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(annotations))
	for _, annotation := range annotations {
		result = append(result, map[string]interface{}{
			"key":   annotation.Key,
			"value": annotation.Value,
		})
	}
	return result
}

// convertLabelsToTerraformFormat converts AlertRuleLabel slice to Terraform format
func convertLabelsToTerraformFormat(labels []armsAPI.AlertRuleLabel) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(labels))
	for _, label := range labels {
		result = append(result, map[string]interface{}{
			"key":   label.Key,
			"value": label.Value,
		})
	}
	return result
}

// convertTagsToTerraformFormat converts AlertRuleTag slice to Terraform format
func convertTagsToTerraformFormat(tags []armsAPI.AlertRuleTag) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tags))
	for _, tag := range tags {
		result = append(result, map[string]interface{}{
			"key":   tag.Key,
			"value": tag.Value,
		})
	}
	return result
}
