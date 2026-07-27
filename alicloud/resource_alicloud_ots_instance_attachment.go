package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunTablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudOtsInstanceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunOtsInstanceAttachmentCreate,
		Read:   resourceAliyunOtsInstanceAttachmentRead,
		Delete: resourceAliyunOtsInstanceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSInstanceName,
			},

			"vpc_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliyunOtsInstanceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	vpcService := VpcService{client}
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	request := &aliyunTablestoreAPI.BindInstanceRequest{
		InstanceName:    d.Get("instance_name").(string),
		InstanceVpcName: d.Get("vpc_name").(string),
		VirtualSwitchId: d.Get("vswitch_id").(string),
	}

	if vsw, err := vpcService.DescribeVSwitch(d.Get("vswitch_id").(string)); err != nil {
		return WrapError(err)
	} else {
		request.VpcId = vsw.VpcId
	}

	// The legacy BindInstance2Vpc RPC API rejects newer regions (e.g.
	// ap-southeast-3) with InvalidVersion, so bind through the modern
	// console OpenAPI (tablestore-20201209) instead.
	attachment, err := otsService.GetAPI().BindInstanceToVpcV2(request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_instance_attachment", "BindInstance2Vpc", AlibabaCloudSdkGoERROR)
	}
	addDebug("BindInstance2Vpc", attachment, request)

	d.SetId(request.InstanceName)
	return resourceAliyunOtsInstanceAttachmentRead(d, meta)
}

func resourceAliyunOtsInstanceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}
	object, err := otsService.DescribeOtsInstanceAttachment(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	// There is a bug that inst does not contain instance name and vswitch ID, so this resource does not support import function.
	//d.Set("instance_name", inst.InstanceName)
	d.Set("vpc_name", object.InstanceVpcName)
	d.Set("vpc_id", object.VpcId)
	return nil
}

func resourceAliyunOtsInstanceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}
	object, err := otsService.DescribeOtsInstanceAttachment(d.Id())
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	// The legacy UnbindInstance2Vpc RPC API rejects newer regions (e.g.
	// ap-southeast-3) with InvalidVersion, so unbind through the modern
	// console OpenAPI (tablestore-20201209) instead.
	request := &aliyunTablestoreAPI.UnbindInstanceRequest{
		InstanceName:    d.Id(),
		InstanceVpcName: object.InstanceVpcName,
	}
	if err := otsService.GetAPI().UnbindInstanceFromVpcV2(request); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UnbindInstance2Vpc", AlibabaCloudSdkGoERROR)
	}
	addDebug("UnbindInstance2Vpc", "Success", request)
	return WrapError(otsService.WaitForOtsInstanceVpc(d.Id(), Deleted, DefaultTimeout))
}
