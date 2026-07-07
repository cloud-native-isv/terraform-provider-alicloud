package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgResourceGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgResourceGroupsRead,
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
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_group_config": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_list": {
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

func dataSourceAliCloudAdbpgResourceGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allGroups, err := adbpgService.ListAdbpgResourceGroups(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var names []string
	s := make([]map[string]interface{}, 0)

	for _, g := range allGroups {
		if nameRegex != nil && !nameRegex.MatchString(g.ResourceGroupName) {
			continue
		}
		mapping := map[string]interface{}{
			"resource_group_name":   g.ResourceGroupName,
			"resource_group_config": g.ResourceGroupConfig,
			"status":                g.Status,
			"role_list":             g.RoleList,
		}
		names = append(names, g.ResourceGroupName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(names))
	d.Set("groups", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
