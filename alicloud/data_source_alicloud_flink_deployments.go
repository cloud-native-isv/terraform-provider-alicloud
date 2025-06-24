package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudFlinkDeployments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkDeploymentsRead,
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
			"deployments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"deployment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"workspace_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"job_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"execution_mode": {
							Type:     schema.TypeString,
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
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modified_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudFlinkDeploymentsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	// Get all deployments
	deployments, err := flinkService.ListDeployments(namespaceName)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_deployments", "ListDeployments", AlibabaCloudSdkGoERROR)
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

	var deploymentMaps []map[string]interface{}
	var filteredIds []string
	var filteredNames []string

	for _, deployment := range deployments {
		deploymentId := fmt.Sprintf("%s:%s", namespaceName, deployment.DeploymentId)

		// Apply filters
		if len(idsMap) > 0 {
			if _, ok := idsMap[deploymentId]; !ok {
				continue
			}
		}

		if len(namesMap) > 0 {
			if _, ok := namesMap[deployment.Name]; !ok {
				continue
			}
		}

		deploymentMap := map[string]interface{}{
			"id":             deploymentId,
			"deployment_id":  deployment.DeploymentId,
			"name":           deployment.Name,
			"namespace":      deployment.Namespace,
			"workspace_id":   workspaceId,
			"status":         deployment.Status,
			"job_id":         deployment.JobID,
			"engine_version": deployment.EngineVersion,
			"execution_mode": deployment.ExecutionMode,
			"creator":        deployment.Creator,
			"creator_name":   deployment.CreatorName,
			"modifier":       deployment.Modifier,
			"modifier_name":  deployment.ModifierName,
			"create_time":    deployment.CreatedAt,
			"modified_time":  deployment.ModifiedAt,
		}

		deploymentMaps = append(deploymentMaps, deploymentMap)
		filteredIds = append(filteredIds, deploymentId)
		filteredNames = append(filteredNames, deployment.Name)
	}

	d.SetId(fmt.Sprintf("%s:%s:%d", workspaceId, namespaceName, time.Now().Unix()))

	if err := d.Set("ids", filteredIds); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", filteredNames); err != nil {
		return WrapError(err)
	}
	if err := d.Set("deployments", deploymentMaps); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), deploymentMaps); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
