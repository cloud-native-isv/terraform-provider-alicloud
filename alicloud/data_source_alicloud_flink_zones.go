package alicloud

import (
	"fmt"
	"sort"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
)

func dataSourceAlicloudFlinkZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkZonesRead,

		Schema: map[string]*schema.Schema{
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
	response, err := flinkService.DescribeSupportedZones(request)
	if err != nil {
		return WrapError(err)
	}

	var zoneIds []string
	var zones []map[string]interface{}
	for _, zone := range response.Body.Zones {
		zoneIds = append(zoneIds, zone.ZoneId)
		zones = append(zones, map[string]interface{}{
			"id": zone.ZoneId,
		})
	}

	sort.Strings(zoneIds)

	d.SetId(dataResourceIdHash(zoneIds))
	if err := d.Set("zones", zones); err != nil {
		return WrapError(err)
	}
	if err := d.Set("ids", zoneIds); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), zones)
	}

	return nil
}