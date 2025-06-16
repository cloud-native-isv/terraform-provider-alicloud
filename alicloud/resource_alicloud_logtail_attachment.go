package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudLogtailAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogtailAttachmentCreate,
		Read:   resourceAlicloudLogtailAttachmentRead,
		Delete: resourceAlicloudLogtailAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"logtail_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"machine_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAlicloudLogtailAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	projectName := d.Get("project").(string)
	configName := d.Get("logtail_config_name").(string)
	machineGroupName := d.Get("machine_group_name").(string)

	// Validate attachment parameters
	if err := slsService.ValidateSlsLogtailAttachment(projectName, configName, machineGroupName); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "ValidateAttachment", AlibabaCloudSdkGoERROR)
	}

	// Create logtail attachment
	_, err = slsService.CreateSlsLogtailAttachment(projectName, configName, machineGroupName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "CreateLogtailAttachment", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s%s%s%s%s", projectName, COLON_SEPARATED, configName, COLON_SEPARATED, machineGroupName))

	// Use StateRefreshFunc to wait for attachment to be ready
	stateConf := BuildStateConf([]string{}, []string{"active"}, d.Timeout(schema.TimeoutCreate), 5*time.Second,
		slsService.SlsLogtailAttachmentStateRefreshFunc(d.Id(), "status", []string{}))

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudLogtailAttachmentRead(d, meta)
}

func resourceAlicloudLogtailAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	object, err := slsService.DescribeSlsLogtailAttachment(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	d.Set("project", parts[0])
	d.Set("logtail_config_name", parts[1])
	d.Set("machine_group_name", parts[2])

	// Set computed attributes from the object
	if v, ok := object["status"]; ok {
		d.Set("status", v)
	}
	if v, ok := object["createTime"]; ok {
		d.Set("create_time", v)
	}
	if v, ok := object["lastModifyTime"]; ok {
		d.Set("last_modify_time", v)
	}

	return nil
}

func resourceAlicloudLogtailAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	projectName := parts[0]
	configName := parts[1]
	machineGroupName := parts[2]

	// Delete logtail attachment
	if err := slsService.DeleteSlsLogtailAttachment(projectName, configName, machineGroupName); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_attachment", "DeleteLogtailAttachment", AlibabaCloudSdkGoERROR)
	}

	// Use StateRefreshFunc to wait for attachment to be deleted
	stateConf := BuildStateConf([]string{"active"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second,
		slsService.SlsLogtailAttachmentStateRefreshFunc(d.Id(), "status", []string{}))

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
