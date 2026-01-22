# Feature Specification: Split OTS Instance Resources

**Feature Branch**: `011-split-ots-instance`
**Created**: 2026-01-20
**Status**: Draft
**Input**: User description: "将 'alicloud_ots_instance': resourceAliCloudOtsInstance(),这个resource拆分成alicloud_ots_instance和alicloud_ots_instance_vcu两个resource，当前他们的逻辑都混在一起，需要将他们从resourceAliCloudOtsInstance中拆分出来，alicloud_ots_instance还是使用resourceAliCloudOtsInstance作为实现，alicloud_ots_instance_vcu使用一个新的resourceAliCloudOtsInstanceVCU作为实现函数。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create VCU Instance (Priority: P1)

Users can create and manage Tablestore instances with VCU specification using a dedicated resource `alicloud_ots_instance_vcu`.

**Why this priority**: VCU instances have specific parameters (like VCU quota) that differ from standard instances. Separation clarifies usage.

**Independent Test**: Can be tested by creating `alicloud_ots_instance_vcu` resource in Terraform and verifying it exists in Aliyun.

**Acceptance Scenarios**:

1. **Given** valid VCU configuration, **When** applying `alicloud_ots_instance_vcu` resource, **Then** a VCU instance is created successfully.
2. **Given** existing VCU instance, **When** updating VCU-specific parameters (e.g., elastic limit), **Then** the instance is updated correctly.
3. **Given** existing VCU instance, **When** destroying the resource, **Then** the instance is deleted.

---

### User Story 2 - Maintain Standard Instance (Priority: P1)

Users can continue to create and manage standard Tablestore instances (SSD, HYBRID) using the existing `alicloud_ots_instance` resource, but with VCU logic removed from its implementation.

**Why this priority**: Ensures backward compatibility for existing workflows while cleaning up the codebase.

**Independent Test**: Can be tested by creating `alicloud_ots_instance` (SSD/HYBRID) and verifying functionality remains unchanged.

**Acceptance Scenarios**:

1. **Given** valid SSD/HYBRID configuration, **When** applying `alicloud_ots_instance`, **Then** the instance is created successfully.
2. **Given** existing SSD/HYBRID instance, **When** updating, **Then** updates apply correctly.
3. **Given** VCU configuration in `alicloud_ots_instance` (if decided to forbid), **When** applying, **Then** it should either fail with helpful error or (if backward compat required) continue to work but warn.
4. **Negative Test**: **Given** a configuration for `alicloud_ots_instance` attempting to set `instance_type` to "VCU" or using VCU-specific fields, **When** running `terraform plan`, **Then** the provider must return a validation error.

### Edge Cases

- **Existing State**: How does `alicloud_ots_instance` behave if it already manages a VCU instance in state?
  - *Mitigation*: **Hard Break**. The resource will no longer support VCU instances. Users MUST migrate their state to `alicloud_ots_instance_vcu` manually (e.g., using `terraform state mv` or import) before upgrading. The provider will returns errors or remove VCU fields from the legacy resource schema.
- **Cross Usage**: User tries to create VCU instance with `alicloud_ots_instance`.
  - *Desired*: Error or validation failure guiding them to `alicloud_ots_instance_vcu`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a new resource `alicloud_ots_instance_vcu`.
- **FR-002**: `alicloud_ots_instance_vcu` MUST create Tablestore instances with "VCU" specification automatically. The `instance_specification` argument MUST be removed from this resource's schema (hidden from user) and hardcoded in the API call.
- **FR-003**: `alicloud_ots_instance_vcu` MUST support VCU-specific parameters (e.g., `elastic_vcu_upper_limit`).
- **FR-004**: The existing `alicloud_ots_instance` resource MUST REMOVE support for "VCU" specification. It only supports "SSD" and "HYBRID".
- **FR-005**: Logic specific to VCU MUST be moved/copied to `resourceAliCloudOtsInstanceVCU` and cleaned up from `resourceAliCloudOtsInstance` where appropriate to separate concerns.
- **FR-006**: Both resources MUST support standard Terraform lifecycle: Create, Read, Update, Delete.
- **FR-007**: The existing `alicloud_ots_instance` resource MUST REMOVE fields that are specific to VCU instances (e.g., `elastic_vcu_upper_limit`, `vcu_quota`) from its schema.
- **FR-008**: Both resources MUST support common instance attributes appropriately: `name`, `description`, `tags`, `access_log_enabled`, `network_type_acl`, etc.

### Key Entities

- **OTS Instance**: Represented by `alicloud_ots_instance` (standard types) and `alicloud_ots_instance_vcu` (VCU type).
- **Instance Specification**: Differentiator (SSD/HYBRID vs VCU).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can successfully create a VCU instance using `alicloud_ots_instance_vcu`.
- **SC-002**: Developers can successfully create an SSD instance using `alicloud_ots_instance`.
- **SC-003**: `terraform plan` shows correct separate resources for respective configurations.
- **SC-004**: Codebase contains two distinct resource implementation functions: `resourceAliCloudOtsInstance` and `resourceAliCloudOtsInstanceVCU`.

## Clarifications

### Session 2026-01-20

- Q: How should the `alicloud_ots_instance` resource handle existing VCU instances found in the state file? → A: **Hard Break**: Remove VCU support entirely. Users must manually migrate state (`terraform state mv`) immediately.
- Q: How should the `instance_specification` parameter be handled in the new `alicloud_ots_instance_vcu` resource? → A: **Remove/Hardcode**: Remove `instance_specification` from the schema. Hardcode "VCU" in the backend API call.
- Q: Should VCU-specific fields (e.g., `elastic_vcu_upper_limit`) be removed from the legacy `alicloud_ots_instance` schema? → A: **Remove**: Remove fields that only apply to VCU instances.

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
