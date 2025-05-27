package alicloud

import (
	"time"

	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"namespace_id": {
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

	namespace := d.Get("namespace_id").(string)
	workspaceId := d.Get("workspace_id").(string)
	jarUri := d.Get("jar_uri").(string)

	jobName := d.Get("job_name").(string)

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

	deployment.Artifact = artifact

	// Set parallelism

	streamingResourceSetting := &ververica.StreamingResourceSetting{}

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
	namespace := d.Get("namespace_id").(string)

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
	d.Set("namespace_id", *deployment.Namespace)

	if deployment.CreatedAt != nil {
		d.Set("create_time", *deployment.CreatedAt)
	}

	if deployment.ModifiedAt != nil {
		d.Set("update_time", *deployment.ModifiedAt)
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
	namespace := d.Get("namespace_id").(string)
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

		deployment.Artifact = artifact
		update = true
	}

	if d.HasChange("parallelism") {
		streamingResourceSetting := &ververica.StreamingResourceSetting{}
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
	namespace := d.Get("namespace_id").(string)

	// Check if there's an active job that needs to be stopped first
	if err != nil {
		if IsExpectedErrors(err, []string{"EntityNotExist.Deployment"}) {
			return nil
		}
		return WrapError(err)
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
