# Tasks: Split AliKafka Instance Resource

**Feature Branch**: `006-split-alikafka-instance`
**Status**: Planned

## Phase 1: Setup
*Goal: Initialize project structure and prepare for implementation.*

- [ ] T001 Create `alicloud/resource_alicloud_alikafka_deployment.go` with basic resource structure
- [ ] T002 Update `alicloud/service_alicloud_alikafka_types.go` to include `StopInstanceRequest` struct
- [ ] T003 Update `alicloud/service_alicloud_alikafka.go` to include `StopInstance` method stub

## Phase 2: Foundational
*Goal: Implement core service methods and types required for the split.*

- [ ] T004 Implement `StopInstance` logic in `alicloud/service_alicloud_alikafka.go` calling the appropriate API
- [ ] T005 Verify `StartInstance` in `alicloud/service_alicloud_alikafka.go` supports all necessary parameters for deployment

## Phase 3: User Story 1 - Create Instance Only
*Goal: Modify `alicloud_alikafka_instance` to only create the instance without starting it.*

**Independent Test**: Create an `alicloud_alikafka_instance` resource. Verify via API or console that the instance exists but is in a non-running state.

- [ ] T006 [US1] Modify `resourceAliCloudAlikafkaInstanceCreate` in `alicloud/resource_alicloud_alikafka_instance.go` to remove `StartInstance` call
- [ ] T007 [US1] Modify `resourceAliCloudAlikafkaInstanceCreate` in `alicloud/resource_alicloud_alikafka_instance.go` to remove waiting for "Running" state
- [ ] T008 [US1] Modify `resourceAliCloudAlikafkaInstanceRead` in `alicloud/resource_alicloud_alikafka_instance.go` to not enforce "Stopped" state (FR-009)
- [ ] T009 [US1] Update schema of `alicloud_alikafka_instance` in `alicloud/resource_alicloud_alikafka_instance.go` to remove deployment-specific fields (e.g., `vswitch_id` if strictly moved, or mark as optional/computed if keeping for compat - spec says strict split implies moving them, but let's check data model. Data model says `vswitch_id` moved to deployment. So remove `vswitch_id` from instance schema or make it optional/computed for reference? Spec says "Strict Split". Let's remove `vswitch_id` from `Required` in instance, maybe keep as `Computed` if API returns it, or remove entirely if it's input-only for Start.)
- [ ] T010 [US1] Create acceptance test for `alicloud_alikafka_instance` verifying creation without start

## Phase 4: User Story 2 - Deploy Instance
*Goal: Implement `alicloud_alikafka_deployment` to start the instance.*

**Independent Test**: Create an `alicloud_alikafka_instance` and an `alicloud_alikafka_deployment` referencing it. Verify the instance transitions to "Running".

- [ ] T011 [US2] Implement `resourceAliCloudAlikafkaDeploymentCreate` in `alicloud/resource_alicloud_alikafka_deployment.go` calling `StartInstance`
- [ ] T012 [US2] Implement `resourceAliCloudAlikafkaDeploymentRead` in `alicloud/resource_alicloud_alikafka_deployment.go` to check "Running" state and detect drift (FR-010)
- [ ] T013 [US2] Define schema for `alicloud_alikafka_deployment` in `alicloud/resource_alicloud_alikafka_deployment.go` including `instance_id`, `vswitch_id`, etc.
- [ ] T014 [US2] Create acceptance test for `alicloud_alikafka_deployment` verifying instance start

## Phase 5: User Story 3 - Stop Instance
*Goal: Implement `alicloud_alikafka_deployment` deletion to stop the instance.*

**Independent Test**: Destroy the `alicloud_alikafka_deployment` resource. Verify the instance transitions to "Stopped" state.

- [ ] T015 [US3] Implement `resourceAliCloudAlikafkaDeploymentDelete` in `alicloud/resource_alicloud_alikafka_deployment.go` calling `StopInstance`
- [ ] T016 [US3] Implement waiting logic in `Delete` to ensure instance reaches "Stopped" state
- [ ] T017 [US3] Create acceptance test for `alicloud_alikafka_deployment` destruction verifying instance stop

## Phase 6: Polish & Cross-Cutting
*Goal: Finalize documentation, cleanup, and full integration testing.*

- [ ] T018 Register `alicloud_alikafka_deployment` in `alicloud/provider.go` (or wherever resources are registered)
- [ ] T019 Update documentation for `alicloud_alikafka_instance` reflecting the breaking change
- [ ] T020 Create documentation for `alicloud_alikafka_deployment`
- [ ] T021 Run full acceptance test suite for AliKafka

## Dependencies

1. **Phase 1 & 2** must be completed first.
2. **Phase 3 (US1)** and **Phase 4 (US2)** can be developed somewhat in parallel, but US2 depends on US1's output (instance ID) for testing.
3. **Phase 5 (US3)** depends on Phase 4 (US2) implementation.

## Implementation Strategy

1. **MVP**: Implement Phase 1, 2, and 3. This allows creating instances (even if they can't be started via TF yet, they can be started manually).
2. **Full Feature**: Implement Phase 4 and 5 to complete the lifecycle management.
