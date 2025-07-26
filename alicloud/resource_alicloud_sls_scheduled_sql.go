// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSlsScheduledSQL() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSlsScheduledSQLCreate,
		Read:   resourceAliCloudSlsScheduledSQLRead,
		Update: resourceAliCloudSlsScheduledSQLUpdate,
		Delete: resourceAliCloudSlsScheduledSQLDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schedule": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"run_immediately": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"cron_expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"delay": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"interval": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"scheduled_sql_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"script": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"resource_pool": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"from_time_expr": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dest_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"to_time": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"max_run_time_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"data_format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sql_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_logstore": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"dest_logstore": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dest_endpoint": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"to_time_expr": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"from_time": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"dest_project": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"scheduled_sql_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudSlsScheduledSQLCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	projectName := d.Get("project").(string)
	scheduledSQLName := d.Get("scheduled_sql_name").(string)

	// Create ScheduledSQL struct from Terraform data
	scheduledSQL := &aliyunSlsAPI.ScheduledSQL{
		Name:        scheduledSQLName,
		DisplayName: d.Get("display_name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		scheduledSQL.Description = v.(string)
	}

	// Set schedule configuration
	if scheduleList := d.Get("schedule").([]interface{}); len(scheduleList) > 0 {
		scheduleData := scheduleList[0].(map[string]interface{})
		scheduledSQL.Schedule = &aliyunSlsAPI.ScheduledSQLSchedule{}

		if v, ok := scheduleData["type"]; ok && v.(string) != "" {
			scheduledSQL.Schedule.Type = v.(string)
		}
		if v, ok := scheduleData["cron_expression"]; ok && v.(string) != "" {
			scheduledSQL.Schedule.CronExpression = v.(string)
		}
		if v, ok := scheduleData["time_zone"]; ok && v.(string) != "" {
			scheduledSQL.Schedule.TimeZone = v.(string)
		}
		if v, ok := scheduleData["interval"]; ok && v.(string) != "" {
			scheduledSQL.Schedule.Interval = v.(string)
		}
		if v, ok := scheduleData["delay"]; ok {
			scheduledSQL.Schedule.Delay = int32(v.(int))
		}
		if v, ok := scheduleData["run_immediately"]; ok {
			scheduledSQL.Schedule.RunImmediately = v.(bool)
		}
	}

	// Set SQL configuration
	if configList := d.Get("scheduled_sql_configuration").([]interface{}); len(configList) > 0 {
		configData := configList[0].(map[string]interface{})
		scheduledSQL.Configuration = &aliyunSlsAPI.ScheduledSQLConfiguration{}

		if v, ok := configData["script"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.Script = v.(string)
		}
		if v, ok := configData["sql_type"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.SqlType = v.(string)
		}
		if v, ok := configData["source_logstore"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.SourceLogstore = v.(string)
		}
		if v, ok := configData["dest_project"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.DestProject = v.(string)
		}
		if v, ok := configData["dest_logstore"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.DestLogstore = v.(string)
		}
		if v, ok := configData["dest_endpoint"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.DestEndpoint = v.(string)
		}
		if v, ok := configData["role_arn"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.RoleArn = v.(string)
		}
		if v, ok := configData["dest_role_arn"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.DestRoleArn = v.(string)
		}
		if v, ok := configData["from_time_expr"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.FromTimeExpr = v.(string)
		}
		if v, ok := configData["to_time_expr"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.ToTimeExpr = v.(string)
		}
		if v, ok := configData["data_format"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.DataFormat = v.(string)
		}
		if v, ok := configData["resource_pool"]; ok && v.(string) != "" {
			scheduledSQL.Configuration.ResourcePool = v.(string)
		}
		if v, ok := configData["max_retries"]; ok {
			scheduledSQL.Configuration.MaxRetries = int64(v.(int))
		}
		if v, ok := configData["max_run_time_in_seconds"]; ok {
			scheduledSQL.Configuration.MaxRunTimeInSeconds = int64(v.(int))
		}
		if v, ok := configData["from_time"]; ok {
			scheduledSQL.Configuration.FromTime = int64(v.(int))
		}
		if v, ok := configData["to_time"]; ok {
			scheduledSQL.Configuration.ToTime = int64(v.(int))
		}
		if v, ok := configData["parameters"]; ok {
			if params, ok := v.(map[string]interface{}); ok {
				scheduledSQL.Configuration.Parameters = params
			}
		}
	}

	// Create the scheduled SQL
	err = slsService.CreateSlsScheduledSQL(projectName, scheduledSQL)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "CreateSlsScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s", projectName, scheduledSQLName))

	// Wait for the resource to be available using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"ENABLED", "DISABLED"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsScheduledSQLStateRefreshFunc(d.Id(), "$.status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSlsScheduledSQLRead(d, meta)
}

func resourceAliCloudSlsScheduledSQLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	scheduledSQL, err := slsService.DescribeSlsScheduledSQL(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_sls_scheduled_sql DescribeSlsScheduledSQL Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if scheduledSQL == nil {
		d.SetId("")
		return nil
	}

	// Set basic attributes
	d.Set("description", scheduledSQL.Description)
	d.Set("display_name", scheduledSQL.DisplayName)

	// Set schedule configuration
	if scheduledSQL.Schedule != nil {
		scheduleMap := map[string]interface{}{
			"type":            scheduledSQL.Schedule.Type,
			"cron_expression": scheduledSQL.Schedule.CronExpression,
			"time_zone":       scheduledSQL.Schedule.TimeZone,
			"interval":        scheduledSQL.Schedule.Interval,
			"delay":           scheduledSQL.Schedule.Delay,
			"run_immediately": scheduledSQL.Schedule.RunImmediately,
		}
		d.Set("schedule", []map[string]interface{}{scheduleMap})
	}

	// Set SQL configuration
	if scheduledSQL.Configuration != nil {
		config := scheduledSQL.Configuration
		configMap := map[string]interface{}{
			"script":                  config.Script,
			"sql_type":                config.SqlType,
			"source_logstore":         config.SourceLogstore,
			"dest_project":            config.DestProject,
			"dest_logstore":           config.DestLogstore,
			"dest_endpoint":           config.DestEndpoint,
			"role_arn":                config.RoleArn,
			"dest_role_arn":           config.DestRoleArn,
			"from_time_expr":          config.FromTimeExpr,
			"to_time_expr":            config.ToTimeExpr,
			"data_format":             config.DataFormat,
			"resource_pool":           config.ResourcePool,
			"max_retries":             config.MaxRetries,
			"max_run_time_in_seconds": config.MaxRunTimeInSeconds,
			"from_time":               config.FromTime,
			"to_time":                 config.ToTime,
			"parameters":              config.Parameters,
		}
		d.Set("scheduled_sql_configuration", []map[string]interface{}{configMap})
	}

	// Set project and scheduled SQL name from ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) == 2 {
		d.Set("project", parts[0])
		d.Set("scheduled_sql_name", parts[1])
	}

	return nil
}

func resourceAliCloudSlsScheduledSQLUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapErrorf(Error("invalid scheduled SQL ID format"), DefaultErrorMsg, "alicloud_sls_scheduled_sql", "Update", AlibabaCloudSdkGoERROR)
	}

	projectName := parts[0]
	scheduledSQLName := parts[1]
	update := false

	// Create ScheduledSQL struct for update
	scheduledSQL := &aliyunSlsAPI.ScheduledSQL{
		Name:        scheduledSQLName,
		DisplayName: d.Get("display_name").(string),
	}

	if d.HasChange("display_name") {
		update = true
	}

	if d.HasChange("description") {
		update = true
		scheduledSQL.Description = d.Get("description").(string)
	}

	// Set schedule configuration
	if d.HasChange("schedule") {
		update = true
		if scheduleList := d.Get("schedule").([]interface{}); len(scheduleList) > 0 {
			scheduleData := scheduleList[0].(map[string]interface{})
			scheduledSQL.Schedule = &aliyunSlsAPI.ScheduledSQLSchedule{}

			if v, ok := scheduleData["type"]; ok && v.(string) != "" {
				scheduledSQL.Schedule.Type = v.(string)
			}
			if v, ok := scheduleData["cron_expression"]; ok && v.(string) != "" {
				scheduledSQL.Schedule.CronExpression = v.(string)
			}
			if v, ok := scheduleData["time_zone"]; ok && v.(string) != "" {
				scheduledSQL.Schedule.TimeZone = v.(string)
			}
			if v, ok := scheduleData["interval"]; ok && v.(string) != "" {
				scheduledSQL.Schedule.Interval = v.(string)
			}
			if v, ok := scheduleData["delay"]; ok {
				scheduledSQL.Schedule.Delay = int32(v.(int))
			}
			if v, ok := scheduleData["run_immediately"]; ok {
				scheduledSQL.Schedule.RunImmediately = v.(bool)
			}
		}
	}

	// Set SQL configuration
	if d.HasChange("scheduled_sql_configuration") {
		update = true
		if configList := d.Get("scheduled_sql_configuration").([]interface{}); len(configList) > 0 {
			configData := configList[0].(map[string]interface{})
			scheduledSQL.Configuration = &aliyunSlsAPI.ScheduledSQLConfiguration{}

			if v, ok := configData["script"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.Script = v.(string)
			}
			if v, ok := configData["sql_type"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.SqlType = v.(string)
			}
			if v, ok := configData["source_logstore"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.SourceLogstore = v.(string)
			}
			if v, ok := configData["dest_project"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.DestProject = v.(string)
			}
			if v, ok := configData["dest_logstore"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.DestLogstore = v.(string)
			}
			if v, ok := configData["dest_endpoint"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.DestEndpoint = v.(string)
			}
			if v, ok := configData["role_arn"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.RoleArn = v.(string)
			}
			if v, ok := configData["dest_role_arn"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.DestRoleArn = v.(string)
			}
			if v, ok := configData["from_time_expr"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.FromTimeExpr = v.(string)
			}
			if v, ok := configData["to_time_expr"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.ToTimeExpr = v.(string)
			}
			if v, ok := configData["data_format"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.DataFormat = v.(string)
			}
			if v, ok := configData["resource_pool"]; ok && v.(string) != "" {
				scheduledSQL.Configuration.ResourcePool = v.(string)
			}
			if v, ok := configData["max_retries"]; ok {
				scheduledSQL.Configuration.MaxRetries = int64(v.(int))
			}
			if v, ok := configData["max_run_time_in_seconds"]; ok {
				scheduledSQL.Configuration.MaxRunTimeInSeconds = int64(v.(int))
			}
			if v, ok := configData["from_time"]; ok {
				scheduledSQL.Configuration.FromTime = int64(v.(int))
			}
			if v, ok := configData["to_time"]; ok {
				scheduledSQL.Configuration.ToTime = int64(v.(int))
			}
			if v, ok := configData["parameters"]; ok {
				if params, ok := v.(map[string]interface{}); ok {
					scheduledSQL.Configuration.Parameters = params
				}
			}
		}
	}

	// Update the scheduled SQL if needed
	if update {
		err = slsService.UpdateSlsScheduledSQL(projectName, scheduledSQLName, scheduledSQL)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSlsScheduledSQL", AlibabaCloudSdkGoERROR)
		}
	}

	// Handle status changes (enable/disable)
	if d.HasChange("status") {
		target := d.Get("status").(string)

		// Wait for update to complete first
		stateConf := BuildStateConf([]string{}, []string{"ENABLED", "DISABLED"}, d.Timeout(schema.TimeoutUpdate), 1*time.Second, slsService.SlsScheduledSQLStateRefreshFunc(d.Id(), "$.status", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		// Change status if needed
		if target == "ENABLED" {
			err = slsService.EnableSlsScheduledSQL(projectName, scheduledSQLName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "EnableSlsScheduledSQL", AlibabaCloudSdkGoERROR)
			}
		} else if target == "DISABLED" {
			err = slsService.DisableSlsScheduledSQL(projectName, scheduledSQLName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DisableSlsScheduledSQL", AlibabaCloudSdkGoERROR)
			}
		}
	}

	return resourceAliCloudSlsScheduledSQLRead(d, meta)
}

func resourceAliCloudSlsScheduledSQLDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_scheduled_sql", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapErrorf(Error("invalid scheduled SQL ID format"), DefaultErrorMsg, "alicloud_sls_scheduled_sql", "Delete", AlibabaCloudSdkGoERROR)
	}

	projectName := parts[0]
	scheduledSQLName := parts[1]

	// Delete the scheduled SQL
	err = slsService.DeleteSlsScheduledSQL(projectName, scheduledSQLName)
	if err != nil {
		if IsExpectedErrors(err, []string{"403", "ResourceNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSlsScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	return nil
}
