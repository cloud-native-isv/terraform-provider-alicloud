package alicloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var cmsScopeFiles = []string{
	"alicloud/service_alicloud_cms_addon.go",
	"alicloud/service_alicloud_cms_agg_task.go",
	"alicloud/service_alicloud_cms_alert.go",
	"alicloud/service_alicloud_cms_cloud_resource.go",
	"alicloud/service_alicloud_cms_context.go",
	"alicloud/service_alicloud_cms_context_store.go",
	"alicloud/service_alicloud_cms_dataset.go",
	"alicloud/service_alicloud_cms_delivery_task.go",
	"alicloud/service_alicloud_cms_entity_store.go",
	"alicloud/service_alicloud_cms_integration_policy.go",
	"alicloud/service_alicloud_cms_memory.go",
	"alicloud/service_alicloud_cms_memory_store.go",
	"alicloud/service_alicloud_cms_pipeline.go",
	"alicloud/service_alicloud_cms_prometheus.go",
	"alicloud/service_alicloud_cms_service.go",
	"alicloud/service_alicloud_cms_umodel.go",
	"alicloud/service_alicloud_cms_workspace.go",
	"alicloud/resource_alicloud_cms_addon.go",
	"alicloud/resource_alicloud_cms_agg_task.go",
	"alicloud/resource_alicloud_cms_alert.go",
	"alicloud/resource_alicloud_cms_cloud_resource.go",
	"alicloud/resource_alicloud_cms_context.go",
	"alicloud/resource_alicloud_cms_context_store.go",
	"alicloud/resource_alicloud_cms_dataset.go",
	"alicloud/resource_alicloud_cms_delivery_task.go",
	"alicloud/resource_alicloud_cms_entity_store.go",
	"alicloud/resource_alicloud_cms_integration_policy.go",
	"alicloud/resource_alicloud_cms_memory.go",
	"alicloud/resource_alicloud_cms_memory_store.go",
	"alicloud/resource_alicloud_cms_pipeline.go",
	"alicloud/resource_alicloud_cms_prometheus.go",
	"alicloud/resource_alicloud_cms_service.go",
	"alicloud/resource_alicloud_cms_umodel.go",
	"alicloud/resource_alicloud_cms_workspace.go",
	"alicloud/data_source_alicloud_cms_addon.go",
	"alicloud/data_source_alicloud_cms_agg_task.go",
	"alicloud/data_source_alicloud_cms_alert.go",
	"alicloud/data_source_alicloud_cms_cloud_resource.go",
	"alicloud/data_source_alicloud_cms_context.go",
	"alicloud/data_source_alicloud_cms_context_store.go",
	"alicloud/data_source_alicloud_cms_dataset.go",
	"alicloud/data_source_alicloud_cms_delivery_task.go",
	"alicloud/data_source_alicloud_cms_entity_store.go",
	"alicloud/data_source_alicloud_cms_integration_policy.go",
	"alicloud/data_source_alicloud_cms_memory.go",
	"alicloud/data_source_alicloud_cms_memory_store.go",
	"alicloud/data_source_alicloud_cms_pipeline.go",
	"alicloud/data_source_alicloud_cms_prometheus.go",
	"alicloud/data_source_alicloud_cms_service.go",
	"alicloud/data_source_alicloud_cms_umodel.go",
	"alicloud/data_source_alicloud_cms_workspace.go",
}

func TestCmsScopeFileCount(t *testing.T) {
	if len(cmsScopeFiles) != 51 {
		t.Fatalf("expected 51 cms scope files, got %d", len(cmsScopeFiles))
	}
}

func TestCmsResourceAndDataSourceNoDirectRpcOrSdkCalls(t *testing.T) {
	for _, file := range cmsScopeFiles {
		name := filepath.Base(file)
		if !strings.HasPrefix(name, "resource_") && !strings.HasPrefix(name, "data_source_") {
			continue
		}
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		text := string(content)
		if strings.Contains(text, "RpcPost(") || strings.Contains(text, "WithCmsClient(") {
			t.Fatalf("expected %s not to use direct rpc/sdk client", name)
		}
	}
}
