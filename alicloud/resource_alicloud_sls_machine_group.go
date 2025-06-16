package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudLogMachineGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogMachineGroupCreate,
		Read:   resourceAlicloudLogMachineGroupRead,
		Update: resourceAlicloudLogMachineGroupUpdate,
		Delete: resourceAlicloudLogMachineGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"identify_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      aliyunSlsAPI.MachineIDTypeIP,
				ValidateFunc: validation.StringInSlice([]string{aliyunSlsAPI.MachineIDTypeIP, aliyunSlsAPI.MachineIDTypeUserDefined}, false),
			},
			"identify_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"topic": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAlicloudLogMachineGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Get("project").(string)
	machineGroupName := d.Get("name").(string)

	// Build machine group object
	machineGroup := &aliyunSlsAPI.MachineGroup{
		Name:          machineGroupName,
		MachineIDType: d.Get("identify_type").(string),
		MachineIDList: expandStringList(d.Get("identify_list").([]interface{})),
	}

	// Set topic if provided
	if topic, ok := d.GetOk("topic"); ok {
		machineGroup.Attribute = &aliyunSlsAPI.MachineGroupAttribute{
			GroupTopic: topic.(string),
		}
	}

	// Create machine group
	err = slsService.CreateSlsMachineGroup(projectName, machineGroup)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_machine_group", "CreateMachineGroup", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using the composite format: project:machine_group_name
	d.SetId(slsService.BuildMachineGroupId(projectName, machineGroupName))

	// Wait for machine group to be available
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsMachineGroupStateRefreshFunc(projectName, machineGroupName, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudLogMachineGroupRead(d, meta)
}

func resourceAlicloudLogMachineGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse the composite ID
	projectName, machineGroupName, err := slsService.ParseMachineGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Get machine group details
	machineGroup, err := slsService.DescribeSlsMachineGroup(projectName, machineGroupName)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set resource attributes
	d.Set("project", projectName)
	d.Set("name", machineGroup.Name)
	d.Set("identify_type", machineGroup.MachineIDType)
	d.Set("identify_list", machineGroup.MachineIDList)

	// Set topic if available
	if machineGroup.Attribute != nil {
		d.Set("topic", machineGroup.Attribute.GroupTopic)
	}

	return nil
}

func resourceAlicloudLogMachineGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	// Parse the composite ID
	projectName, machineGroupName, err := slsService.ParseMachineGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if d.HasChanges("identify_type", "identify_list", "topic") {
		// Build updated machine group object
		machineGroup := &aliyunSlsAPI.MachineGroup{
			Name:          machineGroupName,
			MachineIDType: d.Get("identify_type").(string),
			MachineIDList: expandStringList(d.Get("identify_list").([]interface{})),
		}

		// Set topic if provided
		if topic, ok := d.GetOk("topic"); ok {
			machineGroup.Attribute = &aliyunSlsAPI.MachineGroupAttribute{
				GroupTopic: topic.(string),
			}
		}

		// Update machine group
		err := slsService.UpdateSlsMachineGroup(projectName, machineGroupName, machineGroup)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateMachineGroup", AlibabaCloudSdkGoERROR)
		}

		// Wait for machine group to be updated
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.SlsMachineGroupStateRefreshFunc(projectName, machineGroupName, []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAlicloudLogMachineGroupRead(d, meta)
}

func resourceAlicloudLogMachineGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	// Parse the composite ID
	projectName, machineGroupName, err := slsService.ParseMachineGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Delete machine group
	err = slsService.DeleteSlsMachineGroup(projectName, machineGroupName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteMachineGroup", AlibabaCloudSdkGoERROR)
	}

	// Wait for machine group to be deleted
	return WrapError(slsService.WaitForSlsMachineGroup(projectName, machineGroupName, "Deleted", DefaultTimeoutMedium))
}
