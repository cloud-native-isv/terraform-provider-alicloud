package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Flink instance.",
				ForceNew:    true,
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
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
				Description: "The IDs of the vSwitches for the Flink instance.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the security group.",
				ForceNew:    true,
			},
			"architecture_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "X86",
				Description:  "The architecture type of the Flink instance.",
				ValidateFunc: validation.StringInSlice([]string{"X86", "ARM"}, false),
				ForceNew:     true,
			},
			"auto_renew": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the instance automatically renews.",
				ForceNew:    true,
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POST",
				Description:  "The billing method of the instance.",
				ValidateFunc: validation.StringInSlice([]string{"POST", "PRE"}, false),
				ForceNew:     true,
			},
			"duration": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The subscription duration.",
				ForceNew:    true,
			},
			"pricing_cycle": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Month",
				Description: "The billing cycle for Subscription instances.",
				ForceNew:    true,
			},
			"extra": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional configuration for the instance.",
				ForceNew:    true,
			},
			"ha": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "CPU specifications for HA resources.",
									},
									"memory": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Memory specifications for HA resources.",
									},
								},
							},
							Description: "HA resource specifications.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The IDs of the vSwitches for high availability.",
						},
						"zone_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The zone ID for high availability.",
						},
					},
				},
				Description: "High availability configuration.",
				ForceNew:    true,
			},
			"monitor_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The monitoring type of the instance.",
				ForceNew:     true,
				Default:      "ARMS",
				ValidateFunc: validation.StringInSlice([]string{"ARMS", "TAIHAO"}, true),
			},
			"promotion_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The promotion code.",
				ForceNew:    true,
			},
			"use_promotion_code": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use promotion code.",
				ForceNew:    true,
			},
			"storage": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oss_bucket": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The OSS bucket name for the Flink instance.",
						},
					},
				},
				Description: "Storage configuration of oss bucket for the Flink instance.",
				ForceNew:    true,
			},
			"resource": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "CPU units in millicores.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Memory in MB.",
						},
					},
				},
				Description: "Resource specifications for the Flink instance.",
				ForceNew:    true,
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
	workspaceRequest := &aliyunAPI.Workspace{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		ResourceGroupID: d.Get("resource_group_id").(string),
		ZoneID:          d.Get("zone_id").(string),
		VPCID:           d.Get("vpc_id").(string),
		Region:          client.RegionId,
	}

	// Handle vswitch_ids
	if vswitchIds := d.Get("vswitch_ids").([]interface{}); len(vswitchIds) > 0 {
		workspaceRequest.VSwitchIDs = make([]string, len(vswitchIds))
		for i, v := range vswitchIds {
			workspaceRequest.VSwitchIDs[i] = v.(string)
		}
	}

	// Initialize Network structure and set VPCID and VSwitchIDs
	workspaceRequest.Network = &aliyunAPI.WorkspaceNetwork{
		VPCID:      d.Get("vpc_id").(string),
		VSwitchIDs: workspaceRequest.VSwitchIDs,
	}

	// Handle security_group_id
	if sgId, ok := d.GetOk("security_group_id"); ok {
		workspaceRequest.Network.SecurityGroupID = sgId.(string)
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
		workspaceRequest.ResourceSpec = &aliyunAPI.ResourceSpec{
			Cpu:      float64(resourceMap["cpu"].(int)),
			MemoryGB: float64(resourceMap["memory"].(int)),
		}
	}

	// Handle storage configuration
	if storageList := d.Get("storage").([]interface{}); len(storageList) > 0 {
		storageMap := storageList[0].(map[string]interface{})
		workspaceRequest.Storage = &aliyunAPI.Storage{
			Oss: &aliyunAPI.OssConfig{
				Bucket: storageMap["oss_bucket"].(string),
			},
		}
	}

	// Handle HA configuration
	if haList := d.Get("ha").([]interface{}); len(haList) > 0 {
		haMap := haList[0].(map[string]interface{})

		// Set high availability flag
		workspaceRequest.HA = &aliyunAPI.HighAvailability{
			Enabled: true,
			ZoneID:  haMap["zone_id"].(string),
		}

		// Handle HA vswitch IDs
		if haVswitchIds := haMap["vswitch_ids"].([]interface{}); len(haVswitchIds) > 0 {
			workspaceRequest.HA.VSwitchIDs = make([]string, len(haVswitchIds))
			for i, v := range haVswitchIds {
				workspaceRequest.HA.VSwitchIDs[i] = v.(string)
			}
		}

		// Handle HA resource specs
		if resourceList := haMap["resource"].([]interface{}); len(resourceList) > 0 {
			resourceMap := resourceList[0].(map[string]interface{})
			workspaceRequest.HA.ResourceSpec = &aliyunAPI.ResourceSpec{
				Cpu:      float64(resourceMap["cpu"].(int)),
				MemoryGB: float64(resourceMap["memory"].(int)),
			}
		}
	}

	var workspace *aliyunAPI.Workspace
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateInstance(workspaceRequest)
		if err != nil {
			if IsExpectedErrors(err, []string{"903021"}) {
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

	if workspace == nil || workspace.ID == "" || workspace.ResourceID == "" {
		return WrapError(Error("Failed to get instance ID from workspace"))
	}

	d.SetId(workspace.ID)

	// Set resource_id after creation
	d.Set("resource_id", workspace.ResourceID)

	// Wait for the instance to be in running state
	stateConf := resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"RUNNING"},
		Refresh:    flinkService.FlinkWorkspaceStateRefreshFunc(d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

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
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_workspace DescribeFlinkWorkspace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes from workspace workspace
	d.Set("name", workspace.Name)
	d.Set("description", workspace.Description)
	d.Set("resource_group_id", workspace.ResourceGroupID)
	d.Set("zone_id", workspace.ZoneID)
	d.Set("vpc_id", workspace.VPCID)
	d.Set("vswitch_ids", workspace.VSwitchIDs)

	// Handle security group from Network structure
	if workspace.Network != nil && workspace.Network.SecurityGroupID != "" {
		d.Set("security_group_id", workspace.Network.SecurityGroupID)
	}

	d.Set("architecture_type", workspace.ArchitectureType)
	d.Set("charge_type", workspace.ChargeType)
	d.Set("status", workspace.Status)

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

	// Set resource_id
	d.Set("resource_id", workspace.ResourceID)

	// Set HA configuration
	if workspace.HA != nil {
		haConfig := map[string]interface{}{
			"zone_id":     workspace.HA.ZoneID,
			"vswitch_ids": workspace.HA.VSwitchIDs,
		}

		// Add resource info
		if workspace.HA.ResourceSpec != nil {
			haResourceSpec := map[string]interface{}{
				"cpu":    int(workspace.HA.ResourceSpec.Cpu),
				"memory": int(workspace.HA.ResourceSpec.MemoryGB),
			}
			haConfig["resource"] = []interface{}{haResourceSpec}
		}

		d.Set("ha", []interface{}{haConfig})
	}

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
		if !IsExpectedErrors(err, []string{"903021"}) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteInstance", AlibabaCloudSdkGoERROR)
		}
		// Error 903021 means deletion was successful, continue to wait for state change
	}

	// Use a customized state refresh function that handles the not found case properly
	stateConf := &resource.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			// Check if the workspace still exists
			workspace, err := flinkService.DescribeFlinkWorkspace(d.Id())
			if err != nil {
				if NotFoundError(err) {
					// Resource is gone, which is what we want
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			// If we can still get the workspace, it's still being deleted
			return workspace, workspace.Status, nil
		},
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
