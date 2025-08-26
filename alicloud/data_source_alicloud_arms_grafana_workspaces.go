package alicloud

import (
	"regexp"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsGrafanaWorkspaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsGrafanaWorkspacesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"aliyun_lang": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "zh",
				ValidateFunc: validation.StringInSlice([]string{"zh", "en"}, false),
				Description:  "The language of the query. Valid values: zh (Chinese), en (English). Default: zh.",
			},
			"tags": tagsSchemaForceNew(),
			"enable_details": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Grafana workspace.",
						},
						"grafana_workspace_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Grafana workspace.",
						},
						"grafana_workspace_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Grafana workspace.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the Grafana workspace.",
						},
						"grafana_workspace_edition": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The edition of the Grafana workspace.",
						},
						"grafana_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The version of Grafana.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the Grafana workspace.",
						},
						"grafana_workspace_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP address of the Grafana workspace.",
						},
						"grafana_workspace_domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The domain of the Grafana workspace.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time.",
						},
						"commercial": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether it's a commercial workspace.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The protocol used by the workspace.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The tags of the Grafana workspace.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsGrafanaWorkspacesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get parameters
	resourceGroupId := ""
	if v, ok := d.GetOk("resource_group_id"); ok {
		resourceGroupId = v.(string)
	}

	aliyunLang := d.Get("aliyun_lang").(string)

	enableDetails := d.Get("enable_details").(bool)

	// Get all workspaces using Service layer
	var workspaces []*aliyunArmsAPI.GrafanaWorkspace
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		result, err := service.ListGrafanaWorkspace(resourceGroupId, aliyunLang)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		workspaces = result
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_grafana_workspaces", "ListGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}

	// Filter by name regex
	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		nameRegex = r
	}

	// Filter by IDs
	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	// Apply filters and handle details
	var basicWorkspaces []*aliyunArmsAPI.GrafanaWorkspace
	var detailedWorkspaces []*aliyunArmsAPI.GrafanaWorkspaceDetail

	for _, workspace := range workspaces {
		// Filter by name regex
		if nameRegex != nil && !nameRegex.MatchString(tea.StringValue(workspace.GrafanaWorkspaceName)) {
			continue
		}

		// Filter by IDs
		if len(idsMap) > 0 {
			if _, ok := idsMap[tea.StringValue(workspace.GrafanaWorkspaceId)]; !ok {
				continue
			}
		}

		// If enable_details is true, get detailed information for each workspace
		if enableDetails {
			detailedWorkspace, err := service.DescribeArmsGrafanaWorkspace(tea.StringValue(workspace.GrafanaWorkspaceId))
			if err != nil {
				if IsNotFoundError(err) {
					// Skip if workspace is deleted during processing
					continue
				}
				return WrapErrorf(err, DefaultErrorMsg, tea.StringValue(workspace.GrafanaWorkspaceId), "DescribeArmsGrafanaWorkspace", AlibabaCloudSdkGoERROR)
			}
			detailedWorkspaces = append(detailedWorkspaces, detailedWorkspace)
		} else {
			basicWorkspaces = append(basicWorkspaces, workspace)
		}
	}

	// Build result
	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	if enableDetails {
		for _, workspace := range detailedWorkspaces {
			mapping := map[string]interface{}{
				"id":                        workspace.GrafanaWorkspaceId,
				"grafana_workspace_id":      workspace.GrafanaWorkspaceId,
				"grafana_workspace_name":    workspace.GrafanaWorkspaceName,
				"description":               workspace.Description,
				"grafana_workspace_edition": workspace.GrafanaWorkspaceEdition,
				"grafana_version":           workspace.GrafanaVersion,
				"status":                    workspace.Status,
				"grafana_workspace_ip":      workspace.GrafanaWorkspaceIp,
				"grafana_workspace_domain":  workspace.GrafanaWorkspaceDomain,
				"region_id":                 workspace.RegionId,
				"resource_group_id":         workspace.ResourceGroupId,
				"create_time":               timeToString(workspace.GmtCreate),
				"commercial":                workspace.Commercial,
				"protocol":                  workspace.Protocol,
			}

			// Convert tags to map[string]interface{}
			tagsMap := make(map[string]interface{})
			if workspace.Tags != nil {
				for _, tag := range workspace.Tags {
					tagsMap[tea.StringValue(tag.Key)] = tea.StringValue(tag.Value)
				}
			}
			mapping["tags"] = tagsMap

			ids = append(ids, tea.StringValue(workspace.GrafanaWorkspaceId))
			names = append(names, workspace.GrafanaWorkspaceName)
			s = append(s, mapping)
		}
	} else {
		for _, workspace := range basicWorkspaces {
			mapping := map[string]interface{}{
				"id":                        workspace.GrafanaWorkspaceId,
				"grafana_workspace_id":      workspace.GrafanaWorkspaceId,
				"grafana_workspace_name":    workspace.GrafanaWorkspaceName,
				"description":               workspace.Description,
				"grafana_workspace_edition": workspace.GrafanaWorkspaceEdition,
				"grafana_version":           workspace.GrafanaVersion,
				"status":                    workspace.Status,
				"region_id":                 workspace.RegionId,
				"resource_group_id":         workspace.ResourceGroupId,
				"create_time":               timeToString(workspace.GmtCreate),
			}

			ids = append(ids, tea.StringValue(workspace.GrafanaWorkspaceId))
			names = append(names, workspace.GrafanaWorkspaceName)
			s = append(s, mapping)
		}
	}

	d.SetId(dataResourceIdHash(ids))

	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("workspaces", s); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}

func timeToString(t *int64) string {
	if t == nil {
		return ""
	}
	return time.Unix(*t/1000, 0).Format(time.RFC3339)
}
