package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsAddon manages a CMS 2.0 addon release through the
// cws-lib-go cms API layer. The addon catalog entry itself is read-only
// (Helm-chart-like), so the managed lifecycle is the release identified
// by the (addon_name, version) pair. The composite resource ID is
// "<addon_name>:<version>".
func resourceAliCloudCmsAddon() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsAddonCreate,
		Read:   resourceAliCloudCmsAddonRead,
		Update: resourceAliCloudCmsAddonUpdate,
		Delete: resourceAliCloudCmsAddonDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the CMS addon.",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The version of the addon release.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the addon release.",
			},
			"release_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The release status reported by the service.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the addon release.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the addon release.",
			},
		},
	}
}

func cmsAddonId(addonName, version string) string {
	return addonName + ":" + version
}

func parseCmsAddonId(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid CMS addon id %q, expected <addon_name>:<version>", id)
	}
	return parts[0], parts[1], nil
}

func resourceAliCloudCmsAddonCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	addon := &cmsapi.CmsAddons{
		AddonName:   d.Get("addon_name").(string),
		Version:     d.Get("version").(string),
		Description: d.Get("description").(string),
	}
	if _, err := service.CreateCmsAddonRelease(addon); err != nil {
		return WrapError(err)
	}

	d.SetId(cmsAddonId(addon.AddonName, addon.Version))
	return resourceAliCloudCmsAddonRead(d, meta)
}

func resourceAliCloudCmsAddonRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	addonName, version, err := parseCmsAddonId(d.Id())
	if err != nil {
		return WrapError(err)
	}
	addon, err := service.GetCmsAddonRelease(addonName, version)
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("addon_name", firstNonEmptyString(addon.AddonName, addonName))
	d.Set("version", firstNonEmptyString(addon.Version, version))
	if addon.Description != "" {
		d.Set("description", addon.Description)
	}
	d.Set("release_status", addon.ReleaseStatus)
	d.Set("create_time", addon.CreateTime)
	d.Set("update_time", addon.UpdateTime)
	return nil
}

func resourceAliCloudCmsAddonUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("description") {
		addon := &cmsapi.CmsAddons{
			AddonName:   d.Get("addon_name").(string),
			Version:     d.Get("version").(string),
			Description: d.Get("description").(string),
		}
		if _, err := service.UpdateCmsAddonRelease(addon); err != nil {
			return WrapError(err)
		}
	}
	return resourceAliCloudCmsAddonRead(d, meta)
}

func resourceAliCloudCmsAddonDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	addonName, version, err := parseCmsAddonId(d.Id())
	if err != nil {
		return WrapError(err)
	}
	if err := service.DeleteCmsAddonRelease(addonName, version); err != nil {
		return WrapError(err)
	}
	return nil
}

// firstNonEmptyString returns the first non-empty string of the given
// values, or the empty string when all of them are empty.
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
