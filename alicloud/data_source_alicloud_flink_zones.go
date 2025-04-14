package alicloud

import (
	"sort"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudFlinkZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkZonesRead,

		Schema: map[string]*schema.Schema{
			"architecture_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The architecture type of Flink instances.",
				ValidateFunc: validation.StringInSlice([]string{"X86", "ARM"}, true),
				ForceNew:     true,
				Default:      "X86",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region to query available zones from, e.g. cn-beijing.",
				ForceNew:    true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFlinkZonesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	request := &foasconsole.DescribeSupportedZonesRequest{}

	// Set request parameters from schema if provided
	if v, ok := d.GetOk("architecture_type"); ok {
		architectureType := v.(string)
		request.ArchitectureType = &architectureType
	}

	if v, ok := d.GetOk("region"); ok {
		region := v.(string)
		request.Region = &region
	}

	response, err := flinkService.DescribeSupportedZones(request)
	if err != nil {
		return WrapError(err)
	}

	var ids []string
	var zoneList []map[string]interface{}

	if response.Body.Success != nil && *response.Body.Success && response.Body.ZoneIds != nil {
		// 提取并排序可用区ID
		var zoneIds []string
		for _, zonePtr := range response.Body.ZoneIds {
			if zonePtr != nil {
				zoneIds = append(zoneIds, *zonePtr)
			}
		}
		sort.Strings(zoneIds)

		for _, zoneId := range zoneIds {
			ids = append(ids, zoneId)
			zone := map[string]interface{}{
				"id": zoneId,
			}
			zoneList = append(zoneList, zone)
		}
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}
	if err := d.Set("zones", zoneList); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), zoneList)
	}

	return nil
}
