package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkWorkspaceCreate,
		Read:   resourceAliCloudFlinkWorkspaceRead,
		Delete: resourceAliCloudFlinkWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Flink instance.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the resource group.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The zone ID where the Flink instance is located.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VPC ID of the Flink instance.",
			},
			"vswitch_ids": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The IDs of the vSwitches for the Flink instance.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of the security group.",
			},
			"architecture_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "X86",
				ValidateFunc: validation.StringInSlice([]string{"X86", "ARM"}, false),
				Description:  "The architecture type of the Flink instance.",
			},
			"auto_renew": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether the instance automatically renews.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "POST",
				ValidateFunc: validation.StringInSlice([]string{"POST", "PRE"}, false),
				Description:  "The billing method of the instance.",
			},
			"duration": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     1,
				Description: "The subscription duration.",
			},
			"pricing_cycle": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "Month",
				Description: "The billing cycle for Subscription instances.",
			},
			"extra": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Additional configuration for the instance.",
			},
			"ha": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										ForceNew:    true,
										Description: "CPU specifications for HA resources.",
									},
									"memory": {
										Type:        schema.TypeInt,
										Required:    true,
										ForceNew:    true,
										Description: "Memory specifications for HA resources.",
									},
								},
							},
							Description: "HA resource specifications.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Required:    true,
							ForceNew:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The IDs of the vSwitches for high availability.",
						},
						"zone_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The zone ID for high availability.",
						},
					},
				},
				Description: "High availability configuration.",
			},
			"monitor_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ARMS",
				ValidateFunc: validation.StringInSlice([]string{"ARMS", "TAIHAO"}, true),
				Description:  "The monitoring type of the instance.",
			},
			"promotion_code": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The promotion code.",
			},
			"use_promotion_code": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Whether to use promotion code.",
			},
			"storage": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oss_bucket": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The OSS bucket name for the Flink instance.",
						},
					},
				},
				Description: "Storage configuration of oss bucket for the Flink instance.",
			},
			"resource": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
							Description: "CPU units in millicores.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
							Description: "Memory in MB.",
						},
					},
				},
				Description: "Resource specifications for the Flink instance.",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource ID of the Flink workspace instance.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create workspace request using cws-lib-go types
	workspaceRequest := &aliyunFlinkAPI.Workspace{
		Name:            d.Get("name").(string),
		ResourceGroupId: d.Get("resource_group_id").(string),
		ZoneId:          d.Get("zone_id").(string),
		VpcId:           d.Get("vpc_id").(string),
		Region:          client.RegionId,
	}

	// Handle vswitch_ids
	if vswitchIds := d.Get("vswitch_ids").([]interface{}); len(vswitchIds) > 0 {
		workspaceRequest.VSwitchIds = make([]string, len(vswitchIds))
		for i, v := range vswitchIds {
			workspaceRequest.VSwitchIds[i] = v.(string)
		}
	}

	// Handle security_group_id
	if sgId, ok := d.GetOk("security_group_id"); ok {
		if workspaceRequest.SecurityGroupInfo == nil {
			workspaceRequest.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{}
		}
		workspaceRequest.SecurityGroupInfo.SecurityGroupId = sgId.(string)
	}

	// Handle architecture_type
	if archType, ok := d.GetOk("architecture_type"); ok {
		workspaceRequest.ArchitectureType = archType.(string)
	}

	// Handle charge_type
	if chargeType, ok := d.GetOk("charge_type"); ok {
		workspaceRequest.ChargeType = chargeType.(string)
	}

	// Handle resource configuration
	if resourceList := d.Get("resource").([]interface{}); len(resourceList) > 0 {
		resourceMap := resourceList[0].(map[string]interface{})
		workspaceRequest.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{
			Cpu:      float64(resourceMap["cpu"].(int)),
			MemoryGB: float64(resourceMap["memory"].(int)),
		}
	}

	// Handle storage configuration
	if storageList := d.Get("storage").([]interface{}); len(storageList) > 0 {
		storageMap := storageList[0].(map[string]interface{})
		workspaceRequest.Storage = &aliyunFlinkAPI.Storage{
			Oss: &aliyunFlinkAPI.OSSStorage{
				Bucket: storageMap["oss_bucket"].(string),
			},
		}
	}

	// Handle HA configuration
	if haList := d.Get("ha").([]interface{}); len(haList) > 0 {
		haMap := haList[0].(map[string]interface{})

		// Set high availability flag
		workspaceRequest.HighAvailability = &aliyunFlinkAPI.HighAvailability{
			Enabled: true,
			ZoneId:  haMap["zone_id"].(string),
		}

		// Handle HA vswitch IDs
		if haVswitchIds := haMap["vswitch_ids"].([]interface{}); len(haVswitchIds) > 0 {
			workspaceRequest.HighAvailability.VSwitchIds = make([]string, len(haVswitchIds))
			for i, v := range haVswitchIds {
				workspaceRequest.HighAvailability.VSwitchIds[i] = v.(string)
			}
		}

		// Handle HA resource specs
		if resourceList := haMap["resource"].([]interface{}); len(resourceList) > 0 {
			resourceMap := resourceList[0].(map[string]interface{})
			workspaceRequest.HighAvailability.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{
				Cpu:      float64(resourceMap["cpu"].(int)),
				MemoryGB: float64(resourceMap["memory"].(int)),
			}
		}
	}

	// Create the workspace with retry mechanism
	var workspace *aliyunFlinkAPI.Workspace
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateInstance(workspaceRequest)
		if err != nil {
			if IsNotFoundError(err) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		workspace = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_workspace", "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	if workspace == nil || workspace.Id == "" {
		return WrapError(Error("Failed to get instance ID from workspace"))
	}

	d.SetId(workspace.Id)

	// Wait for the instance to be in running state using service layer function
	if err := flinkService.WaitForWorkspaceStarting(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkWorkspaceRead(d, meta)
}

func resourceAliCloudFlinkWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, err := flinkService.DescribeFlinkWorkspace(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_workspace DescribeFlinkWorkspace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes from workspace
	d.Set("name", workspace.Name)
	d.Set("resource_group_id", workspace.ResourceGroupId)

	// Ensure zone_id is always set to prevent forces replacement
	if workspace.ZoneId != "" {
		d.Set("zone_id", workspace.ZoneId)
	} else {
		// If API doesn't return zone_id, preserve the configured value
		if configuredZoneId := d.Get("zone_id").(string); configuredZoneId != "" {
			d.Set("zone_id", configuredZoneId)
		}
	}

	d.Set("vpc_id", workspace.VpcId)

	// Handle vswitch_ids with proper type conversion
	if workspace.VSwitchIds != nil && len(workspace.VSwitchIds) > 0 {
		d.Set("vswitch_ids", workspace.VSwitchIds)
	} else {
		// Preserve configured vswitch_ids if API doesn't return them
		if configuredVSwitchIds := d.Get("vswitch_ids").([]interface{}); len(configuredVSwitchIds) > 0 {
			// Keep the configured values to prevent plan changes
		}
	}

	// Handle security group from SecurityGroupInfo structure
	if workspace.SecurityGroupInfo != nil && workspace.SecurityGroupInfo.SecurityGroupId != "" {
		d.Set("security_group_id", workspace.SecurityGroupInfo.SecurityGroupId)
	}

	// Set fields that are returned by the API with fallback to configured values
	if workspace.ArchitectureType != "" {
		d.Set("architecture_type", workspace.ArchitectureType)
	}
	if workspace.ChargeType != "" {
		d.Set("charge_type", workspace.ChargeType)
	}
	if workspace.MonitorType != "" {
		d.Set("monitor_type", workspace.MonitorType)
	}

	// Always set resource_id if available
	if workspace.ResourceId != "" {
		d.Set("resource_id", workspace.ResourceId)
	}

	// Set resource configuration
	if workspace.ResourceSpec != nil {
		resourceConfig := map[string]interface{}{
			"cpu":    int(workspace.ResourceSpec.Cpu),
			"memory": int(workspace.ResourceSpec.MemoryGB),
		}
		d.Set("resource", []interface{}{resourceConfig})
	}

	// Set storage configuration
	if workspace.Storage != nil && workspace.Storage.Oss != nil {
		storageConfig := map[string]interface{}{
			"oss_bucket": workspace.Storage.Oss.Bucket,
		}
		d.Set("storage", []interface{}{storageConfig})
	}

	// Set HA configuration
	if workspace.HighAvailability != nil && workspace.HighAvailability.Enabled {
		haConfig := map[string]interface{}{
			"zone_id":     workspace.HighAvailability.ZoneId,
			"vswitch_ids": workspace.HighAvailability.VSwitchIds,
		}

		// Add resource info if available
		if workspace.HighAvailability.ResourceSpec != nil {
			haResourceSpec := map[string]interface{}{
				"cpu":    int(workspace.HighAvailability.ResourceSpec.Cpu),
				"memory": int(workspace.HighAvailability.ResourceSpec.MemoryGB),
			}
			haConfig["resource"] = []interface{}{haResourceSpec}
		}

		d.Set("ha", []interface{}{haConfig})
	} else {
		// Set empty HA config if not enabled
		d.Set("ha", []interface{}{})
	}

	// Handle input-only fields that are not returned by the API
	// These fields are used only during creation and are not returned by DescribeInstances
	// We need to preserve their configured values to avoid terraform plan showing changes

	// For auto_renew: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("auto_renew"); !ok {
		// Only set default if not already configured
		d.Set("auto_renew", true)
	}

	// For duration: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("duration"); !ok {
		// Only set default if not already configured
		d.Set("duration", 1)
	}

	// For pricing_cycle: preserve the configured value since it's not returned by the API
	if _, ok := d.GetOk("pricing_cycle"); !ok {
		// Only set default if not already configured
		d.Set("pricing_cycle", "Month")
	}

	// For extra, promotion_code, use_promotion_code: these are input-only fields
	// They don't need to be set here as they are only used during creation

	return nil
}

func resourceAliCloudFlinkWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteInstance(d.Id())
	if err != nil {
		if !IsNotFoundError(err) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteInstance", AlibabaCloudSdkGoERROR)
		}
		// IsNotFoundError means deletion was successful or resource already gone
	}

	// Wait for the workspace to be completely deleted using service layer function
	if err := flinkService.WaitForWorkspaceDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
