# Data Model: CMS Provider 封装能力改造

## Overview

本数据模型描述计划阶段需要追踪的设计对象，不代表新增运行时存储。所有实体用于指导后续任务拆分、代码审查和验收证据收集。

## Entity: CmsProviderFile

**Purpose**: 表示用户脚本枚举出的一个 Provider 文件，是范围判定和验收追踪的最小单位。

| Field | Type | Required | Description |
|---|---|---|---|
| `path` | string | yes | 文件路径，例如 `alicloud/resource_alicloud_cms_workspace.go` |
| `file_kind` | enum | yes | `service`, `resource`, `data_source` |
| `object_group` | enum | yes | `addon`, `agg_task`, `alert`, `cloud_resource`, `context`, `context_store`, `dataset`, `delivery_task`, `entity_store`, `integration_policy`, `memory`, `memory_store`, `pipeline`, `prometheus`, `service`, `umodel`, `workspace` |
| `current_state` | enum | yes | `placeholder`, `implemented_direct_call`, `implemented_service_layer`, `registered`, `unregistered` |
| `scope_decision` | enum | yes | `migrate_now`, `no_change_placeholder`, `enable_later`, `needs_api_gap_resolution` |
| `registration_name` | string | no | Provider 注册名称，如 `alicloud_cms_service` |
| `notes` | string | no | 兼容性、缺口或特殊风险说明 |

**Validation Rules**:

- `path` 必须属于用户脚本枚举出的 51 个文件之一。
- `file_kind=resource` 的文件只有在 schema、CRUD、服务能力和测试都存在时才能标记 `registered`。
- `file_kind=data_source` 的文件只有在 schema、read、服务能力和测试都存在时才能标记 `registered`。
- 占位文件必须显式标记 `no_change_placeholder` 或 `enable_later`，不得静默跳过。

## Entity: CmsObjectGroup

**Purpose**: 表示 CMS 业务对象分组，用于连接 Provider 文件、服务方法和 CWS-Lib-Go API wrapper。

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | enum | yes | CMS 对象组名称 |
| `api_package_area` | string | yes | 对应 API wrapper 文件前缀，如 `alicloud_cms_workspace_*` |
| `terraform_resource_name` | string | no | 启用时的资源名 |
| `terraform_data_source_name` | string | no | 启用时的数据源名 |
| `lifecycle_support` | set(enum) | yes | `create`, `read`, `update`, `delete`, `list`, `query`, `enable`, `disable`, `status_update` |
| `enabled_status` | enum | yes | `enabled`, `placeholder_only`, `planned_enablement`, `blocked_by_gap` |
| `compatibility_policy` | enum | yes | `preserve_existing`, `new_capability_requires_spec`, `internal_only` |

**Validation Rules**:

- 每个 `CmsObjectGroup` 必须映射到 3 个 `CmsProviderFile`（service/resource/data_source）。
- `enabled_status=enabled` 时必须有 Terraform 注册名和验证记录。
- `new_capability_requires_spec` 表示不能在本次内部重构中直接暴露新用户可见能力。

## Entity: CmsServiceCapability

**Purpose**: 描述 Provider 服务层可复用能力，是 Resource/DataSource 到 CWS-Lib-Go API 的唯一入口。

| Field | Type | Required | Description |
|---|---|---|---|
| `method_name` | string | yes | 服务方法名，例如 `OpenCmsService`、`DescribeCmsWorkspace` |
| `object_group` | enum | yes | 关联 CMS 对象组 |
| `operation` | enum | yes | `create`, `read`, `update`, `delete`, `list`, `query`, `wait`, `state_refresh`, `enable` |
| `input_type` | string | yes | 强类型输入或 documented dynamic exception |
| `output_type` | string | yes | 强类型输出或 documented dynamic exception |
| `api_wrapper_method` | string | yes | CWS-Lib-Go API 方法名 |
| `handles_pagination` | boolean | yes | 是否返回完整结果集 |
| `handles_not_found` | boolean | yes | 是否归一对象不存在语义 |
| `handles_retry` | boolean | yes | 是否封装重试或可重试错误 |
| `test_reference` | string | no | 对应测试名或验收记录 |

**Validation Rules**:

- 新增 Provider 服务方法不得以 `map[string]interface{}` 作为主要 request/response 类型。
- 需要动态结构时，`input_type` 或 `output_type` 必须包含 `documented dynamic exception` 并关联原因。
- `list`/`query` 能力必须确保调用方无需自行分页。

## Entity: CwsCmsApiWrapper

**Purpose**: 表示本地 CWS-Lib-Go CMS API wrapper 能力。

| Field | Type | Required | Description |
|---|---|---|---|
| `package_path` | string | yes | `pkg/cws-lib-go/lib/cloud/aliyun/api/cms` |
| `method_name` | string | yes | API wrapper 方法名 |
| `object_group` | enum | yes | CMS 对象组 |
| `sdk_action` | string | no | 底层 SDK action 名，如可从 wrapper 推导 |
| `typed_input` | string | yes | API 层输入类型 |
| `typed_output` | string | yes | API 层输出类型 |
| `pagination_boundary` | enum | yes | `api`, `service`, `not_applicable`, `needs_review` |
| `test_file` | string | no | API wrapper 测试文件 |

**Validation Rules**:

- Provider 层只能通过 `CmsServiceCapability` 间接使用该实体。
- 若 API wrapper 缺少所需操作，必须生成 `needs_api_gap_resolution` 任务，而不是在资源/数据源直接调用 SDK。

## Entity: TerraformOperationContract

**Purpose**: 表示用户可见 Terraform 行为与服务能力之间的契约。

| Field | Type | Required | Description |
|---|---|---|---|
| `terraform_name` | string | yes | 资源或数据源名称 |
| `operation` | enum | yes | `create`, `read`, `update`, `delete`, `import`, `list`, `enable` |
| `schema_fields` | list(string) | yes | 相关 schema 字段 |
| `service_method` | string | yes | 调用的服务层方法 |
| `state_effect` | string | yes | ID、Computed 字段或 state 清理行为 |
| `compatibility_requirement` | string | yes | 对既有行为的兼容承诺 |
| `acceptance_evidence` | string | no | 测试、手工验证或代码审查证据 |

**Validation Rules**:

- 对已启用能力，`compatibility_requirement` 必须说明 ID、schema 和错误反馈兼容性。
- `read` 在对象不存在时必须明确 `d.SetId("")` 或数据源空结果行为。
- `create/update/delete` 如涉及异步状态，必须关联 `wait` 或 `state_refresh` 服务能力。

## Entity: MigrationEvidence

**Purpose**: 记录改造证据，用于满足 FR-006、FR-007、SC-001、SC-003 和 SC-004。

| Field | Type | Required | Description |
|---|---|---|---|
| `object_group` | enum | yes | CMS 对象组 |
| `provider_files` | list(string) | yes | 相关 Provider 文件 |
| `decision` | enum | yes | `migrated`, `placeholder_no_user_visible_change`, `blocked`, `deferred` |
| `reason` | string | yes | 判定原因 |
| `service_methods` | list(string) | no | 已补齐服务方法 |
| `api_methods` | list(string) | no | 使用的 API wrapper 方法 |
| `tests` | list(string) | no | 验证项 |
| `review_notes` | string | no | 代码审查或兼容性说明 |

**Validation Rules**:

- 17 个对象组必须各有至少一条 `MigrationEvidence`。
- `decision=blocked` 必须附带缺口和后续任务。
- `decision=migrated` 必须包含服务方法、API 方法和测试证据。

## Relationships

- `CmsObjectGroup` 1..1 -> 3 `CmsProviderFile`
- `CmsObjectGroup` 1..N -> `CmsServiceCapability`
- `CmsServiceCapability` N..1 -> `CwsCmsApiWrapper`
- `TerraformOperationContract` N..1 -> `CmsServiceCapability`
- `MigrationEvidence` N..1 -> `CmsObjectGroup`

## State Transitions

### CmsProviderFile

```text
placeholder -> no_change_placeholder
placeholder -> enable_later -> implemented_service_layer -> registered
implemented_direct_call -> implemented_service_layer -> registered
implemented_service_layer -> registered
implemented_service_layer -> needs_api_gap_resolution
```

### MigrationEvidence

```text
pending_scope_decision -> placeholder_no_user_visible_change
pending_scope_decision -> migrated
pending_scope_decision -> blocked -> migrated
pending_scope_decision -> deferred
```
