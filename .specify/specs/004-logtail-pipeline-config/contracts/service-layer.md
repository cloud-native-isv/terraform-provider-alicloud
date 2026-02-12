# Contracts: Service Layer (SLS Logtail Pipeline Config)

**Date**: 2026-02-12  
**Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](../requirements.md)

## Layering Contract

- Resource 层（`alicloud/resource_alicloud_logtail_config.go`）只负责：schema、CRUD 编排、状态写入/读取。
- Service 层（`alicloud/service_alicloud_sls_logtail_config.go`）负责：生命周期操作封装、错误语义统一、等待/刷新封装。
- API 层由 `pkg/cws-lib-go` 提供：对接 SLS SDK，并暴露强类型方法。

## Service Methods (Logical)

> 方法名可在实现中微调，但语义与职责必须保持一致。

- `DescribeSlsLogtailPipelineConfig(id string) (*LogtailPipelineConfig, error)`
  - 输入：Terraform 资源 ID（建议保持 `project:config:<name>` 格式）。
  - 行为：解析 ID 后调用 API Get；NotFound 时返回可识别的 NotFound 错误（供资源层清理 state）。

- `CreateSlsLogtailPipelineConfig(project string, cfg *LogtailPipelineConfig) error`
  - 行为：调用 API Create；对可重试错误进行重试包装（由资源层或 service 统一约定）。

- `UpdateSlsLogtailPipelineConfig(project string, cfg *LogtailPipelineConfig) error`
  - 行为：调用 API Update；更新后可选等待读回一致。

- `DeleteSlsLogtailPipelineConfig(project, name string) error`
  - 行为：调用 API Delete；NotFound 视为已删除。

- `LogtailPipelineConfigStateRefreshFunc(id string, failStates []string) StateRefreshFunc`
  - NotFound：必须返回 `(nil, "", nil)`。
  - 失败态：命中 `failStates` 时返回可诊断错误。

## Error Semantics

- NotFound：资源层读取到 NotFound 时应 `d.SetId("")`。
- 参数错误：返回清晰可操作的错误信息（英文），指出字段与原因。
