// Package alicloud provides resources for Alibaba Cloud products
package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudLogProjectLogging() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogProjectLoggingCreate,
		Read:   resourceAlicloudLogProjectLoggingRead,
		Update: resourceAlicloudLogProjectLoggingUpdate,
		Delete: resourceAlicloudLogProjectLoggingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The project name to which the logging configurations belong.",
			},
			"logging_details": {
				Type:     schema.TypeList,
				Required: true,
				Optional: false,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The type of service log, such as operation_log, consumer_group, etc.",
						},
						"logstore": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The logstore to store the service logs.",
						},
					},
				},
				Description: "Configuration details for the service logging.",
			},
		},
	}
}

func resourceAlicloudLogProjectLoggingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsServiceV2, err := NewSlsServiceV2(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Get("project").(string)
	
	// Convert from schema format to API format
	loggingDetails := convertLoggingDetailsSchemaToAPI(d)
	
	_, err = slsServiceV2.CreateSlsLogging(project, loggingDetails)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project_logging", "CreateSlsLogging", AlibabaCloudSdkGoERROR)
	}
	
	d.SetId(project)
	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsServiceV2, err := NewSlsServiceV2(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Id()
	
	loggingResp, err := slsServiceV2.GetSlsLogging(project)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	
	d.Set("project", project)
	
	// Convert API response to schema format
	rawLoggingDetails, ok := loggingResp["loggingDetails"].([]interface{})
	if !ok {
		return WrapError(fmt.Errorf("failed to parse logging details from response"))
	}
	
	loggingDetailsSet := make([]map[string]interface{}, 0)
	for _, item := range rawLoggingDetails {
		if detail, ok := item.(map[string]interface{}); ok {
			loggingDetail := map[string]interface{}{
				"type":     detail["type"],
				"logstore": detail["logstore"],
			}
			loggingDetailsSet = append(loggingDetailsSet, loggingDetail)
		}
	}
	
	if err := d.Set("logging_details", loggingDetailsSet); err != nil {
		return WrapError(err)
	}
	
	return nil
}

func resourceAlicloudLogProjectLoggingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsServiceV2, err := NewSlsServiceV2(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Id()
	
	if d.HasChange("logging_details") {
		// Convert from schema format to API format
		loggingDetails := convertLoggingDetailsSchemaToAPI(d)
		
		_, err = slsServiceV2.UpdateSlsLogging(project, loggingDetails)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSlsLogging", AlibabaCloudSdkGoERROR)
		}
	}
	
	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsServiceV2, err := NewSlsServiceV2(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Id()
	
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := slsServiceV2.DeleteSlsLogging(project)
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSlsLogging", AlibabaCloudSdkGoERROR)
	}
	
	return nil
}

// Helper function to convert schema format to API format
func convertLoggingDetailsSchemaToAPI(d *schema.ResourceData) []map[string]interface{} {
	loggingDetailsSet := d.Get("logging_details").(*schema.Set).List()
	loggingDetails := make([]map[string]interface{}, 0, len(loggingDetailsSet))
	
	for _, item := range loggingDetailsSet {
		if m, ok := item.(map[string]interface{}); ok {
			detail := map[string]interface{}{
				"type":     m["type"],
				"logstore": m["logstore"],
			}
			loggingDetails = append(loggingDetails, detail)
		}
	}
	
	return loggingDetails
}