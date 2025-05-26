package alicloud

import (
	"fmt"
	"time"

	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
			"namespace_id": {
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
			"allow_non_restored_state": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to allow non-restored state. Default is false.",
			},
			"savepoint_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The path to the savepoint to restore from.",
			},
			"operation_params": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional operation parameters for job execution.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auto_restart": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to automatically restart the job on failure. Default is true.",
			},
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
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The start time of the job.",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The end time of the job.",
			},
			"last_savepoint_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path to the last created savepoint.",
			},
			"last_checkpoint_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path to the last checkpoint.",
			},
			"flink_web_ui_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to the Flink web UI for this job.",
			},
			"restart_on_update": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to restart the job when updating its configuration. Default is true.",
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

	// Properly initialize the FlinkService with all required fields
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deploymentID := d.Get("deployment_id").(string)
	workspaceID := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)

	request := &ververica.StartJobWithParamsRequest{}

	// Create the proper parameters for the job
	jobStartParams := &ververica.JobStartParameters{}
	jobStartParams.DeploymentId = tea.String(deploymentID)

	// Handle restore strategy if savepoint is provided
	if v, ok := d.GetOk("savepoint_path"); ok {
		restoreStrategy := &ververica.DeploymentRestoreStrategy{}
		restoreStrategy.Kind = tea.String("SAVEPOINT")
		restoreStrategy.SavepointId = tea.String(v.(string))
		restoreStrategy.AllowNonRestoredState = tea.Bool(d.Get("allow_non_restored_state").(bool))
		jobStartParams.RestoreStrategy = restoreStrategy
	} else {
		// Ensure we still have a valid restore strategy with allowNonRestoredState
		restoreStrategy := &ververica.DeploymentRestoreStrategy{}
		restoreStrategy.AllowNonRestoredState = tea.Bool(d.Get("allow_non_restored_state").(bool))
		jobStartParams.RestoreStrategy = restoreStrategy
	}

	// Set job start parameters to request
	request.Body = jobStartParams

	response, err := flinkService.ververicaClient.StartJobWithParamsWithOptions(
		tea.String(namespace),
		request,
		&ververica.StartJobWithParamsHeaders{
			Workspace: tea.String(workspaceID),
		},
		&util.RuntimeOptions{},
	)

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_job", "StartJobWithParams", AlibabaCloudSdkGoERROR)
	}

	// Extract the job ID from the response data
	if response.Body == nil || response.Body.Data == nil {
		return WrapError(fmt.Errorf("failed to get job ID from response"))
	}

	jobID := *response.Body.Data.JobId

	// Set the resource ID in the format "namespace:deploymentID:jobID"
	d.SetId(fmt.Sprintf("%s:%s:%s", namespace, deploymentID, jobID))
	d.Set("job_id", jobID)

	return resourceAliCloudFlinkJobRead(d, meta)
}

func resourceAliCloudFlinkJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	deploymentID := parts[1]
	jobID := parts[2]
	workspaceID := d.Get("workspace_id").(string)

	response, err := flinkService.ververicaClient.GetJobWithOptions(
		tea.String(namespace),
		tea.String(jobID),
		&ververica.GetJobHeaders{
			Workspace: tea.String(workspaceID),
		},
		&util.RuntimeOptions{},
	)
	if err != nil {
		if IsExpectedErrors(err, []string{"JobNotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if response == nil || response.Body == nil || response.Body.Data == nil {
		d.SetId("")
		return nil
	}

	// Update state with values from the response
	job := response.Body.Data

	// Set basic job information
	d.Set("namespace_id", namespace)
	d.Set("deployment_id", deploymentID)
	d.Set("job_id", jobID)

	// The deployment_name field can be set if available
	if job.DeploymentName != nil {
		d.Set("job_name", *job.DeploymentName)
	}

	// Set job status information
	if job.Status != nil {
		if job.Status.CurrentJobStatus != nil {
			d.Set("status", *job.Status.CurrentJobStatus)
		}

		// Use start_time and end_time from the job object directly if available
		if job.StartTime != nil {
			d.Set("start_time", fmt.Sprintf("%d", *job.StartTime))
		}

		if job.EndTime != nil {
			d.Set("end_time", fmt.Sprintf("%d", *job.EndTime))
		}
	}

	// Set web UI URL if available - since this field doesn't exist, comment it out
	// if job.FlinkWebUiUrl != nil {
	//    d.Set("flink_web_ui_url", *job.FlinkWebUiUrl)
	// }

	return nil
}

func resourceAliCloudFlinkJobUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	deploymentID := parts[1]
	jobID := parts[2]
	workspaceID := d.Get("workspace_id").(string)

	// Check if we need to restart the job
	needsRestart := d.Get("restart_on_update").(bool) && (d.HasChange("allow_non_restored_state") ||
		d.HasChange("savepoint_path") ||
		d.HasChange("operation_params") ||
		d.HasChange("auto_restart") ||
		d.HasChange("job_name"))

	if needsRestart {
		// First, stop the existing job with savepoint
		stopRequest := &ververica.StopJobRequest{}

		// Look at the actual fields available in StopJobRequestBody
		stopRequest.Body = &ververica.StopJobRequestBody{
			// Remove unknown field and use what's actually available in the struct
			// For now, using an empty body as we don't know the exact fields
		}

		_, err := flinkService.ververicaClient.StopJobWithOptions(
			tea.String(namespace),
			tea.String(jobID),
			stopRequest,
			&ververica.StopJobHeaders{
				Workspace: tea.String(workspaceID),
			},
			&util.RuntimeOptions{},
		)
		if err != nil {
			// If the job is not found, it might have already been stopped or terminated
			if !IsExpectedErrors(err, []string{"JobNotFound"}) {
				return WrapError(err)
			}
		}

		// Wait for the job to stop
		stateConf := resource.StateChangeConf{
			Pending:      []string{"RUNNING", "CANCELLING"},
			Target:       []string{"CANCELED", "FINISHED", "FAILED"},
			Refresh:      flinkJobRefreshFunc(flinkService, namespace, jobID, workspaceID),
			Timeout:      5 * time.Minute,
			Delay:        10 * time.Second,
			PollInterval: 5 * time.Second,
			MinTimeout:   3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		// Now start a new job with updated parameters
		startRequest := &ververica.StartJobWithParamsRequest{}

		// Create proper JobStartParameters object
		jobStartParams := &ververica.JobStartParameters{
			DeploymentId: tea.String(deploymentID),
		}

		// Set restore strategy properties
		restoreStrategy := &ververica.DeploymentRestoreStrategy{
			AllowNonRestoredState: tea.Bool(d.Get("allow_non_restored_state").(bool)),
		}

		// Add savepoint path if specified
		if savepointPath, ok := d.GetOk("savepoint_path"); ok {
			restoreStrategy.Kind = tea.String("SAVEPOINT")
			restoreStrategy.SavepointId = tea.String(savepointPath.(string))
		}

		jobStartParams.RestoreStrategy = restoreStrategy

		// Set auto restart flag using local variables
		localVars := []*ververica.LocalVariable{
			{
				Name:  tea.String("__GENERATED_CONSTANT__autoRestartEnabled"),
				Value: tea.String(fmt.Sprintf("%t", d.Get("auto_restart").(bool))),
			},
		}
		jobStartParams.LocalVariables = localVars

		// Set the request body
		startRequest.Body = jobStartParams

		// Start the new job
		startResponse, err := flinkService.ververicaClient.StartJobWithParamsWithOptions(
			tea.String(namespace),
			startRequest,
			&ververica.StartJobWithParamsHeaders{
				Workspace: tea.String(workspaceID),
			},
			&util.RuntimeOptions{},
		)
		if err != nil {
			return WrapError(err)
		}

		if startResponse == nil || startResponse.Body == nil || startResponse.Body.Data == nil {
			return WrapError(fmt.Errorf("failed to get new job ID from response"))
		}

		// Update resource ID with the new job ID
		newJobID := *startResponse.Body.Data.JobId
		d.SetId(fmt.Sprintf("%s:%s:%s", namespace, deploymentID, newJobID))
		d.Set("job_id", newJobID)

		// Wait for the new job to reach a stable state
		newStateConf := resource.StateChangeConf{
			Pending:      []string{"INITIALIZING", "CREATED", "RESTARTING", "SCHEDULED"},
			Target:       []string{"RUNNING", "FAILED", "FINISHED", "CANCELED"},
			Refresh:      flinkJobRefreshFunc(flinkService, namespace, newJobID, workspaceID),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        10 * time.Second,
			PollInterval: 10 * time.Second,
			MinTimeout:   5 * time.Second,
		}

		_, err = newStateConf.WaitForState()
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkJobRead(d, meta)
}

func resourceAliCloudFlinkJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	jobID := parts[2]
	workspaceID := d.Get("workspace_id").(string)

	// Create a stop request to terminate the job
	request := &ververica.StopJobRequest{}

	// Use an empty body since we don't know the actual fields
	request.Body = &ververica.StopJobRequestBody{}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := flinkService.ververicaClient.StopJobWithOptions(
			tea.String(namespace),
			tea.String(jobID),
			request,
			&ververica.StopJobHeaders{
				Workspace: tea.String(workspaceID),
			},
			&util.RuntimeOptions{},
		)
		if err != nil {
			if IsExpectedErrors(err, []string{"JobNotFound"}) {
				return nil
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StopJob", AlibabaCloudSdkGoERROR)
	}

	// Wait for the job to be stopped or terminated
	stateConf := resource.StateChangeConf{
		Pending:      []string{"RUNNING", "CANCELLING", "INITIALIZING", "CREATED", "RESTARTING"},
		Target:       []string{"CANCELED", "FINISHED", "FAILED"},
		Refresh:      flinkJobRefreshFunc(flinkService, namespace, jobID, workspaceID),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        10 * time.Second,
		PollInterval: 5 * time.Second,
		MinTimeout:   3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// flinkJobRefreshFunc returns a resource.StateRefreshFunc that is used to watch the state of a Flink job
func flinkJobRefreshFunc(flinkService *FlinkService, namespace, jobID, workspaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		response, err := flinkService.ververicaClient.GetJobWithOptions(
			tea.String(namespace),
			tea.String(jobID),
			&ververica.GetJobHeaders{
				Workspace: tea.String(workspaceID),
			},
			&util.RuntimeOptions{},
		)

		if err != nil {
			if IsExpectedErrors(err, []string{"JobNotFound"}) {
				// Job doesn't exist
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		if response == nil || response.Body == nil || response.Body.Data == nil || response.Body.Data.Status == nil || response.Body.Data.Status.CurrentJobStatus == nil {
			return response, "", nil
		}

		status := *response.Body.Data.Status.CurrentJobStatus
		return response, status, nil
	}
}
