package alicloud

import (
	"log"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAccountRead,

		Schema: map[string]*schema.Schema{
			// Computed values
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAliCloudAccountRead(d *schema.ResourceData, meta interface{}) error {
	accountId, err := meta.(*connectivity.AliyunClient).AccountId()

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] alicloud_account - account ID found: %#v", accountId)

	d.SetId(accountId)

	return nil
}
