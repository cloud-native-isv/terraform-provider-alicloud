# Data Model: 006-unify-log-prefix

## 1. NamingMappingItem（命名映射项）

表示一个历史 `alicloud_sls_*` 入口与目标 `alicloud_log_*` 入口之间的关系。

### Fields

- `capability_key` (string, required)
  - 能力唯一标识（如 `scheduled_sql`、`etl`、`project`）
- `legacy_name` (string, required)
  - 历史名称，格式必须匹配 `^alicloud_sls_[a-z0-9_]+$`
- `canonical_name` (string, required)
  - 统一名称，格式必须匹配 `^alicloud_log_[a-z0-9_]+$`
- `mapping_type` (enum, required)
  - `one_to_one` | `many_to_one` | `retired`
- `status` (enum, required)
  - `unified` | `pending` | `retired_with_note`
- `replacement_hint` (string, required)
  - 用户可直接执行的替代提示
- `effective_version` (string, required)
  - 生效版本标识（最近发布版本）
- `updated_at` (datetime, required)

### Validation Rules

- `legacy_name` 必须唯一。
- `canonical_name` 在 `mapping_type=one_to_one` 时必须唯一。
- 若 `mapping_type=retired`，必须填写 `replacement_hint` 或下线说明。

---

## 2. GovernanceCatalog（治理清单）

表示统一治理范围与阶段状态的集合视图。

### Fields

- `catalog_id` (string, required)
- `scope_rule` (string, required)
  - 固定值：`all alicloud_sls_* included`
- `included_items` (array[NamingMappingItem], required)
- `excluded_items` (array[string], required)
  - 本需求下通常为空
- `pending_items` (array[NamingMappingItem], required)
- `release_note_ref` (string, required)
  - 发布说明引用
- `last_reviewed_at` (datetime, required)

### Validation Rules

- `scope_rule` 不可变更为其他策略。
- 所有识别到的 `alicloud_sls_*` 项必须在 `included_items` 或 `pending_items` 出现。

---

## 3. MigrationGuidanceItem（迁移指引项）

描述一个历史入口迁移到统一入口的执行步骤。

### Fields

- `legacy_name` (string, required)
- `canonical_name` (string, required)
- `strategy` (enum, required)
  - 固定值：`destroy_recreate`
- `preconditions` (array[string], required)
- `execution_steps` (array[string], required)
- `expected_result` (string, required)
- `risk_notice` (array[string], required)
- `rollback_guidance` (array[string], required)

### Validation Rules

- `strategy` 只能是 `destroy_recreate`。
- `execution_steps` 至少包含：删除旧资源定义、替换命名、重建、验证。

---

## State Transitions

### NamingMappingItem.status

- `pending -> unified`
  - 条件：`canonical_name` 已注册可用，文档与发布说明已同步。
- `pending -> retired_with_note`
  - 条件：确认无直接替代，已提供下线说明。
- `unified -> retired_with_note`
  - 条件：能力整体下线，发布说明已生效。

### GovernanceCatalog

- 每次发布前必须执行一致性检查：
  - `count(alicloud_sls_*) == count(included_items) + count(pending_items)`

---

## Relationship Overview

- 一个 `GovernanceCatalog` 包含多个 `NamingMappingItem`。
- 一个 `NamingMappingItem` 对应一个 `MigrationGuidanceItem`（`retired` 可仅保留下线说明）。
