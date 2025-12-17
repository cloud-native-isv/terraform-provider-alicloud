# Feature Specification: Implement KafkaService methods using cws-lib-go

**Feature Branch**: `005-implement-kafka-service`  
**Created**: 2025-12-16  
**Status**: Draft  
**Input**: User description: "将KafkaService中的函数实现补全，当前很多函数都存在如下注释 // CWS-Lib-Go KafkaAPI does not yet expose instance order creation. // Keep signature for forward compatibility but fail fast at runtime. 这些注释是在cws-lib-go代码中还不支持对应的函数时留下的，现在aliyun/api/kafka中的cws-lib-go代码库已经完善了kafka相关的API，需要更新KafkaService中的实现。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Kafka Instance (Priority: P1)

As a Terraform user, I want to create Kafka instances (both PrePaid and PostPaid) so that I can use message queuing services.

**Why this priority**: Core functionality for provisioning resources.

**Independent Test**: Verify that `CreatePostPayOrder` and `CreatePrePayOrder` methods in `KafkaService` correctly invoke the `cws-lib-go` API.

**Acceptance Scenarios**:

1. **Given** a valid PostPaid instance configuration, **When** `CreatePostPayOrder` is called, **Then** it should invoke `KafkaAPI.CreateOrder` with `PaidType=PostPaid` and return the Order ID.
2. **Given** a valid PrePaid instance configuration, **When** `CreatePrePayOrder` is called, **Then** it should invoke `KafkaAPI.CreateOrder` with `PaidType=PrePaid` and return the Order ID.

---

### User Story 2 - Manage Kafka Instance Lifecycle (Priority: P2)

As a Terraform user, I want to start, modify, and upgrade Kafka instances so that I can manage my infrastructure.

**Why this priority**: Essential for day-to-day operations and maintenance.

**Independent Test**: Verify `StartInstance`, `ModifyInstanceName`, and `UpgradeInstanceVersion` methods.

**Acceptance Scenarios**:

1. **Given** a created instance, **When** `StartInstance` is called, **Then** it should invoke `KafkaAPI.StartInstance`.
2. **Given** an existing instance, **When** `ModifyInstanceName` is called, **Then** it should invoke `KafkaAPI.ModifyInstanceName`.
3. **Given** an existing instance, **When** `UpgradeInstanceVersion` is called, **Then** it should invoke `KafkaAPI.UpgradeInstanceVersion`.

### Edge Cases

- What happens when the API returns an error? The service methods must wrap and return the error.
- What happens when optional parameters are missing? The service methods should handle nil or empty values gracefully when constructing the API request.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `KafkaService` MUST be updated to initialize and hold a reference to `cws-lib-go`'s `KafkaAPI`.
- **FR-002**: `CreatePostPayOrder` MUST accept a strong type (e.g., `*KafkaOrder`) and validate it before calling `KafkaAPI.CreateOrder` with `PaidType=PostPaid`.
- **FR-003**: `CreatePrePayOrder` MUST accept a strong type (e.g., `*KafkaOrder`) and validate it before calling `KafkaAPI.CreateOrder` with `PaidType=PrePaid`.
- **FR-004**: `StartInstance` MUST accept a strong type (e.g., `*StartInstanceRequest`) and validate it before calling `KafkaAPI.StartInstance`.
- **FR-005**: `ModifyInstanceName` MUST accept a strong type (e.g., `*ModifyInstanceNameRequest`) and validate it before calling `KafkaAPI.ModifyInstanceName`.
- **FR-006**: `UpgradeInstanceVersion` MUST accept a strong type (e.g., `*UpgradeInstanceVersionRequest`) and validate it before calling `KafkaAPI.UpgradeInstanceVersion`.
- **FR-007**: `UpgradePostPayOrder` and `UpgradePrePayOrder` should remain unimplemented if the underlying API does not support them, returning a clear error message.

### Key Entities *(include if feature involves data)*

- **KafkaOrder**: Struct used to pass order details to `CreateOrder`.
- **KafkaInstance**: Entity representing the Kafka instance.
- **StartInstanceRequest**: Struct for starting an instance.
- **ModifyInstanceNameRequest**: Struct for modifying instance name.
- **UpgradeInstanceVersionRequest**: Struct for upgrading instance version.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `KafkaService` compiles successfully with the new implementation.
- **SC-002**: All previously stubbed methods (except those truly missing in API) are implemented and call the `cws-lib-go` API.
- **SC-003**: Error handling is consistent with the provider's patterns (wrapping errors).

## Clarifications

### Session 2025-12-16
- Q: How should `KafkaService` methods handle input parameters? → A: Refactor to use strong types (e.g., `KafkaInstance`, `KafkaOrder`, `StartInstanceRequest`) instead of `map[string]interface{}`, and perform validation on key fields.

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
