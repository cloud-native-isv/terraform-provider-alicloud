package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
				Description: "The name of the Flink workspace.",
			},
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespace of the Flink member.",
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
				Description:  "The role of the member, valid values are: EDITOR, OWNER, VIEWER.",
				ValidateFunc: validation.StringInSlice([]string{"EDITOR", "OWNER", "VIEWER"}, false),
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

	workspace := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create a request map for the service method
	memberData := map[string]interface{}{
		"member": name,
		"role":   role,
	}

	// Pass workspace, namespace and request to the service method
	_, err = flinkService.CreateMember(workspace, namespace, memberData)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(workspace + "/" + namespace + "/" + name)
	return resourceAliCloudFlinkMemberRead(d, meta)
}

func resourceAliCloudFlinkMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse the ID to get workspace, namespace and member name
	workspace := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)

	// Use GetMember method with string values instead of pointers
	response, err := flinkService.GetMember(workspace, namespace, name)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes based on response
	if response != nil {
		// The response is now a map[string]interface{}, not a struct with Body field
		d.Set("role", response["role"])
	}

	return nil
}

func resourceAliCloudFlinkMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create a map with the member data
	memberData := map[string]interface{}{
		"member": name,
		"role":   role,
	}

	// Pass workspace, namespace and map to the service method
	_, err = flinkService.UpdateMember(workspace, namespace, memberData)
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

	workspace := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)

	// Use DeleteMember method with string values instead of pointers
	err = flinkService.DeleteMember(workspace, namespace, name)
	return WrapError(err)
}
