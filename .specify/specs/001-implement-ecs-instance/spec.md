# Feature Specification: Implement alicloud_ecs_instance Resource

**Feature Branch**: `001-implement-ecs-instance`
**Created**: 2026-01-22
**Status**: Draft
**Input**: User description: "Complete the implementation of resourceAliCloudEcsInstance. It differs from resourceAliCloudInstance by using CreateInstance instead of RunInstances, but other logic is the same."

## Clarifications

### Session 2026-01-22

- Q: The native `CreateInstance` API creates instances in a `Stopped` state, whereas the existing `alicloud_instance` resource (based on `RunInstances`) defaults to `Running`. Should `alicloud_ecs_instance` default to `Running` (implicitly auto-start after creation) or `Stopped`?
    - A: **Running** (Auto-start) - Consistency with the existing `alicloud_instance` resource ensures a familiar user experience.
- Q: `alicloud_instance` contains a large number of fields, some of which are legacy or deprecated (e.g., Classic network params). Should the new `alicloud_ecs_instance` strictly replicate the *full* schema for maximum compatibility, or implement a *Clean/Modern* subset?
  - A: **Clean Subset** - Avoids legacy debt; focuses on modern VPC-based ECS features.
- Q: The `CreateInstance` API does not automatically allocate a public IP or wait for ENI attachment completion in the same way `RunInstances` might bundle these. Should the resource implementation explicitly wait for network interface readiness before transitioning to `Running` state?
  - A: **Wait for Network** - Ensures instance is fully reachable and network dependencies are satisfied before resource is marked ready.
- Q: `alicloud_instance` implements complex provider-side logic to generate default names (`instance_name`, `host_name`) and passwords if they are unspecified. Should the new `alicloud_ecs_instance` rely on pure **API-side defaults** (lighter implementation) or retain **Provider-side generation** (consistent with legacy behavior)?
  - A: **API Defaults** - Keeps the resource implementation lightweight and aligned with the "Clean Subset" strategy.
- Q: `CreateInstance` allows minimal configuration of data disks and ENIs inline. For attached resources (Data Disks, Secondary ENIs), should the provider strictly use **Inline Configuration** (passed directly to `CreateInstance`) or separate **Create-then-Attach** logic (using `AttachDisk`/`AttachNetworkInterface`)?
  - A: **Create then Attach** - Reduces complexity of the initial creation call and allows for more robust error handling for attachments.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create ECS Instance (Priority: P1)

As a Terraform user, I want to define and provision an ECS instance using the `alicloud_ecs_instance` resource, so that I can manage my infrastructure using the `CreateInstance` API semantics.

**Why this priority**: Core functionality of the resource.

**Independent Test**: Create a `.tf` file with `alicloud_ecs_instance`, run `terraform apply`, and verify the instance exists in the cloud Console.

**Acceptance Scenarios**:

1. **Given** a valid `alicloud_ecs_instance` configuration, **When** `terraform apply` is executed, **Then** an ECS instance is created using `CreateInstance` API.
2. **Given** the instance is created (which defaults to Stopped state via `CreateInstance`), **When** the `status` is set to "Running" (default), **Then** the provider automatically starts the instance to "Running" state.

---

### User Story 2 - Lifecycle Management (Priority: P2)

As a Terraform user, I want to update and delete the `alicloud_ecs_instance`, so that I can manage the full lifecycle of the resource.

**Why this priority**: Essential for IaC resource to be useful.

**Independent Test**: Update a tag or modify description in `.tf`, run `apply`. Then destroy the resource.

**Acceptance Scenarios**:

1. **Given** an existing `alicloud_ecs_instance`, **When** I change the `description` and `terraform apply`, **Then** the instance description is updated.
2. **Given** an existing `alicloud_ecs_instance`, **When** I run `terraform destroy`, **Then** the instance is terminated.

### Edge Cases

- **Create Fails**: If `CreateInstance` fails (e.g., quota exceeded), Terraform should report a clear error.
- **Start Fails**: If `StartInstance` fails after creation, the resource should be tainted or report error.
- **Instance Already Exists**: Standard Terraform duplicate handling.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The resource `alicloud_ecs_instance` MUST use the Alibaba Cloud SDK `create_instance` (or `CreateInstance`) method to provision the instance.
- **FR-002**: The resource MUST default to `Running` status. It must automatically call `StartInstance` after creation and wait for the `Running` state, unless explicitly configured otherwise (e.g., status set to `Stopped`).
- **FR-003**: The schema MUST be a modern clean subset of `alicloud_instance`, excluding deprecated or Classic network-only fields. The focus MUST be on VPC-based instances.
- **FR-004**: The resource CRUD operations (Read, Update, Delete) MUST function identically to `alicloud_instance` where the underlying API supports it.
- **FR-005**: The resource creation flow MUST explicitly wait for primary network interface readiness before transitioning the instance to `Running` state.
- **FR-006**: The resource MUST rely on API defaults for optional fields like `instance_name` and `host_name`, avoiding complex provider-side default generation logic.
- **FR-007**: The resource MUST uses a "Create then Attach" strategy for secondary resources (Data Disks, Secondary ENIs), invoking `CreateInstance` for the instance first, followed by separate API calls for attachments.

### Success Criteria

1.  Successful creation of ECS instance using the new resource.
2.  Verification that `CreateInstance` API was used.
3.  The resource supports basic updates (e.g., tags, name).
4.  The resource supports deletion.

### Key Entities

- **ECS Instance**: The cloud resource being managed.
- **Terraform Resource (`alicloud_ecs_instance`)**: The representation in Terraform.
