package alicloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudLogProjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudLogProjectsRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "Disable"}, true),
				ForceNew:     true,
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_modify_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudLogProjectsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService := NewLogService(client)

	var objects []map[string]interface{}
	var logProjectNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		logProjectNameRegex = r
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}
	status, statusOk := d.GetOk("status")
	var response []string
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		raw, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			return slsClient.ListLogProjects(ctx, "", "")
		})
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if projects, ok := raw.([]*aliyunSlsAPI.LogProject); ok {
			response = make([]string, len(projects))
			for i, project := range projects {
				response[i] = project.ProjectName
			}
		}
		return nil
	})
	addDebug("ListProject", response)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_projects", "ListProject", AliyunLogGoSdkERROR)
	}

	for _, projectName := range response {
		if logProjectNameRegex != nil {
			if !logProjectNameRegex.MatchString(projectName) {
				continue
			}
		}
		if len(idsMap) > 0 {
			if _, ok := idsMap[projectName]; !ok {
				continue
			}
		}

		project, err := logService.DescribeLogProject(projectName)
		if err != nil {
			if NotFoundError(err) {
				log.Printf("The project '%s' is no exist! \n", projectName)
				continue
			}
			return WrapError(err)
		}

		// Use jsonpath to extract status for compatibility
		if statusOk && status.(string) != "" {
			if projectStatus, exists := project["status"]; exists && status.(string) != projectStatus.(string) {
				continue
			}
		}

		objects = append(objects, project)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		mapping := map[string]interface{}{
			"id":               object["projectName"],
			"description":      object["description"],
			"last_modify_time": object["lastModifyTime"],
			"owner":            object["owner"],
			"project_name":     object["projectName"],
			"region":           object["region"],
			"status":           object["status"],
		}

		// Get project policy if exists
		if policy, exists := object["policy"]; exists && policy != nil {
			mapping["policy"] = policy
		}

		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, object["projectName"])
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("projects", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
