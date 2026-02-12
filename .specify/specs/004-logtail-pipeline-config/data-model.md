# Data Model: 004-logtail-pipeline-config

**Date**: 2026-02-12  
**Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](requirements.md)

## Entities

### LogtailPipelineConfig (Terraform Domain)

**描述**: Terraform 资源 `alicloud_logtail_config` 所管理的“采集/处理/输出”流水线配置。

**字段**:

- `project` (string, required): SLS Project 名称。
- `name` (string, required): 配置名称；在同一 Project 内唯一；创建后不可变更。
- `inputs` (list<Plugin>, required, min 1): 输入插件链。
- `processors` (list<Plugin>, optional): 处理插件链。
- `flushers` (list<Plugin>, required, min 1): 输出插件链。
- `aggregators` (list<Plugin>, optional): 聚合插件链。
- `global_json` (string, optional): 全局设置（JSON）。
- `task_json` (string, optional): 任务设置（JSON）。
- `log_sample` (string, optional): 样例日志。
- `create_time` (int64, read-only): 创建时间。
- `last_modify_time` (int64, read-only): 最后修改时间。

**校验规则**:

- `project`、`name` 必填。
- `inputs`、`flushers` 至少 1 个。
- `*_json` 与 `config_json` 必须是合法 JSON（若提供）。

**规范化规则**:

- 所有 JSON 字段进入 state 前做 canonical normalization。
- 插件链顺序保持输入顺序（不排序）。

---

### Plugin

**描述**: 流水线组件（inputs/processors/flushers/aggregators）的单个插件。

**字段**:

- `type` (string, required): 插件类型标识。
- `config_json` (string, optional): 插件配置明细 JSON。

**约束**:

- `type` 必填。
- `config_json` 若为空，表示使用插件默认配置或仅由 `type` 决定行为（以远端能力为准）。

## Relationships

- 一个 `LogtailPipelineConfig` 归属一个 `project`。
- 插件链按顺序执行；因此列表顺序具有语义。

## State Transitions

- Create: 配置创建成功后可被读取与更新。
- Update: 配置内容更新并反映在 `last_modify_time`。
- Delete: 删除后读取应表现为资源不存在。
