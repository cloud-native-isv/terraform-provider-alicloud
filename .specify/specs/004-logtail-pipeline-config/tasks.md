# Tasks: 004-logtail-pipeline-config（SLS Logtail Pipeline Config Migration）

**Input**: 设计文档位于 `.specify/specs/004-logtail-pipeline-config/`（plan.md, requirements.md, research.md, data-model.md, contracts/, quickstart.md）

**范围**: 以 `alicloud_logtail_config` 资源为中心，迁移到新版 Logtail Pipeline Config API，并重做 schema（不保证向前兼容）。

---

## Phase 1: Setup（准备与对齐）

- [x] T001 复核需求与验收场景：`.specify/specs/004-logtail-pipeline-config/requirements.md`
- [x] T002 复核实现计划与约束（分层/强类型/等待）：`.specify/specs/004-logtail-pipeline-config/plan.md`
- [x] T003 复核数据模型与 schema/service 合约：`.specify/specs/004-logtail-pipeline-config/data-model.md`、`.specify/specs/004-logtail-pipeline-config/contracts/terraform-schema.md`、`.specify/specs/004-logtail-pipeline-config/contracts/service-layer.md`
- [x] T004 复核 Feature 010 状态与更新时间（如需则更新）：`.specify/memory/features.md`、`.specify/memory/features/010.md`

---

## Phase 2: Foundational（阻塞性基础能力）

> 目标：在不触碰 Terraform 资源 CRUD 细节前，先把“强类型领域模型 + 映射边界 + Service 新 API 封装 + JSON 规范化”准备好。

- [x] T005 [P] 定义 Terraform 领域模型与插件结构体（强类型外壳）：`alicloud/sls_logtail_pipeline_config_model.go`
- [x] T006 [P] 实现 JSON 规范化与解析辅助（复用/封装 `normalizeJsonString`，补齐 object-only 校验等）：`alicloud/sls_logtail_pipeline_config_json.go`
- [x] T007 [P] 实现 Terraform 领域模型 ↔ cws-lib-go `slsAPI.LogtailPipelineConfig` 的双向转换（弱结构仅限转换边界内部）：`alicloud/sls_logtail_pipeline_config_mapping.go`
- [x] T008 [P] 为映射与 JSON 规范化编写单元测试（先写测试，确保在实现前失败）：`alicloud/sls_logtail_pipeline_config_mapping_test.go`
- [x] T009 新增/封装 Pipeline Config 的 Service 层方法（Get/Create/Update/Delete + RefreshFunc），并保持 NotFound 语义符合合约：`alicloud/service_alicloud_sls_logtail_pipeline_config.go`
- [x] T010 为 Service 层 NotFound/等待语义与 ID 解析编写单元测试：`alicloud/service_alicloud_sls_logtail_pipeline_config_test.go`
- [x] T011 确认现有旧版 LogtailConfig Service/数据源不被破坏（如新增新文件则确保无编译冲突）：`alicloud/service_alicloud_sls_logtail_config.go`、`alicloud/data_source_alicloud_sls_logtail_config.go`

**Checkpoint**: 基础层完成后，US1 的资源实现可直接基于强类型模型与 Service 方法落地。

---

## Phase 3: User Story 1 - 管理 Logtail Pipeline Config（Priority: P1）🎯 MVP

**Goal**: `alicloud_logtail_config` 支持新版 Pipeline Config 的 Create/Read/Update/Delete + Import。

**Independent Test**: 参见 `.specify/specs/004-logtail-pipeline-config/requirements.md`（全新 Project/Logstore：apply 成功；read 一致；import 后 plan 无差异；update 生效；destroy 成功）。

### Tests for US1（先写测试）

- [x] T012 [P] [US1] 添加资源验收测试骨架（可先 skip，但覆盖 CRUD + Import 场景与配置生成）：`alicloud/resource_alicloud_logtail_config_test.go`
- [x] T013 [P] [US1] 添加资源层展开/回填（expand/flatten）关键路径单测（不依赖真实云端）：`alicloud/resource_alicloud_logtail_config_mapping_test.go`

### Implementation for US1

- [x] T014 [US1] 重做资源 schema 为“插件链 + JSON 字段 + 时间字段”（不兼容旧 schema）：`alicloud/resource_alicloud_logtail_config.go`
- [x] T015 [US1] 实现 Create：从 schema expand 为领域模型，调用 `CreateLogtailPipelineConfig`，并使用等待/刷新确保可读：`alicloud/resource_alicloud_logtail_config.go`
- [x] T016 [US1] 实现 Read：调用 `GetLogtailPipelineConfig`，按合约回填 state（含 JSON canonical normalization）：`alicloud/resource_alicloud_logtail_config.go`
- [x] T017 [US1] 实现 Update：对可更新字段做原地更新（其余 ForceNew），并在更新后等待读回一致：`alicloud/resource_alicloud_logtail_config.go`
- [x] T018 [US1] 实现 Delete：调用 `DeleteLogtailPipelineConfig`，NotFound 视为成功，并等待资源消失：`alicloud/resource_alicloud_logtail_config.go`
- [x] T019 [US1] 实现 Import：ID 采用 `<project>:config:<name>` 并确保 Read 可完整回填（必要时补齐 `project`/`name` 的解析）：`alicloud/resource_alicloud_logtail_config.go`

**Checkpoint**: US1 完成后，资源应具备完整生命周期能力，且 Import → plan 无差异（在规范化完成前可能仍需 US2 增强）。

---

## Phase 4: User Story 2 - 清晰的 Schema 与可预期的 Diff（Priority: P2）

**Goal**: schema 表达清晰，JSON 与可选字段规范化后避免永久 diff；连续两次 plan 无差异。

**Independent Test**: 按 `.specify/specs/004-logtail-pipeline-config/quickstart.md` 重复执行两次 `terraform plan` 输出应为无变更。

### Tests for US2（先写测试）

- [x] T020 [P] [US2] 补齐 JSON 规范化边界测试（空字符串/空对象/排序/空白/数组顺序保持）：`alicloud/sls_logtail_pipeline_config_json_test.go`
- [ ] T021 [P] [US2] 补齐 Read 回填稳定性测试（同一远端对象多次 flatten 输出必须一致）：`alicloud/resource_alicloud_logtail_config_mapping_test.go`

### Implementation for US2

- [x] T022 [US2] 为所有 JSON 字段（`global_json`/`task_json`/`config_json`）添加 `StateFunc` + JSON 校验，确保入库即规范化：`alicloud/resource_alicloud_logtail_config.go`
- [x] T023 [US2] 避免“读回默认值导致 drift”：Read 回填时仅在远端存在该字段时写入；空值策略保持一致：`alicloud/resource_alicloud_logtail_config.go`
- [x] T024 [US2] 避免“空列表/空 block”导致 diff：对 `processors`/`aggregators` 的 absent vs empty 行为做一致化处理（必要时在 flatten 中返回 nil）：`alicloud/resource_alicloud_logtail_config.go`、`alicloud/sls_logtail_pipeline_config_mapping.go`
- [ ] T025 [US2] 如仍存在等价 JSON diff，补充 `DiffSuppressFunc`（仅限等价场景；不改变插件链顺序语义）：`alicloud/resource_alicloud_logtail_config.go`
- [x] T026 [US2] 手工验证无永久 diff（按 quickstart 步骤记录结果/截图链接可放入本 spec 目录）：`.specify/specs/004-logtail-pipeline-config/quickstart.md`

**Checkpoint**: 在无变更情况下，连续两次 plan 不产生差异；Import 后 plan 不产生差异。

---

## Phase 5: User Story 3 - 仅使用新版能力并可验证（Priority: P3）

**Goal**: 本资源与其新 Service 层仅使用新版 Logtail Pipeline Config API；可通过检索与测试验证。

**Independent Test**: 在资源相关文件范围内检索不再出现 `GetLogtailConfig/CreateLogtailConfig/UpdateLogtailConfig/DeleteLogtailConfig`。

- [ ] T027 [P] [US3] 清理资源层对旧类型/旧字段的残留引用（例如 `input_type`/`output_type`/`input_detail`/`output_detail`）：`alicloud/resource_alicloud_logtail_config.go`
- [x] T028 [P] [US3] 确认新 Service 层仅调用 `Get/Create/Update/DeleteLogtailPipelineConfig`（必要时为旧能力保留独立文件，避免混用）：`alicloud/service_alicloud_sls_logtail_pipeline_config.go`、`alicloud/service_alicloud_sls_logtail_config.go`
- [x] T029 [P] [US3] 增加回归保护测试：对资源/新 Service 相关文件做关键字检索断言（或在 CI 脚本中加入 grep 检查）：`alicloud/service_alicloud_sls_logtail_pipeline_config_test.go`

---

## Phase 6: Polish & Cross-Cutting Concerns（收尾与质量门禁）

- [ ] T030 [P] 增加用户使用说明与示例（明确不兼容变更、新 schema 样例、Import ID 格式）：`docs/resources/alicloud_logtail_config.md`
- [x] T031 [P] 更新变更记录（注明 breaking change 与迁移到 Pipeline Config）：`CHANGELOG.md`
- [x] T032 运行并修复单元测试/编译（以 Makefile 为准）：`Makefile`
- [x] T033 复核等待/刷新实现符合项目规范（对照文档逐条检查）：`docs/wait_for_state.md`
- [x] T034 复核分层/强类型约束符合 Constitution（记录任何局部例外）：`.specify/memory/constitution.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup（Phase 1）→ Foundational（Phase 2）→ US1（Phase 3）→ US2（Phase 4）→ US3（Phase 5）→ Polish（Phase 6）

### User Story Dependencies

- US1（P1）依赖 Foundational（Phase 2）
- US2（P2）依赖 US1（至少 schema 与 CRUD 已落地）
- US3（P3）可与 US2 并行收尾，但建议在 US1 完成后尽早做检索与清理

---

## Parallel Execution Examples

### Foundational（Phase 2）

- 可并行：T005（model）+ T006（json helpers）+ T007（mapping）+ T008（mapping tests）

### US1（Phase 3）

- 可并行：T012（验收骨架）+ T013（mapping 单测）

---

## Implementation Strategy（MVP 优先）

1. Phase 1 + Phase 2：先把模型/映射/Service 新 API 与测试打好地基
2. Phase 3（US1）：完成 CRUD + Import（MVP）
3. **暂停验证**：按 `.specify/specs/004-logtail-pipeline-config/quickstart.md` 做 smoke（重点：Import 与 plan 稳定性）
4. Phase 4（US2）：消除永久 diff 与空值/默认值漂移
5. Phase 5（US3）：确保仅用新版能力并可验证
6. Phase 6：文档/Changelog/质量门禁
