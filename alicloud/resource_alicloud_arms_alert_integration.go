package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudArmsAlertIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsIntegrationCreate,
		Read:   resourceAliCloudArmsIntegrationRead,
		Update: resourceAliCloudArmsIntegrationUpdate,
		Delete: resourceAliCloudArmsIntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"integration_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"integration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"CLOUDWATCH", "DATADOG", "GRAFANA", "PROMETHEUS", "WEBHOOKS"}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"config": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Active",
				ValidateFunc: validation.StringInSlice([]string{"Active", "Inactive"}, false),
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudArmsIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build integration object using strong types
	integration := &aliyunArmsAPI.AlertIntegration{
		IntegrationName:        d.Get("integration_name").(string),
		IntegrationProductType: d.Get("integration_type").(string),
		Description:            d.Get("description").(string),
	}

	// Create integration using Service layer
	result, err := service.CreateArmsIntegration(integration)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_integration", "CreateIntegration", AlibabaCloudSdkGoERROR)
	}

	// Set ID from result
	d.SetId(fmt.Sprintf("%d", result.IntegrationId))

	// Wait for integration to be ready
	err = service.WaitForArmsIntegrationCreated(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudArmsIntegrationRead(d, meta)
}

func resourceAliCloudArmsIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := service.DescribeArmsIntegration(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_integration DescribeArmsIntegration Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all necessary fields using strong types
	d.Set("integration_name", object.IntegrationName)
	d.Set("integration_type", object.IntegrationProductType)
	d.Set("description", object.Description)
	d.Set("status", func() string {
		if object.State {
			return "Active"
		}
		return "Inactive"
	}())

	if object.CreateTime != nil {
		d.Set("create_time", object.CreateTime.Format("2006-01-02 15:04:05"))
	}
	if object.UpdateTime != nil {
		d.Set("update_time", object.UpdateTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func resourceAliCloudArmsIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse integration ID
	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Check if update is needed
	if d.HasChange("description") || d.HasChange("config") || d.HasChange("status") {
		// Build integration object for update
		integration := &aliyunArmsAPI.AlertIntegration{
			IntegrationId:          integrationId,
			IntegrationName:        d.Get("integration_name").(string),
			IntegrationProductType: d.Get("integration_type").(string),
			Description:            d.Get("description").(string),
			State:                  d.Get("status").(string) == "Active",
		}

		// Update integration using Service layer
		_, err := service.UpdateArmsIntegration(integration)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateIntegration", AlibabaCloudSdkGoERROR)
		}

		// Wait for integration to be updated
		err = service.WaitForArmsIntegrationCreated(d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudArmsIntegrationRead(d, meta)
}

func resourceAliCloudArmsIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse integration ID
	integrationId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Delete integration using Service layer
	err = service.DeleteArmsIntegration(integrationId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteIntegration", AlibabaCloudSdkGoERROR)
	}

	// Wait for integration to be deleted
	err = service.WaitForArmsIntegrationDeleted(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
