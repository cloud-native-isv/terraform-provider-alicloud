# Feature Specification: Split AliKafka Instance Resource

**Feature Branch**: `006-split-alikafka-instance`  
**Created**: 2025-12-23  
**Status**: Draft  
**Input**: User description: "将"alicloud_alikafka_instance": resourceAliCloudAlikafkaInstance(),这个resource拆分成alicloud_alikafka_instance和alicloud_alikafka_deployment两个resource，当前的主要实现是在func resourceAliCloudAlikafkaInstanceCreate中先使用CreatePostPayOrder和CreatePrePayOrder创建instance对象，然后使用kafkaService.StartInstance启动实例，我需要将这个流程改为alicloud_alikafka_instance调用CreatePrePayOrder API创建实例，alicloud_alikafka_deployment调用kafkaService.StartInstance和kafkaService.StopInstance部署实例，将这两个流程解耦，以实现更灵活的资源管理。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Instance Only (Priority: P1)

As a user, I want to create an AliKafka instance without immediately starting it, so that I can configure it or manage costs before it becomes active.

**Why this priority**: This is the core of the decoupling request. It allows separation of creation and deployment.

**Independent Test**: Create an `alicloud_alikafka_instance` resource. Verify via API or console that the instance exists but is in a non-running state (e.g., "Created" or "Stopped").

**Acceptance Scenarios**:

1. **Given** a valid AliKafka instance configuration, **When** `terraform apply` is run with `alicloud_alikafka_instance`, **Then** the instance is created successfully.
2. **Given** the instance is created, **When** checking the instance status, **Then** it is NOT in "Running" state (unless the API defaults to running, in which case the provider must not explicitly wait for it or trigger start if possible, but based on the request, the start is a separate step).

### User Story 2 - Deploy Instance (Priority: P1)

As a user, I want to start a created AliKafka instance using a separate resource, so that I can control when the service becomes available.

**Why this priority**: Enables the second half of the decoupled workflow.

**Independent Test**: Create an `alicloud_alikafka_instance` and an `alicloud_alikafka_deployment` referencing it. Verify the instance transitions to "Running".

**Acceptance Scenarios**:

1. **Given** an existing `alicloud_alikafka_instance` ID, **When** `terraform apply` is run with `alicloud_alikafka_deployment`, **Then** the instance transitions to "Running" state.
2. **Given** the deployment resource exists, **When** `terraform plan` is run, **Then** no changes are detected if the instance is still running.

### User Story 3 - Stop Instance (Priority: P2)

As a user, I want to stop a running instance by removing the deployment resource, so that I can save costs or perform maintenance without deleting the instance data.

**Why this priority**: Provides lifecycle management for the running state.

**Independent Test**: Destroy the `alicloud_alikafka_deployment` resource. Verify the instance transitions to "Stopped" (or equivalent) state but is not deleted.

**Acceptance Scenarios**:

1. **Given** a running instance managed by `alicloud_alikafka_deployment`, **When** `terraform destroy` is run for the deployment resource, **Then** the instance is stopped.
2. **Given** the instance is stopped, **When** checking the instance existence, **Then** the instance still exists.

### Edge Cases

- **Deployment for non-existent instance**: Should fail with a clear error.
- **Deployment for already running instance**: Should succeed (adopt state) or fail depending on implementation preference (usually adopt).
- **Deleting instance while deployment exists**: Terraform dependency graph should handle this (deployment deleted first). If forced, deployment might lose its target.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `alicloud_alikafka_instance` resource MUST create an AliKafka instance using `CreatePrePayOrder` or `CreatePostPayOrder` APIs (depending on configuration).
- **FR-002**: The `alicloud_alikafka_instance` resource MUST NOT explicitly call `StartInstance` during creation.
- **FR-003**: A new resource `alicloud_alikafka_deployment` MUST be implemented.
- **FR-004**: `alicloud_alikafka_deployment` MUST accept an `instance_id` parameter (Required, ForceNew).
- **FR-005**: `alicloud_alikafka_deployment` MUST call `kafkaService.StartInstance` during its Create phase.
- **FR-006**: `alicloud_alikafka_deployment` MUST call `kafkaService.StopInstance` during its Delete phase.
- **FR-007**: `alicloud_alikafka_deployment` MUST wait for the instance to reach "Running" state during Create.
- **FR-008**: `alicloud_alikafka_deployment` MUST wait for the instance to reach a non-running state (e.g., "Stopped") during Delete.
- **FR-009**: `alicloud_alikafka_instance` Read operation MUST NOT enforce "Stopped" state; it MUST succeed if the instance exists regardless of status.
- **FR-010**: `alicloud_alikafka_deployment` Read operation MUST detect drift if the instance is not in "Running" state (setting state to trigger update).

### Key Entities *(include if feature involves data)*

- **AliKafka Instance**: The cloud resource representing the Kafka cluster.
- **AliKafka Deployment**: A logical resource representing the "Running" state of the instance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully create an instance (`alicloud_alikafka_instance`) without it entering the "Running" state (assuming API supports this separation).
- **SC-002**: Users can successfully start an instance by applying `alicloud_alikafka_deployment`.
- **SC-003**: Users can successfully stop an instance by destroying `alicloud_alikafka_deployment`.
- **SC-004**: The `alicloud_alikafka_instance` code no longer contains `StartInstance` logic in its Create method.

## Clarifications

### Session 2025-12-23

- Q: Backward Compatibility & Breaking Change: Removing `StartInstance` from `alicloud_alikafka_instance` is a breaking change. How should this be handled? → A: **Strict Split**: `alicloud_alikafka_instance` *only* creates (stops after CreateOrder). Users *must* add `alicloud_alikafka_deployment` to start it.
- Q: `alicloud_alikafka_instance` State Enforcement: How should `alicloud_alikafka_instance` behave if it finds the instance is already "Running"? → A: **Existence Only**: The resource checks if the instance exists. It ignores whether it is Running or Stopped. It does NOT attempt to stop a running instance.
- Q: Drift Detection for `alicloud_alikafka_deployment`: If an instance is stopped manually, what should `terraform plan` detect? → A: **Detect and Correct**: Plan shows a diff (status changed from Running to Stopped) and apply restarts the instance.

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
