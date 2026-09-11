package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const dataSourceAliCloudCSKubernetesClusterName = "alicloud_cs_kubernetes_cluster"

func dataSourceAliCloudCSKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudCSKubernetesClusterRead,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the cluster.",
			},

			// Computed values from DescribeClusterDetail
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
			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_url": {
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
			"container_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"node_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"worker_ram_role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"duration": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"weekly_period": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

// flattenAckMaintenanceWindow converts the maintenance window into the data
// source's maintenance_window block. Exposed for offline table-driven tests.
func flattenAckMaintenanceWindow(window *aliyunAckAPI.AckMaintenanceWindow) []interface{} {
	if window == nil {
		return []interface{}{}
	}
	weeklyPeriod := make([]interface{}, 0, len(window.WeeklyPeriod))
	for _, v := range window.WeeklyPeriod {
		weeklyPeriod = append(weeklyPeriod, v)
	}
	return []interface{}{map[string]interface{}{
		"enable":        window.Enable,
		"start_time":    window.StartTime,
		"duration":      window.Duration,
		"weekly_period": weeklyPeriod,
	}}
}

func dataSourceAliCloudCSKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	cluster, err := ackService.GetAPI().DescribeClusterDetail(clusterId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, dataSourceAliCloudCSKubernetesClusterName, "DescribeClusterDetail", err)
	}

	d.SetId(cluster.ClusterId)
	d.Set("name", cluster.Name)
	d.Set("state", cluster.State)
	d.Set("cluster_spec", cluster.ClusterSpec)
	d.Set("cluster_type", cluster.ClusterType)
	d.Set("profile", cluster.Profile)
	d.Set("version", cluster.CurrentVersion)
	d.Set("region_id", cluster.RegionId)
	d.Set("zone_id", cluster.ZoneId)
	d.Set("master_url", cluster.MasterUrl)
	d.Set("vpc_id", cluster.VpcId)
	d.Set("container_cidr", cluster.ContainerCidr)
	d.Set("service_cidr", cluster.ServiceCidr)
	d.Set("subnet_cidr", cluster.SubnetCidr)
	d.Set("security_group_id", cluster.SecurityGroupId)
	d.Set("resource_group_id", cluster.ResourceGroupId)
	d.Set("deletion_protection", cluster.DeletionProtection)
	d.Set("node_count", int(cluster.Size))
	d.Set("worker_ram_role_name", cluster.WorkerRamRoleName)
	d.Set("timezone", cluster.Timezone)
	d.Set("tags", flattenAckTags(cluster.Tags))

	vswitchIds := make([]interface{}, 0, len(cluster.VswitchIds))
	for _, v := range cluster.VswitchIds {
		vswitchIds = append(vswitchIds, v)
	}
	d.Set("vswitch_ids", vswitchIds)

	d.Set("maintenance_window", flattenAckMaintenanceWindow(cluster.MaintenanceWindow))
	return nil
}
