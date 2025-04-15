package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudFlinkNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkNamespaceCreate,
		Read:   resourceAliCloudFlinkNamespaceRead,
		Update: resourceAliCloudFlinkNamespaceUpdate,
		Delete: resourceAliCloudFlinkNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink workspace",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Flink namespace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the namespace",
			},
			"ha":{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: " ",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	ha := d.Get("ha").(string)

			// type CreateNamespaceRequest struct {
		// 	// if can be null:
		// 	// true
		// 	Ha *bool `json:"Ha,omitempty" xml:"Ha,omitempty"`
		// 	// This parameter is required.
		// 	//
		// 	// example:
		// 	//
		// 	// f-cn-wwo36qj4g06
		// 	InstanceId *string `json:"InstanceId,omitempty" xml:"InstanceId,omitempty"`
		// 	// This parameter is required.
		// 	//
		// 	// example:
		// 	//
		// 	// di-593440390152545
		// 	Namespace *string `json:"Namespace,omitempty" xml:"Namespace,omitempty"`
		// 	// This parameter is required.
		// 	//
		// 	// example:
		// 	//
		// 	// cn-shenzhen
		// 	Region       *string                             `json:"Region,omitempty" xml:"Region,omitempty"`
		// 	ResourceSpec *CreateNamespaceRequestResourceSpec `json:"ResourceSpec,omitempty" xml:"ResourceSpec,omitempty" type:"Struct"`
		// }
	// Create request and set parameters directly according to API structure
	request := &foasconsole.CreateNamespaceRequest{
		Ha = 
	}
	request.InstanceId = &workspaceId
	request.Namespace = &name
	
	// Add description if provided
	if description != "" {
		// Note: Description field does not exist directly on the request
		// Need to add it as part of ResourceSpec if needed
	}

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := flinkService.CreateNamespace(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "ThrottlingException"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "CreateNamespace", AlibabaCloudSdkGoERROR)
	}
	
	// Set ID using format workspaceId/name
	d.SetId(fmt.Sprintf("%s/%s", workspaceId, name))
	
	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource id: %s", d.Id()))
	}
	workspaceId, name := parts[0], parts[1]

	// Create request and set parameters directly based on API structure
	request := &foasconsole.DescribeNamespacesRequest{}
	request.InstanceId = &workspaceId

	var namespaceFound bool = false
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err := flinkService.DescribeNamespaces(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException"}) {
				time.Sleep(time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		
		// Check if the response has namespaces
		if raw != nil && raw.Body != nil && raw.Body.Namespaces != nil {
			// Iterate through namespaces to find the one matching the name
			for _, ns := range raw.Body.Namespaces {
				if ns.Namespace != nil && *ns.Namespace == name {
					d.Set("workspace_id", workspaceId)
					d.Set("name", *ns.Namespace)
					// Description is not directly available in the API response
					namespaceFound = true
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "DescribeNamespaces", AlibabaCloudSdkGoERROR)
	}

	if !namespaceFound {
		d.SetId("")
		log.Printf("[WARN] Flink Namespace (%s) not found, removing from state", d.Id())
		return nil
	}

	return nil
}

func resourceAliCloudFlinkNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	// Currently, namespace properties can't be updated after creation
	// This is a placeholder for future API support
	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource id: %s", d.Id()))
	}
	workspaceId, name := parts[0], parts[1]
	
	request := &foasconsole.DeleteNamespaceRequest{}
	request.InstanceId = &workspaceId
	request.Namespace = &name
	
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := flinkService.DeleteNamespace(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "ThrottlingException"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteNamespace", AlibabaCloudSdkGoERROR)
	}
	
	return WrapError(nil)
}
