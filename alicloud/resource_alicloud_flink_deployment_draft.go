package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudFlinkDeploymentDraft() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentDraftCreate,
		Read:   resourceAliCloudFlinkDeploymentDraftRead,
		Update: resourceAliCloudFlinkDeploymentDraftUpdate,
		Delete: resourceAliCloudFlinkDeploymentDraftDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Flink workspace.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the namespace.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the deployment draft.",
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Associated deployment ID.",
			},
			"artifact_uri": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URI of the deployment artifact (e.g., JAR file in OSS).",
			},
			"flink_configuration": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of Flink configuration properties.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"job_manager_resource_spec": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "CPU cores for JobManager.",
						},
						"memory": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Memory for JobManager (e.g., \"1g\").",
						},
					},
				},
			},
			"task_manager_resource_spec": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "CPU cores for TaskManager.",
						},
						"memory": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Memory for TaskManager (e.g., \"2g\").",
						},
					},
				},
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of environment variables.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the deployment draft.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the deployment draft.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the deployment draft.",
			},
			"draft_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the deployment draft.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkDeploymentDraftCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)
	artifactURI := d.Get("artifact_uri").(string)

	// Create deployment draft using cws-lib-go service
	request := &aliyunFlinkAPI.Deployment{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      name,
	}

	// Set artifact
	request.Artifact = &aliyunFlinkAPI.Artifact{
		Kind: "JAR",
		JarArtifact: &aliyunFlinkAPI.JarArtifact{
			JarUri: artifactURI,
		},
	}

	// Set deployment ID if provided
	if deploymentId, ok := d.GetOk("deployment_id"); ok {
		request.ReferencedDeploymentDraftId = deploymentId.(string)
	}

	// Handle resource specifications
	if jmSpecs, ok := d.GetOk("job_manager_resource_spec"); ok {
		jmSpecList := jmSpecs.([]interface{})
		if len(jmSpecList) > 0 {
			jmSpec := jmSpecList[0].(map[string]interface{})
			if request.StreamingResourceSetting == nil {
				request.StreamingResourceSetting = &aliyunFlinkAPI.StreamingResourceSetting{
					ResourceSettingMode:  "BASIC",
					BasicResourceSetting: &aliyunFlinkAPI.BasicResourceSetting{},
				}
			}
			if request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec == nil {
				request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &aliyunFlinkAPI.ResourceSettingSpec{}
			}
			request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec.CPU = jmSpec["cpu"].(float64)
			request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec.Memory = jmSpec["memory"].(string)
		}
	}

	if tmSpecs, ok := d.GetOk("task_manager_resource_spec"); ok {
		tmSpecList := tmSpecs.([]interface{})
		if len(tmSpecList) > 0 {
			tmSpec := tmSpecList[0].(map[string]interface{})
			if request.StreamingResourceSetting == nil {
				request.StreamingResourceSetting = &aliyunFlinkAPI.StreamingResourceSetting{
					ResourceSettingMode:  "BASIC",
					BasicResourceSetting: &aliyunFlinkAPI.BasicResourceSetting{},
				}
			}
			if request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec == nil {
				request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &aliyunFlinkAPI.ResourceSettingSpec{}
			}
			request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec.CPU = tmSpec["cpu"].(float64)
			request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec.Memory = tmSpec["memory"].(string)
		}
	}

	// Handle Flink configuration
	if flinkConf, ok := d.GetOk("flink_configuration"); ok {
		request.FlinkConf = make(map[string]string)
		for k, v := range flinkConf.(map[string]interface{}) {
			request.FlinkConf[k] = v.(string)
		}
	}

	// Create the deployment draft
	var response *aliyunFlinkAPI.DeploymentDraft
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateDeploymentDraft(workspaceId, namespaceName, &aliyunFlinkAPI.DeploymentDraft{
			Workspace:                workspaceId,
			Namespace:                namespaceName,
			Name:                     name,
			Artifact:                 request.Artifact,
			FlinkConf:                request.FlinkConf,
			StreamingResourceSetting: request.StreamingResourceSetting,
			ReferencedDeploymentId:   request.ReferencedDeploymentDraftId,
		})
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_draft", "CreateDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.DeploymentDraftId == "" {
		return WrapError(Error("Failed to get deployment draft ID from response"))
	}

	d.SetId(fmt.Sprintf("%s:%s", namespaceName, response.DeploymentDraftId))
	d.Set("draft_id", response.DeploymentDraftId)

	// Wait for deployment draft creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkDeploymentDraftStateRefreshFunc(workspaceId, namespaceName, response.DeploymentDraftId, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkDeploymentDraftRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDraftRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	// Use FlinkService method instead of directly accessing VervericaClient
	deploymentDraft, err := flinkService.GetDeploymentDraft(workspaceID, namespace, draftID)
	if err != nil {
		if IsExpectedErrors(err, []string{"DraftNotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if deploymentDraft == nil {
		d.SetId("")
		return nil
	}

	// Update state with values from the response
	d.Set("namespace_name", deploymentDraft.Namespace)
	d.Set("workspace_id", deploymentDraft.Workspace)
	d.Set("name", deploymentDraft.Name)
	d.Set("draft_id", deploymentDraft.DeploymentDraftId)

	if deploymentDraft.ReferencedDeploymentId != "" {
		d.Set("deployment_id", deploymentDraft.ReferencedDeploymentId)
	}

	if deploymentDraft.CreatedAt > 0 {
		// Convert from int64 to string if needed
		createdAtStr := fmt.Sprintf("%d", deploymentDraft.CreatedAt)
		d.Set("create_time", createdAtStr)
	}

	if deploymentDraft.ModifiedAt > 0 {
		// Convert from int64 to string if needed
		modifiedAtStr := fmt.Sprintf("%d", deploymentDraft.ModifiedAt)
		d.Set("update_time", modifiedAtStr)
	}

	// Set artifact URI
	if deploymentDraft.Artifact != nil {
		if deploymentDraft.Artifact.JarArtifact != nil {
			d.Set("artifact_uri", deploymentDraft.Artifact.JarArtifact.JarUri)
		}
	}

	return nil
}

func resourceAliCloudFlinkDeploymentDraftUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	// Use service methods instead of directly accessing Ververica client
	// First, get the current deployment draft
	deploymentDraft, err := flinkService.GetDeploymentDraft(workspaceID, namespace, draftID)
	if err != nil {
		return WrapError(err)
	}

	// Only update fields that have changed
	update := false

	if d.HasChange("name") {
		deploymentDraft.Name = d.Get("name").(string)
		update = true
	}

	if d.HasChange("deployment_id") {
		if deploymentID, ok := d.GetOk("deployment_id"); ok {
			deploymentDraft.ReferencedDeploymentId = deploymentID.(string)
		} else {
			deploymentDraft.ReferencedDeploymentId = ""
		}
		update = true
	}

	if d.HasChange("artifact_uri") {
		if deploymentDraft.Artifact == nil {
			deploymentDraft.Artifact = &aliyunFlinkAPI.Artifact{
				Kind:        "JAR",
				JarArtifact: &aliyunFlinkAPI.JarArtifact{},
			}
		}

		if deploymentDraft.Artifact.JarArtifact == nil {
			deploymentDraft.Artifact.JarArtifact = &aliyunFlinkAPI.JarArtifact{}
		}

		deploymentDraft.Artifact.JarArtifact.JarUri = d.Get("artifact_uri").(string)
		update = true
	}

	// Handle Flink configuration if changed
	if d.HasChange("flink_configuration") {
		flinkConf := d.Get("flink_configuration").(map[string]interface{})
		deploymentDraft.FlinkConf = make(map[string]string)
		for k, v := range flinkConf {
			deploymentDraft.FlinkConf[k] = v.(string)
		}
		update = true
	}

	// Handle resource specifications if changed
	if d.HasChange("job_manager_resource_spec") || d.HasChange("task_manager_resource_spec") {
		if deploymentDraft.StreamingResourceSetting == nil {
			deploymentDraft.StreamingResourceSetting = &aliyunFlinkAPI.StreamingResourceSetting{
				ResourceSettingMode:  "BASIC",
				BasicResourceSetting: &aliyunFlinkAPI.BasicResourceSetting{},
			}
		}

		// Job Manager resource spec
		if d.HasChange("job_manager_resource_spec") {
			jmSpecs := d.Get("job_manager_resource_spec").([]interface{})
			if len(jmSpecs) > 0 {
				jmSpec := jmSpecs[0].(map[string]interface{})
				if deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec == nil {
					deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &aliyunFlinkAPI.ResourceSettingSpec{}
				}
				deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec.CPU = jmSpec["cpu"].(float64)
				deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec.Memory = jmSpec["memory"].(string)
			}
		}

		// Task Manager resource spec
		if d.HasChange("task_manager_resource_spec") {
			tmSpecs := d.Get("task_manager_resource_spec").([]interface{})
			if len(tmSpecs) > 0 {
				tmSpec := tmSpecs[0].(map[string]interface{})
				if deploymentDraft.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec == nil {
					deploymentDraft.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &aliyunFlinkAPI.ResourceSettingSpec{}
				}
				deploymentDraft.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec.CPU = tmSpec["cpu"].(float64)
				deploymentDraft.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec.Memory = tmSpec["memory"].(string)
			}
		}
		update = true
	}

	if update {
		// Call the service method to update the deployment draft
		_, err = flinkService.UpdateDeploymentDraft(workspaceID, namespace, draftID, deploymentDraft)
		if err != nil {
			return WrapError(err)
		}

		// Wait a moment for the update to be processed
		time.Sleep(5 * time.Second)
	}

	return resourceAliCloudFlinkDeploymentDraftRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDraftDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		// Use service method instead of directly accessing VervericaClient
		err := flinkService.DeleteDeploymentDraft(workspaceID, namespace, draftID)
		if err != nil {
			if IsExpectedErrors(err, []string{"DraftNotFound"}) {
				return nil
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	return nil
}
