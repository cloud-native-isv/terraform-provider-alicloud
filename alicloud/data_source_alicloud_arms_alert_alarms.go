package alicloud

import (
	"regexp"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// AlarmWithContext wraps AlertAlarm with context information
type AlarmWithContext struct {
	Alarm   *aliyunArmsAPI.AlertAlarm
	EventId string
	AlertId int64
}

func dataSourceAliCloudArmsAlertAlarms() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertAlarmsRead,
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
			"alarm_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alarms": {
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
						"alarm_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alarm_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alarm_content": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"send_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notify_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notify_object": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"event_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertAlarmsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS service
	service, err := NewArmsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_alarms", "NewArmsService", AlibabaCloudSdkGoERROR)
	}

	var objects []*AlarmWithContext
	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		nameRegex = r
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_alarms", "ParseInt", AlibabaCloudSdkGoERROR)
	}

	status := ""
	if v, ok := d.GetOk("status"); ok {
		status = v.(string)
	}

	// Get alarms using ARMS service layer
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		// First get all alert events for this alert ID to get the context
		filters := map[string]interface{}{
			"alertId": alertIdInt,
		}

		alertEvents, err := service.DescribeArmsAllAlertEvents(filters)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		// Extract alarms with context from events
		for _, event := range alertEvents {
			if event.AlertId == alertIdInt && event.Alarms != nil {
				for _, alarm := range event.Alarms {
					// Apply status filter
					if status != "" && alarm.Status != status {
						continue
					}

					// Apply name regex filter (using alarm content)
					if nameRegex != nil && !nameRegex.MatchString(alarm.AlarmContent) {
						continue
					}

					// Apply IDs filter - construct ID from event context
					alarmId := EncodeArmsAlertAlarmId(event.AlertId, event.EventId, alarm.AlarmId)
					if len(idsMap) > 0 {
						if _, ok := idsMap[alarmId]; !ok {
							continue
						}
					}

					// Create alarm with context for proper ID encoding
					alarmWithContext := &AlarmWithContext{
						Alarm:   alarm,
						EventId: event.EventId,
						AlertId: event.AlertId,
					}
					objects = append(objects, alarmWithContext)
				}
			}
		}

		return nil
	})

	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_alarms", "DescribeArmsAlertAlarmsByAlertId", AlibabaCloudSdkGoERROR)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, alarmWithContext := range objects {
		alarm := alarmWithContext.Alarm
		alarmId := EncodeArmsAlertAlarmId(alarmWithContext.AlertId, alarmWithContext.EventId, alarm.AlarmId)

		mapping := map[string]interface{}{
			"id":            alarmId,
			"alert_id":      strconv.FormatInt(alarmWithContext.AlertId, 10),
			"alarm_id":      strconv.FormatInt(alarm.AlarmId, 10),
			"alarm_type":    alarm.ContactType, // Use ContactType as alarm type
			"status":        alarm.Status,
			"alarm_content": alarm.AlarmContent,
			"target":        alarm.ContactValue, // Use ContactValue as target
			"send_time":     alarm.SendTime,
			"create_time":   alarm.CreateTime,
			"notify_type":   alarm.ContactType,
			"notify_object": alarm.ContactName,
			"description":   alarm.Solution,
			"event_id":      alarmWithContext.EventId,
		}

		ids = append(ids, alarmId)
		names = append(names, alarm.ContactValue) // Use ContactValue as name
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("alarms", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
