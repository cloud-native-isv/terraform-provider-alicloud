package alicloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
						// Add other fields as needed
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

	// 2. 获取所有Flink实例（分页处理）
	region := client.RegionId // 默认使用客户端的区域
	instances, err := flinkService.ListInstances(region)
	if err != nil {
		return WrapError(err)
	}

	// 3. 过滤和映射结果
	var workspaces []map[string]interface{}
	ids := make([]string, 0)
	names := make([]string, 0)
	for _, instance := range instances {
		if instance == nil || instance.InstanceId == nil || instance.InstanceName == nil || instance.ClusterStatus == nil {
			continue
		}
		instanceId := *instance.InstanceId
		instanceName := *instance.InstanceName
		status := *instance.ClusterStatus

		workspaces = append(workspaces, map[string]interface{}{
			"id":      instanceId,
			"name":    instanceName,
			"status":  status,
			// Add other fields here
		})
		ids = append(ids, instanceId)
		names = append(names, instanceName)
	}

	// 4. 设置数据源返回值
	d.SetId("flink_workspaces")
	if err := d.Set("ids", ids); err != nil {
		return err
	}
	if err := d.Set("names", names); err != nil {
		return err
	}
	if err := d.Set("workspaces", workspaces); err != nil {
		return err
	}

	return nil
}