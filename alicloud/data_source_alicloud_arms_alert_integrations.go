package alicloud

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudArmsAlertIntegrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudArmsIntegrationsRead,
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

func dataSourceAliCloudArmsIntegrationsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService := ArmsService{client: client}

	var filteredIntegrations []*aliyunArmsAPI.AlertIntegration
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

	// Get integrations from service layer using strong types
	integrations, err := armsService.ListArmsIntegrations()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_integrations", "ListArmsIntegrations", AlibabaCloudSdkGoERROR)
	}

	// Process and filter integrations using strong types
	for _, integration := range integrations {
		// Apply name regex filter
		if integrationNameRegex != nil && !integrationNameRegex.MatchString(integration.IntegrationName) {
			continue
		}

		// Apply IDs filter
		if len(idsMap) > 0 {
			integrationIdStr := strconv.FormatInt(integration.IntegrationId, 10)
			if _, ok := idsMap[integrationIdStr]; !ok {
				continue
			}
		}

		// Apply integration type filter
		if v, ok := d.GetOk("integration_type"); ok {
			if integration.IntegrationProductType != v.(string) {
				continue
			}
		}

		// Apply status filter
		if v, ok := d.GetOk("status"); ok {
			expectedStatus := v.(string)
			var currentStatus string
			if integration.State {
				currentStatus = "Active"
			} else {
				currentStatus = "Inactive"
			}
			if currentStatus != expectedStatus {
				continue
			}
		}

		filteredIntegrations = append(filteredIntegrations, integration)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, integration := range filteredIntegrations {
		integrationIdStr := strconv.FormatInt(integration.IntegrationId, 10)

		mapping := map[string]interface{}{
			"id":               integrationIdStr,
			"integration_id":   integrationIdStr,
			"integration_name": integration.IntegrationName,
			"integration_type": integration.IntegrationProductType,
			"status":           getIntegrationStatusFromBool(integration.State),
			"create_time":      formatTimeFromPtr(integration.CreateTime),
		}

		if integration.Description != "" {
			mapping["description"] = integration.Description
		}

		// Set config field if ApiEndpoint exists
		if integration.ApiEndpoint != "" {
			mapping["config"] = fmt.Sprintf(`{"apiEndpoint":"%s"}`, integration.ApiEndpoint)
		}

		// For update_time, use UpdateTime if available, otherwise use create_time
		if integration.UpdateTime != nil {
			mapping["update_time"] = formatTimeFromPtr(integration.UpdateTime)
		} else {
			mapping["update_time"] = mapping["create_time"]
		}

		ids = append(ids, integrationIdStr)
		names = append(names, integration.IntegrationName)
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

// Helper function to convert boolean state to status string
func getIntegrationStatusFromBool(state bool) string {
	if state {
		return "Active"
	}
	return "Inactive"
}

// Helper function to format time pointer to string
func formatTimeFromPtr(timePtr *time.Time) string {
	if timePtr == nil {
		return ""
	}
	return timePtr.Format(time.RFC3339)
}
