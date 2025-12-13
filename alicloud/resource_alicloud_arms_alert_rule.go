package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	armsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudArmsAlertRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertRuleCreate,
		Read:   resourceAliCloudArmsAlertRuleRead,
		Update: resourceAliCloudArmsAlertRuleUpdate,
		Delete: resourceAliCloudArmsAlertRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"alert_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "The name of the alert rule.",
			},
			"level": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"P1", "P2", "P3", "P4"}, false),
				Description:  "The severity level of the alert rule. Valid values: P1, P2, P3, P4.",
			},
			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The message template of the alert rule.",
			},
			"alert_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "PROMETHEUS_MONITORING_ALERT_RULE",
				ValidateFunc: validation.StringInSlice([]string{"PROMETHEUS_MONITORING_ALERT_RULE", "APPLICATION_MONITORING_ALERT_RULE", "BROWSER_MONITORING_ALERT_RULE"}, false),
				Description:  "The type of the alert rule.",
			},
			"alert_check_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "STATIC",
				ValidateFunc: validation.StringInSlice([]string{"STATIC", "CUSTOM"}, false),
				Description:  "The check type of the alert rule.",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The cluster ID for Prometheus alert rules.",
			},
			"promql": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The PromQL expression for Prometheus alert rules.",
			},
			"duration": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1m",
				Description: "The duration threshold for the alert rule (e.g., 1m, 5m, 10m).",
			},
			"auto_add_new_application": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to automatically add new applications to the alert rule.",
			},
			"pids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Process IDs for application monitoring alert rules.",
			},
			"annotations": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Alert rule annotations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Annotation key.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Annotation value.",
						},
					},
				},
			},
			"labels": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Alert rule labels.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Label key.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Label value.",
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Alert rule tags.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Tag key.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Tag value.",
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "RUNNING",
				ValidateFunc: validation.StringInSlice([]string{"RUNNING", "STOPPED"}, false),
				Description:  "The status of the alert rule.",
			},
			"notify_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ALERT_MANAGER",
				ValidateFunc: validation.StringInSlice([]string{"ALERT_MANAGER", "DISPATCH_RULE"}, false),
				Description:  "The notification type of the alert rule.",
			},
			"dispatch_rule_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The dispatch rule ID for the alert rule.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID.",
			},
			// Computed fields
			"alert_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the alert rule.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region of the alert rule.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the alert rule.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The update time of the alert rule.",
			},
		},
	}
}

func resourceAliCloudArmsAlertRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build AlertRule object from schema data
	alertRule, err := buildAlertRuleFromSchema(d)
	if err != nil {
		return WrapError(err)
	}

	// Create alert rule using service layer
	var result *armsAPI.AlertRule
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = service.CreateArmsAlertRule(alertRule)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable", "SystemBusy"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_rule", "CreateArmsAlertRule", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using the alert ID
	d.SetId(strconv.FormatInt(result.AlertId, 10))

	return resourceAliCloudArmsAlertRuleRead(d, meta)
}

func resourceAliCloudArmsAlertRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get alert rule from service layer
	object, err := service.DescribeArmsAlertRule(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_rule DescribeArmsAlertRule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all schema fields from AlertRule object
	d.Set("alert_id", strconv.FormatInt(object.AlertId, 10))
	d.Set("alert_name", object.AlertName)
	d.Set("alert_type", object.AlertType)
	d.Set("alert_check_type", object.AlertCheckType)
	d.Set("level", object.Level)
	d.Set("message", object.Message)
	d.Set("duration", object.Duration)
	d.Set("promql", object.PromQL)
	d.Set("cluster_id", object.ClusterId)
	d.Set("status", object.Status)
	d.Set("auto_add_new_application", object.AutoAddNewApplication)
	d.Set("pids", object.Pids)
	d.Set("notify_type", object.NotifyType)
	d.Set("dispatch_rule_id", strconv.FormatInt(object.DispatchRuleId, 10))
	d.Set("region", object.Region)
	d.Set("resource_group_id", object.ResourceGroupId)
	d.Set("create_time", object.CreateTime)
	d.Set("update_time", object.UpdateTime)

	// Convert annotations to terraform format
	if len(object.Annotations) > 0 {
		annotations := make([]map[string]interface{}, 0, len(object.Annotations))
		for _, annotation := range object.Annotations {
			annotations = append(annotations, map[string]interface{}{
				"key":   annotation.Key,
				"value": annotation.Value,
			})
		}
		d.Set("annotations", annotations)
	}

	// Convert labels to terraform format
	if len(object.Labels) > 0 {
		labels := make([]map[string]interface{}, 0, len(object.Labels))
		for _, label := range object.Labels {
			labels = append(labels, map[string]interface{}{
				"key":   label.Key,
				"value": label.Value,
			})
		}
		d.Set("labels", labels)
	}

	// Convert tags to terraform format
	if len(object.Tags) > 0 {
		tags := make([]map[string]interface{}, 0, len(object.Tags))
		for _, tag := range object.Tags {
			tags = append(tags, map[string]interface{}{
				"key":   tag.Key,
				"value": tag.Value,
			})
		}
		d.Set("tags", tags)
	}

	return nil
}

func resourceAliCloudArmsAlertRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse alert ID from resource ID
	alertId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapError(fmt.Errorf("invalid alert rule ID format: %s", d.Id()))
	}

	// Build AlertRule object from schema data
	alertRule, err := buildAlertRuleFromSchema(d)
	if err != nil {
		return WrapError(err)
	}

	// Set the alert ID for update
	alertRule.AlertId = alertId

	// Update alert rule using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = service.UpdateArmsAlertRule(alertRule)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable", "SystemBusy"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateArmsAlertRule", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudArmsAlertRuleRead(d, meta)
}

func resourceAliCloudArmsAlertRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse alert ID from resource ID
	alertId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapError(fmt.Errorf("invalid alert rule ID format: %s", d.Id()))
	}

	// Delete alert rule using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = service.DeleteArmsAlertRule(alertId)
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable", "SystemBusy"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteArmsAlertRule", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// =============================================================================
// Helper Functions for AlertRule Resource
// =============================================================================

// buildAlertRuleFromSchema converts Terraform schema data to AlertRule object
func buildAlertRuleFromSchema(d *schema.ResourceData) (*armsAPI.AlertRule, error) {
	alertRule := &armsAPI.AlertRule{
		AlertName:             d.Get("alert_name").(string),
		Level:                 d.Get("level").(string),
		AlertType:             d.Get("alert_type").(string),
		AlertCheckType:        d.Get("alert_check_type").(string),
		AutoAddNewApplication: d.Get("auto_add_new_application").(bool),
		NotifyType:            d.Get("notify_type").(string),
	}

	// Set optional string fields
	if v, ok := d.GetOk("message"); ok {
		alertRule.Message = v.(string)
	}
	if v, ok := d.GetOk("cluster_id"); ok {
		alertRule.ClusterId = v.(string)
	}
	if v, ok := d.GetOk("promql"); ok {
		alertRule.PromQL = v.(string)
	}
	if v, ok := d.GetOk("duration"); ok {
		alertRule.Duration = v.(string)
	}
	if v, ok := d.GetOk("status"); ok {
		alertRule.Status = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		alertRule.ResourceGroupId = v.(string)
	}

	// Set dispatch rule ID if provided
	if v, ok := d.GetOk("dispatch_rule_id"); ok {
		dispatchRuleId, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid dispatch_rule_id format: %s", v.(string))
		}
		alertRule.DispatchRuleId = dispatchRuleId
	}

	// Set PIDs array
	if v, ok := d.GetOk("pids"); ok {
		pids := make([]string, 0)
		for _, pid := range v.([]interface{}) {
			if pid != nil {
				pids = append(pids, pid.(string))
			}
		}
		alertRule.Pids = pids
	}

	// Set annotations
	if v, ok := d.GetOk("annotations"); ok {
		annotations := make([]armsAPI.AlertRuleAnnotation, 0)
		for _, annotation := range v.([]interface{}) {
			if annotationMap, ok := annotation.(map[string]interface{}); ok {
				annotations = append(annotations, armsAPI.AlertRuleAnnotation{
					Key:   annotationMap["key"].(string),
					Value: annotationMap["value"].(string),
				})
			}
		}
		alertRule.Annotations = annotations
	}

	// Set labels
	if v, ok := d.GetOk("labels"); ok {
		labels := make([]armsAPI.AlertRuleLabel, 0)
		for _, label := range v.([]interface{}) {
			if labelMap, ok := label.(map[string]interface{}); ok {
				labels = append(labels, armsAPI.AlertRuleLabel{
					Key:   labelMap["key"].(string),
					Value: labelMap["value"].(string),
				})
			}
		}
		alertRule.Labels = labels
	}

	// Set tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make([]armsAPI.AlertRuleTag, 0)
		for _, tag := range v.([]interface{}) {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				tags = append(tags, armsAPI.AlertRuleTag{
					Key:   tagMap["key"].(string),
					Value: tagMap["value"].(string),
				})
			}
		}
		alertRule.Tags = tags
	}

	return alertRule, nil
}
