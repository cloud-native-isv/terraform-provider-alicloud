package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertContactSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertContactScheduleCreate,
		Read:   resourceAliCloudArmsAlertContactScheduleRead,
		Update: resourceAliCloudArmsAlertContactScheduleUpdate,
		Delete: resourceAliCloudArmsAlertContactScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"schedule_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alert_robot_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Asia/Shanghai",
			},
			"schedule_rules": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceAliCloudArmsAlertContactScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateContactSchedule"
	request := make(map[string]interface{})
	var err error

	request["ScheduleName"] = d.Get("schedule_name")
	if v, ok := d.GetOk("description"); ok {
		request["Description"] = v
	}
	if v, ok := d.GetOk("alert_robot_id"); ok {
		request["AlertRobotId"] = v
	}
	if v, ok := d.GetOkExists("state"); ok {
		request["State"] = v
	}
	if v, ok := d.GetOk("time_zone"); ok {
		request["TimeZone"] = v
	}
	if v, ok := d.GetOk("schedule_rules"); ok {
		request["ScheduleRules"] = v
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact_schedule", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["ScheduleId"]))

	return resourceAliCloudArmsAlertContactScheduleRead(d, meta)
}

func resourceAliCloudArmsAlertContactScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)
	contactSchedule, err := armsService.DescribeArmsAlertContactSchedule(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_contact_schedule armsService.DescribeArmsAlertContactSchedule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("schedule_name", contactSchedule.ScheduleName)
	d.Set("description", contactSchedule.Description)
	if contactSchedule.AlertRobotId > 0 {
		d.Set("alert_robot_id", fmt.Sprint(contactSchedule.AlertRobotId))
	}
	// Note: State field may need to be mapped from schedule status or a specific field
	// d.Set("state", contactSchedule.State)
	// Note: TimeZone field may not be available in the current AlertContactSchedule struct
	// d.Set("time_zone", contactSchedule.TimeZone)
	// Note: ScheduleRules field may not be available in the current AlertContactSchedule struct
	// d.Set("schedule_rules", contactSchedule.ScheduleRules)
	// Note: CreateTime and UpdateTime fields may not be available in the current AlertContactSchedule struct
	// d.Set("create_time", contactSchedule.CreateTime)
	// d.Set("update_time", contactSchedule.UpdateTime)

	return nil
}

func resourceAliCloudArmsAlertContactScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	var err error
	update := false
	request := map[string]interface{}{
		"ScheduleId": d.Id(),
	}

	if d.HasChange("schedule_name") {
		update = true
	}
	request["ScheduleName"] = d.Get("schedule_name")

	if d.HasChange("description") {
		update = true
	}
	if v, ok := d.GetOk("description"); ok {
		request["Description"] = v
	}

	if d.HasChange("alert_robot_id") {
		update = true
	}
	if v, ok := d.GetOk("alert_robot_id"); ok {
		request["AlertRobotId"] = v
	}

	if d.HasChange("state") {
		update = true
	}
	if v, ok := d.GetOkExists("state"); ok {
		request["State"] = v
	}

	if d.HasChange("time_zone") {
		update = true
	}
	if v, ok := d.GetOk("time_zone"); ok {
		request["TimeZone"] = v
	}

	if d.HasChange("schedule_rules") {
		update = true
	}
	if v, ok := d.GetOk("schedule_rules"); ok {
		request["ScheduleRules"] = v
	}

	request["RegionId"] = client.RegionId

	if update {
		action := "UpdateContactSchedule"
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

	return resourceAliCloudArmsAlertContactScheduleRead(d, meta)
}

func resourceAliCloudArmsAlertContactScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteContactSchedule"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"ScheduleId": d.Id(),
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
