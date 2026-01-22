# Tasks: Split OTS Instance Resources

**Feature Branch**: `011-split-ots-instance`
**Status**: Generated

## Phase 1: Setup

1. **Initialize Feature Workspace**
   - [ ] T001 Initialize branch and folder structure (already done by tools, verifying check) in `.specify/specs/011-split-ots-instance/`

## Phase 2: Foundational

1. **Design & Plan Review**
   - [ ] T002 Verify `resource_alicloud_ots_instance.go` content matches expected legacy state in `alicloud/resource_alicloud_ots_instance.go`
   - [ ] T003 Ensure `tablestoreAPI` imports are available for new resource in `alicloud/resource_alicloud_ots_instance.go`

## Phase 3: User Story 1 (Priority P1) - Create VCU Instance

1. **VCU Resource Implementation (New)**
   - [ ] T004 [US1] Create new resource file `alicloud/resource_alicloud_ots_instance_vcu.go` with strict VCU schema
   - [ ] T005 [US1] Implement `Create` function hardcoding "VCU" spec and supporting `elastic_vcu_upper_limit` in `alicloud/resource_alicloud_ots_instance_vcu.go`
   - [ ] T006 [US1] Implement `Read` function to map API response to VCU schema (only expose VCU fields) in `alicloud/resource_alicloud_ots_instance_vcu.go`
   - [ ] T007 [US1] Implement `Update` function for VCU-specific fields (e.g. `elastic_vcu_upper_limit`) in `alicloud/resource_alicloud_ots_instance_vcu.go`
   - [ ] T008 [US1] Implement `Delete` function reusing existing delete logic in `alicloud/resource_alicloud_ots_instance_vcu.go`

2. **Register Resource**
   - [ ] T009 [US1] Register `alicloud_ots_instance_vcu` in provider resource map in `alicloud/provider.go`

## Phase 4: User Story 2 (Priority P1) - Maintain Standard Instance

1. **Legacy Resource Cleanup**
   - [ ] T010 [US2] update schema in `alicloud/resource_alicloud_ots_instance.go`: remove `elastic_vcu_upper_limit` and `vcu_quota` optional/computed fields
   - [ ] T011 [US2] update validation in `alicloud/resource_alicloud_ots_instance.go`: remove "VCU" from `instance_specification` allowed values
   - [ ] T012 [US2] update `Update` function in `alicloud/resource_alicloud_ots_instance.go`: remove handling for `elastic_vcu_upper_limit`
   - [ ] T013 [US2] update `Read` function in `alicloud/resource_alicloud_ots_instance.go`: remove setting of VCU-only fields from API response

## Phase 5: Testing & Acceptance

1. **Acceptance Testing**
   - [ ] T014 [US1] Create acceptance test for VCU instance lifecycle (Create/Read/Update/Delete) in `alicloud/resource_alicloud_ots_instance_vcu_test.go`
   - [ ] T015 [US2] Create/Update acceptance test for Standard (SSD/HYBRID) instance to verify VCU parameters are rejected in `alicloud/resource_alicloud_ots_instance_test.go`
   - [ ] T015b [Test] Verify common fields (Title, Description, Tags) function correctly on both in `alicloud/resource_alicloud_ots_instance_test.go` and `alicloud/resource_alicloud_ots_instance_vcu_test.go`
   - [ ] T015c [Test] Negative testing: Ensure `alicloud_ots_instance` raises error if `instance_type`="VCU" is manually attempted (via `MakeTestSteps`) in `alicloud/resource_alicloud_ots_instance_test.go`
   - [ ] T016 [US1] Verify manual documentation or migration note added to `website/docs/r/ots_instance.html.markdown` (if applicable) or `docs/`

2. **Final Verification**
   - [ ] T017 Run all OTS acceptance tests to ensure no regression for standard instances (run `make testacc` targeting OTS)
   - [ ] T018 Verify `terraform plan` output for both resource types

## Dependencies

- **T004-T008** (VCU Resource) depend on **T002-T003** (Foundations)
- **T010-T013** (Legacy Cleanup) can run in parallel with VCU implementation
- **T014-T018** (Testing) depend on implementations T008 and T013

## Parallel Execution Opportunities

- T004-T008 (New Resource) and T010-T013 (Refactor Legacy) can be developed simultaneously.
- T014 and T015 (Tests) can be written alongside implementations.

## Implementation Strategy

1. **MVP**: Implement `alicloud_ots_instance_vcu` first to prove VCU capability works independently.
2. **Cleanup**: Refactor `alicloud_ots_instance` to remove VCU support (Hard Break).
3. **Verify**: Run acceptance tests.
