package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
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
			"namespace_name": {
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

	// Create deployment request using cws-lib-go types
	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	request := &aliyunAPI.Deployment{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      d.Get("job_name").(string),
	}

	// Handle artifact configuration
	if jarUri, ok := d.GetOk("jar_uri"); ok {
		request.Artifact = &aliyunAPI.Artifact{
			Kind: "JAR",
			JarArtifact: &aliyunAPI.JarArtifact{
				JarUri: jarUri.(string),
			},
		}

		if entryClass, ok := d.GetOk("entry_class"); ok {
			request.Artifact.JarArtifact.EntryClass = entryClass.(string)
		}
	}

	// Handle streaming resource setting for parallelism
	if parallelism, ok := d.GetOk("parallelism"); ok {
		request.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
			ResourceSettingMode: "BASIC",
			BasicResourceSetting: &aliyunAPI.BasicResourceSetting{
				Parallelism: parallelism.(int),
			},
		}
	}

	// Handle Flink configuration
	if flinkConf := d.Get("flink_conf").(map[string]interface{}); len(flinkConf) > 0 {
		request.FlinkConf = make(map[string]string)
		for k, v := range flinkConf {
			request.FlinkConf[k] = v.(string)
		}
	}

	// Create deployment
	var response *aliyunAPI.Deployment
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateDeployment(&namespaceName, request)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment", "CreateDeployment", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.DeploymentId == "" {
		return WrapError(Error("Failed to get deployment ID from response"))
	}

	// Set composite ID: namespace:deploymentId
	d.SetId(namespaceName + ":" + response.DeploymentId)

	// Wait for deployment creation to complete
	stateConf := resource.StateChangeConf{
		Pending:    []string{"CREATING", "STARTING"},
		Target:     []string{"CREATED", "RUNNING"},
		Refresh:    flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"FAILED"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_deployment GetDeployment Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Parse namespace and deployment ID from composite ID
	namespaceName, deploymentId, err := parseDeploymentId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Set attributes from deployment deployment using correct field names
	d.Set("workspace_id", deployment.Workspace)
	d.Set("namespace_name", namespaceName)
	d.Set("deployment_id", deploymentId)
	d.Set("job_name", deployment.Name)
	d.Set("create_time", deployment.CreatedAt)
	d.Set("update_time", deployment.ModifiedAt)

	// Handle artifact fields
	if deployment.Artifact != nil && deployment.Artifact.JarArtifact != nil {
		d.Set("jar_uri", deployment.Artifact.JarArtifact.JarUri)
		d.Set("entry_class", deployment.Artifact.JarArtifact.EntryClass)
	}

	// Handle parallelism from streaming resource setting
	if deployment.StreamingResourceSetting != nil &&
		deployment.StreamingResourceSetting.BasicResourceSetting != nil {
		d.Set("parallelism", deployment.StreamingResourceSetting.BasicResourceSetting.Parallelism)
	}

	// Handle job status from job summary
	if deployment.JobSummary != nil {
		d.Set("status", deployment.Status())
	}

	if deployment.FlinkConf != nil {
		flinkConf := make(map[string]interface{})
		for k, v := range deployment.FlinkConf {
			flinkConf[k] = v
		}
		d.Set("flink_conf", flinkConf)
	}

	return nil
}

func resourceAliCloudFlinkDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get current deployment
	deployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		return WrapError(err)
	}

	hasChanged := false

	// Update deployment properties if changed
	if d.HasChange("jar_uri") || d.HasChange("entry_class") || d.HasChange("parallelism") || d.HasChange("flink_conf") {
		if jarUri, ok := d.GetOk("jar_uri"); ok {
			if deployment.Artifact == nil {
				deployment.Artifact = &aliyunAPI.Artifact{
					Kind:        "JAR",
					JarArtifact: &aliyunAPI.JarArtifact{},
				}
			}
			if deployment.Artifact.JarArtifact == nil {
				deployment.Artifact.JarArtifact = &aliyunAPI.JarArtifact{}
			}
			deployment.Artifact.JarArtifact.JarUri = jarUri.(string)
		}

		if entryClass, ok := d.GetOk("entry_class"); ok {
			if deployment.Artifact == nil {
				deployment.Artifact = &aliyunAPI.Artifact{
					Kind:        "JAR",
					JarArtifact: &aliyunAPI.JarArtifact{},
				}
			}
			if deployment.Artifact.JarArtifact == nil {
				deployment.Artifact.JarArtifact = &aliyunAPI.JarArtifact{}
			}
			deployment.Artifact.JarArtifact.EntryClass = entryClass.(string)
		}

		if parallelism, ok := d.GetOk("parallelism"); ok {
			if deployment.StreamingResourceSetting == nil {
				deployment.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
					ResourceSettingMode:  "BASIC",
					BasicResourceSetting: &aliyunAPI.BasicResourceSetting{},
				}
			}
			if deployment.StreamingResourceSetting.BasicResourceSetting == nil {
				deployment.StreamingResourceSetting.BasicResourceSetting = &aliyunAPI.BasicResourceSetting{}
			}
			deployment.StreamingResourceSetting.BasicResourceSetting.Parallelism = parallelism.(int)
		}

		if flinkConf := d.Get("flink_conf").(map[string]interface{}); len(flinkConf) > 0 {
			deployment.FlinkConf = make(map[string]string)
			for k, v := range flinkConf {
				deployment.FlinkConf[k] = v.(string)
			}
		}

		hasChanged = true
	}

	if hasChanged {
		// Update deployment
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.UpdateDeployment(deployment)
			if err != nil {
				if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDeployment", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := resource.StateChangeConf{
			Pending:    []string{"UPDATING"},
			Target:     []string{"RUNNING", "CREATED"},
			Refresh:    flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"FAILED"}),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse namespace and deployment ID from composite ID
	namespaceName, deploymentId, err := parseDeploymentId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteDeployment(namespaceName, deploymentId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDeployment.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeployment", AlibabaCloudSdkGoERROR)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{},
		Refresh:    flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
