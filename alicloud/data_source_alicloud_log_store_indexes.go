package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAlicloudLogStoreIndexes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudLogStoreIndexesRead,
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
			"ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_text_len": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"log_reduce": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"log_reduce_black_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"log_reduce_white_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"line": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"case_sensitive": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"exclude_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"keys": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alias": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"case_sensitive": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"token": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"doc_value": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudLogStoreIndexesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Get("project").(string)
	logstore := d.Get("logstore").(string)

	// Get logstore index configuration
	index, err := slsService.GetSlsLogStoreIndex(project, logstore)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_store_indexes", "GetSlsLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using project and logstore
	d.SetId(resource.PrefixedUniqueId(project + ":" + logstore + ":"))

	// Set basic index properties
	d.Set("project", project)
	d.Set("logstore", logstore)
	d.Set("ttl", index.TTL)
	d.Set("last_modify_time", formatUnixTimestamp(index.LastModifyTime))
	d.Set("max_text_len", index.MaxTextLen)
	d.Set("log_reduce", index.LogReduce)

	// Set log reduce lists
	if index.LogReduceBlackList != nil {
		d.Set("log_reduce_black_list", index.LogReduceBlackList)
	}
	if index.LogReduceWhiteList != nil {
		d.Set("log_reduce_white_list", index.LogReduceWhiteList)
	}

	// Set line index configuration
	if index.Line != nil {
		lineConfig := make(map[string]interface{})
		lineConfig["token"] = index.Line.Token
		lineConfig["case_sensitive"] = index.Line.CaseSensitive
		if index.Line.IncludeKeys != nil {
			lineConfig["include_keys"] = index.Line.IncludeKeys
		}
		if index.Line.ExcludeKeys != nil {
			lineConfig["exclude_keys"] = index.Line.ExcludeKeys
		}
		d.Set("line", []interface{}{lineConfig})
	}

	// Set field index configurations
	if index.Keys != nil {
		keysMap := make(map[string]interface{})
		for keyName, keyConfig := range index.Keys {
			keyMap := make(map[string]interface{})
			keyMap["type"] = keyConfig.Type
			keyMap["alias"] = keyConfig.Alias
			keyMap["case_sensitive"] = keyConfig.CaseSensitive
			keyMap["doc_value"] = keyConfig.DocValue
			if keyConfig.Token != nil {
				keyMap["token"] = keyConfig.Token
			}
			keysMap[keyName] = keyMap
		}
		d.Set("keys", keysMap)
	}

	return nil
}

// formatUnixTimestamp converts Unix timestamp to readable format
func formatUnixTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
