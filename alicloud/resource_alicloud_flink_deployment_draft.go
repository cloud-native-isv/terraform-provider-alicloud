package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
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
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the namespace.",
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

	workspaceID := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)
	artifactURI := d.Get("artifact_uri").(string)

	// Create deployment draft using cws-lib-go service
	request := &aliyunAPI.Deployment{
		Workspace: workspaceID,
		Namespace: namespace,
		Name:      name,
	}

	// Set artifact
	request.Artifact = &aliyunAPI.Artifact{
		Kind: "JAR",
		JarArtifact: &aliyunAPI.JarArtifact{
			JarUri: artifactURI,
		},
	}

	// Set deployment ID if provided
	if deploymentID, ok := d.GetOk("deployment_id"); ok {
		request.ReferencedDeploymentDraftId = deploymentID.(string)
	}

	// Handle resource specifications
	if jmSpecs, ok := d.GetOk("job_manager_resource_spec"); ok {
		jmSpecList := jmSpecs.([]interface{})
		if len(jmSpecList) > 0 {
			jmSpec := jmSpecList[0].(map[string]interface{})
			if request.StreamingResourceSetting == nil {
				request.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
					ResourceSettingMode:  "BASIC",
					BasicResourceSetting: &aliyunAPI.BasicResourceSetting{},
				}
			}
			if request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec == nil {
				request.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &aliyunAPI.ResourceSettingSpec{}
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
				request.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
					ResourceSettingMode:  "BASIC",
					BasicResourceSetting: &aliyunAPI.BasicResourceSetting{},
				}
			}
			if request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec == nil {
				request.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &aliyunAPI.ResourceSettingSpec{}
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
	response, err := flinkService.CreateDeploymentDraft(namespace, &aliyunAPI.DeploymentDraft{
		Workspace:                workspaceID,
		Namespace:                namespace,
		Name:                     name,
		Artifact:                 request.Artifact,
		FlinkConf:                request.FlinkConf,
		StreamingResourceSetting: request.StreamingResourceSetting,
		ReferencedDeploymentId:   request.ReferencedDeploymentDraftId,
	})
	if err != nil {
		return WrapError(err)
	}

	if response == nil || response.DeploymentDraftId == "" {
		return WrapError(fmt.Errorf("failed to get deployment draft ID from response"))
	}

	d.SetId(fmt.Sprintf("%s:%s", namespace, response.DeploymentDraftId))
	d.Set("draft_id", response.DeploymentDraftId)

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
	d.Set("namespace_id", deploymentDraft.Namespace)
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
			deploymentDraft.Artifact = &aliyunAPI.Artifact{
				Kind:        "JAR",
				JarArtifact: &aliyunAPI.JarArtifact{},
			}
		}

		if deploymentDraft.Artifact.JarArtifact == nil {
			deploymentDraft.Artifact.JarArtifact = &aliyunAPI.JarArtifact{}
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
			deploymentDraft.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
				ResourceSettingMode:  "BASIC",
				BasicResourceSetting: &aliyunAPI.BasicResourceSetting{},
			}
		}

		// Job Manager resource spec
		if d.HasChange("job_manager_resource_spec") {
			jmSpecs := d.Get("job_manager_resource_spec").([]interface{})
			if len(jmSpecs) > 0 {
				jmSpec := jmSpecs[0].(map[string]interface{})
				if deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec == nil {
					deploymentDraft.StreamingResourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &aliyunAPI.ResourceSettingSpec{}
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
					deploymentDraft.StreamingResourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &aliyunAPI.ResourceSettingSpec{}
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
