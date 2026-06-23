package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudAdbpgZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgZonesRead,
		Schema: map[string]*schema.Schema{
			"multi": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
						"multi_zone_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgZonesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	resources, err := adbpgService.ListAdbpgAvailableResources(client.RegionId)
	if err != nil {
		return WrapError(err)
	}

	multi := d.Get("multi").(bool)

	var ids []string
	var zones []map[string]interface{}

	for _, res := range resources {
		if multi && len(res.ZoneId) <= 1 {
			continue
		}

		zone := map[string]interface{}{
			"id":             res.ZoneId,
			"multi_zone_ids": []string{res.ZoneId},
		}
		ids = append(ids, res.ZoneId)
		zones = append(zones, zone)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("ids", ids)
	d.Set("zones", zones)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), zones)
	}

	return nil
}
