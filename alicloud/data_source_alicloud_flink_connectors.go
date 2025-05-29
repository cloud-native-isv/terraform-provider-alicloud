package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// dataSourceAlicloudFlinkConnectors provides the data source implementation for Alibaba Cloud Flink connectors
func dataSourceAlicloudFlinkConnectors() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkConnectorsRead,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace_name": {
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
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"connectors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"properties": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"dependencies": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"supported_formats": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"lookup": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"source": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sink": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"creator": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creator_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modifier_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudFlinkConnectorsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	response, err := flinkService.ListCustomConnectors(workspaceId, namespaceName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connectors", "ListCustomConnectors", AliyunLogGoSdkERROR)
	}

	if response == nil {
		return WrapErrorf(fmt.Errorf("empty response"), DefaultErrorMsg, "alicloud_flink_connectors", "ListCustomConnectors", AliyunLogGoSdkERROR)
	}

	var filteredConnectorIds []string
	var filteredConnectorNames []string
	var filteredConnectors []map[string]interface{}

	nameRegex, nameRegexOk := d.GetOk("name_regex")
	var r *regexp.Regexp
	if nameRegexOk {
		r = regexp.MustCompile(nameRegex.(string))
	}

	connectorType := d.Get("type").(string)

	for _, connector := range response {
		if connector.Name == "" {
			continue
		}

		// Apply name_regex filter if set
		if nameRegexOk && !r.MatchString(connector.Name) {
			continue
		}

		// Apply type filter if set
		if connectorType != "" && connector.Type != connectorType {
			continue
		}

		mapping := map[string]interface{}{
			"id":   fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, connector.Name),
			"name": connector.Name,
			"type": connector.Type,
		}

		if connector.Creator != "" {
			mapping["creator"] = connector.Creator
		}

		if connector.Description != "" {
			mapping["description"] = connector.Description
		}

		// Handle dependencies
		if connector.Dependencies != nil && len(connector.Dependencies) > 0 {
			mapping["dependencies"] = connector.Dependencies
		}

		filteredConnectorIds = append(filteredConnectorIds, mapping["id"].(string))
		filteredConnectorNames = append(filteredConnectorNames, mapping["name"].(string))
		filteredConnectors = append(filteredConnectors, mapping)
	}

	d.SetId(dataResourceIdHash(filteredConnectorIds))
	if err := d.Set("ids", filteredConnectorIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", filteredConnectorNames); err != nil {
		return WrapError(err)
	}
	if err := d.Set("connectors", filteredConnectors); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), filteredConnectors)
	}

	return nil
}
