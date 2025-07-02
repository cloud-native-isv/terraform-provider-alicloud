package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudFlinkZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkZonesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
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
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"deprecated": {
							Type:     schema.TypeBool,
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

	// Get all supported zones
	zones, err := flinkService.DescribeSupportedZones()
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_zones", "DescribeSupportedZones", AlibabaCloudSdkGoERROR)
	}

	// Filter results if ids are provided
	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	var zoneMaps []map[string]interface{}
	var filteredIds []string

	for _, zone := range zones {
		// Apply filters
		if len(idsMap) > 0 {
			if _, ok := idsMap[zone.ZoneID]; !ok {
				continue
			}
		}

		zoneMap := map[string]interface{}{
			"id":         zone.ZoneID,
			"zone_id":    zone.ZoneID,
			"zone_name":  zone.ZoneName,
			"deprecated": zone.Deprecated,
		}

		zoneMaps = append(zoneMaps, zoneMap)
		filteredIds = append(filteredIds, zone.ZoneID)
	}

	d.SetId(fmt.Sprintf("flink_zones_%d", time.Now().Unix()))

	if err := d.Set("ids", filteredIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("zones", zoneMaps); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), zoneMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
