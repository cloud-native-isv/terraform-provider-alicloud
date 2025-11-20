# Feature Specification: Kafka Provider Refactoring

**Feature Branch**: `002-refactor-kafka-provider`  
**Created**: 2025-11-17  
**Status**: Draft  
**Input**: User description: "refactor kafka相关provider的实现，包括如下文件中的代码，data_source_alicloud_alikafka_consumer_groups.go data_source_alicloud_alikafka_instances.go data_source_alicloud_alikafka_sasl_acls.go data_source_alicloud_alikafka_sasl_users.go data_source_alicloud_alikafka_topics.go resource_alicloud_alikafka_consumer_group.go resource_alicloud_alikafka_instance_allowed_ip_attachment.go resource_alicloud_alikafka_instance.go resource_alicloud_alikafka_sasl_acl.go resource_alicloud_alikafka_sasl_user.go resource_alicloud_alikafka_topic.go service_alicloud_alikafka.go，使用/cws_data/cws-lib-go/lib/cloud/aliyun/api/kafka中cws-lib-go库提供的API函数。代码风格可以参考data_source_alicloud_flink_workspaces.go resource_alicloud_flink_workspace.go service_alicloud_flink_workspace.go中的实现"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Modernized Kafka Provider Implementation (Priority: P1)

As a Terraform user managing Alibaba Cloud Kafka resources, I want the provider to use the modern cws-lib-go API layer instead of direct SDK calls, so that I benefit from better error handling, consistent API patterns, and improved maintainability.

**Why this priority**: This is the core technical debt refactoring that enables all other improvements and ensures consistency with the provider's architecture guidelines.

**Independent Test**: Can be fully tested by running existing Kafka provider acceptance tests and verifying that all functionality works identically with the new implementation.

**Acceptance Scenarios**:

1. **Given** an existing Terraform configuration using alicloud_alikafka resources, **When** I run `terraform plan` and `terraform apply`, **Then** the operations succeed with the same behavior as before
2. **Given** a new Terraform configuration for Kafka resources, **When** I create, read, update, and delete resources, **Then** all operations work correctly using the new cws-lib-go API layer

---

### User Story 2 - Consistent Error Handling and State Management (Priority: P2)

As a DevOps engineer, I want consistent error handling and state management across all Kafka resources, so that troubleshooting is easier and resource state transitions are reliable.

**Why this priority**: Consistent error handling improves user experience and reduces support burden, while proper state management ensures resource reliability.

**Independent Test**: Can be tested by simulating error conditions and verifying that appropriate error messages are returned, and by testing state transitions during resource creation/deletion.

**Acceptance Scenarios**:

1. **Given** a Kafka resource creation that fails due to quota limits, **When** I run `terraform apply`, **Then** I receive a clear, actionable error message
2. **Given** a Kafka resource being deleted, **When** I monitor the deletion process, **Then** the provider properly waits for the resource to be completely removed before completing

---

### User Story 3 - Proper Layered Architecture Compliance (Priority: P3)

As a provider maintainer, I want the Kafka implementation to follow the Provider → Resource/DataSource → Service → API layered architecture, so that the codebase remains maintainable and follows established patterns.

**Why this priority**: Architectural compliance ensures long-term maintainability and makes it easier for new developers to understand and contribute to the codebase.

**Independent Test**: Can be verified by code review ensuring that Resources/DataSources only call Service layer functions, and Service layer only calls cws-lib-go API functions.

**Acceptance Scenarios**:

1. **Given** the refactored Kafka provider code, **When** I examine the call hierarchy, **Then** I see clear separation between Resource/DataSource, Service, and API layers
2. **Given** a need to add new Kafka functionality, **When** I look at the code structure, **Then** I can easily identify where to add new methods following the established pattern

---

### Edge Cases

- **Error Code Differences**: When the cws-lib-go API returns different error codes than the old SDK, the provider will map these error codes to maintain consistent error handling behavior for end users.
- **API Rate Limiting**: The system will handle API rate limiting with the new cws-lib-go implementation by implementing exponential backoff with jitter, consistent with other provider resources.
- **Field/Structure Differences**: When cws-lib-go API responses have different field names or structures, the provider will map these fields to maintain backward compatibility with existing Terraform configurations.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All Kafka resources (instances, topics, consumer groups, SASL users, SASL ACLs, IP attachments) MUST use cws-lib-go API functions instead of direct alikafka SDK calls
- **FR-002**: Service layer functions MUST follow the same pattern as Flink service (Describe*, Create*, Delete*, WaitFor* functions)
- **FR-003**: All Resources and DataSources MUST call Service layer functions only, never directly access APIs or SDKs
- **FR-004**: Error handling MUST use the standard WrapError/WrapErrorf patterns with appropriate error classification
- **FR-005**: State management MUST use proper StateRefreshFunc and WaitFor* patterns for resource creation and deletion
- **FR-006**: All ID encoding/decoding functions MUST follow the standard pattern (Encode*Id/Decode*Id)
- **FR-007**: Timeout configurations MUST be set for all CRUD operations (defaults: Create 10m, Update 10m, Delete 5m; resource-specific overrides allowed with justification)
- **FR-008**: All existing functionality MUST be preserved with identical behavior

### Non-functional Requirements (Constitution-aligned)

- **NFR-001 (Layering)**: All Kafka components MUST adhere to Provider → Resource/DataSource → Service → API layering; no direct SDK/HTTP calls from Resources/DataSources.
- **NFR-002 (State Mgmt)**: Create/Delete flows MUST use Service-layer wait functions with proper timeouts; no direct Read invocation inside Create.
- **NFR-003 (Error Handling)**: MUST use wrapped errors and helper predicates; define retryable errors and backoff policies.
- **NFR-004 (Strong Typing)**: MUST use cws-lib-go strong types exclusively; avoid `map[string]interface{}` completely.
- **NFR-005 (Build & Size)**: `make` MUST succeed after refactoring; code files exceeding 1000 LOC MUST be split by module responsibility.

### Key Entities

- **KafkaInstance**: Represents an Alibaba Cloud Kafka instance with properties like instanceId, region, diskSize, ioMax, etc.
- **KafkaTopic**: Represents a Kafka topic within an instance with properties like topic name, partition count, etc.
- **KafkaConsumerGroup**: Represents a consumer group with consumerId and associated metadata
- **KafkaSaslUser**: Represents a SASL user with username and authentication details
- **KafkaSaslAcl**: Represents SASL ACL rules with resource type, operation type, and permissions
- **KafkaAllowedIp**: Represents IP allowlist configuration for Kafka instances

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All existing Kafka provider acceptance tests pass with 100% success rate
- **SC-002**: Code compilation succeeds with `make` command without any errors or warnings
- **SC-003**: Resource creation time remains within 10% of current performance baseline
- **SC-004**: Error messages are consistent with other provider resources and provide actionable information
- **SC-005**: Code review shows 100% compliance with layered architecture guidelines
- **SC-006**: No functional regressions detected in manual testing of all Kafka resource types
