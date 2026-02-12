# Research Notes: 004-logtail-pipeline-config

**Date**: 2026-02-12  
**Branch**: 004-logtail-pipeline-config  
**Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](requirements.md)

## Sources

- [docs/development_guide.md](../../docs/development_guide.md): 分层架构、强类型约束、ID 编解码与状态等待规范。
- [docs/wait_for_state.md](../../docs/wait_for_state.md): `StateRefreshFunc` 返回值与 NotFound 行为约定。
- `pkg/cws-lib-go/lib/cloud/aliyun/api/sls/alicloud_sls_logtail_pipeline_api.go`: 新版 Logtail Pipeline Config 的 CRUD/LIST。
- `pkg/cws-lib-go/lib/cloud/aliyun/api/sls/alicloud_sls_logtail_types.go`: 新旧 Logtail 类型定义（旧版标记 deprecated，新版结构为插件链）。

## Findings

1. **新版能力入口已就绪**：`cws-lib-go` 提供 `Get/Create/Update/DeleteLogtailPipelineConfig`，并将旧版 `LogtailConfig` 标记为 Deprecated。
2. **新版数据结构差异显著**：新版 `LogtailPipelineConfig` 由 `inputs/processors/flushers/aggregators/global/task` 组成，且插件内容在 `cws-lib-go` 中表示为 `map` 形式。
3. **Provider 侧强类型约束的处理方式**：为遵循“Service 方法签名强类型”，计划在 Provider 内部使用强类型领域模型承载插件的 `type + config_json`，仅在转换到 `cws-lib-go` 时临时构造 map。
4. **稳定 diff 的关键点**：所有 JSON 字段必须做规范化；插件链顺序保持输入顺序（不排序）。
5. **StateRefreshFunc NotFound 约定**：NotFound 场景必须返回 `(nil, "", nil)`，而非返回非 nil object。

## Decisions

- 资源 schema 以“插件链 block + JSON 明细”的模式重做，不做旧 schema 兼容。
- Service 层对外仅暴露强类型模型；弱结构仅存在于转换边界内部。

## Open Questions

- 无（本次 plan 阶段可在不新增澄清点的前提下进入 tasks）。
