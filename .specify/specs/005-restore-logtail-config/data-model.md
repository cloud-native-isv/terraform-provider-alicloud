# Data Model: 005-restore-logtail-config

**Date**: 2026-02-13  
**Spec**: [.specify/specs/005-restore-logtail-config/requirements.md](requirements.md)

## Entities

### 1) LegacyLogtailConfigResource

**Description**: 面向存量用户的旧资源实体，对应 Terraform 资源名 `alicloud_logtail_config`，目标是升级不破坏。

**Core Fields**:
- `project` (string, required, ForceNew)
- `name` (string, required, ForceNew)
- `logstore` / `input` / `filter` / `output` 等旧语义字段（按现有旧实现保持）
- 只读元数据字段（如创建时间、修改时间，若旧实现已提供）

**Validation Rules**:
- 维持旧实现校验行为，不新增破坏性必填项
- 升级后已有配置必须可被解析与读取

**Behavior Rules**:
- 兼容窗口内持续可用（至少 2 个小版本）
- 不新增运行时弃用告警

---

### 2) PipelineLogtailConfigResource

**Description**: 面向新版能力的资源实体，对应 Terraform 资源名 `alicloud_logtail_pipeline_config`。

**Core Fields**:
- `project` (string, required, ForceNew)
- `name` (string, required, ForceNew)
- pipeline 语义字段（如 `inputs`、`processors`、`flushers`、`aggregators`）
- 可选 JSON 扩展字段（如 `global_json`、`task_json`，按当前实现）

**Validation Rules**:
- 新资源字段按新版语义校验
- 用户误用旧语义字段时需返回可操作错误

**Behavior Rules**:
- 与旧资源并存，不复用旧资源名
- 支持完整生命周期（create/read/update/delete）

---

### 3) TerraformResourceState

**Description**: Terraform 状态实体，保存资源类型、实例 ID、属性快照。

**Core Fields**:
- `resource_type` (`alicloud_logtail_config` | `alicloud_logtail_pipeline_config`)
- `id` (string)
- `attributes` (map-like state data managed by Terraform core)

**State Consistency Rules**:
- 新旧资源可在同一 workspace 并存
- 当同时指向同一远端对象时，采用“最后一次成功写入生效”
- 后续读取结果必须能反映最近一次成功写入

## Relationships

- 一个远端 Logtail 配置对象可以被两个 Terraform 资源类型分别引用（并存场景）
- 每个 Terraform 资源实例在 state 中独立记录，互不覆盖

## State Transitions

### Legacy Resource
- `Create` -> `Readable`
- `Readable` -> `Updated`
- `Readable/Updated` -> `Deleted`
- 兼容窗口内不因版本升级进入不可识别状态

### Pipeline Resource
- `Create` -> `Readable`
- `Readable` -> `Updated`
- `Readable/Updated` -> `Deleted`

### Coexistence Transition
- `Legacy Updated` then `Pipeline Updated` => 远端状态以 Pipeline 最后写入为准
- `Pipeline Updated` then `Legacy Updated` => 远端状态以 Legacy 最后写入为准
