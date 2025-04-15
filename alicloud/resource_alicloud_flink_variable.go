package alicloud

import (
	"fmt"
	"strings"

	"github.com/alibabacloud-go/tea/tea"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudFlinkVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkVariableCreate,
		Read:   resourceAliCloudFlinkVariableRead,
		Update: resourceAliCloudFlinkVariableUpdate,
		Delete: resourceAliCloudFlinkVariableDelete,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink workspace",
			},
			"namespace": {
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
		return err
	}
	workspace, namespace, varName := splitVariableID(d.Id())
	variable, err := flinkService.GetVariable(tea.String(workspace), tea.String(namespace), tea.String(varName))
	if err != nil {
		return err
	}

	d.Set("workspace", workspace)
	d.Set("namespace", namespace)
	d.Set("name", variable.Name)
	d.Set("value", variable.Value)
	d.Set("description", variable.Description)

	return nil
}

func resourceAliCloudFlinkVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return err
	}

	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	_, err = flinkService.CreateVariable(tea.String(workspace), tea.String(namespace), &ververica.Variable{
		Name:        tea.String(name),
		Value:       tea.String(value),
		Description: tea.String(description),
		Kind:        tea.String(kind),
	})
	if err != nil {
		return err
	}
	d.SetId(joinVariableID(workspace, namespace, name))
	return nil
}

func resourceAliCloudFlinkVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return err
	}

	workspace, namespace, varName := splitVariableID(d.Id())
	value := d.Get("value").(string)
	description := d.Get("description").(string)
	kind := d.Get("kind").(string)

	flinkService.UpdateVariable(tea.String(workspace), tea.String(namespace), tea.String(varName), &ververica.Variable{
		Name:        tea.String(varName),
		Value:       tea.String(value),
		Description: tea.String(description),
		Kind:        tea.String(kind),
	})

	return nil
}

func resourceAliCloudFlinkVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return err
	}
	workspace, namespace, varName := splitVariableID(d.Id())
	flinkService.DeleteVariable(tea.String(workspace), tea.String(namespace), tea.String(varName))
	if err != nil {
		return err
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
