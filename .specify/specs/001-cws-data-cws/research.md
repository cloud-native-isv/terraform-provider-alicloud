# Phase 0 Research

本研究文档用于消解 Technical Context 中的所有 NEEDS CLARIFICATION 并为设计产物提供依据。

## Unknowns → Decisions

1) CWS-Lib-Go 中的 VPC API 覆盖范围
- Decision: 本次改造以 EIP、NAT Gateway、SNAT Entry 与 EIP Association 的核心 API 为主（Allocate/Associate/Unassociate/Release/Describe；Create/Delete/Modify/Describe；Create/Delete/Modify/Describe）。增强类 API（例如高精度监控、删除保护、转发模式等）暂列入需求文档，后续在 CWS-Lib-Go 中补齐后再集成。
- Rationale: 优先保障核心资源管理能力与兼容性，最小化迁移风险。
- Alternatives considered: 直接调用官方 SDK 或 RPC；但与宪章分层冲突且缺乏统一重试/错误封装。

2) 资源层与 Service 层接口契约
- Decision: 资源层仅负责 Terraform schema 映射与 CRUD 生命周期，所有云端交互通过新增的 `service_alicloud_vpc_*.go` 实现（含分页、错误判定、状态刷新与等待函数）。
- Rationale: 符合治理的分层设计；提升可维护性与复用性。
- Alternatives considered: 在资源层直接拼装请求；不利于统一重试与状态处理。

3) 状态机与等待策略
- Decision: 为各资源在 Service 层实现 `*StateRefreshFunc` 与 `WaitFor*`，Pending/Target/Fail 状态遵循阿里云资源语义（如 EIP 分配后应达到 `Available`；NAT Gateway 创建后达到 `Available`）。
- Rationale: 避免在 Create 内直接 Read；在 Delete 使用 StateChangeConf 轮询直至删除成功或超时。
- Alternatives considered: 固定 sleep + Read；在大规模场景不稳定且不弹性。

4) 错误与重试规范
- Decision: 统一使用 `IsNotFoundError/IsAlreadyExistError/NeedRetry` 与 `WrapError/WrapErrorf`，结合常见可重试错误（ServiceUnavailable、Throttling 等）通过 `resource.Retry` 实现指数退避（或固定重试间隔）。
- Rationale: 与宪章一致并提升稳定性。
- Alternatives considered: 以 `IsExpectedErrors` 列表硬编码；可读性与可维护性较差。

5) 分页
- Decision: 所有 List/Describe 的分页逻辑仅存在于 Service 层 `*_api.go`（或 service 文件内专门函数），对外返回完整聚合结果。
- Rationale: 规避调用侧重复分页实现，保持一致性。
- Alternatives considered: 由资源层处理分页；增加重复与耦合。

6) ID 编码/解码
- Decision: Service 层为涉及复合标识的对象提供 `Encode*Id/Decode*Id`（如 `workspaceId:namespace:jobId` 模式），VPC 资源常用阿里云原生 Id（AllocationId、NatGatewayId、SnatEntryId 等）则直接使用并保持 Provider 一致风格（字段名使用 `Id`）。
- Rationale: 与宪章约定一致；对复合 Id 保持统一规范。
- Alternatives considered: 在资源层散落处理；不利于一致性与调试。

7) 测试策略
- Decision: 单元测试覆盖 Service 层方法（分页、错误、状态机）；验收测试覆盖四类资源的 Happy Path 与 1-2 个边界（如解绑不存在的关联、重复释放等）。
- Rationale: 保障核心路径正确性与回归稳定性。
- Alternatives considered: 仅依赖手工验证；风险较高。

## Best Practices
- Terraform Provider：
  - Create 后调用 Service 层 `WaitFor*`；Delete 使用 `StateChangeConf`；Read 补齐 computed 字段。
  - 不在 Create 中直接调用 Read。
  - 尽可能实现幂等（重复调用 Create 在 AlreadyExist 时走 Read）。
- CWS-Lib-Go：
  - 统一封装分页与重试；类型安全与错误包装；提供清晰的方法命名与注释。

## Integrations/Patterns
- 与现有 Provider 错误与日志框架保持一致（`WrapErrorf(err, DefaultErrorMsg, ...)`）。
- 继续使用 `timeouts` 与 Terraform `resource.Retry` 配置，遵循既有超时约定。

## Decisions Summary
- 使用 CWS-Lib-Go VPC API，资源层 → Service 层 → API 层，杜绝直连 SDK/RPC。
- 增强 API 先文档化，等 CWS-Lib-Go 实现后再集成。
- Service 层承担分页、状态机、错误与重试，资源层轻薄化。
