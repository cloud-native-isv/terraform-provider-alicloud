package alicloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsAlertIntegrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsAlertIntegrationsRead,
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
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"integration_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"integration_product_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
						"integration_product_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"api_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"short_token": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"liveness": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auto_recover": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"recover_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"duplicate_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudArmsAlertIntegrationsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsAPI, err := arms.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_integrations", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	var objects []*arms.AlertIntegration
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

	// Get integration filters
	integrationName := ""
	if v, ok := d.GetOk("integration_name"); ok {
		integrationName = v.(string)
	}

	integrationProductType := ""
	if v, ok := d.GetOk("integration_product_type"); ok {
		integrationProductType = v.(string)
	}

	// List integrations using ARMS API
	page := int64(1)
	pageSize := int64(100)

	for {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			var integrations []*arms.AlertIntegration

			if integrationProductType != "" {
				// Filter by product type
				integrations, _, err = armsAPI.ListIntegrations(integrationProductType, page, pageSize, true, integrationName)
			} else {
				// Get all integrations
				integrations, err = armsAPI.ListAllIntegrations(page, pageSize)
			}

			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			// Filter results
			for _, integration := range integrations {
				// Apply name regex filter
				if integrationNameRegex != nil && !integrationNameRegex.MatchString(integration.IntegrationName) {
					continue
				}

				// Apply IDs filter
				integrationIdStr := fmt.Sprint(integration.IntegrationId)
				if len(idsMap) > 0 {
					if _, ok := idsMap[integrationIdStr]; !ok {
						continue
					}
				}

				// Apply integration name filter (if not already applied in API call)
				if integrationName != "" && integrationProductType == "" && integration.IntegrationName != integrationName {
					continue
				}

				objects = append(objects, integration)
			}

			// Check if we need to continue pagination
			if int64(len(integrations)) < pageSize {
				return nil
			}

			return nil
		})

		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_integrations", "ListIntegrations", AlibabaCloudSdkGoERROR)
		}

		// Continue pagination if we got a full page
		if len(objects) >= int(pageSize) {
			page++
		} else {
			break
		}
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, object := range objects {
		mapping := map[string]interface{}{
			"id":                       fmt.Sprint(object.IntegrationId),
			"integration_id":           fmt.Sprint(object.IntegrationId),
			"integration_name":         object.IntegrationName,
			"integration_product_type": object.IntegrationProductType,
			"api_endpoint":             object.ApiEndpoint,
			"short_token":              object.ShortToken,
			"state":                    object.State,
			"liveness":                 object.Liveness,
			"create_time":              object.CreateTime,
		}

		// Set integration detail fields if available
		if object.IntegrationDetail != nil {
			mapping["description"] = object.IntegrationDetail.Description
			mapping["auto_recover"] = object.IntegrationDetail.AutoRecover
			mapping["recover_time"] = object.IntegrationDetail.RecoverTime
			mapping["duplicate_key"] = object.IntegrationDetail.DuplicateKey
		} else {
			mapping["description"] = ""
			mapping["auto_recover"] = false
			mapping["recover_time"] = 0
			mapping["duplicate_key"] = ""
		}

		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, object.IntegrationName)
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
