package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgSecurityIpArray() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgSecurityIpArrayCreate,
		Read:   resourceAliCloudAdbpgSecurityIpArrayRead,
		Update: resourceAliCloudAdbpgSecurityIpArrayUpdate,
		Delete: resourceAliCloudAdbpgSecurityIpArrayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_instance_ip_array_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
			},
			"security_ip_list": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAliCloudAdbpgSecurityIpArrayCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	arrayName := d.Get("db_instance_ip_array_name").(string)
	ipList := d.Get("security_ip_list").(string)

	if err := adbpgService.ModifyAdbpgSecurityIps(instanceId, ipList, arrayName); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, arrayName))

	return resourceAliCloudAdbpgSecurityIpArrayRead(d, meta)
}

func resourceAliCloudAdbpgSecurityIpArrayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	detail, err := adbpgService.DescribeAdbpgInstance(parts[0])
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Instance %s not found, removing security IP array from state", parts[0])
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("db_instance_id", parts[0])
	d.Set("db_instance_ip_array_name", parts[1])
	d.Set("security_ip_list", detail.SecurityIPList)

	return nil
}

func resourceAliCloudAdbpgSecurityIpArrayUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("security_ip_list") {
		if err := adbpgService.ModifyAdbpgSecurityIps(parts[0], d.Get("security_ip_list").(string), parts[1]); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudAdbpgSecurityIpArrayRead(d, meta)
}

func resourceAliCloudAdbpgSecurityIpArrayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	return WrapError(adbpgService.ModifyAdbpgSecurityIps(parts[0], "127.0.0.1", parts[1]))
}
