# Feature Specification: 更新VPC相关资源以使用新的CWS-Lib-Go API

**Feature Branch**: `001-cws-data-cws`  
**Created**: November 4, 2025  
**Status**: Draft  
**Input**: User description: "根据/cws_data/cws-lib-go/lib/cloud/aliyun/api/vpc中新添加的vpc相关api，更新resourceAliCloudNatGateway，resourceAliCloudSagSnatEntry，resourceAliCloudEipAddress、resourceAliCloudEipAssociation等实现，对于需要实现但是当前/cws_data/cws-lib-go/lib/cloud/aliyun/api/vpc中还没有实现的API可以写成markdown格式的需求文档。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 更新EIP地址资源 (Priority: P1)

作为Terraform用户，我希望能够使用最新的阿里云VPC API来管理EIP地址资源，以便获得更好的性能和新功能支持。

**Why this priority**: EIP地址是VPC中最基础的资源之一，许多其他资源都依赖于它，更新它将为其他资源的更新奠定基础。

**Independent Test**: 可以通过创建、读取、更新和删除EIP地址来完全测试此功能，并验证其与阿里云API的兼容性。

**Acceptance Scenarios**:

1. **Given** 用户配置了EIP地址资源，**When** 执行terraform apply，**Then** 应该成功创建EIP地址并使用新的CWS-Lib-Go API
2. **Given** 已存在的EIP地址，**When** 执行terraform plan，**Then** 应该正确识别资源状态而无需不必要的更改

---

### User Story 2 - 更新NAT网关资源 (Priority: P1)

作为Terraform用户，我希望能够使用最新的阿里云VPC API来管理NAT网关资源，以便利用新功能和改进的性能。

**Why this priority**: NAT网关是VPC中的核心网络组件，更新它将提升网络连接的可靠性和性能。

**Independent Test**: 可以通过创建、配置和删除NAT网关来完全测试此功能。

**Acceptance Scenarios**:

1. **Given** 用户配置了NAT网关资源，**When** 执行terraform apply，**Then** 应该成功创建NAT网关并使用新的CWS-Lib-Go API
2. **Given** 已存在的NAT网关，**When** 修改其规格或属性，**Then** 应该成功更新资源而不会中断现有连接

---

### User Story 3 - 更新SNAT条目资源 (Priority: P2)

作为Terraform用户，我希望能够使用最新的阿里云VPC API来管理SNAT条目，以便更好地控制网络流量。

**Why this priority**: SNAT条目是NAT网关的重要组成部分，更新它可以提供更好的网络流量管理能力。

**Independent Test**: 可以通过创建、修改和删除SNAT条目来测试此功能。

**Acceptance Scenarios**:

1. **Given** 用户配置了SNAT条目资源，**When** 执行terraform apply，**Then** 应该成功创建SNAT条目并使用新的CWS-Lib-Go API
2. **Given** 已存在的SNAT条目，**When** 修改源CIDR或SNAT IP，**Then** 应该成功更新规则

---

### User Story 4 - 更新EIP关联资源 (Priority: P2)

作为Terraform用户，我希望能够使用最新的阿里云VPC API来管理EIP与实例的关联，以确保网络连接的稳定性。

**Why this priority**: EIP关联是连接云资源与公网的重要操作，更新它可以提升关联操作的可靠性。

**Independent Test**: 可以通过将EIP关联到不同类型的实例（如ECS、NAT网关等）来测试此功能。

**Acceptance Scenarios**:

1. **Given** 用户配置了EIP关联资源，**When** 执行terraform apply，**Then** 应该成功将EIP与实例关联并使用新的CWS-Lib-Go API
2. **Given** 已存在的EIP关联，**When** 解除关联，**Then** 应该成功分离EIP与实例

### Edge Cases

- 当CWS-Lib-Go中的API与当前直接调用的API行为不一致时如何处理？
- 如何处理API版本兼容性问题？
- 当新的API缺少某些当前实现的功能时怎么办？

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须使用CWS-Lib-Go封装的VPC API替代直接的RPC调用
- **FR-002**: 系统必须保持与现有Terraform配置的向后兼容性
- **FR-003**: 用户必须能够通过terraform plan/apply管理EIP地址资源
- **FR-004**: 系统必须支持NAT网关的所有现有功能和属性
- **FR-005**: 系统必须支持SNAT条目的创建、修改和删除
- **FR-006**: 系统必须支持EIP与各种实例类型的关联操作
- **FR-007**: 系统必须正确处理资源状态刷新和等待逻辑
- **FR-008**: 系统必须保持现有的错误处理和重试机制

### Key Entities

- **EIP地址**: 弹性公网IP地址资源，包含IP地址、带宽、付费类型等属性
- **NAT网关**: 网络地址转换网关，包含规格、类型、VPC关联等属性
- **SNAT条目**: 源网络地址转换规则，定义内部网络到外部网络的地址转换
- **EIP关联**: EIP地址与云资源实例的绑定关系

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 使用新API的资源操作响应时间提升20%以上
- **SC-002**: 资源创建成功率保持在99%以上
- **SC-003**: 95%的用户能够成功迁移至使用新API的资源版本
- **SC-004**: 减少与API调用相关的错误率50%以上

# 缺失API需求文档

## 需要实现的API列表

### 1. EIP地址相关API
- `AllocateEipAddress`: 分配EIP地址
- `AssociateEipAddress`: 绑定EIP到实例
- `UnassociateEipAddress`: 解绑EIP
- `ReleaseEipAddress`: 释放EIP
- `ModifyEipAddressAttribute`: 修改EIP属性
- `DescribeEipAddresses`: 查询EIP信息

### 2. NAT网关相关API
- `CreateNatGateway`: 创建NAT网关
- `DeleteNatGateway`: 删除NAT网关
- `ModifyNatGatewayAttribute`: 修改NAT网关属性
- `ModifyNatGatewaySpec`: 修改NAT网关规格
- `DescribeNatGateways`: 查询NAT网关信息

### 3. SNAT条目相关API
- `CreateSnatEntry`: 创建SNAT条目
- `DeleteSnatEntry`: 删除SNAT条目
- `ModifySnatEntry`: 修改SNAT条目
- `DescribeSnatTableEntries`: 查询SNAT条目信息

## 当前可能缺失的API

### 1. 增强的EIP管理API
- `EnableHighDefinitionMonitorLog`: 启用高精度监控日志
- `DisableHighDefinitionMonitorLog`: 禁用高精度监控日志
- `DescribeHighDefinitionMonitorLogAttribute`: 查询高精度监控日志属性

### 2. 增强的NAT网关API
- `UpdateNatGatewayNatType`: 更新NAT网关类型
- `DeletionProtection`: 启用/禁用删除保护
- `ModifyEipForwardMode`: 修改EIP转发模式

### 3. 增强的标签管理API
- `TagResources`: 为资源添加标签的增强版本
- `UnTagResources`: 为资源移除标签的增强版本
- `ListTagResources`: 列出资源标签的增强版本

这些API需要在CWS-Lib-Go中实现，以便Terraform Provider能够使用它们来提供更丰富的功能。
