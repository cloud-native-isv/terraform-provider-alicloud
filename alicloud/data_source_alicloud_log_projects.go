package alicloud

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
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
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	request := map[string]interface{}{}
	if v, ok := d.GetOk("name_regex"); ok {
		request["projectName"] = v.(string)
	}

	var allProjects []map[string]interface{}

	var projectNames []string
	if v, ok := d.GetOk("ids"); ok {
		for _, item := range v.([]interface{}) {
			if item != nil {
				projectNames = append(projectNames, strings.Trim(item.(string), `"`))
			}
		}
	}

	if len(projectNames) > 0 {
		// Get specific projects by names
		for _, name := range projectNames {
			project, err := slsService.DescribeLogProject(name)
			if err != nil {
				if NotFoundError(err) {
					continue
				}
				return WrapError(err)
			}
			// Convert LogProject to map[string]interface{}
			projectMap := convertLogProjectToMap(project)
			allProjects = append(allProjects, projectMap)
		}
	} else {
		// List all projects
		projects, err := slsService.ListProjects()
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_projects", "ListProjects", AliyunLogGoSdkERROR)
		}

		for _, project := range projects {
			projectMap := convertLogProjectToMap(project)
			allProjects = append(allProjects, projectMap)
		}
	}

	var filteredProjects []map[string]interface{}
	nameRegex, ok := d.GetOk("name_regex")
	if ok && nameRegex.(string) != "" {
		r, err := regexp.Compile(nameRegex.(string))
		if err != nil {
			return WrapError(err)
		}
		for _, project := range allProjects {
			if projectName, exists := project["project_name"]; exists {
				if r.MatchString(projectName.(string)) {
					filteredProjects = append(filteredProjects, project)
				}
			}
		}
	} else {
		filteredProjects = allProjects
	}

	return logProjectsDecriptionAttributes(d, filteredProjects, meta)
}

// convertLogProjectToMap converts aliyunSlsAPI.LogProject to map[string]interface{} for compatibility
func convertLogProjectToMap(project *aliyunSlsAPI.LogProject) map[string]interface{} {
	return map[string]interface{}{
		"id":               project.ProjectName,
		"description":      project.Description,
		"last_modify_time": project.LastModifyTime,
		"owner":            project.Owner,
		"project_name":     project.ProjectName,
		"region":           project.Region,
		"status":           project.Status,
	}
}

func logProjectsDecriptionAttributes(d *schema.ResourceData, projects []map[string]interface{}, meta interface{}) error {
	var ids []string
	var names []interface{}
	var s []map[string]interface{}

	for _, project := range projects {
		mapping := map[string]interface{}{
			"id":               project["id"],
			"description":      project["description"],
			"last_modify_time": project["last_modify_time"],
			"owner":            project["owner"],
			"project_name":     project["project_name"],
			"region":           project["region"],
			"status":           project["status"],
		}

		// Handle policy if exists
		if policy, ok := project["policy"]; ok {
			mapping["policy"] = policy
		}

		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, project["project_name"])
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
