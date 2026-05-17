# CMS Migration Evidence

## Summary

- Requirement Key: `007-refactor-cms-api`
- Feature: `005 / CWS-Lib-Go Integration`
- Evidence Status: in progress

## Evidence Template

For each object group:

- Decision: `migrated` / `placeholder_no_user_visible_change` / `blocked` / `deferred`
- Reason
- Provider files
- Service methods
- API methods
- Tests
- Review notes

---

## addon

- Decision: blocked
- Reason: Provider service skeleton implemented, but typed cws-lib-go cms integration remains blocked by unresolved compile issues in local cws-lib-go cms package.
- Provider files: `alicloud/service_alicloud_cms_addon.go`, `alicloud/resource_alicloud_cms_addon.go`, `alicloud/data_source_alicloud_cms_addon.go`
- Service methods: `ListCmsAddons`, `ListCmsAddonReleases`, `GetCmsAddonRelease`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## agg_task

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_agg_task.go`, `alicloud/resource_alicloud_cms_agg_task.go`, `alicloud/data_source_alicloud_cms_agg_task.go`
- Service methods: `ListCmsAggTaskGroups`, `GetCmsAggTaskGroup`, `UpdateCmsAggTaskGroupStatus`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## alert

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_alert.go`, `alicloud/resource_alicloud_cms_alert.go`, `alicloud/data_source_alicloud_cms_alert.go`
- Service methods: `ListCmsAlertRules`, `ManageCmsAlertRules`, `ListCmsAlertWebhooks`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## cloud_resource

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_cloud_resource.go`, `alicloud/resource_alicloud_cms_cloud_resource.go`, `alicloud/data_source_alicloud_cms_cloud_resource.go`
- Service methods: `ListCmsCloudResources`, `GetCmsCloudResource`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## context

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_context.go`, `alicloud/resource_alicloud_cms_context.go`, `alicloud/data_source_alicloud_cms_context.go`
- Service methods: `ListCmsContexts`, `GetCmsContext`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## context_store

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_context_store.go`, `alicloud/resource_alicloud_cms_context_store.go`, `alicloud/data_source_alicloud_cms_context_store.go`
- Service methods: `ListCmsContextStores`, `GetCmsContextStore`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## dataset

- Decision: blocked
- Reason: Provider service skeleton implemented; dynamic query semantics preserved but typed integration blocked by cws compile errors.
- Provider files: `alicloud/service_alicloud_cms_dataset.go`, `alicloud/resource_alicloud_cms_dataset.go`, `alicloud/data_source_alicloud_cms_dataset.go`
- Service methods: `ListCmsDatasets`, `GetCmsDataset`, `ExecuteCmsDatasetQuery`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## delivery_task

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_delivery_task.go`, `alicloud/resource_alicloud_cms_delivery_task.go`, `alicloud/data_source_alicloud_cms_delivery_task.go`
- Service methods: `ListCmsDeliveryTasks`, `GetCmsDeliveryTask`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## entity_store

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_entity_store.go`, `alicloud/resource_alicloud_cms_entity_store.go`, `alicloud/data_source_alicloud_cms_entity_store.go`
- Service methods: `ListCmsEntityStores`, `GetCmsEntityStore`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## integration_policy

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_integration_policy.go`, `alicloud/resource_alicloud_cms_integration_policy.go`, `alicloud/data_source_alicloud_cms_integration_policy.go`
- Service methods: `ListCmsIntegrationPolicies`, `GetCmsIntegrationPolicy`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## memory

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_memory.go`, `alicloud/resource_alicloud_cms_memory.go`, `alicloud/data_source_alicloud_cms_memory.go`
- Service methods: `ListCmsMemories`, `GetCmsMemory`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## memory_store

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_memory_store.go`, `alicloud/resource_alicloud_cms_memory_store.go`, `alicloud/data_source_alicloud_cms_memory_store.go`
- Service methods: `ListCmsMemoryStores`, `GetCmsMemoryStore`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## pipeline

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_pipeline.go`, `alicloud/resource_alicloud_cms_pipeline.go`, `alicloud/data_source_alicloud_cms_pipeline.go`
- Service methods: `ListCmsPipelines`, `GetCmsPipeline`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## prometheus

- Decision: blocked
- Reason: Provider service skeleton implemented; settings payload remains dynamic by API semantics and typed integration is blocked by cws compile issues.
- Provider files: `alicloud/service_alicloud_cms_prometheus.go`, `alicloud/resource_alicloud_cms_prometheus.go`, `alicloud/data_source_alicloud_cms_prometheus.go`
- Service methods: `ListCmsPrometheusViews`, `GetCmsPrometheusInstance`, `ListCmsPrometheusVirtualInstances`, `UpdateCmsPrometheusUserSetting`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## service

- Decision: migrated
- Reason: `alicloud_cms_service` data source has been refactored to call `CmsService` methods and no longer invokes `RpcPost` directly in data source layer.
- Provider files: `alicloud/service_alicloud_cms_service.go`, `alicloud/resource_alicloud_cms_service.go`, `alicloud/data_source_alicloud_cms_service.go`
- Service methods: `NewCmsService`, `OpenCmsService`, `GetCmsServiceStatus`
- API methods: `OpenCmsService` wrapper added in `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api.go` (task-level implementation)
- Tests: `alicloud/data_source_alicloud_cms_service_test.go`, `alicloud/cms_layering_test.go`, `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api_test.go`
- Review notes: provider-side integration currently avoids direct `cws-lib-go/cms` import because the local `pkg/cws-lib-go/lib/cloud/aliyun/api/cms` package has pre-existing compile blockers unrelated to this slice.

## umodel

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_umodel.go`, `alicloud/resource_alicloud_cms_umodel.go`, `alicloud/data_source_alicloud_cms_umodel.go`
- Service methods: `ListCmsUmodels`, `GetCmsUmodel`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## workspace

- Decision: blocked
- Reason: Provider service skeleton implemented, waiting on cws-lib-go cms compile unblock.
- Provider files: `alicloud/service_alicloud_cms_workspace.go`, `alicloud/resource_alicloud_cms_workspace.go`, `alicloud/data_source_alicloud_cms_workspace.go`
- Service methods: `ListCmsWorkspaces`, `GetCmsWorkspace`, `PutCmsWorkspace`, `DeleteCmsWorkspace`
- API methods: pending typed integration
- Tests: `alicloud/service_alicloud_cms_capability_test.go`
- Review notes: blocked_by_cws_compile

## Dynamic Exception Log

- dataset query dynamic payload: allowed by API semantics, constrained at service boundary (`ExecuteCmsDatasetQuery`).
- alert notify strategy dynamic payload: allowed by API semantics, constrained at service boundary (`ManageCmsAlertRules`).
- prometheus setting dynamic payload: allowed by API semantics, constrained at service boundary (`UpdateCmsPrometheusUserSetting`).
- service observability dynamic payload: remains in cws wrapper scope and blocked from provider integration until cws compile issues are resolved.

## Compatibility Audit (US3)

- Enabled in-scope Terraform name remains `alicloud_cms_service`.
- `alicloud/data_source_alicloud_cms_service.go` keeps compatibility IDs: `CmsServiceHasNotBeenOpened` and `CmsServiceHasBeenOpened`.
- Status compatibility remains `Opened` when service enable operation succeeds.
- Placeholder object groups remain unregistered in `alicloud/provider.go`.

## Blockers

- Provider targeted test command `go test ./alicloud -run ...` fails with module path/setup error: `stat .../pkg/cws-lib-go/lib/cloud/aliyun/api/alicloud: directory not found`.
- cws-lib-go CMS targeted test fails because remaining pre-existing SDK field mismatches in pipeline_utils.go and prometheus_api.go remain outside this spec's change scope.
- Phase 6 wrapper test command `go test ./cms -run Test -count=1` compiled 7 files we fixed (dataset_utils, delivery_task_utils, entity_store_utils, integration_policy_api, memory_utils, memory_store_utils, memory_api), but still fails due to remaining pre-existing errors in `alicloud_cms_pipeline_utils.go` and `alicloud_cms_prometheus_api.go` (CreatePipelineResponseBody has no PipelineId; CreatePrometheusView signature removed workspace param; etc.).
- Phase 6 provider test command `go test ./alicloud ...` fails because unrelated existing tests in the same package do not compile (for example `resource_alicloud_alikafka_instance_test.go` and `resource_alicloud_logtail_*_test.go` report undefined symbols).
- Phase 6 root make validation `make -C /cws_data/terraform-provider-alicloud test` fails in `fmtcheck` because script `/cws_data/terraform-provider-alicloud/scripts/gofmtcheck.sh` is missing in current workspace checkout.

## T044 Fix Summary

Resolved pre-existing SDK field mismatches across 7 cws-lib-go CMS files:
- `alicloud_cms_dataset_utils.go`: CreateDatasetResponseBody (only RequestId), GetDatasetResponseBody (no DatasetId field, used DatasetName as ID), ListDatasetsResponseBodyDatasets (no DatasetId)
- `alicloud_cms_delivery_task_utils.go`: CreateDeliveryTaskRequest (Description→TaskDescription), GetDeliveryTaskResponseBody (nested under DeliveryTask sub-object)
- `alicloud_cms_entity_store_utils.go`: CreateEntityStoreRequest (empty struct), GetEntityStoreResponseBody (only WorkspaceName/RegionId/RequestId), GetEntityStoreDataResponseBody (Data/Header instead of Entities)
- `alicloud_cms_integration_policy_api.go`: 10 methods with `workspace` param removed, updated to new SDK signatures (CreateIntegrationPolicyResponseBody.Policy.PolicyId, GetIntegrationVersionForCSResponseBody.IntegrationVersion, etc.)
- `alicloud_cms_memory_utils.go`: SearchMemoriesRequest (TopK/Query instead of MaxResults/NextToken), SearchMemoriesResponseBody (Results instead of Memories), GetMemoryResponseBody (Id/Memory instead of MemoryId/MemoryContent), GetMemoriesResponseBody (Results instead of Memories), GetMemoryHistoryResponseBody (Results instead of HistoryItems)
- `alicloud_cms_memory_store_utils.go`: ListMemoryStoresRequest (MemoryStoreName instead of Query), CreateMemoryStoreResponseBody (no MemoryStoreId), GetMemoryStoreResponseBody (no MemoryStoreId), ListMemoryStoresResponseBodyMemoryStores (no MemoryStoreId/Status)
- `alicloud_cms_memory_api.go`: SearchMemories pagination loop removed (NextToken not available in new SDK), switched to single-call with updated `memoriesFromSearchResponse` signature
