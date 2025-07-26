package alicloud

import (
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

	return resourceAliCloudOssBucketRead(d, meta)
}

func resourceAliCloudOssBucketRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	object, err := ossService.DescribeOssBucket(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("bucket", d.Id())
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

	d.Partial(false)
	return resourceAliCloudOssBucketRead(d, meta)
}

func resourceAliCloudOssBucketDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)
	var requestInfo *oss.Client
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

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return nil, ossClient.DeleteBucket(d.Id())
		})
		if err != nil {
			if IsExpectedErrors(err, []string{"BucketNotEmpty"}) {
				if d.Get("force_destroy").(bool) {
					raw, er := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
						bucket, _ := ossClient.Bucket(d.Id())
						lor, err := bucket.ListObjectVersions()
						if err != nil {
							return nil, WrapErrorf(err, DefaultErrorMsg, d.Id(), "ListObjectVersions", AliyunOssGoSdk)
						}
						addDebug("ListObjectVersions", lor, requestInfo)
						objectsToDelete := make([]oss.DeleteObject, 0)
						for _, object := range lor.ObjectDeleteMarkers {
							objectsToDelete = append(objectsToDelete, oss.DeleteObject{
								Key:       object.Key,
								VersionId: object.VersionId,
							})
						}

						for _, object := range lor.ObjectVersions {
							objectsToDelete = append(objectsToDelete, oss.DeleteObject{
								Key:       object.Key,
								VersionId: object.VersionId,
							})
						}
						return bucket.DeleteObjectVersions(objectsToDelete)
					})
					if er != nil {
						return resource.NonRetryableError(er)
					}
					addDebug("DeleteObjectVersions", raw, requestInfo, map[string]string{"bucketName": d.Id()})
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
