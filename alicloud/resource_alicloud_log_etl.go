package alicloud

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudLogETL() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogETLCreate,
		Read:   resourceAliCloudLogETLRead,
		Update: resourceAliCloudLogETLUpdate,
		Delete: resourceAliCloudLogETLDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"etl_config": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"STARTING", "RUNNING", "STOPPING", "STOPPED"}, false),
						},
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"schedule": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "Resident",
									},
									"interval": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"configuration": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  aliyunSlsAPI.ETLVersion,
									},
									"script": {
										Type:     schema.TypeString,
										Required: true,
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{Type: schema.TypeString},
									},
									"from_time": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"to_time": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"logstore": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"sinks": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"type": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  aliyunSlsAPI.ETLSinksType,
												},
												"project": {
													Type:     schema.TypeString,
													Required: true,
												},
												"logstore": {
													Type:     schema.TypeString,
													Required: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"description": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudLogETLCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	etlJob, err := createETLJob(d, meta)
	if err != nil {
		return err
	}

	project := d.Get("project_name").(string)
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := slsService.CreateSlsETL(project, &etlJob)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if debugOn() {
			logstore := ""
			if etlJob.Configuration != nil {
				logstore = etlJob.Configuration.Logstore
			}
			addDebug("CreateETL", nil, nil, map[string]interface{}{
				"project_name": project,
				"logstore":     logstore,
			})
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_etl", "CreateETL", AliyunLogGoSdkERROR)
	}
	d.SetId(fmt.Sprintf("%s%s%s", project, COLON_SEPARATED, etlJob.Name))
	targetStatus := normalizeTargetStatus(etlJob.Status)
	stateConf := BuildStateConf([]string{}, []string{targetStatus}, d.Timeout(schema.TimeoutCreate), 5*time.Second, logService.SlsETLStateRefreshFunc(project, etlJob.Name, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return resourceAliCloudLogETLRead(d, meta)
}

func resourceAliCloudLogETLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	etl, err := logService.DescribeSlsETL(parts[0], parts[1])
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_log_etl SlsService.DescribeLogEtl Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	d.Set("project_name", parts[0])
	if err := d.Set("etl_config", []interface{}{flattenETLToTerraformMap(etl)}); err != nil {
		return WrapError(err)
	}
	return nil
}

func resourceAliCloudLogETLUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		etl, err := createETLJob(d, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		desiredStatus := normalizeTargetStatus(etl.Status)
		if desiredStatus == "STOPPED" {
			err = slsService.UpdateSlsETL(parts[0], parts[1], &etl)
		} else {
			err = slsService.StopSlsETL(parts[0], parts[1])
			if err == nil {
				err = slsService.UpdateSlsETL(parts[0], parts[1], &etl)
				if err == nil {
					err = slsService.StartSlsETL(parts[0], parts[1])
					if err == nil {
						stateConf := BuildStateConf([]string{}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, logService.SlsETLStateRefreshFunc(parts[0], parts[1], []string{}))
						if _, err := stateConf.WaitForState(); err != nil {
							return resource.NonRetryableError(WrapErrorf(err, IdMsg, d.Id()))
						}
					}
				}
			}
		}
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateLogETL", AliyunLogGoSdkERROR)
	}
	return resourceAliCloudLogETLRead(d, meta)
}

func resourceAliCloudLogETLDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := slsService.DeleteSlsETL(parts[0], parts[1])
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if debugOn() {
			addDebug("DeleteLogETL", nil, nil, map[string]interface{}{
				"project_name": parts[0],
				"etl_name":     parts[1],
			})
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_etl", "DeleteLogETL", AliyunLogGoSdkERROR)
	}
	return WrapError(logService.WaitForSlsETL(d.Id(), Deleted, DefaultTimeout))
}

func createETLJob(d *schema.ResourceData, meta interface{}) (aliyunSlsAPI.ETL, error) {
	_ = meta.(*connectivity.AliyunClient)
	etlConfigRaw, ok := d.GetOk("etl_config")
	if !ok || len(etlConfigRaw.([]interface{})) == 0 || etlConfigRaw.([]interface{})[0] == nil {
		return aliyunSlsAPI.ETL{}, WrapError(Error("etl_config is required"))
	}
	etlConfig := etlConfigRaw.([]interface{})[0].(map[string]interface{})
	configurationRaw := etlConfig["configuration"].([]interface{})
	if len(configurationRaw) == 0 || configurationRaw[0] == nil {
		return aliyunSlsAPI.ETL{}, WrapError(Error("etl_config.configuration is required"))
	}
	configurationMap := configurationRaw[0].(map[string]interface{})

	parms := map[string]string{}
	if temp, ok := configurationMap["parameters"].(map[string]interface{}); ok {
		for k, v := range temp {
			parms[k] = v.(string)
		}
	}
	parameterKeys := make([]string, 0, len(parms))
	for key := range parms {
		parameterKeys = append(parameterKeys, key)
	}
	sort.Strings(parameterKeys)
	parameterValues := make([]string, 0, len(parameterKeys))
	for _, key := range parameterKeys {
		parameterValues = append(parameterValues, fmt.Sprintf("%s=%s", key, parms[key]))
	}

	etlSinks := make([]aliyunSlsAPI.ETLSink, 0)
	for _, sinkRaw := range configurationMap["sinks"].(*schema.Set).List() {
		sinkMap := sinkRaw.(map[string]interface{})
		sink := aliyunSlsAPI.ETLSink{
			Name:        sinkMap["name"].(string),
			Type:        sinkMap["type"].(string),
			Project:     sinkMap["project"].(string),
			Logstore:    sinkMap["logstore"].(string),
			RoleArn:     sinkMap["role_arn"].(string),
			Description: sinkMap["description"].(string),
		}
		etlSinks = append(etlSinks, sink)
	}

	configuration := aliyunSlsAPI.ETLConfiguration{
		FromTime:   int64(configurationMap["from_time"].(int)),
		Logstore:   configurationMap["logstore"].(string),
		Parameters: parameterValues,
		Script:     configurationMap["script"].(string),
		ToTime:     int64(configurationMap["to_time"].(int)),
		Version:    configurationMap["version"].(string),
		RoleArn:    configurationMap["role_arn"].(string),
		Sinks:      etlSinks,
	}

	schedule := aliyunSlsAPI.ETLSchedule{Type: "Resident"}
	if scheduleRaw, ok := etlConfig["schedule"].([]interface{}); ok && len(scheduleRaw) > 0 && scheduleRaw[0] != nil {
		scheduleMap := scheduleRaw[0].(map[string]interface{})
		if v, ok := scheduleMap["type"].(string); ok && v != "" {
			schedule.Type = v
		}
		if v, ok := scheduleMap["interval"].(string); ok {
			schedule.Interval = v
		}
	}

	etlJob := aliyunSlsAPI.ETL{
		Configuration: &configuration,
		DisplayName:   etlConfig["display_name"].(string),
		Description:   etlConfig["description"].(string),
		Name:          etlConfig["name"].(string),
		Schedule:      &schedule,
		Status:        etlConfig["status"].(string),
		CreateTime:    int64(etlConfig["create_time"].(int)),
	}
	if etlJob.Status == "" {
		etlJob.Status = "RUNNING"
	}
	return etlJob, nil
}

func flattenETLToTerraformMap(etl *aliyunSlsAPI.ETL) map[string]interface{} {
	result := map[string]interface{}{
		"name":        etl.Name,
		"display_name": etl.DisplayName,
		"description": etl.Description,
		"status":      etl.Status,
		"create_time": int(etl.CreateTime),
	}

	schedule := map[string]interface{}{"type": "Resident", "interval": ""}
	if etl.Schedule != nil {
		schedule["type"] = etl.Schedule.Type
		schedule["interval"] = etl.Schedule.Interval
	}
	result["schedule"] = []interface{}{schedule}

	configuration := map[string]interface{}{}
	if etl.Configuration != nil {
		configuration["version"] = etl.Configuration.Version
		if configuration["version"].(string) == "" {
			configuration["version"] = aliyunSlsAPI.ETLVersion
		}
		configuration["script"] = etl.Configuration.Script
		configuration["parameters"] = convertETLParametersSliceToTerraformMap(etl.Configuration.Parameters)
		configuration["from_time"] = int(etl.Configuration.FromTime)
		configuration["to_time"] = int(etl.Configuration.ToTime)
		configuration["logstore"] = etl.Configuration.Logstore
		configuration["role_arn"] = etl.Configuration.RoleArn
		sinks := make([]map[string]interface{}, 0)
		for _, sink := range etl.Configuration.Sinks {
			sinks = append(sinks, map[string]interface{}{
				"name":        sink.Name,
				"type":        sink.Type,
				"project":     sink.Project,
				"logstore":    sink.Logstore,
				"role_arn":    sink.RoleArn,
				"description": sink.Description,
			})
		}
		configuration["sinks"] = sinks
	} else {
		configuration["version"] = aliyunSlsAPI.ETLVersion
		configuration["script"] = ""
		configuration["parameters"] = map[string]string{}
		configuration["from_time"] = 0
		configuration["to_time"] = 0
		configuration["logstore"] = ""
		configuration["role_arn"] = ""
		configuration["sinks"] = []map[string]interface{}{}
	}
	result["configuration"] = []interface{}{configuration}
	return result
}

func normalizeTargetStatus(status string) string {
	switch strings.ToUpper(status) {
	case "STOPPED", "STOPPING":
		return "STOPPED"
	default:
		return "RUNNING"
	}
}

func convertETLParametersSliceToTerraformMap(parameters []string) map[string]string {
	result := make(map[string]string)
	for idx, parameter := range parameters {
		if parameter == "" {
			continue
		}
		parts := strings.SplitN(parameter, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
			continue
		}
		result["param_"+strconv.Itoa(idx)] = parameter
	}
	return result
}
