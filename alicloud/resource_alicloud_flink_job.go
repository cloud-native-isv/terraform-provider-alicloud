package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkJobCreate,
		Read:   resourceAliCloudFlinkJobRead,
		Update: resourceAliCloudFlinkJobUpdate,
		Delete: resourceAliCloudFlinkJobDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The workspace ID where the Flink job will be deployed.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespace name where the Flink job will be deployed.",
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The deployment ID for the Flink job.",
			},
			"restore_strategy": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The restore strategy for the Flink job.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"NONE", "LATEST_SAVEPOINT", "FROM_SAVEPOINT", "LATEST_STATE"}, false),
							Description:  "The restore strategy kind.",
						},
						"savepoint_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The savepoint ID to restore from.",
						},
					},
				},
			},
			"local_variables": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Local variables for the Flink job.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Variable name.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Variable value.",
						},
					},
				},
			},
			"with_savepoint": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to create a savepoint when stopping the job.",
			},
			// Computed fields - read-only, returned from the API
			"job_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Flink job.",
			},
			"parallelism": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The parallelism of the Flink job.",
			},
			"max_parallelism": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum parallelism of the Flink job.",
			},
			"execution_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The execution mode of the Flink job.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The engine version of the Flink job.",
			},
			"session_cluster_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The session cluster name for the Flink job.",
			},
			"flink_conf": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The Flink configuration.",
			},
			"user_flink_conf": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The user Flink configuration.",
			},
			"job_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Flink job.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the Flink job.",
			},
			"deployment_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the deployment.",
			},
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The start time of the Flink job.",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The end time of the Flink job.",
			},
			"duration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The duration of the Flink job.",
			},
			"creator": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator of the Flink job.",
			},
			"creator_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator name of the Flink job.",
			},
			"modifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier of the Flink job.",
			},
			"modifier_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier name of the Flink job.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the Flink job.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modification time of the Flink job.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	deploymentId := d.Get("deployment_id").(string)

	params := &aliyunFlinkAPI.JobStartParameters{
		WorkspaceId:  workspaceId,
		Namespace:    namespaceName,
		DeploymentId: deploymentId,
	}

	restoreList := d.Get("restore_strategy").([]interface{})
	if len(restoreList) > 0 {
		restoreMap := restoreList[0].(map[string]interface{})
		params.RestoreStrategy = &aliyunFlinkAPI.DeploymentRestoreStrategy{
			Kind: restoreMap["kind"].(string),
		}
		if savepointId, exists := restoreMap["savepoint_id"]; exists && savepointId.(string) != "" {
			params.RestoreStrategy.SavepointId = savepointId.(string)
		}
	}

	if v, ok := d.GetOk("local_variables"); ok {
		variableSet := v.(*schema.Set)
		localVars := make([]*aliyunFlinkAPI.LocalVariable, 0, variableSet.Len())
		for _, varInterface := range variableSet.List() {
			varMap := varInterface.(map[string]interface{})
			localVars = append(localVars, &aliyunFlinkAPI.LocalVariable{
				Name:  varMap["name"].(string),
				Value: varMap["value"].(string),
			})
		}
		params.LocalVariables = localVars
	}

	// Check if there's already a running job in the same deployment
	jobs, err := flinkService.ListJobs(workspaceId, namespaceName, deploymentId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, deploymentId), "ListJobs", AlibabaCloudSdkGoERROR)
	}

	// Stop any existing running jobs in the same deployment
	for _, job := range jobs {
		if job.Status != nil && job.Status.CurrentJobStatus != nil && *job.Status.CurrentJobStatus == "RUNNING" {
			// Stop the existing running job
			existingJobId := EncodeJobId(workspaceId, namespaceName, job.JobId)
			withSavepoint := true // Use savepoint when stopping existing job
			stopErr := flinkService.StopJob(existingJobId, withSavepoint)
			if stopErr != nil {
				if !IsNotFoundError(stopErr) {
					return WrapErrorf(stopErr, DefaultErrorMsg, existingJobId, "StopJob", AlibabaCloudSdkGoERROR)
				}
			}

			// Wait for the existing job to stop
			if err := flinkService.WaitForFlinkJobStopping(existingJobId, 5*time.Minute); err != nil {
				return WrapErrorf(err, IdMsg, existingJobId)
			}
		}
	}

	job, err := flinkService.StartJob(params)
	if err != nil {
		// Handle the specific error "Existing job count exceed limit"
		if IsExpectedErrors(err, FlinkJobErrors) {
			// If we still get this error after stopping existing jobs,
			// it might be a race condition or the job is in a transitional state
			// Retry once more after a short delay
			time.Sleep(10 * time.Second)

			// Try to list jobs again and stop any that might have appeared
			jobs, listErr := flinkService.ListJobs(workspaceId, namespaceName, deploymentId)
			if listErr == nil {
				for _, job := range jobs {
					if job.Status != nil && job.Status.CurrentJobStatus != nil && *job.Status.CurrentJobStatus == "RUNNING" {
						existingJobId := EncodeJobId(workspaceId, namespaceName, job.JobId)
						withSavepoint := true
						stopErr := flinkService.StopJob(existingJobId, withSavepoint)
						if stopErr == nil {
							flinkService.WaitForFlinkJobStopping(existingJobId, 2*time.Minute)
						}
					}
				}
			}

			// Try to start the job again
			job, err = flinkService.StartJob(params)
			if err != nil {
				return WrapError(err)
			}
		} else {
			return WrapError(err)
		}
	}

	d.SetId(EncodeJobId(workspaceId, namespaceName, job.JobId))

	// Use abstracted wait method from service layer
	if err := flinkService.WaitForFlinkJobCreating(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAliCloudFlinkJobRead(d, meta)
}

func resourceAliCloudFlinkJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	job, err := flinkService.DescribeFlinkJob(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Check if the parent deployment still exists
	// If deployment is deleted, the job should also be considered deleted
	if job.DeploymentId != "" {
		workspaceId := job.Workspace
		namespaceName := job.Namespace
		deploymentId := job.DeploymentId

		deploymentResourceId := fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, deploymentId)
		_, deploymentErr := flinkService.GetDeployment(deploymentResourceId)
		if deploymentErr != nil && IsNotFoundError(deploymentErr) {
			// Parent deployment no longer exists, remove job from state
			d.SetId("")
			return nil
		}
	}

	d.Set("workspace_id", job.Workspace)
	d.Set("namespace_name", job.Namespace)
	d.Set("deployment_id", job.DeploymentId)
	d.Set("job_id", job.JobId)
	d.Set("deployment_name", job.DeploymentName)

	if job.Status != nil {
		d.Set("status", job.Status.CurrentJobStatus)
	}

	d.Set("parallelism", int(job.Parallelism))
	d.Set("max_parallelism", int(job.MaxParallelism))

	if job.StartTime > 0 {
		d.Set("start_time", fmt.Sprintf("%d", job.StartTime))
	}
	if job.EndTime > 0 {
		d.Set("end_time", fmt.Sprintf("%d", job.EndTime))
	}
	if job.Duration > 0 {
		d.Set("duration", fmt.Sprintf("%d", job.Duration))
	}

	d.Set("execution_mode", job.ExecutionMode)
	d.Set("engine_version", job.EngineVersion)
	d.Set("session_cluster_name", job.SessionClusterName)
	d.Set("creator", job.Creator)
	d.Set("creator_name", job.CreatorName)
	d.Set("modifier", job.Modifier)
	d.Set("modifier_name", job.ModifierName)
	d.Set("created_at", job.CreatedAt)
	d.Set("modified_at", job.ModifiedAt)

	// Set FlinkConf and UserFlinkConf fields (computed only)
	if job.FlinkConf != nil {
		flinkConfMap := make(map[string]interface{})
		for k, v := range job.FlinkConf {
			flinkConfMap[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("flink_conf", flinkConfMap); err != nil {
			return WrapError(err)
		}
	} else {
		if err := d.Set("flink_conf", nil); err != nil {
			return WrapError(err)
		}
	}

	if job.UserFlinkConf != nil {
		userFlinkConfMap := make(map[string]interface{})
		for k, v := range job.UserFlinkConf {
			userFlinkConfMap[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("user_flink_conf", userFlinkConfMap); err != nil {
			return WrapError(err)
		}
	} else {
		if err := d.Set("user_flink_conf", nil); err != nil {
			return WrapError(err)
		}
	}

	if _, ok := d.GetOk("with_savepoint"); !ok {
		d.Set("with_savepoint", true)
	}

	if job.RestoreStrategy != nil {
		restoreStrategy := []map[string]interface{}{
			{
				"kind":         job.RestoreStrategy.Kind,
				"savepoint_id": job.RestoreStrategy.SavepointId,
			},
		}
		if err := d.Set("restore_strategy", restoreStrategy); err != nil {
			return WrapError(err)
		}
	}

	if len(job.LocalVariables) > 0 {
		localVars := make([]map[string]interface{}, 0, len(job.LocalVariables))
		for _, variable := range job.LocalVariables {
			localVars = append(localVars, map[string]interface{}{
				"name":  variable.Name,
				"value": variable.Value,
			})
		}

		localVarsInterface := make([]interface{}, len(localVars))
		for i, v := range localVars {
			localVarsInterface[i] = v
		}

		if err := d.Set("local_variables", schema.NewSet(schema.HashResource(resourceAliCloudFlinkJob().Schema["local_variables"].Elem.(*schema.Resource)), localVarsInterface)); err != nil {
			return WrapError(err)
		}
	} else {
		if err := d.Set("local_variables", nil); err != nil {
			return WrapError(err)
		}
	}

	return nil
}

func resourceAliCloudFlinkJobUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*connectivity.AliyunClient)
	// flinkService, err := NewFlinkService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }

	// Since parallelism is now ForceNew: true, only with_savepoint can be updated
	if d.HasChange("with_savepoint") {
		// with_savepoint is only used during deletion, no API call needed for update
		return resourceAliCloudFlinkJobRead(d, meta)
	}

	// No updatable changes detected
	return resourceAliCloudFlinkJobRead(d, meta)
}

func resourceAliCloudFlinkJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	withSavepoint := d.Get("with_savepoint").(bool)

	// StopJob now handles status checking internally
	stopErr := flinkService.StopJob(d.Id(), withSavepoint)
	if stopErr != nil {
		if IsNotFoundError(stopErr) {
			return nil
		}
		return WrapError(stopErr)
	}

	// Use abstracted wait method from service layer for stopping
	if err := flinkService.WaitForFlinkJobStopping(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return err
	}

	err = flinkService.DeleteJob(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	// Use abstracted wait method from service layer for deleting
	if err := flinkService.WaitForFlinkJobDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}

	return nil
}
