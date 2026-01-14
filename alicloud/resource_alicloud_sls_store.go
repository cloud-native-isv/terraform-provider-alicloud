// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudLogStore() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSlsLogStoreCreate,
		Read:   resourceAliCloudSlsLogStoreRead,
		Update: resourceAliCloudSlsLogStoreUpdate,
		Delete: resourceAliCloudSlsLogStoreDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"append_meta": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"auto_split": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enable_web_tracking": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"encrypt_conf": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypt_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"enable": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"user_cmk_info": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cmk_key_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"region_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"arn": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"hot_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"infrequent_access_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  365,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3650,
			},
			"logstore_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"logstore_name", "name"},
				ForceNew:     true,
			},
			"max_split_shard_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: IntBetween(0, 256),
			},
			"metering_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					}
					return old != "" && new != "" && old == new
				},
			},
			"project_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"project_name", "project"},
				ForceNew:     true,
			},
			"shard_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
				ForceNew: true,
			},
			"telemetry_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"project": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				Deprecated: "Field 'project' has been deprecated since provider version 1.215.0. New field 'project_name' instead.",
				ForceNew:   true,
			},
			"name": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				Deprecated: "Field 'name' has been deprecated since provider version 1.215.0. New field 'logstore_name' instead.",
				ForceNew:   true,
			},
			"shards": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"begin_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudSlsLogStoreCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	projectName := d.Get("project_name").(string)
	if v, ok := d.GetOkExists("project"); ok {
		projectName = v.(string)
	}

	logstore := buildLogStore(d)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		createErr := slsService.CreateLogStore(projectName, logstore)
		if createErr != nil {
			// Handle LogStoreAlreadyExist error by importing existing resource (preserve prior behavior)
			if IsExpectedErrors(createErr, []string{"LogStoreAlreadyExist"}) {
				log.Printf("[INFO] LogStore %s already exists, importing existing resource", logstore.LogstoreName)
				d.SetId(fmt.Sprintf("%s:%s", projectName, logstore.LogstoreName))
				return nil
			}
			if NeedRetry(createErr) {
				wait()
				return resource.RetryableError(createErr)
			}
			return resource.NonRetryableError(createErr)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "CreateLogStore", AlibabaCloudSdkGoERROR)
	}

	// Set ID if not already set (for new resources)
	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%s:%s", projectName, logstore.LogstoreName))
	}

	if v, ok := d.GetOk("max_split_shard_count"); ok {
		d.Set("max_split_shard_count", v)
	}

	// Wait for the logstore to be available using state refresh function
	stateConf := BuildStateConf([]string{}, []string{"available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.LogStoreStateRefreshFunc(d.Id(), "logstoreName", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSlsLogStoreUpdate(d, meta)
}

func resourceAliCloudSlsLogStoreRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	logstore, err := slsService.DescribeLogStoreById(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_log_store DescribeLogStore Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set logstore attributes using struct fields instead of map access
	d.Set("append_meta", logstore.AppendMeta)
	d.Set("auto_split", logstore.AutoSplit)
	d.Set("create_time", logstore.CreateTime)
	d.Set("enable_web_tracking", logstore.EnableTracking)
	if logstore.HotTtl != nil {
		d.Set("hot_ttl", *logstore.HotTtl)
	}
	if logstore.InfrequentAccessTTL != nil {
		d.Set("infrequent_access_ttl", *logstore.InfrequentAccessTTL)
	}
	d.Set("max_split_shard_count", logstore.MaxSplitShard)
	d.Set("mode", logstore.Mode)
	d.Set("ttl", logstore.Ttl)
	d.Set("shard_count", logstore.ShardCount)
	d.Set("telemetry_type", logstore.TelemetryType)
	d.Set("logstore_name", logstore.LogstoreName)

	// Handle encryption configuration
	encryptConfMaps := make([]map[string]interface{}, 0)
	if logstore.EncryptConf != nil {
		encryptConfMap := make(map[string]interface{})
		encryptConfMap["enable"] = logstore.EncryptConf.Enable
		encryptConfMap["encrypt_type"] = logstore.EncryptConf.EncryptType

		userCmkInfoMaps := make([]map[string]interface{}, 0)
		if logstore.EncryptConf.UserCmkInfo != nil {
			userCmkInfoMap := make(map[string]interface{})
			userCmkInfoMap["arn"] = logstore.EncryptConf.UserCmkInfo.Arn
			userCmkInfoMap["cmk_key_id"] = logstore.EncryptConf.UserCmkInfo.CmkKeyId
			userCmkInfoMap["region_id"] = logstore.EncryptConf.UserCmkInfo.RegionId
			userCmkInfoMaps = append(userCmkInfoMaps, userCmkInfoMap)
		}
		encryptConfMap["user_cmk_info"] = userCmkInfoMaps
		encryptConfMaps = append(encryptConfMaps, encryptConfMap)

		if err := d.Set("encrypt_conf", encryptConfMaps); err != nil {
			return err
		}
	}

	// Get metering mode information
	meteringMode, err := slsService.DescribeGetLogStoreMeteringMode(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if meteringMode != nil {
		d.Set("metering_mode", meteringMode.MeteringMode)
	}

	parts := strings.Split(d.Id(), ":")
	d.Set("project_name", parts[0])
	d.Set("logstore_name", parts[1])

	d.Set("project", d.Get("project_name"))
	d.Set("name", d.Get("logstore_name"))

	// Get shard information using the SLS service
	// Add retry logic for temporary LogStoreNotExist errors, especially for newly created resources
	var shards []*aliyunSlsAPI.LogStoreShard
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		var getErr error
		shards, getErr = slsService.GetLogStoreShards(parts[0], parts[1])
		if getErr != nil {
			// For new resources, retry on LogStoreNotExist errors as the logstore may not be fully initialized yet
			if d.IsNewResource() && NotFoundError(getErr) {
				log.Printf("[DEBUG] Resource alicloud_log_store GetLogStoreShards returned LogStoreNotExist for new resource, retrying...")
				return resource.RetryableError(getErr)
			}
			// For existing resources or other errors, return non-retryable error
			return resource.NonRetryableError(WrapErrorf(getErr, DefaultErrorMsg, "alicloud_log_store", "GetLogStoreShards", AliyunLogGoSdkERROR))
		}
		return nil
	})
	if err != nil {
		return err
	}

	var shardList []map[string]interface{}
	activeCount := 0
	for _, s := range shards {
		mapping := map[string]interface{}{
			"id":        s.ShardId,
			"status":    s.Status,
			"begin_key": s.InclusiveBeginKey,
			"end_key":   s.ExclusiveEndKey,
		}
		shardList = append(shardList, mapping)
		if strings.ToLower(s.Status) == "readwrite" {
			activeCount++
		}
	}
	d.Set("shards", shardList)
	d.Set("shard_count", activeCount)
	return nil
}

func resourceAliCloudSlsLogStoreUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	update := false
	d.Partial(true)
	parts := strings.Split(d.Id(), ":")
	projectName := parts[0]
	logstoreName := parts[1]

	if !d.IsNewResource() && d.HasChange("auto_split") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("append_meta") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("hot_ttl") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("mode") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("ttl") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("max_split_shard_count") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("enable_web_tracking") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("encrypt_conf") {
		update = true
	}
	if !d.IsNewResource() && d.HasChange("infrequent_access_ttl") {
		update = true
	}

	// Apply updates via strongly-typed API when any field changes
	if update {
		logstoreObj := buildLogStore(d)
		// Ensure name is correct
		logstoreObj.LogstoreName = logstoreName

		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			updateErr := slsService.UpdateLogStore(projectName, logstoreName, logstoreObj)
			if updateErr != nil {
				if NeedRetry(updateErr) {
					wait()
					return resource.RetryableError(updateErr)
				}
				return resource.NonRetryableError(updateErr)
			}
			return nil
		})
		if err != nil {
			action := fmt.Sprintf("/logstores/%s", logstoreName)
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		if v, ok := d.GetOk("max_split_shard_count"); ok {
			d.Set("max_split_shard_count", v)
		}

		stateConf := BuildStateConf([]string{}, []string{"available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.LogStoreStateRefreshFunc(d.Id(), "logstoreName", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	// Metering mode update via strongly-typed API
	update = false
	meteringModeObj, _ := slsService.DescribeGetLogStoreMeteringMode(d.Id())
	if d.HasChange("metering_mode") && meteringModeObj != nil && meteringModeObj.MeteringMode != d.Get("metering_mode").(string) {
		update = true
	}

	if update {
		metering := &aliyunSlsAPI.LogStoreMeteringMode{MeteringMode: d.Get("metering_mode").(string)}
		wait := incrementalWait(3*time.Second, 5*time.Second)
		action := fmt.Sprintf("/logstores/%s/meteringmode", logstoreName)

		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			updateErr := slsService.UpdateLogStoreMeteringMode(projectName, logstoreName, metering)
			if updateErr != nil {
				if NeedRetry(updateErr) {
					wait()
					return resource.RetryableError(updateErr)
				}
				return resource.NonRetryableError(updateErr)
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{}, []string{"available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.LogStoreStateRefreshFunc(d.Id(), "logstoreName", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	d.Partial(false)
	return resourceAliCloudSlsLogStoreRead(d, meta)
}

func resourceAliCloudSlsLogStoreDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	parts := strings.Split(d.Id(), ":")
	logstore := parts[1]
	action := fmt.Sprintf("/logstores/%s", logstore)

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		delErr := slsService.DeleteLogStore(parts[0], logstore)
		if delErr != nil {
			if NeedRetry(delErr) {
				wait()
				return resource.RetryableError(delErr)
			}
			return resource.NonRetryableError(delErr)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

func buildLogStore(d *schema.ResourceData) *aliyunSlsAPI.LogStore {
	logstore := &aliyunSlsAPI.LogStore{
		LogstoreName:   d.Get("logstore_name").(string),
		Ttl:            int32(d.Get("ttl").(int)),
		ShardCount:     int32(d.Get("shard_count").(int)),
		EnableTracking: d.Get("enable_web_tracking").(bool),
		AutoSplit:      d.Get("auto_split").(bool),
		MaxSplitShard:  int32(d.Get("max_split_shard_count").(int)),
		AppendMeta:     d.Get("append_meta").(bool),
		TelemetryType:  d.Get("telemetry_type").(string),
		Mode:           d.Get("mode").(string),
	}
	if v, ok := d.GetOkExists("name"); ok {
		logstore.LogstoreName = v.(string)
	}
	if hotTTL, ok := d.GetOk("hot_ttl"); ok {
		logstore.HotTtl = tea.Int32(int32(hotTTL.(int)))
	}
	if infrequentAccessTTL, ok := d.GetOk("infrequent_access_ttl"); ok {
		logstore.InfrequentAccessTTL = tea.Int32(int32(infrequentAccessTTL.(int)))
	}
	if encrypt := buildEncrypt(d); encrypt != nil {
		logstore.EncryptConf = encrypt
	}

	return logstore
}

func buildEncrypt(d *schema.ResourceData) *aliyunSlsAPI.EncryptConf {
	var encryptConf *aliyunSlsAPI.EncryptConf
	if field, ok := d.GetOk("encrypt_conf"); ok {
		encryptConf = new(aliyunSlsAPI.EncryptConf)
		value := field.([]interface{})[0].(map[string]interface{})
		encryptConf.Enable = value["enable"].(bool)
		encryptConf.EncryptType = value["encrypt_type"].(string)
		cmkInfo := value["user_cmk_info"].([]interface{})
		if len(cmkInfo) > 0 {
			cmk := cmkInfo[0].(map[string]interface{})
			encryptConf.UserCmkInfo = &aliyunSlsAPI.UserCmkInfo{
				CmkKeyId: cmk["cmk_key_id"].(string),
				Arn:      cmk["arn"].(string),
				RegionId: cmk["region_id"].(string),
			}
		}
	}
	return encryptConf
}
