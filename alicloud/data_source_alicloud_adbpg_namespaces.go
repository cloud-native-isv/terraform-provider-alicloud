package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgNamespaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgNamespacesRead,
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
			"namespaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
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

func dataSourceAliCloudAdbpgNamespacesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allNamespaces, err := adbpgService.ListAdbpgNamespaces(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var names []string
	s := make([]map[string]interface{}, 0)

	for _, ns := range allNamespaces {
		if nameRegex != nil && !nameRegex.MatchString(ns.NamespaceName) {
			continue
		}
		mapping := map[string]interface{}{
			"namespace_name": ns.NamespaceName,
			"description":    ns.Description,
			"create_time":    ns.CreateTime,
		}
		names = append(names, ns.NamespaceName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(names))
	d.Set("namespaces", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
