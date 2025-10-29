---
description: "Tasks for implementing alicloud_sls_consumer_group"
---

# Tasks: Add Terraform resource alicloud_sls_consumer_group

**Input**: Design documents from `/.specify/specs/003-add-alicloud-sls/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec; test tasks are omitted by default. You can add acceptance tests later if needed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Feature scaffolding and basic structure

 - [X] T001 [P] Create docs & examples directories if missing: `docs/r/`, `examples/sls/consumer_group/`
 - [X] T002 [P] Add resource docs stub: `docs/r/sls_consumer_group.md` (describe schema, ForceNew fields, import ID format)
 - [X] T003 [P] Add example HCL: `examples/sls/consumer_group/main.tf` (minimal working sample per quickstart)
- [ ] T004 Run build to confirm baseline compiles: `make` (from repo root)

**Checkpoint**: Repo builds cleanly; doc/example skeletons ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core provider hooks and service scaffolding

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Verify provider resource registration (add if missing) in `alicloud/provider.go` for key `alicloud_sls_consumer_group`
- [X] T006 [P] Create service file scaffold: `alicloud/service_alicloud_sls_consumer_group.go` with package `alicloud` and TODO headers
- [X] T007 [P] Create resource file scaffold: `alicloud/resource_alicloud_sls_consumer_group.go` with package `alicloud` and TODO headers
- [ ] T008 Run `make` to ensure scaffolds compile and imports resolve

**Checkpoint**: Provider recognizes the new resource symbol; empty scaffolds compile

---

## Phase 3: User Story 1 - 以基础参数创建并管理 SLS 消费组 (Priority: P1) 🎯 MVP

**Goal**: CRUD 管理消费组：创建、更新（行为参数）、删除，满足基本超时与状态回填；Create 遇已存在时 adopt 并对齐行为参数

**Independent Test**: 使用 quickstart 的最小 HCL 在一个现有 Project/Logstore 下完成一次 plan/apply/destroy 流程，并验证读取字段与远端一致、无漂移

### Implementation for User Story 1

- [X] T009 [P] [US1] Service: 定义 ID 编解码 `EncodeSlsConsumerGroupId(project, logstore, consumerGroup)` / `DecodeSlsConsumerGroupId(id)` in `alicloud/service_alicloud_sls_consumer_group.go`
- [X] T010 [US1] Service: 定义 CRUD 方法签名并通过 CWS-Lib-Go API 层调用（Create/Describe/Update/Delete，占位实现） in `service_alicloud_sls_consumer_group.go`
- [X] T011 [US1] Service: 实现 StateRefreshFunc 与 WaitFor*（Creating/Deleting） in `service_alicloud_sls_consumer_group.go`
- [X] T012 [P] [US1] Resource: 定义 Schema（Required: project, logstore, consumer_group; Optional+Computed: timeout, order），并设置 ForceNew 于标识字段 in `resource_alicloud_sls_consumer_group.go`
- [X] T013 [US1] Resource: Create — 构建请求、调用 service.Create、资源 adopt 行为（存在则收敛 timeout/order）、WaitForCreating、最后 Read in `resource_alicloud_sls_consumer_group.go`
- [X] T014 [US1] Resource: Read — 调用 service.Describe，设置全部字段（含 computed），处理 not found → `d.SetId("")` in `resource_alicloud_sls_consumer_group.go`
- [X] T015 [US1] Resource: Update — 仅允许更新 timeout/order，使用 service.Update + WaitFor（若需要） in `resource_alicloud_sls_consumer_group.go`
- [X] T016 [US1] Resource: Delete — 调用 service.Delete + StateChangeConf 等待删除完成 in `resource_alicloud_sls_consumer_group.go`
- [X] T017 [US1] Resource: Timeouts 配置（Create/Update/Delete 默认值）与日志记录、WrapError/WrapErrorf 使用 in `resource_alicloud_sls_consumer_group.go`
- [X] T018 [US1] Docs: 完善 `docs/r/sls_consumer_group.md`（描述字段、ForceNew、adopt 行为与超时）
- [X] T019 [US1] Example: 更新 `examples/sls/consumer_group/main.tf` 以匹配最终 Schema 和导入说明
- [ ] T020 Build & smoke: 运行 `make`，确保编译通过；对 example 做一次本地验证（仅语法层）

**Checkpoint**: US1 可独立验证：最小配置 CRUD + adopt 行为，plan/apply/destroy 流程通畅

---

## Phase 4: User Story 2 - 导入现有消费组纳入托管 (Priority: P2)

**Goal**: 支持 `terraform import` 以 `project:logstore:consumer_group` 格式进行导入；导入后 plan 不出现非预期变更

**Independent Test**: 准备仅包含 id 的占位配置，执行 `terraform import`，随后 `terraform plan` 无非预期变更

### Implementation for User Story 2

- [X] T021 [P] [US2] Resource: Importer — `State: schema.ImportStatePassthrough` 并在 Read 中解析 `id` via `DecodeSlsConsumerGroupId` in `resource_alicloud_sls_consumer_group.go`
- [X] T022 [US2] Service/Resource: Import 错误提示优化（解析失败、目标不存在时的提示） in `service_alicloud_sls_consumer_group.go` & `resource_alicloud_sls_consumer_group.go`
- [X] T023 [US2] Docs: 在 `docs/r/sls_consumer_group.md` 更新 Import 示例与常见错误提示
- [ ] T024 Build: 运行 `make` 验证无新增编译问题

**Checkpoint**: US2 可独立验证：按指定三段式 ID 成功导入并与远端一致

---

## Phase 5: User Story 3 - 输入校验与可用性反馈 (Priority: P3)

**Goal**: 在 plan/apply 前尽早发现无效输入；对暂时性错误自动重试并在超时后提供清晰信息；统一错误处理

**Independent Test**: 对非法名称、缺字段、越界数值分别执行 plan/apply 观察错误提示；模拟系统繁忙验证重试逻辑（如可行）

### Implementation for User Story 3

- [X] T025 [P] [US3] Resource: Schema 校验 — `validation.StringMatch`/正则校验 consumer_group 命名与长度；timeout 合理区间检查 in `resource_alicloud_sls_consumer_group.go`
- [X] T026 [US3] Service/Resource: 错误处理模式 — 使用 `IsNotFoundError/IsAlreadyExistError/NeedRetry` 和 `WrapError/WrapErrorf`；实现 `resource.Retry` 针对 `ServiceUnavailable/Throttling/SystemBusy/OperationConflict` 等 in both files
- [X] T027 [US3] Service: 完善 WaitFor 目标/失败状态与轮询间隔，确保在超时边界条件下行为稳定 in `service_alicloud_sls_consumer_group.go`
- [X] T028 [US3] Docs: 在 `docs/r/sls_consumer_group.md` 增补校验与错误处理说明（含常见报错与修复建议）
- [ ] T029 Build: 运行 `make`，确保所有改动编译通过

**Checkpoint**: US3 可独立验证：plan 阶段即拦截无效输入；遇到暂时性错误具备重试与明确反馈

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: 文档、示例与可观察性完善

- [ ] T030 [P] 更新 `CHANGELOG.md` 新增资源记录
- [ ] T031 代码清理与注释（英文注释、日志前缀一致）
- [ ] T032 [P] 文档链接校对与 examples 手动走查
- [ ] T033 （可选）增加验收测试样例（需要外部环境与凭据）
- [ ] T034 最终一次 `make` 并准备提交 PR（遵循治理规范）

---

## Dependencies & Execution Order

### Phase Dependencies
- Setup (Phase 1): 无依赖
- Foundational (Phase 2): 依赖 Setup 完成 — 阻塞所有用户故事
- User Stories (Phase 3+): 均依赖 Foundational 完成
- Polish (Final Phase): 依赖所需用户故事完成

### User Story Dependencies
- US1 (P1): Foundational 完成后可开始；独立验证
- US2 (P2): Foundational 完成后可开始；可与 US1 并行，但推荐在 US1 合入后执行
- US3 (P3): Foundational 完成后可开始；与 US1/US2 无强耦合

### Within Each User Story
- 无测试要求情况下：按 Service → Resource → Docs/Examples → Build 顺序；跨文件的任务标记 [P]
- 同一文件内任务顺序化，避免冲突

### Parallel Opportunities
- 跨文件任务均已以 [P] 标注，可并行推进（如 Service 与 Resource 的不同子任务、Docs/Examples）

---

## Parallel Example: User Story 1

```bash
# 可并行的任务（不同文件）：
T009 Service: ID 编解码           # service_alicloud_sls_consumer_group.go
T012 Resource: 定义 Schema       # resource_alicloud_sls_consumer_group.go
T018 Docs: 完善资源文档          # docs/r/sls_consumer_group.md
T019 Example: 更新示例           # examples/sls/consumer_group/main.tf
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)
1. 完成 Phase 1/2
2. 完成 US1 的 Service/Resource/Docs/Examples 实现
3. `make` 构建并对 example 做语法级校验
4. 暂停并验证 US1 独立可用

### Incremental Delivery
1. 完成 Setup + Foundational → 基础就绪
2. 交付 US1 → 验证 → 合并
3. 交付 US2 → 验证 Import → 合并
4. 交付 US3 → 验证校验与重试 → 合并

### Notes
- 严格遵循 Constitution：分层、状态管理、错误处理、命名与超时
- 每个阶段完成后执行一次 `make`，确保编译通过
