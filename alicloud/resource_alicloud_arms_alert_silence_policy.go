package alicloud

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertSilencePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertSilencePolicyCreate,
		Read:   resourceAliCloudArmsAlertSilencePolicyRead,
		Update: resourceAliCloudArmsAlertSilencePolicyUpdate,
		Delete: resourceAliCloudArmsAlertSilencePolicyDelete,
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

func resourceAliCloudArmsAlertSilencePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build the policy object using strong types
	policy := &aliyunArmsAPI.AlertSilencePolicy{
		SilenceName:       d.Get("silence_name").(string),
		State:             d.Get("state").(string),
		EffectiveTimeType: d.Get("effective_time_type").(string),
	}

	// Set optional fields
	if v, ok := d.GetOk("time_period"); ok {
		policy.TimePeriod = v.(string)
	}
	if v, ok := d.GetOk("time_slots"); ok {
		policy.TimeSlots = v.(string)
	}
	if v, ok := d.GetOk("start_time"); ok {
		policy.StartTime = v.(string)
	}
	if v, ok := d.GetOk("end_time"); ok {
		policy.EndTime = v.(string)
	}
	if v, ok := d.GetOk("comment"); ok {
		policy.Comment = v.(string)
	}
	// TODO: Parse matching_rules from string to proper structure
	if v, ok := d.GetOk("matching_rules"); ok {
		// For now, keep as placeholder since we need to parse the string format
		_ = v.(string)
		// policy.MatchingRules = parseMatchingRules(v.(string))
	}

	// Create the silence policy using service layer
	var result *aliyunArmsAPI.AlertSilencePolicy
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = armsService.CreateArmsAlertSilencePolicy(policy)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_silence_policy", "CreateArmsAlertSilencePolicy", AlibabaCloudSdkGoERROR)
	}

	// Set the resource ID
	d.SetId(fmt.Sprintf("%d", result.SilenceId))

	return resourceAliCloudArmsAlertSilencePolicyRead(d, meta)
}

func resourceAliCloudArmsAlertSilencePolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}
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

func resourceAliCloudArmsAlertSilencePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Check if any changes need to be applied
	update := false
	policy := &aliyunArmsAPI.AlertSilencePolicy{}

	if d.HasChange("silence_name") {
		update = true
	}
	policy.SilenceName = d.Get("silence_name").(string)

	if d.HasChange("state") {
		update = true
	}
	if v, ok := d.GetOk("state"); ok {
		policy.State = v.(string)
	}

	if d.HasChange("effective_time_type") {
		update = true
	}
	policy.EffectiveTimeType = d.Get("effective_time_type").(string)

	if d.HasChange("time_period") {
		update = true
	}
	if v, ok := d.GetOk("time_period"); ok {
		policy.TimePeriod = v.(string)
	}

	if d.HasChange("time_slots") {
		update = true
	}
	if v, ok := d.GetOk("time_slots"); ok {
		policy.TimeSlots = v.(string)
	}

	if d.HasChange("start_time") {
		update = true
	}
	if v, ok := d.GetOk("start_time"); ok {
		policy.StartTime = v.(string)
	}

	if d.HasChange("end_time") {
		update = true
	}
	if v, ok := d.GetOk("end_time"); ok {
		policy.EndTime = v.(string)
	}

	if d.HasChange("comment") {
		update = true
	}
	if v, ok := d.GetOk("comment"); ok {
		policy.Comment = v.(string)
	}

	if d.HasChange("matching_rules") {
		update = true
		// TODO: Parse matching_rules from string to proper structure
		if v, ok := d.GetOk("matching_rules"); ok {
			_ = v.(string)
			// policy.MatchingRules = parseMatchingRules(v.(string))
		}
	}

	if update {
		// Parse silence ID from resource ID
		silenceId, err := strconv.ParseInt(d.Id(), 10, 64)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ParseInt", AlibabaCloudSdkGoERROR)
		}

		// Update using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err = armsService.UpdateArmsAlertSilencePolicy(silenceId, policy)
			if err != nil {
				if NeedRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateArmsAlertSilencePolicy", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudArmsAlertSilencePolicyRead(d, meta)
}

func resourceAliCloudArmsAlertSilencePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse silence ID from resource ID
	silenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ParseInt", AlibabaCloudSdkGoERROR)
	}

	// Delete using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = armsService.DeleteArmsAlertSilencePolicy(silenceId)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteArmsAlertSilencePolicy", AlibabaCloudSdkGoERROR)
	}

	return nil
}
