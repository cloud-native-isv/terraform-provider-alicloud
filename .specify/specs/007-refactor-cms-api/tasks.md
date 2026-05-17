# Tasks: CMS Provider 封装能力改造

**Requirement ID**: 007  
**Requirement Key**: 007-refactor-cms-api  
**Related Feature**: 005 CWS-Lib-Go Integration  
**Input**: Design documents from `.specify/specs/007-refactor-cms-api/`  
**Prerequisites**: `plan.md`, `requirements.md`, `research.md`, `data-model.md`, `contracts/cms-provider-contract.yaml`, `quickstart.md`

**User Input Analysis**: `$ARGUMENTS` 为空；按默认工作流从现有设计产物生成完整可执行任务清单。  
**Input Handling Strategy**: 无额外背景、任务大纲或可执行任务条目需要合并；保持按用户故事优先级组织。

**Tests**: 本规格、宪法和计划均要求 TDD/可验证改造证据，因此每个用户故事均包含测试或审计任务，测试任务必须先于对应实现任务完成并在实现前失败。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Definition of Done (DoD)

- 51 个脚本枚举 CMS Provider 文件均有范围判定和迁移证据。
- 已启用 CMS 能力不再从资源层或数据源层直接调用 SDK/RPC。
- 新增 Provider 服务层方法优先使用 CWS-Lib-Go CMS typed structs。
- `alicloud_cms_service` 数据源保持 schema、ID、status 和错误反馈兼容。
- 单元测试、分层审计测试和可运行的 targeted Go 测试通过；无法运行的验收测试必须记录阻塞原因。
- Feature 005 任务阶段备注和 Feature Index 日期已同步。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel when files do not conflict and prerequisite tasks are complete.
- **[Story]**: User story label for story phases only.
- Every task includes at least one exact repository file path.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish traceability inputs and scope baselines before changing code.

- [X] T001 Create the initial 51-file CMS scope matrix in `.specify/specs/007-refactor-cms-api/cms-migration-matrix.md` using the exact script-expanded file list from `requirements.md`
- [X] T002 [P] Record current Provider registration and direct-call findings from `alicloud/provider.go` and `alicloud/data_source_alicloud_cms_service.go` in `.specify/specs/007-refactor-cms-api/cms-migration-matrix.md`
- [X] T003 [P] Record CWS-Lib-Go CMS API wrapper method inventory from `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api.go` and `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_service_api.go` in `.specify/specs/007-refactor-cms-api/cms-migration-matrix.md`
- [X] T004 [P] Record task-stage Feature review inputs in `.specify/specs/007-refactor-cms-api/feature-ref.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared helper and audit structures required before any user-story implementation.

**⚠️ CRITICAL**: No user story implementation should begin until this phase is complete.

- [X] T005 Create CMS API construction helper and test seam for `connectivity.AliyunClient` credentials in `alicloud/service_alicloud_cms_common.go`
- [X] T006 Create the migration evidence template with 17 object-group sections in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`
- [X] T007 [P] Add the cross-file CMS scope and direct-call audit test scaffold in `alicloud/cms_layering_test.go`
- [X] T008 [P] Add the CMS service capability matrix test scaffold in `alicloud/service_alicloud_cms_capability_test.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in priority order or parallel by story.

---

## Phase 3: User Story 1 - CMS 代码统一使用封装能力 (Priority: P1) 🎯 MVP

**Goal**: Ensure the currently enabled CMS data source and all script-scoped CMS files are governed by the service/API wrapper boundary.

**Independent Test**: Run CMS layering tests to confirm the 51 files are audited and no enabled resource/data source bypasses service/API wrapper calls; run `alicloud_cms_service` data source tests to confirm behavior remains compatible.

### Tests for User Story 1 (MANDATORY) ⚠️

> Write these tests first and confirm they fail before implementation.

- [X] T009 [P] [US1] Add failing CWS API tests for `OpenCmsService` and `GetCmsServiceStatus` behavior in `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api_test.go`
- [X] T010 [P] [US1] Add failing data source tests for `enable=Off`, `enable=On`, `status`, and ID compatibility in `alicloud/data_source_alicloud_cms_service_test.go`
- [X] T011 [P] [US1] Add failing scope and no-direct-RPC audit assertions for the 51 CMS files in `alicloud/cms_layering_test.go`

### Implementation for User Story 1

- [X] T012 [US1] Implement missing CWS-Lib-Go `OpenCmsService` wrapper and response normalization in `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api.go`
- [X] T013 [US1] Implement Provider service methods for opening and reading CMS service state through CWS-Lib-Go in `alicloud/service_alicloud_cms_service.go`
- [X] T014 [US1] Refactor `alicloud_cms_service` data source to call `CmsService` methods, remove `client.RpcPost`, and remove the duplicate package declaration in `alicloud/data_source_alicloud_cms_service.go`
- [X] T015 [US1] Update migration evidence for the `service` object group as `migrated` in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`
- [X] T016 [US1] Record US1 validation commands and results for `alicloud_cms_service` in `.specify/specs/007-refactor-cms-api/cms-migration-matrix.md`

**Checkpoint**: User Story 1 delivers MVP value and is independently testable.

---

## Phase 4: User Story 2 - 服务层补齐 CMS 生命周期能力 (Priority: P2)

**Goal**: Provide reusable CMS service-layer methods for all 17 CMS object groups, backed by local CWS-Lib-Go CMS API wrapper methods and scoped by lifecycle capability.

**Independent Test**: Run service capability matrix tests to confirm each object group has a service-layer decision, wrapper mapping, lifecycle operations, pagination/not-found handling policy, and documented dynamic-type exceptions where required.

### Tests for User Story 2 (MANDATORY) ⚠️

> Write these tests first and confirm they fail before implementation.

- [X] T017 [P] [US2] Add failing lifecycle coverage assertions for all 17 CMS object groups in `alicloud/service_alicloud_cms_capability_test.go`
- [X] T018 [P] [US2] Add failing API wrapper method inventory assertions for object-group mapping in `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api_test.go`

### Implementation for User Story 2

- [X] T019 [P] [US2] Implement typed service-layer wrapper methods for addon release/query operations in `alicloud/service_alicloud_cms_addon.go`
- [X] T020 [P] [US2] Implement typed service-layer wrapper methods for aggregate task group lifecycle/status operations in `alicloud/service_alicloud_cms_agg_task.go`
- [X] T021 [P] [US2] Implement typed service-layer wrapper methods for alert webhook/rule/query operations in `alicloud/service_alicloud_cms_alert.go`
- [X] T022 [P] [US2] Implement typed service-layer wrapper methods for cloud resource query/lifecycle operations in `alicloud/service_alicloud_cms_cloud_resource.go`
- [X] T023 [P] [US2] Implement typed service-layer wrapper methods for context query/lifecycle operations in `alicloud/service_alicloud_cms_context.go`
- [X] T024 [P] [US2] Implement typed service-layer wrapper methods for context store query/lifecycle operations in `alicloud/service_alicloud_cms_context_store.go`
- [X] T025 [P] [US2] Implement typed service-layer wrapper methods for dataset create/read/update/delete/list/query operations in `alicloud/service_alicloud_cms_dataset.go`
- [X] T026 [P] [US2] Implement typed service-layer wrapper methods for delivery task lifecycle/list operations in `alicloud/service_alicloud_cms_delivery_task.go`
- [X] T027 [P] [US2] Implement typed service-layer wrapper methods for entity store lifecycle/list operations in `alicloud/service_alicloud_cms_entity_store.go`
- [X] T028 [P] [US2] Implement typed service-layer wrapper methods for integration policy lifecycle/list operations in `alicloud/service_alicloud_cms_integration_policy.go`
- [X] T029 [P] [US2] Implement typed service-layer wrapper methods for memory lifecycle/list operations in `alicloud/service_alicloud_cms_memory.go`
- [X] T030 [P] [US2] Implement typed service-layer wrapper methods for memory store lifecycle/list operations in `alicloud/service_alicloud_cms_memory_store.go`
- [X] T031 [P] [US2] Implement typed service-layer wrapper methods for pipeline lifecycle/list operations in `alicloud/service_alicloud_cms_pipeline.go`
- [X] T032 [P] [US2] Implement typed service-layer wrapper methods for Prometheus view/instance/virtual-instance/user-setting operations in `alicloud/service_alicloud_cms_prometheus.go`
- [X] T033 [P] [US2] Implement typed service-layer wrapper methods for umodel lifecycle/list operations in `alicloud/service_alicloud_cms_umodel.go`
- [X] T034 [P] [US2] Implement typed service-layer wrapper methods for workspace put/get/list/delete operations in `alicloud/service_alicloud_cms_workspace.go`
- [X] T035 [US2] Update all 17 object-group service capability rows and pagination/not-found/retry policies in `.specify/specs/007-refactor-cms-api/cms-migration-matrix.md`
- [X] T036 [US2] Document dynamic type exceptions for dataset query, alert notify strategy, Prometheus settings, and service observability config in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`

**Checkpoint**: Service layer has complete reusable capability coverage without requiring resource/data source layers to know API details.

---

## Phase 5: User Story 3 - CMS 用户体验保持兼容 (Priority: P3)

**Goal**: Prove that internal migration does not regress user-visible CMS behavior, provider registration, schema semantics, ID handling, or error feedback.

**Independent Test**: Run Provider compatibility tests and optional acceptance checks for enabled CMS capabilities; confirm placeholder files remain unregistered unless explicitly enabled with tests and docs.

### Tests for User Story 3 (MANDATORY) ⚠️

> Write these tests first and confirm they fail before implementation.

- [X] T037 [P] [US3] Add failing Provider registration compatibility tests for enabled CMS names in `alicloud/provider_cms_compatibility_test.go`
- [X] T038 [P] [US3] Add failing schema and ID regression tests for `alicloud_cms_service` in `alicloud/data_source_alicloud_cms_service_test.go`
- [X] T039 [US3] Add failing placeholder-not-registered regression assertions for script-scoped placeholder files in `alicloud/provider_cms_compatibility_test.go`

### Implementation for User Story 3

- [X] T040 [US3] Add a compatibility example for the `alicloud_cms_service` data source in `examples/cms/cms_service/main.tf`
- [X] T041 [US3] Audit existing CMS Terraform names, schema fields, IDs, and error messages against current behavior in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`
- [X] T042 [US3] Record optional acceptance-test command and any credential/environment blocker for CMS compatibility in `.specify/specs/007-refactor-cms-api/quickstart.md`

**Checkpoint**: CMS user-visible behavior remains compatible and documented.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final formatting, validation, feature tracking, and quality gates across all user stories.

- [X] T043 [P] Run `gofmt` and `goimports` cleanup for CMS Provider files in `alicloud/service_alicloud_cms_service.go` and `alicloud/data_source_alicloud_cms_service.go`
- [X] T044 [P] Run CWS-Lib-Go CMS API tests and fix failures in `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api_test.go`
- [X] T045 Run Provider CMS tests and fix failures in `alicloud/cms_layering_test.go`, `alicloud/service_alicloud_cms_capability_test.go`, `alicloud/provider_cms_compatibility_test.go`, and `alicloud/data_source_alicloud_cms_service_test.go`
- [X] T046 Run `make test` and `make`; record any environment blocker in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`
- [X] T047 Update task-stage Feature 005 notes in `.specify/memory/features/005.md`
- [X] T048 Update Feature Index last-updated date for Feature 005 in `.specify/memory/features.md`
- [X] T049 Validate that every task in `.specify/specs/007-refactor-cms-api/tasks.md` follows `- [ ] T### [P?] [US?] Description with file path`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - blocks all user stories.
- **User Stories (Phase 3+)**: Depend on Foundational completion.
- **Polish (Phase 6)**: Depends on completed target user stories.

### User Story Dependencies

- **User Story 1 (P1)**: Starts after Foundational; MVP and first implementation slice.
- **User Story 2 (P2)**: Starts after Foundational; can run in parallel with US1 only after shared helper/test seams exist, but should integrate US1's service helper conventions.
- **User Story 3 (P3)**: Starts after Foundational; best run after US1 for accurate compatibility evidence.

### Within Each User Story

- Tests/audits must be written before implementation.
- API wrapper gaps must be fixed before Provider service methods that call them.
- Service methods must exist before resource/data source refactors.
- Evidence files must be updated before considering the story complete.

---

## Parallel Opportunities

- T002, T003, and T004 can run in parallel after T001 is created.
- T007 and T008 can run in parallel after T005 and T006 are defined.
- US1 test tasks T009, T010, and T011 can run in parallel.
- US2 service implementation tasks T019 through T034 can run in parallel after T017 and T018 define expected coverage.
- US3 test tasks T037 and T038 can run in parallel; T039 shares `alicloud/provider_cms_compatibility_test.go` with T037 and should be sequenced after T037.
- T043 and T044 can run in parallel during polish; T045 must follow relevant test file edits.

---

## Parallel Example: User Story 1

```text
Task: T009 Add CWS API tests in pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_api_test.go
Task: T010 Add data source tests in alicloud/data_source_alicloud_cms_service_test.go
Task: T011 Add layering tests in alicloud/cms_layering_test.go
```

---

## Parallel Example: User Story 2

```text
Task: T019 Implement addon service in alicloud/service_alicloud_cms_addon.go
Task: T025 Implement dataset service in alicloud/service_alicloud_cms_dataset.go
Task: T032 Implement Prometheus service in alicloud/service_alicloud_cms_prometheus.go
Task: T034 Implement workspace service in alicloud/service_alicloud_cms_workspace.go
```

---

## Parallel Example: User Story 3

```text
Task: T037 Add Provider registration tests in alicloud/provider_cms_compatibility_test.go
Task: T038 Add cms_service schema and ID tests in alicloud/data_source_alicloud_cms_service_test.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 / US1.
3. Validate that `alicloud_cms_service` no longer uses direct `RpcPost` and that all 51 files have scope decisions.
4. Stop and review before enabling broader CMS object work.

### Incremental Delivery

1. Foundation ready: scope matrix, helper seam, evidence template, audit test scaffolds.
2. US1: migrate enabled data source and prove layering.
3. US2: fill service-layer methods for each object group without exposing new user-visible resources by default.
4. US3: prove compatibility and registration boundaries.
5. Polish: format, tests, make targets, feature memory, final task-format validation.

### Parallel Team Strategy

1. One developer owns CWS API wrapper and `alicloud_cms_service` migration.
2. Multiple developers split US2 object-group service files because each file is independent after the shared helper exists.
3. One developer owns compatibility tests and examples.
4. Integrate through shared evidence files and the CMS scope matrix.

---

## Notes

- Keep placeholder resource/data source files unregistered unless a later task adds schema, service methods, tests, docs, and provider registration evidence.
- Do not add direct `client.RpcPost`, `WithCmsClient`, or raw SDK calls in resource/data source CMS files.
- Prefer CWS-Lib-Go CMS typed structs; document any dynamic payload exception in `.specify/specs/007-refactor-cms-api/cms-migration-evidence.md`.
- Commit after each completed user story or coherent task group.
