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
				"architecture_type": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The architecture type of Flink instances.",
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The region to query available zones from, e.g. cn-beijing.",
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
		for _, zonePtr := range response.Body.ZoneIds {
			if zonePtr != nil {
				ids = append(ids, *zonePtr)
				
				zone := map[string]interface{}{
					"id": *zonePtr,
				}
				zoneList = append(zoneList, zone)
			}
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