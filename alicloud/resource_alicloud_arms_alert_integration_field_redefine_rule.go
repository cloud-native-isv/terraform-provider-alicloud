package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudArmsAlertIntegrationFieldRedefineRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertIntegrationFieldRedefineRuleCreate,
		Read:   resourceAliCloudArmsAlertIntegrationFieldRedefineRuleRead,
		Update: resourceAliCloudArmsAlertIntegrationFieldRedefineRuleUpdate,
		Delete: resourceAliCloudArmsAlertIntegrationFieldRedefineRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the alert integration",
			},
			"auto_recover": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to enable auto recovery",
			},
			"recover_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(60, 86400),
				Description:  "Auto recovery time in seconds (60-86400)",
			},
			"field_redefine_rules": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Field redefine rules configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field",
						},
						"field_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"STRING", "NUMBER", "BOOLEAN"}, false),
							Description:  "The type of the field (STRING, NUMBER, BOOLEAN)",
						},
						"redefine_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"JSON_PATH", "CONSTANT", "MAPPING"}, false),
							Description:  "The redefine type (JSON_PATH, CONSTANT, MAPPING)",
						},
						"json_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "JSON path expression (used when redefine_type is JSON_PATH)",
						},
						"expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Constant expression (used when redefine_type is CONSTANT)",
						},
						"mapping_rules": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Mapping rules (used when redefine_type is MAPPING)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Original value to map from",
									},
									"mapping_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Value to map to",
									},
								},
							},
						},
					},
				},
			},
			"extended_field_redefine_rules": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Extended field redefine rules configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field",
						},
						"field_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"STRING", "NUMBER", "BOOLEAN"}, false),
							Description:  "The type of the field (STRING, NUMBER, BOOLEAN)",
						},
						"redefine_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"JSON_PATH", "CONSTANT", "MAPPING"}, false),
							Description:  "The redefine type (JSON_PATH, CONSTANT, MAPPING)",
						},
						"json_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "JSON path expression (used when redefine_type is JSON_PATH)",
						},
						"expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Constant expression (used when redefine_type is CONSTANT)",
						},
						"mapping_rules": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Mapping rules (used when redefine_type is MAPPING)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Original value to map from",
									},
									"mapping_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Value to map to",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudArmsAlertIntegrationFieldRedefineRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	integrationId := d.Get("integration_id").(string)

	// Create uses the same logic as Update - modifying the integration
	err := updateIntegrationFieldRedefineRules(d, client, integrationId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration_field_redefine_rule", "CreateFieldRedefineRules", AlibabaCloudSdkGoERROR)
	}

	// Use integration_id as the resource ID
	d.SetId(integrationId)

	return resourceAliCloudArmsAlertIntegrationFieldRedefineRuleRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationFieldRedefineRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := armsService.DescribeArmsIntegration(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_integration_field_redefine_rule armsService.DescribeArmsIntegration Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("integration_id", object.IntegrationId)
	d.Set("auto_recover", object.AutoRecover)
	d.Set("recover_time", object.RecoverTime)

	// Set field_redefine_rules and extended_field_redefine_rules from object
	if len(object.FieldRedefineRules) > 0 {
		d.Set("field_redefine_rules", flattenAlertIntegrationFieldRedefineRules(object.FieldRedefineRules))
	}

	if len(object.ExtendedFieldRedefineRules) > 0 {
		d.Set("extended_field_redefine_rules", flattenAlertIntegrationFieldRedefineRules(object.ExtendedFieldRedefineRules))
	}

	return nil
}

func resourceAliCloudArmsAlertIntegrationFieldRedefineRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	integrationId := d.Id()

	err := updateIntegrationFieldRedefineRules(d, client, integrationId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateFieldRedefineRules", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudArmsAlertIntegrationFieldRedefineRuleRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationFieldRedefineRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	integrationId := d.Id()

	// Delete means clearing all field redefine rules
	request := map[string]interface{}{
		"RegionId":      client.RegionId,
		"IntegrationId": integrationId,
		"AutoRecover":   false,
		"RecoverTime":   300,
		// Clear all rules by not sending them
	}

	action := "UpdateIntegration"
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		if IsExpectedErrors(err, []string{"404"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// Helper function to update integration with field redefine rules
func updateIntegrationFieldRedefineRules(d *schema.ResourceData, client *connectivity.AliyunClient, integrationId string) error {
	request := map[string]interface{}{
		"RegionId":      client.RegionId,
		"IntegrationId": integrationId,
	}

	if v, ok := d.GetOk("auto_recover"); ok {
		request["AutoRecover"] = v
	}

	if v, ok := d.GetOk("recover_time"); ok {
		request["RecoverTime"] = v
	}

	if v, ok := d.GetOk("field_redefine_rules"); ok {
		rules := expandFieldRedefineRules(v.([]interface{}))
		request["FieldRedefineRules"] = rules
	}

	if v, ok := d.GetOk("extended_field_redefine_rules"); ok {
		rules := expandFieldRedefineRules(v.([]interface{}))
		request["ExtendedFieldRedefineRules"] = rules
	}

	action := "UpdateIntegration"
	wait := incrementalWait(3*time.Second, 3*time.Second)
	return resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
}

// Helper function to expand field redefine rules from schema to API format
func expandFieldRedefineRules(rules []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		apiRule := map[string]interface{}{
			"FieldName":    ruleMap["field_name"],
			"FieldType":    ruleMap["field_type"],
			"RedefineType": ruleMap["redefine_type"],
		}

		if jsonPath, ok := ruleMap["json_path"]; ok && jsonPath != "" {
			apiRule["JsonPath"] = jsonPath
		}

		if expression, ok := ruleMap["expression"]; ok && expression != "" {
			apiRule["Expression"] = expression
		}

		if mappingRules, ok := ruleMap["mapping_rules"]; ok {
			apiRule["MappingRules"] = expandMappingRules(mappingRules.([]interface{}))
		}

		result = append(result, apiRule)
	}

	return result
}

// Helper function to expand mapping rules
func expandMappingRules(rules []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		result = append(result, map[string]interface{}{
			"OriginValue":  ruleMap["origin_value"],
			"MappingValue": ruleMap["mapping_value"],
		})
	}

	return result
}

// Helper function to flatten field redefine rules from API format to schema
func flattenFieldRedefineRules(rules []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		schemaRule := map[string]interface{}{
			"field_name":    ruleMap["FieldName"],
			"field_type":    ruleMap["FieldType"],
			"redefine_type": ruleMap["RedefineType"],
		}

		if jsonPath, ok := ruleMap["JsonPath"]; ok {
			schemaRule["json_path"] = jsonPath
		}

		if expression, ok := ruleMap["Expression"]; ok {
			schemaRule["expression"] = expression
		}

		if mappingRules, ok := ruleMap["MappingRules"]; ok {
			schemaRule["mapping_rules"] = flattenMappingRules(mappingRules.([]interface{}))
		}

		result = append(result, schemaRule)
	}

	return result
}

// Helper function to flatten mapping rules
func flattenMappingRules(rules []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		result = append(result, map[string]interface{}{
			"origin_value":  ruleMap["OriginValue"],
			"mapping_value": ruleMap["MappingValue"],
		})
	}

	return result
}

// Helper function to flatten AlertIntegrationFieldRedefineRule from API format to schema
func flattenAlertIntegrationFieldRedefineRules(rules []aliyunArmsAPI.AlertIntegrationFieldRedefineRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		schemaRule := map[string]interface{}{
			"field_name":    rule.FieldName,
			"field_type":    rule.FieldType,
			"redefine_type": rule.RedefineType,
		}

		if rule.JsonPath != "" {
			schemaRule["json_path"] = rule.JsonPath
		}

		if rule.Expression != "" {
			schemaRule["expression"] = rule.Expression
		}

		if len(rule.MappingRules) > 0 {
			schemaRule["mapping_rules"] = flattenAlertIntegrationMappingRules(rule.MappingRules)
		}

		result = append(result, schemaRule)
	}

	return result
}

// Helper function to flatten AlertIntegrationMappingRule
func flattenAlertIntegrationMappingRules(rules []aliyunArmsAPI.AlertIntegrationMappingRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))

	for _, rule := range rules {
		result = append(result, map[string]interface{}{
			"origin_value":  rule.OriginValue,
			"mapping_value": rule.MappingValue,
		})
	}

	return result
}
