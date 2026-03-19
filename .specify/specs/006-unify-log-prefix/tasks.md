---
description: "Task list for 006-unify-log-prefix implementation"
---

# Tasks: 统一日志服务资源前缀

**Requirement ID**: 006  
**Requirement Key**: 006-unify-log-prefix  
**Related Feature**: 011 SLS Log Prefix Unification  
**Input**: Design documents from `.specify/specs/006-unify-log-prefix/`  
**Prerequisites**: plan.md, requirements.md, data-model.md, contracts/openapi.yaml, quickstart.md

**Arguments Analysis**: 本次未提供 `$ARGUMENTS`，按默认流程生成完整任务列表。  
**Tests**: 本需求包含迁移与停用行为风险，保留关键自动化与手工验证任务。

## Definition of Done (DoD)

- [ ] 统一命名入口实现完成并通过代码评审
- [ ] `alicloud_sls_*` 停用行为具备明确替代提示
- [ ] 迁移文档与治理清单已同步发布信息
- [ ] `make test` 通过，关键 SLS 相关用例通过
- [ ] 各用户故事可独立验证并满足 requirements.md 成功标准

## Format: `[ID] [P?] [Story] Description`

- `[P]` 表示可并行执行（不同文件且无未完成依赖）
- `[USx]` 仅用于用户故事阶段任务
- 每个任务描述包含明确文件路径

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立命名统一实施基线与任务输入清单

- [ ] T001 盘点当前 SLS 命名入口并生成基线清单到 `.specify/specs/006-unify-log-prefix/contracts/naming-inventory.md`
- [ ] T002 建立历史到统一命名映射初稿到 `.specify/specs/006-unify-log-prefix/contracts/naming-mapping.md`
- [ ] T003 [P] 提取 provider 注册点中的 `alicloud_sls_*` 键并记录到 `.specify/specs/006-unify-log-prefix/contracts/provider-registry-baseline.md`（来源 `alicloud/provider.go`）
- [ ] T004 [P] 从 `quickstart.md` 生成迁移验证检查单到 `.specify/specs/006-unify-log-prefix/checklists/migration-validation.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 构建所有用户故事共用的阻塞性基础能力

**⚠️ CRITICAL**: 完成本阶段后方可进入用户故事实现

- [ ] T005 在 `alicloud/provider.go` 新增集中式命名映射注释区并标记治理入口（resource/data source 分开）
- [ ] T006 [P] 新增统一前缀校验与替代提示辅助函数到 `alicloud/common.go`
- [ ] T007 [P] 为 `alicloud/provider.go` 增加旧前缀停用错误输出路径（统一调用 T006 中的提示函数）
- [ ] T008 [P] 新增命名映射一致性单测文件 `alicloud/prefix_unification_mapping_test.go`
- [ ] T009 将治理清单回写规则补充到 `.specify/specs/006-unify-log-prefix/contracts/naming-governance-rules.md`

**Checkpoint**: 停用提示基础设施与映射基线可复用，用户故事可独立推进

---

## Phase 3: User Story 1 - 使用统一前缀完成新建与维护 (Priority: P1) 🎯 MVP

**Goal**: 日志服务能力统一以 `alicloud_log_*` 暴露，避免新建/维护时出现双命名入口。

**Independent Test**: 仅保留 `alicloud_log_*` 为可发现入口，`alicloud_sls_*` 不再作为唯一可用入口。

### Tests for User Story 1

- [ ] T010 [P] [US1] 在 `alicloud/prefix_unification_mapping_test.go` 添加 `alicloud_sls_* -> alicloud_log_*` 映射覆盖测试
- [ ] T011 [P] [US1] 在 `alicloud/provider_sls_prefix_registration_test.go` 增加 provider 注册键一致性测试（聚焦 `alicloud/provider.go`）

### Implementation for User Story 1

- [ ] T012 [US1] 更新 `alicloud/provider.go` 的 data source 注册：统一切换到 `alicloud_log_*` 命名入口
- [ ] T013 [US1] 更新 `alicloud/provider.go` 的 resource 注册：统一切换到 `alicloud_log_*` 命名入口
- [ ] T014 [P] [US1] 在 `alicloud/resource_alicloud_sls_alert.go` 与 `alicloud/resource_alicloud_sls_scheduled_sql.go` 对齐统一命名导出函数名
- [ ] T015 [P] [US1] 在 `alicloud/data_source_alicloud_sls_alerts.go` 与 `alicloud/data_source_alicloud_sls_projects.go` 对齐统一命名导出函数名
- [ ] T016 [US1] 更新映射文档 `.specify/specs/006-unify-log-prefix/contracts/naming-mapping.md`，标记 US1 覆盖项为 `unified`
- [ ] T017 [US1] 执行并记录 US1 独立验证到 `.specify/specs/006-unify-log-prefix/checklists/us1-verification.md`

**Checkpoint**: 用户可通过统一前缀完成日志服务能力检索与配置

---

## Phase 4: User Story 2 - 老配置可平滑过渡到统一前缀 (Priority: P2)

**Goal**: 为存量 `alicloud_sls_*` 用户提供可执行的销毁重建迁移路径与可操作提示。

**Independent Test**: 用户按迁移步骤替换命名并重建后，配置校验通过且核心结果一致。

### Tests for User Story 2

- [ ] T018 [P] [US2] 在 `alicloud/prefix_unification_mapping_test.go` 添加旧前缀停用错误提示断言（含替代名称）
- [ ] T019 [P] [US2] 新增迁移路径回归测试 `alicloud/prefix_unification_migration_test.go`（覆盖销毁重建流程）

### Implementation for User Story 2

- [ ] T020 [US2] 在 `alicloud/provider.go` 对 `alicloud_sls_*` 入口接入明确报错与替代建议
- [ ] T021 [US2] 更新迁移手册 `.specify/specs/006-unify-log-prefix/quickstart.md`，固定“销毁重建”步骤与风险提示
- [ ] T022 [US2] 将迁移样例补充到 `.specify/specs/006-unify-log-prefix/contracts/migration-examples.md`
- [ ] T023 [US2] 执行并记录 US2 独立验证到 `.specify/specs/006-unify-log-prefix/checklists/us2-verification.md`

**Checkpoint**: 历史用户可按指南在统一版本完成迁移与验证

---

## Phase 5: User Story 3 - 团队可治理命名一致性 (Priority: P3)

**Goal**: 输出可持续治理的命名清单、状态与发布同步机制。

**Independent Test**: 维护者可从治理清单准确查看每项能力目标命名与当前阶段状态。

### Implementation for User Story 3

- [ ] T024 [P] [US3] 在 `.specify/specs/006-unify-log-prefix/contracts/naming-governance-rules.md` 增加状态机规则（`pending/unified/retired_with_note`）
- [ ] T025 [US3] 更新 `.specify/specs/006-unify-log-prefix/contracts/naming-mapping.md`，补齐纳入项/下线说明/更新时间字段
- [ ] T026 [P] [US3] 在 `.specify/specs/006-unify-log-prefix/contracts/openapi.yaml` 对齐治理与校验接口字段（含 releaseNoteRef）
- [ ] T027 [US3] 输出发布同步模板到 `.specify/specs/006-unify-log-prefix/contracts/release-note-template.md`
- [ ] T028 [US3] 执行并记录 US3 独立验证到 `.specify/specs/006-unify-log-prefix/checklists/us3-verification.md`

**Checkpoint**: 命名治理清单可持续维护且可用于版本沟通

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 完成跨故事质量收敛与发布前验证

- [ ] T029 [P] 统一更新 spec 文档交叉引用路径（`plan.md`、`data-model.md`、`quickstart.md`、`contracts/openapi.yaml`）
- [ ] T030 运行 `go test ./alicloud -run PrefixUnification -v` 并记录结果到 `.specify/specs/006-unify-log-prefix/checklists/test-report.md`
- [ ] T031 运行 `make test` 并记录结果到 `.specify/specs/006-unify-log-prefix/checklists/test-report.md`
- [ ] T032 执行 `quickstart.md` 手工回归并记录到 `.specify/specs/006-unify-log-prefix/checklists/manual-regression.md`
- [ ] T033 同步 feature 跟踪（`tasks` 阶段备注）到 `.specify/memory/features/011.md` 与 `.specify/memory/features.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2 → User Stories (Phase 3/4/5) → Phase 6
- User Story phases均依赖 Phase 2 完成

### User Story Dependencies

- **US1 (P1)**: 无对其他故事依赖，是 MVP
- **US2 (P2)**: 依赖 US1 完成统一入口后再定义停用与迁移提示
- **US3 (P3)**: 可与 US2 并行推进文档治理，但最终状态回写依赖 US1/US2 的实际结果

### Within Each Story

- 先测试/校验任务，再实施任务
- 先 provider 注册与公共提示逻辑，再资源/数据源局部调整
- 每个用户故事完成后立即执行独立验证任务

---

## Parallel Execution Opportunities

- **Setup**: T003、T004 可并行
- **Foundational**: T006、T007、T008 可并行
- **US1**: T010、T011、T014、T015 可并行
- **US2**: T018、T019 可并行
- **US3**: T024、T026 可并行
- **Polish**: T029 与测试准备可并行，T031 需在实现完成后执行

---

## Parallel Example: User Story 1

```bash
# 并行编写 US1 测试
Task: T010 [US1] alicloud/prefix_unification_mapping_test.go
Task: T011 [US1] alicloud/provider_sls_prefix_registration_test.go

# 并行处理 US1 资源/数据源命名对齐
Task: T014 [US1] alicloud/resource_alicloud_sls_alert.go + alicloud/resource_alicloud_sls_scheduled_sql.go
Task: T015 [US1] alicloud/data_source_alicloud_sls_alerts.go + alicloud/data_source_alicloud_sls_projects.go
```

---

## Implementation Strategy

### MVP First (US1)

1. 完成 Phase 1、Phase 2
2. 完成 US1（Phase 3）
3. 立即执行 T017 验证统一前缀入口
4. 在 MVP 通过后再推进迁移与治理故事

### Incremental Delivery

1. US1：统一入口可用
2. US2：迁移路径与停用提示可用
3. US3：治理清单与发布机制可持续运行

### Validation Gates

- 每个故事完成后都有独立验证文件
- Phase 6 才执行全局 `make test` 与手工回归
- 最终将任务阶段变化回写 Feature 011
