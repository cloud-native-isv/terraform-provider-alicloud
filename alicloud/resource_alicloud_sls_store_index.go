package alicloud

import (
	"fmt"
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

	// Check if index already exists
	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := slsService.DescribeSlsLogStoreIndex(id)
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "GetIndex", AliyunLogGoSdkERROR))
			}
			if !IsExpectedErrors(err, []string{"IndexConfigNotExist"}) {
				return resource.NonRetryableError(WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store", "GetIndex", AliyunLogGoSdkERROR))
			}
		} else {
			return resource.NonRetryableError(WrapError(Error("There is already existing an index in the store %s. Please import it using id '%s%s%s'.",
				logstore, project, COLON_SEPARATED, logstore)))
		}
		return nil
	}); err != nil {
		return err
	}

	var index aliyunSlsAPI.LogStoreIndex
	if fullOk {
		index.Line = buildIndexLine(d)
	}
	if fieldOk {
		index.Keys = buildIndexKeys(d)
	}

	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		err := slsService.aliyunSlsAPI.CreateLogStoreIndex(project, logstore, &index)
		if err != nil {
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

	// Wait for index to be created and available using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.LogStoreIndexStateRefreshFunc(id, "$.project", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
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

	var index *aliyunSlsAPI.LogStoreIndex
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		index, err = slsService.DescribeSlsLogStoreIndex(d.Id())
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			if IsExpectedErrors(err, []string{"IndexConfigNotExist"}) {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapError(err)
	}

	if index == nil {
		d.SetId("")
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

	// Get current index using structured data
	currentIndex, err := slsService.DescribeSlsLogStoreIndex(d.Id())
	if err != nil {
		return WrapError(err)
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
		if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			err := slsService.aliyunSlsAPI.UpdateLogStoreIndex(project, logstore, &index)
			if err != nil {
				if IsExpectedErrors(err, []string{LogClientTimeout}) {
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

		// Wait for index update to complete using StateRefreshFunc
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, slsService.LogStoreIndexStateRefreshFunc(d.Id(), "$.project", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
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
	_, err = slsService.DescribeSlsLogStoreIndex(d.Id())
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

	// Wait for index to be completely deleted using StateRefreshFunc
	stateConf := BuildStateConf([]string{"Available"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, slsService.LogStoreIndexStateRefreshFunc(d.Id(), "$.project", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
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
