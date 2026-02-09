---

description: "Task list for AliKafka 资源分层改造"
---

# Tasks: AliKafka 资源分层改造

**Input**: Design documents from `.specify/specs/002-alikafka-service-layer/`
**Prerequisites**: plan.md (required), requirements.md (required for user stories), data-model.md, contracts/, quickstart.md
**Tests**: Required by Constitution Principle III (Test-First Development).
**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 项目初始化与任务前置核对

- [x] T001 完成资源/服务/契约映射表并记录于 .specify/specs/002-alikafka-service-layer/feature-ref.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 所有用户故事的基础设施，必须先完成

- [x] T002 创建 AliKafka 测试公共夹具与助手在 alicloud/alikafka_test.go
- [x] T003 在 alicloud/service_alicloud_alikafka.go 中补充通用分页与重试辅助函数（供服务层统一调用）

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - 统一分层架构 (Priority: P1) 🎯 MVP

**Goal**: 7 个 AliKafka 资源仅通过服务层完成核心生命周期操作。

**Independent Test**: `go test ./alicloud -run AliKafka`（新增测试覆盖通过），并基于 quickstart.md 做最小验证。

### Tests for User Story 1 (MANDATORY)

- [x] T004 [P] [US1] 添加实例服务层生命周期单测于 alicloud/service_alicloud_alikafka_instance_test.go
- [x] T005 [P] [US1] 添加主题服务层生命周期单测于 alicloud/service_alicloud_alikafka_topic_test.go
- [x] T006 [P] [US1] 添加 SASL 用户服务层生命周期单测于 alicloud/service_alicloud_alikafka_sasl_user_test.go
- [x] T007 [P] [US1] 添加 SASL ACL 服务层生命周期单测于 alicloud/service_alicloud_alikafka_sasl_acl_test.go
- [x] T008 [P] [US1] 添加消费者组/部署/白名单服务层单测于 alicloud/service_alicloud_alikafka_consumer_group_test.go

### Implementation for User Story 1

- [x] T009 [P] [US1] 在 alicloud/service_alicloud_alikafka_instance.go 实现实例 Create/Describe/Delete/List 及 WaitFor/StateRefresh
- [x] T010 [P] [US1] 在 alicloud/service_alicloud_alikafka_topic.go 实现主题 Create/Describe/Delete/List 及 WaitFor/StateRefresh
- [x] T011 [P] [US1] 在 alicloud/service_alicloud_alikafka_sasl_user.go 实现 SASL 用户 Create/Describe/Delete/List 及 WaitFor/StateRefresh
- [x] T012 [P] [US1] 在 alicloud/service_alicloud_alikafka_sasl_acl.go 实现 SASL ACL Create/Delete/List 及 WaitFor/StateRefresh
- [x] T013 [P] [US1] 在 alicloud/service_alicloud_alikafka_consumer_group.go 实现消费者组 List/Describe 并封装分页
- [x] T014 [P] [US1] 在 alicloud/service_alicloud_alikafka.go 实现部署 List 并封装分页
- [x] T015 [P] [US1] 在 alicloud/service_alicloud_alikafka.go 实现实例白名单 Attach/Detach/List 并封装分页
- [x] T016 [US1] 重构 alicloud/resource_alicloud_alikafka_instance.go 仅通过服务层调用并使用 WaitFor
- [x] T017 [US1] 重构 alicloud/resource_alicloud_alikafka_topic.go 仅通过服务层调用并使用 WaitFor
- [x] T018 [US1] 重构 alicloud/resource_alicloud_alikafka_sasl_user.go 仅通过服务层调用并使用 WaitFor
- [x] T019 [US1] 重构 alicloud/resource_alicloud_alikafka_sasl_acl.go 仅通过服务层调用并使用 WaitFor
- [x] T020 [US1] 重构 alicloud/resource_alicloud_alikafka_consumer_group.go 仅通过服务层调用
- [x] T021 [US1] 重构 alicloud/resource_alicloud_alikafka_deployment.go 仅通过服务层调用
- [x] T022 [US1] 重构 alicloud/resource_alicloud_alikafka_instance_allowed_ip_attachment.go 仅通过服务层调用
- [x] T023 [US1] 依据 .specify/specs/002-alikafka-service-layer/quickstart.md 进行手工验收记录

**Checkpoint**: User Story 1 should be fully functional and independently testable

---

## Phase 4: User Story 2 - 服务层能力补齐 (Priority: P2)

**Goal**: 服务层提供完备可复用能力，覆盖 ID 编解码、统一错误处理与分页封装。

**Independent Test**: 服务层辅助函数单测通过，资源层无需自定义分页/重试逻辑。

### Tests for User Story 2 (MANDATORY)

- [x] T024 [P] [US2] 添加 ID Encode/Decode 单测于 alicloud/service_alicloud_alikafka_types_test.go

### Implementation for User Story 2

- [x] T025 [P] [US2] 在 alicloud/service_alicloud_alikafka_types.go 为实例/主题/用户/ACL/消费者组/部署/白名单补充 Encode/Decode Id
- [x] T026 [P] [US2] 统一服务层错误处理与重试策略于 alicloud/service_alicloud_alikafka*.go
- [x] T027 [US2] 清理资源层分页/重试逻辑（若存在）并确保统一调用服务层 in alicloud/resource_alicloud_alikafka_*.go

**Checkpoint**: Service layer capabilities complete and reusable

---

## Phase 5: User Story 3 - 用户体验不回退 (Priority: P3)

**Goal**: 既有配置保持兼容，核心操作路径不回退。

**Independent Test**: 既有配置样例可通过最小回归流程。

### Tests for User Story 3 (MANDATORY)

- [x] T028 [P] [US3] 添加 AliKafka 实例验收测试骨架于 alicloud/resource_alicloud_alikafka_instance_test.go
- [x] T029 [P] [US3] 添加主题字段映射回归测试于 alicloud/resource_alicloud_alikafka_topic_test.go
- [x] T030 [P] [US3] 添加 AliKafka 实例 testacc 于 alicloud/resource_alicloud_alikafka_instance_test.go
- [x] T031 [US3] 执行 AliKafka 实例 testacc 并记录结果于 .specify/specs/002-alikafka-service-layer/quickstart.md

### Implementation for User Story 3

- [x] T032 [US3] 复核 7 个资源字段映射与描述，确保向后兼容于 alicloud/resource_alicloud_alikafka_*.go
- [x] T033 [US3] 更新兼容性注意事项与验收说明至 .specify/specs/002-alikafka-service-layer/quickstart.md

**Checkpoint**: Backward compatibility verified

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 全局质量收尾

- [x] T034 [P] 补充开发指南与分层约束说明于 docs/development_guide.md
- [x] T035 执行最小回归检查并记录于 .specify/specs/002-alikafka-service-layer/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup completion
- **User Stories (Phase 3+)**: Depend on Foundational completion
- **Polish (Phase 6)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational (Phase 2)
- **User Story 2 (P2)**: Starts after User Story 1 to avoid API churn
- **User Story 3 (P3)**: Starts after User Story 1, may run in parallel with User Story 2 if stable

### Parallel Opportunities

- Setup and Foundational tasks marked [P] can run in parallel
- Service-layer tasks for different files in US1/US2 can run in parallel
- US3 tests can run in parallel with US2 once US1 stabilizes

---

## Parallel Example: User Story 1

- Task: "添加实例服务层生命周期单测于 alicloud/service_alicloud_alikafka_instance_test.go"
- Task: "添加主题服务层生命周期单测于 alicloud/service_alicloud_alikafka_topic_test.go"
- Task: "在 alicloud/service_alicloud_alikafka_instance.go 实现实例 Create/Describe/Delete/List 及 WaitFor/StateRefresh"
- Task: "在 alicloud/service_alicloud_alikafka_topic.go 实现主题 Create/Describe/Delete/List 及 WaitFor/StateRefresh"

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. Validate US1 independently (tests + quickstart)

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → MVP ready
3. Add User Story 2 → Test independently
4. Add User Story 3 → Test independently
5. Polish phase for cross-cutting improvements
