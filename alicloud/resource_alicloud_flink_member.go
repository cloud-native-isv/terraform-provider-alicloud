package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkMemberCreate,
		Read:   resourceAliCloudFlinkMemberRead,
		Update: resourceAliCloudFlinkMemberUpdate,
		Delete: resourceAliCloudFlinkMemberDelete,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Flink workspaceId.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespaceName of the Flink member.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the member.",
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The role of the member, valid values are: editor, owner, viewer.",
				ValidateFunc: validation.StringInSlice([]string{"editor", "owner", "viewer"}, false),
			},
		},
	}
}

func resourceAliCloudFlinkMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create a Member struct for the service method
	member := &aliyunFlinkAPI.Member{
		Member: name,
		Role:   role,
	}

	// Pass workspaceId, namespaceName and Member struct to the service method
	_, err = flinkService.CreateMember(workspaceId, namespaceName, member)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(workspaceId + "/" + namespaceName + "/" + name)

	// Wait for member creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkMemberStateRefreshFunc(workspaceId, namespaceName, name, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkMemberRead(d, meta)
}

func resourceAliCloudFlinkMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse the ID to get workspaceId, namespaceName and member name
	// ID format: workspace_id/namespace_name/member_name
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return WrapError(fmt.Errorf("invalid resource id format: %s, expected workspace_id/namespace_name/member_name", d.Id()))
	}

	workspaceId := parts[0]
	namespaceName := parts[1]
	name := parts[2]

	// Use GetMember method with parsed values
	response, err := flinkService.GetMember(workspaceId, namespaceName, name)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes based on response
	if response != nil {
		d.Set("workspace_id", workspaceId)
		d.Set("namespace_name", namespaceName)
		d.Set("name", name)
		d.Set("role", response.Role)
	}

	return nil
}

func resourceAliCloudFlinkMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create a Member struct for the service method
	member := &aliyunFlinkAPI.Member{
		Member: name,
		Role:   role,
	}

	// Pass workspaceId, namespaceName and Member struct to the service method
	_, err = flinkService.UpdateMember(workspaceId, namespaceName, member)
	if err != nil {
		return WrapError(err)
	}

	return resourceAliCloudFlinkMemberRead(d, meta)
}

func resourceAliCloudFlinkMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)

	// Use DeleteMember method with string values instead of pointers
	err = flinkService.DeleteMember(workspaceId, namespaceName, name)
	return WrapError(err)
}
