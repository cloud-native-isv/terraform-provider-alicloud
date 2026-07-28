package alicloud

import (
	"sort"
	"strings"

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

	availableZones, err := adbpgService.ListAdbpgZones(client.RegionId)
	if err != nil {
		return WrapError(err)
	}

	multi := d.Get("multi").(bool)
	var zoneIds []string
	for _, zoneId := range availableZones {
		if multi && strings.Contains(zoneId, MULTI_IZ_SYMBOL) {
			zoneIds = append(zoneIds, zoneId)
			continue
		}
		if !multi && !strings.Contains(zoneId, MULTI_IZ_SYMBOL) {
			zoneIds = append(zoneIds, zoneId)
		}
	}
	if len(zoneIds) > 0 {
		sort.Strings(zoneIds)
	}

	var zones []map[string]interface{}
	if !multi {
		for _, zoneId := range zoneIds {
			zones = append(zones, map[string]interface{}{"id": zoneId})
		}
	} else {
		for _, zoneId := range zoneIds {
			zones = append(zones, map[string]interface{}{
				"id":             zoneId,
				"multi_zone_ids": splitMultiZoneId(zoneId),
			})
		}
	}

	d.SetId(dataResourceIdHash(zoneIds))
	if err := d.Set("ids", zoneIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("zones", zones); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), zones)
	}

	return nil
}
