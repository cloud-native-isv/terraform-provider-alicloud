package alicloud

import (
	"time"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"oss_bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vswitch_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10000,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  40000,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"architecture_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_renew": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"charge_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"extra": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ha": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ha_resource_spec": {
				Type:     schema.TypeList, // Need detailed spec struct handling
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Define fields here based on HaResourceSpec struct
					},
				},
			},
			"ha_vswitch_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ha_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"monitor_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pricing_cycle": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"promotion_code": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_spec": {
				Type:     schema.TypeList, // For ResourceSpec struct
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"storage": {
				Type:     schema.TypeList, // For Storage struct
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Define storage fields here
					},
				},
			},
			"use_promotion_code": {
				Type:     schema.TypeBool,
				Optional: true,
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

	ossBucket := d.Get("oss_bucket").(string) // 提取OSS bucket值到storage结构
	request.Storage = &foasconsole.CreateInstanceRequestStorage{
		Oss: &foasconsole.CreateInstanceRequestStorageOss{
			Bucket: tea.String(ossBucket),
		},
	}

	// Handle vswitch_ids as list
	vswitchIds := d.Get("vswitch_ids").([]interface{})
	request.VSwitchIds = make([]*string, 0, len(vswitchIds))
	for _, v := range vswitchIds {
		request.VSwitchIds = append(request.VSwitchIds, tea.String(v.(string)))
	}

	// Handle nested ResourceSpec
	resourceSpec := d.Get("resource_spec").([]interface{})[0].(map[string]interface{})
	request.ResourceSpec = &foasconsole.CreateInstanceRequestResourceSpec{
		Cpu:     tea.Int32(int32(resourceSpec["cpu"].(int))),
		MemoryGB:  tea.Int32(int32(resourceSpec["memory"].(int))),
	}

	// Handle boolean fields with proper conversion
	request.AutoRenew = tea.Bool(d.Get("auto_renew").(bool))
	request.Ha = tea.Bool(d.Get("ha").(bool))
	request.UsePromotionCode = tea.Bool(d.Get("use_promotion_code").(bool))

	// Handle optional fields
	if val, ok := d.GetOk("duration"); ok {
		request.Duration = tea.Int32(int32(val.(int)))
	}

	if val, ok := d.GetOk("promotion_code"); ok {
		request.PromotionCode = tea.String(val.(string))
	}

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

	request.ChargeType = tea.String(d.Get("charge_type").(string))
	request.Region = tea.String(d.Get("region").(string))

	response, err := flinkService.CreateInstance(&request)
	if err != nil {
		return WrapError(err)
	}

	if response.Body.Success == nil || !*response.Body.Success {
		return WrapErrorf(err, "CreateInstance failed with response: %v", response)
	}

	d.SetId(*response.Body.OrderInfo.InstanceId)

	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, flinkService.FlinkWorkspaceStateRefreshFunc(d.Id()))
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
	request.Region = tea.String(d.Get("region").(string))
	
	// Call the API to get the instance
	response, err := flinkService.DescribeInstances(request)
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

	// Set attributes from the instance
	d.Set("name", instance.InstanceName)
	if instance.ResourceSpec != nil {
		d.Set("cpu", instance.ResourceSpec.Cpu)
		d.Set("memory", instance.ResourceSpec.MemoryGB)
	}
	d.Set("charge_type", instance.ChargeType)
	d.Set("vpc_id", instance.VpcId)
	d.Set("zone_id", instance.ZoneId)
	
	// Set VSwitchIds
	if instance.VSwitchIds != nil {
		d.Set("vswitch_ids", instance.VSwitchIds)
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
	
	// Set other attributes
	d.Set("ha", instance.Ha)
	d.Set("region", instance.Region)
	
	return nil
}

func resourceAliCloudFlinkWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	request := &foasconsole.DeleteInstanceRequest{
		InstanceId: tea.String(d.Id()),
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, e := flinkService.DeleteInstance(request)
		if e != nil {
			if IsExpectedErrors(e, []string{"IncorrectInstanceStatus"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	})
	if err != nil {
		return WrapError(err)
	}

	// 添加删除后的状态检查
	stateConf := BuildStateConf([]string{"DELETING"}, []string{}, d.Timeout(schema.TimeoutDelete), 10*time.Second, flinkService.FlinkWorkspaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		if IsExpectedErrors(err, []string{"InvalidInstance.NotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, "Error deleting Flink workspace: %s", d.Id())
	}

	d.SetId("")
	return nil
}

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &foasconsole.DescribeInstancesRequest{}
		request.InstanceId = tea.String(id)
		
		response, err := s.DescribeInstances(request)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		
		// Check if the instance was found
		if response == nil || response.Body == nil || len(response.Body.Instances) == 0 {
			return nil, "", nil
		}
		
		// Get the first instance (should be the only one since we're querying by ID)
		instance := response.Body.Instances[0]
		// Use ClusterStatus instead of Status which doesn't exist
		return instance, *instance.ClusterStatus, nil
	}
}
