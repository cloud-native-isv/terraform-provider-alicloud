# Data Model — 自动接管已存在数据库（alicloud_db_database）

本文件抽取 Feature Spec 的实体、字段、关系与校验规则，并描述关键状态转移。

## Entities

### 1) Database（数据库）
- Identity
  - `instance_id` (string, required): 目标数据库实例 ID。
  - `name` (string, required): 数据库名，正则 `^[a-z][a-z0-9_-]*[a-z0-9]$`。
  - `id` (string, computed): 资源唯一标识 `instanceId:dbName`。
- Attributes
  - `character_set` (string, optional, ForceNew): 
    - MySQL/SQLServer：单值，如 `utf8`。
    - PostgreSQL：复合值 `CharacterSetName,Collate,Ctype`（与现有 Read 行为兼容）。
  - `description` (string, optional): 描述信息。
- Provider-Meta（Plan 透明性）
  - `adopt_existing` (bool, Optional+Computed): Plan 阶段提示将接管现有数据库。
  - `adoption_notice` (string, Optional+Computed): 面向用户的友好提示语，Plan/Apply 均可见。

### 2) Desired Configuration（期望配置）
- 由用户在 Terraform HCL 中声明：`instance_id`, `name`, `character_set?`, `description?`。
- 用于 Create/Update 与接管时差异校验。

### 3) Adoption Outcome（接管结果）
- `adopted` (bool): 是否发生接管。
- `adopted_at` (timestamp): 接管时间（可在日志与提示中体现；不持久化到 state）。
- `diff_summary` (string): 差异摘要（只读提示用；不作为 state 持久字段）。

## Validation Rules

- `name`：必须匹配正则；遵循现有校验。
- `character_set`：
  - MySQL/SQLServer：大小写不敏感；ForceNew。
  - PostgreSQL：`Charset,Collate,Ctype` 三元组 ForceNew。
- 不可变字段冲突（接管时）：失败 + 指引（FR-003）。
- 可变字段（如 `description`）在接管发生的同一轮不自动对齐（FR-004）。

## Relationships

- Database ∈ Instance（同一实例内按 `name` 唯一）。

## State Transitions

- Create Path
  - 输入 Desired Configuration → Service 层查询是否已存在：
    - 存在 → 设置 `id`，标记 adopt；进入 Read 同步状态（不修改可变字段）。
    - 不存在 → 调用 Create；WaitFor Ready；进入 Read 同步状态。
- Update Path
  - `description` 变更 → 调用 Modify Description；等待稳定（若需要）；Read。
- Delete Path
  - WaitFor Instance Running → Delete；WaitFor Deleted。

## Notes

- 仅在 Plan 中做只读探测；失败时降级提示但不让 Plan 失败。
- Apply 中遵循重试与等待规范，保障确定性。