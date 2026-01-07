# Implementation Tasks - Tablestore VCU Instance Support

> **Feature**: Tablestore VCU Instance Support (007-ots-vcu-support)

## Phase 1: Setup & Pre-computation

**Goal**: Verify existing codebase capabilities and prepare for VCU-specific logic.

- [x] T001 Verify API constants and enum capability in `aliyun/api/tablestore/alicloud_tablestore_instance_api.go` and `cws-lib-go` references.
- [x] T002 Verify `Read` logic specifically for `elastic_vcu_upper_limit` population in `alicloud/service_alicloud_ots_instance.go`.
- [x] T003 Check dependencies and ensure `CreateVCUInstance` and `UpdateInstanceElasticVCUUpperLimit` are available in API layer (`aliyun/api/tablestore/alicloud_tablestore_instance_api.go`).

## Phase 2: Foundational Components

**Goal**: Ensure core service layer functions are robust for VCU operations.

- [x] T004 [P] Review and Refine `resourceAliCloudOtsInstanceCreate` in `alicloud/resource_alicloud_ots_instance.go` to ensure correct flow for VCU instances (implied handled by API, but verify error propagation).
- [x] T005 [P] Verify `resourceAliCloudOtsInstanceUpdate` uses `UpdateInstanceElasticVCUUpperLimit` correctly when only limit changes.
- [x] T006 Ensure schema validation for `instance_specification` correctly includes "VCU" (it is present in `alicloud/resource_alicloud_ots_instance.go` but double check logic).

## Phase 3: User Story 1 - Create VCU Instance

**Goal**: Enable creation of VCU instances with optional elastic limits.
**Priority**: P1

- [x] T007 [US1] Create acceptance test `TestAccAlicloudOtsInstance_vcu` in `alicloud/resource_alicloud_ots_instance_test.go` (create new test file if needed).
- [x] T008 [US1] Run acceptance test `TestAccAlicloudOtsInstance_vcu` and verify creation succeeds.
- [x] T009 [US1] Verify `terraform plan` is clean (no diff) after creation (State Management check).

## Phase 4: User Story 2 - Update VCU Configuration

**Goal**: Allow updating elastic VCU limits without recreation.
**Priority**: P2

- [x] T010 [US2] Create acceptance test `TestAccAlicloudOtsInstance_updateVcuLimit` in `alicloud/resource_alicloud_ots_instance_test.go`.
- [x] T011 [US2] Implement logic to handle updates if missing (current analysis suggests it exists, so this is a verification task during test run).
- [x] T012 [US2] Run acceptance test `TestAccAlicloudOtsInstance_updateVcuLimit` and verify update occurs via API call.

## Phase 5: Polish & Cross-Cutting

**Goal**: Final cleanups and documentation.

- [x] T013 [P] Verify error messages are clear when `elastic_vcu_upper_limit` is set for non-VCU instances (Edge Case).
- [x] T014 Enforce strong typing and build verification by running `make` in `/cws_data/terraform-provider-alicloud`.

## Implementation Strategy

- **Sequential Execution**: First verify logical correctness of existing code (which seems to have VCU fragments), then create tests to prove it works.
- **Fail Fast**: Use acceptance tests to identify gaps in the "already implemented" logic.

## Dependencies

- US1 (Create) must pass before US2 (Update) verification is reliable.
