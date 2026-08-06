package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsEntityStore manages a CMS 2.0 entity store through
// the cws-lib-go cms API layer. The entity store API exposes
// Create/Get/Delete only and no Update action, so every user-writable
// attribute is ForceNew and the Update hook is intentionally omitted.
func resourceAliCloudCmsEntityStore() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsEntityStoreCreate,
		Read:   resourceAliCloudCmsEntityStoreRead,
		Delete: resourceAliCloudCmsEntityStoreDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the entity store.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The description of the entity store.",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "The metadata labels attached to the entity store.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the entity store.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the entity store.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the entity store.",
			},
		},
	}
}

// cmsEntityStoreMetadata converts the Terraform TypeMap value into the
// map[string]string expected by the cws-lib-go typed struct.
func cmsEntityStoreMetadata(raw map[string]interface{}) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	result := make(map[string]string, len(raw))
	for key, value := range raw {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

func resourceAliCloudCmsEntityStoreCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	store := &cmsapi.CmsEntityStore{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Metadata:    cmsEntityStoreMetadata(d.Get("metadata").(map[string]interface{})),
	}
	result, err := service.CreateCmsEntityStore(store)
	if err != nil {
		return WrapError(err)
	}

	id := result.StoreId
	if id == "" {
		id = store.Name
	}
	d.SetId(id)
	return resourceAliCloudCmsEntityStoreRead(d, meta)
}

func resourceAliCloudCmsEntityStoreRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	store, err := service.GetCmsEntityStore(d.Id())
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if store.Name != "" {
		d.Set("name", store.Name)
	}
	if store.Description != "" {
		d.Set("description", store.Description)
	}
	if len(store.Metadata) > 0 {
		d.Set("metadata", store.Metadata)
	}
	d.Set("status", store.Status)
	d.Set("create_time", store.CreateTime)
	d.Set("update_time", store.UpdateTime)
	return nil
}

func resourceAliCloudCmsEntityStoreDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := service.DeleteCmsEntityStore(d.Id()); err != nil {
		return WrapError(err)
	}
	return nil
}
