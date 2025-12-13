// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

func resourceAliCloudFCAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCAliasCreate,
		Read:   resourceAliCloudFCAliasRead,
		Update: resourceAliCloudFCAliasUpdate,
		Delete: resourceAliCloudFCAliasDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"additional_version_weight": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeFloat},
			},
			"alias_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAliCloudFCAliasCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)

	log.Printf("[DEBUG] Creating FC Alias for function: %s", functionName)

	// Build alias from schema
	alias := fcService.BuildCreateAliasInputFromSchema(d)

	// Create alias using service layer
	var result *aliyunFCAPI.Alias
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = fcService.CreateFCAlias(functionName, alias)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Alias creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Alias creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_alias", "CreateAlias", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	if result != nil && result.AliasName != nil {
		d.SetId(EncodeAliasResourceId(functionName, *result.AliasName))
		log.Printf("[DEBUG] FC Alias created successfully: %s:%s", functionName, *result.AliasName)
	} else {
		return WrapErrorf(fmt.Errorf("failed to get alias name from create response"), DefaultErrorMsg, "alicloud_fc_alias", "CreateAlias", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudFCAliasRead(d, meta)
}

func resourceAliCloudFCAliasRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	// Decode the resource ID to get function name and alias name
	functionName, aliasName, err := DecodeAliasResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	objectRaw, err := fcService.DescribeFCAlias(functionName, aliasName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_alias DescribeFCAlias Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use the service layer helper to set schema fields
	err = fcService.SetSchemaFromAlias(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	// Set function_name from resource ID
	d.Set("function_name", functionName)

	return nil
}

func resourceAliCloudFCAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName, aliasName, err := DecodeAliasResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	log.Printf("[DEBUG] Updating FC Alias: %s:%s", functionName, aliasName)

	// Check if any field has changed
	if d.HasChange("version_id") || d.HasChange("description") || d.HasChange("additional_version_weight") {
		// Build update alias from schema
		alias := fcService.BuildUpdateAliasInputFromSchema(d)

		// Update alias using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCAlias(functionName, aliasName, alias)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Alias update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Alias update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateAlias", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] FC Alias updated successfully: %s:%s", functionName, aliasName)
	}

	return resourceAliCloudFCAliasRead(d, meta)
}

func resourceAliCloudFCAliasDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName, aliasName, err := DecodeAliasResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	log.Printf("[DEBUG] Deleting FC Alias: %s:%s", functionName, aliasName)

	// Delete alias using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCAlias(functionName, aliasName)
		if err != nil {
			if NotFoundError(err) {
				log.Printf("[DEBUG] FC Alias not found during deletion: %s:%s", functionName, aliasName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Alias deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Alias deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAlias", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC Alias deleted successfully: %s:%s", functionName, aliasName)

	return nil
}
