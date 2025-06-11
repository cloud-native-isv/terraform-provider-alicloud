package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudArmsAlertFieldRedefineRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudArmsAlertFieldRedefineRuleCreate,
		Read:   resourceAlicloudArmsAlertFieldRedefineRuleRead,
		Update: resourceAlicloudArmsAlertFieldRedefineRuleUpdate,
		Delete: resourceAlicloudArmsAlertFieldRedefineRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"redefine_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"field_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"field_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"match_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"json_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAlicloudArmsAlertFieldRedefineRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateAlertFieldRedefineRule"
	request := make(map[string]interface{})
	var err error

	request["IntegrationId"] = d.Get("integration_id")
	request["Name"] = d.Get("name")
	request["RedefineType"] = d.Get("redefine_type")
	request["FieldName"] = d.Get("field_name")
	request["FieldType"] = d.Get("field_type")
	if v, ok := d.GetOk("match_expression"); ok {
		request["MatchExpression"] = v
	}
	if v, ok := d.GetOk("expression"); ok {
		request["Expression"] = v
	}
	if v, ok := d.GetOk("json_path"); ok {
		request["JsonPath"] = v
	}
	request["RegionId"] = client.RegionId

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_field_redefine_rule", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["Id"]))

	return resourceAlicloudArmsAlertFieldRedefineRuleRead(d, meta)
}

func resourceAlicloudArmsAlertFieldRedefineRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)
	object, err := armsService.DescribeArmsAlertFieldRedefineRule(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_field_redefine_rule armsService.DescribeArmsAlertFieldRedefineRule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("integration_id", fmt.Sprint(object["IntegrationId"]))
	d.Set("name", object["Name"])
	d.Set("redefine_type", object["RedefineType"])
	d.Set("field_name", object["FieldName"])
	d.Set("field_type", object["FieldType"])
	d.Set("match_expression", object["MatchExpression"])
	d.Set("expression", object["Expression"])
	d.Set("json_path", object["JsonPath"])

	return nil
}

func resourceAlicloudArmsAlertFieldRedefineRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	var err error
	update := false
	request := map[string]interface{}{
		"Id": d.Id(),
	}

	if d.HasChange("name") {
		update = true
	}
	request["Name"] = d.Get("name")

	if d.HasChange("redefine_type") {
		update = true
	}
	request["RedefineType"] = d.Get("redefine_type")

	if d.HasChange("field_name") {
		update = true
	}
	request["FieldName"] = d.Get("field_name")

	if d.HasChange("field_type") {
		update = true
	}
	request["FieldType"] = d.Get("field_type")

	if d.HasChange("match_expression") {
		update = true
	}
	if v, ok := d.GetOk("match_expression"); ok {
		request["MatchExpression"] = v
	}

	if d.HasChange("expression") {
		update = true
	}
	if v, ok := d.GetOk("expression"); ok {
		request["Expression"] = v
	}

	if d.HasChange("json_path") {
		update = true
	}
	if v, ok := d.GetOk("json_path"); ok {
		request["JsonPath"] = v
	}

	request["RegionId"] = client.RegionId

	if update {
		action := "UpdateAlertFieldRedefineRule"
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

	return resourceAlicloudArmsAlertFieldRedefineRuleRead(d, meta)
}

func resourceAlicloudArmsAlertFieldRedefineRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteAlertFieldRedefineRule"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"Id": d.Id(),
	}

	request["RegionId"] = client.RegionId
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
