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

func dataSourceAliCloudArmsAlertItems() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertItemsRead,
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
			"alert_type": {
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
			"items": {
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
						"alert_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rule_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
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
						"is_enable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"contact_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"webhook": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ding_robot_webhook": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"wechat_robot_webhook": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slack_webhook": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"feishu_robot_webhook": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertItemsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS service
	service, err := NewArmsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_items", "NewArmsService", AlibabaCloudSdkGoERROR)
	}

	var objects []*aliyunArmsAPI.AlertItem
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

	// Get filter parameters
	alertName := ""
	if v, ok := d.GetOk("alert_name"); ok {
		alertName = v.(string)
	}

	alertType := ""
	if v, ok := d.GetOk("alert_type"); ok {
		alertType = v.(string)
	}

	status := ""
	if v, ok := d.GetOk("status"); ok {
		status = v.(string)
	}

	// Get all alert items using service layer
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		items, err := service.DescribeArmsAllAlertItems()
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		// Filter results
		for _, item := range items {
			// Apply alert name filter
			if alertName != "" && item.AlertName != alertName {
				continue
			}

			// Apply alert type filter
			if alertType != "" && item.AlertType != alertType {
				continue
			}

			// Apply status filter
			if status != "" && strconv.FormatInt(item.Status, 10) != status {
				continue
			}

			// Apply name regex filter
			if alertNameRegex != nil && !alertNameRegex.MatchString(item.AlertName) {
				continue
			}

			// Apply IDs filter
			itemId := EncodeArmsAlertItemId(item.AlertId)
			if len(idsMap) > 0 {
				if _, ok := idsMap[itemId]; !ok {
					continue
				}
			}

			objects = append(objects, item)
		}

		return nil
	})

	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_items", "DescribeArmsAllAlertItems", AlibabaCloudSdkGoERROR)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, object := range objects {
		itemId := EncodeArmsAlertItemId(object.AlertId)

		mapping := map[string]interface{}{
			"id":                   itemId,
			"alert_id":             strconv.FormatInt(object.AlertId, 10),
			"alert_name":           object.AlertName,
			"alert_type":           object.AlertType,
			"status":               strconv.FormatInt(object.Status, 10),
			"rule_id":              "", // Field not available in AlertItem
			"rule_name":            "", // Field not available in AlertItem
			"cluster_id":           object.ClusterId,
			"cluster_name":         "", // Field not available in AlertItem
			"region_id":            object.RegionId,
			"create_time":          object.CreateTime,
			"update_time":          object.UpdateTime,
			"is_enable":            false,           // Field not available in AlertItem
			"level":                object.Severity, // Use Severity as level
			"contact_groups":       []string{},      // Field not available in AlertItem
			"webhook":              "",              // Field not available in AlertItem
			"ding_robot_webhook":   "",              // Field not available in AlertItem
			"wechat_robot_webhook": "",              // Field not available in AlertItem
			"slack_webhook":        "",              // Field not available in AlertItem
			"feishu_robot_webhook": "",              // Field not available in AlertItem
		}

		ids = append(ids, itemId)
		names = append(names, object.AlertName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("items", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
