package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgVectorCollections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgVectorCollectionsRead,
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "public",
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
			"collections": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"collection_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dimension": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"document_count": {
							Type:     schema.TypeInt,
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

func dataSourceAliCloudAdbpgVectorCollectionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	namespace := d.Get("namespace").(string)
	allCollections, err := adbpgService.ListAdbpgVectorCollections(instanceId, namespace)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var names []string
	s := make([]map[string]interface{}, 0)

	for _, c := range allCollections {
		if nameRegex != nil && !nameRegex.MatchString(c.CollectionName) {
			continue
		}
		mapping := map[string]interface{}{
			"collection_name": c.CollectionName,
			"dimension":       int(c.Dimension),
			"status":          c.Status,
			"document_count":  int(c.DocumentCount),
			"create_time":     c.CreateTime,
		}
		names = append(names, c.CollectionName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(names))
	d.Set("collections", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
