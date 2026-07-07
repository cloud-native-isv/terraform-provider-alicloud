package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgDataSources() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgDataSourcesRead,
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_source_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_source_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_source_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgDataSourcesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allSources, err := adbpgService.ListAdbpgDataSources(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var ids []string
	s := make([]map[string]interface{}, 0)

	for _, ds := range allSources {
		if nameRegex != nil && !nameRegex.MatchString(ds.DataSourceName) {
			continue
		}
		mapping := map[string]interface{}{
			"data_source_id":   ds.DataSourceId,
			"data_source_name": ds.DataSourceName,
			"data_source_type": ds.DataSourceType,
			"status":           ds.Status,
			"create_time":      ds.CreateTime,
		}
		ids = append(ids, ds.DataSourceId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("sources", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
