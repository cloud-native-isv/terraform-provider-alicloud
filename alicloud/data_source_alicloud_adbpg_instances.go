package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgInstancesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"description_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"db_instance_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_network_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vswitch_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"pay_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	query := &adbpg.AdbpgInstanceQuery{}

	if v, ok := d.GetOk("resource_group_id"); ok {
		query.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("tags"); ok {
		tagsMap := v.(map[string]interface{})
		adbpgTags := make([]adbpg.Tag, 0, len(tagsMap))
		for k, val := range tagsMap {
			adbpgTags = append(adbpgTags, adbpg.Tag{Key: k, Value: val.(string)})
		}
		query.Tags = adbpgTags
	}

	allInstances, err := adbpgService.ListAdbpgInstances(query)
	if err != nil {
		return WrapError(err)
	}

	var filteredInstances []adbpg.AdbpgInstance
	idsMap := make(map[string]bool)
	if v, ok := d.GetOk("ids"); ok {
		for _, id := range v.([]interface{}) {
			idsMap[id.(string)] = true
		}
	}

	var descRegex *regexp.Regexp
	if v, ok := d.GetOk("description_regex"); ok {
		descRegex = regexp.MustCompile(v.(string))
	}

	statusFilter := ""
	if v, ok := d.GetOk("status"); ok {
		statusFilter = v.(string)
	}

	for _, inst := range allInstances {
		if len(idsMap) > 0 && !idsMap[inst.DBInstanceId] {
			continue
		}
		if descRegex != nil && !descRegex.MatchString(inst.DBInstanceDescription) {
			continue
		}
		if statusFilter != "" && inst.DBInstanceStatus != statusFilter {
			continue
		}
		filteredInstances = append(filteredInstances, inst)
	}

	ids := make([]string, 0, len(filteredInstances))
	s := make([]map[string]interface{}, 0, len(filteredInstances))

	for _, inst := range filteredInstances {
		mapping := map[string]interface{}{
			"id":                    inst.DBInstanceId,
			"description":           inst.DBInstanceDescription,
			"status":                inst.DBInstanceStatus,
			"engine":                inst.Engine,
			"engine_version":        inst.EngineVersion,
			"db_instance_mode":      inst.DBInstanceMode,
			"instance_network_type": inst.InstanceNetworkType,
			"region_id":             inst.RegionId,
			"zone_id":               inst.ZoneId,
			"vpc_id":                inst.VpcId,
			"vswitch_id":            inst.VSwitchId,
			"pay_type":              inst.PayType,
			"creation_time":         inst.CreateTime,
			"resource_group_id":     inst.ResourceGroupId,
		}
		ids = append(ids, inst.DBInstanceId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("ids", ids)
	d.Set("instances", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
