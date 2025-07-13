package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertIntegrationCreate,
		Read:   resourceAliCloudArmsAlertIntegrationRead,
		Update: resourceAliCloudArmsAlertIntegrationUpdate,
		Delete: resourceAliCloudArmsAlertIntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"integration_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"integration_product_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_recover": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"recover_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"duplicate_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"liveness": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudArmsAlertIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationName := d.Get("integration_name").(string)
	integrationProductType := d.Get("integration_product_type").(string)
	description := d.Get("description").(string)
	autoRecover := d.Get("auto_recover").(bool)
	recoverTime := int64(d.Get("recover_time").(int))

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		integration, err := armsAPI.CreateIntegration(integrationName, integrationProductType, description, autoRecover, recoverTime)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		d.SetId(fmt.Sprint(integration.IntegrationId))
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "CreateIntegration", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudArmsAlertIntegrationRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	integration, err := armsAPI.GetIntegrationById(integrationId, true)
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_integration GetIntegrationByID Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("integration_name", integration.IntegrationName)
	d.Set("integration_product_type", integration.IntegrationProductType)
	d.Set("state", integration.State)
	d.Set("api_endpoint", integration.ApiEndpoint)
	d.Set("short_token", integration.ShortToken)
	d.Set("liveness", integration.Liveness)
	d.Set("create_time", integration.CreateTime)

	// Set integration detail fields if available
	if integration.IntegrationDetail != nil {
		d.Set("description", integration.IntegrationDetail.Description)
		d.Set("auto_recover", integration.IntegrationDetail.AutoRecover)
		d.Set("recover_time", integration.IntegrationDetail.RecoverTime)
		d.Set("duplicate_key", integration.IntegrationDetail.DuplicateKey)
	}

	return nil
}

func resourceAliCloudArmsAlertIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	update := false
	description := d.Get("description").(string)
	autoRecover := d.Get("auto_recover").(bool)
	recoverTime := int64(d.Get("recover_time").(int))

	if d.HasChange("integration_name") || d.HasChange("description") || d.HasChange("auto_recover") || d.HasChange("recover_time") || d.HasChange("duplicate_key") || d.HasChange("state") {
		update = true
	}

	if update {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			err := armsAPI.UpdateIntegration(integrationId, description, autoRecover, recoverTime)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateIntegration", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudArmsAlertIntegrationRead(d, meta)
}

func resourceAliCloudArmsAlertIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integration", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := armsAPI.DeleteIntegration(integrationId)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteIntegration", AlibabaCloudSdkGoERROR)
	}

	return nil
}
