# Implementation Plan: 004-logtail-pipeline-config

**Branch**: `004-logtail-pipeline-config` | **Date**: 2026-02-12 | **Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](requirements.md)
**Input**: Specification from `.specify/specs/004-logtail-pipeline-config/requirements.md`

## Summary

把 `alicloud_logtail_config` 资源与 `alicloud/service_alicloud_sls_logtail_config.go` 的旧版 LogtailConfig API 调用，全面替换为新版 Logtail Pipeline Config API（通过 `pkg/cws-lib-go` 暴露的 `Get/Create/Update/DeleteLogtailPipelineConfig`）。

由于新版数据结构与旧版完全不同，本次同时重做资源 schema：用“插件流水线（inputs/processors/flushers/aggregators）+ 可选 global/task 设置 + log_sample”的模型表达配置；并对 JSON 字段做规范化，消除排序/空白导致的永久 diff。

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**:
- Terraform Plugin SDK（`github.com/hashicorp/terraform-plugin-sdk`）
- CWS-Lib-Go（`pkg/cws-lib-go`，SLS API 封装与类型）
  
**Storage**: N/A  
**Testing**: `go test ./...` + Terraform Acceptance Tests（`make testacc`）  
**Target Platform**: Terraform Provider（alicloud）  
**Project Type**: 单体 Go Provider（源码位于 `alicloud/`）  
**Performance Goals**: 等价配置连续两次 `plan` 不产生差异；Create/Update 后在合理时间内可稳定读回一致状态  
**Constraints**:
- 严格分层：Resource -> Service -> API（cws-lib-go）-> SDK
- 强类型：Resource/Service 的核心入参出参不以 `map[string]interface{}` 作为主要载体
- 状态等待：遵循项目 `WaitForState` / `StateRefreshFunc` 编码规范
  
**Scale/Scope**: 仅覆盖 `alicloud_logtail_config` 资源与其 Service；不做向前兼容

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Layered Architecture & Library-First**: 资源层只做 Terraform 编排；Service 层统一调用 `cws-lib-go`。
- **Strong Typing Constraints**: Service 层对外暴露明确的领域模型；弱结构仅在“JSON <-> cws-lib-go 插件 map”转换边界内部出现。
- **Test-First Development**: 变更前先补齐映射/规范化的单元测试与验收测试骨架。
- **State Management**: Create/Update/Delete 的等待与 NotFound 处理遵循 [docs/wait_for_state.md](../../docs/wait_for_state.md)。
- **Feature-Centric Development**: Feature 010 已登记；本次 plan 不引入新 Feature。

**Gates Status**: ✅ All gates pass（存在一处“插件配置弱结构”的局部例外，已在 Complexity Tracking 说明）

## Project Structure

### Documentation (this spec)

```text
.specify/specs/004-logtail-pipeline-config/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── service-layer.md
│   └── terraform-schema.md
├── feature-ref.md
└── tasks.md
```

### Source Code (repository root)

```text
alicloud/
├── resource_alicloud_logtail_config.go
├── service_alicloud_sls_logtail_config.go
└── (to add) resource_alicloud_logtail_config_test.go

pkg/cws-lib-go/lib/cloud/aliyun/api/sls/
├── alicloud_sls_logtail_pipeline_api.go
└── alicloud_sls_logtail_types.go
```

**Structure Decision**: 使用现有 Provider 结构；本次变更集中在 Resource/Service 两层，并复用 `pkg/cws-lib-go` 的新版 API。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 新版插件配置需要在转换边界内部临时使用 `map[string]interface{}` | `cws-lib-go` 的 `LogtailPipelineConfig` 插件链以 `[]map[string]interface{}` 表达；Provider 需要对 Terraform 的 JSON 字段进行编解码以组装请求 | 将所有插件配置静态化会导致 schema 过度膨胀且无法覆盖插件生态；因此采用“结构化外层 + JSON 明细”的折中，并把弱结构限制在转换边界内部 |

## Phase 0: Research Review & Context

本 spec 目录未预先提供 `research.md`；已通过以下现有资料完成技术澄清与决策：

- 项目权威开发约束：[docs/development_guide.md](../../docs/development_guide.md)
- 状态等待与 Refresh 规范：[docs/wait_for_state.md](../../docs/wait_for_state.md)
- `pkg/cws-lib-go` 已实现的新版 Logtail Pipeline Config API 与类型

输出产物：本次将补充生成 `research.md`，记录上述结论与关键引用点。

## Phase 1: Design & Contracts

### 1) 新 schema 设计（不兼容旧版）

资源：`alicloud_logtail_config`

顶层字段：
- `project` (required, ForceNew)
- `name` (required, ForceNew)
- `inputs` (required, list block, MinItems=1)
- `processors` (optional, list block)
- `flushers` (required, list block, MinItems=1)
- `aggregators` (optional, list block)
- `global_json` (optional, JSON string)
- `task_json` (optional, JSON string)
- `log_sample` (optional)
- `create_time` / `last_modify_time` (computed)

插件 block：
- `type` (required)
- `config_json` (optional, JSON string)

规范化：所有 JSON 字段进入 state 前做 canonical normalization；插件链顺序保持输入顺序（不排序）。

### 2) Service 迁移设计

将 Service 层的旧调用替换为新版：
- `GetLogtailConfig` -> `GetLogtailPipelineConfig`
- `CreateLogtailConfig` -> `CreateLogtailPipelineConfig`
- `UpdateLogtailConfig` -> `UpdateLogtailPipelineConfig`
- `DeleteLogtailConfig` -> `DeleteLogtailPipelineConfig`

并按项目规范更新 `StateRefreshFunc`：NotFound 返回 `(nil, "", nil)`。

### 3) Contracts（计划产物）

- `contracts/service-layer.md`: Resource/Service/API 职责边界、方法语义、NotFound/重试语义
- `contracts/terraform-schema.md`: schema 结构、校验规则、JSON 规范化与读回映射规则

## Phase 2: Planning (Work Packages)

> 具体可执行任务拆分留给 `/speckit.tasks`；此处只给出工作包与验收点。

### WP-A: Service 层 API 替换与等待

- 替换 CRUD 调用为 Pipeline Config 新版
- 增补/修正等待与刷新逻辑（Create/Update/Delete 后一致性）
- 验收：`Read` 可稳定读回；NotFound 行为符合规范

### WP-B: Resource schema 重做与双向映射

- 实现新 schema（插件链 + JSON 字段）
- 实现 schema <-> 模型 <-> `slsAPI.LogtailPipelineConfig` 的映射与 JSON 规范化
- 验收：连续两次 `plan` 无差异；Update 能正确触发远端更新

### WP-C: 测试（先行）

- 单元测试：JSON 规范化、插件链映射、边界错误信息
- 验收测试：Create/Read/Update/Import/Delete

### WP-D: 文档与示例

- 更新资源文档与示例，明确不兼容变更与推荐写法

### Quality Gates

- `make` 通过
- `make test` 通过
- 在具备凭证的环境下，相关验收测试通过
