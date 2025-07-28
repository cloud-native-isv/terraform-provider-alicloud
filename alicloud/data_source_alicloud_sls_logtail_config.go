package alicloud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudLogLogtailConfigs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudLogLogtailConfigsRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the log project.",
			},
			"logstore_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter configs by logstore name.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter logtail configs by name.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			"names": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of logtail config names.",
			},
			"ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of logtail config IDs.",
			},
			"configs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of logtail configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the logtail config.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the logtail config.",
						},
						"project_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the log project.",
						},
						"logstore_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The target logstore name.",
						},
						"input_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The input type of the logtail config.",
						},
						"output_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The output type of the logtail config.",
						},
						"log_sample": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The log sample of the logtail config.",
						},
						"create_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The creation time of the logtail config (Unix timestamp).",
						},
						"last_modify_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The last modification time of the logtail config (Unix timestamp).",
						},
						"input_detail": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The input detail configuration of the logtail config (JSON string).",
						},
						"output_detail": {
							Type:        schema.TypeList,
							Computed:    true,
							MaxItems:    1,
							Description: "The output detail configuration of the logtail config.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"project_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The output project name.",
									},
									"logstore_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The output logstore name.",
									},
									"compress_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The compress type.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudLogLogtailConfigsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Get("project_name").(string)
	logstoreName := d.Get("logstore_name").(string)
	nameRegex, hasNameRegex := d.GetOk("name_regex")

	var allConfigNames []string
	var err error

	if hasNameRegex && nameRegex.(string) != "" {
		// name_regex 模式：获取所有 config 并正则过滤
		err = resource.Retry(2*time.Minute, func() *resource.RetryError {
			var requestInfo *sls.Client
			raw, err := client.WithLogClient(func(slsClient *sls.Client) (interface{}, error) {
				requestInfo = slsClient
				configs, _, err := slsClient.ListConfig(projectName, 0, 500)
				return configs, err
			})
			if err != nil {
				if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			if debugOn() {
				addDebug("ListConfig", raw, requestInfo, map[string]string{"project": projectName})
			}
			if configs, ok := raw.([]string); ok {
				allConfigNames = configs
			} else {
				return resource.NonRetryableError(fmt.Errorf("unexpected response type from ListConfig"))
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_logtail_configs", "ListConfig", AliyunLogGoSdkERROR)
		}

		// 正则过滤
		r, err := regexp.Compile(nameRegex.(string))
		if err != nil {
			return WrapError(err)
		}
		var filteredConfigs []string
		for _, configName := range allConfigNames {
			if r.MatchString(configName) {
				filteredConfigs = append(filteredConfigs, configName)
			}
		}
		allConfigNames = filteredConfigs
	} else {
		// names 模式：遍历 names 列表，单独获取
		var configNames []string
		if v, ok := d.GetOk("names"); ok {
			for _, item := range v.([]interface{}) {
				if item != nil {
					configNames = append(configNames, item.(string))
				}
			}
		}
		if len(configNames) == 0 {
			// 没有 names，直接返回空
			d.SetId("")
			d.Set("ids", []string{})
			d.Set("names", []string{})
			d.Set("configs", []map[string]interface{}{})
			return nil
		}
		allConfigNames = configNames
	}

	var ids []string
	var names []string
	var configs []map[string]interface{}

	for _, configName := range allConfigNames {
		id := fmt.Sprintf("%s:%s", projectName, configName)

		// 获取详细信息
		var logtailConfig *sls.LogConfig
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			var requestInfo *sls.Client
			raw, err := client.WithLogClient(func(slsClient *sls.Client) (interface{}, error) {
				requestInfo = slsClient
				return slsClient.GetConfig(projectName, configName)
			})
			if err != nil {
				if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
					return resource.RetryableError(err)
				}
				if IsExpectedErrors(err, []string{"ConfigNotExist"}) {
					return resource.NonRetryableError(WrapErrorf(err, NotFoundMsg, AliyunLogGoSdkERROR))
				}
				return resource.NonRetryableError(err)
			}
			if debugOn() {
				addDebug("GetConfig", raw, requestInfo, map[string]string{"project": projectName, "config": configName})
			}
			if config, ok := raw.(*sls.LogConfig); ok {
				logtailConfig = config
			} else {
				return resource.NonRetryableError(fmt.Errorf("unexpected response type from GetConfig"))
			}
			return nil
		})

		if err != nil {
			if IsNotFoundError(err) {
				continue
			}
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_logtail_configs", "GetConfig", AliyunLogGoSdkERROR)
		}

		if logstoreName != "" && logtailConfig.OutputDetail.LogStoreName != logstoreName {
			continue
		}

		var inputDetailStr string
		if logtailConfig.InputDetail != nil {
			if inputDetailBytes, err := json.Marshal(logtailConfig.InputDetail); err == nil {
				inputDetailStr = string(inputDetailBytes)
			}
		}

		outputDetailList := make([]map[string]interface{}, 1)
		outputDetailList[0] = map[string]interface{}{
			"project_name":  logtailConfig.OutputDetail.ProjectName,
			"logstore_name": logtailConfig.OutputDetail.LogStoreName,
			"compress_type": logtailConfig.OutputDetail.CompressType,
		}

		configMapping := map[string]interface{}{
			"id":               id,
			"name":             logtailConfig.Name,
			"project_name":     projectName,
			"logstore_name":    logtailConfig.OutputDetail.LogStoreName,
			"input_type":       logtailConfig.InputType,
			"output_type":      logtailConfig.OutputType,
			"log_sample":       logtailConfig.LogSample,
			"create_time":      int(logtailConfig.CreateTime),
			"last_modify_time": int(logtailConfig.LastModifyTime),
			"input_detail":     inputDetailStr,
			"output_detail":    outputDetailList,
		}

		ids = append(ids, id)
		names = append(names, configName)
		configs = append(configs, configMapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}
	if err := d.Set("configs", configs); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), configs)
	}

	return nil
}
