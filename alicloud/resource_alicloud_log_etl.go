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
			"etl_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project": {
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
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Resident",
			},

			"etl_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  aliyunSlsAPI.ETLType,
			},

			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"STARTING", "RUNNING", "STOPPING", "STOPPED"}, false),
			},

			"create_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"last_modified_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"access_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"access_key_secret": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_encrypted_access_key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: kmsDiffSuppressFunc,
			},

			"kms_encryption_access_key_id_context": {
				Type:     schema.TypeMap,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("kms_encrypted_access_key_id").(string) == ""
				},
				Elem: schema.TypeString,
			},

			"kms_encrypted_access_key_secret": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: kmsDiffSuppressFunc,
			},

			"kms_encryption_access_key_secret_context": {
				Type:     schema.TypeMap,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("kms_encrypted_access_key_secret").(string) == ""
				},
				Elem: schema.TypeString,
			},

			"from_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"to_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"script": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  aliyunSlsAPI.ETLVersion,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etl_sinks": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_encrypted_access_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: kmsDiffSuppressFunc,
						},
						"access_key_secret": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_encrypted_access_key_secret": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: kmsDiffSuppressFunc,
						},
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
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
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  aliyunSlsAPI.ETLSinksType,
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
	etlJob, err := getETLJob(d, meta)
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
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
			addDebug("CreateETL", nil, nil, map[string]interface{}{
				"project":  project,
				"logstore": d.Get("logstore").(string),
			})
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_etl", "CreateETL", AliyunLogGoSdkERROR)
	}
	d.SetId(fmt.Sprintf("%s%s%s", project, COLON_SEPARATED, d.Get("etl_name").(string)))
	stateConf := BuildStateConf([]string{}, []string{"RUNNING"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, logService.SlsETLStateRefreshFunc(project, d.Get("etl_name").(string), []string{}))
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
	d.Set("etl_name", parts[1])
	d.Set("project", parts[0])

	// Handle struct response from new service
	d.Set("display_name", etl.DisplayName)
	d.Set("description", etl.Description)
	d.Set("status", etl.Status)

	if etl.CreateTime > 0 {
		d.Set("create_time", int(etl.CreateTime))
	}

	if etl.CreateTime > 0 {
		d.Set("last_modified_time", int(etl.CreateTime))
	}

	if etl.Configuration != nil {
		d.Set("from_time", int(etl.Configuration.FromTime))
		d.Set("to_time", int(etl.Configuration.ToTime))
		d.Set("script", etl.Configuration.Script)
		if etl.Configuration.Version != "" {
			d.Set("version", etl.Configuration.Version)
		} else {
			d.Set("version", aliyunSlsAPI.ETLVersion)
		}

		d.Set("logstore", etl.Configuration.Logstore)
		d.Set("parameters", convertETLParametersSliceToTerraformMap(etl.Configuration.Parameters))

		d.Set("role_arn", etl.Configuration.RoleArn)
		if etl.Configuration.Sinks != nil {
			var etl_sinks []map[string]interface{}
			for _, sink := range etl.Configuration.Sinks {
				endpoint := ""
				if currentSinks, ok := d.GetOk("etl_sinks"); ok {
					for _, currentSinkRaw := range currentSinks.(*schema.Set).List() {
						currentSink := currentSinkRaw.(map[string]interface{})
						if currentSink["name"].(string) == sink.Name {
							if endpointValue, endpointOk := currentSink["endpoint"].(string); endpointOk {
								endpoint = endpointValue
							}
							break
						}
					}
				}
				temp := map[string]interface{}{
					"name":     sink.Name,
					"endpoint": endpoint,
					"project":  sink.Project,
					"logstore": sink.Logstore,
					"role_arn": sink.RoleArn,
					"type":     sink.Type,
				}
				etl_sinks = append(etl_sinks, temp)
			}
			d.Set("etl_sinks", etl_sinks)
		}
	}

	if etl.Schedule != nil {
		d.Set("schedule", etl.Schedule.Type)
	}

	d.Set("etl_type", aliyunSlsAPI.ETLType)

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
	if d.HasChange("status") {
		status := d.Get("status").(string)
		if status == "STARTING" || status == "RUNNING" {
			if err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				err := slsService.StartSlsETL(parts[0], d.Get("etl_name").(string))
				if err != nil {
					if IsExpectedErrors(err, []string{LogClientTimeout}) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			}); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StartLogETL", AliyunLogGoSdkERROR)
			}
			stateConf := BuildStateConf([]string{}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, logService.SlsETLStateRefreshFunc(parts[0], parts[1], []string{}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		} else if status == "STOPPING" || status == "STOPPED" {
			if err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				err := slsService.StopSlsETL(parts[0], d.Get("etl_name").(string))
				if err != nil {
					if IsExpectedErrors(err, []string{LogClientTimeout}) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			}); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StopLogETL", AliyunLogGoSdkERROR)
			}
			stateConf := BuildStateConf([]string{}, []string{"STOPPED"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, logService.SlsETLStateRefreshFunc(parts[0], parts[1], []string{}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}

		}
		return resourceAliCloudLogETLRead(d, meta)
	}

	if err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		etl, err := getETLJob(d, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		status := d.Get("status").(string)
		if status == "STOPPING" || status == "STOPPED" {
			err = slsService.UpdateSlsETL(parts[0], parts[1], &etl)
		} else {
			err = slsService.StopSlsETL(parts[0], parts[1])
			if err == nil {
				err = slsService.UpdateSlsETL(parts[0], parts[1], &etl)
				if err == nil {
					err = slsService.StartSlsETL(parts[0], parts[1])
					if err == nil {
						logService, err := NewSlsService(client)
						if err != nil {
							return resource.NonRetryableError(WrapError(err))
						}
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
				"elt_name":     parts[1],
			})
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_etl", "DeleteLogETL", AliyunLogGoSdkERROR)
	}
	return WrapError(logService.WaitForSlsETL(d.Id(), Deleted, DefaultTimeout))
}

func getETLJob(d *schema.ResourceData, meta interface{}) (aliyunSlsAPI.ETL, error) {
	client := meta.(*connectivity.AliyunClient)
	var etlSinks = []aliyunSlsAPI.ETLSink{}
	var config = aliyunSlsAPI.ETLConfiguration{}
	var etlJob = aliyunSlsAPI.ETL{}
	schedule := aliyunSlsAPI.ETLSchedule{
		Type: d.Get("schedule").(string),
	}
	parms := map[string]string{}
	if temp, ok := d.GetOk("parameters"); ok {
		for k, v := range temp.(map[string]interface{}) {
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
	config = aliyunSlsAPI.ETLConfiguration{
		FromTime:   int64(d.Get("from_time").(int)),
		Logstore:   d.Get("logstore").(string),
		Parameters: parameterValues,
		Script:     d.Get("script").(string),
		ToTime:     int64(d.Get("to_time").(int)),
		Version:    d.Get("version").(string),
	}
	for _, f := range d.Get("etl_sinks").(*schema.Set).List() {
		v := f.(map[string]interface{})
		sink := aliyunSlsAPI.ETLSink{
			Name:     v["name"].(string),
			Type:     v["type"].(string),
			Project:  v["project"].(string),
			Logstore: v["logstore"].(string),
		}
		sinkResult, err := permissionParameterCheck(v, client, d)
		if err != nil {
			return etlJob, WrapError(err)
		}
		if len(sinkResult) == 1 {
			sink.RoleArn = sinkResult["roleArn"]
		} else {
			// Note: These fields may need to be added to ETLSink type if they don't exist
			// sink.AccessKeyId = sinkResult["akId"]
			// sink.AccessKeySecret = sinkResult["ak"]
		}
		etlSinks = append(etlSinks, sink)
	}
	config.Sinks = etlSinks

	configResult, err := permissionParameterCheck(nil, client, d)
	if err != nil {
		return etlJob, WrapError(err)
	}
	if len(configResult) == 1 {
		config.RoleArn = configResult["roleArn"]
	} else {
		// Note: These fields may need to be added to ETLConfiguration type if they don't exist
		// config.AccessKeyId = configResult["akId"]
		// config.AccessKeySecret = configResult["ak"]
	}
	etlJob = aliyunSlsAPI.ETL{
		Configuration: &config,
		DisplayName:   d.Get("display_name").(string),
		Description:   d.Get("description").(string),
		Name:          d.Get("etl_name").(string),
		Schedule:      &schedule,
		Status:        d.Get("status").(string),
		CreateTime:    int64(d.Get("create_time").(int)),
	}
	return etlJob, nil
}

func permissionParameterCheck(v map[string]interface{}, client *connectivity.AliyunClient, d *schema.ResourceData) (map[string]string, error) {
	if v != nil {
		akId := v["access_key_id"].(string)
		ak := v["access_key_secret"].(string)
		roleArn := v["role_arn"].(string)
		if akId != "" && ak != "" {
			if roleArn != "" {
				return nil, Error("(access_key_id, access_key_secret), (role_arn) only one can be selected to fill into the sink")
			}
			return map[string]string{"akId": akId, "ak": ak}, nil
		}
		if roleArn != "" {
			return map[string]string{"roleArn": roleArn}, nil
		}
		kmsAkId := v["kms_encrypted_access_key_id"].(string)
		kmsAk := v["kms_encrypted_access_key_secret"].(string)
		if kmsAkId != "" && kmsAk != "" {
			kmsService := KmsService{client}
			akIdResp, akIdErr := kmsService.Decrypt(kmsAkId, d.Get("kms_encryption_access_key_id_context").(map[string]interface{}))
			if akIdErr != nil {
				return nil, akIdErr
			}
			akResp, akErr := kmsService.Decrypt(kmsAk, d.Get("kms_encryption_access_key_secret_context").(map[string]interface{}))
			if akErr != nil {
				return nil, akErr
			}
			return map[string]string{"akId": akIdResp, "ak": akResp}, nil
		}
		return nil, Error("(access_key_id, access_key_secret),(kms_encrypted_access_key_id, kms_encrypted_access_key_secret, kms_encryption_access_key_id_context, kms_encryption_access_key_secret_context),(role_arn) must fill in one of them into sink")
	} else {
		akId := d.Get("access_key_id").(string)
		ak := d.Get("access_key_secret").(string)
		roleArn := d.Get("role_arn").(string)
		if akId != "" && ak != "" {
			if roleArn != "" {
				return nil, Error("(access_key_id, access_key_secret), (role_arn) only one can be selected to fill into the configuration")
			}
			return map[string]string{"akId": akId, "ak": ak}, nil
		}
		if roleArn != "" {
			return map[string]string{"roleArn": roleArn}, nil
		}
		kmsAkId := d.Get("kms_encrypted_access_key_id").(string)
		kmsAk := d.Get("kms_encrypted_access_key_secret").(string)
		if kmsAkId != "" && kmsAk != "" {
			kmsService := KmsService{client}
			akIdResp, akIdErr := kmsService.Decrypt(kmsAkId, d.Get("kms_encryption_access_key_id_context").(map[string]interface{}))
			if akIdErr != nil {
				return nil, akIdErr
			}
			akResp, akErr := kmsService.Decrypt(kmsAk, d.Get("kms_encryption_access_key_secret_context").(map[string]interface{}))
			if akErr != nil {
				return nil, akErr
			}
			return map[string]string{"akId": akIdResp, "ak": akResp}, nil
		}
		return nil, Error("(access_key_id, access_key_secret),(kms_encrypted_access_key_id, kms_encrypted_access_key_secret, kms_encryption_access_key_id_context, kms_encryption_access_key_secret_context),(role_arn) must fill in one of them into configuration")
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
