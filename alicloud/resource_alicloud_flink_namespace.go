package alicloud

import (
	"fmt"
	"strings"
	"time"

	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
			"workspace_id": {
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
			"ha": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cpu": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "CPU number",
			},
			"memory": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "memory size in GB",
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

	region := client.RegionId
	workspace := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	ha := d.Get("ha").(bool)
	cpu := d.Get("cpu").(int32)
	memory := d.Get("memory").(int32)

	request := &foasconsole.CreateNamespaceRequest{
		Region:     tea.String(region),
		InstanceId: tea.String(workspace),
		Namespace:  tea.String(name),
		Ha:         tea.Bool(ha),
		ResourceSpec: &foasconsole.CreateNamespaceRequestResourceSpec{
			Cpu:      tea.Int32(cpu),
			MemoryGB: tea.Int32(memory),
		},
	}

	_, err = flinkService.CreateNamespace(request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "CreateNamespace", AlibabaCloudSdkGoERROR)
	}

	d.SetId(joinNamespaceID(workspace, name))
	return resourceAliCloudFlinkNamespaceRead(d, meta)
}

func resourceAliCloudFlinkNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}
	id := d.Id()
	workspace, namespace := splitNamespaceID(id)

	_namespace, derr := flinkService.GetNamespace(workspace, namespace)
	if derr != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "DescribeNamespaces", AlibabaCloudSdkGoERROR)
	}

	d.Set("cpu", _namespace.ResourceSpec.Cpu)
	d.Set("memory", _namespace.ResourceSpec.MemoryGB)

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

	id := d.Id()
	workspace, namespace := splitNamespaceID(id)
	region := client.RegionId

	request := &foasconsole.DeleteNamespaceRequest{
		InstanceId: tea.String(workspace),
		Namespace:  tea.String(namespace),
		Region:     tea.String(region),
	}
	_, err = flinkService.DeleteNamespace(request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "DeleteNamespace", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func splitNamespaceID(id string) (string, string) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		panic(fmt.Errorf("invalid ID format: %s, should be <workspace>/<namespace>", id))
	}
	return parts[0], parts[1]
}

func joinNamespaceID(workspace, namespace string) string {
	return fmt.Sprintf("%s/%s", workspace, namespace)
}
