package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
)

func resourceAliCloudFlinkMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkMemberCreate,
		Read:   resourceAliCloudFlinkMemberRead,
		Update: resourceAliCloudFlinkMemberUpdate,
		Delete: resourceAliCloudFlinkMemberDelete,

		Schema: map[string]*schema.Schema{
			"member_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "member",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudFlinkMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService := FlinkService{ververicaClient: meta.(*connectivity.AliyunClient).ververicaClient}

	request := &ververica.CreateMemberRequest{
		MemberId:   d.Get("member_id").(string),
		InstanceId: d.Get("instance_id").(string),
		Role:       d.Get("role").(string),
	}

	response, err := flinkService.ververicaClient.CreateMember(request)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(response.MemberId)
	return resourceAliCloudFlinkMemberRead(d, meta)
}

func resourceAliCloudFlinkMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService := FlinkService{ververicaClient: meta.(*connectivity.AliyunClient).ververicaClient}

	request := &ververica.DescribeMemberRequest{
		MemberId:   d.Id(),
		InstanceId: d.Get("instance_id").(string),
	}

	response, err := flinkService.ververicaClient.DescribeMember(request)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("status", response.Status)
	d.Set("role", response.Role)
	return nil
}

func resourceAliCloudFlinkMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService := FlinkService{ververicaClient: meta.(*connectivity.AliyunClient).ververicaClient}

	request := &ververica.UpdateMemberRequest{
		MemberId:   d.Id(),
		InstanceId: d.Get("instance_id").(string),
		Role:       d.Get("role").(string),
	}

	_, err := flinkService.ververicaClient.UpdateMember(request)
	return WrapError(err)
}

func resourceAliCloudFlinkMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService := FlinkService{ververicaClient: meta.(*connectivity.AliyunClient).ververicaClient}

	request := &ververica.DeleteMemberRequest{
		MemberId:   d.Id(),
		InstanceId: d.Get("instance_id").(string),
	}

	_, err := flinkService.ververicaClient.DeleteMember(request)
	return WrapError(err)
}