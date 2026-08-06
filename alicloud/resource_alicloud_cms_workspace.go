package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsWorkspace manages a CMS 2.0 workspace through the
// cws-lib-go cms API layer. The CMS workspace API only exposes an
// idempotent Put plus Get/Delete and has no dedicated Update action, so
// every user-writable attribute is ForceNew and the Update hook is
// intentionally omitted.
func resourceAliCloudCmsWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsWorkspaceCreate,
		Read:   resourceAliCloudCmsWorkspaceRead,
		Delete: resourceAliCloudCmsWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique name of the CMS workspace.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The description of the CMS workspace.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name reported by the CMS workspace.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region in which the CMS workspace lives.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the CMS workspace.",
			},
		},
	}
}

func resourceAliCloudCmsWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := &cmsapi.CmsWorkspace{
		Workspace:   d.Get("workspace").(string),
		Description: d.Get("description").(string),
	}
	result, err := service.PutCmsWorkspace(workspace)
	if err != nil {
		return WrapError(err)
	}

	id := result.WorkspaceId
	if id == "" {
		id = workspace.Workspace
	}
	d.SetId(id)
	return resourceAliCloudCmsWorkspaceRead(d, meta)
}

func resourceAliCloudCmsWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, err := service.GetCmsWorkspace(d.Id())
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace", workspace.Workspace)
	d.Set("name", workspace.Name)
	d.Set("description", workspace.Description)
	d.Set("region_id", workspace.RegionId)
	d.Set("status", workspace.Status)
	return nil
}

func resourceAliCloudCmsWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := service.DeleteCmsWorkspace(d.Id()); err != nil {
		return WrapError(err)
	}
	return nil
}
