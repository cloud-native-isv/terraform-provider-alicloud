package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgSsl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgSslCreate,
		Read:   resourceAliCloudAdbpgSslRead,
		Update: resourceAliCloudAdbpgSslUpdate,
		Delete: resourceAliCloudAdbpgSslDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ssl_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"ssl_expired": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgSslCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	sslEnabled := d.Get("ssl_enabled").(bool)

	if err := adbpgService.ModifyAdbpgSSL(instanceId, sslEnabled); err != nil {
		return WrapError(err)
	}

	d.SetId(instanceId)

	return resourceAliCloudAdbpgSslRead(d, meta)
}

func resourceAliCloudAdbpgSslRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	config, err := adbpgService.DescribeAdbpgSSL(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG SSL for instance %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("db_instance_id", d.Id())
	d.Set("ssl_enabled", config.SSLEnabled)
	d.Set("ssl_expired", config.SSLExpired)

	return nil
}

func resourceAliCloudAdbpgSslUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("ssl_enabled") {
		if err := adbpgService.ModifyAdbpgSSL(d.Id(), d.Get("ssl_enabled").(bool)); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudAdbpgSslRead(d, meta)
}

func resourceAliCloudAdbpgSslDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	return WrapError(adbpgService.ModifyAdbpgSSL(d.Id(), false))
}
