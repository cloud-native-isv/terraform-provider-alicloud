package alicloud

import (
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudLogStores() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudLogStoresRead,
		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"stores": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"store_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"shard_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"last_modify_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enable_tracking": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"auto_split": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_split_shard": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"append_meta": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"hot_ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"infrequent_access_ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"encrypt_conf": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"encrypt_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_cmk_info": {
										Type:     schema.TypeList,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cmk_key_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"arn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"region_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"product_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"processor_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"telemetry_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudLogStoresRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	var logStoreNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		logStoreNameRegex = r
	}

	project := d.Get("project").(string)
	var stores []*aliyunSlsAPI.LogStore

	// Check if specific store names are provided
	var storeNames []string
	if v, ok := d.GetOk("ids"); ok {
		for _, item := range v.([]interface{}) {
			if item != nil {
				storeNames = append(storeNames, item.(string))
			}
		}
	}

	if len(storeNames) > 0 {
		// Get specific stores by names
		for _, storeName := range storeNames {
			err = resource.Retry(2*time.Minute, func() *resource.RetryError {
				store, err := slsService.GetLogStore(project, storeName)
				if err != nil {
					if IsExpectedErrors(err, []string{LogClientTimeout}) {
						time.Sleep(5 * time.Second)
						return resource.RetryableError(err)
					}
					if NotFoundError(err) {
						return resource.NonRetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				stores = append(stores, store)
				return nil
			})
			if err != nil {
				if NotFoundError(err) {
					continue
				}
				return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_stores", "GetLogStore", AliyunLogGoSdkERROR)
			}
		}
	} else {
		// List all stores
		err = resource.Retry(2*time.Minute, func() *resource.RetryError {
			response, err := slsService.ListLogStores(project, "", "", "")
			if err != nil {
				if IsExpectedErrors(err, []string{LogClientTimeout}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				} else if NotFoundError(err) {
					return nil
				} else {
					return resource.NonRetryableError(err)
				}
			}
			stores = response
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_stores", "ListLogStore", AliyunLogGoSdkERROR)
		}
	}

	addDebug("LogStores", stores)

	// Filter stores based on name regex
	var filteredStores []*aliyunSlsAPI.LogStore
	for _, store := range stores {
		if logStoreNameRegex != nil {
			if !logStoreNameRegex.MatchString(store.LogstoreName) {
				continue
			}
		}
		filteredStores = append(filteredStores, store)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)
	for _, store := range filteredStores {
		mapping := map[string]interface{}{
			"id":                    store.LogstoreName,
			"store_name":            store.LogstoreName,
			"project_name":          store.ProjectName,
			"ttl":                   store.Ttl,
			"shard_count":           store.ShardCount,
			"create_time":           store.CreateTime,
			"last_modify_time":      store.LastModifyTime,
			"enable_tracking":       store.EnableTracking,
			"auto_split":            store.AutoSplit,
			"max_split_shard":       store.MaxSplitShard,
			"append_meta":           store.AppendMeta,
			"hot_ttl":               store.HotTtl,
			"infrequent_access_ttl": store.InfrequentAccessTTL,
			"mode":                  store.Mode,
			"product_type":          store.ProductType,
			"processor_id":          store.ProcessorId,
			"telemetry_type":        store.TelemetryType,
		}

		// Handle encryption configuration
		if store.EncryptConf != nil {
			encryptConf := map[string]interface{}{
				"enable":       store.EncryptConf.Enable,
				"encrypt_type": store.EncryptConf.EncryptType,
			}

			if store.EncryptConf.UserCmkInfo != nil {
				userCmkInfo := map[string]interface{}{
					"cmk_key_id": store.EncryptConf.UserCmkInfo.CmkKeyId,
					"arn":        store.EncryptConf.UserCmkInfo.Arn,
					"region_id":  store.EncryptConf.UserCmkInfo.RegionId,
				}
				encryptConf["user_cmk_info"] = []map[string]interface{}{userCmkInfo}
			}
			mapping["encrypt_conf"] = []map[string]interface{}{encryptConf}
		}

		ids = append(ids, store.LogstoreName)
		names = append(names, store.LogstoreName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("stores", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
