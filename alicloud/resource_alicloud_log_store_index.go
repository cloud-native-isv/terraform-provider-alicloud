package alicloud

import (
	"context"
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

	_, fullOk := d.GetOk("full_text")
	_, fieldOk := d.GetOk("field_search")
	if !fullOk && !fieldOk {
		return WrapError(Error("At least one of the 'full_text' and 'field_search' should be specified."))
	}

	project := d.Get("project").(string)
	logstore := d.Get("logstore").(string)

	// Check if index already exists
	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			return slsClient.GetIndex(ctx, project, logstore)
		})
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
		_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			return nil, slsClient.CreateIndex(ctx, project, logstore, &index)
		})
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

	d.SetId(fmt.Sprintf("%s%s%s", project, COLON_SEPARATED, logstore))

	return resourceAlicloudLogStoreIndexRead(d, meta)
}

func resourceAlicloudLogStoreIndexRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	project := parts[0]
	logstore := parts[1]

	var index *aliyunSlsAPI.LogStoreIndex
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		raw, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			return slsClient.GetIndex(ctx, project, logstore)
		})
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
		index = raw.(*aliyunSlsAPI.LogStoreIndex)
		return nil
	})

	if err != nil {
		return WrapError(err)
	}

	if index == nil {
		d.SetId("")
		return nil
	}

	if line := index.Line; line != nil {
		mapping := map[string]interface{}{
			"case_sensitive":  line.CaseSensitive,
			"include_chinese": false, // The new API doesn't have Chn field, setting default
			"token":           strings.Join(line.Token, ""),
		}
		if err := d.Set("full_text", []map[string]interface{}{mapping}); err != nil {
			return WrapError(err)
		}
	}
	if keys := index.Keys; keys != nil {
		var keySet []map[string]interface{}
		for k, v := range keys {
			mapping := map[string]interface{}{
				"name":             k,
				"type":             v.Type,
				"alias":            v.Alias,
				"case_sensitive":   v.CaseSensitive,
				"include_chinese":  false, // The new API doesn't have Chn field, setting default
				"token":            strings.Join(v.Token, ""),
				"enable_analytics": v.DocValue,
			}
			if len(v.JsonKeys) > 0 {
				var result = []map[string]interface{}{}
				for k1, v1 := range v.JsonKeys {
					var value = map[string]interface{}{}
					value["doc_value"] = v1.DocValue
					value["alias"] = v1.Alias
					value["type"] = v1.Type
					value["name"] = k1
					result = append(result, value)
				}
				mapping["json_keys"] = result
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

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	project := parts[0]
	logstore := parts[1]

	// Get current index
	var index *aliyunSlsAPI.LogStoreIndex
	_, err = client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
		ctx := context.Background()
		var getErr error
		index, getErr = slsClient.GetIndex(ctx, project, logstore)
		return index, getErr
	})
	if err != nil {
		return WrapError(err)
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
			_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
				ctx := context.Background()
				return nil, slsClient.UpdateIndex(ctx, project, logstore, index)
			})
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
	}

	return resourceAlicloudLogStoreIndexRead(d, meta)
}

func resourceAlicloudLogStoreIndexDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	project := parts[0]
	logstore := parts[1]

	// Check if index exists
	_, err = client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
		ctx := context.Background()
		return slsClient.GetIndex(ctx, project, logstore)
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"IndexConfigNotExist"}) {
			return nil
		}
		return WrapError(err)
	}

	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			return nil, slsClient.DeleteIndex(ctx, project, logstore)
		})
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
