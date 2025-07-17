package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudOtsInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudOtsInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				ForceNew: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"tags": tagsSchema(),
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the Tablestore instance.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Tablestore instance.",
						},
						"alias_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The alias name of the Tablestore instance.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the Tablestore instance.",
						},
						"cluster_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster type of the Tablestore instance.",
						},
						"storage_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The storage type of the Tablestore instance.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the Tablestore instance.",
						},
						"network": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network type of the Tablestore instance.",
						},
						"network_type_acl": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The network type ACL of the Tablestore instance.",
						},
						"network_source_acl": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The network source ACL of the Tablestore instance.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID of the Tablestore instance.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID of the Tablestore instance.",
						},
						"payment_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The payment type of the Tablestore instance.",
						},
						"is_multi_az": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the Tablestore instance is multi-AZ.",
						},
						"table_quota": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The table quota of the Tablestore instance.",
						},
						"vcu_quota": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The VCU quota of the Tablestore instance.",
						},
						"elastic_vcu_upper_limit": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "The elastic VCU upper limit of the Tablestore instance.",
						},
						"policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy of the Tablestore instance.",
						},
						"policy_version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The policy version of the Tablestore instance.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time of the Tablestore instance.",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user ID of the Tablestore instance.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The tags of the Tablestore instance.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudOtsInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all instances using the new ListOtsInstance method
	allInstances, err := otsService.ListOtsInstance()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_instances", "ListOtsInstance", AlibabaCloudSdkGoERROR)
	}

	// Apply filters
	var filteredInstances []tablestoreAPI.TablestoreInstance

	// Prepare IDs filter
	idsMap := make(map[string]bool)
	if v, ok := d.GetOk("ids"); ok && len(v.([]interface{})) > 0 {
		for _, x := range v.([]interface{}) {
			if x == nil {
				continue
			}
			idsMap[x.(string)] = true
		}
	}

	// Prepare name regex filter
	var nameReg *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok && v.(string) != "" {
		nameReg = regexp.MustCompile(v.(string))
	}

	// Prepare tags filter
	var tagsFilter map[string]interface{}
	if v, ok := d.GetOk("tags"); ok {
		if vmap, ok := v.(map[string]interface{}); ok && len(vmap) > 0 {
			tagsFilter = vmap
		}
	}

	// Apply all filters
	for _, instance := range allInstances {
		// Apply IDs filter
		if len(idsMap) > 0 {
			if _, ok := idsMap[instance.InstanceName]; !ok {
				continue
			}
		}

		// Apply name regex filter
		if nameReg != nil && !nameReg.MatchString(instance.InstanceName) {
			continue
		}

		// Apply tags filter
		if tagsFilter != nil {
			instanceTagsMap := convertTablestoreInstanceTagsToMap(instance.Tags)
			if !otsInstanceTagsMapEqual(tagsFilter, instanceTagsMap) {
				continue
			}
		}

		filteredInstances = append(filteredInstances, instance)
	}

	return otsInstancesDescriptionAttributes(d, filteredInstances, meta)
}

func otsInstancesDescriptionAttributes(d *schema.ResourceData, instances []tablestoreAPI.TablestoreInstance, meta interface{}) error {
	var ids []string
	var names []string
	var s []map[string]interface{}

	for _, instance := range instances {
		mapping := map[string]interface{}{
			"id":                      instance.InstanceName,
			"name":                    instance.InstanceName,
			"alias_name":              instance.AliasName,
			"description":             instance.InstanceDescription,
			"status":                  instance.InstanceStatus,
			"network":                 instance.Network,
			"network_type_acl":        convertStringSliceToSet(instance.NetworkTypeACL),
			"network_source_acl":      convertStringSliceToSet(instance.NetworkSourceACL),
			"region_id":               instance.RegionId,
			"resource_group_id":       instance.ResourceGroupId,
			"payment_type":            instance.PaymentType,
			"is_multi_az":             instance.IsMultiAZ,
			"table_quota":             instance.TableQuota,
			"vcu_quota":               instance.VCUQuota,
			"elastic_vcu_upper_limit": instance.ElasticVCUUpperLimit,
			"policy":                  instance.Policy,
			"policy_version":          instance.PolicyVersion,
			"create_time":             instance.CreateTime.Format("2006-01-02T15:04:05Z"),
			"user_id":                 instance.UserId,
			"tags":                    convertTablestoreInstanceTagsToMap(instance.Tags),
		}

		names = append(names, instance.InstanceName)
		ids = append(ids, instance.InstanceName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("instances", s); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	// Write to output file if specified
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}

// Helper function to compare tags maps for OTS instances
func otsInstanceTagsMapEqual(expected map[string]interface{}, actual map[string]interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}

	for key, expectedValue := range expected {
		if actualValue, ok := actual[key]; !ok || actualValue != expectedValue {
			return false
		}
	}

	return true
}
