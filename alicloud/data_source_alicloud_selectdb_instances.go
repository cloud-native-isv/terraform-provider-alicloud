package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudSelectDBInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudSelectDBInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"tags": tagsSchema(),
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_minor_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"payment_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cpu_prepaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_prepaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cache_size_prepaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_count_prepaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cpu_postpaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_postpaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cache_size_postpaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_count_postpaid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vswitch_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sub_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gmt_created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gmt_modified": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gmt_expired": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lock_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lock_reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudSelectDBInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get filter parameters
	var idsFilter []string
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv != nil {
				idsFilter = append(idsFilter, vv.(string))
			}
		}
	}

	// List all instances
	instances, err := selectDBService.DescribeSelectDBInstances(int32(1), int32(50))
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_selectdb_instances", AlibabaCloudSdkGoERROR)
	}

	var filteredInstances []selectdb.Instance
	for _, instance := range instances {
		// Apply ID filter if specified
		if len(idsFilter) > 0 {
			found := false
			for _, filterId := range idsFilter {
				if instance.Id == filterId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filteredInstances = append(filteredInstances, instance)
	}

	ids := make([]string, 0)
	s := make([]map[string]interface{}, 0)

	for _, instance := range filteredInstances {
		mapping := map[string]interface{}{
			// New field names
			"id":                   instance.Id,
			"instance_id":          instance.Id,
			"instance_name":        instance.Name,
			"engine":               instance.Engine,
			"engine_version":       instance.EngineVersion,
			"engine_minor_version": instance.EngineMinorVersion,
			"instance_description": instance.Name,
			"status":               instance.Status,
			"payment_type":         convertChargeTypeToPaymentType(instance.ChargeType),
			"region_id":            instance.RegionId,
			"zone_id":              instance.ZoneId,
			"vpc_id":               instance.VpcId,
			"vswitch_id":           instance.VswitchId,
			"sub_domain":           instance.SubDomain,
			"gmt_created":          instance.GmtCreated,
			"gmt_modified":         instance.GmtModified,
			"gmt_expired":          instance.ExpireTime,
			"lock_mode":            fmt.Sprintf("%d", instance.LockMode),
			"lock_reason":          instance.LockReason,

			// Resource configuration
			"cpu_prepaid":            0,
			"memory_prepaid":         0,
			"cache_size_prepaid":     0,
			"cluster_count_prepaid":  0,
			"cpu_postpaid":           instance.ResourceCpu,
			"memory_postpaid":        instance.ResourceMemory,
			"cache_size_postpaid":    instance.StorageSize,
			"cluster_count_postpaid": instance.ClusterCount,
		}

		ids = append(ids, instance.Id)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("instances", s); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
