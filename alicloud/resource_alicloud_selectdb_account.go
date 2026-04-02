package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSelectDBAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBAccountCreate,
		Read:   resourceAliCloudSelectDBAccountRead,
		Update: resourceAliCloudSelectDBAccountUpdate,
		Delete: resourceAliCloudSelectDBAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB instance.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The admin username of the SelectDB instance.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The admin password for the SelectDB instance account.",
			},
		},
	}
}

func resourceAliCloudSelectDBAccountCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	if err := selectDBService.WaitForSelectDBInstanceUpdated(instanceId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, instanceId)
	}

	if err := selectDBService.resetSelectDBInstancePassword(instanceId, username, password); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceId, "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
	}

	if err := selectDBService.WaitForSelectDBInstanceUpdated(instanceId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, instanceId)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, username))
	return resourceAliCloudSelectDBAccountRead(d, meta)
}

func resourceAliCloudSelectDBAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	instanceId := parts[0]
	username := parts[1]

	_, err = selectDBService.DescribeSelectDBInstance(instanceId)
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_selectdb_account not found!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("instance_id", instanceId)
	d.Set("username", username)
	return nil
}

func resourceAliCloudSelectDBAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	instanceId := parts[0]
	username := parts[1]

	if d.HasChange("password") {
		password := d.Get("password").(string)
		if password == "" {
			return WrapErrorf(fmt.Errorf("password must be provided when resetting account password"), DefaultErrorMsg, instanceId, "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
		}

		if err := selectDBService.resetSelectDBInstancePassword(instanceId, username, password); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, instanceId, "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
		}

		if err := selectDBService.WaitForSelectDBInstanceUpdated(instanceId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return WrapErrorf(err, IdMsg, instanceId)
		}
	}

	return resourceAliCloudSelectDBAccountRead(d, meta)
}

func resourceAliCloudSelectDBAccountDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
