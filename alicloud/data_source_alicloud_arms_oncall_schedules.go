package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsOncallSchedules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsOncallSchedulesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of OnCall Schedule IDs.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
				Description:  "A regex string to filter results by OnCall Schedule name.",
			},
			"names": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of OnCall Schedule names.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the OnCall Schedule.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			"schedules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of OnCall Schedules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the OnCall Schedule.",
						},
						"schedule_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the OnCall Schedule.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the OnCall Schedule.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the OnCall Schedule.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsOncallSchedulesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsOnCallScheduleService(client)
	if err != nil {
		return WrapError(err)
	}

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

	// Get name filter
	var nameFilter string
	if v, ok := d.GetOk("name"); ok {
		nameFilter = v.(string)
	}

	// List all on-call schedules with pagination
	page := int64(1)
	size := int64(100) // Use larger page size for data source
	var allSchedules []map[string]interface{}

	for {
		schedules, totalCount, err := service.DescribeArmsOnCallSchedules(page, size, nameFilter)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_oncall_schedules", "DescribeArmsOnCallSchedules", AlibabaCloudSdkGoERROR)
		}

		for _, schedule := range schedules {
			scheduleIdStr := fmt.Sprintf("%d", schedule.Id)

			// Filter by name regex
			if nameRegex != nil && !nameRegex.MatchString(schedule.Name) {
				continue
			}

			// Filter by IDs
			if len(idsMap) > 0 {
				if _, ok := idsMap[scheduleIdStr]; !ok {
					continue
				}
			}

			mapping := map[string]interface{}{
				"id":          scheduleIdStr,
				"schedule_id": schedule.Id,
				"name":        schedule.Name,
				"description": schedule.Description,
			}
			allSchedules = append(allSchedules, mapping)
		}

		// Check if we've retrieved all schedules
		if int64(len(schedules)) < size || totalCount <= page*size {
			break
		}
		page++
	}

	// Extract IDs and names for the computed attributes
	ids := make([]string, 0, len(allSchedules))
	names := make([]string, 0, len(allSchedules))

	for _, schedule := range allSchedules {
		ids = append(ids, schedule["id"].(string))
		names = append(names, schedule["name"].(string))
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("schedules", allSchedules); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), allSchedules)
	}

	return nil
}
