# VPC EIP CWS-Lib-Go 迁移说明

本文档描述 Terraform Provider Alicloud 在 EIP 资源上的迁移要求与 API 映射，帮助在 CWS-Lib-Go 中对齐/补齐必要能力。

更新时间: 2025-11-04
适用范围: `alicloud/service_alicloud_vpc_eip.go`, `alicloud/resource_alicloud_eip_address.go`

## 背景

- Provider 侧已引入 `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc` 并在 Service 层完成初始化。
- 现阶段 EIP Service 仍通过旧版 RPC 进行实际调用，仅在 Service 内部封装，Resource 层已完成解耦。
- 目标是将 Service 内部全部切换为 CWS-Lib-Go VPC API 强类型调用，统一错误处理与重试策略。

## 目标 API 与契约

以下 API 名称参考 `/.specify/specs/001-cws-data-cws/contracts/openapi.yaml` 与 `spec.md`，建议在 CWS-Lib-Go 中提供同名/等价方法：

1) AllocateEipAddress
- 入参: 
  - bandwidth (int)
  - internetChargeType (string: PayByTraffic | PayByBandwidth)
  - description (string, optional)
  - name (string, optional)
  - tags ([]Tag, optional)
- 返回: 
  - allocationId (string)
  - ipAddress (string, optional)

2) DescribeEipAddress
- 入参: allocationId (string)
- 返回: 结构体 (建议字段)
  - AllocationId (string)
  - IpAddress (string)
  - Status (string) // Available | InUse | Allocating | Associating | Unassociating | Released | Failed
  - Bandwidth (int)
  - ChargeType (string)
  - InternetChargeType (string)
  - ResourceGroupId (string, optional)
  - Tags ([]Tag)

3) ModifyEipAddressAttribute
- 入参: allocationId (string), attrs(结构体或参数):
  - Bandwidth (int, optional)
  - Description (string, optional)
  - Name (string, optional)
  - InternetChargeType (string, optional)
- 返回: error

4) ReleaseEipAddress
- 入参: allocationId (string)
- 返回: error

5) AssociateEipAddress
- 入参: 
  - allocationId (string)
  - instanceId (string)
  - instanceType (string) // ECS, SLB, NatGateway, NetworkInterface 等
  - privateIpAddress (string, optional) // 绑定 ENI 的场景
- 返回: error

6) UnassociateEipAddress
- 入参: allocationId (string)
- 返回: error

备注: 若已有等价方法但命名不同，请在 provider 侧对接时适配；若缺失请补齐。

## 错误语义与重试建议

- 需要支持识别 NotFound、AlreadyExist、Retryable 错误：
  - NotFound: 当 Describe 不存在或 Release/Unassociate 已经生效
  - Retryable: ServiceUnavailable, Throttling, SystemBusy, OperationConflict, LastTokenProcessing, IncorrectStatus.* 等
- Provider 侧调用通过 `resource.Retry` + `NeedRetry(err)` 实现指数回退，建议 cws-lib-go 内部也做轻量重试（可选），但需保证幂等。

## 状态机语义

- Pending: [Allocating, Associating, Unassociating]
- Target: [Available, InUse]
- Fail: [Released, Failed]
- Describe 返回对象中的 `Status` 字段需符合上述值域，以便 Provider 的 StateRefreshFunc 正确轮询。

## Provider 侧映射（当前 → 目标）

- AllocateEipAddress: 由 `VpcEipService.AllocateEipAddress(map[string]interface{})` 调用 vpcAPI.AllocateEipAddress(...)，返回 AllocationId
- DescribeEipAddress: 由 `VpcEipService.DescribeEipAddress(string)` 调用 vpcAPI.Get/DescribeEipAddress(...)，返回强类型对象；Provider 将其转为 map[string]interface{} 以保持兼容
- ModifyEipAddressAttribute: 直接调用 vpcAPI.ModifyEipAddressAttribute(...)
- ReleaseEipAddress: 直接调用 vpcAPI.ReleaseEipAddress(...)
- AssociateEipAddress: 直接调用 vpcAPI.AssociateEipAddress(...)
- UnassociateEipAddress: 直接调用 vpcAPI.UnassociateEipAddress(...)

## 迁移步骤与验证

1) 在 `VpcEipService` 中增加 `GetAPI()` 以便资源/服务内部可直接使用强类型 API（已完成）。
2) 将 `DescribeEipAddress` 首先切换至 vpcAPI，返回值转为 map，以降低变更面。
3) 依次替换 Allocate/Modify/Release/Associate/Unassociate，保留 Provider 侧重试与错误包装逻辑。
4) 运行 `make`，并对 EIP 资源执行 Create/Read/Update/Delete 的本地烟测。

## 兼容性要求

- 不改变 Terraform schema 字段名与含义；ChargeType/InternetChargeType 的值域需与历史对齐（必要时做转换）。
- ID 字段统一使用 `Id`（非 `ID`）。

## 后续工作

- 若 cws-lib-go VPC API 缺失上述方法或返回字段，请据此文档完善；Provider 侧将在方法可用后立即切换实现。
- 完成 EIP 后，复用同一 API 客户端按相同模式迁移 NAT 网关与 SNAT 条目。
