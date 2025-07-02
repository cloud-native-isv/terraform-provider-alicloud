package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertNotificationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertNotificationPolicyCreate,
		Read:   resourceAliCloudArmsAlertNotificationPolicyRead,
		Update: resourceAliCloudArmsAlertNotificationPolicyUpdate,
		Delete: resourceAliCloudArmsAlertNotificationPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"notification_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"send_recover_message": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"repeat_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  120,
			},
			"escalation_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudArmsAlertNotificationPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateNotificationPolicy"
	request := make(map[string]interface{})
	var err error

	request["Name"] = d.Get("notification_policy_name")
	if v, ok := d.GetOkExists("send_recover_message"); ok {
		request["SendRecoverMessage"] = v
	}
	if v, ok := d.GetOk("repeat_interval"); ok {
		request["RepeatInterval"] = v
	}
	if v, ok := d.GetOk("escalation_policy_id"); ok {
		request["EscalationPolicyId"] = v
	}
	if v, ok := d.GetOkExists("state"); ok {
		request["IsEnable"] = v
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_notification_policy", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["Id"]))

	return resourceAliCloudArmsAlertNotificationPolicyRead(d, meta)
}

func resourceAliCloudArmsAlertNotificationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)
	object, err := armsService.DescribeArmsAlertNotificationPolicy(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_notification_policy armsService.DescribeArmsAlertNotificationPolicy Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("notification_policy_name", object["Name"])
	d.Set("send_recover_message", object["SendRecoverMessage"])
	d.Set("repeat_interval", object["RepeatInterval"])
	d.Set("escalation_policy_id", fmt.Sprint(object["EscalationPolicyId"]))
	d.Set("state", object["State"])
	d.Set("create_time", fmt.Sprint(object["CreateTime"]))
	d.Set("update_time", fmt.Sprint(object["UpdateTime"]))

	return nil
}

func resourceAliCloudArmsAlertNotificationPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	var err error
	update := false
	request := map[string]interface{}{
		"Id": d.Id(),
	}

	if d.HasChange("notification_policy_name") {
		update = true
	}
	request["Name"] = d.Get("notification_policy_name")

	if d.HasChange("send_recover_message") {
		update = true
	}
	if v, ok := d.GetOkExists("send_recover_message"); ok {
		request["SendRecoverMessage"] = v
	}

	if d.HasChange("repeat_interval") {
		update = true
	}
	if v, ok := d.GetOk("repeat_interval"); ok {
		request["RepeatInterval"] = v
	}

	if d.HasChange("escalation_policy_id") {
		update = true
	}
	if v, ok := d.GetOk("escalation_policy_id"); ok {
		request["EscalationPolicyId"] = v
	}

	if d.HasChange("state") {
		update = true
	}
	if v, ok := d.GetOkExists("state"); ok {
		request["IsEnable"] = v
	}

	request["RegionId"] = client.RegionId

	if update {
		action := "UpdateNotificationPolicy"
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

	return resourceAliCloudArmsAlertNotificationPolicyRead(d, meta)
}

func resourceAliCloudArmsAlertNotificationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteNotificationPolicy"
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
