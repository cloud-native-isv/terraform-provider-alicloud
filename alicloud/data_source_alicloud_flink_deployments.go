package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAlicloudFlinkDeployments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkDeploymentsRead,
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
			"deployments": {
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

func dataSourceAlicloudFlinkDeploymentsRead(d *schema.ResourceData, meta interface{}) error {
	// 1. 初始化Flink服务客户端
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// 2. 获取所有Flink实例（分页处理）
	pagination := &aliyunFlinkAPI.PaginationRequest{
		PageIndex: 1,
		PageSize:  50,
	}
	response, err := flinkService.ListInstances(pagination)
	if err != nil {
		return WrapError(err)
	}

	// 3. 过滤和映射结果
	var deployments []map[string]interface{}
	ids := make([]string, 0)
	names := make([]string, 0)

	if response != nil {
		for _, instance := range response.Data {
			instanceId := instance.ID
			instanceName := instance.Name
			status := instance.Status

			deployments = append(deployments, map[string]interface{}{
				"id":     instanceId,
				"name":   instanceName,
				"status": status,
				// Add other fields here
			})
			ids = append(ids, instanceId)
			names = append(names, instanceName)
		}
	}

	// 4. 设置数据源返回值
	d.SetId("flink_deployments")
	if err := d.Set("ids", ids); err != nil {
		return err
	}
	if err := d.Set("names", names); err != nil {
		return err
	}
	if err := d.Set("deployments", deployments); err != nil {
		return err
	}

	return nil
}
