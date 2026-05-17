# Research: CMS Provider 封装能力改造

**Date**: 2026-05-15  
**Branch**: `007-refactor-cms-api`  
**Spec**: `requirements.md`

## Research Sources

- `requirements.md`: 用户脚本枚举范围、用户故事、FR/SC。
- `.specify/memory/constitution.md`: 分层、强类型、测试优先、Feature-Centric 约束。
- `.specify/memory/features.md` 与 `.specify/memory/features/005.md`: 本规格绑定 Feature 005 / CWS-Lib-Go Integration。
- `README.md`: Go/Provider 构建与测试说明、项目结构、Feature 概览。
- `docs/development_guide.md`: Resource/DataSource -> Service -> API -> SDK 分层、分页、强类型、错误处理与 WaitFor 规范。
- `docs/wait_for_state.md`: `StateRefreshFunc` 与 `WaitFor*` 编码规范。
- `alicloud/{service,resource,data_source}_alicloud_cms_*.go`: 用户脚本范围内 Provider 文件。
- `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/*`: 本地 CMS API wrapper、类型、转换函数和测试。

## Scope Baseline

用户脚本展开为 51 个 Provider 文件：

- 17 个服务层文件：`service_alicloud_cms_{addon,agg_task,alert,cloud_resource,context,context_store,dataset,delivery_task,entity_store,integration_policy,memory,memory_store,pipeline,prometheus,service,umodel,workspace}.go`
- 17 个资源层文件：`resource_alicloud_cms_{...}.go`
- 17 个数据源层文件：`data_source_alicloud_cms_{...}.go`

源代码检查结果：

- 50 个文件当前仅包含 `package alicloud`，属于占位/待启用状态。
- `alicloud/data_source_alicloud_cms_service.go` 当前已有 `alicloud_cms_service` 数据源实现，但直接调用 `client.RpcPost("Cms", "2019-01-01", "OpenCmsService", ...)`，不符合分层原则；文件末尾还存在重复 `package alicloud`，后续实现前必须修复。
- `provider.go` 当前注册的是既有 CMS 资源/数据源集合；脚本范围内明确注册的用户可见能力为 `alicloud_cms_service` 数据源。

## Existing CMS API Wrapper Coverage

本地目录 `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/` 已包含与脚本范围对应的 CMS API wrapper：

| Object Group | Local API Coverage | Notes |
|---|---|---|
| addon | `GetAddon`, `GetAddonSchema`, `GetAddonCodeTemplate`, `GetAddonRelease`, `CreateAddonRelease`, `UpdateAddonRelease`, `DeleteAddonRelease`, `ListAddonReleases`, `ListAddons` | 发布/查询双路径 |
| agg_task | `GetAggTaskGroup`, `CreateAggTaskGroup`, `UpdateAggTaskGroup`, `UpdateAggTaskGroupStatus`, `DeleteAggTaskGroup`, `ListAggTaskGroups` | API 文件命名为 `agg_task_group` |
| alert | `ListAlertActions`, webhook CRUD/list, robots/contact list, rule manage/query, notify strategy update | 部分策略接口为动态结构 |
| cloud_resource | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| context | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| context_store | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| dataset | `CreateDataset`, `UpdateDataset`, `DeleteDataset`, `GetDataset`, `ListDatasets`, `ExecuteQuery` | `ExecuteQuery` 语义动态 |
| delivery_task | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| entity_store | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| integration_policy | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| memory | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| memory_store | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| pipeline | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| prometheus | Prometheus view/instance/virtual instance/user setting methods and list methods | 包含多个子对象 |
| service | API/types/tests present | 用于替换直接 `OpenCmsService` RPC |
| umodel | API/types/utils/tests present | 需在任务阶段核对方法粒度 |
| workspace | `PutWorkspace`, `GetWorkspace`, `ListWorkspaces`, `DeleteWorkspace` | 支持工作空间生命周期 |

## Decisions

### Decision 1: 以“范围矩阵”驱动迁移顺序

**Decision**: 在任务阶段先建立 51 个文件的范围矩阵，逐项标记为“已启用改造、无需改造占位、待启用补齐”。

**Rationale**: 需求 SC-001 和 FR-009 要求 100% 范围判定与可追踪映射；当前大多数文件是占位，若直接按资源实现会误导交付状态。

**Alternatives considered**:

- 直接实现所有 17 个 CMS 对象的 Terraform 资源/数据源：被拒绝，因为需求假设“不新增用户可见能力，除非现有范围明确需要启用”。
- 仅修改 `data_source_alicloud_cms_service.go`：被拒绝，因为无法满足对全部脚本枚举文件的范围判定与证据要求。

### Decision 2: 先迁移当前已启用数据源 `alicloud_cms_service`

**Decision**: 以 `data_source_alicloud_cms_service.go` 为首个实际代码迁移对象，新增/补齐 `service_alicloud_cms_service.go` 中的服务方法，由服务方法调用 CWS-Lib-Go CMS service API。

**Rationale**: 该数据源当前直接调用 `RpcPost`，是明确的宪法 I/开发指南 1.1 违例；也是脚本范围内唯一已注册的用户可见能力。

**Alternatives considered**:

- 保留直接 RPC，仅记录例外：被拒绝，因为本规格目标就是迁移到封装能力，且已有本地 API wrapper。

### Decision 3: Provider 层保持兼容，不主动批量注册占位资源/数据源

**Decision**: 除非任务阶段证明某对象已具备完整 schema、服务方法、测试和兼容验收，否则不在 `provider.go` 中新增脚本范围内占位资源/数据源注册。

**Rationale**: 避免将内部改造误变为新用户可见功能；满足 FR-005 和边界条件中“占位内容或暂未注册能力”的区分要求。

### Decision 4: Service 方法提供强类型边界，动态结构仅限 API 语义必要处

**Decision**: Provider 服务层优先使用 CWS-Lib-Go CMS typed structs；若 CMS API wrapper 已因远端 API 动态查询语义使用 `map[string]any`，Provider 服务层不得扩大动态结构传播范围，并需在任务/代码审查记录中标记兼容例外。

**Rationale**: 宪法 II 和 Feature 008 禁止新增弱类型主要载体；但 CMS `ExecuteQuery`、通知策略等远端接口存在动态语义，需要隔离而非伪造类型。

### Decision 5: 等待、分页和错误归一放在 Service/API 层

**Decision**: List/query 完整结果、not-found 判断、状态等待和删除消失语义由 API wrapper 或 Provider service 封装；Resource/DataSource 只做 Terraform schema、diff 和状态写入。

**Rationale**: 与 `docs/development_guide.md`、`docs/wait_for_state.md` 保持一致，降低重复逻辑和回归风险。

## Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| 占位文件被误判为已交付资源/数据源 | 用户可见能力不完整 | 通过范围矩阵和 provider registration 审计标记待启用 |
| CWS-Lib-Go wrapper 方法与现有 Terraform 字段语义不完全一致 | 兼容性回退 | 服务层适配字段名、默认值、ID 编解码和错误语义 |
| CMS 部分接口异步或无更新能力 | CRUD 设计不匹配 | 对每个对象记录生命周期能力；仅实现实际支持操作 |
| 动态查询/策略接口使用弱类型 | 强类型原则风险 | 将弱类型限制在 API wrapper 语义边界，并记录例外 |
| 接受测试依赖真实云资源 | 本地验证不稳定 | 优先补齐单元/服务转换测试；AccTest 作为凭据可用时的发布门禁 |

## Open Items

无阻塞性 `NEEDS CLARIFICATION`。任务阶段仍需逐对象核对 CWS-Lib-Go CMS wrapper 的具体方法签名、SDK 错误码映射和 provider schema 是否应启用。
