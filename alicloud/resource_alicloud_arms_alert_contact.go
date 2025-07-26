package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertContact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertContactCreate,
		Read:   resourceAliCloudArmsAlertContactRead,
		Update: resourceAliCloudArmsAlertContactUpdate,
		Delete: resourceAliCloudArmsAlertContactDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"alert_contact_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ding_robot_webhook_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_num": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"receive_system_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAliCloudArmsAlertContactCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateAlertContact"
	request := make(map[string]interface{})
	var err error
	if v, ok := d.GetOk("alert_contact_name"); ok {
		request["ContactName"] = v
	}
	if v, ok := d.GetOk("ding_robot_webhook_url"); ok {
		request["DingRobotWebhookUrl"] = v
	} else if v, ok := d.GetOk("email"); ok && v.(string) == "" {
		if v, ok := d.GetOk("phone_num"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "ding_robot_webhook_url", "email", d.Get("email"), "phone_num", d.Get("phone_num")))
		}
	}
	if v, ok := d.GetOk("email"); ok {
		request["Email"] = v
	} else if v, ok := d.GetOk("ding_robot_webhook_url"); ok && v.(string) == "" {
		if v, ok := d.GetOk("phone_num"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "email", "ding_robot_webhook_url", d.Get("ding_robot_webhook_url"), "phone_num", d.Get("phone_num")))
		}
	}
	if v, ok := d.GetOk("phone_num"); ok {
		request["Phone"] = v
	} else if v, ok := d.GetOk("ding_robot_webhook_url"); ok && v.(string) == "" {
		if v, ok := d.GetOk("email"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "phone_num", "ding_robot_webhook_url", d.Get("ding_robot_webhook_url"), "email", d.Get("email")))
		}
	}
	if v, ok := d.GetOkExists("receive_system_notification"); ok {
		request["SystemNoc"] = v
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["ContactId"]))

	return resourceAliCloudArmsAlertContactRead(d, meta)
}

func resourceAliCloudArmsAlertContactRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)

	_, err := armsService.DescribeArmsAlertContact(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_contact armsService.DescribeArmsAlertContact Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	// TODO: use strong typed alertContact
	// d.Set("alert_contact_name", alertContact.ContactName)
	// d.Set("ding_robot_webhook_url", alertContact.Webhook)
	// d.Set("email", alertContact.Email)
	// d.Set("phone_num", alertContact.Phone)
	// d.Set("receive_system_notification", alertContact.SystemNoc)

	return nil
}

func resourceAliCloudArmsAlertContactUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var err error
	var response map[string]interface{}
	update := false
	request := map[string]interface{}{
		"ContactId": d.Id(),
	}
	request["RegionId"] = client.RegionId
	if d.HasChange("alert_contact_name") {
		update = true
	}
	if v, ok := d.GetOk("alert_contact_name"); ok {
		request["ContactName"] = v
	}
	if d.HasChange("ding_robot_webhook_url") {
		update = true
	}
	if v, ok := d.GetOk("ding_robot_webhook_url"); ok {
		request["DingRobotWebhookUrl"] = v
	} else if v, ok := d.GetOk("email"); ok && v.(string) == "" {
		if v, ok := d.GetOk("phone_num"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "ding_robot_webhook_url", "email", d.Get("email"), "phone_num", d.Get("phone_num")))
		}
	}
	if d.HasChange("email") {
		update = true
	}
	if v, ok := d.GetOk("email"); ok {
		request["Email"] = v
	} else if v, ok := d.GetOk("ding_robot_webhook_url"); ok && v.(string) == "" {
		if v, ok := d.GetOk("phone_num"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "email", "ding_robot_webhook_url", d.Get("ding_robot_webhook_url"), "phone_num", d.Get("phone_num")))
		}
	}
	if d.HasChange("phone_num") {
		update = true
	}
	if v, ok := d.GetOk("phone_num"); ok {
		request["Phone"] = v
	} else if v, ok := d.GetOk("ding_robot_webhook_url"); ok && v.(string) == "" {
		if v, ok := d.GetOk("email"); ok && v.(string) == "" {
			return WrapError(fmt.Errorf("attribute '%s' is required when '%s' is %v and '%s' is %v ", "phone_num", "ding_robot_webhook_url", d.Get("ding_robot_webhook_url"), "email", d.Get("email")))
		}
	}
	if d.HasChange("receive_system_notification") || d.IsNewResource() {
		update = true
	}
	if v, ok := d.GetOkExists("receive_system_notification"); ok {
		request["SystemNoc"] = v
	}
	if update {
		action := "UpdateAlertContact"
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
	return resourceAliCloudArmsAlertContactRead(d, meta)
}

func resourceAliCloudArmsAlertContactDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteAlertContact"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"ContactId": d.Id(),
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
