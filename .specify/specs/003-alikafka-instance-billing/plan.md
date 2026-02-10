# Implementation Plan: AliKafka 实例计费组合一致性

**Branch**: `003-alikafka-instance-billing` | **Date**: 2026-02-10 | **Spec**: [.specify/specs/003-alikafka-instance-billing/requirements.md](.specify/specs/003-alikafka-instance-billing/requirements.md)
**Input**: Specification from `.specify/specs/003-alikafka-instance-billing/requirements.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

对 AliKafka 实例创建路径进行计费组合一致性改造，统一使用同一 CreateInstance API 覆盖预留型（预付费/后付费）与 Serverless（后付费）三种合法组合；在资源层/服务层加入组合校验与默认行为保持，明确拒绝 Serverless + 预付费。

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**: Terraform Plugin SDK、cws-lib-go、Alibaba Cloud SDK（含本地 sdk/）  
**Storage**: N/A（状态由 Terraform 管理）  
**Testing**: go test、make test、make testacc  
**Target Platform**: Linux/macOS/Windows（Terraform Provider）
**Project Type**: 单体 Go 模块（Terraform Provider）  
**Performance Goals**: 组合校验在创建前完成（SC-002 要求 2 秒内返回清晰错误）  
**Constraints**: 资源层仅调用服务层；服务层使用 cws-lib-go 强类型；分页与重试封装在服务层；使用 WaitFor/StateRefresh 规范；禁止新增弱类型 map；不支持组合必须在创建前拒绝；缺省参数沿用既有默认行为  
**Scale/Scope**: AliKafka 实例创建路径的计费组合校验与服务层/资源层一致性更新

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
├── resource_alicloud_alikafka_instance.go
├── service_alicloud_alikafka_instance.go
└── service_alicloud_alikafka*.go

docs/
├── development_guide.md
└── wait_for_state.md

pkg/
└── cws-lib-go/
```

**Structure Decision**: 单体 Go Provider；变更集中在 AliKafka 实例资源与服务层。

## Complexity Tracking

N/A

## Phase 0: Research Review & Context

- 已完成项目文档与 Feature 记忆检索（README、docs、features）。
- 未发现需要额外研究才能确定的关键技术决策；无需 research.md。

## Phase 1: Design & Contracts

1. 生成数据模型（data-model.md），覆盖实例类型、计费方式与创建结果。
2. 基于功能需求产出服务层操作契约（contracts/）。
3. 生成 quickstart.md，说明最小验收步骤与测试路径。
4. 更新 AI agent 上下文（如项目提供脚本）。

## Phase 2: Planning

- 拆解任务到 /speckit.tasks，覆盖资源层与服务层的创建逻辑、组合校验与状态回显。
- 明确测试范围（最小回归与关键组合验收用例）。
