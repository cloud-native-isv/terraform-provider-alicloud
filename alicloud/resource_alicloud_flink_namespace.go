package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"resource_spec": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"memory_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ha": {
				Type:     schema.TypeBool,
				Computed: true,
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

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	namespace := &aliyunFlinkAPI.Namespace{
		Name:        namespaceName,
		WorkspaceId: workspaceId,
	}

	// Handle resource specification
	if resourceSpecList := d.Get("resource_spec").([]interface{}); len(resourceSpecList) > 0 {
		resourceSpecMap := resourceSpecList[0].(map[string]interface{})
		namespace.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{
			Cpu:      float64(resourceSpecMap["cpu"].(int)),
			MemoryGB: float64(resourceSpecMap["memory_gb"].(int)),
		}
	}

	// Create namespace
	_, err = flinkService.CreateNamespace(workspaceId, namespace)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "CreateNamespace", AlibabaCloudSdkGoERROR)
	}

	d.SetId(workspaceId + ":" + namespaceName)

	// Wait for namespace to be available
	stateConf := BuildStateConf([]string{"CREATING"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkNamespaceStateRefreshFunc(workspaceId, namespaceName, []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	namespace, err := flinkService.GetNamespace(workspaceId, namespaceName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_namespace GetNamespace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespace.Name)
	d.Set("status", namespace.Status)
	d.Set("ha", namespace.Ha)

	// Set resource specification
	if namespace.ResourceSpec != nil {
		resourceSpec := map[string]interface{}{
			"cpu":       int(namespace.ResourceSpec.Cpu),
			"memory_gb": int(namespace.ResourceSpec.MemoryGB),
		}
		d.Set("resource_spec", []interface{}{resourceSpec})
	}

	return nil
}

func resourceAliCloudFlinkNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("resource_spec") {
		namespace := &aliyunFlinkAPI.Namespace{
			Name:        namespaceName,
			WorkspaceId: workspaceId,
		}

		// Handle resource specification
		if resourceSpecList := d.Get("resource_spec").([]interface{}); len(resourceSpecList) > 0 {
			resourceSpecMap := resourceSpecList[0].(map[string]interface{})
			namespace.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{
				Cpu:      float64(resourceSpecMap["cpu"].(int)),
				MemoryGB: float64(resourceSpecMap["memory_gb"].(int)),
			}
		}

		// Update namespace - Note: Update might not be supported for all fields
		_, err := flinkService.CreateNamespace(workspaceId, namespace)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateNamespace", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, flinkService.FlinkNamespaceStateRefreshFunc(workspaceId, namespaceName, []string{"FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, err := parseNamespaceResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteNamespace(workspaceId, namespaceName)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteNamespace", AlibabaCloudSdkGoERROR)
	}

	// Wait for namespace to be deleted
	stateConf := &resource.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			namespace, err := flinkService.GetNamespace(workspaceId, namespaceName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return namespace, namespace.Status, nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper function to parse namespace resource ID
func parseNamespaceResourceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid namespace resource ID format: %s, expected workspace_id:namespace_name", id)
	}
	return parts[0], parts[1], nil
}
