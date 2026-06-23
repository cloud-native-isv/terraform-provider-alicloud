package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgAccountCreate,
		Read:   resourceAliCloudAdbpgAccountRead,
		Update: resourceAliCloudAdbpgAccountUpdate,
		Delete: resourceAliCloudAdbpgAccountDelete,
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
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"account_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"account_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgAccountCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	input := &adbpg.AdbpgAccountCreate{
		AccountName:     d.Get("account_name").(string),
		AccountPassword: d.Get("account_password").(string),
	}
	if v, ok := d.GetOk("account_type"); ok {
		input.AccountType = v.(string)
	}
	if v, ok := d.GetOk("account_description"); ok {
		input.AccountDescription = v.(string)
	}

	if err := adbpgService.CreateAdbpgAccount(instanceId, input); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, input.AccountName))

	return resourceAliCloudAdbpgAccountRead(d, meta)
}

func resourceAliCloudAdbpgAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	account, err := adbpgService.DescribeAdbpgAccount(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Account %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	d.Set("db_instance_id", parts[0])
	d.Set("account_name", account.AccountName)
	d.Set("account_type", account.AccountType)
	d.Set("account_description", account.AccountDescription)
	d.Set("status", account.AccountStatus)

	return nil
}

func resourceAliCloudAdbpgAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	accountName := parts[1]

	if d.HasChange("account_password") {
		if err := adbpgService.ResetAdbpgAccountPassword(instanceId, accountName, d.Get("account_password").(string)); err != nil {
			return WrapError(err)
		}
	}

	if d.HasChange("account_description") {
		if err := adbpgService.ModifyAdbpgAccountDescription(instanceId, accountName, d.Get("account_description").(string)); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudAdbpgAccountRead(d, meta)
}

func resourceAliCloudAdbpgAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	return WrapError(adbpgService.DeleteAdbpgAccount(parts[0], parts[1]))
}
