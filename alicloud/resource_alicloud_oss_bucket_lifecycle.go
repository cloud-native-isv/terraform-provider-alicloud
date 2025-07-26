// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOssBucketLifecycle() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOssBucketLifecycleCreate,
		Read:   resourceAliCloudOssBucketLifecycleRead,
		Update: resourceAliCloudOssBucketLifecycleUpdate,
		Delete: resourceAliCloudOssBucketLifecycleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"expiration": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"created_before_date": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"transitions": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"created_before_date": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"IA", "Archive", "ColdArchive", "DeepColdArchive"}, false),
									},
									"is_access_time": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"return_to_std_when_visit": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"allow_small_file": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"abort_multipart_upload": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"created_before_date": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"IA", "Archive", "ColdArchive", "DeepColdArchive"}, false),
									},
									"is_access_time": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"return_to_std_when_visit": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"filter": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"not": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"tag": {
													Type:     schema.TypeSet,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key": {
																Type:     schema.TypeString,
																Required: true,
															},
															"value": {
																Type:     schema.TypeString,
																Required: true,
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
						"tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudOssBucketLifecycleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	action := fmt.Sprintf("/?lifecycle")
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	body := make(map[string]interface{})
	hostMap := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	hostMap["bucket"] = StringPointer(d.Get("bucket").(string))

	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("rule"); !IsNil(v) {
		if v, ok := d.GetOk("rule"); ok {
			localData, err := jsonpath.Get("$", v)
			if err != nil {
				localData = make([]interface{}, 0)
			}
			localMaps := make([]interface{}, 0)
			for _, dataLoop := range localData.(*schema.Set).List() {
				dataLoopTmp := make(map[string]interface{})
				if dataLoop != nil {
					dataLoopTmp = dataLoop.(map[string]interface{})
				}
				dataLoopMap := make(map[string]interface{})

				if v, ok := dataLoopTmp["id"]; ok && v != "" {
					dataLoopMap["ID"] = v
				}
				if v, ok := dataLoopTmp["prefix"]; ok && v != "" {
					dataLoopMap["Prefix"] = v
				}
				if v, ok := dataLoopTmp["enabled"]; ok {
					if v.(bool) {
						dataLoopMap["Status"] = "Enabled"
					} else {
						dataLoopMap["Status"] = "Disabled"
					}
				}

				// Handle expiration
				if v, ok := dataLoopTmp["expiration"]; ok && v != nil {
					expirationSet := v.(*schema.Set)
					if expirationSet.Len() > 0 {
						expiration := expirationSet.List()[0].(map[string]interface{})
						expirationMap := make(map[string]interface{})
						if date, ok := expiration["date"]; ok && date != "" {
							expirationMap["Date"] = date
						}
						if days, ok := expiration["days"]; ok && days != 0 {
							expirationMap["Days"] = days
						}
						if createdBeforeDate, ok := expiration["created_before_date"]; ok && createdBeforeDate != "" {
							expirationMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if expiredObjectDeleteMarker, ok := expiration["expired_object_delete_marker"]; ok {
							expirationMap["ExpiredObjectDeleteMarker"] = expiredObjectDeleteMarker
						}
						if len(expirationMap) > 0 {
							dataLoopMap["Expiration"] = expirationMap
						}
					}
				}

				// Handle transitions
				if v, ok := dataLoopTmp["transitions"]; ok && v != nil {
					transitionsSet := v.(*schema.Set)
					transitionMaps := make([]interface{}, 0)
					for _, transition := range transitionsSet.List() {
						transitionMap := make(map[string]interface{})
						transitionTmp := transition.(map[string]interface{})

						if createdBeforeDate, ok := transitionTmp["created_before_date"]; ok && createdBeforeDate != "" {
							transitionMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if days, ok := transitionTmp["days"]; ok && days != 0 {
							transitionMap["Days"] = days
						}
						if storageClass, ok := transitionTmp["storage_class"]; ok {
							transitionMap["StorageClass"] = storageClass
						}
						if isAccessTime, ok := transitionTmp["is_access_time"]; ok {
							transitionMap["IsAccessTime"] = isAccessTime
						}
						if returnToStdWhenVisit, ok := transitionTmp["return_to_std_when_visit"]; ok {
							transitionMap["ReturnToStdWhenVisit"] = returnToStdWhenVisit
						}
						if allowSmallFile, ok := transitionTmp["allow_small_file"]; ok {
							transitionMap["AllowSmallFile"] = allowSmallFile
						}

						transitionMaps = append(transitionMaps, transitionMap)
					}
					if len(transitionMaps) > 0 {
						dataLoopMap["Transition"] = transitionMaps
					}
				}

				// Handle abort_multipart_upload
				if v, ok := dataLoopTmp["abort_multipart_upload"]; ok && v != nil {
					abortSet := v.(*schema.Set)
					if abortSet.Len() > 0 {
						abort := abortSet.List()[0].(map[string]interface{})
						abortMap := make(map[string]interface{})
						if days, ok := abort["days"]; ok {
							abortMap["Days"] = days
						}
						if createdBeforeDate, ok := abort["created_before_date"]; ok && createdBeforeDate != "" {
							abortMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if len(abortMap) > 0 {
							dataLoopMap["AbortMultipartUpload"] = abortMap
						}
					}
				}

				// Handle noncurrent_version_expiration
				if v, ok := dataLoopTmp["noncurrent_version_expiration"]; ok && v != nil {
					nvExpirationSet := v.(*schema.Set)
					if nvExpirationSet.Len() > 0 {
						nvExpiration := nvExpirationSet.List()[0].(map[string]interface{})
						nvExpirationMap := make(map[string]interface{})
						if days, ok := nvExpiration["days"]; ok {
							nvExpirationMap["NoncurrentDays"] = days
						}
						if len(nvExpirationMap) > 0 {
							dataLoopMap["NoncurrentVersionExpiration"] = nvExpirationMap
						}
					}
				}

				// Handle noncurrent_version_transition
				if v, ok := dataLoopTmp["noncurrent_version_transition"]; ok && v != nil {
					nvTransitionsSet := v.(*schema.Set)
					nvTransitionMaps := make([]interface{}, 0)
					for _, nvTransition := range nvTransitionsSet.List() {
						nvTransitionMap := make(map[string]interface{})
						nvTransitionTmp := nvTransition.(map[string]interface{})

						if days, ok := nvTransitionTmp["days"]; ok {
							nvTransitionMap["NoncurrentDays"] = days
						}
						if storageClass, ok := nvTransitionTmp["storage_class"]; ok {
							nvTransitionMap["StorageClass"] = storageClass
						}
						if isAccessTime, ok := nvTransitionTmp["is_access_time"]; ok {
							nvTransitionMap["IsAccessTime"] = isAccessTime
						}
						if returnToStdWhenVisit, ok := nvTransitionTmp["return_to_std_when_visit"]; ok {
							nvTransitionMap["ReturnToStdWhenVisit"] = returnToStdWhenVisit
						}

						nvTransitionMaps = append(nvTransitionMaps, nvTransitionMap)
					}
					if len(nvTransitionMaps) > 0 {
						dataLoopMap["NoncurrentVersionTransition"] = nvTransitionMaps
					}
				}

				// Handle filter
				if v, ok := dataLoopTmp["filter"]; ok && v != nil {
					filterSet := v.(*schema.Set)
					if filterSet.Len() > 0 {
						filter := filterSet.List()[0].(map[string]interface{})
						filterMap := make(map[string]interface{})

						if prefix, ok := filter["prefix"]; ok && prefix != "" {
							filterMap["Prefix"] = prefix
						}

						if tags, ok := filter["tag"]; ok && tags != nil {
							tagsSet := tags.(*schema.Set)
							tagMaps := make([]interface{}, 0)
							for _, tag := range tagsSet.List() {
								tagMap := make(map[string]interface{})
								tagTmp := tag.(map[string]interface{})
								tagMap["Key"] = tagTmp["key"]
								tagMap["Value"] = tagTmp["value"]
								tagMaps = append(tagMaps, tagMap)
							}
							if len(tagMaps) > 0 {
								filterMap["Tag"] = tagMaps
							}
						}

						if not, ok := filter["not"]; ok && not != nil {
							notSet := not.(*schema.Set)
							if notSet.Len() > 0 {
								notFilter := notSet.List()[0].(map[string]interface{})
								notMap := make(map[string]interface{})

								if prefix, ok := notFilter["prefix"]; ok && prefix != "" {
									notMap["Prefix"] = prefix
								}

								if tag, ok := notFilter["tag"]; ok && tag != nil {
									tagSet := tag.(*schema.Set)
									if tagSet.Len() > 0 {
										tagTmp := tagSet.List()[0].(map[string]interface{})
										tagMap := make(map[string]interface{})
										tagMap["Key"] = tagTmp["key"]
										tagMap["Value"] = tagTmp["value"]
										notMap["Tag"] = tagMap
									}
								}

								if len(notMap) > 0 {
									filterMap["Not"] = notMap
								}
							}
						}

						if len(filterMap) > 0 {
							dataLoopMap["Filter"] = filterMap
						}
					}
				}

				// Handle tags
				if v, ok := dataLoopTmp["tags"]; ok && v != nil {
					tags := v.(map[string]interface{})
					tagMaps := make([]interface{}, 0)
					for k, v := range tags {
						tagMap := make(map[string]interface{})
						tagMap["Key"] = k
						tagMap["Value"] = v
						tagMaps = append(tagMaps, tagMap)
					}
					if len(tagMaps) > 0 {
						dataLoopMap["Tag"] = tagMaps
					}
				}

				localMaps = append(localMaps, dataLoopMap)
			}
			objectDataLocalMap["Rule"] = localMaps
		}
	}

	request["LifecycleConfiguration"] = objectDataLocalMap
	body = request

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("PUT", "2019-05-17", "PutBucketLifecycle", action), query, body, nil, hostMap, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket_lifecycle", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(*hostMap["bucket"]))

	ossService := NewOssService(client)
	stateConf := BuildStateConf([]string{}, []string{"#CHECKSET"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, ossService.OssBucketLifecycleStateRefreshFunc(d.Id(), "#Rule", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudOssBucketLifecycleRead(d, meta)
}

func resourceAliCloudOssBucketLifecycleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := NewOssService(client)

	objectRaw, err := ossService.DescribeOssBucketLifecycle(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_oss_bucket_lifecycle DescribeOssBucketLifecycle Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	ruleRaw := objectRaw["Rule"]
	ruleMaps := make([]map[string]interface{}, 0)
	if ruleRaw != nil {
		for _, ruleChildRaw := range ruleRaw.([]interface{}) {
			ruleMap := make(map[string]interface{})
			ruleChildRaw := ruleChildRaw.(map[string]interface{})

			ruleMap["id"] = ruleChildRaw["ID"]
			ruleMap["prefix"] = ruleChildRaw["Prefix"]

			if status, ok := ruleChildRaw["Status"]; ok {
				ruleMap["enabled"] = status == "Enabled"
			}

			// Handle expiration
			if expiration, ok := ruleChildRaw["Expiration"]; ok && expiration != nil {
				expirationMap := expiration.(map[string]interface{})
				expirationResult := make(map[string]interface{})

				if date, ok := expirationMap["Date"]; ok {
					expirationResult["date"] = date
				}
				if days, ok := expirationMap["Days"]; ok {
					expirationResult["days"] = days
				}
				if createdBeforeDate, ok := expirationMap["CreatedBeforeDate"]; ok {
					expirationResult["created_before_date"] = createdBeforeDate
				}
				if expiredObjectDeleteMarker, ok := expirationMap["ExpiredObjectDeleteMarker"]; ok {
					expirationResult["expired_object_delete_marker"] = expiredObjectDeleteMarker
				}

				if len(expirationResult) > 0 {
					ruleMap["expiration"] = []interface{}{expirationResult}
				}
			}

			// Handle transitions
			if transitions, ok := ruleChildRaw["Transition"]; ok && transitions != nil {
				transitionMaps := make([]interface{}, 0)
				for _, transition := range transitions.([]interface{}) {
					transitionMap := transition.(map[string]interface{})
					transitionResult := make(map[string]interface{})

					if createdBeforeDate, ok := transitionMap["CreatedBeforeDate"]; ok {
						transitionResult["created_before_date"] = createdBeforeDate
					}
					if days, ok := transitionMap["Days"]; ok {
						transitionResult["days"] = days
					}
					if storageClass, ok := transitionMap["StorageClass"]; ok {
						transitionResult["storage_class"] = storageClass
					}
					if isAccessTime, ok := transitionMap["IsAccessTime"]; ok {
						transitionResult["is_access_time"] = isAccessTime
					}
					if returnToStdWhenVisit, ok := transitionMap["ReturnToStdWhenVisit"]; ok {
						transitionResult["return_to_std_when_visit"] = returnToStdWhenVisit
					}
					if allowSmallFile, ok := transitionMap["AllowSmallFile"]; ok {
						transitionResult["allow_small_file"] = allowSmallFile
					}

					transitionMaps = append(transitionMaps, transitionResult)
				}
				if len(transitionMaps) > 0 {
					ruleMap["transitions"] = transitionMaps
				}
			}

			// Handle abort_multipart_upload
			if abortMultipartUpload, ok := ruleChildRaw["AbortMultipartUpload"]; ok && abortMultipartUpload != nil {
				abortMap := abortMultipartUpload.(map[string]interface{})
				abortResult := make(map[string]interface{})

				if days, ok := abortMap["Days"]; ok {
					abortResult["days"] = days
				}
				if createdBeforeDate, ok := abortMap["CreatedBeforeDate"]; ok {
					abortResult["created_before_date"] = createdBeforeDate
				}

				if len(abortResult) > 0 {
					ruleMap["abort_multipart_upload"] = []interface{}{abortResult}
				}
			}

			// Handle noncurrent_version_expiration
			if nvExpiration, ok := ruleChildRaw["NoncurrentVersionExpiration"]; ok && nvExpiration != nil {
				nvExpirationMap := nvExpiration.(map[string]interface{})
				nvExpirationResult := make(map[string]interface{})

				if days, ok := nvExpirationMap["NoncurrentDays"]; ok {
					nvExpirationResult["days"] = days
				}

				if len(nvExpirationResult) > 0 {
					ruleMap["noncurrent_version_expiration"] = []interface{}{nvExpirationResult}
				}
			}

			// Handle noncurrent_version_transition
			if nvTransitions, ok := ruleChildRaw["NoncurrentVersionTransition"]; ok && nvTransitions != nil {
				nvTransitionMaps := make([]interface{}, 0)
				for _, nvTransition := range nvTransitions.([]interface{}) {
					nvTransitionMap := nvTransition.(map[string]interface{})
					nvTransitionResult := make(map[string]interface{})

					if days, ok := nvTransitionMap["NoncurrentDays"]; ok {
						nvTransitionResult["days"] = days
					}
					if storageClass, ok := nvTransitionMap["StorageClass"]; ok {
						nvTransitionResult["storage_class"] = storageClass
					}
					if isAccessTime, ok := nvTransitionMap["IsAccessTime"]; ok {
						nvTransitionResult["is_access_time"] = isAccessTime
					}
					if returnToStdWhenVisit, ok := nvTransitionMap["ReturnToStdWhenVisit"]; ok {
						nvTransitionResult["return_to_std_when_visit"] = returnToStdWhenVisit
					}

					nvTransitionMaps = append(nvTransitionMaps, nvTransitionResult)
				}
				if len(nvTransitionMaps) > 0 {
					ruleMap["noncurrent_version_transition"] = nvTransitionMaps
				}
			}

			// Handle filter
			if filter, ok := ruleChildRaw["Filter"]; ok && filter != nil {
				filterMap := filter.(map[string]interface{})
				filterResult := make(map[string]interface{})

				if prefix, ok := filterMap["Prefix"]; ok {
					filterResult["prefix"] = prefix
				}

				if tags, ok := filterMap["Tag"]; ok && tags != nil {
					tagMaps := make([]interface{}, 0)
					for _, tag := range tags.([]interface{}) {
						tagMap := tag.(map[string]interface{})
						tagResult := make(map[string]interface{})
						tagResult["key"] = tagMap["Key"]
						tagResult["value"] = tagMap["Value"]
						tagMaps = append(tagMaps, tagResult)
					}
					if len(tagMaps) > 0 {
						filterResult["tag"] = tagMaps
					}
				}

				if not, ok := filterMap["Not"]; ok && not != nil {
					notMap := not.(map[string]interface{})
					notResult := make(map[string]interface{})

					if prefix, ok := notMap["Prefix"]; ok {
						notResult["prefix"] = prefix
					}

					if tag, ok := notMap["Tag"]; ok && tag != nil {
						tagMap := tag.(map[string]interface{})
						tagResult := make(map[string]interface{})
						tagResult["key"] = tagMap["Key"]
						tagResult["value"] = tagMap["Value"]
						notResult["tag"] = []interface{}{tagResult}
					}

					if len(notResult) > 0 {
						filterResult["not"] = []interface{}{notResult}
					}
				}

				if len(filterResult) > 0 {
					ruleMap["filter"] = []interface{}{filterResult}
				}
			}

			// Handle tags
			if tags, ok := ruleChildRaw["Tag"]; ok && tags != nil {
				tagsMap := make(map[string]interface{})
				for _, tag := range tags.([]interface{}) {
					tagMap := tag.(map[string]interface{})
					tagsMap[tagMap["Key"].(string)] = tagMap["Value"]
				}
				if len(tagsMap) > 0 {
					ruleMap["tags"] = tagsMap
				}
			}

			ruleMaps = append(ruleMaps, ruleMap)
		}
	}

	if err := d.Set("rule", ruleMaps); err != nil {
		return err
	}

	d.Set("bucket", d.Id())

	return nil
}

func resourceAliCloudOssBucketLifecycleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	var body map[string]interface{}
	update := false

	var err error
	action := fmt.Sprintf("/?lifecycle")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	body = make(map[string]interface{})
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(d.Id())

	objectDataLocalMap := make(map[string]interface{})

	if d.HasChange("rule") {
		update = true
	}

	if v := d.Get("rule"); v != nil {
		// Use the same conversion logic as in Create
		if v, ok := d.GetOk("rule"); ok {
			localData, err := jsonpath.Get("$", v)
			if err != nil {
				localData = make([]interface{}, 0)
			}
			localMaps := make([]interface{}, 0)
			for _, dataLoop := range localData.(*schema.Set).List() {
				dataLoopTmp := make(map[string]interface{})
				if dataLoop != nil {
					dataLoopTmp = dataLoop.(map[string]interface{})
				}
				dataLoopMap := make(map[string]interface{})

				if v, ok := dataLoopTmp["id"]; ok && v != "" {
					dataLoopMap["ID"] = v
				}
				if v, ok := dataLoopTmp["prefix"]; ok && v != "" {
					dataLoopMap["Prefix"] = v
				}
				if v, ok := dataLoopTmp["enabled"]; ok {
					if v.(bool) {
						dataLoopMap["Status"] = "Enabled"
					} else {
						dataLoopMap["Status"] = "Disabled"
					}
				}

				// Handle expiration
				if v, ok := dataLoopTmp["expiration"]; ok && v != nil {
					expirationSet := v.(*schema.Set)
					if expirationSet.Len() > 0 {
						expiration := expirationSet.List()[0].(map[string]interface{})
						expirationMap := make(map[string]interface{})
						if date, ok := expiration["date"]; ok && date != "" {
							expirationMap["Date"] = date
						}
						if days, ok := expiration["days"]; ok && days != 0 {
							expirationMap["Days"] = days
						}
						if createdBeforeDate, ok := expiration["created_before_date"]; ok && createdBeforeDate != "" {
							expirationMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if expiredObjectDeleteMarker, ok := expiration["expired_object_delete_marker"]; ok {
							expirationMap["ExpiredObjectDeleteMarker"] = expiredObjectDeleteMarker
						}
						if len(expirationMap) > 0 {
							dataLoopMap["Expiration"] = expirationMap
						}
					}
				}

				// Handle transitions
				if v, ok := dataLoopTmp["transitions"]; ok && v != nil {
					transitionsSet := v.(*schema.Set)
					transitionMaps := make([]interface{}, 0)
					for _, transition := range transitionsSet.List() {
						transitionMap := make(map[string]interface{})
						transitionTmp := transition.(map[string]interface{})

						if createdBeforeDate, ok := transitionTmp["created_before_date"]; ok && createdBeforeDate != "" {
							transitionMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if days, ok := transitionTmp["days"]; ok && days != 0 {
							transitionMap["Days"] = days
						}
						if storageClass, ok := transitionTmp["storage_class"]; ok {
							transitionMap["StorageClass"] = storageClass
						}
						if isAccessTime, ok := transitionTmp["is_access_time"]; ok {
							transitionMap["IsAccessTime"] = isAccessTime
						}
						if returnToStdWhenVisit, ok := transitionTmp["return_to_std_when_visit"]; ok {
							transitionMap["ReturnToStdWhenVisit"] = returnToStdWhenVisit
						}
						if allowSmallFile, ok := transitionTmp["allow_small_file"]; ok {
							transitionMap["AllowSmallFile"] = allowSmallFile
						}

						transitionMaps = append(transitionMaps, transitionMap)
					}
					if len(transitionMaps) > 0 {
						dataLoopMap["Transition"] = transitionMaps
					}
				}

				// Handle abort_multipart_upload
				if v, ok := dataLoopTmp["abort_multipart_upload"]; ok && v != nil {
					abortSet := v.(*schema.Set)
					if abortSet.Len() > 0 {
						abort := abortSet.List()[0].(map[string]interface{})
						abortMap := make(map[string]interface{})
						if days, ok := abort["days"]; ok {
							abortMap["Days"] = days
						}
						if createdBeforeDate, ok := abort["created_before_date"]; ok && createdBeforeDate != "" {
							abortMap["CreatedBeforeDate"] = createdBeforeDate
						}
						if len(abortMap) > 0 {
							dataLoopMap["AbortMultipartUpload"] = abortMap
						}
					}
				}

				// Handle noncurrent_version_expiration
				if v, ok := dataLoopTmp["noncurrent_version_expiration"]; ok && v != nil {
					nvExpirationSet := v.(*schema.Set)
					if nvExpirationSet.Len() > 0 {
						nvExpiration := nvExpirationSet.List()[0].(map[string]interface{})
						nvExpirationMap := make(map[string]interface{})
						if days, ok := nvExpiration["days"]; ok {
							nvExpirationMap["NoncurrentDays"] = days
						}
						if len(nvExpirationMap) > 0 {
							dataLoopMap["NoncurrentVersionExpiration"] = nvExpirationMap
						}
					}
				}

				// Handle noncurrent_version_transition
				if v, ok := dataLoopTmp["noncurrent_version_transition"]; ok && v != nil {
					nvTransitionsSet := v.(*schema.Set)
					nvTransitionMaps := make([]interface{}, 0)
					for _, nvTransition := range nvTransitionsSet.List() {
						nvTransitionMap := make(map[string]interface{})
						nvTransitionTmp := nvTransition.(map[string]interface{})

						if days, ok := nvTransitionTmp["days"]; ok {
							nvTransitionMap["NoncurrentDays"] = days
						}
						if storageClass, ok := nvTransitionTmp["storage_class"]; ok {
							nvTransitionMap["StorageClass"] = storageClass
						}
						if isAccessTime, ok := nvTransitionTmp["is_access_time"]; ok {
							nvTransitionMap["IsAccessTime"] = isAccessTime
						}
						if returnToStdWhenVisit, ok := nvTransitionTmp["return_to_std_when_visit"]; ok {
							nvTransitionMap["ReturnToStdWhenVisit"] = returnToStdWhenVisit
						}

						nvTransitionMaps = append(nvTransitionMaps, nvTransitionMap)
					}
					if len(nvTransitionMaps) > 0 {
						dataLoopMap["NoncurrentVersionTransition"] = nvTransitionMaps
					}
				}

				// Handle filter
				if v, ok := dataLoopTmp["filter"]; ok && v != nil {
					filterSet := v.(*schema.Set)
					if filterSet.Len() > 0 {
						filter := filterSet.List()[0].(map[string]interface{})
						filterMap := make(map[string]interface{})

						if prefix, ok := filter["prefix"]; ok && prefix != "" {
							filterMap["Prefix"] = prefix
						}

						if tags, ok := filter["tag"]; ok && tags != nil {
							tagsSet := tags.(*schema.Set)
							tagMaps := make([]interface{}, 0)
							for _, tag := range tagsSet.List() {
								tagMap := make(map[string]interface{})
								tagTmp := tag.(map[string]interface{})
								tagMap["Key"] = tagTmp["key"]
								tagMap["Value"] = tagTmp["value"]
								tagMaps = append(tagMaps, tagMap)
							}
							if len(tagMaps) > 0 {
								filterMap["Tag"] = tagMaps
							}
						}

						if not, ok := filter["not"]; ok && not != nil {
							notSet := not.(*schema.Set)
							if notSet.Len() > 0 {
								notFilter := notSet.List()[0].(map[string]interface{})
								notMap := make(map[string]interface{})

								if prefix, ok := notFilter["prefix"]; ok && prefix != "" {
									notMap["Prefix"] = prefix
								}

								if tag, ok := notFilter["tag"]; ok && tag != nil {
									tagSet := tag.(*schema.Set)
									if tagSet.Len() > 0 {
										tagTmp := tagSet.List()[0].(map[string]interface{})
										tagMap := make(map[string]interface{})
										tagMap["Key"] = tagTmp["key"]
										tagMap["Value"] = tagTmp["value"]
										notMap["Tag"] = tagMap
									}
								}

								if len(notMap) > 0 {
									filterMap["Not"] = notMap
								}
							}
						}

						if len(filterMap) > 0 {
							dataLoopMap["Filter"] = filterMap
						}
					}
				}

				// Handle tags
				if v, ok := dataLoopTmp["tags"]; ok && v != nil {
					tags := v.(map[string]interface{})
					tagMaps := make([]interface{}, 0)
					for k, v := range tags {
						tagMap := make(map[string]interface{})
						tagMap["Key"] = k
						tagMap["Value"] = v
						tagMaps = append(tagMaps, tagMap)
					}
					if len(tagMaps) > 0 {
						dataLoopMap["Tag"] = tagMaps
					}
				}

				localMaps = append(localMaps, dataLoopMap)
			}
			objectDataLocalMap["Rule"] = localMaps
		}
	}

	request["LifecycleConfiguration"] = objectDataLocalMap
	body = request

	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.Do("Oss", xmlParam("PUT", "2019-05-17", "PutBucketLifecycle", action), query, body, nil, hostMap, false)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		addDebug(action, response, request)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}

		ossService := NewOssService(client)
		stateConf := BuildStateConf([]string{}, []string{"#CHECKSET"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, ossService.OssBucketLifecycleStateRefreshFunc(d.Id(), "#Rule", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudOssBucketLifecycleRead(d, meta)
}

func resourceAliCloudOssBucketLifecycleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := fmt.Sprintf("/?lifecycle")
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	hostMap := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	hostMap["bucket"] = StringPointer(d.Id())

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("DELETE", "2019-05-17", "DeleteBucketLifecycle", action), query, nil, nil, hostMap, false)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}
