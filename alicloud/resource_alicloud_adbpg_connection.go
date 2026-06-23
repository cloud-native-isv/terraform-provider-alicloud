package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgConnectionCreate,
		Read:   resourceAliCloudAdbpgConnectionRead,
		Delete: resourceAliCloudAdbpgConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_string_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	prefix := d.Get("connection_string_prefix").(string)

	if err := adbpgService.AllocateAdbpgPublicConnection(instanceId, prefix); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, prefix))

	if err := adbpgService.WaitForAdbpgInstanceRunning(instanceId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudAdbpgConnectionRead(d, meta)
}

func resourceAliCloudAdbpgConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	info, err := adbpgService.DescribeAdbpgPublicConnection(parts[0])
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Connection %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("db_instance_id", parts[0])
	d.Set("connection_string_prefix", parts[1])
	d.Set("connection_string", info.ConnectionString)
	d.Set("ip_address", info.IPAddress)
	d.Set("port", info.Port)

	return nil
}

func resourceAliCloudAdbpgConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	info, err := adbpgService.DescribeAdbpgPublicConnection(parts[0])
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	return WrapError(adbpgService.ReleaseAdbpgPublicConnection(parts[0], info.ConnectionString))
}
