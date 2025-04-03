package alicloud

import (
	"time"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
)

func resourceAliCloudFlinkDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentCreate,
		Read:   resourceAliCloudFlinkDeploymentRead,
		Update: resourceAliCloudFlinkDeploymentUpdate,
		Delete: resourceAliCloudFlinkDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"entry_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parallelism": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"jar_uri": {
				Type:     schema.TypeString,
				Required: true,
			},
			"jar_artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"program_args": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	namespace := d.Get("namespace").(string)
	workspaceId := d.Get("workspace_id").(string)
	jarUri := d.Get("jar_uri").(string)
	entryClass := d.Get("entry_class").(string)
	jobName := d.Get("job_name").(string)
	parallelism := d.Get("parallelism").(int)
	programArgs := d.Get("program_args").(string)
	
	deploymentName := d.Get("deployment_name").(string)
	if deploymentName == "" {
		deploymentName = jobName
	}

	// Create the deployment request with the correct structure
	request := &ververica.CreateDeploymentRequest{}
	
	// Create a Deployment object for the Body field
	deployment := &ververica.Deployment{}
	
	// Set deployment name
	deployment.Name = &deploymentName
	
	// Set workspace ID
	deployment.Workspace = &workspaceId
	
	// Create artifact for JAR info
	artifact := &ververica.Artifact{}
	artifact.Kind = stringPointer("JAR")
	artifact.JarUri = &jarUri
	artifact.EntryClass = &entryClass
	
	if programArgs != "" {
		artifact.ProgramArgs = &programArgs
	}
	
	deployment.Artifact = artifact
	
	// Set parallelism
	pInt64 := int64(parallelism)
	streamingResourceSetting := &ververica.StreamingResourceSetting{}
	streamingResourceSetting.Parallelism = &pInt64
	deployment.StreamingResourceSetting = streamingResourceSetting
	
	// Set the deployment as the body of the request
	request.Body = deployment
	
	// Create the deployment
	response, err := flinkService.CreateDeployment(&namespace, request)
	if err != nil {
		return WrapError(err)
	}

	// Access fields from the correct response structure
	d.SetId(*response.Body.Data.DeploymentId)
	d.Set("deployment_id", *response.Body.Data.DeploymentId)
	d.Set("jar_artifact_name", jarUri)

	// Create an adapter function to convert ResourceStateRefreshFunc to resource.StateRefreshFunc
	refreshFunc := func() (interface{}, string, error) {
		return flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{})()
	}

	// Wait for deployment creation to complete
	stateConf := resource.StateChangeConf{
		Pending:      []string{},
		Target:       []string{"CREATED"},
		Refresh:      refreshFunc,
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        5 * time.Second,
		PollInterval: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

// Helper function to create a string pointer
func stringPointer(s string) *string {
	return &s
}

func resourceAliCloudFlinkDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deploymentId := d.Id()
	namespace := d.Get("namespace").(string)
	
	response, err := flinkService.GetDeployment(&namespace, &deploymentId)
	if err != nil {
		if IsExpectedErrors(err, []string{"EntityNotExist.Deployment"}) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if response.Body.Data == nil {
		d.SetId("")
		return nil
	}

	// Access fields from the correct response structure
	deployment := response.Body.Data
	d.Set("deployment_id", *deployment.DeploymentId)
	d.Set("deployment_name", *deployment.Name)
	d.Set("workspace_id", *deployment.Workspace)
	d.Set("namespace", *deployment.Namespace)
	
	// Handle JAR information
	if deployment.Artifact != nil {
		if deployment.Artifact.JarUri != nil {
			d.Set("jar_uri", *deployment.Artifact.JarUri)
			d.Set("jar_artifact_name", *deployment.Artifact.JarUri)
		}
		
		if deployment.Artifact.EntryClass != nil {
			d.Set("entry_class", *deployment.Artifact.EntryClass)
		}
		
		if deployment.Artifact.ProgramArgs != nil {
			d.Set("program_args", *deployment.Artifact.ProgramArgs)
		}
	}
	
	// Handle parallelism
	if deployment.StreamingResourceSetting != nil && deployment.StreamingResourceSetting.Parallelism != nil {
		d.Set("parallelism", *deployment.StreamingResourceSetting.Parallelism)
	}
	
	if deployment.CreatedAt != nil {
		d.Set("create_time", *deployment.CreatedAt)
	}
	
	if deployment.ModifiedAt != nil {
		d.Set("update_time", *deployment.ModifiedAt)
	}
	
	// Get job status
	if deployment.JobSummary != nil && deployment.JobSummary.Status != nil {
		d.Set("status", *deployment.JobSummary.Status)
	} else {
		d.Set("status", "CREATED")
	}

	return nil
}

func resourceAliCloudFlinkDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deploymentId := d.Id()
	namespace := d.Get("namespace").(string)
	d.Partial(true)

	// Create update request with the correct structure
	request := &ververica.UpdateDeploymentRequest{}
	
	// Create a Deployment object for the Body field
	deployment := &ververica.Deployment{}
	update := false

	// Set the deployment ID
	deployment.DeploymentId = &deploymentId

	if d.HasChange("job_name") || d.HasChange("deployment_name") {
		deploymentName := d.Get("deployment_name").(string)
		if deploymentName == "" {
			deploymentName = d.Get("job_name").(string)
		}
		deployment.Name = &deploymentName
		update = true
	}

	if d.HasChange("jar_uri") || d.HasChange("entry_class") || d.HasChange("program_args") {
		artifact := &ververica.Artifact{}
		artifact.Kind = stringPointer("JAR")
		
		if d.HasChange("jar_uri") {
			jarUri := d.Get("jar_uri").(string)
			artifact.JarUri = &jarUri
		}
		
		if d.HasChange("entry_class") {
			entryClass := d.Get("entry_class").(string)
			artifact.EntryClass = &entryClass
		}
		
		if d.HasChange("program_args") {
			programArgs := d.Get("program_args").(string)
			artifact.ProgramArgs = &programArgs
		}
		
		deployment.Artifact = artifact
		update = true
	}

	if d.HasChange("parallelism") {
		parallelism := d.Get("parallelism").(int)
		pInt64 := int64(parallelism)
		streamingResourceSetting := &ververica.StreamingResourceSetting{}
		streamingResourceSetting.Parallelism = &pInt64
		deployment.StreamingResourceSetting = streamingResourceSetting
		update = true
	}

	if update {
		// Set the deployment as the body of the request
		request.Body = deployment
		
		_, err := flinkService.UpdateDeployment(&namespace, &deploymentId, request)
		if err != nil {
			return WrapError(err)
		}

		 // Create an adapter function to convert ResourceStateRefreshFunc to resource.StateRefreshFunc
		refreshFunc := func() (interface{}, string, error) {
			return flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{})()
		}

		// Wait for update to complete
		stateConf := resource.StateChangeConf{
			Pending:      []string{},
			Target:       []string{"CREATED"},
			Refresh:      refreshFunc,
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			Delay:        5 * time.Second,
			PollInterval: 5 * time.Second,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	d.Partial(false)
	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deploymentId := d.Id()
	namespace := d.Get("namespace").(string)

	// Check if there's an active job that needs to be stopped first
	response, err := flinkService.GetDeployment(&namespace, &deploymentId)
	if err != nil {
		if IsExpectedErrors(err, []string{"EntityNotExist.Deployment"}) {
			return nil
		}
		return WrapError(err)
	}

	// Check if there's an active job running
	if response.Body.Data != nil && response.Body.Data.JobSummary != nil && 
	   response.Body.Data.JobSummary.Status != nil && *response.Body.Data.JobSummary.Status == "RUNNING" {
		// Get the job ID from deployment
		jobId := ""
		if response.Body.Data.JobSummary.JobId != nil {
			jobId = *response.Body.Data.JobSummary.JobId
		}
		
		if jobId != "" {
			// If the job is running, stop it first
			stopRequest := &ververica.StopJobRequest{
				SavepointPath: nil, // No savepoint path specified
			}
			
			_, err := flinkService.StopJob(&namespace, &jobId, stopRequest)
			if err != nil && !IsExpectedErrors(err, []string{"JobNotFound", "JobNotRunning"}) {
				return WrapError(err)
			}
			
			// Create an adapter function to convert ResourceStateRefreshFunc to resource.StateRefreshFunc for job state
			refreshJobFunc := func() (interface{}, string, error) {
				return flinkService.FlinkJobStateRefreshFunc(namespace, jobId, []string{})()
			}
			
			// Wait for the job to stop
			jobStateConf := resource.StateChangeConf{
				Pending:      []string{"RUNNING", "CANCELLING"},
				Target:       []string{"CANCELED", "FINISHED", "FAILED"},
				Refresh:      refreshJobFunc,
				Timeout:      d.Timeout(schema.TimeoutDelete),
				Delay:        5 * time.Second,
				PollInterval: 5 * time.Second,
			}
			
			if _, err := jobStateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}
	}

	// Now delete the deployment
	_, err = flinkService.DeleteDeployment(&namespace, &deploymentId)
	if err != nil {
		if !IsExpectedErrors(err, []string{"EntityNotExist.Deployment"}) {
			return WrapError(err)
		}
	}

	// Create an adapter function to convert ResourceStateRefreshFunc to resource.StateRefreshFunc
	refreshDeploymentFunc := func() (interface{}, string, error) {
		return flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"EntityNotExist.Deployment"})()
	}
	
	// Wait for the deployment to be deleted
	deploymentStateConf := resource.StateChangeConf{
		Pending:      []string{},
		Target:       []string{},
		Refresh:      refreshDeploymentFunc,
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        5 * time.Second,
		PollInterval: 5 * time.Second,
	}
	
	if _, err := deploymentStateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}