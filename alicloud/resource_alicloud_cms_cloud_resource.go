package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsCloudResource registers a cloud resource with CMS 2.0
// through the cws-lib-go cms API layer. The CMS cloud resource API exposes
// Create/Get/Delete only and no general Update action (ChangeResourceGroup
// is a group-move, not an attribute update), so every user-writable
// attribute is ForceNew and the Update hook is intentionally omitted.
func resourceAliCloudCmsCloudResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsCloudResourceCreate,
		Read:   resourceAliCloudCmsCloudResourceRead,
		Delete: resourceAliCloudCmsCloudResourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud resource identifier registered with CMS.",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of the cloud resource.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The display name of the cloud resource registration.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The region in which the cloud resource lives.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The resource group of the cloud resource.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the cloud resource registration.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the registration.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the registration.",
			},
		},
	}
}

func resourceAliCloudCmsCloudResourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	resource := &cmsapi.CmsCloudResource{
		ResourceId:      d.Get("resource_id").(string),
		ResourceType:    d.Get("resource_type").(string),
		Name:            d.Get("name").(string),
		RegionId:        d.Get("region_id").(string),
		ResourceGroupId: d.Get("resource_group_id").(string),
	}
	result, err := service.CreateCmsCloudResource(resource)
	if err != nil {
		return WrapError(err)
	}

	id := result.ResourceId
	if id == "" {
		id = resource.ResourceId
	}
	d.SetId(id)
	return resourceAliCloudCmsCloudResourceRead(d, meta)
}

func resourceAliCloudCmsCloudResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	resource, err := service.GetCmsCloudResource(d.Id())
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("resource_id", firstNonEmptyString(resource.ResourceId, d.Id()))
	if resource.ResourceType != "" {
		d.Set("resource_type", resource.ResourceType)
	}
	if resource.Name != "" {
		d.Set("name", resource.Name)
	}
	if resource.RegionId != "" {
		d.Set("region_id", resource.RegionId)
	}
	if resource.ResourceGroupId != "" {
		d.Set("resource_group_id", resource.ResourceGroupId)
	}
	d.Set("status", resource.Status)
	d.Set("create_time", resource.CreateTime)
	d.Set("update_time", resource.UpdateTime)
	return nil
}

func resourceAliCloudCmsCloudResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := service.DeleteCmsCloudResource(d.Id()); err != nil {
		return WrapError(err)
	}
	return nil
}
