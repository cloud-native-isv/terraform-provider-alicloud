package alicloud

import (
	"sort"

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

	response, err := flinkService.DescribeSupportedZones()
	if err != nil {
		return WrapError(err)
	}

	var ids []string
	var zoneList []map[string]interface{}

	if response != nil {
		// Extract zone IDs and sort them
		var zoneIds []string
		for _, zone := range response {
			if zone != nil && zone.ZoneID != "" {
				zoneIds = append(zoneIds, zone.ZoneID)
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
