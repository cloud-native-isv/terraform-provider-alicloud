# Tasks: 005-restore-logtail-config（Logtail 旧版兼容恢复 + 新版独立命名）

**Input**: 设计文档位于 `.specify/specs/005-restore-logtail-config/`（plan.md, requirements.md, data-model.md, contracts/openapi.yaml, quickstart.md）

**$ARGUMENTS 分析结果**: 本次未提供额外参数（无背景补充、无任务大纲、无额外任务条目），按现有设计文档生成完整任务清单。

**组织方式**: 按用户故事分组，确保每个故事可独立实现与独立验证。

---

## Phase 1: Setup（准备与复核）

**Purpose**: 对齐需求、计划、契约和 Feature 索引，形成可执行基线。

- [x] T001 复核需求与验收场景定义，确认 US1/US2/US3 独立验收口径：`.specify/specs/005-restore-logtail-config/requirements.md`
- [x] T002 复核实现计划中的分层/强类型/兼容窗口约束：`.specify/specs/005-restore-logtail-config/plan.md`
- [x] T003 [P] 复核数据模型与并存规则（最后一次成功写入生效）：`.specify/specs/005-restore-logtail-config/data-model.md`
- [x] T004 [P] 复核契约端点与 quickstart 验证路径：`.specify/specs/005-restore-logtail-config/contracts/openapi.yaml`、`.specify/specs/005-restore-logtail-config/quickstart.md`
- [x] T005 复核 Feature 列表是否需新增/合并/拆分；如有变更同步更新：`.specify/memory/features.md`、`.specify/memory/features/010.md`

---

## Phase 2: Foundational（阻塞性基础改造）

**Purpose**: 完成双资源并存的基础骨架与共享能力，阻塞后续所有用户故事。

**⚠️ CRITICAL**: 本阶段完成前不进入任何用户故事实现。

- [x] T006 在 Provider 注册层补齐双资源入口并保持向后兼容映射：`alicloud/provider.go`
- [x] T007 [P] 抽离/整理 Pipeline 资源共享映射与 JSON 处理边界（仅限映射边界可弱结构）：`alicloud/sls_logtail_pipeline_config_mapping.go`、`alicloud/sls_logtail_pipeline_config_json.go`
- [x] T008 [P] 对齐旧版与新版 Service 分层职责与调用边界（Resource -> Service -> API）：`alicloud/service_alicloud_sls_logtail_config.go`、`alicloud/service_alicloud_sls_logtail_pipeline_config.go`
- [x] T009 建立（或重命名）新版资源文件骨架并接入资源工厂方法：`alicloud/resource_alicloud_logtail_pipeline_config.go`
- [x] T010 为双资源共存基础增加最小回归测试入口（编译级 + 资源注册可见性）：`alicloud/resource_alicloud_logtail_config_test.go`、`alicloud/resource_alicloud_logtail_pipeline_config_test.go`

**Checkpoint**: 双资源并存基础可编译、可注册、分层边界清晰。

---

## Phase 3: User Story 1 - 保持既有配置可持续使用 (Priority: P1) 🎯 MVP

**Goal**: 恢复 `alicloud_logtail_config` 旧版兼容行为，确保存量配置升级后不阻断。

**Independent Test**: 使用仅含 `alicloud_logtail_config` 的历史配置执行 init/plan/apply，不出现未知资源类型，read/update 行为连续。

### Tests for User Story 1（先测后改）

- [x] T011 [P] [US1] 补充旧资源兼容回归测试（plan/read/update 基线场景）：`alicloud/resource_alicloud_logtail_config_test.go`
- [x] T012 [P] [US1] 补充旧资源 schema 映射与回填稳定性单测：`alicloud/resource_alicloud_logtail_config_mapping_test.go`

### Implementation for User Story 1

- [x] T013 [US1] 恢复 `alicloud_logtail_config` 的旧版 schema 与字段语义：`alicloud/resource_alicloud_logtail_config.go`
- [x] T014 [US1] 恢复旧资源 Create/Read/Update/Delete 调用链到旧版 LogtailConfig Service：`alicloud/resource_alicloud_logtail_config.go`
- [x] T015 [US1] 修复旧资源 Import/ID 解析与状态回填一致性：`alicloud/resource_alicloud_logtail_config.go`、`alicloud/service_alicloud_sls_logtail_config.go`
- [x] T016 [US1] 校验旧资源只读/刷新流程在升级场景不触发破坏性漂移：`alicloud/resource_alicloud_logtail_config.go`

**Checkpoint**: 旧资源单独可用且升级无阻断。

---

## Phase 4: User Story 2 - 新版能力以新资源名独立提供 (Priority: P1)

**Goal**: 以 `alicloud_logtail_pipeline_config` 独立承载新版能力并支持完整生命周期。

**Independent Test**: 使用仅含 `alicloud_logtail_pipeline_config` 的配置完成 create/read/update/delete，重复 plan 无非预期差异。

### Tests for User Story 2（先测后改）

- [x] T017 [P] [US2] 新增新版资源生命周期验收测试骨架（CRUD + Import）：`alicloud/resource_alicloud_logtail_pipeline_config_test.go`
- [x] T018 [P] [US2] 新增新版资源 schema 映射与 JSON 规范化单测：`alicloud/sls_logtail_pipeline_config_mapping_test.go`

### Implementation for User Story 2

- [x] T019 [US2] 将当前新版实现迁移到新资源文件并暴露资源工厂：`alicloud/resource_alicloud_logtail_pipeline_config.go`
- [x] T020 [US2] 在新版资源中实现 Create/Read/Update/Delete/Import 与等待逻辑：`alicloud/resource_alicloud_logtail_pipeline_config.go`
- [x] T021 [US2] 确保新版资源仅调用 Pipeline Service，不混用旧 API：`alicloud/resource_alicloud_logtail_pipeline_config.go`、`alicloud/service_alicloud_sls_logtail_pipeline_config.go`
- [x] T022 [US2] 清理旧资源文件中的新版残留实现，避免命名与语义混淆：`alicloud/resource_alicloud_logtail_config.go`

**Checkpoint**: 新资源可独立生命周期管理，且与旧资源语义边界明确。

---

## Phase 5: User Story 3 - 升级路径清晰可控 (Priority: P2)

**Goal**: 在同一 workspace 支持新旧资源并存治理，提供可操作错误与迁移说明。

**Independent Test**: 同时声明新旧资源并指向同一远端对象，分别更新后结果符合“最后一次成功写入生效”，并有可理解错误反馈。

### Tests for User Story 3（先测后改）

- [x] T023 [P] [US3] 新增并存场景验收测试（同对象双资源 + 最后写入生效）：`alicloud/resource_alicloud_logtail_config_test.go`、`alicloud/resource_alicloud_logtail_pipeline_config_test.go`
- [x] T024 [P] [US3] 新增混用语义错误反馈测试（错误信息可操作）：`alicloud/resource_alicloud_logtail_pipeline_config_test.go`

### Implementation for User Story 3

- [x] T025 [US3] 在资源/服务层补齐并存读写一致性处理与冲突场景行为说明：`alicloud/resource_alicloud_logtail_config.go`、`alicloud/resource_alicloud_logtail_pipeline_config.go`、`alicloud/service_alicloud_sls_logtail_config.go`、`alicloud/service_alicloud_sls_logtail_pipeline_config.go`
- [x] T026 [US3] 统一混用新旧语义时的错误消息文本，确保清晰可操作：`alicloud/resource_alicloud_logtail_config.go`、`alicloud/resource_alicloud_logtail_pipeline_config.go`
- [x] T027 [US3] 更新迁移文档与兼容窗口说明（至少 2 个小版本，文档提示弃用）：`README.md`、`CHANGELOG.md`
- [x] T028 [US3] 补充本 spec 的执行说明与并存验证记录模板：`.specify/specs/005-restore-logtail-config/quickstart.md`

**Checkpoint**: 并存治理可验证，迁移路径文档清晰。

---

## Phase 6: Polish & Cross-Cutting Concerns（收尾与质量门禁）

**Purpose**: 完成跨故事质量收敛与发布准备。

- [x] T029 [P] 运行编译与单元测试并修复回归：`Makefile`、`alicloud/`
- [x] T030 [P] 在具备凭证环境执行相关验收测试并记录结果：`alicloud/resource_alicloud_logtail_config_test.go`、`alicloud/resource_alicloud_logtail_pipeline_config_test.go`
- [x] T031 复核 WaitFor/Refresh 语义与 NotFound 处理符合规范：`docs/wait_for_state.md`、`alicloud/service_alicloud_sls_logtail_config.go`、`alicloud/service_alicloud_sls_logtail_pipeline_config.go`
- [x] T032 复核分层与强类型约束符合宪章并记录结论：`.specify/memory/constitution.md`、`docs/development_guide.md`
- [x] T033 更新 Feature 010 任务拆分关键变化备注（若无变更则显式记录无新增 Feature）：`.specify/memory/features/010.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1（Setup）→ Phase 2（Foundational）→ Phase 3/4/5（User Stories）→ Phase 6（Polish）
- Phase 2 是所有用户故事的阻塞前置。

### User Story Dependencies

- **US1 (P1)**: 依赖 Phase 2；无对其他故事的强依赖（MVP 核心）。
- **US2 (P1)**: 依赖 Phase 2；建议在 US1 稳定后并行推进。
- **US3 (P2)**: 依赖 US1 + US2（需要双资源都可用后再验证并存治理）。

### Within Each User Story

- 先测试任务，再实现任务。
- 先 schema/映射，再 CRUD/状态一致性。
- 文档与发布说明在故事功能稳定后补齐。

---

## Parallel Opportunities

- Setup 并行：T003、T004。
- Foundational 并行：T007、T008（不同文件边界）。
- US1 并行：T011、T012。
- US2 并行：T017、T018。
- US3 并行：T023、T024。
- Polish 并行：T029、T030。

---

## Parallel Example: User Story 2

- Task: T017 [US2] `alicloud/resource_alicloud_logtail_pipeline_config_test.go`
- Task: T018 [US2] `alicloud/sls_logtail_pipeline_config_mapping_test.go`

---

## Implementation Strategy

### MVP First（仅 US1）

1. 完成 Phase 1 + Phase 2。
2. 完成 US1（T011-T016）。
3. 立刻按 quickstart 做独立验证，确保旧资源升级不阻断。

### Incremental Delivery

1. 在 MVP 稳定后交付 US2（新版独立资源）。
2. 再交付 US3（并存治理 + 迁移文档）。
3. 最后执行 Polish 质量门禁并准备发布。

### Suggested MVP Scope

- 建议 MVP 范围：**Phase 1 + Phase 2 + US1（T001-T016）**。
- 该范围已覆盖最高业务风险（存量升级阻断）。
