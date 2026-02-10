# Requirements Specification: AliKafka 实例计费组合一致性

**Requirement Branch**: `003-alikafka-instance-billing`  
**Created**: 2026-02-10  
**Status**: Draft  
**Input**: User description: "参考pkg/cws-lib-go/.specify/specs/007-kafka-instance-billing/目录中的实现文档，根据最新的cws-lib-go层的kafka instance api实现更新resourceAliCloudAlikafkaInstance和KafkaService中方法的实现，重点在于使用同一个CreateInstance API创建Prepaid和Postpaid、Serverless和reserved的四中Instance组合创建。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 创建预留型实例 (Priority: P1)

作为 Terraform 使用者，我需要创建预留型 Kafka 实例并选择预付费或后付费，以便满足不同的采购与成本管理需求。

**Why this priority**: 预留型实例是最常见的生产形态，创建路径必须稳定可靠。

**Independent Test**: 仅通过创建预留型实例（分别选预付费/后付费）即可验证资源创建与结果回显。

**Acceptance Scenarios**:

1. **Given** 选择了预留型与后付费，**When** 发起实例创建，**Then** 创建成功并返回实例标识用于后续管理。
2. **Given** 选择了预留型与预付费，**When** 发起实例创建，**Then** 创建成功并返回订单标识或实例标识以便追踪。

---

### User Story 2 - 创建 Serverless 实例 (Priority: P2)

作为 Terraform 使用者，我需要创建 Serverless Kafka 实例并选择可用的计费方式，以便在弹性场景下快速启用实例。

**Why this priority**: Serverless 是独立产品形态，确保其创建流程可用对用户价值明显。

**Independent Test**: 仅通过创建 Serverless 实例即可验证该形态的创建与结果回显。

**Acceptance Scenarios**:

1. **Given** 选择了 Serverless 与后付费，**When** 发起实例创建，**Then** 创建成功并返回实例标识用于后续管理。

---

### User Story 3 - 不支持组合的明确提示 (Priority: P3)

作为平台运维人员，我需要在用户选择不支持的计费组合时收到清晰错误提示，以便快速定位并纠正配置。

**Why this priority**: 避免误配置导致的失败重试和支持成本上升。

**Independent Test**: 使用不支持的组合创建实例即可验证错误提示清晰可理解。

**Acceptance Scenarios**:

1. **Given** 选择了不支持的实例类型与计费组合，**When** 发起实例创建，**Then** 返回明确错误且不产生实例。

### Edge Cases

- 未提供计费方式时，应保持既有默认行为且结果可预测。
- 仅提供实例类型但缺少计费参数时，系统应使用既有默认值并可回显最终选择。
- 选择不支持的组合时，必须在创建前给出明确错误且不产生实例。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 允许用户在实例创建时选择实例类型（预留型、Serverless）。
- **FR-002**: 系统 MUST 允许用户在实例创建时选择计费方式（预付费、后付费）。
- **FR-003**: 系统 MUST 支持预留型 + 后付费与预留型 + 预付费两种组合的实例创建。
- **FR-004**: 系统 MUST 支持 Serverless + 后付费组合的实例创建。
- **FR-005**: 系统 MUST 拒绝 Serverless + 预付费组合的创建请求，并返回清晰错误信息。
- **FR-006**: 对于成功的创建请求，系统 MUST 返回可用于后续管理的实例标识；若产生订单，必须返回订单标识或等价追踪信息。
- **FR-007**: 系统 MUST 在资源状态中回显最终选择的实例类型与计费方式，便于用户确认。
- **FR-008**: 当计费方式或实例类型未提供时，系统 MUST 沿用既有默认行为并确保结果可预测。

### Key Entities *(include if requirement involves data)*

- **Instance Type**: 实例类型（预留型、Serverless）。
- **Billing Type**: 计费方式（预付费、后付费）。
- **Instance Creation Request**: 包含实例类型、计费方式与必要规格的创建请求。
- **Creation Result**: 创建结果，包含实例标识与可选的订单标识。
- **Validation Error**: 表达不支持组合或参数缺失的错误信息。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 预留型 + 预付费、预留型 + 后付费、Serverless + 后付费三种合法组合的验收创建成功率达到 100%。
- **SC-002**: 不支持的计费组合在 2 秒内返回清晰错误提示且不产生实例。
- **SC-003**: 用户在一次 Terraform 申请中即可完成目标组合创建的比例达到 95%。
- **SC-004**: 因计费组合错误导致的支持工单数量在发布后 30 天内下降 30%。

## Assumptions

- 未提供计费方式或实例类型时，沿用当前资源的默认行为，不改变既有用户体验。
- 合法计费组合与限制以云端产品规则为准。

## Dependencies

- 依赖云端 Kafka 产品对实例类型与计费组合的规则与可用性。

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
