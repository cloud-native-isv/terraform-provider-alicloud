package alicloud

import (
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
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
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Flink workspace.",
			},
			"namespace": {
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

	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create the CreateMemberRequest with correct fields
	request := &ververica.CreateMemberRequest{}

	// Create a Member object to set in the Body field
	memberObj := &ververica.Member{
		Member: &name,
		Role:   &role,
	}

	// Set the Member object as the request Body
	request.Body = memberObj

	// Pass workspace, namespace and request to the service method
	_, err = flinkService.CreateMember(&workspace, &namespace, request)
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
	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)

	// Use GetMember method that accepts workspace, namespace and member as string pointers
	response, err := flinkService.GetMember(&workspace, &namespace, &name)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes based on response
	if response != nil && response.Body != nil && response.Body.Data != nil {
		if response.Body.Data.Role != nil {
			d.Set("role", *response.Body.Data.Role)
		}
	}

	return nil
}

func resourceAliCloudFlinkMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Create the UpdateMemberRequest with correct parameters
	request := &ververica.UpdateMemberRequest{}

	// Create a Member object with updated values
	memberObj := &ververica.Member{
		Member: &name,
		Role:   &role,
	}

	// Set the Member object as the request Body
	request.Body = memberObj

	// Pass workspace, namespace and request to the service method
	_, err = flinkService.UpdateMember(&workspace, &namespace, request)
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

	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)

	// Use DeleteMember method that accepts workspace, namespace and member as string pointers
	_, err = flinkService.DeleteMember(&workspace, &namespace, &name)
	return WrapError(err)
}
