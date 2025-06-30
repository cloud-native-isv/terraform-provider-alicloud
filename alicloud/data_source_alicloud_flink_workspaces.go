package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAlicloudFlinkWorkspaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkWorkspacesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				Computed: true,
			},
			"workspaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"architecture_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ask_cluster_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"charge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"monitor_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"order_state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uid": {
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
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cluster_state": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"resource_spec": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeFloat},
						},
						"storage": {
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

func dataSourceAlicloudFlinkWorkspacesRead(d *schema.ResourceData, meta interface{}) error {
	// 1. 初始化Flink服务客户端
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// 2. 获取所有Flink实例（直接获取，无需分页）
	workspaces, err := flinkService.ListInstances()
	if err != nil {
		return WrapError(err)
	}

	// 3. 过滤和映射结果
	var workspaceMaps []map[string]interface{}
	ids := make([]string, 0)
	names := make([]string, 0)

	for _, instance := range workspaces {
		workspace := map[string]interface{}{
			"id":                instance.ID,
			"name":              instance.Name,
			"status":            instance.Status,
			"region":            instance.Region,
			"architecture_type": instance.ArchitectureType,
			"ask_cluster_id":    instance.AskClusterID,
			"charge_type":       instance.ChargeType,
			"monitor_type":      instance.MonitorType,
			"order_state":       instance.OrderState,
			"resource_id":       instance.ResourceID,
			"resource_group_id": instance.ResourceGroupID,
			"uid":               instance.UID,
			"vpc_id":            instance.VPCID,
			"create_time":       instance.CreateTime,
		}

		// Add VSwitchIDs as a list
		if instance.VSwitchIDs != nil && len(instance.VSwitchIDs) > 0 {
			workspace["vswitch_ids"] = instance.VSwitchIDs
		}

		// Add Storage as a map
		if instance.Storage != nil {
			storage := map[string]interface{}{
				"fully_managed": fmt.Sprintf("%t", instance.Storage.FullyManaged),
			}

			if instance.Storage.Oss != nil {
				storage["oss_bucket"] = instance.Storage.Oss.Bucket
			}

			workspace["storage"] = storage
		}

		workspaceMaps = append(workspaceMaps, workspace)
		ids = append(ids, instance.ID)
		names = append(names, instance.Name)
	}

	// 4. 设置数据源返回值
	d.SetId("flink_workspaces")
	if err := d.Set("ids", ids); err != nil {
		return err
	}
	if err := d.Set("names", names); err != nil {
		return err
	}
	if err := d.Set("workspaces", workspaceMaps); err != nil {
		return err
	}

	return nil
}
