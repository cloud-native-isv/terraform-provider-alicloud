# Tasks: Implement alicloud_ecs_instance Resource

**Feature**: Implement alicloud_ecs_instance Resource (001-implement-ecs-instance)

## Phase 1: Setup
Initialization and project structure verification.

- [X] T001 Verify project build system and ensure clean state in `alicloud/` directory

## Phase 2: Foundational (Blocking)
Service layer implementation for ECS operations.

- [ ] T002 [P] Implement `CreateInstance` method using CWS-Lib-Go in `alicloud/service_alicloud_ecs.go`
- [ ] T003 [P] Implement `DescribeInstance` method with pagination logic in `alicloud/service_alicloud_ecs.go`
- [ ] T004 [P] Implement `DeleteInstance` method in `alicloud/service_alicloud_ecs.go`
- [ ] T005 [P] Implement `StartInstance` method in `alicloud/service_alicloud_ecs.go`
- [ ] T006 [P] Implement `StopInstance` method in `alicloud/service_alicloud_ecs.go`
- [ ] T007 [P] Implement `WaitForInstanceRunning` state refresh function in `alicloud/service_alicloud_ecs.go`
- [ ] T008 [P] Implement `WaitForNetworkReady` state refresh function in `alicloud/service_alicloud_ecs.go`

## Phase 3: User Story 1 - Create ECS Instance
**Goal**: Provision an ECS instance using CreateInstance API and ensure it reaches Running state.
**Test Criteria**: `terraform apply` creates instance, public/private IPs populated, status is Running.

- [ ] T009 [US1] Create resource skeleton with schema definition in `alicloud/resource_alicloud_ecs_instance.go`
- [ ] T010 [US1] Implement `create` function using Service layer and wait logic in `alicloud/resource_alicloud_ecs_instance.go`
- [ ] T011 [US1] Implement `read` function to sync state in `alicloud/resource_alicloud_ecs_instance.go`
- [ ] T012 [US1] Register new resource in `alicloud/provider.go`
- [ ] T013 [US1] Create acceptance test file `alicloud/resource_alicloud_ecs_instance_test.go`
- [ ] T014 [US1] Add `TestAccAlicloudEcsInstance_basic` test case for creation verification in `alicloud/resource_alicloud_ecs_instance_test.go`

## Phase 4: User Story 2 - Lifecycle Management
**Goal**: Support updates and deletion of the ECS instance.
**Test Criteria**: `terraform destroy` terminates instance; changing tags/description updates instance.

- [ ] T015 [US2] Implement `update` function for basic fields (name, tags, description) in `alicloud/resource_alicloud_ecs_instance.go`
- [ ] T016 [US2] Implement `delete` function with graceful termination in `alicloud/resource_alicloud_ecs_instance.go`
- [ ] T017 [US2] Add `TestAccAlicloudEcsInstance_update` test case in `alicloud/resource_alicloud_ecs_instance_test.go`

## Final Phase: Polish
Cross-cutting concerns and final validation.

- [ ] T018 Verify full compilation and run lint checks via `make`

## Dependencies

1. **Foundational** (T001-T008) must complete before **User Story 1**.
2. **User Story 1** (T009-T014) must complete before **User Story 2**.
3. **Drafting T009** (Schema) can start early but requires T002/T003 for implementation.

## Implementation Strategy
- **MVP**: Complete **User Story 1** first. This delivers the core value (CreateInstance support).
- **Incremental**: Add Lifecycle (US2) only after Create is stable.
