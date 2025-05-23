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
			"namespace": {
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
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceID := d.Get("workspace_id").(string)
	namespace := d.Get("namespace").(string)
	deploymentID := d.Get("deployment_id").(string)

	// Get the deployment details to fetch information like the deployment name (if job_name is not specified)
	deploymentResponse, err := flinkService.GetDeployment(tea.String(namespace), tea.String(deploymentID))
	if err != nil {
		return WrapError(fmt.Errorf("error getting deployment details: %s", err))
	}

	if deploymentResponse == nil || deploymentResponse.Body == nil || deploymentResponse.Body.Data == nil {
		return WrapError(fmt.Errorf("deployment not found or invalid response"))
	}

	// Set job name to deployment name if not specified
	jobName := d.Get("job_name").(string)
	if jobName == "" && deploymentResponse.Body.Data.Name != nil {
		jobName = *deploymentResponse.Body.Data.Name
		d.Set("job_name", jobName)
	}

	// Create job start request
	request := &ververica.StartJobWithParamsRequest{}
	startParams := &ververica.StartJobParams{}
	startParams.DeploymentId = tea.String(deploymentID)

	// Set optional job parameters
	if jobName != "" {
		startParams.JobName = tea.String(jobName)
	}

	startParams.AllowNonRestoredState = tea.Bool(d.Get("allow_non_restored_state").(bool))

	if savepointPath, ok := d.GetOk("savepoint_path"); ok {
		startParams.SavepointPath = tea.String(savepointPath.(string))
	}

	// Set operation parameters if provided
	if opParams, ok := d.GetOk("operation_params"); ok {
		paramsMap := opParams.(map[string]interface{})
		params := make(map[string]*string)
		for k, v := range paramsMap {
			params[k] = tea.String(v.(string))
		}
		startParams.Params = params
	}

	// Set auto-restart behavior
	startParams.AutoRestartEnabled = tea.Bool(d.Get("auto_restart").(bool))

	// Set the start params as the body of the request
	request.Body = startParams

	// Start the job with workspace header
	response, err := flinkService.ververicaClient.StartJobWithParamsWithOptions(
		tea.String(namespace),
		request,
		&ververica.StartJobWithParamsHeaders{
			Workspace: tea.String(workspaceID),
		},
		&util.RuntimeOptions{},
	)
	if err != nil {
		return WrapError(err)
	}

	if response == nil || response.Body == nil || response.Body.Data == nil || response.Body.Data.JobId == nil {
		return WrapError(fmt.Errorf("failed to get job ID from response"))
	}

	jobID := *response.Body.Data.JobId
	d.SetId(fmt.Sprintf("%s:%s:%s", namespace, deploymentID, jobID))
	d.Set("job_id", jobID)

	// Wait for the job to reach a stable state
	stateConf := resource.StateChangeConf{
		Pending:      []string{"INITIALIZING", "CREATED", "RESTARTING", "CANCELLING", "SCHEDULED"},
		Target:       []string{"RUNNING", "FAILED", "FINISHED", "CANCELED"},
		Refresh:      flinkJobRefreshFunc(flinkService, namespace, jobID, workspaceID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
		MinTimeout:   5 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

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
	d.Set("namespace", job.Namespace)
	d.Set("deployment_id", job.DeploymentId)
	d.Set("job_id", job.JobId)

	if job.Name != nil {
		d.Set("job_name", *job.Name)
	}

	// Set job status information
	if job.Status != nil {
		if job.Status.CurrentJobStatus != nil {
			d.Set("status", *job.Status.CurrentJobStatus)
		}

		if job.Status.StartTime != nil {
			d.Set("start_time", *job.Status.StartTime)
		}

		if job.Status.EndTime != nil {
			d.Set("end_time", *job.Status.EndTime)
		}

		if job.Status.LastSavepointPath != nil {
			d.Set("last_savepoint_path", *job.Status.LastSavepointPath)
		}

		if job.Status.LastCheckpointPath != nil {
			d.Set("last_checkpoint_path", *job.Status.LastCheckpointPath)
		}
	}

	// Set Flink Web UI URL if available
	if job.FlinkWebUiUrl != nil {
		d.Set("flink_web_ui_url", *job.FlinkWebUiUrl)
	}

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
		stopRequest.Body = &ververica.JobStopParams{
			SavepointEnabled: tea.Bool(true),
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
		startParams := &ververica.StartJobParams{}
		startParams.DeploymentId = tea.String(deploymentID)

		// Set job name if specified
		if jobName, ok := d.GetOk("job_name"); ok {
			startParams.JobName = tea.String(jobName.(string))
		}

		// Set allow non-restored state flag
		startParams.AllowNonRestoredState = tea.Bool(d.Get("allow_non_restored_state").(bool))

		// Set savepoint path if specified
		if savepointPath, ok := d.GetOk("savepoint_path"); ok {
			startParams.SavepointPath = tea.String(savepointPath.(string))
		}

		// Set operation parameters if provided
		if opParams, ok := d.GetOk("operation_params"); ok {
			paramsMap := opParams.(map[string]interface{})
			params := make(map[string]*string)
			for k, v := range paramsMap {
				params[k] = tea.String(v.(string))
			}
			startParams.Params = params
		}

		// Set auto-restart behavior
		startParams.AutoRestartEnabled = tea.Bool(d.Get("auto_restart").(bool))

		// Set the start params as the body of the request
		startRequest.Body = startParams

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

		if startResponse == nil || startResponse.Body == nil || startResponse.Body.Data == nil || startResponse.Body.Data.JobId == nil {
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
	request.Body = &ververica.JobStopParams{
		SavepointEnabled: tea.Bool(true), // Try to create a savepoint when stopping
	}

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
