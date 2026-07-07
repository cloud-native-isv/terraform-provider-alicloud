package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgSupabaseProjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgSupabaseProjectsRead,
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
			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"plan": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
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

func dataSourceAliCloudAdbpgSupabaseProjectsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allProjects, err := adbpgService.ListAdbpgSupabaseProjects(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var ids []string
	s := make([]map[string]interface{}, 0)

	for _, p := range allProjects {
		if nameRegex != nil && !nameRegex.MatchString(p.ProjectName) {
			continue
		}
		mapping := map[string]interface{}{
			"project_id":   p.ProjectId,
			"project_name": p.ProjectName,
			"plan":         p.Plan,
			"region_id":    p.RegionId,
			"status":       p.Status,
			"create_time":  p.CreateTime,
		}
		ids = append(ids, p.ProjectId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("projects", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
