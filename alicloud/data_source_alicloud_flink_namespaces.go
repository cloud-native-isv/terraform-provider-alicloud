package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAlicloudFlinkNamespaces() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudFlinkNamespacesRead,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Flink workspace",
			},
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
			"namespaces": {
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

func dataSourceAlicloudFlinkNamespacesRead(d *schema.ResourceData, meta interface{}) error {
	// 1. 初始化Flink服务客户端
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace").(string)

	// 2. 获取所有Flink实例（分页处理）
	namespaces, err := flinkService.ListNamespaces(workspace)
	if err != nil {
		return WrapError(err)
	}

	// 3. 过滤和映射结果
	var namespaceMaps []map[string]interface{}
	ids := make([]string, 0)
	names := make([]string, 0)
	for _, namespace := range namespaces {
		if namespace == nil || namespace.Namespace == nil || namespace.ResourceSpec == nil  {
			continue
		}
		namespaceMaps = append(namespaceMaps, map[string]interface{}{
			"id":     *namespace.Namespace,
			"name":   *namespace.Namespace,
			"status": *namespace.Status,
		})
		ids = append(ids, *namespace.Namespace)
		names = append(names, *namespace.Namespace)
	}

	if err := d.Set("ids", ids); err != nil {
		return err
	}
	if err := d.Set("names", names); err != nil {
		return err
	}
	if err := d.Set("namespaces", namespaceMaps); err != nil {
		return err
	}

	return nil
}
