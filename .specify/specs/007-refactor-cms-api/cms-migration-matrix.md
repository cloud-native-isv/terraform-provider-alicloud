# CMS Migration Matrix

## Scope Baseline (51 Files)

| Object Group | Service File | Resource File | Data Source File | Current State | Registration | Scope Decision |
|---|---|---|---|---|---|---|
| addon | alicloud/service_alicloud_cms_addon.go | alicloud/resource_alicloud_cms_addon.go | alicloud/data_source_alicloud_cms_addon.go | placeholder | unregistered | no_change_placeholder |
| agg_task | alicloud/service_alicloud_cms_agg_task.go | alicloud/resource_alicloud_cms_agg_task.go | alicloud/data_source_alicloud_cms_agg_task.go | placeholder | unregistered | no_change_placeholder |
| alert | alicloud/service_alicloud_cms_alert.go | alicloud/resource_alicloud_cms_alert.go | alicloud/data_source_alicloud_cms_alert.go | placeholder | unregistered | no_change_placeholder |
| cloud_resource | alicloud/service_alicloud_cms_cloud_resource.go | alicloud/resource_alicloud_cms_cloud_resource.go | alicloud/data_source_alicloud_cms_cloud_resource.go | placeholder | unregistered | no_change_placeholder |
| context | alicloud/service_alicloud_cms_context.go | alicloud/resource_alicloud_cms_context.go | alicloud/data_source_alicloud_cms_context.go | placeholder | unregistered | no_change_placeholder |
| context_store | alicloud/service_alicloud_cms_context_store.go | alicloud/resource_alicloud_cms_context_store.go | alicloud/data_source_alicloud_cms_context_store.go | placeholder | unregistered | no_change_placeholder |
| dataset | alicloud/service_alicloud_cms_dataset.go | alicloud/resource_alicloud_cms_dataset.go | alicloud/data_source_alicloud_cms_dataset.go | placeholder | unregistered | no_change_placeholder |
| delivery_task | alicloud/service_alicloud_cms_delivery_task.go | alicloud/resource_alicloud_cms_delivery_task.go | alicloud/data_source_alicloud_cms_delivery_task.go | placeholder | unregistered | no_change_placeholder |
| entity_store | alicloud/service_alicloud_cms_entity_store.go | alicloud/resource_alicloud_cms_entity_store.go | alicloud/data_source_alicloud_cms_entity_store.go | placeholder | unregistered | no_change_placeholder |
| integration_policy | alicloud/service_alicloud_cms_integration_policy.go | alicloud/resource_alicloud_cms_integration_policy.go | alicloud/data_source_alicloud_cms_integration_policy.go | placeholder | unregistered | no_change_placeholder |
| memory | alicloud/service_alicloud_cms_memory.go | alicloud/resource_alicloud_cms_memory.go | alicloud/data_source_alicloud_cms_memory.go | placeholder | unregistered | no_change_placeholder |
| memory_store | alicloud/service_alicloud_cms_memory_store.go | alicloud/resource_alicloud_cms_memory_store.go | alicloud/data_source_alicloud_cms_memory_store.go | placeholder | unregistered | no_change_placeholder |
| pipeline | alicloud/service_alicloud_cms_pipeline.go | alicloud/resource_alicloud_cms_pipeline.go | alicloud/data_source_alicloud_cms_pipeline.go | placeholder | unregistered | no_change_placeholder |
| prometheus | alicloud/service_alicloud_cms_prometheus.go | alicloud/resource_alicloud_cms_prometheus.go | alicloud/data_source_alicloud_cms_prometheus.go | placeholder | unregistered | no_change_placeholder |
| service | alicloud/service_alicloud_cms_service.go | alicloud/resource_alicloud_cms_service.go | alicloud/data_source_alicloud_cms_service.go | implemented_direct_call (data source) | data source registered as `alicloud_cms_service` | migrate_now |
| umodel | alicloud/service_alicloud_cms_umodel.go | alicloud/resource_alicloud_cms_umodel.go | alicloud/data_source_alicloud_cms_umodel.go | placeholder | unregistered | no_change_placeholder |
| workspace | alicloud/service_alicloud_cms_workspace.go | alicloud/resource_alicloud_cms_workspace.go | alicloud/data_source_alicloud_cms_workspace.go | placeholder | unregistered | no_change_placeholder |

## Provider Registration Findings

- Enabled CMS data source in scope: `alicloud_cms_service` from `alicloud/provider.go`.
- No script-scoped CMS resource from this matrix is currently registered in `alicloud/provider.go`.

## Direct Call Findings

- `alicloud/data_source_alicloud_cms_service.go` currently uses direct `client.RpcPost("Cms", "2019-01-01", "OpenCmsService", ...)`.
- `alicloud/data_source_alicloud_cms_service.go` contains duplicate trailing `package alicloud` declaration.

## CWS-Lib-Go CMS API Inventory (Task Stage)

| Object Group | API File Area | Key Methods |
|---|---|---|
| addon | `alicloud_cms_addon_api.go` | `GetAddon`, `CreateAddonRelease`, `UpdateAddonRelease`, `DeleteAddonRelease`, `ListAddons` |
| agg_task | `alicloud_cms_agg_task_group_api.go` | `GetAggTaskGroup`, `CreateAggTaskGroup`, `UpdateAggTaskGroup`, `UpdateAggTaskGroupStatus`, `DeleteAggTaskGroup`, `ListAggTaskGroups` |
| alert | `alicloud_cms_alert_api.go` | `ListAlertActions`, `CreateAlertWebhook`, `UpdateAlertWebhook`, `DeleteAlertWebhooks`, `ManageAlertRules`, `QueryAlertRules` |
| cloud_resource | `alicloud_cms_cloud_resource_api.go` | CMS cloud resource list/get/create/update/delete methods |
| context | `alicloud_cms_context_api.go` | CMS context list/get/create/update/delete methods |
| context_store | `alicloud_cms_context_store_api.go` | CMS context store list/get/create/update/delete methods |
| dataset | `alicloud_cms_dataset_api.go` | `CreateDataset`, `GetDataset`, `UpdateDataset`, `DeleteDataset`, `ListDatasets`, `ExecuteQuery` |
| delivery_task | `alicloud_cms_delivery_task_api.go` | CMS delivery task list/get/create/update/delete methods |
| entity_store | `alicloud_cms_entity_store_api.go` | CMS entity store list/get/create/update/delete methods |
| integration_policy | `alicloud_cms_integration_policy_api.go` | CMS integration policy list/get/create/update/delete methods |
| memory | `alicloud_cms_memory_api.go` | CMS memory list/get/create/update/delete methods |
| memory_store | `alicloud_cms_memory_store_api.go` | CMS memory store list/get/create/update/delete methods |
| pipeline | `alicloud_cms_pipeline_api.go` | CMS pipeline list/get/create/update/delete methods |
| prometheus | `alicloud_cms_prometheus_api.go` | `CreatePrometheusView`, `GetPrometheusInstance`, `ListPrometheusVirtualInstances`, `UpdatePrometheusUserSetting` |
| service | `alicloud_cms_api.go`, `alicloud_cms_service_api.go` | `GetCmsServiceStatus`, `CreateService`, `GetService`, `UpdateService`, `DeleteService`, `ListServices` |
| umodel | `alicloud_cms_umodel_api.go` | CMS umodel list/get/create/update/delete methods |
| workspace | `alicloud_cms_workspace_api.go` | `PutWorkspace`, `GetWorkspace`, `ListWorkspaces`, `DeleteWorkspace` |

## Service Capability Rows (US2)

| Object Group | Provider Service Methods | Pagination Policy | NotFound Policy | Retry Policy | Status |
|---|---|---|---|---|---|
| addon | `ListCmsAddons`, `ListCmsAddonReleases`, `GetCmsAddonRelease` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| agg_task | `ListCmsAggTaskGroups`, `GetCmsAggTaskGroup`, `UpdateCmsAggTaskGroupStatus` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| alert | `ListCmsAlertRules`, `ManageCmsAlertRules`, `ListCmsAlertWebhooks` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| cloud_resource | `ListCmsCloudResources`, `GetCmsCloudResource` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| context | `ListCmsContexts`, `GetCmsContext` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| context_store | `ListCmsContextStores`, `GetCmsContextStore` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| dataset | `ListCmsDatasets`, `GetCmsDataset`, `ExecuteCmsDatasetQuery` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| delivery_task | `ListCmsDeliveryTasks`, `GetCmsDeliveryTask` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| entity_store | `ListCmsEntityStores`, `GetCmsEntityStore` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| integration_policy | `ListCmsIntegrationPolicies`, `GetCmsIntegrationPolicy` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| memory | `ListCmsMemories`, `GetCmsMemory` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| memory_store | `ListCmsMemoryStores`, `GetCmsMemoryStore` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| pipeline | `ListCmsPipelines`, `GetCmsPipeline` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| prometheus | `ListCmsPrometheusViews`, `GetCmsPrometheusInstance`, `ListCmsPrometheusVirtualInstances`, `UpdateCmsPrometheusUserSetting` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| service | `OpenCmsService`, `GetCmsServiceStatus` | service/api layer owns | service layer owns | service layer owns | migrated |
| umodel | `ListCmsUmodels`, `GetCmsUmodel` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |
| workspace | `ListCmsWorkspaces`, `GetCmsWorkspace`, `PutCmsWorkspace`, `DeleteCmsWorkspace` | service/api layer owns | service layer owns | service layer owns | blocked_by_cws_compile |

## Validation Log

- US1 provider test command:
	- `go test ./alicloud -run 'TestDataSourceAliCloudCmsServiceSchema|TestDataSourceAliCloudCmsServiceLayeringAudit|TestCmsScopeFileCount|TestCmsResourceAndDataSourceNoDirectRpcOrSdkCalls' -count=1`
	- Result: failed due repository/module setup error `stat .../pkg/cws-lib-go/lib/cloud/aliyun/api/alicloud: directory not found`.
- US1 cws-lib-go test command:
	- `cd pkg/cws-lib-go/lib/cloud/aliyun/api && go test ./cms -run TestCmsAPIServiceActivationMethodsCoverage -count=1`
	- Result: failed due pre-existing compile blockers in unrelated CMS wrapper files.
- Pending: US2 service capability rows update.
