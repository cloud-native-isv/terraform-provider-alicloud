# Implementation Plan: 005-restore-logtail-config

**Branch**: `005-restore-logtail-config` | **Date**: 2026-02-13 | **Spec**: [.specify/specs/005-restore-logtail-config/requirements.md](requirements.md)
**Input**: Specification from `.specify/specs/005-restore-logtail-config/requirements.md`

## Summary

恢复 `alicloud_logtail_config` 的旧版兼容实现，确保存量 Terraform 配置升级后可继续工作；同时将当前“新版 Pipeline Config 实现”以新资源名 `alicloud_logtail_pipeline_config` 独立提供，避免命名冲突与语义混淆。

实现路径：
1. 资源注册层新增/恢复双资源并存：旧资源名与新资源名同时可用。
2. 文件与实现命名完成边界重构：`resource_alicloud_logtail_config.go` 保留旧行为；新版迁移到 `resource_alicloud_logtail_pipeline_config.go`。
3. Service 层保持分层调用与强类型约束；并存场景遵循“最后一次成功写入生效”。
4. 补充文档：明确兼容窗口（至少 2 个小版本）与迁移指引，仅文档提示弃用，不增加运行时提示。

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**:
- Terraform Plugin SDK (`github.com/hashicorp/terraform-plugin-sdk`)
- Alibaba Cloud Go SDK（项目既有依赖）
- CWS-Lib-Go（Service/API 分层能力，按现有代码路径使用）

**Storage**: N/A（Provider 无本地持久化，状态由 Terraform state 管理）  
**Testing**: `go test ./alicloud`、`make test`、相关资源验收测试（`TF_ACC=1`）  
**Target Platform**: Terraform Provider for Alibaba Cloud（Linux/macOS CI 与开发环境）
**Project Type**: 单体 Go Provider（核心源码位于 `alicloud/`）  
**Performance Goals**:
- 存量 `alicloud_logtail_config` 配置升级后 `plan` 不出现“未知资源类型”阻断
- 新旧资源在稳定读回后，重复 `plan` 不产生非预期漂移

**Constraints**:
- 必须遵循 `Resource -> Service -> API -> SDK` 分层
- 新增代码不得以 `map[string]interface{}` 作为主要请求/响应载体
- 兼容窗口至少 2 个小版本，仅文档标注弃用，不加运行时 deprecation 提示
- 本次不实现自动/半自动状态迁移

**Scale/Scope**:
- 仅覆盖 Logtail Config / Pipeline Config 相关资源、服务层与文档
- 不扩展其他 SLS 资源能力

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **I. Layered Architecture & Library-First Design**: 资源层只负责 Terraform schema 与编排，业务逻辑与 API 调用保留在 Service 层。
- **II. Standardized Interfaces**: 维持旧资源接口稳定，同时引入新资源名承载新版能力；分页/重试逻辑不下沉到资源层。
- **III. Test-First Development**: 在实现前补齐/调整单测与验收测试用例，覆盖旧资源回归与新资源生命周期。
- **IV. Integration & Contract Testing**: 并存场景、同对象写入冲突场景纳入契约与验收验证。
- **V. Observability, Versioning & Simplicity**: 保持日志与错误信息可诊断，文档明确兼容窗口与升级路径。
- **VI. Continuous Integration & Quality Gates**: 以 `make` 与相关测试通过作为合并门槛。
- **VII. Feature-Centric Development**: 复核 Feature 010，无新增/拆分 Feature；将状态从 Planned 更新为 Implemented，并记录本次关键变化。

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/005-restore-logtail-config/
├── plan.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── openapi.yaml
└── tasks.md (由 /speckit.tasks 生成)
```

### Source Code (repository root)

```text
alicloud/
├── provider.go (或资源注册入口文件)
├── resource_alicloud_logtail_config.go
├── resource_alicloud_logtail_pipeline_config.go
├── service_alicloud_sls_logtail_config.go
├── service_alicloud_sls_logtail_pipeline_config.go
└── *_test.go

docs/
└── （必要时补充变更说明）
```

**Structure Decision**: 采用现有 Provider 单体结构，仅在 `alicloud/` 内完成资源命名边界恢复与并存治理，并在 spec 文档目录输出设计产物。

## Complexity Tracking

N/A

## Phase 0: Research Review & Context

### Information Gathering

- 已读取并对齐权威约束：`.specify/memory/constitution.md`、`docs/development_guide.md`、`docs/wait_for_state.md`
- 已读取项目上下文：`README.md`
- 已读取 Feature 体系：`.specify/memory/features.md` 与 `.specify/memory/features/*.md`
- 已读取当前规格：`.specify/specs/005-restore-logtail-config/requirements.md`
- 当前 spec 目录无 `research.md`，通过现有文档与既有 spec（004/005）信息完成关键澄清，不存在阻断性未知项

### Clarification Resolution

- 并存规则：新旧资源允许同时指向同一远端对象，语义为“最后一次成功写入生效”。
- 弃用策略：仅文档提示，不新增运行时警告。
- 兼容窗口：至少 2 个小版本。
- 迁移范围：不包含自动状态迁移。

## Phase 1: Design & Contracts

### 1) Data Model

提炼三个核心实体：
1. Legacy Logtail Config Resource（旧资源兼容面）
2. Pipeline Logtail Config Resource（新资源能力面）
3. Terraform Resource State（并存与可追踪性）

详见 `data-model.md`。

### 2) API Contracts

将用户动作抽象为契约端点：
- 解析旧资源配置
- 解析新资源配置
- 新旧资源 CRUD
- 并存冲突写入校验
- 兼容窗口元数据查询

契约文件输出为 OpenAPI：`contracts/openapi.yaml`。

### 3) Quickstart

提供最小验证路径：
- 旧资源回归验证（plan/apply 不阻断）
- 新资源生命周期验证（create/update/delete）
- 同对象并存验证（最后写入生效）

详见 `quickstart.md`。

### Post-Design Constitution Re-check

设计产物与宪章保持一致：
- 分层架构保持不变
- 强类型约束未放宽
- 测试与契约覆盖关键流
- Feature 索引已同步更新

结果：✅ Pass

## Phase 2: Planning (Work Packages)

### WP-1 资源命名边界恢复

- 恢复旧资源入口：`alicloud_logtail_config`
- 新版实现改名并独立注册：`alicloud_logtail_pipeline_config`
- 完成文件命名调整与注册映射

### WP-2 服务层与状态一致性

- 对齐旧/新资源各自 Service 调用路径
- 并存场景读取一致性与错误信息可操作化
- 维持 `WaitFor` 与 NotFound 语义符合规范

### WP-3 测试与回归

- 旧资源兼容回归（plan/read/update）
- 新资源全生命周期测试
- 同对象并存冲突写入规则验证

### WP-4 文档与发布说明

- 资源文档标注旧资源弃用状态与兼容窗口
- 发布说明增加迁移指引与范围边界（不含自动迁移）

### Quality Gates

- `make` 通过
- `make test` 通过
- 相关资源验收测试在具备凭证环境下通过
