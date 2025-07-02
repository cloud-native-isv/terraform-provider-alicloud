package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"alert_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "The name of the alert rule.",
			},
			"severity": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"P1", "P2", "P3", "P4", "P5", "P6"}, false),
				Description:  "The severity level of the alert rule. Valid values: P1, P2, P3, P4, P5, P6.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the alert rule.",
			},
			"integration_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ARMS",
				Description: "The integration type of the alert rule.",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The cluster ID associated with the alert rule.",
			},
			"expression": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The PromQL expression for the alert rule.",
			},
			"duration": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
				Description: "The duration for which the condition must be true to fire the alert (in seconds).",
			},
			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The message template for the alert.",
			},
			"check_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "CUSTOM",
				Description: "The check type of the alert rule.",
			},
			"alert_group": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "The alert group ID.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The labels for the alert rule.",
			},
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The owner of the alert rule.",
			},
			"handler": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The current handler of the alert rule.",
			},
			"solution": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The solution for the alert.",
			},
			// Computed fields
			"alert_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the alert rule.",
			},
			"state": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The state of the alert rule. 0=pending, 1=processing, 2=resolved.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the alert rule.",
			},
			"dispatch_rule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The dispatch rule ID associated with the alert.",
			},
			"dispatch_rule_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The dispatch rule name associated with the alert.",
			},
		},
	}
}

func resourceAliCloudArmsAlertRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)

	// Prepare rule parameters
	rule := make(map[string]interface{})

	// Set expression/PromQL
	if v, ok := d.GetOk("expression"); ok {
		rule["expression"] = v
	}

	// Set duration
	if v, ok := d.GetOk("duration"); ok {
		rule["duration"] = v
	}

	// Set message
	if v, ok := d.GetOk("message"); ok {
		rule["message"] = v
	}

	// Set check type
	if v, ok := d.GetOk("check_type"); ok {
		rule["check_type"] = v
	}

	// Set alert group
	if v, ok := d.GetOk("alert_group"); ok {
		rule["alert_group"] = v
	}

	// Set labels
	if v, ok := d.GetOk("labels"); ok {
		rule["labels"] = v
	}

	// Call service function to create alert rule
	alertId, err := armsService.CreateArmsAlertRule(
		d.Get("alert_name").(string),
		d.Get("severity").(string),
		d.Get("description").(string),
		d.Get("integration_type").(string),
		d.Get("cluster_id").(string),
		rule,
	)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_rule", "CreateArmsAlertRule", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(alertId))

	return resourceAliCloudArmsAlertRuleRead(d, meta)
}

func resourceAliCloudArmsAlertRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)

	alertId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapError(err)
	}

	object, err := armsService.DescribeArmsAlertRule(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_rule armsService.DescribeArmsAlertRule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("alert_id", alertId)
	d.Set("alert_name", object["AlertName"])
	d.Set("severity", object["Severity"])
	d.Set("description", object["Describe"])
	d.Set("owner", object["Owner"])
	d.Set("handler", object["Handler"])
	d.Set("solution", object["Solution"])
	d.Set("create_time", object["CreateTime"])

	if state, ok := object["State"]; ok {
		d.Set("state", formatInt(state))
	}

	if dispatchRuleId, ok := object["DispatchRuleId"]; ok {
		d.Set("dispatch_rule_id", fmt.Sprint(formatInt(dispatchRuleId)))
	}

	if dispatchRuleName, ok := object["DispatchRuleName"]; ok {
		d.Set("dispatch_rule_name", dispatchRuleName)
	}

	return nil
}

func resourceAliCloudArmsAlertRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}

	alertId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapError(err)
	}

	update := false
	request := map[string]interface{}{
		"AlertId":   alertId,
		"AlertType": "PROMETHEUS_MONITORING_ALERT_RULE",
		"RegionId":  client.RegionId,
	}

	if d.HasChange("alert_name") {
		update = true
		request["AlertName"] = d.Get("alert_name")
	}

	if d.HasChange("severity") {
		update = true
		request["Level"] = d.Get("severity")
	}

	if d.HasChange("description") {
		update = true
		request["Message"] = d.Get("description")
	}

	if d.HasChange("cluster_id") {
		update = true
		request["ClusterId"] = d.Get("cluster_id")
	}

	if d.HasChange("expression") {
		update = true
		request["PromQL"] = d.Get("expression")
	}

	if d.HasChange("duration") {
		update = true
		request["Duration"] = d.Get("duration")
	}

	if d.HasChange("message") {
		update = true
		request["Message"] = d.Get("message")
	}

	if d.HasChange("check_type") {
		update = true
		request["AlertCheckType"] = d.Get("check_type")
	}

	if d.HasChange("alert_group") {
		update = true
		request["AlertGroup"] = d.Get("alert_group")
	}

	if d.HasChange("labels") {
		update = true
		if v, ok := d.GetOk("labels"); ok {
			labelsMap := v.(map[string]interface{})
			if len(labelsMap) > 0 {
				labelsMaps := make([]map[string]interface{}, 0)
				for key, value := range labelsMap {
					labelsMaps = append(labelsMaps, map[string]interface{}{
						"name":  key,
						"value": fmt.Sprintf("%v", value),
					})
				}
				if labelString, err := convertArrayObjectToJsonString(labelsMaps); err == nil {
					request["Labels"] = labelString
				} else {
					return WrapError(err)
				}
			}
		}
	}

	if update {
		action := "CreateOrUpdateAlertRule"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
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
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudArmsAlertRuleRead(d, meta)
}

func resourceAliCloudArmsAlertRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteAlertRule"
	var response map[string]interface{}

	alertId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapError(err)
	}

	request := map[string]interface{}{
		"AlertId": alertId,
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}
