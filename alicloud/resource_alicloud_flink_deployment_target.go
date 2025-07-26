package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkDeploymentTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentTargetCreate,
		Read:   resourceAliCloudFlinkDeploymentTargetRead,
		Update: resourceAliCloudFlinkDeploymentTargetUpdate,
		Delete: resourceAliCloudFlinkDeploymentTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"quota": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"limit": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"memory_gb": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"disk": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"used": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"memory_gb": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"disk": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudFlinkDeploymentTargetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	deploymentTargetService := NewFlinkDeploymentTargetService(client)

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	targetName := d.Get("name").(string)

	target := &flink.DeploymentTarget{
		Name:      targetName,
		Namespace: namespaceName,
	}

	// Set quota if provided
	if v, ok := d.GetOk("quota"); ok {
		target.Quota = expandResourceQuota(v.([]interface{}))
	}

	_, err := deploymentTargetService.CreateFlinkDeploymentTarget(workspaceId, namespaceName, target)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_target", "CreateDeploymentTarget", AlibabaCloudSdkGoERROR)
	}

	d.SetId(formatDeploymentTargetId(workspaceId, namespaceName, targetName))

	return resourceAliCloudFlinkDeploymentTargetRead(d, meta)
}

func resourceAliCloudFlinkDeploymentTargetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	deploymentTargetService := NewFlinkDeploymentTargetService(client)

	object, err := deploymentTargetService.DescribeFlinkDeploymentTarget(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	workspaceId, namespaceName, targetName, err := parseDeploymentTargetId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("name", targetName)

	if object.Quota != nil {
		d.Set("quota", flattenResourceQuota(object.Quota))
	}

	return nil
}

func resourceAliCloudFlinkDeploymentTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	deploymentTargetService := NewFlinkDeploymentTargetService(client)

	workspaceId, namespaceName, targetName, err := parseDeploymentTargetId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	updateRequest := &flink.DeploymentTarget{
		Name:      targetName,
		Namespace: namespaceName,
	}

	update := false

	if d.HasChange("quota") {
		if v, ok := d.GetOk("quota"); ok {
			updateRequest.Quota = expandResourceQuota(v.([]interface{}))
		}
		update = true
	}

	if update {
		_, err := deploymentTargetService.UpdateFlinkDeploymentTarget(workspaceId, namespaceName, updateRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDeploymentTarget", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudFlinkDeploymentTargetRead(d, meta)
}

func resourceAliCloudFlinkDeploymentTargetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	deploymentTargetService := NewFlinkDeploymentTargetService(client)

	workspaceId, namespaceName, targetName, err := parseDeploymentTargetId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = deploymentTargetService.DeleteFlinkDeploymentTarget(workspaceId, namespaceName, targetName)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDeploymentTarget.NotFound", "DeploymentTargetNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeploymentTarget", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// Helper functions for expanding and flattening schema data

func expandResourceQuota(configured []interface{}) *flink.ResourceQuota {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	quota := &flink.ResourceQuota{}

	if v, ok := raw["limit"]; ok {
		quota.Limit = expandResourceSpec(v.([]interface{}))
	}

	if v, ok := raw["used"]; ok {
		quota.Used = expandResourceSpec(v.([]interface{}))
	}

	return quota
}

func expandResourceSpec(configured []interface{}) *flink.ResourceSpec {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	spec := &flink.ResourceSpec{}

	if v, ok := raw["cpu"]; ok && v.(float64) > 0 {
		spec.Cpu = v.(float64)
	}

	if v, ok := raw["memory_gb"]; ok && v.(float64) > 0 {
		spec.MemoryGB = v.(float64)
	}

	if v, ok := raw["disk"]; ok && v.(int) > 0 {
		spec.Disk = int32(v.(int))
	}

	return spec
}

func flattenResourceQuota(quota *flink.ResourceQuota) []interface{} {
	if quota == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}

	if quota.Limit != nil {
		result["limit"] = flattenResourceSpec(quota.Limit)
	}

	if quota.Used != nil {
		result["used"] = flattenResourceSpec(quota.Used)
	}

	return []interface{}{result}
}

func flattenResourceSpec(spec *flink.ResourceSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"cpu":       spec.Cpu,
			"memory_gb": spec.MemoryGB,
			"disk":      int(spec.Disk),
		},
	}
}
