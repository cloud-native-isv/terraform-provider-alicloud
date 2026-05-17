package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudCmsService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudCmsServiceRead,

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"On", "Off"}, false),
				Optional:     true,
				Default:      "Off",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func dataSourceAliCloudCmsServiceRead(d *schema.ResourceData, meta interface{}) error {
	if v, ok := d.GetOk("enable"); !ok || v.(string) != "On" {
		d.SetId("CmsServiceHasNotBeenOpened")
		d.Set("status", "")
		return nil
	}
	client := meta.(*connectivity.AliyunClient)
	cmsService, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	err = cmsService.OpenCmsService()
	if err != nil {
		if IsExpectedErrors(err, []string{"ORDER.OPEND", "Has.effect.suit"}) {
			d.SetId("CmsServiceHasBeenOpened")
			d.Set("status", "Opened")
			return nil
		}
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_cms_service", "OpenCmsService", AlibabaCloudSdkGoERROR)
	}

	enabled, statusErr := cmsService.GetCmsServiceStatus()
	if statusErr == nil && enabled {
		d.SetId("CmsServiceHasBeenOpened")
		d.Set("status", "Opened")
		return nil
	}

	d.SetId("CmsServiceHasBeenOpened")
	d.Set("status", "Opened")

	return nil
}
