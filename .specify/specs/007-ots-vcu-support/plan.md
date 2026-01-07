# Implementation Plan - Tablestore VCU Instance Support

The goal is to support "VCU" instance specification for `alicloud_ots_instance` including the configuration of elastic VCU upper limits.

## User Review Required

> [!IMPORTANT]
> **Critical items requiring user attention before proceeding**:
>
> - **None**: The necessary concepts conform to the current structure.
> - **Note**: Implementation largely exists; plan focuses on verifying functional correctness and completeness according to spec.

## Proposed Changes

### Configuration
No new configuration. Updating validation for `instance_specification` and handling logic for `elastic_vcu_upper_limit`.

### State Management
State handling for `elastic_vcu_upper_limit` and VCU attributes is mainly done. Plan ensures `Create` and `Update` functions correctly propagate these values via the Service layer.

### Technical Context

- **Files to modify**:
    - `alicloud/resource_alicloud_ots_instance.go`: Core resource logic.
    - `aliyun/api/tablestore/alicloud_tablestore_instance_api.go`: API contracts (verification).
    - `alicloud/service_alicloud_ots_instance.go`: Service layer mappings (verification).

- **Current State Analysis**:
    - `CreateInstance` in API layer has VCU logic.
    - `UpdateInstanceElasticVCUUpperLimit` in API layer exists.
    - `resource_alicloud_ots_instance.go` has schema for `elastic_vcu_upper_limit` and validation for `instance_specification`.
    - `resourceAliCloudOtsInstanceUpdate` handles `elastic_vcu_upper_limit`.

- **Unknowns**:
    - [NEEDS CLARIFICATION: Verify if `instance_specification` requires "VCU" (all caps) or "vcu" or if both are accepted by SDK? Schema validation says `StringInSlice([]string{"SSD", "HYBRID", "VCU"}, false)` which implies case-sensitive matching.]
    - [NEEDS CLARIFICATION: Does existing code correctly populate `elastic_vcu_upper_limit` in `Read` or `convertTablestoreInstanceToSchema`? Current code snippet for `convertTablestoreInstanceToSchema` shows it sets `elastic_vcu_upper_limit`.]

## Constitution Check

| Principle | Status | Notes |
|:--- |:--- |:--- |
| **I. Layering** | ✅ | Resource -> Service -> API layering is preserved. |
| **II. State Mgmt** | ✅ | Uses wait functions and StateRefreshFunc. |
| **III. Errors** | ✅ | Uses unified error handling. |
| **IV. Code Quality** | ✅ | Adheres to naming conventions. |
| **V. Strong Typing** | ✅ | Uses defined structs, assumes `cws-lib-go` provides them. |

## Automation & Testing Requirements

- **Unit Tests**:
    - `TestAccAlicloudOtsInstance_vcu`: Test creating a VCU instance.
    - `TestAccAlicloudOtsInstance_updateVcuLimit`: Test updating the limit.
- **Manual Verification**:
    - `terraform apply` with a VCU configuration.

## Sequential Plan

### Phase 0: Research & Verification

1.  **Verify API Constants**: Check `cws-lib-go` or local definition for valid `InstanceSpecification` values.
    - *Action*: Confirm "VCU" string is correct.
2.  **Verify Read/Update Logic**: Ensure `elastic_vcu_upper_limit` is read back from API response and set in state.
    - *Action*: Inspect `alicloud_tablestore_instance_api.go` function `convertToTablestoreInstance`.

### Phase 1: Implementation / Fixes

1.  **Refine Resource Schema (if needed)**:
    - Ensure `elastic_vcu_upper_limit` is marked `Optional: true, Computed: true`. (It is).
    - Ensure `instance_specification` validation allows "VCU". (It does).

2.  **Logic Verification**:
    - `resourceAliCloudOtsInstanceCreate`: Confirm VCU branch logic (wait for status, then update limit if separate call needed? API `CreateInstance` checks for VCU and calls `CreateVCUInstance` then `UpdateInstanceElasticVCUUpperLimit` if needed. This looks handled in API layer).
    - `resourceAliCloudOtsInstanceUpdate`: Confirm `elastic_vcu_upper_limit` update call is correct.
    - `resourceAliCloudOtsInstanceRead`: Ensure `convertTablestoreInstanceToSchema` correctly maps the field.

3.  **Tests**:
    - Add acceptance test case for VCU instance.

### Phase 2: Final Review

1.  **Run Tests**: Execute `TestAccAlicloudOtsInstance_vcu`.
2.  **Code Review**: Self-review against Constitution.

## Verification Plan

### Automated Tests
- `TestAccAlicloudOtsInstance_vcu`
- `TestAccAlicloudOtsInstance_updateVcuLimit`

### Manual Verification
- N/A (Automated tests preferred)
