package alicloud

import (
	"regexp"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsAlertIntegrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertIntegrationsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"integration_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Active", "Inactive"}, false),
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"integrations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the alert integration.",
						},
						"integration_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the alert integration.",
						},
						"integration_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the alert integration.",
						},
						"integration_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The service type of the alert integration.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the alert integration.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the alert integration. Valid values: Active, Inactive.",
						},
						"api_endpoint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The API endpoint URL of the alert integration.",
						},
						"short_token": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The authentication token of the alert integration.",
						},
						"liveness": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The health status of the alert integration. Valid values: ready, unhealthy, unknown.",
						},
						"auto_recover": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether alert events are automatically cleared.",
						},
						"recover_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The period of time within which alert events are automatically cleared. Unit: minutes.",
						},
						"duplicate_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The fields whose values are deduplicated.",
						},
						"field_redefine_rules": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The predefined mapped fields of the alert source.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The ID of the field redefine rule.",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the field redefine rule.",
									},
									"field_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The target field name.",
									},
									"field_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The field type. Valid values: LABEL, ANNOTATION.",
									},
									"redefine_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The redefine type. Valid values: EXTRACT, MAP, ADD.",
									},
									"json_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The JSON path expression.",
									},
									"expression": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The template expression.",
									},
									"match_expression": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The match expression.",
									},
								},
							},
						},
						"extended_field_redefine_rules": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The extended mapped fields of the alert source.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The ID of the extended field redefine rule.",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the extended field redefine rule.",
									},
									"field_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The target field name.",
									},
									"field_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The field type. Valid values: LABEL, ANNOTATION.",
									},
									"redefine_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The redefine type. Valid values: EXTRACT, MAP, ADD.",
									},
									"json_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The JSON path expression.",
									},
									"expression": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The template expression.",
									},
									"match_expression": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The match expression.",
									},
								},
							},
						},
						"statistics": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The statistics information: [total_count, error_count].",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when the alert integration was created.",
						},
						"update_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when the alert integration was last updated.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertIntegrationsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	var filteredIntegrations []*aliyunArmsAPI.AlertIntegration
	var integrationNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		integrationNameRegex = r
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	// Get integrations from service layer using strong types
	integrations, err := armsService.ListArmsAlertIntegrations()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_integrations", "ListArmsAlertIntegrations", AlibabaCloudSdkGoERROR)
	}

	// Process and filter integrations using strong types
	for _, integration := range integrations {
		// Apply name regex filter
		if integrationNameRegex != nil && !integrationNameRegex.MatchString(integration.IntegrationName) {
			continue
		}

		// Apply IDs filter
		if len(idsMap) > 0 {
			integrationIdStr := strconv.FormatInt(integration.IntegrationId, 10)
			if _, ok := idsMap[integrationIdStr]; !ok {
				continue
			}
		}

		// Apply integration type filter
		if v, ok := d.GetOk("integration_type"); ok {
			if integration.IntegrationProductType != v.(string) {
				continue
			}
		}

		// Apply status filter
		if v, ok := d.GetOk("status"); ok {
			expectedStatus := v.(string)
			var currentStatus string
			if integration.State {
				currentStatus = "Active"
			} else {
				currentStatus = "Inactive"
			}
			if currentStatus != expectedStatus {
				continue
			}
		}

		filteredIntegrations = append(filteredIntegrations, integration)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	integrationsList := make([]interface{}, 0)

	for _, integration := range filteredIntegrations {
		integrationIdStr := strconv.FormatInt(integration.IntegrationId, 10)

		integrationMap := buildIntegrationMap(integration)

		ids = append(ids, integrationIdStr)
		names = append(names, integration.IntegrationName)
		integrationsList = append(integrationsList, integrationMap)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("integrations", integrationsList); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), integrationsList)
	}

	return nil
}

// buildIntegrationMap builds a map from AlertIntegration struct using strong types
func buildIntegrationMap(integration *aliyunArmsAPI.AlertIntegration) map[string]interface{} {
	integrationIdStr := strconv.FormatInt(integration.IntegrationId, 10)

	result := map[string]interface{}{
		"id":               integrationIdStr,
		"integration_id":   integrationIdStr,
		"integration_name": integration.IntegrationName,
		"integration_type": integration.IntegrationProductType,
		"status":           getIntegrationStatusFromBool(integration.State),
		"create_time":      formatTimeFromPtr(integration.CreateTime),
	}

	// Optional basic fields
	if integration.Description != "" {
		result["description"] = integration.Description
	}
	if integration.ApiEndpoint != "" {
		result["api_endpoint"] = integration.ApiEndpoint
	}
	if integration.ShortToken != "" {
		result["short_token"] = integration.ShortToken
	}
	if integration.Liveness != "" {
		result["liveness"] = integration.Liveness
	}
	if integration.DuplicateKey != "" {
		result["duplicate_key"] = integration.DuplicateKey
	}
	if integration.Region != "" {
		result["region"] = integration.Region
	}
	if integration.ResourceGroupId != "" {
		result["resource_group_id"] = integration.ResourceGroupId
	}

	// Boolean and numeric fields
	result["auto_recover"] = integration.AutoRecover
	result["recover_time"] = integration.RecoverTime

	// Statistics array
	if len(integration.Statistics) > 0 {
		result["statistics"] = integration.Statistics
	}

	// Field redefine rules
	if len(integration.FieldRedefineRules) > 0 {
		result["field_redefine_rules"] = buildFieldRedefineRulesList(integration.FieldRedefineRules)
	}

	// Extended field redefine rules
	if len(integration.ExtendedFieldRedefineRules) > 0 {
		result["extended_field_redefine_rules"] = buildFieldRedefineRulesList(integration.ExtendedFieldRedefineRules)
	}

	// Update time handling
	if integration.UpdateTime != nil {
		result["update_time"] = formatTimeFromPtr(integration.UpdateTime)
	} else {
		result["update_time"] = result["create_time"]
	}

	return result
}

// buildFieldRedefineRulesList builds field redefine rules list from AlertIntegrationFieldRedefineRule structs
func buildFieldRedefineRulesList(rules []aliyunArmsAPI.AlertIntegrationFieldRedefineRule) []interface{} {
	result := make([]interface{}, 0, len(rules))

	for _, rule := range rules {
		ruleMap := map[string]interface{}{
			"id":            rule.Id,
			"name":          rule.Name,
			"field_name":    rule.FieldName,
			"field_type":    rule.FieldType,
			"redefine_type": rule.RedefineType,
		}

		// Optional fields
		if rule.JsonPath != "" {
			ruleMap["json_path"] = rule.JsonPath
		}
		if rule.Expression != "" {
			ruleMap["expression"] = rule.Expression
		}
		if rule.MatchExpression != "" {
			ruleMap["match_expression"] = rule.MatchExpression
		}

		result = append(result, ruleMap)
	}

	return result
}

// Helper function to convert boolean state to status string
func getIntegrationStatusFromBool(state bool) string {
	if state {
		return "Active"
	}
	return "Inactive"
}

// Helper function to format time pointer to string
func formatTimeFromPtr(timePtr *time.Time) string {
	if timePtr == nil {
		return ""
	}
	return timePtr.Format(time.RFC3339)
}
