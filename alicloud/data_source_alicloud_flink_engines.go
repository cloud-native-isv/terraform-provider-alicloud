package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudFlinkEngines() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkEnginesRead,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The workspace ID of Flink.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter results by engine version name.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			"ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of engine version IDs.",
			},
			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of engine version names.",
			},
			"engines": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of Flink engine versions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the engine version.",
						},
						"engine_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The version of the Flink engine.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the engine version (STABLE, BETA, DEPRECATED, EXPIRED).",
						},
						"use_for_sql_deployments": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this engine version can be used for SQL deployments.",
						},
						"support_native_savepoint": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this engine version supports native savepoints.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudFlinkEnginesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)

	// Use FlinkService to list engines
	engines, err := flinkService.ListEngines(workspaceId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_engines", "ListEngines", AlibabaCloudSdkGoERROR)
	}

	var filteredEngines []*aliyunAPI.FlinkEngine
	nameRegex, nameRegexOk := d.GetOk("name_regex")

	// Apply name regex filter if specified
	for _, engine := range engines {
		if nameRegexOk && nameRegex != nil {
			r, err := regexp.Compile(nameRegex.(string))
			if err != nil {
				return WrapError(err)
			}
			if !r.MatchString(engine.EngineVersion) {
				continue
			}
		}
		filteredEngines = append(filteredEngines, engine)
	}

	ids := make([]string, 0)
	names := make([]string, 0)
	s := make([]map[string]interface{}, 0)

	for _, engine := range filteredEngines {
		mapping := map[string]interface{}{
			"id":             engine.EngineVersion,
			"engine_version": engine.EngineVersion,
			"status":         engine.Status,
		}

		// Set default values for features
		mapping["use_for_sql_deployments"] = false
		mapping["support_native_savepoint"] = false

		// Extract features if available
		if engine.Features != nil {
			if engine.Features.UseForSqlDeployments {
				mapping["use_for_sql_deployments"] = engine.Features.UseForSqlDeployments
			}
			if engine.Features.SupportNativeSavepoint {
				mapping["support_native_savepoint"] = engine.Features.SupportNativeSavepoint
			}
		}

		ids = append(ids, engine.EngineVersion)
		names = append(names, engine.EngineVersion)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("engines", s); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
