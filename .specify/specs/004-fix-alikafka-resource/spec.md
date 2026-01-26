# Feature Specification: Fix AliKafka Resource Implementation

**Feature Branch**: `004-fix-alikafka-resource`  
**Created**: 2026-01-25  
**Status**: Draft  
**Input**: User description: "全面修复和优化alicloud_alikafka相关resource实现函数中的问题，当前主要存在的问题包括：1）state状态同步的问题，有很多字段在apply之后没有进行持久化，导致tfstate文件中没有记录的值。2）provider实现没有更新大API层的最新实现，pkg/cws-lib-go/lib/cloud/aliyun/api/kafka这个子模块目录中提供了API层的封装，需要将更新到API中。3）状态同步问题，aliyun上的kafka实例存在多种状态，包括刚创建订单（resourceAliCloudAlikafkaInstanceCreate）之后的未部署状态，启动Instance（resourceAliCloudAlikafkaDeploymentCreate）之后的运行状态，这些状态需要进行同步之后才能进行下一步。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Complete State Persistence (Priority: P1)

As a Terraform user managing AliKafka instances, I want all resource attributes to be properly synchronized and persisted in the Terraform state file after any operation (create, update, read), so that my infrastructure state accurately reflects the actual cloud resources and prevents drift or unexpected behavior during subsequent applies.

**Why this priority**: This is critical for Terraform's core functionality - state consistency. Without proper state persistence, users cannot reliably manage their infrastructure, leading to potential data loss, configuration drift, and operational issues.

**Independent Test**: Can be fully tested by creating an AliKafka instance with all available attributes, verifying that all computed and configured fields are present in the tfstate file, and confirming that subsequent terraform plan operations show no unexpected changes.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with `alicloud_alikafka_instance` resource with all supported attributes, **When** `terraform apply` is executed successfully, **Then** all computed fields (vpc_id, vswitch_id, zone_id, endpoints, etc.) are stored in the tfstate file
2. **Given** an existing AliKafka instance managed by Terraform, **When** `terraform refresh` is executed, **Then** all resource attributes are correctly synchronized from the cloud provider to the local state without data loss

---

### User Story 2 - Modern API Integration (Priority: P1)

As a Terraform provider maintainer, I want the AliKafka resource implementation to use the latest CWS-Lib-Go API layer (`pkg/cws-lib-go/lib/cloud/aliyun/api/kafka`) instead of legacy direct API calls, so that the provider benefits from standardized error handling, type safety, and ongoing API improvements.

**Why this priority**: Using the modern API layer ensures consistency with other resources in the provider, reduces maintenance overhead, and provides better error handling and type safety. This is foundational for long-term maintainability.

**Independent Test**: Can be fully tested by verifying that all AliKafka resource operations (create, read, update, delete) exclusively use the CWS-Lib-Go Kafka API functions rather than direct HTTP calls or legacy SDK methods.

**Acceptance Scenarios**:

1. **Given** the AliKafka resource implementation code, **When** reviewed for API usage patterns, **Then** all API calls use the `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka` package functions
2. **Given** a test environment with the updated provider, **When** executing AliKafka resource operations, **Then** all operations succeed using the modern API layer without falling back to legacy methods

---

### User Story 3 - Proper State Lifecycle Management (Priority: P1)

As a Terraform user deploying AliKafka instances, I want the provider to properly handle the multi-stage lifecycle of Kafka instances (order creation → deployment → running state), so that Terraform waits for each stage to complete before proceeding and accurately reflects the current state at all times.

**Why this priority**: AliKafka instances have complex state transitions that must be properly handled to prevent Terraform from proceeding before resources are ready, which could cause failures in dependent resources or inconsistent state.

**Independent Test**: Can be fully tested by creating an AliKafka instance and verifying that Terraform properly waits for the order creation to complete, then waits for deployment to complete, and finally confirms the instance is in a running state before marking the resource as created.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration creating an `alicloud_alikafka_instance`, **When** `terraform apply` is executed, **Then** the provider first waits for the order to be processed and instance ID to be available, then proceeds to deployment if needed
2. **Given** an AliKafka instance in various states (creating, deploying, running, stopped), **When** `terraform refresh` is executed, **Then** the provider correctly identifies and reports the current state without errors

---

### Edge Cases

- What happens when the AliKafka instance creation fails during the order processing phase?
- How does the system handle partial deployment where some zones succeed but others fail?
- What occurs when the API returns inconsistent state information between different endpoints?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST persist all computed resource attributes in Terraform state after create/update operations, including all fields available in the CWS-Lib-Go API responses. This includes basic fields (vpc_id, vswitch_id, zone_id, endpoints, security_group, service_version, config, kms_key_id, eip_max) and derived complex fields like vswitch_ids and selected_zones (derived from basic zone and network information). Quota-related fields (topic_num_of_buy, topic_used, topic_left, partition_used, partition_left, group_used, group_left, is_partition_buy) should be included if they are available in the API response.
- **FR-002**: System MUST use only the CWS-Lib-Go Kafka API layer (`github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka`) for all AliKafka resource operations, specifically using CreateInstance directly for all creation operations (bypassing the order system entirely), GetInstance for read operations, UpdateInstance for updates, and StartInstance/StopInstance for deployment operations. This eliminates any direct HTTP calls or legacy SDK usage.
- **FR-003**: System MUST implement proper state transition handling for AliKafka instances with explicit state definitions: state 0 = Order Processing (waiting for instance ID assignment), state 2 = Creating (infrastructure provisioning), state 3 = Configuring (applying instance settings), state 4 = Starting (service initialization), state 5 = Running (fully operational). The system must recognize these states and wait appropriately during create/update operations.
- **FR-004**: System MUST ensure that `alicloud_alikafka_instance` and `alicloud_alikafka_deployment` resources share consistent state management logic and use the same underlying Service layer functions, with both resources using the same GetInstance function for state synchronization.
- **FR-005**: System MUST implement comprehensive error handling using the standard error classification patterns (IsNotFoundError, IsAlreadyExistError, NeedRetry) as defined in the development guide, with appropriate error handling for each CWS-Lib-Go API function call.
- **FR-006**: System MUST handle instance creation using CreateInstance directly (not order-based operations), and deployment operations using StartInstance/StopInstance, ensuring that deployment only occurs when necessary and state is synchronized appropriately through the shared GetInstance function.
- **FR-007**: System MUST maintain backward compatibility with existing Terraform configurations while fixing the state persistence issues

### Key Entities

- **AliKafkaInstance**: Represents a Kafka instance in Alibaba Cloud with attributes including deployment type, disk configuration, network settings, security configuration, and operational status
- **AliKafkaDeployment**: Represents the deployment configuration of a Kafka instance, including VPC/VSwitch settings, zone configuration, and runtime parameters
- **KafkaInstanceState**: Represents the lifecycle state of a Kafka instance, transitioning through order processing, creation, configuration, startup, and running states

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of computed resource attributes are consistently persisted in Terraform state files after all operations (create, update, refresh)
- **SC-002**: All AliKafka resource operations execute successfully using only the CWS-Lib-Go API layer with no legacy API calls detected in code review
- **SC-003**: Terraform apply operations for AliKafka resources complete within expected timeframes (under 60 minutes for create, under 120 minutes for update) while properly waiting for all state transitions
- **SC-004**: Users report zero state drift issues related to missing or inconsistent AliKafka resource attributes in subsequent terraform plan operations
- **SC-005**: The provider passes all existing acceptance tests while adding new tests specifically covering state persistence and API layer integration

## Clarifications

### Session 2026-01-25

- Q: What should be the precise definition and handling of each AliKafka instance state (0, 2, 3, 4, 5) in terms of user-visible behavior, valid transitions, and Terraform state management? → A: Define explicit state mappings: 0=Order Processing (waiting for instance ID), 2=Creating (infrastructure provisioning), 3=Configuring (applying settings), 4=Starting (service initialization), 5=Running (fully operational)
- Q: Which specific CWS-Lib-Go Kafka API functions should be used for each AliKafka resource operation (create, read, update, delete, deploy, undeploy) to ensure proper state management and consistency with the two-step order→instance creation process? → A: Use CreateInstance directly for all creation operations, bypassing the order system entirely since CWS-Lib-Go provides this function.
- Q: Which fields should be persisted in the Terraform state for AliKafka instances? Should we include only the fields directly available in the CWS-Lib-Go KafkaInstance struct, or also include derived/complex fields like vswitch_ids, selected_zones, cross_zone, and quota-related fields (topic_num_of_buy, partition_used, etc.) that may require additional API calls or processing? → A: Include all fields that are available in the CWS-Lib-Go API responses, and derive complex fields like vswitch_ids and selected_zones from the basic zone and network information. Include quota-related fields if they are available in the API response.
