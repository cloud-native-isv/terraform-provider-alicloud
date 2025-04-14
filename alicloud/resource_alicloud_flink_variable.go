package flink

import (
	"context"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAliCloudFlinkVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAliCloudFlinkVariableCreate,
		ReadContext:   resourceAliCloudFlinkVariableRead,
		UpdateContext: resourceAliCloudFlinkVariableUpdate,
		DeleteContext: resourceAliCloudFlinkVariableDelete,
		Schema: map[string]*schema.Schema{
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
			"flink_instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink instance",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the variable",
			},
		},
	}
}

func resourceAliCloudFlinkVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.AliyunClient)
	request := map[string]interface{}{
		"name":             d.Get("name"),
		"value":            d.Get("value"),
		"flink_instance_id": d.Get("flink_instance_id"),
		"description":      d.Get("description"),
	}

	response, err := client.FlinkService.CreateVariable(request)
	if err != nil {
		return diag.Errorf("failed to create Flink variable: %s", err)
	}

	variableId := response["VariableId"].(string)
	d.SetId(strings.Join([]string{request["flink_instance_id"].(string), variableId}, "/"))

	return resourceAliCloudFlinkVariableRead(ctx, d, meta)
}

func resourceAliCloudFlinkVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.AliyunClient)
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format, expected <flink_instance_id>/<variable_id>")
	}
	namespace, variableId := parts[0], parts[1]

	response, err := client.FlinkService.DescribeVariable(variableId)
	if err != nil {
		if !connectivity.IsResourceNotFoundException(err) {
			return diag.Errorf("failed to describe Flink variable: %s", err)
		}
		d.SetId("")
		return nil
	}

	d.Set("name", response["Name"])
	d.Set("value", response["Value"])
	d.Set("flink_instance_id", response["FlinkInstanceId"])
	d.Set("description", response["Description"])

	return nil
}

func resourceAliCloudFlinkVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.AliyunClient)
	request := map[string]interface{}{
		"id":        d.Id(),
		"value":     d.Get("value"),
		"description": d.Get("description"),
	}

	_, err := client.FlinkService.UpdateVariable(request)
	if err != nil {
		return diag.Errorf("failed to update Flink variable: %s", err)
	}

	return resourceAliCloudFlinkVariableRead(ctx, d, meta)
}

func resourceAliCloudFlinkVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*connectivity.AliyunClient)
	_, variableId := splitVariableID(d.Id())

	err := client.FlinkService.DeleteVariable(variableId)
	if err != nil {
		if connectivity.IsResourceNotFoundException(err) {
			return nil
		}
		return diag.Errorf("failed to delete Flink variable: %s", err)
	}

	return nil
}

func splitVariableID(id string) (string, string) {
	parts := strings.Split(id, "/")
	return parts[0], parts[1]
}