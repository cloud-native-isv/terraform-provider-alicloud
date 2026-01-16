# Tasks: Use CreateInstance API for Instance Resource

**Feature**: Use CreateInstance API for Instance Resource
**Status**: Planned

## Phase 1: Setup
- [ ] T001 Verify project builds and tests pass locally before starting changes to `alicloud/resource_alicloud_instance.go`

## Phase 2: Foundation (Service Layer)
**Goal**: Enable Service layer to interact with CreateInstance and StartInstance APIs.

- [ ] T002 [US1] Add `CreateInstance` method to `EcsService` in `alicloud/service_alicloud_ecs.go` using SDK request/response types
- [ ] T003 [US1] Add `StartInstance` method to `EcsService` in `alicloud/service_alicloud_ecs.go`
- [ ] T004 [US1] Verify `WaitForEcsInstance` in `alicloud/service_alicloud_ecs.go` supports waiting for "Stopped" and "Running" states (Existing code likely covers this, but verify/enhance if needed)

## Phase 3: User Story 1 - Create ECS Instance
**Goal**: Refactor resource creation to use new API flow.
**Tests**: `TestAccAlicloudInstance_basic`

- [ ] T005 [US1] Remove or Comment out legacy `RunInstances` logic in `resourceAliCloudInstanceCreate` in `alicloud/resource_alicloud_instance.go`
- [ ] T006 [US1] Initialize `ecs.CreateInstanceRequest` and map simple fields (ImageId, InstanceType, Name, Description, HostName, Password) in `alicloud/resource_alicloud_instance.go`
- [ ] T007 [US1] Map Network/VPC fields (VSwitchId, PrivateIpAddress) and Disks (SystemDisk, DataDisks) to request in `alicloud/resource_alicloud_instance.go`
- [ ] T008 [US1] Map ChargeType, Period, AutoRenew fields to request in `alicloud/resource_alicloud_instance.go`
- [ ] T009 [US1] Implement Security Group logic: Map first Group ID to request, store others for later in `alicloud/resource_alicloud_instance.go`
- [ ] T010 [US1] Implement `CreateInstance` API call and Error Handling in `alicloud/resource_alicloud_instance.go`
- [ ] T011 [US1] Implement Wait logic: Wait for instance to reach `Stopped` status after creation in `alicloud/resource_alicloud_instance.go`
- [ ] T012 [US1] Implement Post-Create Security Group Association: Loop through remaining groups and call `JoinSecurityGroup` in `alicloud/resource_alicloud_instance.go`
- [ ] T013 [US1] Implement Start Logic: Call `StartInstance` and Wait for `Running` status in `alicloud/resource_alicloud_instance.go`
- [ ] T014 [US1] Finalize State: Set Resource ID and call Read in `alicloud/resource_alicloud_instance.go`

## Phase 4: Polish & Verify
- [ ] T015 Run acceptance test `TestAccAlicloudInstance_basic` to verify basic creation flow
- [ ] T016 Run acceptance test `TestAccAlicloudInstance_vpc` (or similar) to verify VPC/VSwitch creation
- [ ] T017 Verify "Stopped" status behavior (ensure it started) manually or via test modification if needed
- [ ] T018 Verify Multiple Security Groups via test or manual verification

## Dependencies
- US1 (T005-T014) depends on Foundation (T002-T004)

## Implementation Strategy
1.  **Service Layer**: Add missing API wrappers.
2.  **Resource Refactor**: Replace the `RunInstances` block with the new multi-step flow.
3.  **Verification**: Ensure the instance ends up Running and fully configured.
