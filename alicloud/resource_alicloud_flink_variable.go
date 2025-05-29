package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudFlinkVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkVariableCreate,
		Read:   resourceAliCloudFlinkVariableRead,
		Update: resourceAliCloudFlinkVariableUpdate,
		Delete: resourceAliCloudFlinkVariableDelete,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink workspace",
			},
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink namespace",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kind of the Flink variable",
				Default:     "Plain",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Flink variable",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the Flink variable",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the variable",
			},
		},
	}
}

func resourceAliCloudFlinkVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, namespace, varName := splitVariableID(d.Id())
	variable, err := flinkService.GetVariable(workspace, namespace, varName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "GetVariable", AlibabaCloudSdkGoERROR)
	}

	d.Set("workspace_id", workspace)
	d.Set("namespace_id", namespace)
	d.Set("name", variable.Name)
	d.Set("value", variable.Value)
	d.Set("description", variable.Description)
	d.Set("kind", variable.Kind)

	return nil
}

func resourceAliCloudFlinkVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace_id").(string)
	namespace := d.Get("namespace_id").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	variable := &aliyunAPI.Variable{
		Name:        name,
		Value:       value,
		Description: description,
		Kind:        kind,
	}

	_, err = flinkService.CreateVariable(workspace, namespace, variable)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "CreateVariable", AlibabaCloudSdkGoERROR)
	}

	d.SetId(joinVariableID(workspace, namespace, name))
	return resourceAliCloudFlinkVariableRead(d, meta)
}

func resourceAliCloudFlinkVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, namespace, varName := splitVariableID(d.Id())
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	variable := &aliyunAPI.Variable{
		Name:        varName,
		Value:       value,
		Description: description,
		Kind:        kind,
	}

	_, err = flinkService.UpdateVariable(workspace, namespace, variable)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "UpdateVariable", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudFlinkVariableRead(d, meta)
}

func resourceAliCloudFlinkVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, namespace, varName := splitVariableID(d.Id())
	err = flinkService.DeleteVariable(workspace, namespace, varName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "DeleteVariable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func splitVariableID(id string) (string, string, string) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 {
		panic(fmt.Errorf("invalid ID format: %s, should be <workspace>/<namespace>/<variable name>", id))
	}
	return parts[0], parts[1], parts[2]
}

func joinVariableID(workspace, namespace, varName string) string {
	return fmt.Sprintf("%s/%s/%s", workspace, namespace, varName)
}
