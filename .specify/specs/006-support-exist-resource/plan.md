# Implementation Plan: 自动接管已存在数据库（alicloud_db_database）

**Branch**: `006-support-exist-resource` | **Date**: 2025-10-31 | **Spec**: /.specify/specs/006-support-exist-resource/spec.md
**Input**: Feature specification from `/.specify/specs/006-support-exist-resource/spec.md`

## Summary

目标：当用户声明创建已存在的数据库（同一实例内同名）时，Provider 自动接管现有数据库为受管对象，而非重复创建或“静默保存”。

技术路径（高层）：
- 在 Create 路径添加“存在性检测 + 接管”分支：通过 Service 层调用 cws-lib-go RDS API 查询实例内是否已存在该 DB；若存在，直接设置资源 Id 并进入状态对齐；若不存在，执行创建 + 等待就绪。
- 在 Plan 透明性方面：通过 CustomizeDiff 执行只读存在性探测，并在计划输出中呈现“将接管已存在数据库”的提示字段（只读、不会导致实际变更）。
- 全面对齐架构与治理规范：移除直接 RpcPost，统一走 Service 层；规范错误处理、状态等待与超时设置；遵循 Id 命名约定。

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: Terraform Plugin SDK v1.17.2；cws-lib-go `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api`（RDS）  
**Storage**: N/A（Provider 内部状态 + Terraform state）  
**Testing**: go test；terraform-plugin-test（集成测试）；手工验收用例（spec 中验收场景）  
**Target Platform**: Linux（Terraform Provider 运行环境）  
**Project Type**: Terraform Provider（单仓多资源）  
**Performance Goals**: 接管路径 90% ≤ 2min、99% ≤ 5min（含限流与状态轮询）；Plan 阶段探测应轻量且受速率限制保护  
**Constraints**: 遵循阿里云 RDS API 速率限制；不可变字段冲突采取“失败 + 指引”保守策略；严禁破坏性变更  
**Scale/Scope**: 适配常见引擎（MySQL/PostgreSQL/SQLServer），判断范围限定在“目标实例内”

未知项（初始）：
- 各引擎不可变字段清单与规范化提示文案（NEEDS CLARIFICATION → 在 research.md 解决）
- Plan 阶段“提示字段”的最小侵入设计（字段命名、是否显示、对 diff 的影响）（NEEDS CLARIFICATION → 在 research.md 选择方案）

## Constitution Check

必须符合的门禁（来自宪法）：
1) 架构分层：Resource/DataSource 层仅调用 Service 层；Service 层使用 cws-lib-go 封装，避免直接 HTTP/第三方 SDK。
2) 状态管理：Create 不直接调用 Read；使用 StateRefreshFunc/WaitFor 等待目标状态；资源不存在时 d.SetId("")；Read 设置所有 computed 字段。
3) 错误处理：优先使用封装的错误判断（IsNotFoundError/IsAlreadyExistError/NeedRetry）；使用 WrapError/WrapErrorf；合理重试常见错误（Throttling、ServiceUnavailable、SystemBusy...）。
4) 代码质量与命名：ID 字段使用 Id；Schema 含 Description；超时设置齐备；文件>1000 行需拆分。
5) 验证：变更后执行 `make` 编译校验。

门禁评估（设计前）：
- 当前实现存在违例：Create 使用 `client.RpcPost`（未走 Service 层；与宪法不符）。
- 计划中的整改：重构 Create 至 Service 层 + cws-lib-go；增加 WaitFor 创建等待；补充错误与重试规范；通过 CustomizeDiff 实现 Plan 透明性但仅做读取。

结论（设计后复核）：PASS

复核要点：
- 已制定 Service 层改造与错误/重试/等待规范，满足架构分层与状态管理门禁。
- Plan 透明性通过 CustomizeDiff 的只读探测与 Optional+Computed 提示字段实现，不引入副作用，满足 FR-009。
- 命名、描述与超时等质量约束在设计中明确约束；>1000 行拆分暂不涉及。

## Project Structure

### Documentation (this feature)

```
/.specify/specs/006-support-exist-resource/
├── plan.md              # 本文件（/speckit.plan 输出）
├── research.md          # Phase 0 输出（研究与抉择）
├── data-model.md        # Phase 1 输出（实体/状态/校验）
├── quickstart.md        # Phase 1 输出（使用与验证）
└── contracts/           # Phase 1 输出（OpenAPI/GraphQL 合约）
```

### Source Code (repository root)

```
alicloud/
├── resource_alicloud_db_database.go        # Resource 层（目标重构：Create 接管/创建、CustomizeDiff 只读探测）
├── service_alicloud_rds_base.go            # Service 层（已接入 cws-lib-go）
├── resource_alicloud_rds_db_proxy*.go      # 参考 RDS Service 使用与等待模式
├── common.go                               # 错误包装/工具方法
└── provider.go                             # 资源注册
```

**Structure Decision**: 继续沿用当前 Provider 单仓结构；对目标资源文件进行内聚改造并补充必要的 Service 层方法与等待函数，遵循“Resource 调 Service、Service 调 API”的分层。

## Complexity Tracking

当前无需要豁免的复杂度设计；CustomizeDiff 仅用于“只读存在性提示”，不会引入副作用或破坏幂等。
