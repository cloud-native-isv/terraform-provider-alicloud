package alicloud

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudArmsAlertActivities() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudArmsAlertActivitiesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"alert_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"activity_type": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"handler_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"activities": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alert_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"handler_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"content": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudArmsAlertActivitiesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &arms.ArmsCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_activities", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	var objects []arms.AlertActivity
	var handlerNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		handlerNameRegex = r
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	// Get parameters
	alertId := d.Get("alert_id").(string)
	alertIdInt, err := strconv.ParseInt(alertId, 10, 64)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_activities", "ParseInt", AlibabaCloudSdkGoERROR)
	}

	var activityType *int64
	if v, ok := d.GetOk("activity_type"); ok {
		activityTypeVal := int64(v.(int))
		activityType = &activityTypeVal
	}

	handlerName := ""
	if v, ok := d.GetOk("handler_name"); ok {
		handlerName = v.(string)
	}

	// Get activities using ARMS API
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		activities, err := armsAPI.GetAlertActivities(alertIdInt)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		// Filter results
		for _, activity := range activities {
			// Apply activity type filter
			if activityType != nil && activity.Type != *activityType {
				continue
			}

			// Apply handler name filter
			if handlerName != "" && activity.HandlerName != handlerName {
				continue
			}

			// Apply name regex filter
			if handlerNameRegex != nil && !handlerNameRegex.MatchString(activity.HandlerName) {
				continue
			}

			// Apply IDs filter
			activityId := fmt.Sprintf("%s_%s_%d", alertId, activity.Time, activity.Type)
			if len(idsMap) > 0 {
				if _, ok := idsMap[activityId]; !ok {
					continue
				}
			}

			objects = append(objects, activity)
		}

		return nil
	})

	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_activities", "GetAlertActivities", AlibabaCloudSdkGoERROR)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, object := range objects {
		activityId := fmt.Sprintf("%s_%s_%d", alertId, object.Time, object.Type)
		mapping := map[string]interface{}{
			"id":           activityId,
			"alert_id":     alertId,
			"time":         object.Time,
			"type":         int(object.Type),
			"handler_name": object.HandlerName,
			"description":  object.Description,
			"content":      object.Content,
		}

		ids = append(ids, activityId)
		names = append(names, object.HandlerName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("activities", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
