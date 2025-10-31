# Phase 0 Research — 自动接管已存在数据库（alicloud_db_database）

本研究旨在解决 Technical Context 中的 NEEDS CLARIFICATION 并给出明确抉择，形成可实施的设计基线。

## 1) 不可变字段清单与策略（按引擎）

- Decision:
  - MySQL: `character_set` 属于不可变字段（创建后不可在线更改）。
  - PostgreSQL: `character_set`（包含 Collate、Ctype 组合）视为不可变集合；按当前实现读法为 `CharacterSetName,Collate,Ctype`。
  - SQL Server: 传入非本地引擎字符集时 SDK 可能回退默认字符集；仍视字符集为不可变（按现有行为保守处理）。
  - 接管场景下，若用户配置与现有数据库在上述不可变集合存在差异：采取“失败 + 指引”保守策略（FR-003）。

- Rationale:
  - 云数据库主流约束与现有 Provider 行为一致；误改风险高，保守策略可保障数据安全。

- Alternatives considered:
  - 允许在接管同一轮尝试对齐不可变字段：不可行（需要重建数据库，风险高且与接管目标冲突）。
  - 忽略差异：不可行（用户预期不明确，破坏基础设施即代码的确定性）。

## 2) Plan 阶段透明性（如何在 Plan 呈现“将接管”）

- Decision:
  - 采用 `CustomizeDiff` 执行只读存在性探测（调用 Service 层 Describe；遵守限流/超时），并设置一个“只读提示字段”以在 Plan 中清晰展示。
  - 字段方案：新增 Optional+Computed 字段 `adoption_notice`（string），仅用于展示计划提示；不会在 Apply 触发实际操作；Read 时也会按最终状态给出稳定值。
  - 同时新增 Optional+Computed 布尔字段 `adopt_existing` 作为机器可读标志，便于用户在 CI 中断言计划行为。

- Rationale:
  - Terraform v1 SDK 的 `CustomizeDiff` 支持基于远端只读检查调整 Diff；通过只读提示字段可在 Plan 层满足 FR-009 的“可见性”要求。

- Alternatives considered:
  - 仅在 Apply 输出 log：不满足 FR-009 在 Plan 阶段的可见性需求。
  - 使用临时本地文件或外部输出：与 Terraform 工作流不契合。

## 3) 错误与重试清单（接管/创建流程）

- Decision:
  - 重试类错误：`ServiceUnavailable`, `ThrottlingException`, `InternalError`, `Throttling`, `SystemBusy`, `OperationConflict`（指数退避，最大持续不超过 Create Timeout）。
  - 不存在/已存在判断：优先使用封装判断 `IsNotFoundError` / `IsAlreadyExistError`，并结合 RDS 特定错误（例如 `InvalidDBName.NotFound`）。
  - 实例状态限制：若实例非 Running，等待至 Running（复用现有 `WaitForDBInstance`）。

- Rationale:
  - 与宪法“错误处理与重试规范”一致，复用既有等待与封装，降低实现复杂度与风险。

- Alternatives considered:
  - 将全部错误上抛：会引入偶发失败，降低用户体验。

## 4) 权限最小集合与提示

- Decision:
  - 最小权限：列举/查询数据库（Describe/List Databases）与创建/删除数据库的 API 权限；Plan 只读探测仅需读取权限。
  - 当资源存在但权限不足：在 Plan/Apply 中给出明确提示，含需要的最小权限与下一步指引（FR-006）。

- Rationale:
  - 准确指示可减少支持成本与反复试错。

- Alternatives considered:
  - 吞掉权限错误：误导性强，不可取。

## 5) 速率限制与健壮性

- Decision:
  - Plan 阶段 CustomizeDiff 探测加节流与最小化请求，失败时降级为“无法确认是否接管”的友好提示，避免计划失败。
  - Apply 阶段严格遵守重试与等待，保障“接管或创建”的确定性。

- Rationale:
  - Plan 应尽量保持可运行与稳定，不因暂时性网络/限流导致失败。

- Alternatives considered:
  - Plan 探测失败即失败整个 Plan：对用户过于严格。

## 6) ID 与状态编码

- Decision:
  - 继续沿用 `instanceId:dbName` 的复合 Id（与现有实现一致）；在 Service 层保留/补充 `Encode*Id` 与 `Decode*Id`，并在 Read/Wait 函数中一致使用。

- Rationale:
  - 与现有资源与工具函数兼容，降低变更面。

- Alternatives considered:
  - 引入更复杂的结构化 Id：无显著收益。

---

## 结论

- 所有 NEEDS CLARIFICATION 已给出明确方案：
  - 不可变字段：按引擎定义，冲突即“失败 + 指引”。
  - Plan 透明性：CustomizeDiff 只读探测 + `adoption_notice`（string）与 `adopt_existing`（bool）。
  - 错误/重试与权限提示：遵循宪法规范，覆盖常见错误码与最小权限说明。

- 接下来可进入 Phase 1 设计与合约/数据模型输出。