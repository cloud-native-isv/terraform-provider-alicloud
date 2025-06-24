package alicloud

import (
	"fmt"
	"time"

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
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
						"jar_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
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
						"lookup": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"supported_formats": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"dependencies": {
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

func dataSourceAlicloudFlinkConnectorsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	// Get all connectors
	connectors, err := flinkService.ListCustomConnectors(workspaceId, namespaceName)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_connectors", "ListCustomConnectors", AlibabaCloudSdkGoERROR)
	}

	// Filter results if ids or names are provided
	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	namesMap := make(map[string]string)
	if v, ok := d.GetOk("names"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			namesMap[vv.(string)] = vv.(string)
		}
	}

	var connectorMaps []map[string]interface{}
	var filteredIds []string
	var filteredNames []string

	for _, connector := range connectors {
		connectorId := fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, connector.Name)

		// Apply filters
		if len(idsMap) > 0 {
			if _, ok := idsMap[connectorId]; !ok {
				continue
			}
		}

		if len(namesMap) > 0 {
			if _, ok := namesMap[connector.Name]; !ok {
				continue
			}
		}

		connectorMap := map[string]interface{}{
			"id":                connectorId,
			"name":              connector.Name,
			"type":              connector.Type,
			"jar_url":           connector.JarUrl,
			"description":       connector.Description,
			"source":            connector.Source,
			"sink":              connector.Sink,
			"lookup":            connector.Lookup,
			"supported_formats": connector.SupportedFormats,
			"dependencies":      connector.Dependencies,
		}

		connectorMaps = append(connectorMaps, connectorMap)
		filteredIds = append(filteredIds, connectorId)
		filteredNames = append(filteredNames, connector.Name)
	}

	d.SetId(fmt.Sprintf("%s:%s:%d", workspaceId, namespaceName, time.Now().Unix()))

	if err := d.Set("ids", filteredIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", filteredNames); err != nil {
		return WrapError(err)
	}
	if err := d.Set("connectors", connectorMaps); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), connectorMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
