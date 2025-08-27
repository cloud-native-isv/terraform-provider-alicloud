package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudArmsAlertIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertIntegrationCreate,
		Read:   resourceAliCloudArmsAlertIntegrationRead,
		Update: resourceAliCloudArmsAlertIntegrationUpdate,
		Delete: resourceAliCloudArmsAlertIntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"integration_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the integration.",
			},
			"product_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ARMS", "CLOUD_MONITOR", "MSE", "ARMS_CLOUD_DIALTEST", "PROMETHEUS",
					"LOG_SERVICE", "CUSTOM", "ARMS_PROMETHEUS", "ARMS_APP_MON",
					"ARMS_FRONT_MON", "ARMS_CUSTOM", "XTRACE", "GRAFANA", "ZABBIX",
					"SKYWALKING", "EVENT_BRIDGE", "NAGIOS", "OPENFALCON", "ARMS_INSIGHTS",
				}, false),
				Description: "The integration type. Valid values: ARMS, CLOUD_MONITOR, MSE, ARMS_CLOUD_DIALTEST, PROMETHEUS, LOG_SERVICE, CUSTOM, ARMS_PROMETHEUS, ARMS_APP_MON, ARMS_FRONT_MON, ARMS_CUSTOM, XTRACE, GRAFANA, ZABBIX, SKYWALKING, EVENT_BRIDGE, NAGIOS, OPENFALCON, ARMS_INSIGHTS.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_recover": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"recover_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"duplicate_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"liveness": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"field_redefine_rules": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Field redefine rules for the integration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field.",
						},
						"field_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the field.",
						},
						"redefine_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The redefine type.",
						},
						"json_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The JSON path.",
						},
						"expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The expression.",
						},
						"mapping_rules": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The mapping rules.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The original value.",
									},
									"mapping_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The mapping value.",
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
				Description: "Extended field redefine rules for the integration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the field.",
						},
						"field_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the field.",
						},
						"redefine_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The redefine type.",
						},
						"json_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The JSON path.",
						},
						"expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The expression.",
						},
						"mapping_rules": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The mapping rules.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The original value.",
									},
									"mapping_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The mapping value.",
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

// Helper function to convert terraform schema list to FieldRedefineRule slice
func expandFieldRedefineRules(rules []interface{}) []arms.AlertIntegrationFieldRedefineRule {
	var result []arms.AlertIntegrationFieldRedefineRule
	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		fieldRule := arms.AlertIntegrationFieldRedefineRule{
			FieldName:    ruleMap["field_name"].(string),
			FieldType:    ruleMap["field_type"].(string),
			RedefineType: ruleMap["redefine_type"].(string),
		}
		if v, ok := ruleMap["json_path"]; ok && v != "" {
			fieldRule.JsonPath = v.(string)
		}
		if v, ok := ruleMap["expression"]; ok && v != "" {
			fieldRule.Expression = v.(string)
		}
		if mappingRulesInterface, ok := ruleMap["mapping_rules"]; ok {
			mappingRules := mappingRulesInterface.([]interface{})
			for _, mappingRule := range mappingRules {
				mappingRuleMap := mappingRule.(map[string]interface{})
				fieldRule.MappingRules = append(fieldRule.MappingRules, arms.AlertIntegrationMappingRule{
					OriginValue:  mappingRuleMap["origin_value"].(string),
					MappingValue: mappingRuleMap["mapping_value"].(string),
				})
			}
		}
		result = append(result, fieldRule)
	}
	return result
}

// Helper function to convert FieldRedefineRule slice to terraform schema list
func flattenFieldRedefineRules(rules []arms.AlertIntegrationFieldRedefineRule) []map[string]interface{} {
	var result []map[string]interface{}
	for _, rule := range rules {
		ruleMap := map[string]interface{}{
			"field_name":    rule.FieldName,
			"field_type":    rule.FieldType,
			"redefine_type": rule.RedefineType,
		}
		if rule.JsonPath != "" {
			ruleMap["json_path"] = rule.JsonPath
		}
		if rule.Expression != "" {
			ruleMap["expression"] = rule.Expression
		}
		var mappingRules []map[string]interface{}
		for _, mappingRule := range rule.MappingRules {
			mappingRules = append(mappingRules, map[string]interface{}{
				"origin_value":  mappingRule.OriginValue,
				"mapping_value": mappingRule.MappingValue,
			})
		}
		if len(mappingRules) > 0 {
			ruleMap["mapping_rules"] = mappingRules
		}
		result = append(result, ruleMap)
	}
	return result
}

func resourceAliCloudArmsAlertIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationName := d.Get("integration_name").(string)
	integrationProductType := d.Get("product_type").(string)
	description := d.Get("description").(string)
	autoRecover := d.Get("auto_recover").(bool)
	recoverTime := int64(d.Get("recover_time").(int))

	// Process field redefine rules
	var fieldRedefineRules []arms.AlertIntegrationFieldRedefineRule
	if v, ok := d.GetOk("field_redefine_rules"); ok {
		fieldRedefineRules = expandFieldRedefineRules(v.([]interface{}))
	}

	// Process extended field redefine rules
	var extendedFieldRedefineRules []arms.AlertIntegrationFieldRedefineRule
	if v, ok := d.GetOk("extended_field_redefine_rules"); ok {
		extendedFieldRedefineRules = expandFieldRedefineRules(v.([]interface{}))
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		// Create AlertIntegration object
		integration := &arms.AlertIntegration{
			IntegrationName:            integrationName,
			IntegrationProductType:     integrationProductType,
			Description:                description,
			AutoRecover:                autoRecover,
			RecoverTime:                recoverTime,
			FieldRedefineRules:         fieldRedefineRules,
			ExtendedFieldRedefineRules: extendedFieldRedefineRules,
		}
		result, err := armsAPI.CreateIntegration(integration)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		d.SetId(fmt.Sprint(result.IntegrationId))
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "CreateIntegration", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudArmsAlertIntegrationRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	integration, err := armsAPI.GetIntegrationById(integrationId, true)
	if err != nil {
		if IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_integration GetIntegrationByID Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("integration_name", integration.IntegrationName)
	d.Set("product_type", integration.IntegrationProductType)
	d.Set("state", integration.State)
	d.Set("api_endpoint", integration.ApiEndpoint)
	d.Set("short_token", integration.ShortToken)
	d.Set("liveness", integration.Liveness)
	d.Set("create_time", integration.CreateTime)

	// Set integration detail fields if available
	if integration.IntegrationDetail != nil {
		d.Set("description", integration.IntegrationDetail.Description)
		d.Set("auto_recover", integration.IntegrationDetail.AutoRecover)
		d.Set("recover_time", integration.IntegrationDetail.RecoverTime)
		d.Set("duplicate_key", integration.IntegrationDetail.DuplicateKey)

		// Set field redefine rules
		if integration.IntegrationDetail.FieldRedefineRules != nil {
			d.Set("field_redefine_rules", flattenFieldRedefineRules(integration.IntegrationDetail.FieldRedefineRules))
		}

		// Set extended field redefine rules
		if integration.IntegrationDetail.ExtendedFieldRedefineRules != nil {
			d.Set("extended_field_redefine_rules", flattenFieldRedefineRules(integration.IntegrationDetail.ExtendedFieldRedefineRules))
		}
	}

	return nil
}

func resourceAliCloudArmsAlertIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	update := false
	integrationName := d.Get("integration_name").(string)
	integrationProductType := d.Get("product_type").(string)
	description := d.Get("description").(string)
	autoRecover := d.Get("auto_recover").(bool)
	recoverTime := int64(d.Get("recover_time").(int))
	duplicateKey := d.Get("duplicate_key").(string)
	state := ""
	if d.Get("state").(bool) {
		state = "active"
	} else {
		state = "inactive"
	}
	liveness := d.Get("liveness").(string) // This is computed, so it might be empty for updates

	// Process field redefine rules
	var fieldRedefineRules []arms.AlertIntegrationFieldRedefineRule
	if v, ok := d.GetOk("field_redefine_rules"); ok {
		fieldRedefineRules = expandFieldRedefineRules(v.([]interface{}))
	}

	// Process extended field redefine rules
	var extendedFieldRedefineRules []arms.AlertIntegrationFieldRedefineRule
	if v, ok := d.GetOk("extended_field_redefine_rules"); ok {
		extendedFieldRedefineRules = expandFieldRedefineRules(v.([]interface{}))
	}

	if d.HasChange("integration_name") || d.HasChange("description") || d.HasChange("auto_recover") || d.HasChange("recover_time") || d.HasChange("duplicate_key") || d.HasChange("state") || d.HasChange("field_redefine_rules") || d.HasChange("extended_field_redefine_rules") {
		update = true
	}

	if update {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			// Create AlertIntegration object
			integration := &arms.AlertIntegration{
				IntegrationId:              integrationId,
				IntegrationName:            integrationName,
				IntegrationProductType:     integrationProductType,
				Description:                description,
				AutoRecover:                autoRecover,
				RecoverTime:                recoverTime,
				DuplicateKey:               duplicateKey,
				State:                      state == "active",
				Liveness:                   liveness,
				FieldRedefineRules:         fieldRedefineRules,
				ExtendedFieldRedefineRules: extendedFieldRedefineRules,
			}
			_, err := armsAPI.UpdateIntegration(integration)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateIntegration", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudArmsAlertIntegrationRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := armsAPI.DeleteIntegration(integrationId)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteIntegration", AlibabaCloudSdkGoERROR)
	}

	return nil
}
