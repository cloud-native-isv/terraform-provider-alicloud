// Package alicloud provides resources for Alibaba Cloud products
package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The project name to which the logging configurations belong.",
			},
			"logging_project": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "The project to store the service logs.",
			},
			"logging_details": {
				Type:     schema.TypeSet,
				Required: true,
				Optional: false,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of service log, such as operation_log, consumer_group, etc.",
						},
						"logstore": {
							Type:        schema.TypeString,
							Required:    true,
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
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Get("project_name").(string)
	logging := createLoggingFromSchema(d)

	_, err = slsService.GetProjectLogging(projectName)
	if err != nil {
		if !NotFoundError(err) {
			return WrapError(err)
		}
		err = slsService.CreateProjectLogging(projectName, logging)
	} else {
		err = slsService.UpdateProjectLogging(projectName, logging)
	}
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project_logging", "CreateOrUpdateProjectLogging", AlibabaCloudSdkGoERROR)
	}

	d.SetId(projectName)
	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Id()

	logging, err := slsService.GetProjectLogging(projectName)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("project_name", projectName)
	d.Set("logging_project", logging.LoggingProject)

	loggingDetailsSet := make([]map[string]interface{}, 0)
	for _, loggingDetail := range logging.LoggingDetails {
		loggingDetailsSet = append(loggingDetailsSet, map[string]interface{}{
			"type":     loggingDetail.Type,
			"logstore": loggingDetail.Logstore,
		})
	}

	if err := d.Set("logging_details", loggingDetailsSet); err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAlicloudLogProjectLoggingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Id()

	if d.HasChange("logging_details") {
		logging := createLoggingFromSchema(d)
		err := slsService.UpdateProjectLogging(projectName, logging)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProjectLogging", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Id()

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := slsService.DeleteProjectLogging(projectName)
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteProjectLogging", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func createLoggingFromSchema(d *schema.ResourceData) *aliyunSlsAPI.LogProjectLogging {
	loggingDetailsSet := d.Get("logging_details").(*schema.Set).List()
	loggingDetails := make([]aliyunSlsAPI.LogProjectLoggingDetails, 0, len(loggingDetailsSet))

	for _, item := range loggingDetailsSet {
		if m, ok := item.(map[string]interface{}); ok {
			loggingDetails = append(loggingDetails, aliyunSlsAPI.LogProjectLoggingDetails{
				Type:     m["type"].(string),
				Logstore: m["logstore"].(string),
			})
		}
	}

	return &aliyunSlsAPI.LogProjectLogging{
		LoggingProject: d.Get("logging_project").(string),
		LoggingDetails: loggingDetails,
	}
}
