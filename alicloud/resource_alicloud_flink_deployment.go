package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
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
	foasService := FoasService{client}

	workspaceId := d.Get("workspace_id").(string)
	namespace := d.Get("namespace").(string)
	jobName := d.Get("job_name").(string)

	request := foasconsole.CreateDeployJobRequest()
	request.WorkspaceId = workspaceId
	request.Namespace = namespace
	request.JobName = jobName
	request.EntryClass = d.Get("entry_class").(string)
	request.JarUri = d.Get("jar_uri").(string)

	if v, ok := d.GetOk("parallelism"); ok {
		request.Parallelism = v.(int)
	}

	if v, ok := d.GetOk("jar_artifact_name"); ok {
		request.JarArtifactName = v.(string)
	}

	if v, ok := d.GetOk("program_args"); ok {
		request.ProgramArgs = v.(string)
	}

	if v, ok := d.GetOk("deployment_name"); ok {
		request.DeploymentName = v.(string)
	} else {
		// Generate a default deployment name if not provided
		request.DeploymentName = fmt.Sprintf("%s-%s", jobName, time.Now().Format("20060102150405"))
	}

	var response *foasconsole.DeployJobResponse
	var err error

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		raw, err := foasService.DeployJob(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationFailed", "ServiceUnavailable"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = raw
		return nil
	})
	
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment", request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", workspaceId, namespace, response.DeploymentId))
	d.Set("deployment_id", response.DeploymentId)

	// Wait for the deployment to be running
	stateConf := BuildStateConf([]string{"CREATING", "STARTING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, foasService.FlinkDeploymentStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	workspaceId := parts[0]
	namespace := parts[1]
	deploymentId := parts[2]

	request := foasconsole.CreateGetDeploymentRequest()
	request.WorkspaceId = workspaceId
	request.Namespace = namespace
	request.DeploymentId = deploymentId

	var deployment *foasconsole.GetDeploymentResponseDeployment
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err := foasService.GetDeployment(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ResourceNotFound", "ServiceUnavailable"}) {
				if NotFoundError(err) {
					log.Printf("[DEBUG] Resource alicloud_flink_deployment foasService.GetDeployment Failed!!! %s", err)
					d.SetId("")
					return nil
				}
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		deployment = raw.Deployment
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	if deployment == nil {
		d.SetId("")
		return nil
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace", namespace)
	d.Set("deployment_id", deploymentId)
	d.Set("job_name", deployment.JobName)
	d.Set("deployment_name", deployment.DeploymentName)
	d.Set("entry_class", deployment.EntryClass)
	d.Set("parallelism", deployment.Parallelism)
	d.Set("jar_uri", deployment.JarUri)
	d.Set("jar_artifact_name", deployment.JarArtifactName)
	d.Set("program_args", deployment.ProgramArgs)
	d.Set("status", deployment.Status)
	d.Set("create_time", deployment.CreateTime)
	d.Set("update_time", deployment.UpdateTime)

	return nil
}

func resourceAliCloudFlinkDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}
	
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	workspaceId := parts[0]
	namespace := parts[1]
	deploymentId := parts[2]

	d.Partial(true)
	
	update := false
	request := foasconsole.CreateUpdateDeploymentRequest()
	request.WorkspaceId = workspaceId
	request.Namespace = namespace
	request.DeploymentId = deploymentId

	if d.HasChange("job_name") {
		request.JobName = d.Get("job_name").(string)
		update = true
	}

	if d.HasChange("entry_class") {
		request.EntryClass = d.Get("entry_class").(string)
		update = true
	}

	if d.HasChange("parallelism") {
		request.Parallelism = d.Get("parallelism").(int)
		update = true
	}

	if d.HasChange("jar_uri") {
		request.JarUri = d.Get("jar_uri").(string)
		update = true
	}

	if d.HasChange("jar_artifact_name") {
		request.JarArtifactName = d.Get("jar_artifact_name").(string)
		update = true
	}

	if d.HasChange("program_args") {
		request.ProgramArgs = d.Get("program_args").(string)
		update = true
	}

	if d.HasChange("deployment_name") {
		request.DeploymentName = d.Get("deployment_name").(string)
		update = true
	}

	if update {
		var response *foasconsole.UpdateDeploymentResponse
		var err error
		
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			raw, err := foasService.UpdateDeployment(request)
			if err != nil {
				if IsExpectedErrors(err, []string{"OperationFailed", "ServiceUnavailable"}) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			response = raw
			return nil
		})
		
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), request.GetActionName(), AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{"UPDATING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, foasService.FlinkDeploymentStateRefreshFunc(d.Id()))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		if d.HasChange("job_name") {
			d.SetPartial("job_name")
		}
		if d.HasChange("entry_class") {
			d.SetPartial("entry_class")
		}
		if d.HasChange("parallelism") {
			d.SetPartial("parallelism")
		}
		if d.HasChange("jar_uri") {
			d.SetPartial("jar_uri")
		}
		if d.HasChange("jar_artifact_name") {
			d.SetPartial("jar_artifact_name")
		}
		if d.HasChange("program_args") {
			d.SetPartial("program_args")
		}
		if d.HasChange("deployment_name") {
			d.SetPartial("deployment_name")
		}
	}

	d.Partial(false)
	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	workspaceId := parts[0]
	namespace := parts[1]
	deploymentId := parts[2]

	request := foasconsole.CreateDeleteDeploymentRequest()
	request.WorkspaceId = workspaceId
	request.Namespace = namespace
	request.DeploymentId = deploymentId

	var response *foasconsole.DeleteDeploymentResponse
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		raw, err := foasService.DeleteDeployment(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationFailed", "ServiceUnavailable"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = raw
		return nil
	})

	if err != nil {
		if IsExpectedErrors(err, []string{"ResourceNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	stateConf := BuildStateConf([]string{"DELETING", "STOPPING"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, foasService.FlinkDeploymentStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}