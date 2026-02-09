# Implementation Plan: AliKafka 资源分层改造

**Branch**: `002-alikafka-service-layer` | **Date**: 2026-02-09 | **Spec**: [.specify/specs/002-alikafka-service-layer/requirements.md](.specify/specs/002-alikafka-service-layer/requirements.md)
**Input**: Specification from `.specify/specs/002-alikafka-service-layer/requirements.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

将 AliKafka 的 7 个资源统一改造为通过服务层完成核心生命周期操作，并在服务层补齐缺失的可复用能力，确保分页/重试/状态刷新等逻辑集中封装且对外行为保持兼容。

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**: Terraform Plugin SDK、cws-lib-go、Alibaba Cloud SDK（含本地 sdk/）  
**Storage**: N/A（状态由 Terraform 管理）  
**Testing**: go test、make test、make testacc  
**Target Platform**: Linux/macOS/Windows（Terraform Provider）
**Project Type**: 单体 Go 模块（Terraform Provider）  
**Performance Goals**: N/A（保持现有性能与行为一致）  
**Constraints**: 资源层仅调用服务层；服务层使用 cws-lib-go 强类型；分页与重试封装在服务层；使用 WaitFor/StateRefresh 规范；禁止新增弱类型 map  
**Scale/Scope**: 7 个 AliKafka 资源的服务层改造与能力补齐

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Feature-Centric Development**: Feature Index is single source of truth; all phases re-evaluate Feature changes.
- **Specification-Driven Development**: Code serves specifications; specifications are executable and generate working systems
- **Intent-Driven Development**: Focus on "what" and "why" before "how"; use rich specifications with guardrails
- **Test-First & Contract-Driven**: TDD flow followed; pure functions have unit tests; critical flows have regression coverage
- **AI Agent Integration**: Only approved agents (GitHub Copilot, Qwen Code, opencode); configuration rejects unsupported providers
- **Continuous Quality & Observability**: Structured logging; semantic versioning; CI quality gates; simple designs (YAGNI)
- **SDD Workflow Compliance**: Follow spec → plan → tasks → implement workflow with proper validation at each phase

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/[###-spec]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
├── feature-ref.md       # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this spec. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
alicloud/
├── resource_alicloud_alikafka_consumer_group.go
├── resource_alicloud_alikafka_deployment.go
├── resource_alicloud_alikafka_instance.go
├── resource_alicloud_alikafka_instance_allowed_ip_attachment.go
├── resource_alicloud_alikafka_sasl_acl.go
├── resource_alicloud_alikafka_sasl_user.go
├── resource_alicloud_alikafka_topic.go
└── service_alicloud_alikafka*.go

docs/
├── development_guide.md
└── wait_for_state.md

pkg/
└── cws-lib-go/
```

**Structure Decision**: 单体 Go Provider；变更集中在 alicloud/ 资源与服务层文件。

## Complexity Tracking

N/A

## Phase 0: Research Review & Context

- 已完成项目文档与 Feature 记忆检索（README、docs、features）。
- 未发现需要额外研究才能确定的关键技术决策；无需 research.md。

## Phase 1: Design & Contracts

1. 生成数据模型（data-model.md），聚焦 AliKafka 核心实体与关系。
2. 基于功能需求产出服务层操作契约（contracts/）。
3. 生成 quickstart.md，说明验证路径与最小验收步骤。
4. 更新 AI agent 上下文（运行项目提供的脚本）。

## Phase 2: Planning

- 拆解任务到 /speckit.tasks，覆盖服务层补齐与资源层替换调用。
- 明确测试范围（最小回归与关键生命周期用例）。
