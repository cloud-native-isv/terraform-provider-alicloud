package alicloud

import (
	"time"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region of the instance.",
				ForceNew:    true,
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
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A mapping of tags to assign to the resource.",
				ForceNew:    true,
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

	request := foasconsole.CreateInstanceRequest{}

	// Populate mandatory fields
	request.InstanceName = tea.String(d.Get("name").(string))
	request.ResourceGroupId = tea.String(d.Get("resource_group_id").(string))
	request.VpcId = tea.String(d.Get("vpc_id").(string))
	request.ZoneId = tea.String(d.Get("zone_id").(string))
	request.Region = tea.String(d.Get("region").(string))

	// Handle storage configuration
	storage := d.Get("storage").([]interface{})
	if len(storage) > 0 {
		storageMap := storage[0].(map[string]interface{})
		if ossBucket, ok := storageMap["oss_bucket"].(string); ok {
			request.Storage = &foasconsole.CreateInstanceRequestStorage{
				Oss: &foasconsole.CreateInstanceRequestStorageOss{
					Bucket: tea.String(ossBucket),
				},
			}
		}
	}

	// Handle vswitch_ids as list
	vswitchIds := d.Get("vswitch_ids").([]interface{})
	request.VSwitchIds = make([]*string, 0, len(vswitchIds))
	for _, v := range vswitchIds {
		request.VSwitchIds = append(request.VSwitchIds, tea.String(v.(string)))
	}

	// Handle nested ResourceSpec
	resourceSpec := d.Get("resource").([]interface{})
	if len(resourceSpec) > 0 {
		resourceSpecMap := resourceSpec[0].(map[string]interface{})
		request.ResourceSpec = &foasconsole.CreateInstanceRequestResourceSpec{
			Cpu:      tea.Int32(int32(resourceSpecMap["cpu"].(int))),
			MemoryGB: tea.Int32(int32(resourceSpecMap["memory"].(int))),
		}
	}

	// Handle boolean fields with proper conversion
	if v, ok := d.GetOk("auto_renew"); ok {
		request.AutoRenew = tea.Bool(v.(bool))
	}
	if v, ok := d.GetOk("ha"); ok {
		request.Ha = tea.Bool(v.(bool))
	}
	if v, ok := d.GetOk("use_promotion_code"); ok {
		request.UsePromotionCode = tea.Bool(v.(bool))
	}

	// Handle optional fields
	// Note: Description field is not available in the API but kept in the schema for future compatibility
	// if v, ok := d.GetOk("description"); ok {
	//     request.Description = tea.String(v.(string))
	// }
	if v, ok := d.GetOk("duration"); ok {
		request.Duration = tea.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("promotion_code"); ok {
		request.PromotionCode = tea.String(v.(string))
	}
	// Note: SecurityGroupId field is not available in the API but kept in the schema for future compatibility
	// if v, ok := d.GetOk("security_group_id"); ok {
	//     request.SecurityGroupId = tea.String(v.(string))
	// }
	if v, ok := d.GetOk("architecture_type"); ok {
		request.ArchitectureType = tea.String(v.(string))
	}
	if v, ok := d.GetOk("monitor_type"); ok {
		request.MonitorType = tea.String(v.(string))
	}
	if v, ok := d.GetOk("pricing_cycle"); ok {
		request.PricingCycle = tea.String(v.(string))
	}
	if v, ok := d.GetOk("extra"); ok {
		request.Extra = tea.String(v.(string))
	}

	// Handle HA fields
	if ha, ok := d.GetOk("ha"); ok {
		haConfig := ha.([]interface{})
		if len(haConfig) > 0 {
			haMap := haConfig[0].(map[string]interface{})
			zoneId := haMap["zone_id"].(string)
			request.HaZoneId = tea.String(zoneId)
			request.Ha = tea.Bool(true) // 启用HA

			// 处理vswitch_ids
			haVswitchIds := haMap["vswitch_ids"].([]interface{})
			request.HaVSwitchIds = make([]*string, 0, len(haVswitchIds))
			for _, v := range haVswitchIds {
				request.HaVSwitchIds = append(request.HaVSwitchIds, tea.String(v.(string)))
			}

			// 处理resource部分
			haResource := haMap["resource"].([]interface{})[0].(map[string]interface{})
			cpu := haResource["cpu"].(int)
			memory := haResource["memory"].(int)
			request.HaResourceSpec = &foasconsole.CreateInstanceRequestHaResourceSpec{
				Cpu:      tea.Int32(int32(cpu)),
				MemoryGB: tea.Int32(int32(memory)),
			}
		}
	}

	// Process tags
	tags := d.Get("tags").(map[string]interface{})
	if len(tags) > 0 {
		request.Tag = make([]*foasconsole.CreateInstanceRequestTag, 0, len(tags))
		for k, v := range tags {
			tag := &foasconsole.CreateInstanceRequestTag{
				Key:   tea.String(k),
				Value: tea.String(v.(string)),
			}
			request.Tag = append(request.Tag, tag)
		}
	}

	// Set required fields
	request.ChargeType = tea.String(d.Get("charge_type").(string))

	region := d.Get("region").(string)
	request.Region = tea.String(region)

	// Make the create request with retry for transient errors
	var response *foasconsole.CreateInstanceResponse
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateInstance(&request)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_instance", "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.Body == nil || response.Body.OrderInfo == nil || response.Body.OrderInfo.InstanceId == nil {
		return WrapError(Error("Failed to get instance ID from response"))
	}

	d.SetId(*response.Body.OrderInfo.InstanceId)

	// Wait for the instance to be in running state
	stateConf := resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"RUNNING"},
		Refresh:    flinkService.FlinkWorkspaceStateRefreshFunc(region, d.Id()),
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

	// Create a describe request for the instance
	request := &foasconsole.DescribeInstancesRequest{}
	request.InstanceId = tea.String(d.Id())

	// Use region if available in state, otherwise use client's region
	if v, ok := d.GetOk("region"); ok {
		request.Region = tea.String(v.(string))
	} else {
		request.Region = tea.String(client.RegionId)
	}

	// Call the API to get the instance with retry for transient errors
	var response *foasconsole.DescribeInstancesResponse
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := flinkService.DescribeInstances(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Check if the instance was found
	if response == nil || response.Body == nil || len(response.Body.Instances) == 0 {
		d.SetId("")
		return nil
	}

	// Get the first instance (should be the only one since we're querying by ID)
	instance := response.Body.Instances[0]

	// Set basic attributes from the instance
	d.Set("name", instance.InstanceName)
	// Description is not available in the API response
	d.Set("resource_group_id", instance.ResourceGroupId)
	d.Set("zone_id", instance.ZoneId)
	d.Set("vpc_id", instance.VpcId)
	d.Set("charge_type", instance.ChargeType)
	d.Set("region", instance.Region)
	d.Set("ha", instance.Ha)
	d.Set("ha_zone_id", instance.HaZoneId)

	// Set VSwitchIds
	if instance.VSwitchIds != nil {
		d.Set("vswitch_ids", instance.VSwitchIds)
	}

	// Set HaVSwitchIds
	if instance.HaVSwitchIds != nil {
		d.Set("ha_vswitch_ids", instance.HaVSwitchIds)
	}

	// Set ResourceSpec
	if instance.ResourceSpec != nil {
		resourceSpec := []interface{}{
			map[string]interface{}{
				"cpu":    instance.ResourceSpec.Cpu,
				"memory": instance.ResourceSpec.MemoryGB,
			},
		}
		d.Set("resource_spec", resourceSpec)
	}

	// Set HaResourceSpec
	if instance.HaResourceSpec != nil {
		haResourceSpec := []interface{}{
			map[string]interface{}{
				"cpu":    instance.HaResourceSpec.Cpu,
				"memory": instance.HaResourceSpec.MemoryGB,
			},
		}
		d.Set("ha_resource_spec", haResourceSpec)
	}

	// Set storage
	if instance.Storage != nil && instance.Storage.Oss != nil && instance.Storage.Oss.Bucket != nil {
		storage := []interface{}{
			map[string]interface{}{
				"oss": []interface{}{
					map[string]interface{}{
						"bucket": *instance.Storage.Oss.Bucket,
					},
				},
			},
		}
		d.Set("storage", storage)
		d.Set("oss_bucket", *instance.Storage.Oss.Bucket)
	}

	// Set tags
	if instance.Tags != nil {
		tags := make(map[string]string)
		for _, tag := range instance.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
		d.Set("tags", tags)
	}

	return nil
}

func resourceAliCloudFlinkWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}
	region := d.Get("region").(string)
	request := &foasconsole.DeleteInstanceRequest{
		InstanceId: tea.String(d.Id()),
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, e := flinkService.DeleteInstance(request)
		if e != nil {
			if IsExpectedErrors(e, []string{"IncorrectInstanceStatus", "ThrottlingException", "OperationConflict"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteInstance", AlibabaCloudSdkGoERROR)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING", "RUNNING"},
		Target:     []string{},
		Refresh:    flinkService.FlinkWorkspaceStateRefreshFunc(region, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		if IsExpectedErrors(err, []string{"InvalidInstance.NotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, "Error deleting Flink instance: %s", d.Id())
	}

	d.SetId("")
	return nil
}
