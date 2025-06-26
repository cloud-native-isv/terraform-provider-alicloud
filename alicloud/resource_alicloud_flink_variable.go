package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
				Description: "ID of the Flink workspaceId",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink namespaceName",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kind of the Flink variable",
				Default:     "Clear",
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

	workspaceId, namespaceName, varName := splitVariableID(d.Id())
	variable, err := flinkService.GetVariable(workspaceId, namespaceName, varName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "GetVariable", AlibabaCloudSdkGoERROR)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
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

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	variable := &aliyunFlinkAPI.Variable{
		Name:        name,
		Value:       value,
		Description: description,
		Kind:        kind,
	}

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := flinkService.CreateVariable(workspaceId, namespaceName, variable)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "CreateVariable", AlibabaCloudSdkGoERROR)
	}

	d.SetId(joinVariableID(workspaceId, namespaceName, name))

	// Use state refresh to wait for the variable to be available using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkVariableStateRefreshFunc(workspaceId, namespaceName, name, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkVariableRead(d, meta)
}

func resourceAliCloudFlinkVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, varName := splitVariableID(d.Id())
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	variable := &aliyunFlinkAPI.Variable{
		Name:        varName,
		Value:       value,
		Description: description,
		Kind:        kind,
	}

	_, err = flinkService.UpdateVariable(workspaceId, namespaceName, variable)
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

	workspaceId, namespaceName, varName := splitVariableID(d.Id())
	err = flinkService.DeleteVariable(workspaceId, namespaceName, varName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_variable", "DeleteVariable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func splitVariableID(id string) (string, string, string) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 {
		panic(fmt.Errorf("invalid ID format: %s, should be <workspaceId>/<namespaceName>/<variable name>", id))
	}
	return parts[0], parts[1], parts[2]
}

func joinVariableID(workspaceId, namespaceName, varName string) string {
	return fmt.Sprintf("%s/%s/%s", workspaceId, namespaceName, varName)
}
