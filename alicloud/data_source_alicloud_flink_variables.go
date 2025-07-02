package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudFlinkVariables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkVariablesRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the Flink workspace",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Name of the Flink namespace",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"variables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the variable",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the variable",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The value of the variable",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the variable",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kind of the variable (Plain, Encrypted, Clear)",
						},
						"workspace_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The workspace ID where the variable belongs",
						},
						"namespace_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The namespace name where the variable belongs",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFlinkVariablesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	// Get all variables from the namespace
	variables, err := flinkService.ListVariables(workspaceId, namespaceName)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_variables", "ListVariables", AlibabaCloudSdkGoERROR)
	}

	// Filter by names if specified
	var filteredVariables []interface{}
	var names []string

	if nameList, ok := d.GetOk("names"); ok {
		targetNames := make(map[string]bool)
		for _, name := range nameList.([]interface{}) {
			targetNames[name.(string)] = true
		}

		for _, variable := range variables {
			if targetNames[variable.Name] {
				variableMap := map[string]interface{}{
					"id":             fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, variable.Name),
					"name":           variable.Name,
					"value":          variable.Value,
					"description":    variable.Description,
					"kind":           variable.Kind,
					"workspace_id":   workspaceId,
					"namespace_name": namespaceName,
				}
				filteredVariables = append(filteredVariables, variableMap)
				names = append(names, variable.Name)
			}
		}
	} else {
		// Return all variables if no filter specified
		for _, variable := range variables {
			variableMap := map[string]interface{}{
				"id":             fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, variable.Name),
				"name":           variable.Name,
				"value":          variable.Value,
				"description":    variable.Description,
				"kind":           variable.Kind,
				"workspace_id":   workspaceId,
				"namespace_name": namespaceName,
			}
			filteredVariables = append(filteredVariables, variableMap)
			names = append(names, variable.Name)
		}
	}

	d.SetId(dataResourceIdHash([]string{workspaceId, namespaceName}))

	if err := d.Set("variables", filteredVariables); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	return nil
}
