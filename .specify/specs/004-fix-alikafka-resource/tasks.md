# Tasks: Fix AliKafka Resource Implementation

**Feature**: 004-fix-alikafka-resource | **Date**: 2026-01-25  
**Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)  
**Input**: User stories from spec with 3 priority P1 requirements

## Implementation Strategy

**MVP First**: Complete User Story 1 (Complete State Persistence) as the minimum viable solution, ensuring all computed fields are properly synchronized to state. This provides immediate value by fixing core state management issues.

**Incremental Delivery**: 
1. Phase 2 foundational tasks (Service layer) complete all stories
2. User Story 1: State persistence fixes
3. User Story 2: API integration updates
4. User Story 3: State lifecycle management

## Phase 1: Setup (project initialization)

- [x] T001 [P] Set up development environment with Go 1.20+
- [x] T002 [P] Verify CWS-Lib-Go Kafka API dependencies in pkg/
- [x] T003 [P] Prepare testing environment for Terraform provider

## Phase 2: Foundational (blocking prerequisites)

- [x] T004 [P] Create service_alicloud_alikafka.go with KafkaService struct
- [x] T005 [P] Implement NewKafkaService factory function
- [x] T006 Implement KafkaService.DescribeInstance method
- [x] T007 Implement KafkaService.CreateInstance method
- [x] T008 Implement KafkaService.UpgradeInstance method
- [x] T009 Implement KafkaService.AliKafkaInstanceStateRefreshFunc for state polling
- [x] T010 Implement KafkaService.WaitForAliKafkaInstanceCreating for state synchronization
- [x] T011 Implement KafkaService.WaitForAliKafkaInstanceUpdating for state synchronization

## Phase 3: User Story 1 - Complete State Persistence (Priority: P1)

**Goal**: As a Terraform user managing AliKafka instances, I want all resource attributes to be properly synchronized and persisted in the Terraform state file after any operation.

**Independent Test**: Can be fully tested by creating an instance and verifying tfstate.

- [x] T012 [P] [US1] Update resource_alicloud_alikafka_instance.go schema to include all computed fields
- [x] T013 [P] [US1] Implement Read method to fetch all fields from CWS-Lib-Go API
- [x] T014 [US1] Update Read method to set vswitch_ids and selected_zones fields
- [x] T015 [P] [US1] Update Create method to call WaitForAliKafkaInstanceCreating
- [x] T016 [P] [US1] Update Update method to handle all field updates
- [x] T017 [US1] Verify all resource attributes are correctly synchronized in tfstate file

## Phase 4: User Story 2 - Modern API Integration (Priority: P1)

**Goal**: As a Terraform provider maintainer, I want the AliKafka resource implementation to use the latest CWS-Lib-Go API layer.

**Independent Test**: Verify code uses only CWS-Lib-Go.

- [x] T018 [P] [US2] Enhance CWS-Lib-Go KafkaInstance struct in pkg/
- [x] T019 [P] [US2] Enhance CWS-Lib-Go CreateInstance API in pkg/
- [x] T020 [P] [US2] Implement CWS-Lib-Go UpgradeInstance API in pkg/
- [x] T021 [P] [US2] Update resource Create to use local KafkaService
- [x] T022 [P] [US2] Update resource Update to use local KafkaService
- [x] T023 [P] [US2] Remove legacy SDK usage from instance resource

## Phase 5: User Story 3 - Proper State Lifecycle Management (Priority: P1)

**Goal**: As a Terraform user deploying AliKafka instances, I want the provider to properly handle the multi-stage lifecycle.

**Independent Test**: Verify wait times and state transitions.

- [x] T024 [P] [US3] Implement WaitForAliKafkaInstanceStopping in service layer
- [x] T025 [P] [US3] Update resource_alicloud_alikafka_deployment.go to use new service
- [x] T026 [P] [US3] Update deployment resource to handle proper state transitions
- [x] T027 [US3] Verify deployment resource waits for appropriate states

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T028 [P] Update resource_alicloud_alikafka_instance.go backward compatibility
- [x] T029 [P] Update resource_alicloud_alikafka_deployment.go backward compatibility
- [x] T030 [P] Update alicloud_alikafka_instance.md documentation
- [x] T031 [P] Update alicloud_alikafka_deployment.md documentation
- [x] T032 [P] Create examples/alicloud_alikafka/main.tf
- [x] T033 [P] Create examples/alicloud_alikafka/variables.tf
- [x] T034 [P] Create examples/alicloud_alikafka/outputs.tf
- [x] T035 Run make command to verify code compiles
- [x] T036 Test resource with real Alibaba Cloud infrastructure (optional)

## Dependencies

### User Story Completion Order
1. **User Story 1** (State Persistence)
2. **User Story 2** (API Integration)
3. **User Story 3** (State Lifecycle)

### Blocking Relations
- Phase 2 blocks all US phases
- US1 enabling for US2/US3

## Parallel Opportunities
- T012 can run parallel with T004-T008
- T018-T020 can run parallel
