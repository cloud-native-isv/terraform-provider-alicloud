package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
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

	workspace := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	ha := d.Get("ha").(bool)
	cpu := d.Get("cpu").(int)
	memory := d.Get("memory").(int)

	// Create namespace using aliyunFlinkAPI.Namespace directly
	namespace := &aliyunFlinkAPI.Namespace{
		Name: name,
		Ha:   ha,
		ResourceSpec: &aliyunFlinkAPI.ResourceSpec{
			Cpu:      float64(cpu),
			MemoryGB: float64(memory),
		},
	}

	_, err = flinkService.CreateNamespace(workspace, namespace)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_namespace", "CreateNamespace", AlibabaCloudSdkGoERROR)
	}

	d.SetId(joinNamespaceID(workspace, name))

	// Wait for namespace creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkNamespaceStateRefreshFunc(workspace, name, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
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

	// Use the refactored DeleteNamespace method with simplified parameters
	err = flinkService.DeleteNamespace(workspace, namespace)
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
