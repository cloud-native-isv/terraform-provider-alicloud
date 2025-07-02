package alicloud

import (
	"regexp"
	"strconv"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	armsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsAlertHistorys() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertHistorysRead,
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
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"dispatch_rule_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rules": {
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
						"alert_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"severity": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"dispatch_rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dispatch_rule_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"solution": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"handler": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"acknowledge_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"recover_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"describe": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertHistorysRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client with credentials
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsClient, err := armsAPI.NewArmsAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_rules", "InitializeArmsAPI", "Failed to initialize ARMS API client")
	}

	// Build filter parameters from input
	var stateFilter *int64
	if v, ok := d.GetOk("state"); ok {
		state := int64(v.(int))
		stateFilter = &state
	}

	severity := ""
	if v, ok := d.GetOk("severity"); ok {
		severity = v.(string)
	}

	alertName := ""
	if v, ok := d.GetOk("alert_name"); ok {
		alertName = v.(string)
	}

	var dispatchRuleId *int64
	if v, ok := d.GetOk("dispatch_rule_id"); ok {
		id, err := strconv.ParseInt(v.(string), 10, 64)
		if err == nil {
			dispatchRuleId = &id
		}
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

	// Call ARMS API to list alert rules
	var allAlerts []*armsAPI.AlertHistory
	page := int64(1)
	size := int64(100) // Use reasonable page size

	for {
		// List alerts without events and activities for better performance
		alerts, totalCount, err := armsClient.ListAlertHistory(page, size, false, false, stateFilter, severity, alertName, "", "", "", "", dispatchRuleId)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_rules", "ListAlertHistory", "Failed to list alert rules")
		}

		// Apply filters
		for _, alert := range alerts {
			// Apply name regex filter if specified
			if alertNameRegex != nil && !alertNameRegex.MatchString(alert.AlertName) {
				continue
			}

			// Apply ID filter if specified
			if len(idsMap) > 0 {
				alertIdStr := strconv.FormatInt(alert.AlertId, 10)
				if _, ok := idsMap[alertIdStr]; !ok {
					continue
				}
			}

			allAlerts = append(allAlerts, alert)
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

	for _, alert := range allAlerts {
		alertIdStr := strconv.FormatInt(alert.AlertId, 10)
		dispatchRuleIdStr := ""
		if alert.DispatchRuleId != 0 {
			dispatchRuleIdStr = strconv.FormatFloat(alert.DispatchRuleId, 'f', 0, 64)
		}

		mapping := map[string]interface{}{
			"id":                 alertIdStr,
			"alert_id":           alertIdStr,
			"alert_name":         alert.AlertName,
			"severity":           alert.Severity,
			"state":              int(alert.State),
			"dispatch_rule_id":   dispatchRuleIdStr,
			"dispatch_rule_name": alert.DispatchRuleName,
			"create_time":        alert.CreateTime,
			"solution":           alert.Solution,
			"owner":              alert.Owner,
			"handler":            alert.Handler,
			"acknowledge_time":   int(alert.AcknowledgeTime),
			"recover_time":       int(alert.RecoverTime),
			"describe":           alert.Describe,
		}

		ids = append(ids, alertIdStr)
		names = append(names, alert.AlertName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("rules", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
