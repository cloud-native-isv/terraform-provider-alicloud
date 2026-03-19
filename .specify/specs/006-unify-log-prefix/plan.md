# Implementation Plan: 006-unify-log-prefix

**Branch**: `006-unify-log-prefix` | **Date**: 2026-03-19 | **Spec**: [.specify/specs/006-unify-log-prefix/requirements.md](requirements.md)
**Input**: Specification from `.specify/specs/006-unify-log-prefix/requirements.md`

## Summary

将所有日志服务相关 Terraform 入口统一到 `alicloud_log_*` 命名，目标版本立即停止 `alicloud_sls_*` 前缀入口；并提供可执行的销毁重建迁移路径、统一映射清单与发布治理流程。

技术方案以“命名入口治理 + 文档与契约先行 + 代码分阶段替换”为核心：
1. 在 provider 注册与文档侧建立 `sls -> log` 的唯一映射来源；
2. 在目标版本对 `alicloud_sls_*` 使用返回明确替代提示；
3. 通过治理清单和发布说明机制确保每次变更可追溯。

## Technical Context

**Language/Version**: Go 1.24（`go.mod`）  
**Primary Dependencies**:
- `github.com/hashicorp/terraform-plugin-sdk`（Terraform Provider 框架）
- `github.com/cloud-native-tools/cws-lib-go`（Service/API 强类型封装）
- `github.com/aliyun/alibaba-cloud-sdk-go` 与日志服务相关 SDK

**Storage**: N/A（Provider 本身无持久化；状态由 Terraform state 承载）  
**Testing**: `make test`、`go test ./alicloud`、必要时 `TF_ACC=1` 验收测试  
**Target Platform**: Linux/macOS 开发环境与 CI；Terraform Provider 运行环境
**Project Type**: 单体 Go Provider（核心目录 `alicloud/`）  
**Performance Goals**:
- 统一后用户检索日志服务能力时仅需 `alicloud_log_*`
- `alicloud_sls_*` 使用失败提示必须包含可操作替代名

**Constraints**:
- 全量纳入规则固定：凡 `alicloud_sls_*` 均纳入治理（FR-009）
- 停用策略为“最近发布版本立即生效”（FR-010）
- 迁移策略固定为“销毁重建”，不承诺 state 迁移（FR-008）
- 保持 `Resource -> Service -> API -> SDK` 分层，不在 Resource 层下沉业务逻辑

**Scale/Scope**:
- 范围：日志服务相关 `resource` + `data source` 的命名入口、映射、文档、发布治理
- 不引入新的日志服务业务能力，仅处理命名一致性与迁移治理

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **I. Layered Architecture & Library-First Design**: 命名统一不改变分层，资源/数据源层仅做入口与 schema 编排。
- **II. Standardized Interfaces**: 对外统一 `alicloud_log_*` 命名，映射关系保持一一对应或显式聚合说明。
- **III. Test-First Development**: 先补充/更新命名映射与兼容性测试，再推进入口替换。
- **IV. Integration & Contract Testing**: 通过契约定义“旧前缀报错 + 替代提示”行为，验收覆盖关键迁移流。
- **V. Observability, Versioning & Simplicity**: 发布说明必须同步停用规则与迁移指引，错误提示结构化可读。
- **VI. Continuous Integration & Quality Gates**: `make` 与相关测试通过作为合并门槛。
- **VII. Feature-Centric Development**: 复核 Feature 列表，确认本次为既有 Feature 011 的计划落地，不新增/合并/淘汰其他 Feature。

**Additional Constraints from $ARGUMENTS**: 无（本次未提供额外 `$ARGUMENTS`）

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/006-unify-log-prefix/
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
├── provider.go
├── resource_alicloud_log_*.go
├── resource_alicloud_sls_*.go
├── data_source_alicloud_log_*.go
├── data_source_alicloud_sls_*.go
├── service_alicloud_sls_*.go
└── *_test.go

docs/
├── development_guide.md
└── wait_for_state.md

.specify/memory/
├── constitution.md
├── features.md
└── features/011.md
```

**Structure Decision**: 保持现有 Provider 单仓结构，在 `alicloud/` 完成命名入口统一与错误提示替代，在 `.specify/specs/006-unify-log-prefix/` 输出设计产物，并在 feature memory 记录治理状态。

## Complexity Tracking

N/A

## Phase 0: Research Review & Context

### Information Gathering

- 已读取需求：`.specify/specs/006-unify-log-prefix/requirements.md`。
- 已读取宪章与开发规范：`.specify/memory/constitution.md`、`docs/development_guide.md`、`docs/wait_for_state.md`。
- 已读取项目上下文：`README.md`、`go.mod`。
- 已读取 Feature 索引与详情：`.specify/memory/features.md` 与 `.specify/memory/features/*.md`。
- 当前 spec 目录无 `research.md`，基于现有文档已消除阻断性未知项。

### Clarification Resolution

- 兼容策略：强制统一，目标版本立即停用 `alicloud_sls_*`。
- 迁移方式：销毁重建，不承诺 `terraform state mv`。
- 统一边界：`alicloud_sls_*` 全量纳入，不设例外。
- 发布策略：不区分 major/minor/patch，最近发布直接生效。

## Phase 1: Design & Contracts

### 1) Data Model

从需求提炼三类核心实体：
1. Naming Mapping Item（命名映射项）
2. Governance Catalog（治理清单）
3. Migration Guidance Item（迁移指引项）

详见 `data-model.md`。

### 2) API Contracts

以治理动作抽象契约接口：
- 查询命名映射清单
- 校验配置中是否存在 `alicloud_sls_*`
- 返回替代建议与迁移指引
- 维护治理清单状态与发布备注

详见 `contracts/openapi.yaml`。

### 3) Quickstart

提供最小可执行路径：
- 扫描并定位历史 `alicloud_sls_*`
- 依据映射替换为 `alicloud_log_*`
- 执行销毁重建与结果验证
- 记录发布说明与治理清单更新

详见 `quickstart.md`。

### Post-Design Constitution Re-check

- 分层架构、强类型、测试与文档约束均保持一致。
- 设计产物覆盖“统一命名 + 停用提示 + 迁移治理”关键流。
- Feature 复核结果：仅更新 Feature 011 状态与备注，无新增/合并/拆分/删除。

结果：✅ Pass

## Phase 2: Planning (Work Packages)

### WP-1 命名映射与范围固化

- 盘点所有 `alicloud_sls_*` resource/data source 并建立映射表
- 标记每项状态（已统一/待替换/下线说明）
- 输出单一权威映射来源用于文档与发布引用

### WP-2 Provider 入口统一与错误提示

- 在目标版本移除/停用 `alicloud_sls_*` 入口
- 为旧前缀使用场景提供明确替代提示（包含目标 `alicloud_log_*`）
- 保持 Resource/Service 分层，不引入弱类型请求载体

### WP-3 迁移路径与回归验证

- 按“销毁重建”策略设计迁移步骤与风险提示
- 覆盖典型场景测试：创建、读取、更新、销毁与升级校验
- 验证旧前缀报错信息可操作性

### WP-4 文档与发布治理

- 更新文档示例，确保仅暴露 `alicloud_log_*`
- 发布说明明确“本次直接生效 + 销毁重建”
- 同步治理清单并记录版本变更

### Quality Gates

- `make` 通过
- `make test` 通过
- 涉及日志服务的关键验收测试在凭证环境通过
