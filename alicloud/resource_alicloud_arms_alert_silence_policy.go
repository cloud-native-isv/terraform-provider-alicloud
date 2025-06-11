package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudArmsAlertSilencePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudArmsAlertSilencePolicyCreate,
		Read:   resourceAlicloudArmsAlertSilencePolicyRead,
		Update: resourceAlicloudArmsAlertSilencePolicyUpdate,
		Delete: resourceAlicloudArmsAlertSilencePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"silence_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ENABLE",
			},
			"effective_time_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"time_period": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"time_slots": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"matching_rules": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceAlicloudArmsAlertSilencePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateSilencePolicy"
	request := make(map[string]interface{})
	var err error

	request["SilenceName"] = d.Get("silence_name")
	if v, ok := d.GetOk("state"); ok {
		request["State"] = v
	}
	request["EffectiveTimeType"] = d.Get("effective_time_type")
	if v, ok := d.GetOk("time_period"); ok {
		request["TimePeriod"] = v
	}
	if v, ok := d.GetOk("time_slots"); ok {
		request["TimeSlots"] = v
	}
	if v, ok := d.GetOk("start_time"); ok {
		request["StartTime"] = v
	}
	if v, ok := d.GetOk("end_time"); ok {
		request["EndTime"] = v
	}
	if v, ok := d.GetOk("comment"); ok {
		request["Comment"] = v
	}
	if v, ok := d.GetOk("matching_rules"); ok {
		request["MatchingRules"] = v
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_silence_policy", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["SilenceId"]))

	return resourceAlicloudArmsAlertSilencePolicyRead(d, meta)
}

func resourceAlicloudArmsAlertSilencePolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := ArmsService{client}
	object, err := armsService.DescribeArmsAlertSilencePolicy(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_silence_policy armsService.DescribeArmsAlertSilencePolicy Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("silence_name", object["SilenceName"])
	d.Set("state", object["State"])
	d.Set("effective_time_type", object["EffectiveTimeType"])
	d.Set("time_period", object["TimePeriod"])
	d.Set("time_slots", object["TimeSlots"])
	d.Set("start_time", fmt.Sprint(object["StartTime"]))
	d.Set("end_time", fmt.Sprint(object["EndTime"]))
	d.Set("comment", object["Comment"])
	d.Set("matching_rules", object["MatchingRules"])
	d.Set("created_by", object["CreatedBy"])
	d.Set("create_time", fmt.Sprint(object["CreateTime"]))
	d.Set("update_time", fmt.Sprint(object["UpdateTime"]))

	return nil
}

func resourceAlicloudArmsAlertSilencePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	var err error
	update := false
	request := map[string]interface{}{
		"SilenceId": d.Id(),
	}

	if d.HasChange("silence_name") {
		update = true
	}
	request["SilenceName"] = d.Get("silence_name")

	if d.HasChange("state") {
		update = true
	}
	if v, ok := d.GetOk("state"); ok {
		request["State"] = v
	}

	if d.HasChange("effective_time_type") {
		update = true
	}
	request["EffectiveTimeType"] = d.Get("effective_time_type")

	if d.HasChange("time_period") {
		update = true
	}
	if v, ok := d.GetOk("time_period"); ok {
		request["TimePeriod"] = v
	}

	if d.HasChange("time_slots") {
		update = true
	}
	if v, ok := d.GetOk("time_slots"); ok {
		request["TimeSlots"] = v
	}

	if d.HasChange("start_time") {
		update = true
	}
	if v, ok := d.GetOk("start_time"); ok {
		request["StartTime"] = v
	}

	if d.HasChange("end_time") {
		update = true
	}
	if v, ok := d.GetOk("end_time"); ok {
		request["EndTime"] = v
	}

	if d.HasChange("comment") {
		update = true
	}
	if v, ok := d.GetOk("comment"); ok {
		request["Comment"] = v
	}

	if d.HasChange("matching_rules") {
		update = true
	}
	if v, ok := d.GetOk("matching_rules"); ok {
		request["MatchingRules"] = v
	}

	request["RegionId"] = client.RegionId

	if update {
		action := "UpdateSilencePolicy"
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

	return resourceAlicloudArmsAlertSilencePolicyRead(d, meta)
}

func resourceAlicloudArmsAlertSilencePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteSilencePolicy"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"SilenceId": d.Id(),
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
