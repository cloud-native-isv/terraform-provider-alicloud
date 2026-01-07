# Feature Specification: Tablestore VCU Instance Support

**Feature Branch**: `007-ots-vcu-support`  
**Created**: 2026-01-06  
**Status**: Draft  
**Input**: User description: "在terraform层支持tablestore 的VCU实例配置"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create VCU Instance (Priority: P1)

Users need to be able to create a Tablestore instance using the VCU (Virtual Compute Unit) pricing model. This allows for serverless, improved elasticity and paying for what is used.

**Why this priority**: Core feature request to support the new instance specification type.

**Independent Test**: Can be fully tested by creating a resource with `instance_specification = "VCU"` and verifying it is created successfully in the provider and API.

**Acceptance Scenarios**:

1. **Given** a new `alicloud_ots_instance` resource configuration with `instance_specification = "VCU"`, **When** `terraform apply` is run, **Then** a Tablestore instance with VCU specification should be created.
2. **Given** a VCU instance configuration, **When** `elastic_vcu_upper_limit` is specified, **Then** the instance should be created with that limit.

---

### User Story 2 - Update VCU Configuration (Priority: P2)

Users need to be able to modify the elastic VCU upper limit of an existing VCU instance to manage costs and scaling limits.

**Why this priority**: Essential management capability for VCU instances.

**Independent Test**: Can be tested by changing `elastic_vcu_upper_limit` on an existing VCU instance and running apply.

**Acceptance Scenarios**:

1. **Given** an existing `alicloud_ots_instance` of type VCU, **When** `elastic_vcu_upper_limit` is changed in configuration and applied, **Then** the instance limit should be updated via API without replacing the instance.

---

### Edge Cases

- What happens when `elastic_vcu_upper_limit` is set for non-VCU instances (e.g., SSD/HYBRID)?
  - Expectation: Should probably be ignored or return an error/warning depending on API behavior (Provider might suppress or validation might fail if added).
- What happens when `instance_specification` is changed from "SSD" to "VCU"?
  - Expectation: `ForceNew` behavior (which is already defined in schema).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The provider MUST support `VCU` as a valid value for `instance_specification` field.
- **FR-002**: The provider MUST support `elastic_vcu_upper_limit` field allowing float values to configure the burst limit.
- **FR-003**: The provider MUST correctly call the VCU-specific creation API (or generic API with correct parameters) when "VCU" is selected.
- **FR-004**: The provider MUST support updating `elastic_vcu_upper_limit` in-place.
- **FR-005**: The provider MUST correctly read back VCU-specific attributes into the state.

### Key Entities

- **OtsInstance**: Represents the Tablestore instance resource.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully provision a Tablestore instance with `instance_specification="VCU"`.
- **SC-002**: Users can successfully set `elastic_vcu_upper_limit` to a valid float value (e.g., 1.0, 2.5).
- **SC-003**: `terraform plan` shows no diff after a successful apply of a VCU instance (idempotency).

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
