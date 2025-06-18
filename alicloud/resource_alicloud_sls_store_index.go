package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudLogStoreIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogStoreIndexCreate,
		Read:   resourceAlicloudLogStoreIndexRead,
		Update: resourceAlicloudLogStoreIndexUpdate,
		Delete: resourceAlicloudLogStoreIndexDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"full_text": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"case_sensitive": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_chinese": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"token": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				MaxItems: 1,
			},
			//field search
			"field_search": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "long",
							ValidateFunc: validation.StringInSlice([]string{"text", "long", "double", "json"}, false),
						},
						"alias": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"case_sensitive": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"include_chinese": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"token": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_analytics": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"json_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "long",
									},
									"alias": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"doc_value": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
								},
							},
						},
					},
				},
				MinItems: 1,
			},
		},
	}
}

func resourceAlicloudLogStoreIndexCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	_, fullOk := d.GetOk("full_text")
	_, fieldOk := d.GetOk("field_search")
	if !fullOk && !fieldOk {
		return WrapError(Error("At least one of the 'full_text' and 'field_search' should be specified."))
	}

	project := d.Get("project").(string)
	logstore := d.Get("logstore").(string)
	id := fmt.Sprintf("%s%s%s", project, COLON_SEPARATED, logstore)

	// First, ensure the log store exists and is available before creating index
	if err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := slsService.DescribeLogStore(project, logstore)
		if err != nil {
			if NotFoundError(err) {
				return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, logstore, "DescribeLogStore", AlibabaCloudSdkGoERROR))
			}
			return resource.NonRetryableError(WrapErrorf(err, DefaultErrorMsg, logstore, "DescribeLogStore", AlibabaCloudSdkGoERROR))
		}
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, logstore, "WaitForLogStoreAvailable", AlibabaCloudSdkGoERROR)
	}

	// Check if index already exists and handle auto-import
	var indexExists bool
	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := slsService.GetSlsLogStoreIndex(project, logstore)
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "GetIndex", AliyunLogGoSdkERROR))
			}
			if !IsExpectedErrors(err, []string{"IndexConfigNotExist", "index config doesn't exist"}) {
				return resource.NonRetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "GetIndex", AliyunLogGoSdkERROR))
			}
		} else {
			// Index already exists, mark for auto-import
			indexExists = true
		}
		return nil
	}); err != nil {
		return err
	}

	if indexExists {
		// Auto-import existing index instead of failing
		log.Printf("[INFO] Index already exists in logstore %s, importing existing resource", logstore)
		d.SetId(id)
		return resourceAlicloudLogStoreIndexRead(d, meta)
	}

	var index aliyunSlsAPI.LogStoreIndex
	if fullOk {
		index.Line = buildIndexLine(d)
	}
	if fieldOk {
		index.Keys = buildIndexKeys(d)
	}

	// Create index with retry logic that handles log store not being ready
	if err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		err := slsService.aliyunSlsAPI.CreateLogStoreIndex(project, logstore, &index)
		if err != nil {
			// Handle specific case where log store is not ready yet
			if strings.Contains(err.Error(), "LogStoreNotExist") || strings.Contains(err.Error(), "not found") {
				return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "CreateIndex", AlibabaCloudSdkGoERROR))
			}
			if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug("CreateIndex", nil)
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "CreateIndex", AliyunLogGoSdkERROR)
	}

	d.SetId(id)

	// Wait for index to be created and available using improved StateRefreshFunc
	if err := slsService.WaitForLogStoreIndexAvailable(id, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudLogStoreIndexRead(d, meta)
}

func resourceAlicloudLogStoreIndexRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	projectName := parts[0]
	logstoreName := parts[1]

	// First check if log store exists using improved error handling
	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := slsService.DescribeLogStore(projectName, logstoreName)
		if err != nil {
			if NotFoundError(err) {
				// Log store doesn't exist, remove index from state
				d.SetId("")
				return nil
			}
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return WrapError(err)
	}

	// If log store doesn't exist, we've already cleared the ID
	if d.Id() == "" {
		return nil
	}

	// Now check the index using improved error handling and state refresh
	var index *aliyunSlsAPI.LogStoreIndex
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		index, err = slsService.GetSlsLogStoreIndex(projectName, logstoreName)
		if err != nil {
			// Handle IndexConfigNotExist and similar errors
			if strings.Contains(err.Error(), "IndexConfigNotExist") ||
				strings.Contains(err.Error(), "not found") ||
				NotFoundError(err) {
				// Index doesn't exist, remove from state
				d.SetId("")
				return nil
			}

			// Handle timeout and retryable errors
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}

			// Handle other retryable errors that might indicate the service is still being set up
			if strings.Contains(err.Error(), "InternalServerError") ||
				strings.Contains(err.Error(), "ServiceUnavailable") {
				time.Sleep(10 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapError(err)
	}

	// If index doesn't exist, we've already cleared the ID
	if d.Id() == "" || index == nil {
		return nil
	}

	// Set full text index configuration using struct fields
	if index.Line != nil {
		mapping := map[string]interface{}{
			"case_sensitive":  index.Line.CaseSensitive,
			"include_chinese": false, // The new API doesn't have Chn field, setting default
		}
		if index.Line.Token != nil && len(index.Line.Token) > 0 {
			tokenStr := ""
			for _, token := range index.Line.Token {
				tokenStr += token
			}
			mapping["token"] = tokenStr
		}
		if err := d.Set("full_text", []map[string]interface{}{mapping}); err != nil {
			return WrapError(err)
		}
	}

	// Set field search index configuration using struct fields
	if index.Keys != nil && len(index.Keys) > 0 {
		var keySet []map[string]interface{}
		for keyName, indexKey := range index.Keys {
			mapping := map[string]interface{}{
				"name":             keyName,
				"type":             indexKey.Type,
				"alias":            indexKey.Alias,
				"case_sensitive":   indexKey.CaseSensitive,
				"include_chinese":  false, // The new API doesn't have Chn field, setting default
				"enable_analytics": indexKey.DocValue,
			}

			// Handle token field
			if indexKey.Token != nil && len(indexKey.Token) > 0 {
				tokenStr := ""
				for _, token := range indexKey.Token {
					tokenStr += token
				}
				mapping["token"] = tokenStr
			}

			// Handle JSON keys for nested fields
			if indexKey.JsonKeys != nil && len(indexKey.JsonKeys) > 0 {
				var jsonKeyResults []map[string]interface{}
				for jsonKeyName, jsonIndexKey := range indexKey.JsonKeys {
					jsonKeyMapping := map[string]interface{}{
						"name":      jsonKeyName,
						"type":      jsonIndexKey.Type,
						"alias":     jsonIndexKey.Alias,
						"doc_value": jsonIndexKey.DocValue,
					}
					jsonKeyResults = append(jsonKeyResults, jsonKeyMapping)
				}
				mapping["json_keys"] = jsonKeyResults
			}
			keySet = append(keySet, mapping)
		}
		if err := d.Set("field_search", keySet); err != nil {
			return WrapError(err)
		}
	}

	d.Set("project", parts[0])
	d.Set("logstore", parts[1])
	return nil
}

func resourceAlicloudLogStoreIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	project := parts[0]
	logstore := parts[1]

	// First ensure the log store still exists before updating index
	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := slsService.DescribeLogStore(project, logstore)
		if err != nil {
			if NotFoundError(err) {
				// Log store doesn't exist, remove index from state
				d.SetId("")
				return nil
			}
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return WrapError(err)
	}

	// If log store doesn't exist, we've already cleared the ID
	if d.Id() == "" {
		return nil
	}

	// Get current index using structured data with improved error handling
	var currentIndex *aliyunSlsAPI.LogStoreIndex
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		currentIndex, err = slsService.GetSlsLogStoreIndex(project, logstore)
		if err != nil {
			if strings.Contains(err.Error(), "IndexConfigNotExist") ||
				strings.Contains(err.Error(), "not found") ||
				NotFoundError(err) {
				// Index doesn't exist, remove from state
				d.SetId("")
				return nil
			}
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapError(err)
	}

	// If index doesn't exist, we've already cleared the ID
	if d.Id() == "" {
		return nil
	}

	var index aliyunSlsAPI.LogStoreIndex
	// Copy existing configuration to preserve unchanged parts
	if currentIndex != nil {
		index = *currentIndex
	}

	update := false
	if d.HasChange("full_text") {
		index.Line = buildIndexLine(d)
		update = true
	}
	if d.HasChange("field_search") {
		index.Keys = buildIndexKeys(d)
		update = true
	}

	if update {
		// Update index with retry logic
		if err := resource.Retry(3*time.Minute, func() *resource.RetryError {
			err := slsService.aliyunSlsAPI.UpdateLogStoreIndex(project, logstore, &index)
			if err != nil {
				// Handle specific errors that might indicate the service is still being set up
				if strings.Contains(err.Error(), "LogStoreNotExist") || strings.Contains(err.Error(), "not found") {
					return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "UpdateIndex", AlibabaCloudSdkGoERROR))
				}
				if IsExpectedErrors(err, []string{LogClientTimeout, "InternalServerError"}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			if debugOn() {
				addDebug("UpdateIndex", map[string]interface{}{
					"project":  parts[0],
					"logstore": parts[1],
					"index":    index,
				})
			}
			return nil
		}); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateIndex", AliyunLogGoSdkERROR)
		}

		// Wait for index update to complete using improved StateRefreshFunc
		if err := slsService.WaitForLogStoreIndexAvailable(d.Id(), int(d.Timeout(schema.TimeoutUpdate).Seconds())); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAlicloudLogStoreIndexRead(d, meta)
}

func resourceAlicloudLogStoreIndexDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	project := parts[0]
	logstore := parts[1]

	// Check if index exists
	_, err = slsService.GetSlsLogStoreIndex(project, logstore)
	if err != nil {
		if IsExpectedErrors(err, []string{"IndexConfigNotExist"}) {
			return nil
		}
		return WrapError(err)
	}

	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		err := slsService.aliyunSlsAPI.DeleteLogStoreIndex(project, logstore)
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if debugOn() {
			addDebug("DeleteIndex", map[string]interface{}{
				"project":  parts[0],
				"logstore": parts[1],
			})
		}
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteIndex", AliyunLogGoSdkERROR)
	}

	// Wait for index to be completely deleted using improved StateRefreshFunc
	if err := slsService.WaitForLogStoreIndexDeleted(d.Id(), int(d.Timeout(schema.TimeoutDelete).Seconds())); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

func buildIndexLine(d *schema.ResourceData) *aliyunSlsAPI.IndexLine {
	if fullText, ok := d.GetOk("full_text"); ok {
		value := fullText.(*schema.Set).List()[0].(map[string]interface{})
		return &aliyunSlsAPI.IndexLine{
			CaseSensitive: value["case_sensitive"].(bool),
			Token:         strings.Split(value["token"].(string), ""),
		}
	}
	return nil
}

func buildIndexKeys(d *schema.ResourceData) map[string]*aliyunSlsAPI.IndexKey {
	keys := make(map[string]*aliyunSlsAPI.IndexKey)
	if field, ok := d.GetOk("field_search"); ok {
		for _, f := range field.(*schema.Set).List() {
			v := f.(map[string]interface{})
			indexKey := &aliyunSlsAPI.IndexKey{
				Type:          v["type"].(string),
				Alias:         v["alias"].(string),
				DocValue:      v["enable_analytics"].(bool),
				Token:         strings.Split(v["token"].(string), ""),
				CaseSensitive: v["case_sensitive"].(bool),
				JsonKeys:      map[string]*aliyunSlsAPI.IndexKey{},
			}
			jsonKeys := v["json_keys"].(*schema.Set).List()
			for _, e := range jsonKeys {
				value := e.(map[string]interface{})
				name := value["name"].(string)
				alias := value["alias"].(string)
				keyType := value["type"].(string)
				docValue := value["doc_value"].(bool)
				indexKey.JsonKeys[name] = &aliyunSlsAPI.IndexKey{
					Type:     keyType,
					Alias:    alias,
					DocValue: docValue,
				}
			}
			keys[v["name"].(string)] = indexKey
		}
	}
	return keys
}
