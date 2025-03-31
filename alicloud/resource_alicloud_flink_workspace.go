package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	foasconsole "github.com/alibabacloud-go/foasconsoleClient-20211028/client"
)

func resourceAliCloudFlinkWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkWorkspaceCreate,
		Read:   resourceAliCloudFlinkWorkspaceRead,
		Update: resourceAliCloudFlinkWorkspaceUpdate,
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
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	request := foasconsole.CreateCreateInstanceRequest()
	request.Name = d.Get("name").(string)
	request.VpcId = d.Get("vpc_id").(string)
	request.VSwitchIds = d.Get("vswitch_ids").([]string)
	request.Description = d.Get("description").(string)

	response, err := foasService.CreateInstance(request)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(response.InstanceId)

	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, foasService.FlinkWorkspaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkWorkspaceRead(d, meta)
}

func resourceAliCloudFlinkWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	instance, err := foasService.DescribeInstance(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("name", instance.Name)
	d.Set("description", instance.Description)
	d.Set("vpc_id", instance.VpcId)
	d.Set("vswitch_ids", instance.VSwitchIds)

	return nil
}

func resourceAliCloudFlinkWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	request := foasconsole.CreateModifyInstanceRequest()
	request.InstanceId = d.Id()

	if d.HasChange("description") {
		request.Description = d.Get("description").(string)
	}

	_, err := foasService.ModifyInstance(request)
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, foasService.FlinkWorkspaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkWorkspaceRead(d, meta)
}

func resourceAliCloudFlinkWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	foasService := FoasService{client}

	request := foasconsole.CreateDeleteInstanceRequest()
	request.InstanceId = d.Id()

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := foasService.DeleteInstance(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"IncorrectInstanceStatus"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{"DELETING"}, []string{}, d.Timeout(schema.TimeoutDelete), 10*time.Second, foasService.FlinkWorkspaceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		if IsExpectedErrors(err, []string{"InvalidInstance.NotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

type FoasService struct {
	*connectivity.AliyunClient
}

func (s *FoasService) CreateInstance(request *foasconsole.CreateInstanceRequest) (*foasconsole.CreateInstanceResponse, error) {
	response, err := s.FoasconsoleClient.CreateInstance(request)
	if err != nil {
		log.Printf("[ERROR] CreateInstance failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DescribeInstance(instanceId string) (*foasconsole.DescribeInstancesResponseInstance, error) {
	request := foasconsole.CreateDescribeInstancesRequest()
	request.InstanceId = instanceId
	response, err := s.FoasconsoleClient.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	for _, inst := range response.Instances {
		if inst.InstanceId == instanceId {
			return &inst, nil
		}
	}
	return nil, WrapErrorf(Error("Instance %s not found", instanceId))
}

func (s *FoasService) ModifyInstance(request *foasconsole.ModifyInstanceRequest) (*foasconsole.ModifyInstanceResponse, error) {
	response, err := s.FoasconsoleClient.ModifyInstance(request)
	if err != nil {
		log.Printf("[ERROR] ModifyInstance failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) DeleteInstance(request *foasconsole.DeleteInstanceRequest) (*foasconsole.DeleteInstanceResponse, error) {
	response, err := s.FoasconsoleClient.DeleteInstance(request)
	if err != nil {
		log.Printf("[ERROR] DeleteInstance failed: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *FoasService) FlinkWorkspaceStateRefreshFunc(instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeInstance(instanceId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "DELETED", nil
			}
			return nil, "", err
		}
		return obj, obj.Status, nil
	}
}