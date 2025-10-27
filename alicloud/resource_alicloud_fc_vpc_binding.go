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
)

func resourceAliCloudFCVpcBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCVpcBindingCreate,
		Read:   resourceAliCloudFCVpcBindingRead,
		Delete: resourceAliCloudFCVpcBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudFCVpcBindingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)
	vpcId := d.Get("vpc_id").(string)

	log.Printf("[DEBUG] Creating FC VPC Binding for function: %s, vpcId: %s", functionName, vpcId)

	// Create VPC binding using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err = fcService.CreateFCVpcBinding(functionName, vpcId)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC VPC Binding creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC VPC Binding creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_vpc_binding", "CreateVpcBinding", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(fmt.Sprintf("%s:%s", functionName, vpcId))
	log.Printf("[DEBUG] FC VPC Binding created successfully for function: %s, vpcId: %s", functionName, vpcId)

	return resourceAliCloudFCVpcBindingRead(d, meta)
}

func resourceAliCloudFCVpcBindingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID to get function name and VPC ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid resource ID format, expected functionName:vpcId")
	}
	functionName := parts[0]
	vpcId := parts[1]

	// Get VPC bindings for the function
	vpcBindings, err := fcService.DescribeFCVpcBinding(functionName)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_vpc_binding DescribeFCVpcBinding Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Check if the VPC ID is in the bindings list
	found := false
	for _, binding := range vpcBindings {
		if binding != nil && *binding == vpcId {
			found = true
			break
		}
	}

	if !found {
		log.Printf("[DEBUG] VPC binding not found for function: %s, vpcId: %s", functionName, vpcId)
		d.SetId("")
		return nil
	}

	// Set schema fields
	d.Set("function_name", functionName)
	d.Set("vpc_id", vpcId)

	return nil
}

func resourceAliCloudFCVpcBindingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID to get function name and VPC ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid resource ID format, expected functionName:vpcId")
	}
	functionName := parts[0]
	vpcId := parts[1]

	log.Printf("[DEBUG] Deleting FC VPC Binding for function: %s, vpcId: %s", functionName, vpcId)

	// Delete VPC binding using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = fcService.DeleteFCVpcBinding(functionName, vpcId)
		if err != nil {
			if IsNotFoundError(err) {
				log.Printf("[DEBUG] FC VPC Binding not found during deletion for function: %s, vpcId: %s", functionName, vpcId)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC VPC Binding deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC VPC Binding deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteVpcBinding", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC VPC Binding deleted successfully for function: %s, vpcId: %s", functionName, vpcId)

	return nil
}
