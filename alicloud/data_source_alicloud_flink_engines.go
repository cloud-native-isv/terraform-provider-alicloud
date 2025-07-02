package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudFlinkEngines() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkEnginesRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
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
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engines": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_for_sql_deployments": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"support_native_savepoint": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFlinkEnginesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)

	// Get all engines
	engines, err := flinkService.ListEngines(workspaceId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_engines", "ListEngines", AlibabaCloudSdkGoERROR)
	}

	// Filter results if ids are provided
	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	var engineMaps []map[string]interface{}
	var filteredIds []string

	for _, engine := range engines {
		engineId := fmt.Sprintf("%s:%s", workspaceId, engine.EngineVersion)

		// Apply filters
		if len(idsMap) > 0 {
			if _, ok := idsMap[engineId]; !ok {
				continue
			}
		}

		engineMap := map[string]interface{}{
			"id":                       engineId,
			"engine_version":           engine.EngineVersion,
			"status":                   engine.Status,
			"use_for_sql_deployments":  false,
			"support_native_savepoint": false,
		}

		// Set features if available
		if engine.Features != nil {
			engineMap["use_for_sql_deployments"] = engine.Features.UseForSqlDeployments
			engineMap["support_native_savepoint"] = engine.Features.SupportNativeSavepoint
		}

		engineMaps = append(engineMaps, engineMap)
		filteredIds = append(filteredIds, engineId)
	}

	d.SetId(fmt.Sprintf("%s:%d", workspaceId, time.Now().Unix()))

	if err := d.Set("ids", filteredIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("engines", engineMaps); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), engineMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
