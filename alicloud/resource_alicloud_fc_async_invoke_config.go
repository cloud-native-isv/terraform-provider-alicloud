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

func resourceAliCloudFCAsyncInvokeConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCAsyncInvokeConfigCreate,
		Read:   resourceAliCloudFCAsyncInvokeConfigRead,
		Update: resourceAliCloudFCAsyncInvokeConfigUpdate,
		Delete: resourceAliCloudFCAsyncInvokeConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"async_task": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_success": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"on_failure": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"max_async_event_age_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: IntBetween(0, 86400),
			},
			"max_async_retry_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

// BuildAsyncConfigFromSchema builds AsyncConfig from Terraform schema data
func BuildAsyncConfigFromSchema(d *schema.ResourceData) *aliyunFCAPI.AsyncConfig {
	config := &aliyunFCAPI.AsyncConfig{}

	if v, ok := d.GetOk("async_task"); ok {
		config.AsyncTask = tea.Bool(v.(bool))
	}

	if v, ok := d.GetOk("max_async_event_age_in_seconds"); ok {
		config.MaxAsyncEventAgeInSeconds = tea.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_async_retry_attempts"); ok {
		config.MaxAsyncRetryAttempts = tea.Int64(int64(v.(int)))
	}

	// Add destination config
	if v, ok := d.GetOk("destination_config"); ok {
		if destConfigs := v.([]interface{}); len(destConfigs) > 0 {
			destConfig := destConfigs[0].(map[string]interface{})
			config.DestinationConfig = &aliyunFCAPI.DestinationConfig{}

			// Add on success
			if onSuccessData, ok := destConfig["on_success"].([]interface{}); ok && len(onSuccessData) > 0 {
				onSuccess := onSuccessData[0].(map[string]interface{})
				config.DestinationConfig.OnSuccess = &aliyunFCAPI.Destination{}
				if dest, ok := onSuccess["destination"].(string); ok && dest != "" {
					config.DestinationConfig.OnSuccess.Destination = tea.String(dest)
				}
			}

			// Add on failure
			if onFailureData, ok := destConfig["on_failure"].([]interface{}); ok && len(onFailureData) > 0 {
				onFailure := onFailureData[0].(map[string]interface{})
				config.DestinationConfig.OnFailure = &aliyunFCAPI.Destination{}
				if dest, ok := onFailure["destination"].(string); ok && dest != "" {
					config.DestinationConfig.OnFailure.Destination = tea.String(dest)
				}
			}
		}
	}

	return config
}

// SetSchemaFromAsyncConfig sets terraform schema data from AsyncConfig
func SetSchemaFromAsyncConfig(d *schema.ResourceData, config *aliyunFCAPI.AsyncConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.AsyncTask != nil {
		d.Set("async_task", *config.AsyncTask)
	}

	if config.CreatedTime != nil {
		d.Set("create_time", *config.CreatedTime)
	}

	if config.FunctionArn != nil {
		d.Set("function_arn", *config.FunctionArn)
	}

	if config.LastModifiedTime != nil {
		d.Set("last_modified_time", *config.LastModifiedTime)
	}

	if config.MaxAsyncEventAgeInSeconds != nil {
		d.Set("max_async_event_age_in_seconds", *config.MaxAsyncEventAgeInSeconds)
	}

	if config.MaxAsyncRetryAttempts != nil {
		d.Set("max_async_retry_attempts", *config.MaxAsyncRetryAttempts)
	}

	// Set destination config
	if config.DestinationConfig != nil {
		destConfigMaps := make([]map[string]interface{}, 0)
		destConfigMap := make(map[string]interface{})

		// Set on success
		if config.DestinationConfig.OnSuccess != nil && config.DestinationConfig.OnSuccess.Destination != nil {
			onSuccessMaps := make([]map[string]interface{}, 0)
			onSuccessMap := make(map[string]interface{})
			onSuccessMap["destination"] = *config.DestinationConfig.OnSuccess.Destination
			onSuccessMaps = append(onSuccessMaps, onSuccessMap)
			destConfigMap["on_success"] = onSuccessMaps
		}

		// Set on failure
		if config.DestinationConfig.OnFailure != nil && config.DestinationConfig.OnFailure.Destination != nil {
			onFailureMaps := make([]map[string]interface{}, 0)
			onFailureMap := make(map[string]interface{})
			onFailureMap["destination"] = *config.DestinationConfig.OnFailure.Destination
			onFailureMaps = append(onFailureMaps, onFailureMap)
			destConfigMap["on_failure"] = onFailureMaps
		}

		if len(destConfigMap) > 0 {
			destConfigMaps = append(destConfigMaps, destConfigMap)
			d.Set("destination_config", destConfigMaps)
		}
	}

	return nil
}

func resourceAliCloudFCAsyncInvokeConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)
	log.Printf("[DEBUG] Creating FC Async Invoke Config for function: %s", functionName)

	// Build async config from schema
	config := BuildAsyncConfigFromSchema(d)

	// Create async config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = fcService.CreateFCAsyncInvokeConfig(functionName, config)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Async Invoke Config creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Async Invoke Config creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_async_invoke_config", "CreateAsyncInvokeConfig", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(functionName)
	log.Printf("[DEBUG] FC Async Invoke Config created successfully for function: %s", functionName)

	return resourceAliCloudFCAsyncInvokeConfigRead(d, meta)
}

func resourceAliCloudFCAsyncInvokeConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	objectRaw, err := fcService.DescribeFCAsyncInvokeConfig(functionName)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_async_invoke_config DescribeFCAsyncInvokeConfig Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use helper to set schema fields
	err = SetSchemaFromAsyncConfig(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	d.Set("function_name", functionName)

	return nil
}

func resourceAliCloudFCAsyncInvokeConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	log.Printf("[DEBUG] Updating FC Async Invoke Config for function: %s", functionName)

	// Check if any field has changed
	if d.HasChange("async_task") || d.HasChange("destination_config") ||
		d.HasChange("max_async_event_age_in_seconds") || d.HasChange("max_async_retry_attempts") {
		// Build async config from schema
		config := BuildAsyncConfigFromSchema(d)

		// Update async config using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCAsyncInvokeConfig(functionName, config)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Async Invoke Config update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Async Invoke Config update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateAsyncInvokeConfig", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] FC Async Invoke Config updated successfully for function: %s", functionName)
	}

	return resourceAliCloudFCAsyncInvokeConfigRead(d, meta)
}

func resourceAliCloudFCAsyncInvokeConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	log.Printf("[DEBUG] Deleting FC Async Invoke Config for function: %s", functionName)

	// Delete async config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCAsyncInvokeConfig(functionName)
		if err != nil {
			if IsNotFoundError(err) {
				log.Printf("[DEBUG] FC Async Invoke Config not found during deletion for function: %s", functionName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Async Invoke Config deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Async Invoke Config deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAsyncInvokeConfig", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC Async Invoke Config deleted successfully for function: %s", functionName)

	return nil
}
