# Contracts: Terraform Schema (alicloud_logtail_config)

**Date**: 2026-02-12  
**Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](../requirements.md)

## Schema Overview (Proposed)

- `project` (string, required, ForceNew)
- `name` (string, required, ForceNew)
- `inputs` (list(block), required, MinItems=1)
- `processors` (list(block), optional)
- `flushers` (list(block), required, MinItems=1)
- `aggregators` (list(block), optional)
- `global_json` (string, optional, JSON)
- `task_json` (string, optional, JSON)
- `log_sample` (string, optional)
- `create_time` (int, computed)
- `last_modify_time` (int, computed)

Plugin block:
- `type` (string, required)
- `config_json` (string, optional, JSON)

## Validation Contract

- `project`、`name` 不能为空。
- `inputs` 与 `flushers` 至少 1 个。
- `global_json`、`task_json`、`config_json`（若提供）必须是合法 JSON。

## Diff / Normalization Contract

- 对所有 JSON 字段进行 canonical normalization，避免空白与 key 顺序导致的永久 diff。
- 不对插件链列表排序；顺序视为有语义。

## Read Mapping Contract

- Read 时从远端返回的插件配置 map 序列化为 JSON 字符串，并在写入 state 前做 canonical normalization。
- `create_time`、`last_modify_time` 永远以远端为准。

## ID Contract

- 资源 ID 建议保持既有风格：`<project>:config:<name>`，用于稳定标识与 Import。
