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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parallelism": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"max_parallelism": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"execution_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"session_cluster_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"with_savepoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"restore_strategy": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"NONE", "LATEST_SAVEPOINT", "FROM_SAVEPOINT", "LATEST_STATE"}, false),
						},
						"savepoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"local_variables": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"flink_conf": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_flink_conf": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
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

	job, err := flinkService.StartJob(params)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, job.JobId))

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
		if NotFoundError(err) {
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
		if deploymentErr != nil && NotFoundError(deploymentErr) {
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
		if NotFoundError(stopErr) {
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
		if NotFoundError(err) {
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
