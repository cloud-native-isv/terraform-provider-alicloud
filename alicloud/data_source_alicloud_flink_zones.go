package alicloud

import (
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
	
	var zones []string
	if *response.Body.Success && response.Body.ZoneIds != nil { // 使用ZoneIds字段
		for _, zonePtr := range response.Body.ZoneIds {
			if zonePtr != nil {
				zones = append(zones, *zonePtr)
			}
		}
	}
	d.Set("zones", zones)
	d.SetId(dataResourceIdHash(zones))
	return nil
}