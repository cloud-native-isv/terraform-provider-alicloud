package alicloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudArmsIntegrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudArmsIntegrationsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"integration_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Active", "Inactive"}, false),
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"integrations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudArmsIntegrationsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	action := "ListIntegrations"
	request := make(map[string]interface{})
	request["RegionId"] = client.RegionId

	if v, ok := d.GetOk("integration_type"); ok {
		request["IntegrationType"] = v
	}
	if v, ok := d.GetOk("status"); ok {
		request["Status"] = v
	}

	var objects []map[string]interface{}
	var integrationNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		integrationNameRegex = r
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	var response map[string]interface{}
	request["Page"] = 1
	request["Size"] = PageSizeXLarge

	for {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			v, err := jsonpath.Get("$.PageInfo.Integrations", response)
			if err != nil {
				return resource.NonRetryableError(WrapErrorf(err, FailedGetAttributeMsg, action, "$.PageInfo.Integrations", response))
			}
			if v != nil {
				for _, integration := range v.([]interface{}) {
					item := integration.(map[string]interface{})
					if integrationNameRegex != nil && !integrationNameRegex.MatchString(fmt.Sprint(item["IntegrationName"])) {
						continue
					}
					if len(idsMap) > 0 {
						if _, ok := idsMap[fmt.Sprint(item["IntegrationId"])]; !ok {
							continue
						}
					}
					objects = append(objects, item)
				}
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_integrations", action, AlibabaCloudSdkGoERROR)
		}

		totalCount, err := jsonpath.Get("$.PageInfo.Total", response)
		if err != nil {
			return WrapErrorf(err, FailedGetAttributeMsg, action, "$.PageInfo.Total", response)
		}

		if len(objects) >= int(totalCount.(float64)) {
			break
		}

		if page, err := jsonpath.Get("$.PageInfo.Page", response); err != nil || page == nil {
			break
		} else {
			request["Page"] = int(page.(float64)) + 1
		}
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		mapping := map[string]interface{}{
			"id":               fmt.Sprint(object["IntegrationId"]),
			"integration_id":   fmt.Sprint(object["IntegrationId"]),
			"integration_name": object["IntegrationName"],
			"integration_type": object["IntegrationType"],
			"description":      object["Description"],
			"status":           object["Status"],
			"create_time":      object["CreateTime"],
			"update_time":      object["UpdateTime"],
			"config":           object["Config"],
		}
		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, object["IntegrationName"])
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("integrations", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
