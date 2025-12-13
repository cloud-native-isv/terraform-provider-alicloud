// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

func resourceAliCloudFCConcurrencyConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCConcurrencyConfigCreate,
		Read:   resourceAliCloudFCConcurrencyConfigRead,
		Update: resourceAliCloudFCConcurrencyConfigUpdate,
		Delete: resourceAliCloudFCConcurrencyConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"reserved_concurrency": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

// BuildConcurrencyConfigFromSchema builds ConcurrencyConfig from Terraform schema data
func BuildConcurrencyConfigFromSchema(d *schema.ResourceData) *aliyunFCAPI.ConcurrencyConfig {
	config := &aliyunFCAPI.ConcurrencyConfig{}

	if v, ok := d.GetOk("reserved_concurrency"); ok {
		reservedConcurrency := int32(v.(int))
		config.ReservedInstanceCount = tea.Int32(reservedConcurrency)
	}

	return config
}

// SetSchemaFromConcurrencyConfig sets terraform schema data from ConcurrencyConfig
func SetSchemaFromConcurrencyConfig(d *schema.ResourceData, config *aliyunFCAPI.ConcurrencyConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Note: FunctionArn is not available in ConcurrencyConfig in FC v3
	// We'll leave it as computed but not set it from the config

	if config.ReservedInstanceCount != nil {
		d.Set("reserved_concurrency", *config.ReservedInstanceCount)
	}

	return nil
}

func resourceAliCloudFCConcurrencyConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)
	log.Printf("[DEBUG] Creating FC Concurrency Config for function: %s", functionName)

	// Build concurrency config from schema
	config := BuildConcurrencyConfigFromSchema(d)

	// Create concurrency config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = fcService.CreateFCConcurrencyConfig(functionName, config)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Concurrency Config creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Concurrency Config creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_concurrency_config", "CreateConcurrencyConfig", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(functionName)
	log.Printf("[DEBUG] FC Concurrency Config created successfully for function: %s", functionName)

	return resourceAliCloudFCConcurrencyConfigRead(d, meta)
}

func resourceAliCloudFCConcurrencyConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	objectRaw, err := fcService.DescribeFCConcurrencyConfig(functionName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_concurrency_config DescribeFCConcurrencyConfig Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use helper to set schema fields
	err = SetSchemaFromConcurrencyConfig(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	d.Set("function_name", functionName)

	return nil
}

func resourceAliCloudFCConcurrencyConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	log.Printf("[DEBUG] Updating FC Concurrency Config for function: %s", functionName)

	// Check if any field has changed
	if d.HasChange("reserved_concurrency") {
		// Build concurrency config from schema
		config := BuildConcurrencyConfigFromSchema(d)

		// Update concurrency config using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCConcurrencyConfig(functionName, config)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Concurrency Config update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Concurrency Config update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateConcurrencyConfig", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] FC Concurrency Config updated successfully for function: %s", functionName)
	}

	return resourceAliCloudFCConcurrencyConfigRead(d, meta)
}

func resourceAliCloudFCConcurrencyConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	log.Printf("[DEBUG] Deleting FC Concurrency Config for function: %s", functionName)

	// Delete concurrency config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCConcurrencyConfig(functionName)
		if err != nil {
			if NotFoundError(err) {
				log.Printf("[DEBUG] FC Concurrency Config not found during deletion for function: %s", functionName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Concurrency Config deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Concurrency Config deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteConcurrencyConfig", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC Concurrency Config deleted successfully for function: %s", functionName)

	return nil
}
