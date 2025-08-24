package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOssBucketLogging() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOssBucketLoggingCreate,
		Read:   resourceAliCloudOssBucketLoggingRead,
		Update: resourceAliCloudOssBucketLoggingUpdate,
		Delete: resourceAliCloudOssBucketLoggingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"target_bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"target_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"user_defined_headers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"user_defined_parameters": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceAliCloudOssBucketLoggingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	bucket := d.Get("bucket").(string)
	targetBucket := d.Get("target_bucket").(string)
	targetPrefix := d.Get("target_prefix").(string)

	// Enable bucket logging
	if err := ossService.PutBucketLogging(bucket, targetBucket, targetPrefix); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_logging", "PutBucketLogging", AliyunOssGoSdk)
	}

	// Set user defined log fields if provided
	if v, ok := d.GetOk("user_defined_headers"); ok || v != nil {
		userDefinedHeaders := expandStringList(v.(*schema.Set).List())
		userDefinedParams := []string{}
		if p, ok := d.GetOk("user_defined_parameters"); ok || p != nil {
			userDefinedParams = expandStringList(p.(*schema.Set).List())
		}

		if err := ossService.PutUserDefinedLogFields(bucket, userDefinedHeaders, userDefinedParams); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_logging", "PutUserDefinedLogFields", AliyunOssGoSdk)
		}
	}

	d.SetId(bucket)

	return resourceAliCloudOssBucketLoggingRead(d, meta)
}

func resourceAliCloudOssBucketLoggingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	bucket := d.Id()

	// Get bucket logging configuration
	loggingResult, err := ossService.GetBucketLoggingV2(bucket)
	if err != nil {
		log.Printf("[WARN] Failed to get bucket logging configuration: %#v", err)
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchLoggingConfiguration"}) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketLoggingV2", AliyunOssGoSdk)
	}

	d.Set("bucket", bucket)

	// If logging is enabled, set the target bucket and prefix
	if loggingResult.BucketLoggingStatus != nil && loggingResult.BucketLoggingStatus.LoggingEnabled != nil {
		if loggingResult.BucketLoggingStatus.LoggingEnabled.TargetBucket != nil {
			d.Set("target_bucket", *loggingResult.BucketLoggingStatus.LoggingEnabled.TargetBucket)
		}
		if loggingResult.BucketLoggingStatus.LoggingEnabled.TargetPrefix != nil {
			d.Set("target_prefix", *loggingResult.BucketLoggingStatus.LoggingEnabled.TargetPrefix)
		}
	} else {
		// If logging is disabled, consider this resource as removed
		d.SetId("")
		return nil
	}

	// Get user defined log fields if they exist
	userDefinedFields, err := ossService.GetUserDefinedLogFieldsV2(bucket)
	if err != nil && !IsExpectedErrors(err, []string{"NoSuchUserDefinedLogFieldsConfig"}) {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetUserDefinedLogFieldsV2", AliyunOssGoSdk)
	}

	// Set user defined headers and parameters if they exist
	if err == nil && userDefinedFields.UserDefinedLogFieldsConfiguration != nil {
		if userDefinedFields.UserDefinedLogFieldsConfiguration.HeaderSet != nil {
			d.Set("user_defined_headers", userDefinedFields.UserDefinedLogFieldsConfiguration.HeaderSet.Headers)
		}
		if userDefinedFields.UserDefinedLogFieldsConfiguration.ParamSet != nil {
			d.Set("user_defined_parameters", userDefinedFields.UserDefinedLogFieldsConfiguration.ParamSet.Parameters)
		}
	}

	return nil
}

func resourceAliCloudOssBucketLoggingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	bucket := d.Id()

	d.Partial(true)

	// Update bucket logging if target_bucket or target_prefix changed
	if d.HasChanges("target_bucket", "target_prefix") {
		targetBucket := d.Get("target_bucket").(string)
		targetPrefix := d.Get("target_prefix").(string)

		if err := ossService.PutBucketLogging(bucket, targetBucket, targetPrefix); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_logging", "PutBucketLogging", AliyunOssGoSdk)
		}
	}

	// Update user defined log fields if they changed
	if d.HasChanges("user_defined_headers", "user_defined_parameters") {
		userDefinedHeaders := []string{}
		if v, ok := d.GetOk("user_defined_headers"); ok {
			userDefinedHeaders = expandStringList(v.(*schema.Set).List())
		}

		userDefinedParams := []string{}
		if v, ok := d.GetOk("user_defined_parameters"); ok {
			userDefinedParams = expandStringList(v.(*schema.Set).List())
		}

		// If both headers and parameters are empty, delete the configuration
		if len(userDefinedHeaders) == 0 && len(userDefinedParams) == 0 {
			if err := ossService.DeleteUserDefinedLogFields(bucket); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_logging", "DeleteUserDefinedLogFields", AliyunOssGoSdk)
			}
		} else {
			// Otherwise, update the configuration
			if err := ossService.PutUserDefinedLogFields(bucket, userDefinedHeaders, userDefinedParams); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_logging", "PutUserDefinedLogFields", AliyunOssGoSdk)
			}
		}
	}

	d.Partial(false)
	return resourceAliCloudOssBucketLoggingRead(d, meta)
}

func resourceAliCloudOssBucketLoggingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	bucket := d.Id()

	// First, remove user defined log fields if they exist
	if _, ok := d.GetOk("user_defined_headers"); ok {
		if err := ossService.DeleteUserDefinedLogFields(bucket); err != nil && !IsExpectedErrors(err, []string{"NoSuchUserDefinedLogFieldsConfig"}) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteUserDefinedLogFields", AliyunOssGoSdk)
		}
	}

	// Then, disable bucket logging
	if err := ossService.DeleteBucketLogging(bucket); err != nil {
		if !IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucketLogging", AliyunOssGoSdk)
		}
	}

	// Wait for the logging to be disabled
	return resource.Retry(3*time.Minute, func() *resource.RetryError {
		logging, err := ossService.GetBucketLoggingV2(bucket)
		if err != nil {
			if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchLoggingConfiguration"}) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if logging.BucketLoggingStatus != nil && logging.BucketLoggingStatus.LoggingEnabled != nil {
			return resource.RetryableError(fmt.Errorf("waiting for bucket %s logging to be deleted", bucket))
		}
		return nil
	})
}
