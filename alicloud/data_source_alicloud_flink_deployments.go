package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAlicloudFlinkDeployments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkDeploymentsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
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
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// Add other fields as needed
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

	namespaceName := d.Get("namespace").(string)

	deployments, err := flinkService.ListDeployments(namespaceName)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_flink_deployments", "ListDeployments", AliyunLogGoSdkERROR)
	}

	var ids []string
	var objects []map[string]interface{}

	for _, item := range deployments {
		object := map[string]interface{}{
			"deployment_id":  item.DeploymentId,
			"name":           item.Name,
			"description":    item.Description,
			"status":         item.Status,
			"engine_version": item.EngineVersion,
			"execution_mode": item.ExecutionMode,
			"create_time":    item.CreateTime,
			"modified_time":  item.ModifiedTime,
		}

		// Add artifact information if available
		if item.Artifact != nil {
			artifactInfo := map[string]interface{}{
				"kind": item.Artifact.Kind,
			}

			// Handle different artifact types
			switch item.Artifact.Kind {
			case "JAR":
				if item.Artifact.JarArtifact != nil {
					artifactInfo["jar_uri"] = item.Artifact.JarArtifact.JarUri
					artifactInfo["entry_class"] = item.Artifact.JarArtifact.EntryClass
					artifactInfo["main_args"] = item.Artifact.JarArtifact.MainArgs
				}
			case "PYTHON":
				if item.Artifact.PythonArtifact != nil {
					artifactInfo["python_artifact_uri"] = item.Artifact.PythonArtifact.PythonArtifactUri
					artifactInfo["entry_module"] = item.Artifact.PythonArtifact.EntryModule
					artifactInfo["main_args"] = item.Artifact.PythonArtifact.MainArgs
				}
			case "SQLSCRIPT":
				if item.Artifact.SqlArtifact != nil {
					artifactInfo["sql_script"] = item.Artifact.SqlArtifact.SqlScript
				}
			}

			object["artifact"] = []map[string]interface{}{artifactInfo}
		}

		objects = append(objects, object)
		ids = append(ids, item.DeploymentId)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("deployments", objects); err != nil {
		return WrapError(err)
	}
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	// output
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), objects)
	}

	return nil
}
