# Feature Specification: Refactor Tablestore Resources

**Feature Branch**: `005-refactor-ots-resources`
**Created**: 2026-01-28
**Status**: Draft
**Input**: User description: "pkg/cws-lib-go submodule refactor split tablestore instance, update provider implementation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Standard Tablestore Instance (Priority: P1)

Users need to be able to create standard Tablestore instances (SSD, HYBRID) using the updated provider which aligns with the underlying library changes.

**Why this priority**: Core functionality broken by library update.

**Independent Test**: Create a resource `alicloud_ots_instance` with type SSD or HYBRID.

**Acceptance Scenarios**:

1. **Given** a terraform configuration for `alicloud_ots_instance` with `instance_specification`="SSD", **When** apply is run, **Then** creating the instance succeeds using `TablestoreInstance` struct.

---

### User Story 2 - Create VCU Tablestore Instance (Priority: P1)

Users need to be able to create VCU Tablestore instances using the updated provider and the designated VCU struct.

**Why this priority**: Core functionality broken by library update.

**Independent Test**: Create a resource `alicloud_ots_instance_vcu`.

**Acceptance Scenarios**:

1. **Given** a terraform configuration for `alicloud_ots_instance_vcu`, **When** apply is run, **Then** creating the instance succeeds using `TablestoreVCUInstance` struct via `CreateVCUInstance` API.

---

### User Story 3 - Read Tablestore Instances (Priority: P1)

Users expect to read back state correctly for both standard and VCU instances.

**Why this priority**: Required for Terraform state management.

**Independent Test**: Refresh state of existing instances.

**Acceptance Scenarios**:

1. **Given** an existing instance, **When** refresh is run, **Then** state is updated using `TablestoreInstanceInfo` Read Model.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `resource_alicloud_ots_instance` Create method MUST use `TablestoreInstance` struct and `CreateInstance` API.
- **FR-002**: `resource_alicloud_ots_instance_vcu` Create method MUST use `TablestoreVCUInstance` struct and `CreateVCUInstance` API.
- **FR-003**: Both resources Read method MUST adapt to `TablestoreInstanceInfo` struct returned by `GetInstance`.
- **FR-004**: Update methods MUST use `TablestoreInstanceUpdate` struct if applicable or individual update methods.

### Key Entities

- **TablestoreInstance**: New struct for standard creation.
- **TablestoreVCUInstance**: New struct for VCU creation.
- **TablestoreInstanceInfo**: Shared struct for reading instance details.
