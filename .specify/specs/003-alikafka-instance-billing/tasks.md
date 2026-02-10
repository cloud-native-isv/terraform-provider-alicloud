---

description: "Task list for AliKafka 实例计费组合一致性"
---

# Tasks: AliKafka 实例计费组合一致性

**Input**: Design documents from `.specify/specs/003-alikafka-instance-billing/`
**Prerequisites**: plan.md (required), requirements.md (required for user stories), data-model.md, contracts/, quickstart.md
**Tests**: Required by Constitution Principle III (Test-First Development).
**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 项目初始化与任务前置核对

- [x] T001 对照 cws-lib-go 计费规范梳理实现要点于 pkg/cws-lib-go/.specify/specs/007-kafka-instance-billing/requirements.md
- [x] T002 盘点现有实例创建路径与状态回显逻辑于 alicloud/resource_alicloud_alikafka_instance.go、alicloud/service_alicloud_alikafka_instance.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 所有用户故事的基础能力，必须先完成

- [x] T003 [P] 统一实例类型与计费方式常量/解析助手于 alicloud/service_alicloud_alikafka_types.go
- [x] T004 [P] 为常量与解析助手补齐单测于 alicloud/service_alicloud_alikafka_types_test.go
- [x] T005 在服务层添加计费组合校验与错误构造函数于 alicloud/service_alicloud_alikafka_instance.go
- [x] T006 为组合校验逻辑补齐单测于 alicloud/service_alicloud_alikafka_instance_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - 创建预留型实例 (Priority: P1) 🎯 MVP

**Goal**: 预留型实例支持预付费与后付费创建，并返回实例/订单标识。

**Independent Test**: 单测覆盖 CreateInstance 请求映射；最小验收按 quickstart.md。

### Tests for User Story 1 (MANDATORY)

- [x] T007 [P] [US1] 添加预留型预付费/后付费请求映射单测于 alicloud/service_alicloud_alikafka_instance_test.go
- [x] T008 [P] [US1] 补充预留型实例验收测试用例于 alicloud/alikafka_test.go

### Implementation for User Story 1

- [x] T009 [US1] 在 alicloud/service_alicloud_alikafka_instance.go 统一使用 CreateInstance API 构建预留型预付费/后付费请求
- [x] T010 [US1] 更新 alicloud/resource_alicloud_alikafka_instance.go 创建流程透传实例类型与计费方式
- [x] T011 [US1] 在 alicloud/resource_alicloud_alikafka_instance.go 回显 instance_type 与 billing_type 状态字段

**Checkpoint**: User Story 1 should be fully functional and independently testable

---

## Phase 4: User Story 2 - 创建 Serverless 实例 (Priority: P2)

**Goal**: Serverless 实例支持后付费创建并返回实例标识。

**Independent Test**: 单测覆盖 Serverless + Postpaid 请求映射；最小验收按 quickstart.md。

### Tests for User Story 2 (MANDATORY)

- [x] T012 [P] [US2] 添加 Serverless + Postpaid 请求映射单测于 alicloud/service_alicloud_alikafka_instance_test.go
- [x] T013 [P] [US2] 补充 Serverless 实例验收测试用例于 alicloud/alikafka_test.go

### Implementation for User Story 2

- [x] T014 [US2] 扩展 alicloud/service_alicloud_alikafka_instance.go 的 CreateInstance 映射以支持 Serverless + Postpaid
- [x] T015 [US2] 更新 alicloud/resource_alicloud_alikafka_instance.go 处理 Serverless 创建与状态回显

**Checkpoint**: User Story 2 should be fully functional and independently testable

---

## Phase 5: User Story 3 - 不支持组合的明确提示 (Priority: P3)

**Goal**: Serverless + Prepaid 在创建前被拒绝并返回清晰错误。

**Independent Test**: 单测验证非法组合返回明确错误；资源层创建不会触发 API 调用。

### Tests for User Story 3 (MANDATORY)

- [x] T016 [P] [US3] 添加 Serverless + Prepaid 非法组合单测于 alicloud/service_alicloud_alikafka_instance_test.go
- [x] T017 [P] [US3] 添加资源层非法组合校验单测于 alicloud/resource_alicloud_alikafka_instance_test.go

### Implementation for User Story 3

- [x] T018 [US3] 在 alicloud/resource_alicloud_alikafka_instance.go 创建前调用组合校验并返回清晰错误
- [x] T019 [US3] 确保错误信息包含非法组合与可用组合提示于 alicloud/service_alicloud_alikafka_instance.go

**Checkpoint**: User Story 3 should be fully functional and independently testable

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 全局质量收尾

- [x] T020 [P] 更新验收与执行记录于 .specify/specs/003-alikafka-instance-billing/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup completion
- **User Stories (Phase 3+)**: Depend on Foundational completion
- **Polish (Phase 6)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational (Phase 2)
- **User Story 2 (P2)**: Starts after User Story 1 to keep创建路径一致
- **User Story 3 (P3)**: Can start after Foundational, but建议在 US1/US2 稳定后完成

### Parallel Opportunities

- Setup/Foundational tasks marked [P] can run in parallel
- Tests within each user story can run in parallel
- US2 tests can begin after US1 基础映射稳定

---

## Parallel Example: User Story 1

- Task: "添加预留型预付费/后付费请求映射单测于 alicloud/service_alicloud_alikafka_instance_test.go"
- Task: "补充预留型实例验收测试用例于 alicloud/alikafka_test.go"
- Task: "在 alicloud/service_alicloud_alikafka_instance.go 统一使用 CreateInstance API 构建预留型预付费/后付费请求"

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
