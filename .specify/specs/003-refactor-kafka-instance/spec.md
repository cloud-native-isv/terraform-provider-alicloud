# Feature Specification: Refactor Kafka Instance API

**Feature Branch**: `003-refactor-kafka-instance`
**Created**: 2025-11-20
**Status**: Draft
**Input**: User description: "将kafka instance resource中的client.RpcPost调用替换为cws-lib-go API的方法，相关代码在/cws_data/cws-lib-go/lib/cloud/aliyun/api/kafka目录，当前项目中已经完成了go mod的replace可以直接使用。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Kafka Instance (Priority: P1)

As a Terraform user, I want to create a Kafka instance (both PrePaid and PostPaid) so that I can use it for message queuing.

**Why this priority**: Core functionality of the resource.

**Independent Test**: Run `terraform apply` with a configuration for `alicloud_alikafka_instance` and verify the instance is created in Alibaba Cloud.

**Acceptance Scenarios**:

1. **Given** a valid Terraform configuration for a PostPaid Kafka instance, **When** I run `terraform apply`, **Then** the instance is created successfully.
2. **Given** a valid Terraform configuration for a PrePaid Kafka instance, **When** I run `terraform apply`, **Then** the instance is created successfully.

---

### User Story 2 - Read Kafka Instance (Priority: P1)

As a Terraform user, I want to read the state of an existing Kafka instance so that Terraform can manage it.

**Why this priority**: Essential for Terraform state management.

**Independent Test**: Run `terraform plan` after creation and verify no changes are detected if the configuration hasn't changed.

**Acceptance Scenarios**:

1. **Given** an existing Kafka instance managed by Terraform, **When** I run `terraform plan`, **Then** the current state matches the configuration.

---

### User Story 3 - Update Kafka Instance (Priority: P2)

As a Terraform user, I want to update the configuration of a Kafka instance (e.g., disk size, partition number) so that I can scale it.

**Why this priority**: Important for lifecycle management.

**Independent Test**: Change a parameter in the Terraform configuration and run `terraform apply`.

**Acceptance Scenarios**:

1. **Given** an existing Kafka instance, **When** I increase the disk size in the configuration and run `terraform apply`, **Then** the instance disk size is updated.
2. **Given** an existing Kafka instance, **When** I change the partition number in the configuration and run `terraform apply`, **Then** the partition number is updated.

---

### User Story 4 - Delete Kafka Instance (Priority: P2)

As a Terraform user, I want to delete a Kafka instance so that I can clean up resources.

**Why this priority**: Essential for resource cleanup.

**Independent Test**: Run `terraform destroy`.

**Acceptance Scenarios**:

1. **Given** an existing Kafka instance, **When** I run `terraform destroy`, **Then** the instance is deleted from Alibaba Cloud.

### Edge Cases

- What happens when the API returns an error (e.g., throttling)? The provider should handle it gracefully (retry or report error).
- What happens when the instance is not found during Read? The provider should remove it from state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `alicloud_alikafka_instance` resource MUST use the `cws-lib-go` library for all API interactions with Alibaba Cloud.
- **FR-002**: The refactoring MUST NOT change the external behavior of the resource (inputs, outputs, state management).
- **FR-003**: The implementation MUST follow the provider's architecture guidelines (Resource -> Service -> API).
- **FR-004**: The implementation MUST handle all existing parameters supported by the resource.
- **FR-005**: The implementation MUST support both PrePaid and PostPaid payment types.
- **FR-006**: The implementation MUST support all update operations currently supported (e.g., `ModifyInstanceName`, `UpgradePostPayOrder`, `UpgradePrePayOrder`, `UpgradeInstanceVersion`, `UpdateInstanceConfig`, `ChangeResourceGroup`, `EnableAutoGroupCreation`, `EnableAutoTopicCreation`).

### Key Entities *(include if feature involves data)*

- **KafkaInstance**: Represents the Alibaba Cloud Kafka instance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of `client.RpcPost` calls in `resource_alicloud_alikafka_instance.go` are replaced with `cws-lib-go` API calls (or Service layer calls wrapping them).
- **SC-002**: Existing acceptance tests for `alicloud_alikafka_instance` pass without modification.
- **SC-003**: The code compiles without errors.
