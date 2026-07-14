package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudOssBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOssBucketCreate,
		Read:   resourceAliCloudOssBucketRead,
		Update: resourceAliCloudOssBucketUpdate,
		Delete: resourceAliCloudOssBucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		// 完整 Timeouts 块 (Create/Update/Delete): 与 ots_table/ots_instance 同形.
		// 实证: 完整块 import 后不残留 `- timeouts {}` (仅声明 Delete 的旧写法才残留);
		// 同时保留 Delete=60min (force_destroy 大桶 PruneBucket 重试窗口) 与用户显式覆盖能力.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: StringLenBetween(3, 63),
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extranet_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"intranet_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_class": {
				Type:     schema.TypeString,
				Default:  oss.StorageStandard,
				Optional: true,
				ForceNew: true,
				ValidateFunc: StringInSlice([]string{
					string(oss.StorageStandard),
					string(oss.StorageIA),
					string(oss.StorageArchive),
					string(oss.StorageColdArchive),
					string(oss.StorageDeepColdArchive),
				}, false),
			},
			"redundancy_type": {
				Type:     schema.TypeString,
				Default:  oss.RedundancyLRS,
				Optional: true,
				ForceNew: true,
				ValidateFunc: StringInSlice([]string{
					string(oss.RedundancyLRS),
					string(oss.RedundancyZRS),
				}, false),
			},

			"tags": tagsSchema(),

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"resource_group_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"versioning": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: StringInSlice([]string{
								"Enabled",
								"Suspended",
							}, false),
						},
					},
				},
			},

			"server_side_encryption_rule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sse_algorithm": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: StringInSlice([]string{
								ServerSideEncryptionAes256,
								ServerSideEncryptionKMS,
								ServerSideEncryptionSM4,
							}, false),
						},
						"kms_master_key_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_data_encryption": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: StringInSlice([]string{
								ServerSideEncryptionSM4,
								"",
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudOssBucketCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var bucketName string
	if v, ok := d.GetOk("bucket"); ok && v != "" {
		bucketName = v.(string)
	} else {
		bucketName = resource.PrefixedUniqueId("tf-oss-bucket-")
		if len(bucketName) > 63 {
			bucketName = bucketName[:63]
		}
	}
	request := map[string]string{"bucketName": bucketName}
	var requestInfo *oss.Client
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.IsBucketExist(request["bucketName"])
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket", "IsBucketExist", AliyunOssGoSdk)
	}
	addDebug("IsBucketExist", raw, requestInfo, request)
	isExist, _ := raw.(bool)
	if isExist {
		return WrapError(Error("[ERROR] The specified bucket name: %#v is not available. The bucket namespace is shared by all users of the OSS system. Please select a different name and try again.", request["bucketName"]))
	}

	options := []oss.Option{
		oss.StorageClass(oss.StorageClassType(d.Get("storage_class").(string))),
		oss.RedundancyType(oss.DataRedundancyType(d.Get("redundancy_type").(string))),
	}

	//resource_group_id
	if resourceGroupId, ok := d.Get("resource_group_id").(string); ok && len(resourceGroupId) > 0 {
		options = append(options, oss.SetHeader("x-oss-resource-group-id", resourceGroupId))
	}

	raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return nil, ossClient.CreateBucket(bucketName, options...)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket", "CreateBucket", AliyunOssGoSdk)
	}
	addDebug("CreateBucket", raw, requestInfo, request)

	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return ossClient.IsBucketExist(request["bucketName"])
		})

		if err != nil {
			return resource.NonRetryableError(err)
		}
		isExist, _ := raw.(bool)
		if !isExist {
			return resource.RetryableError(Error("Trying to ensure new OSS bucket %#v has been created successfully.", request["bucketName"]))
		}
		addDebug("IsBucketExist", raw, requestInfo, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket", "IsBucketExist", AliyunOssGoSdk)
	}

	// Assign the bucket name as the resource ID
	d.SetId(request["bucketName"])

	// 新建后按声明 PUT versioning / SSE; 仅在声明了对应 block 时才动,
	// 避免空 server_side_encryption_rule 触发 DeleteBucketEncryption.
	if _, ok := d.GetOk("versioning"); ok {
		if err := resourceAliCloudOssBucketInlineVersioningUpdate(client, d); err != nil {
			return WrapError(err)
		}
	}
	if _, ok := d.GetOk("server_side_encryption_rule"); ok {
		if err := resourceAliCloudOssBucketInlineEncryptionUpdate(client, d); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudOssBucketRead(d, meta)
}

func resourceAliCloudOssBucketRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	object, err := ossService.DescribeOssBucket(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("bucket", d.Id())
	// force_destroy 是 state-only 删除标志 (API 不返回); 回写 d.Get 保留配置值,
	// import 时取 default false -> 匹配 HCL, 避免永久 +force_destroy diff.
	d.Set("force_destroy", d.Get("force_destroy"))
	d.Set("creation_date", object.BucketInfo.CreationDate.Format("2006-01-02"))
	d.Set("extranet_endpoint", object.BucketInfo.ExtranetEndpoint)
	d.Set("intranet_endpoint", object.BucketInfo.IntranetEndpoint)
	d.Set("location", object.BucketInfo.Location)
	d.Set("owner", object.BucketInfo.Owner.ID)
	d.Set("storage_class", object.BucketInfo.StorageClass)
	d.Set("redundancy_type", object.BucketInfo.RedundancyType)

	request := map[string]string{"bucketName": d.Id()}
	var requestInfo *oss.Client

	// Read tags
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.GetBucketTagging(d.Id())
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketTagging", AliyunOssGoSdk)
	}
	addDebug("GetBucketTagging", raw, requestInfo, request)
	tagging, _ := raw.(oss.GetBucketTaggingResult)
	tagsMap := make(map[string]string)
	if len(tagging.Tags) > 0 {
		for _, t := range tagging.Tags {
			tagsMap[t.Key] = t.Value
		}
	}
	if err := d.Set("tags", tagsMap); err != nil {
		return WrapError(err)
	}

	// Read the bucket resource-group-id
	raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.GetBucketResourceGroup(d.Id())
	})
	if err != nil && !IsExpectedErrors(err, []string{"NotImplemented"}) {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketResourceGroup", AliyunOssGoSdk)
	}
	if err == nil {
		addDebug("GetBucketResourceGroup", raw, requestInfo, request)
		resourceGroup, _ := raw.(oss.GetBucketResourceGroupResult)
		d.Set("resource_group_id", resourceGroup.ResourceGroupId)
	}

	// versioning / server_side_encryption_rule 必须直连经典 ossClient 读真值:
	// ossService.DescribeOssBucket 经 cws-lib-go convertBucketInfoToLegacy 丢弃 Versioning/SseRule,
	// 用它会导致 import 后这两块恒空 -> 与 HCL 永久 diff.
	raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.GetBucketInfo(d.Id())
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketInfo", AliyunOssGoSdk)
	}
	addDebug("GetBucketInfo", raw, requestInfo, request)
	if info, ok := raw.(oss.GetBucketInfoResult); ok {
		if len(info.BucketInfo.SseRule.SSEAlgorithm) > 0 && info.BucketInfo.SseRule.SSEAlgorithm != "None" {
			rule := map[string]interface{}{"sse_algorithm": info.BucketInfo.SseRule.SSEAlgorithm}
			if info.BucketInfo.SseRule.KMSMasterKeyID != "" {
				rule["kms_master_key_id"] = info.BucketInfo.SseRule.KMSMasterKeyID
			}
			if info.BucketInfo.SseRule.KMSDataEncryption != "" {
				rule["kms_data_encryption"] = info.BucketInfo.SseRule.KMSDataEncryption
			}
			d.Set("server_side_encryption_rule", []map[string]interface{}{rule})
		}
		if info.BucketInfo.Versioning != "" {
			d.Set("versioning", []map[string]interface{}{{"status": info.BucketInfo.Versioning}})
		}
	}

	return nil
}

func resourceAliCloudOssBucketUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	d.Partial(true)

	if d.HasChange("tags") {
		if err := resourceAliCloudOssBucketTaggingUpdate(client, d); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
	}

	if !d.IsNewResource() && d.HasChange("resource_group_id") {
		resourceGroupId := d.Get("resource_group_id").(string)
		request := map[string]string{"bucketName": d.Id(), "resourceGroupId": resourceGroupId}
		var requestInfo *oss.Client
		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			requestInfo = ossClient
			return nil, ossClient.PutBucketResourceGroup(d.Id(), oss.PutBucketResourceGroup{
				ResourceGroupId: resourceGroupId,
			})
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "PutBucketResourceGroup", AliyunOssGoSdk)
		}
		addDebug("PutBucketResourceGroup", raw, requestInfo, request)
		d.SetPartial("resource_group_id")
	}

	if d.HasChange("versioning") {
		if err := resourceAliCloudOssBucketInlineVersioningUpdate(client, d); err != nil {
			return WrapError(err)
		}
		d.SetPartial("versioning")
	}

	if d.HasChange("server_side_encryption_rule") {
		if err := resourceAliCloudOssBucketInlineEncryptionUpdate(client, d); err != nil {
			return WrapError(err)
		}
		d.SetPartial("server_side_encryption_rule")
	}

	d.Partial(false)
	return resourceAliCloudOssBucketRead(d, meta)
}

// resourceAliCloudOssBucketInlineVersioningUpdate 按 versioning block 调经典 SetBucketVersioning (移植自 v1.283.0).
func resourceAliCloudOssBucketInlineVersioningUpdate(client *connectivity.AliyunClient, d *schema.ResourceData) error {
	versioning := d.Get("versioning").([]interface{})
	if len(versioning) == 1 {
		var status string
		c := versioning[0].(map[string]interface{})
		if v, ok := c["status"]; ok {
			status = v.(string)
		}
		versioningCfg := oss.VersioningConfig{}
		versioningCfg.Status = status
		var requestInfo *oss.Client
		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			requestInfo = ossClient
			return nil, ossClient.SetBucketVersioning(d.Id(), versioningCfg)
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketVersioning", AliyunOssGoSdk)
		}
		addDebug("SetBucketVersioning", raw, requestInfo, map[string]interface{}{
			"bucketName":       d.Id(),
			"versioningConfig": versioningCfg,
		})
	}
	return nil
}

// resourceAliCloudOssBucketInlineEncryptionUpdate 按 server_side_encryption_rule block 调经典 SetBucketEncryption;
// 空 block 时删除 SSE (移植自 v1.283.0). 注: 与独立子资源 alicloud_oss_bucket_server_side_encryption
// MUST NOT 对同一桶并管 (空配置会 DeleteBucketEncryption 真删云端加密).
func resourceAliCloudOssBucketInlineEncryptionUpdate(client *connectivity.AliyunClient, d *schema.ResourceData) error {
	encryptionRule := d.Get("server_side_encryption_rule").([]interface{})
	var requestInfo *oss.Client
	if len(encryptionRule) == 0 {
		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			requestInfo = ossClient
			return nil, ossClient.DeleteBucketEncryption(d.Id())
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucketEncryption", AliyunOssGoSdk)
		}
		addDebug("DeleteBucketEncryption", raw, requestInfo, map[string]string{"bucketName": d.Id()})
		return nil
	}

	var sseRule oss.ServerEncryptionRule
	c := encryptionRule[0].(map[string]interface{})
	if v, ok := c["sse_algorithm"]; ok {
		sseRule.SSEDefault.SSEAlgorithm = v.(string)
	}
	if v, ok := c["kms_master_key_id"]; ok {
		sseRule.SSEDefault.KMSMasterKeyID = v.(string)
	}
	if v, ok := c["kms_data_encryption"]; ok {
		sseRule.SSEDefault.KMSDataEncryption = v.(string)
	}

	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return nil, ossClient.SetBucketEncryption(d.Id(), sseRule)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketEncryption", AliyunOssGoSdk)
	}
	addDebug("SetBucketEncryption", raw, requestInfo, map[string]interface{}{
		"bucketName":     d.Id(),
		"encryptionRule": sseRule,
	})
	return nil
}

func resourceAliCloudOssBucketDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	var requestInfo *oss.Client
	forceDestroy := d.Get("force_destroy").(bool)
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.IsBucketExist(d.Id())
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "IsBucketExist", AliyunOssGoSdk)
	}
	addDebug("IsBucketExist", raw, requestInfo, map[string]string{"bucketName": d.Id()})

	exist, _ := raw.(bool)
	if !exist {
		return nil
	}

	log.Printf("[DEBUG] OSS bucket delete force_destroy=%t, bucket=%s", forceDestroy, d.Id())
	if forceDestroy {
		// Prune bucket contents before deletion when force_destroy is enabled.
		err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
			err := ossService.PruneBucket(d.Id())
			if err != nil {
				if NeedRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "PruneBucket", "OSS API")
		}
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return nil, ossClient.DeleteBucket(d.Id())
		})
		if err != nil {
			if IsExpectedErrors(err, []string{"BucketNotEmpty"}) {
				if d.Get("force_destroy").(bool) {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}
		addDebug("DeleteBucket", raw, requestInfo, map[string]string{"bucketName": d.Id()})
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucket", AliyunOssGoSdk)
	}
	return WrapError(ossService.WaitForOssBucket(d.Id(), Deleted, DefaultTimeoutMedium))
}

func resourceAliCloudOssBucketTaggingUpdate(client *connectivity.AliyunClient, d *schema.ResourceData) error {
	tagsMap := d.Get("tags").(map[string]interface{})
	var requestInfo *oss.Client
	if tagsMap == nil || len(tagsMap) == 0 {
		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			requestInfo = ossClient
			return nil, ossClient.DeleteBucketTagging(d.Id())
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucketTagging", AliyunOssGoSdk)
		}
		addDebug("DeleteBucketTagging", raw, requestInfo, map[string]string{"bucketName": d.Id()})
		return nil
	}

	// Put tagging
	var bTagging oss.Tagging
	for k, v := range tagsMap {
		bTagging.Tags = append(bTagging.Tags, oss.Tag{
			Key:   k,
			Value: v.(string),
		})
	}
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return nil, ossClient.SetBucketTagging(d.Id(), bTagging)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketTagging", AliyunOssGoSdk)
	}
	addDebug("SetBucketTagging", raw, requestInfo, map[string]interface{}{
		"bucketName": d.Id(),
		"tagging":    bTagging,
	})
	return nil
}
