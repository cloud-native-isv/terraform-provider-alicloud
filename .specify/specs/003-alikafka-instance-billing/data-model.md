# Data Model: AliKafka 实例计费组合一致性

## Entities

### InstanceType
- **Description**: 实例类型枚举。
- **Allowed Values**: Reserved（预留型）、Serverless（Serverless）。

### BillingType
- **Description**: 计费方式枚举。
- **Allowed Values**: Prepaid（预付费）、Postpaid（后付费）。

### InstanceCreationRequest
- **Fields**:
  - instanceType: InstanceType（可选，缺省沿用既有默认行为）
  - billingType: BillingType（可选，缺省沿用既有默认行为）
  - regionId: string
  - name: string
  - spec: string（实例规格，沿用既有字段定义）
- **Relationships**: 生成 InstanceCreationResult。

### InstanceCreationResult
- **Fields**:
  - instanceId: string
  - orderId: string（可选）
  - instanceType: InstanceType
  - billingType: BillingType
- **Relationships**: 与 InstanceCreationRequest 一一对应。

### ValidationError
- **Fields**:
  - code: string
  - message: string
  - invalidCombination: string
  - allowedCombinations: []string

## Validation Rules

- 仅支持 Reserved + Prepaid、Reserved + Postpaid、Serverless + Postpaid。
- Serverless + Prepaid 必须在创建前返回明确错误。
- 当 instanceType 或 billingType 未提供时，沿用既有默认行为并可回显最终选择。
- 成功创建必须返回 instanceId；若产生订单需返回 orderId 或等价标识。

## State Transitions

- Instance: Creating → Running → Updating → Deleting → Deleted
