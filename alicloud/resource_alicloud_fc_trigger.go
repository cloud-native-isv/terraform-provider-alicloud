// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

func resourceAliCloudFCTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCTriggerCreate,
		Read:   resourceAliCloudFCTriggerRead,
		Update: resourceAliCloudFCTriggerUpdate,
		Delete: resourceAliCloudFCTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
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
			"http_trigger": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url_intranet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url_internet": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"invocation_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trigger_config": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := compareJsonTemplateAreEquivalent(old, new)
					return equal
				},
			},
			"trigger_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trigger_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"trigger_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudFCTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)

	log.Printf("[DEBUG] Creating FC Trigger for function: %s", functionName)

	// Build trigger from schema
	trigger := fcService.BuildCreateTriggerInputFromSchema(d)

	// Create trigger using service layer
	var result *aliyunFCAPI.Trigger
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = fcService.CreateFCTrigger(functionName, trigger)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Trigger creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Trigger creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_trigger", "CreateTrigger", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	if result != nil && result.TriggerName != nil {
		d.SetId(EncodeTriggerResourceId(functionName, *result.TriggerName))
		log.Printf("[DEBUG] FC Trigger created successfully: %s:%s", functionName, *result.TriggerName)
	} else {
		return fmt.Errorf("failed to get trigger name from create response")
	}

	// Wait for trigger to be ready
	if trigger.TriggerName != nil {
		log.Printf("[DEBUG] Waiting for FC Trigger to be ready: %s:%s", functionName, *trigger.TriggerName)
		err = fcService.WaitForTriggerCreating(functionName, *trigger.TriggerName, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFCTriggerRead(d, meta)
}

func resourceAliCloudFCTriggerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	objectRaw, err := fcService.DescribeFCTrigger(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_trigger DescribeFCTrigger Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use the service layer helper to set schema fields
	err = fcService.SetSchemaFromTrigger(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	// Set function_name from resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) >= 1 {
		d.Set("function_name", parts[0])
	}

	return nil
}

func resourceAliCloudFCTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName, triggerName, err := DecodeTriggerResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	log.Printf("[DEBUG] Updating FC Trigger: %s:%s", functionName, triggerName)

	// Check if any field has changed
	if d.HasChange("invocation_role") || d.HasChange("trigger_config") || d.HasChange("description") || d.HasChange("qualifier") {
		// Build update trigger from schema
		trigger := fcService.BuildUpdateTriggerInputFromSchema(d)

		// Update trigger using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCTrigger(functionName, triggerName, trigger)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Trigger update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Trigger update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateTrigger", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] Waiting for FC Trigger update to complete: %s:%s", functionName, triggerName)

		// Wait for trigger update to complete
		err = fcService.WaitForTriggerUpdating(functionName, triggerName, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		log.Printf("[DEBUG] FC Trigger updated successfully: %s:%s", functionName, triggerName)
	}

	return resourceAliCloudFCTriggerRead(d, meta)
}

func resourceAliCloudFCTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName, triggerName, err := DecodeTriggerResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	log.Printf("[DEBUG] Deleting FC Trigger: %s:%s", functionName, triggerName)

	// Delete trigger using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCTrigger(functionName, triggerName)
		if err != nil {
			if NotFoundError(err) {
				log.Printf("[DEBUG] FC Trigger not found during deletion: %s:%s", functionName, triggerName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Trigger deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Trigger deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteTrigger", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] Waiting for FC Trigger to be deleted: %s:%s", functionName, triggerName)

	// Wait for trigger to be deleted
	err = fcService.WaitForTriggerDeleting(functionName, triggerName, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	log.Printf("[DEBUG] FC Trigger deleted successfully: %s:%s", functionName, triggerName)

	return nil
}
