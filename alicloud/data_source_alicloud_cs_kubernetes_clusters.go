package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const dataSourceAliCloudCSKubernetesClustersName = "alicloud_cs_kubernetes_clusters"

func dataSourceAliCloudCSKubernetesClusters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudCSKubernetesClustersRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A client-side regular expression used to filter results by cluster name.",
			},
			"cluster_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server-side filter: the cluster type, e.g. `Kubernetes`.",
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server-side filter: the cluster profile, e.g. `Managed` or `Default`.",
			},
			"enable_details": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to enrich every listed cluster with DescribeClusterDetail. Default to false.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Write the result to this file (as JSON).",
			},

			// Computed values
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_spec": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"profile": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vswitch_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"resource_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

// buildAckClusterQuery maps the data source schema onto the server-side
// DescribeClustersV1 filter. Exposed for offline table-driven tests.
func buildAckClusterQuery(d *schema.ResourceData) *aliyunAckAPI.AckClusterQuery {
	query := &aliyunAckAPI.AckClusterQuery{}
	if clusterType, ok := d.GetOk("cluster_type"); ok {
		query.ClusterType = clusterType.(string)
	}
	if profile, ok := d.GetOk("profile"); ok {
		query.Profile = profile.(string)
	}
	return query
}

// filterAckClustersByNameRegex applies the client-side name_regex filter.
// Exposed for offline table-driven tests.
func filterAckClustersByNameRegex(clusters []aliyunAckAPI.AckCluster, nameRegex string) ([]aliyunAckAPI.AckCluster, error) {
	if nameRegex == "" {
		return clusters, nil
	}
	pattern, err := regexp.Compile(nameRegex)
	if err != nil {
		return nil, err
	}
	result := make([]aliyunAckAPI.AckCluster, 0)
	for _, cluster := range clusters {
		if pattern.MatchString(cluster.Name) {
			result = append(result, cluster)
		}
	}
	return result, nil
}

// flattenAckClusters converts the API cluster list into the data source's
// clusters attribute. Exposed for offline table-driven tests.
func flattenAckClusters(clusters []aliyunAckAPI.AckCluster) []interface{} {
	result := make([]interface{}, 0, len(clusters))
	for _, cluster := range clusters {
		vswitchIds := make([]interface{}, 0, len(cluster.VswitchIds))
		for _, v := range cluster.VswitchIds {
			vswitchIds = append(vswitchIds, v)
		}
		result = append(result, map[string]interface{}{
			"cluster_id":        cluster.ClusterId,
			"name":              cluster.Name,
			"state":             cluster.State,
			"cluster_spec":      cluster.ClusterSpec,
			"cluster_type":      cluster.ClusterType,
			"profile":           cluster.Profile,
			"version":           cluster.CurrentVersion,
			"vpc_id":            cluster.VpcId,
			"vswitch_ids":       vswitchIds,
			"resource_group_id": cluster.ResourceGroupId,
			"tags":              flattenAckTags(cluster.Tags),
		})
	}
	return result
}

func dataSourceAliCloudCSKubernetesClustersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	// DescribeClustersV1 is the only supported list API: the legacy
	// DescribeClusters flavor is rejected by the current permission model.
	clusters, err := ackService.GetAPI().DescribeClustersV1(buildAckClusterQuery(d))
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, dataSourceAliCloudCSKubernetesClustersName, "DescribeClustersV1", err)
	}

	nameRegex, _ := d.Get("name_regex").(string)
	clusters, err = filterAckClustersByNameRegex(clusters, nameRegex)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, dataSourceAliCloudCSKubernetesClustersName, "name_regex", err)
	}

	if enableDetails := d.Get("enable_details").(bool); enableDetails {
		detailed := make([]aliyunAckAPI.AckCluster, 0, len(clusters))
		for _, cluster := range clusters {
			detail, err := ackService.GetAPI().DescribeClusterDetail(cluster.ClusterId)
			if err != nil {
				if NotFoundError(err) {
					continue
				}
				return WrapErrorf(err, DataDefaultErrorMsg, dataSourceAliCloudCSKubernetesClustersName, "DescribeClusterDetail", err)
			}
			detailed = append(detailed, *detail)
		}
		clusters = detailed
	}

	ids := make([]interface{}, 0, len(clusters))
	names := make([]interface{}, 0, len(clusters))
	for _, cluster := range clusters {
		ids = append(ids, cluster.ClusterId)
		names = append(names, cluster.Name)
	}
	d.Set("ids", ids)
	d.Set("names", names)
	d.Set("clusters", flattenAckClusters(clusters))
	d.SetId(dataResourceIdHash([]string{nameRegex}))

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), flattenAckClusters(clusters))
	}
	return nil
}
