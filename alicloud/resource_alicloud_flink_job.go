package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
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
			"parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The parallelism level for the job.",
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
	jobName := d.Get("job_name").(string)
	allowNonRestoredState := d.Get("allow_non_restored_state").(bool)
	savepointPath := d.Get("savepoint_path").(string)
	parallelism := d.Get("parallelism").(int)

	// Build job request - using the Job struct from cws-lib-go with correct field names
	request := &aliyunAPI.Job{
		Workspace:      workspaceId,
		Namespace:      namespaceName,
		DeploymentId:   deploymentId,
		DeploymentName: jobName,
	}

	// Handle restore strategy if savepoint path is provided
	if savepointPath != "" {
		request.RestoreStrategy = &aliyunAPI.DeploymentRestoreStrategy{
			Kind:                  "SAVEPOINT",
			AllowNonRestoredState: allowNonRestoredState,
			SavepointId:           savepointPath,
		}
	}

	// Handle streaming resource setting for parallelism
	if parallelism > 0 {
		request.StreamingResourceSetting = &aliyunAPI.StreamingResourceSetting{
			ResourceSettingMode: "BASIC",
			BasicResourceSetting: &aliyunAPI.BasicResourceSetting{
				Parallelism: parallelism,
			},
		}
	}

	// Start job using FlinkService
	job, err := flinkService.StartJobWithParams(namespaceName, request)
	if err != nil {
		return WrapError(err)
	}

	// Set composite ID: namespace:jobId
	d.SetId(fmt.Sprintf("%s:%s", namespaceName, job.JobId))

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

	// Set attributes using correct field names from cws-lib-go Job type
	d.Set("workspace_id", job.Workspace)
	d.Set("namespace_name", job.Namespace)
	d.Set("deployment_id", job.DeploymentId)
	d.Set("job_name", job.DeploymentName)
	d.Set("job_id", job.JobId)

	// Handle job status from job status field
	if job.Status != nil {
		d.Set("status", job.Status.CurrentJobStatus)
	}

	// Handle parallelism from streaming resource setting
	if job.StreamingResourceSetting != nil &&
		job.StreamingResourceSetting.BasicResourceSetting != nil {
		d.Set("parallelism", job.StreamingResourceSetting.BasicResourceSetting.Parallelism)
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

	// Parse composite ID to get namespace and jobId
	namespace, jobId, err := parseJobId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Stop job with savepoint
	err = flinkService.StopJob(namespace, jobId, true)
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
