package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Description: "The ID of the Flink workspace.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the namespace.",
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the associated deployment.",
			},
			"job_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the job. Defaults to the deployment name if not specified.",
			},
			"parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The parallelism level for the job.",
			},
			"max_parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum parallelism level for the job.",
			},
			"execution_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The execution mode of the job (STREAMING or BATCH).",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The Flink engine version.",
			},
			"session_cluster_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The session cluster name for the job.",
			},
			"restore_strategy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The restore strategy for the job.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The restore strategy kind (LATEST_SAVEPOINT, LATEST_CHECKPOINT, SAVEPOINT).",
						},
						"savepoint_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The savepoint ID for restore.",
						},
					},
				},
			},
			"local_variables": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Local variables for the job.",
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
			"flink_conf": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Description: "Flink configuration parameters.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"user_flink_conf": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Description: "User-defined Flink configuration parameters.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			// Computed fields
			"job_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Flink job ID.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the job (e.g., RUNNING, STOPPED, FAILED).",
			},
			"deployment_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployment name associated with the job.",
			},
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The start time of the job (Unix timestamp).",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The end time of the job (Unix timestamp).",
			},
			"duration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The duration of the job in milliseconds.",
			},
			"creator": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator of the job.",
			},
			"creator_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator name of the job.",
			},
			"modifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier of the job.",
			},
			"modifier_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier name of the job.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the job.",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modification time of the job.",
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

	// Get parameters from schema
	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	deploymentId := d.Get("deployment_id").(string)

	// Build job start parameters - using JobStartParameters struct
	params := &aliyunFlinkAPI.JobStartParameters{
		Namespace:    namespaceName,
		DeploymentId: deploymentId,
	}

	// Handle restore strategy
	if v, ok := d.GetOk("restore_strategy"); ok {
		restoreList := v.([]interface{})
		if len(restoreList) > 0 {
			restoreMap := restoreList[0].(map[string]interface{})
			params.RestoreStrategy = &aliyunFlinkAPI.DeploymentRestoreStrategy{
				Kind: restoreMap["kind"].(string),
			}
			if savepointId, exists := restoreMap["savepoint_id"]; exists && savepointId.(string) != "" {
				params.RestoreStrategy.SavepointId = savepointId.(string)
			}
		}
	}

	// Handle local variables
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

	// Start job using FlinkService with JobStartParameters and workspaceId
	job, err := flinkService.StartJob(workspaceId, namespaceName, params)
	if err != nil {
		return WrapError(err)
	}

	// Set composite ID: workspaceId:namespace:jobId
	d.SetId(fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, job.JobId))

	// Wait for job to start using StateRefreshFunc
	stateConf := BuildStateConf([]string{"STARTING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkJobStateRefreshFunc(d.Id(), []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Call Read to sync final state
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
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_job DescribeFlinkJob Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic attributes using correct field names from cws-lib-go Job type
	d.Set("workspace_id", job.Workspace)
	d.Set("namespace_name", job.Namespace)
	d.Set("deployment_id", job.DeploymentId)
	d.Set("job_name", job.JobName)
	d.Set("job_id", job.JobId)
	d.Set("deployment_name", job.DeploymentName)

	// Handle job status from job status field
	if job.Status != nil {
		d.Set("status", job.Status.CurrentJobStatus)
	}

	// Set numeric fields
	d.Set("parallelism", int(job.Parallelism))
	d.Set("max_parallelism", int(job.MaxParallelism))

	// Set time fields - convert Unix timestamps to strings
	if job.StartTime > 0 {
		d.Set("start_time", fmt.Sprintf("%d", job.StartTime))
	}
	if job.EndTime > 0 {
		d.Set("end_time", fmt.Sprintf("%d", job.EndTime))
	}
	if job.Duration > 0 {
		d.Set("duration", fmt.Sprintf("%d", job.Duration))
	}

	// Set string fields
	d.Set("execution_mode", job.ExecutionMode)
	d.Set("engine_version", job.EngineVersion)
	d.Set("session_cluster_name", job.SessionClusterName)
	d.Set("creator", job.Creator)
	d.Set("creator_name", job.CreatorName)
	d.Set("modifier", job.Modifier)
	d.Set("modifier_name", job.ModifierName)
	d.Set("created_at", job.CreatedAt)
	d.Set("modified_at", job.ModifiedAt)

	// Handle restore strategy
	if job.RestoreStrategy != nil {
		restoreStrategy := []map[string]interface{}{
			{
				"kind":         job.RestoreStrategy.Kind,
				"savepoint_id": job.RestoreStrategy.SavepointId,
			},
		}
		d.Set("restore_strategy", restoreStrategy)
	}

	// Handle local variables
	if len(job.LocalVariables) > 0 {
		localVars := make([]map[string]interface{}, 0, len(job.LocalVariables))
		for _, variable := range job.LocalVariables {
			localVars = append(localVars, map[string]interface{}{
				"name":  variable.Name,
				"value": variable.Value,
			})
		}
		d.Set("local_variables", localVars)
	}

	// Handle flink configuration
	if len(job.FlinkConf) > 0 {
		flinkConf := make(map[string]interface{})
		for key, value := range job.FlinkConf {
			flinkConf[key] = fmt.Sprintf("%v", value)
		}
		d.Set("flink_conf", flinkConf)
	}

	// Handle user flink configuration
	if len(job.UserFlinkConf) > 0 {
		userFlinkConf := make(map[string]interface{})
		for key, value := range job.UserFlinkConf {
			userFlinkConf[key] = fmt.Sprintf("%v", value)
		}
		d.Set("user_flink_conf", userFlinkConf)
	}

	return nil
}

func resourceAliCloudFlinkJobUpdate(d *schema.ResourceData, meta interface{}) error {
	// Jobs typically cannot be updated once started, only stopped and restarted
	// For now, return an error indicating updates are not supported
	return WrapError(fmt.Errorf("Flink jobs cannot be updated once started. Please delete and recreate the job."))
}

func resourceAliCloudFlinkJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse composite ID to get workspaceId, namespace and jobId
	workspaceId, namespace, jobId, err := parseJobId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Stop job with savepoint
	err = flinkService.StopJob(workspaceId, namespace, jobId, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidJob.NotFound"}) {
			return nil
		}
		return WrapError(err)
	}

	return nil
}

func buildFlinkJobPropertiesFromSet(propertiesSet *schema.Set) map[string]string {
	properties := make(map[string]string)
	for _, v := range propertiesSet.List() {
		prop := v.(map[string]interface{})
		key := prop["key"].(string)
		value := prop["value"].(string)
		properties[key] = value
	}
	return properties
}

func expandFlinkJobPropertiesFromMap(propertiesMap map[string]interface{}) map[string]string {
	properties := make(map[string]string)
	for key, value := range propertiesMap {
		properties[key] = fmt.Sprintf("%v", value)
	}
	return properties
}

func flattenFlinkJobPropertiesToSet(properties map[string]string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(properties))
	for key, value := range properties {
		result = append(result, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}
	return result
}
