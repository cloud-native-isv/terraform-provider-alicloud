package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	armsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudArmsAlertEvents() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudArmsAlertEventsRead,
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
			"alert_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"severity": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"integration_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alert_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"severity": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"image_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"check": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"class": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"generator_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"receive_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"annotations": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudArmsAlertEventsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client with credentials
	credentials := &armsAPI.ArmsCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsClient, err := armsAPI.NewArmsAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_events", "InitializeArmsAPI", "Failed to initialize ARMS API client")
	}

	// Build filter parameters from input
	var stateFilter *int64
	if v, ok := d.GetOk("state"); ok {
		stateStr := v.(string)
		switch stateStr {
		case "ALERT":
			state := armsAPI.AlertStatePending
			stateFilter = &state
		case "OK":
			state := armsAPI.AlertStateResolved
			stateFilter = &state
		case "SILENCE":
			state := armsAPI.AlertStateProcessing
			stateFilter = &state
		}
	}

	severity := ""
	if v, ok := d.GetOk("severity"); ok {
		severity = v.(string)
	}

	alertName := ""
	if v, ok := d.GetOk("alert_name"); ok {
		alertName = v.(string)
	}

	integrationName := ""
	if v, ok := d.GetOk("integration_name"); ok {
		integrationName = v.(string)
	}

	startTime := ""
	if v, ok := d.GetOk("start_time"); ok {
		startTime = v.(string)
	}

	endTime := ""
	if v, ok := d.GetOk("end_time"); ok {
		endTime = v.(string)
	}

	var alertNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		alertNameRegex = r
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

	// Call ARMS API to list alerts with events
	var allEvents []map[string]interface{}
	page := int64(1)
	size := int64(100) // Use reasonable page size

	for {
		// List alerts with events enabled
		alerts, totalCount, err := armsClient.ListAlertHistory(page, size, true, false, stateFilter, severity, alertName, startTime, endTime, integrationName, "", nil)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_events", "ListAlertHistory", "Failed to list alert events")
		}

		// Extract events from alerts
		for _, alert := range alerts {
			if alert.AlertEvents != nil {
				for _, event := range alert.AlertEvents {
					// Apply name regex filter if specified
					if alertNameRegex != nil && !alertNameRegex.MatchString(event.AlertName) {
						continue
					}

					// Create event ID from combination of fields for uniqueness
					eventId := fmt.Sprintf("%s_%s_%s", event.AlertName, event.IntegrationName, event.StartTime)

					// Apply ID filter if specified
					if len(idsMap) > 0 {
						if _, ok := idsMap[eventId]; !ok {
							continue
						}
					}

					// Convert API event to map format for schema compatibility
					eventMap := map[string]interface{}{
						"EventId":         eventId,
						"AlertName":       event.AlertName,
						"Severity":        event.Severity,
						"State":           event.State,
						"Message":         event.Message,
						"Value":           event.Value,
						"ImageUrl":        event.ImageUrl,
						"Check":           event.Check,
						"Source":          event.Source,
						"Class":           event.Class,
						"Service":         event.Service,
						"StartTime":       event.StartTime,
						"EndTime":         event.EndTime,
						"GeneratorURL":    event.GeneratorURL,
						"ReceiveTime":     event.ReceiveTime,
						"IntegrationName": event.IntegrationName,
						"IntegrationType": event.IntegrationType,
						"Description":     event.Description,
						"Annotations":     event.Annotations,
						"Labels":          event.Labels,
					}
					allEvents = append(allEvents, eventMap)
				}
			}
		}

		// Check if we've retrieved all results
		if totalCount <= int64(len(alerts)) || len(alerts) < int(size) {
			break
		}
		page++
	}

	// Build output data
	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, eventMap := range allEvents {
		mapping := map[string]interface{}{
			"id":               fmt.Sprint(eventMap["EventId"]),
			"alert_name":       eventMap["AlertName"],
			"severity":         eventMap["Severity"],
			"state":            eventMap["State"],
			"message":          eventMap["Message"],
			"value":            eventMap["Value"],
			"image_url":        eventMap["ImageUrl"],
			"check":            eventMap["Check"],
			"source":           eventMap["Source"],
			"class":            eventMap["Class"],
			"service":          eventMap["Service"],
			"start_time":       fmt.Sprint(eventMap["StartTime"]),
			"end_time":         fmt.Sprint(eventMap["EndTime"]),
			"generator_url":    eventMap["GeneratorURL"],
			"receive_time":     fmt.Sprint(eventMap["ReceiveTime"]),
			"integration_name": eventMap["IntegrationName"],
			"integration_type": eventMap["IntegrationType"],
			"description":      eventMap["Description"],
			"annotations":      eventMap["Annotations"],
			"labels":           eventMap["Labels"],
		}
		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, eventMap["AlertName"])
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("events", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
