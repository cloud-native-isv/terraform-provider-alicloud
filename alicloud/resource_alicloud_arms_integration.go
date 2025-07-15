package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudArmsIntegration() *schema.Resource {
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
				ValidateFunc: validation.StringInSlice([]string{"cloudwatch", "datadog", "grafana", "prometheus", "webhooks"}, false),
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
	action := "CreateIntegration"
	request := make(map[string]interface{})

	request["RegionId"] = client.RegionId
	request["IntegrationName"] = d.Get("integration_name")
	request["IntegrationType"] = d.Get("integration_type")
	request["Config"] = d.Get("config")

	if v, ok := d.GetOk("description"); ok {
		request["Description"] = v
	}

	if v, ok := d.GetOk("status"); ok {
		request["Status"] = v
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	var response map[string]interface{}
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_integration", action, AlibabaCloudSdkGoERROR)
	}

	integrationIdResp, err := jsonpath.Get("$.IntegrationId", response)
	if err != nil {
		return WrapErrorf(err, FailedGetAttributeMsg, "alicloud_arms_integration", "$.IntegrationId", response)
	}

	id := fmt.Sprint(integrationIdResp)

	d.SetId(id)

	// Wait for integration to be ready
	armsService := NewArmsService(client)
	stateConf := BuildStateConf([]string{}, []string{"Active"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, armsService.ArmsIntegrationStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudArmsIntegrationRead(d, meta)
}

func resourceAliCloudArmsIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := NewArmsService(client)

	object, err := armsService.DescribeArmsIntegration(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_integration armsService.DescribeArmsIntegration Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("integration_name", object["IntegrationName"])
	d.Set("integration_type", object["IntegrationType"])
	d.Set("description", object["Description"])
	d.Set("config", object["Config"])
	d.Set("status", object["Status"])
	d.Set("create_time", object["CreateTime"])
	d.Set("update_time", object["UpdateTime"])

	return nil
}

func resourceAliCloudArmsIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	update := false

	request := map[string]interface{}{
		"RegionId":      client.RegionId,
		"IntegrationId": d.Id(),
	}

	if d.HasChange("description") {
		update = true
		if v, ok := d.GetOk("description"); ok {
			request["Description"] = v
		}
	}

	if d.HasChange("config") {
		update = true
		request["Config"] = d.Get("config")
	}

	if d.HasChange("status") {
		update = true
		request["Status"] = d.Get("status")
	}

	if update {
		action := "UpdateIntegration"

		wait := incrementalWait(3*time.Second, 3*time.Second)
		err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		// Wait for integration to be updated
		armsService := NewArmsService(client)
		stateConf := BuildStateConf([]string{}, []string{"Active"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, armsService.ArmsIntegrationStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudArmsIntegrationRead(d, meta)
}

func resourceAliCloudArmsIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteIntegration"

	request := map[string]interface{}{
		"RegionId":      client.RegionId,
		"IntegrationId": d.Id(),
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		if IsExpectedErrors(err, []string{"404"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	// Wait for integration to be deleted
	armsService := NewArmsService(client)
	stateConf := BuildStateConf([]string{"Active", "Inactive"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, armsService.ArmsIntegrationStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
