---
description: "Tasks for feature 006-support-exist-resource: 自动接管已存在数据库（alicloud_db_database）"
---

# Tasks: 自动接管已存在数据库（alicloud_db_database）

**Input**: Design documents from `/.specify/specs/006-support-exist-resource/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 未在本功能规范中强制要求测试任务；如需 TDD，请在后续迭代中补充对应测试任务。

**Organization**: 任务严格按用户故事组织，保证每个故事可独立实现与验证。

## Format: `[ID] [P?] [Story] Description`
- **[P]**: 可并行（不同文件且无直接依赖）
- **[Story]**: 任务所属用户故事（US1/US2/US3）
- 任务描述中包含确切文件路径（仓库根相对路径）

---

## Phase 1: Setup (Shared Infrastructure)

目的：项目初始化与基础校验，确保开发环境与依赖就绪。

- [X] T001 进行基线编译校验：在仓库根执行 `make`，确认当前分支可编译通过（不修改代码）。
- [X] T002 [P] 校验依赖库可用：确认 go.mod 中已包含并可解析 `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api`（cws-lib-go）。若缺失则补充，确保 `make` 通过。
- [X] T003 [P] 建立任务追踪文件：在 `/.specify/specs/006-support-exist-resource/` 中确认存在 `plan.md`、`spec.md`、`data-model.md`、`research.md`、`contracts/openapi.yaml`、`quickstart.md`（脚本已识别）。

---

## Phase 2: Foundational (Blocking Prerequisites)

目的：完成所有用户故事的前置阻塞能力（统一 Service 层、等待器、错误与ID编码）。本阶段完成前，不可开始任一用户故事实现。

- [X] T004 在 Service 层新增/补充 RDS Database 能力：创建文件（若更清晰拆分）`alicloud/service_alicloud_rds_database.go`，封装以下方法（使用 cws-lib-go）：
  - DescribeDBDatabase(id string) (map[string]interface{}, error) — 若已存在
  - CreateDBDatabase(req) (*DBDatabase, error)
  - ModifyDBDatabaseDescription(id string, desc string) error
  - DeleteDBDatabase(id string) error
  - 说明：沿用已有 `service_alicloud_rds_base.go` 的 `NewRdsService` 与 API 对接模式
- [X] T005 实现分页/列举封装（如需要）：在 `alicloud/service_alicloud_rds_database.go` 中为“实例内按名称查询”提供 Describe 封装，隐藏分页细节。
- [X] T006 增加 ID 编解码：在 `alicloud/service_alicloud_rds_database.go` 中实现 `EncodeDBId(instanceId, dbName string) string` 与 `DecodeDBId(id string) (instanceId, dbName string, err error)`，格式 `instanceId:dbName`。
- [X] T007 增加状态刷新与等待器：在 `alicloud/service_alicloud_rds_database.go` 中实现 `DBDatabaseStateRefreshFunc(id string, failStates []string)` 与 `WaitForDBDatabaseCreating(id string, timeout time.Duration)`、`WaitForDBDatabaseDeleted(id string, timeout time.Duration)`；遵循宪法中的 `BuildStateConf` 模式。
- [X] T008 统一错误处理：在 Service 层方法中使用 `WrapError/WrapErrorf`，并优先采用 `IsNotFoundError/IsAlreadyExistError/NeedRetry` 与通用可重试错误清单（Throttling、ServiceUnavailable、SystemBusy、OperationConflict 等）。
- [X] T009 [P] 更新/补充 `alicloud/service_alicloud_rds_base.go` 的注释与导出接口，确保 `GetAPI()` 能返回所需 RDS API 客户端用于内部封装调用（保持类型安全）。
- [X] T010 基线编译校验：执行 `make`，确保新增 Service 层代码编译通过。

**Checkpoint**: Service 层能力齐备、等待器就绪、错误与 ID 编码规范统一。

---

## Phase 3: User Story 1 - 已存在数据库应被自动接管（Priority: P1）🎯 MVP

目标：当实例内存在同名数据库时，Create 路径自动接管而非重复创建；Plan 阶段可见“将接管”提示。

独立测试标准：准备一个实例，预先创建 `name = X` 的数据库。Terraform 声明同名资源后：
- Plan 输出明确标注“将接管”提示（不失败）；
- Apply 后 Resource 进入受管状态，未创建重复对象；
- 再次 Plan/Apply 为 No-op。

### 实现任务（US1）

- [X] T011 [US1] 在资源 Schema 新增只读提示字段（Optional+Computed），文件：`alicloud/resource_alicloud_db_database.go`
  - 新增 `adopt_existing` (bool, Optional+Computed, Description: "Whether the provider will adopt an existing database on apply.")
  - 新增 `adoption_notice` (string, Optional+Computed, Description: "Human-readable notice about adoption behavior shown at plan/apply.")
  - 补充各字段 Description，遵循命名与文档规范。
- [X] T012 [US1] 增加 CustomizeDiff（只读探测），文件：`alicloud/resource_alicloud_db_database.go`
  - 在 CustomizeDiff 中调用 Service 层 Describe（节流+错误兜底）
  - 若探测到已存在：设置 `adopt_existing = true` 与 `adoption_notice` 提示；若权限不足则给出友好提示但不失败 Plan（降级策略）
- [X] T013 [US1] 重构 Create：替换 `client.RpcPost("Rds", ...)` 为 Service 层调用，文件：`alicloud/resource_alicloud_db_database.go`
  - 先 Describe：若存在 → 直接 `d.SetId(EncodeDBId(instanceId, name))` 并进入状态对齐（调用 Read 最终同步）
  - 若不存在 → 调用 Service 层 Create；`WaitForDBDatabaseCreating`；最后 Read 同步
  - 统一错误与重试，去除直接 `RpcPost`
- [X] T014 [US1] Delete 等待优化：文件：`alicloud/resource_alicloud_db_database.go`
  - 将 Delete 路径替换为 Service 层 Delete + `WaitForDBDatabaseDeleted`，保留实例 Running 前置等待（如已有则复用 Service 层）
- [X] T015 [P] [US1] 记录 Adoption 日志与用户提示：在 Create adopt 分支添加清晰日志，Read 同步 `adoption_notice` 字段稳定值。
- [X] T016 [US1] 编译与本地验收：执行 `make`；按 `/.specify/specs/006-support-exist-resource/quickstart.md` 验证 Plan/Apply 行为。

**Checkpoint**: US1 完成，具备自动接管与 Plan 可见提示，幂等生效。

---

## Phase 4: User Story 2 - 配置与现状差异的处理（Priority: P2）

目标：对不可变字段冲突采“失败+指引”；对可变字段在接管同轮不自动对齐，仅提示后续操作。

独立测试标准：制造 `character_set` 等不可变冲突场景 → Apply 失败并给出指引；制造 `description` 差异 → 接管成功但不修改描述，提示可后续对齐。

### 实现任务（US2）

- [X] T017 [US2] 在 Create adopt 分支进行不可变字段冲突校验，文件：`alicloud/resource_alicloud_db_database.go`
  - 基于 `research.md` 引擎规则（MySQL/PG/SQLServer）判断 `character_set` 等不可变集合
  - 冲突则 WrapErrorf 明确错误、指出冲突字段并提供建议（保守策略）
- [X] T018 [US2] 保持接管轮不自动对齐可变字段（如 `description`），文件：`alicloud/resource_alicloud_db_database.go`
  - 在 adopt 分支不修改描述，仅在日志与 `adoption_notice` 给予“可在后续变更中对齐”的提示
- [X] T019 [US2] 将 Update 路径改造为 Service 层 `ModifyDBDatabaseDescription` 调用，必要时增加等待，文件：`alicloud/resource_alicloud_db_database.go`
- [X] T020 [US2] 编译与本地验收：`make`，并按 US2 场景验证。

**Checkpoint**: US2 完成，冲突策略与可变字段约束生效。

---

## Phase 5: User Story 3 - 权限与可见性（Priority: P3）

目标：在权限不足时给出一致的、可操作的提示；Plan 探测降级但不中断。

独立测试标准：使用受限权限账号 Plan/Apply，Plan 显示提示不失败；Apply 阶段指示所需最小权限。

### 实现任务（US3）

- [X] T021 [US3] 权限错误映射与提示优化：在 Service 层对常见权限错误（如 `Forbidden.*`）进行识别，文件：`alicloud/service_alicloud_rds_database.go`
  - 在资源层据此生成明确的用户指引（最小权限说明见 `research.md`）
- [X] T022 [US3] CustomizeDiff 探测降级：读取失败（权限/限流）时设置 `adoption_notice` 说明“无法确认是否接管”，但不失败 Plan，文件：`alicloud/resource_alicloud_db_database.go`
- [X] T023 [US3] 编译与本地验收：`make` + 受限权限验证。

**Checkpoint**: US3 完成，权限与可见性体验一致。

---

## Phase N: Polish & Cross-Cutting Concerns

- [X] T024 [P] 文档更新：在 `/.specify/specs/006-support-exist-resource/quickstart.md` 校对示例字段（adopt_existing/adoption_notice）与行为说明。
- [X] T025 [P] 如存在资源文档（`docs/` 下 `alicloud_db_database` 对应文档），补充“自动接管”章节与字段解释。
- [X] T026 代码清理：移除遗留的 `client.RpcPost` 直接调用路径，确保统一从 Service 层访问。
- [X] T027 编译校验与小型冒烟：执行 `make`，对 US1/US2/US3 场景进行一次最小化冒烟验证。

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1)：无依赖，可立即开始。
- Foundational (Phase 2)：依赖 Setup；阻塞所有用户故事。
- User Stories (Phase 3+)：均依赖 Foundational 完成；之后可按优先级顺序或并行推进（US1 → US2 → US3）。
- Polish (Final Phase)：依赖所需用户故事完成。

### User Story Dependencies

- User Story 1 (P1)：仅依赖 Foundational。
- User Story 2 (P2)：依赖 US1（共用接管路径的基本能力），但逻辑可独立验证。
- User Story 3 (P3)：依赖 Foundational；与 US1/US2 仅有提示与映射层面的松耦合集成。

### Within Each User Story

- Schema → CustomizeDiff → Service 调用整合（Create/Update/Delete）→ 等待与错误处理 → Read 同步。
- 完成故事后再进入下一故事；若并行开发，需注意文件冲突合并顺序。

### Parallel Opportunities

- Setup 阶段的 T002/T003 可并行。
- Foundational 阶段中：
  - T009 可并行于其他 Service 实现（注释/接口声明不影响逻辑），其余任务建议串行以降低 API 对接风险。
- US1 中：T015 日志与提示可并行（不影响主要逻辑）；其余强依赖串行。
- Polish 阶段的文档与清理可并行（T024/T025）。

---

## Parallel Example: User Story 1

```bash
# 在 US1 中可并行的任务（示例）：
# 1) 记录 adoption 日志/提示（与核心逻辑相对独立）
Task: T015 [P] [US1] 记录 Adoption 日志与用户提示

# 2) 与主要逻辑串行的任务（需按顺序）
Task: T011 [US1] Schema 字段新增
Task: T012 [US1] CustomizeDiff 只读探测
Task: T013 [US1] Create 路径重构（接管/创建）
Task: T014 [US1] Delete 等待优化
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1（Setup）
2. 完成 Phase 2（Foundational）— 阻塞所有故事
3. 完成 Phase 3（US1）— 自动接管与 Plan 提示
4. STOP & VALIDATE：按 quickstart.md 完成独立验证
5. 若准备就绪即可合入/发布（受控）

### Incremental Delivery

1. 完成 Setup + Foundational
2. US1 → 独立验证 → 发布（MVP）
3. US2 → 独立验证 → 发布
4. US3 → 独立验证 → 发布

### Parallel Team Strategy

- Foundational 阶段由小组协作完成（减少 API/抽象分散实现风险）。
- Foundational 完成后：
  - 开发者 A：US1（核心接管与 Plan）
  - 开发者 B：US2（冲突策略与 Update 重构）
  - 开发者 C：US3（权限映射与可见性）

---

## Notes

- [P] 任务 = 不同文件且无直接依赖。
- 每个用户故事应可独立完成并验证。
- 统一错误处理/日志/等待规范，遵循项目宪法。
- 任何新增公共方法需添加合适注释与 Description 文案。
- 每一阶段结束务必执行 `make` 确认通过。
