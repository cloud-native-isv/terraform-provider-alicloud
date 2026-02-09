# Data Model: AliKafka 资源分层改造

## Entities

### AliKafka 实例
- **Fields**: 实例 Id、名称、规格、状态、地域、创建时间
- **Relationships**: 关联主题、SASL 用户、ACL、消费者组、部署、白名单绑定

### 主题
- **Fields**: 主题名称、分区数、副本数、状态、实例 Id
- **Relationships**: 归属 AliKafka 实例

### SASL 用户
- **Fields**: 用户名、状态、实例 Id
- **Relationships**: 归属 AliKafka 实例；与 SASL ACL 关联

### SASL ACL
- **Fields**: 规则 Id、资源类型、资源名称、权限、用户名、实例 Id
- **Relationships**: 归属 AliKafka 实例；关联 SASL 用户

### 消费者组
- **Fields**: 消费者组 Id、名称、状态、实例 Id
- **Relationships**: 归属 AliKafka 实例

### 部署
- **Fields**: 部署 Id、名称、状态、实例 Id
- **Relationships**: 归属 AliKafka 实例

### 实例白名单绑定
- **Fields**: 绑定 Id、实例 Id、IP 列表、状态
- **Relationships**: 归属 AliKafka 实例

## Validation Rules

- 资源标识字段必须存在且不可为空。
- 所有子资源必须关联到有效的实例 Id。
- 状态字段必须可映射为统一的可读状态字符串。

## State Transitions

- **实例**: Creating → Running → Updating → Deleting → Deleted
- **主题/ACL/用户/消费者组/部署/白名单绑定**: Creating → Active → Updating → Deleting → Deleted
