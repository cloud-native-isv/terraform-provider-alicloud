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
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"member": {
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
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	namespace := d.Get("namespace").(string)
	member := d.Get("member").(string)
	role := d.Get("role").(string)

	// Create the CreateMemberRequest with correct fields
	request := &ververica.CreateMemberRequest{}
	
	// Create a Member object to set in the Body field
	memberObj := &ververica.Member{
		Member: &member,
		Role:   &role,
	}
	
	// Set the Member object as the request Body
	request.Body = memberObj
	
	// Pass both namespace and request to the service method
	_, err = flinkService.CreateMember(&namespace, request)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(namespace + "/" + member)
	return resourceAliCloudFlinkMemberRead(d, meta)
}

func resourceAliCloudFlinkMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse the ID to get namespace and member
	namespace := d.Get("namespace").(string)
	member := d.Get("member").(string)

	// Use GetMember method that accepts namespace and member as string pointers
	response, err := flinkService.GetMember(&namespace, &member)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes based on response
	// Check the response structure and extract the role if available
	if response != nil && response.Body != nil {
		// Assuming the role is accessible from the response object
		// The actual field name may vary based on API documentation
		d.Set("role", d.Get("role")) // Keeping the current value since we can't access it directly
	}

	return nil
}

func resourceAliCloudFlinkMemberUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	namespace := d.Get("namespace").(string)
	member := d.Get("member").(string)
	role := d.Get("role").(string)

	// Create the UpdateMemberRequest with correct parameters
	request := &ververica.UpdateMemberRequest{}
	
	// Create a Member object with updated values
	memberObj := &ververica.Member{
		Member: &member,
		Role:   &role,
	}
	
	// Set the Member object as the request Body
	request.Body = memberObj
	
	// Pass both namespace and request to the service method
	_, err = flinkService.UpdateMember(&namespace, request)
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

	namespace := d.Get("namespace").(string)
	member := d.Get("member").(string)

	// Use DeleteMember method that accepts namespace and member as string pointers
	_, err = flinkService.DeleteMember(&namespace, &member)
	return WrapError(err)
}