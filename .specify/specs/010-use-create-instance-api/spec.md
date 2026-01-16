# Feature Specification: Use CreateInstance API for Instance Resource

**Feature Branch**: `010-use-create-instance-api`  
**Created**: 2026-01-15  
**Status**: Draft  
**Input**: User description: "参考json文件中的接口定义，将resourceAliCloudInstanceCreate中的RunInstances替换为CreateInstance接口，尽可能保证对外接口也就是schema部分的向前兼容。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create ECS Instance (Priority: P1)

As a Terraform user, I want to provision an `alicloud_instance` resource as I did before, but using the `CreateInstance` API backend, so that I can manage my infrastructure using the specified API definitions.

**Why this priority**: Core functionality of the resource.

**Independent Test**: Define a standard `alicloud_instance` resource in a Terraform configuration, run `terraform apply`, and verify the instance is created and running in the Alibaba Cloud console.

**Acceptance Scenarios**:

1. **Given** a valid `alicloud_instance` configuration, **When** `terraform apply` is executed, **Then** the instance is created using `CreateInstance` API and reaches `Running` status.
2. **Given** an invalid configuration (e.g., invalid instance type), **When** `terraform apply` is executed, **Then** the provider returns a proper error message from the API.

### Edge Cases

- What happens when `CreateInstance` creates a generic error? (Should return error to user)
- What happens if `StartInstance` fails? (Should return error and maybe taint resource)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `resourceAliCloudInstanceCreate` function MUST use the `CreateInstance` API instead of `RunInstances`.
- **FR-002**: The implementation MUST map existing Terraform schema attributes to `CreateInstance` request parameters according to the provided API definition.
- **FR-003**: The implementation MUST maintain backward compatibility for the Terraform schema (no attributes removed or renamed).
- **FR-004**: Implementation MUST always call `StartInstance` immediately after creation (once instance is Stopped) to match legacy `RunInstances` behavior, even if `status` is set to "Stopped".
- **FR-005**: The implementation MUST handle the asynchronous nature of `CreateInstance` (waiting for `Stopped` status before starting, then waiting for `Running`).
- **FR-006**: Error handling MUST be updated to reflect `CreateInstance` specific errors if any, though generic API errors should remain similar.
- **FR-007**: If multiple `security_groups` are provided, the implementation MUST create the instance with the first group, then iteratively use `JoinSecurityGroup` to associate the remaining groups.

### Key Entities

- **Instance**: The ECS instance resource being managed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Existing `alicloud_instance` acceptance tests pass without modification to the test cases.
- **SC-002**: The resource creation flow successfully transitions: API Create -> Wait for Stopped -> API Start -> Wait for Running.
- **SC-003**: No regression in supported schema parameters (backward compatibility).

## Clarifications

### Session 2026-01-15

- Q: The `alicloud_instance` resource supports a `status` argument (Default: "Running", Options: "Running", "Stopped"). The new `CreateInstance` API creates instances in a `Stopped` state by default. How should the creation logic handle the `status` argument? → A: **Always Start**: Always call `StartInstance` to match legacy `RunInstances` behavior, ignoring the optimization opportunity for `status="Stopped"`.
- Q: The `alicloud_instance` resource allows configuring multiple `security_groups` (Set). The provided `CreateInstance` API definition only lists a singular `SecurityGroupId` parameter. How should the implementation handle multiple security groups during creation? → A: **Post-Create Association**: Create the instance with the first security group, then iteratively call `JoinSecurityGroup` for the remaining groups before reporting success.
- Q: The `CreateInstance` API documentation states it does **not** allocate a public IP address during creation, even if bandwidth is specified. The legacy `RunInstances` API (and current Terraform resource behavior) allocates a public IP when `internet_max_bandwidth_out` > 0. How should the provider handle public IP allocation? → A: **Ignore Public IP**: Do not allocate a public IP automatically. Users must attach an EIP resource separately (Breaking Change).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `resourceAliCloudInstanceCreate` function MUST use the `CreateInstance` API instead of `RunInstances`.
- **FR-002**: The implementation MUST map existing Terraform schema attributes to `CreateInstance` request parameters according to the provided API definition.
- **FR-003**: The implementation MUST maintain backward compatibility for the Terraform schema (no attributes removed or renamed).
- **FR-004**: Implementation MUST always call `StartInstance` immediately after creation (once instance is Stopped) to match legacy `RunInstances` behavior, even if `status` is set to "Stopped".
- **FR-005**: The implementation MUST handle the asynchronous nature of `CreateInstance` (waiting for `Stopped` status before starting, then waiting for `Running`).
- **FR-006**: Error handling MUST be updated to reflect `CreateInstance` specific errors if any, though generic API errors should remain similar.
- **FR-007**: If multiple `security_groups` are provided, the implementation MUST create the instance with the first group, then iteratively use `JoinSecurityGroup` to associate the remaining groups.
- **FR-008**: If `internet_max_bandwidth_out` > 0 is provided, the implementation MUST NOT automatically allocate a public IP address (unlike legacy `RunInstances` behavior). Users are expected to manage public IPs separately (e.g., via EIP). This is an explicit deviation from behavioral backward compatibility.
