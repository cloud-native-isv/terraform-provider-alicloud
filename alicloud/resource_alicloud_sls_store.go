// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
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
			Update: schema.DefaultTimeout(5 * time.Minute),
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
			},
			"infrequent_access_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
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
			"retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"shard_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "" {
						return false
					}
					return true
				},
				Default: 2,
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

	if v, ok := d.GetOk("telemetry_type"); ok && v == "Metrics" {
		client := meta.(*connectivity.AliyunClient)
		slsService, err := NewSlsService(client)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
		}

		projectName := d.Get("project_name").(string)
		if v, ok := d.GetOkExists("project"); ok {
			projectName = v.(string)
		}
		logstoreName := d.Get("logstore_name").(string)
		if v, ok := d.GetOkExists("name"); ok {
			logstoreName = v.(string)
		}

		logstore := buildLogStore(d)
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			err := slsService.GetAPI().CreateLogStore(projectName, logstore)
			if err != nil {
				// Handle LogStoreAlreadyExist error by importing existing resource
				if IsExpectedErrors(err, []string{"LogStoreAlreadyExist"}) {
					log.Printf("[INFO] LogStore %s already exists, importing existing resource", logstoreName)
					d.SetId(fmt.Sprintf("%s%s%s", projectName, COLON_SEPARATED, logstoreName))
					return nil
				}
				if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
					time.Sleep(10 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "CreateLogStoreV2", AliyunLogGoSdkERROR)
		}

		// Set ID if not already set (for new resources)
		if d.Id() == "" {
			d.SetId(fmt.Sprintf("%s%s%s", projectName, COLON_SEPARATED, logstoreName))
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

	client := meta.(*connectivity.AliyunClient)

	action := fmt.Sprintf("/logstores")
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	body := make(map[string]interface{})
	hostMap := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	request["logstoreName"] = d.Get("logstore_name")
	hostMap["project"] = StringPointer(d.Get("project_name").(string))
	if v, ok := d.GetOkExists("project"); ok {
		hostMap["project"] = tea.String(v.(string))
	}
	if v, ok := d.GetOkExists("name"); ok {
		request["logstoreName"] = v
	}

	request["shardCount"] = 2
	if v, ok := d.GetOkExists("shard_count"); ok {
		request["shardCount"] = v
	}
	if v, ok := d.GetOk("auto_split"); ok {
		request["autoSplit"] = v
	}
	if v, ok := d.GetOk("append_meta"); ok {
		request["appendMeta"] = v
	}
	if v, ok := d.GetOk("telemetry_type"); ok {
		request["telemetryType"] = v
	}
	if v, ok := d.GetOk("hot_ttl"); ok {
		request["hot_ttl"] = v
	}
	if v, ok := d.GetOk("mode"); ok {
		request["mode"] = v
	}
	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("encrypt_conf"); !IsNil(v) {
		enable1, _ := jsonpath.Get("$[0].enable", d.Get("encrypt_conf"))
		if enable1 != nil && enable1 != "" {
			objectDataLocalMap["enable"] = enable1
		}
		encryptType, _ := jsonpath.Get("$[0].encrypt_type", d.Get("encrypt_conf"))
		if encryptType != nil && encryptType != "" {
			objectDataLocalMap["encrypt_type"] = encryptType
		}
		user_cmk_info := make(map[string]interface{})
		cmkKeyId, _ := jsonpath.Get("$[0].user_cmk_info[0].cmk_key_id", d.Get("encrypt_conf"))
		if cmkKeyId != nil && cmkKeyId != "" {
			user_cmk_info["cmk_key_id"] = cmkKeyId
		}
		arn1, _ := jsonpath.Get("$[0].user_cmk_info[0].arn", d.Get("encrypt_conf"))
		if arn1 != nil && arn1 != "" {
			user_cmk_info["arn"] = arn1
		}
		regionId, _ := jsonpath.Get("$[0].user_cmk_info[0].region_id", d.Get("encrypt_conf"))
		if regionId != nil && regionId != "" {
			user_cmk_info["region_id"] = regionId
		}

		user_cmk_info_map, _ := jsonpath.Get("$[0].user_cmk_info[0]", v)
		if !IsNil(user_cmk_info_map) {
			objectDataLocalMap["user_cmk_info"] = user_cmk_info
		}

		request["encrypt_conf"] = objectDataLocalMap
	}

	request["ttl"] = 30
	if v, ok := d.GetOk("retention_period"); ok {
		request["ttl"] = v
	}
	if v, ok := d.GetOk("max_split_shard_count"); ok {
		request["maxSplitShard"] = v
	}
	if v, ok := d.GetOk("enable_web_tracking"); ok {
		request["enable_tracking"] = v
	}
	if v, ok := d.GetOk("infrequent_access_ttl"); ok {
		request["infrequentAccessTTL"] = v
	}

	body = request
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = client.Do("Sls", roaParam("POST", "2020-12-30", "CreateLogStore", action), query, body, nil, hostMap, false)
		if err != nil {
			// Handle LogStoreAlreadyExist error by importing existing resource
			if IsExpectedErrors(err, []string{"LogStoreAlreadyExist"}) {
				log.Printf("[INFO] LogStore %s already exists, importing existing resource", request["logstoreName"])
				d.SetId(fmt.Sprintf("%v:%v", *hostMap["project"], request["logstoreName"]))
				return nil
			}
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", action, AlibabaCloudSdkGoERROR)
	}

	// Set ID if not already set (for new resources)
	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%v:%v", *hostMap["project"], request["logstoreName"]))
	}

	if v, ok := d.GetOk("max_split_shard_count"); ok {
		d.Set("max_split_shard_count", v)
	}

	// Wait for the logstore to be available using state refresh function
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

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
	d.Set("hot_ttl", logstore.HotTtl)
	d.Set("infrequent_access_ttl", logstore.InfrequentAccessTTL)
	d.Set("max_split_shard_count", logstore.MaxSplitShard)
	d.Set("mode", logstore.Mode)
	d.Set("retention_period", logstore.Ttl)
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
	for _, s := range shards {
		mapping := map[string]interface{}{
			"id":        s.ShardId,
			"status":    s.Status,
			"begin_key": s.InclusiveBeginKey,
			"end_key":   s.ExclusiveEndKey,
		}
		shardList = append(shardList, mapping)
	}
	d.Set("shards", shardList)
	return nil
}

func resourceAliCloudSlsLogStoreUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	var body map[string]interface{}
	update := false
	d.Partial(true)
	parts := strings.Split(d.Id(), ":")
	logstore := parts[1]
	action := fmt.Sprintf("/logstores/%s", logstore)
	var err error
	request = make(map[string]interface{})
	query = make(map[string]*string)
	body = make(map[string]interface{})
	hostMap := make(map[string]*string)
	hostMap["project"] = StringPointer(parts[0])

	if !d.IsNewResource() && d.HasChange("auto_split") {
		update = true
	}
	if v, ok := d.GetOk("auto_split"); ok || d.HasChange("auto_split") {
		request["autoSplit"] = v
	}
	if !d.IsNewResource() && d.HasChange("append_meta") {
		update = true
	}
	if v, ok := d.GetOk("append_meta"); ok || d.HasChange("append_meta") {
		request["appendMeta"] = v
	}
	if !d.IsNewResource() && d.HasChange("hot_ttl") {
		update = true
	}
	if v, ok := d.GetOk("hot_ttl"); ok || d.HasChange("hot_ttl") {
		request["hot_ttl"] = v
	}
	if !d.IsNewResource() && d.HasChange("mode") {
		update = true
	}
	if v, ok := d.GetOk("mode"); ok || d.HasChange("mode") {
		request["mode"] = v
	}
	if !d.IsNewResource() && d.HasChange("retention_period") {
		update = true
	}
	request["ttl"] = 30
	if v, ok := d.GetOk("retention_period"); ok {
		request["ttl"] = v
	}
	if !d.IsNewResource() && d.HasChange("max_split_shard_count") {
		update = true
	}
	if v, ok := d.GetOk("max_split_shard_count"); ok || d.HasChange("max_split_shard_count") {
		request["maxSplitShard"] = v
	}
	if !d.IsNewResource() && d.HasChange("enable_web_tracking") {
		update = true
	}
	if v, ok := d.GetOk("enable_web_tracking"); ok || d.HasChange("enable_web_tracking") {
		request["enable_tracking"] = v
	}
	if !d.IsNewResource() && d.HasChange("encrypt_conf") {
		update = true
	}
	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("encrypt_conf"); !IsNil(v) || d.HasChange("encrypt_conf") {
		enable1, _ := jsonpath.Get("$[0].enable", v)
		if enable1 != nil && (d.HasChange("encrypt_conf.0.enable") || enable1 != "") {
			objectDataLocalMap["enable"] = enable1
		}
		encryptType, _ := jsonpath.Get("$[0].encrypt_type", v)
		if encryptType != nil && (d.HasChange("encrypt_conf.0.encrypt_type") || encryptType != "") {
			objectDataLocalMap["encrypt_type"] = encryptType
		}
		user_cmk_info := make(map[string]interface{})
		cmkKeyId, _ := jsonpath.Get("$[0].user_cmk_info[0].cmk_key_id", v)
		if cmkKeyId != nil && (d.HasChange("encrypt_conf.0.user_cmk_info.0.cmk_key_id") || cmkKeyId != "") {
			user_cmk_info["cmk_key_id"] = cmkKeyId
		}
		arn1, _ := jsonpath.Get("$[0].user_cmk_info[0].arn", v)
		if arn1 != nil && (d.HasChange("encrypt_conf.0.user_cmk_info.0.arn") || arn1 != "") {
			user_cmk_info["arn"] = arn1
		}
		regionId, _ := jsonpath.Get("$[0].user_cmk_info[0].region_id", v)
		if regionId != nil && (d.HasChange("encrypt_conf.0.user_cmk_info.0.region_id") || regionId != "") {
			user_cmk_info["region_id"] = regionId
		}

		user_cmk_info_map, _ := jsonpath.Get("$[0].user_cmk_info[0]", v)
		if !IsNil(user_cmk_info_map) {
			objectDataLocalMap["user_cmk_info"] = user_cmk_info
		}

		request["encrypt_conf"] = objectDataLocalMap
	}

	if !d.IsNewResource() && d.HasChange("infrequent_access_ttl") {
		update = true
	}
	if v, ok := d.GetOk("infrequent_access_ttl"); ok || d.HasChange("infrequent_access_ttl") {
		request["infrequentAccessTTL"] = v
	}

	if v, ok := d.GetOk("telemetry_type"); ok && v == "Metrics" {

		projectName := d.Get("project_name").(string)
		if v, ok := d.GetOkExists("project"); ok {
			projectName = v.(string)
		}

		logstore := buildLogStore(d)
		slsService, err := NewSlsService(client)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
		}

		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			err := slsService.GetAPI().UpdateLogStore(projectName, logstore.LogstoreName, logstore)
			if err != nil {
				if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
					time.Sleep(10 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "UpdateLogStore", AliyunLogGoSdkERROR)
		}

		if v, ok := d.GetOk("max_split_shard_count"); ok {
			d.Set("max_split_shard_count", v)
		}

		// Wait for the updated logstore to be available using state refresh function
		stateConf := BuildStateConf([]string{}, []string{"available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.LogStoreStateRefreshFunc(d.Id(), "logstoreName", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		update = false
	}
	body = request
	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.Do("Sls", roaParam("PUT", "2020-12-30", "UpdateLogStore", action), query, body, nil, hostMap, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		if v, ok := d.GetOk("max_split_shard_count"); ok {
			d.Set("max_split_shard_count", v)
		}

		// Wait for the updated logstore to be available using state refresh function
		slsService, err := NewSlsService(client)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
		}

		stateConf := BuildStateConf([]string{}, []string{"available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.LogStoreStateRefreshFunc(d.Id(), "logstoreName", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}
	update = false
	parts = strings.Split(d.Id(), ":")
	logstore = parts[1]
	action = fmt.Sprintf("/logstores/%s/meteringmode", logstore)
	request = make(map[string]interface{})
	query = make(map[string]*string)
	body = make(map[string]interface{})
	hostMap = make(map[string]*string)
	hostMap["project"] = StringPointer(parts[0])

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "NewSlsService", AlibabaCloudSdkGoERROR)
	}
	meteringModeObj, _ := slsService.DescribeGetLogStoreMeteringMode(d.Id())
	if d.HasChange("metering_mode") && meteringModeObj != nil && meteringModeObj.MeteringMode != d.Get("metering_mode").(string) {
		update = true
	}
	request["meteringMode"] = d.Get("metering_mode")

	body = request
	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.Do("Sls", roaParam("PUT", "2020-12-30", "UpdateLogStoreMeteringMode", action), query, body, nil, hostMap, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		// Wait for the metering mode update to complete using state refresh function
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
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	hostMap := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	hostMap["project"] = StringPointer(parts[0])

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.Do("Sls", roaParam("DELETE", "2020-12-30", "DeleteLogStore", action), query, nil, nil, hostMap, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
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
		Ttl:            int32(d.Get("retention_period").(int)),
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
		logstore.HotTtl = int32(hotTTL.(int))
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
