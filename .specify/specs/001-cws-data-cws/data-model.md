# Data Model (Phase 1)

本数据模型从功能需求与用户故事中提取实体、字段与约束，并标注关键状态与关系。

## Entities

### EipAddress
- Fields
  - allocationId (string, required) — 资源唯一标识
  - ipAddress (string, computed)
  - bandwidth (int, optional)
  - internetChargeType (string, optional)
  - name (string, optional)
  - description (string, optional)
  - status (string, computed) — e.g. Allocating → Available → (Associating/Unassociating) → Available
  - tags (map[string]string], optional)
- Validation
  - internetChargeType ∈ {PayByTraffic, PayByBandwidth}
  - bandwidth ≥ 0
- State Transitions
  - Allocate → Available
  - Associate/Unassociate 与 EipAssociation 协同改变（短暂中间态）

### NatGateway
- Fields
  - natGatewayId (string, required)
  - vpcId (string, required)
  - spec (string, optional)
  - natType (string, optional)
  - name (string, optional)
  - description (string, optional)
  - status (string, computed) — Creating → Available → Modifying/Deleting
  - deletionProtection (bool, optional, if supported later)
  - tags (map[string]string, optional)
- Validation
  - vpcId 必填
  - spec 与 natType 与地域能力匹配（由 Service 校验/透传错误）
- State Transitions
  - Create → Available；Modify → Available；Delete → NotFound

### SnatEntry
- Fields
  - snatEntryId (string, required)
  - snatTableId (string, required)
  - sourceCidr (string, required)
  - snatIp (string, required)
  - status (string, computed)
- Validation
  - sourceCidr 为合法 CIDR
  - snatIp 为合法 IP，且与 EIP/NAT 组合有效
- State Transitions
  - Create → Available；Modify → Available；Delete → NotFound

### EipAssociation
- Fields
  - allocationId (string, required)
  - instanceId (string, required)
  - instanceType (string, required) — {EcsInstance, NatGateway, …}
  - privateIpAddress (string, optional)
  - status (string, computed)
- Validation
  - 实例类型与目标资源匹配
  - EIP 与实例必须在可关联状态
- State Transitions
  - Associate → Associated；Unassociate → Disassociated

## Relationships
- EipAddress 1..1 ↔ 0..N EipAssociation（一个 EIP 可多次关联不同实例，但同一时刻通常 0..1 有效绑定，约束由云端与 Service 控制）
- NatGateway 1..N SnatEntry（同一 NAT 可有多个 SNAT 规则）

## Derived/Computed
- status 字段均在 Read 时由 Service Describe 结果设置；
- OpenAPI/GraphQL 合约仅作为文档映射，不直接驱动 Provider 运行时；
- 所有分页在 Service 内部实现，对外返回完整集合。
