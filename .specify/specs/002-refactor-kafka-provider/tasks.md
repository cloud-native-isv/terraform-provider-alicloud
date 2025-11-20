---
description: "Task list for Kafka provider refactoring implementation"
---

# Tasks: Kafka Provider Refactoring

**Input**: Design documents from `/.specify/specs/002-refactor-kafka-provider/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: This is a refactoring task where existing acceptance tests must continue to pass. No new test tasks are generated as the specification focuses on implementation refactoring while maintaining identical behavior.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `alicloud/` at repository root (Terraform provider structure)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the development environment and understand the current implementation

- [x] T001 [P] Review existing Kafka provider files to understand current implementation patterns
- [x] T002 [P] Set up cws-lib-go Kafka API client access pattern based on Flink example
- [x] T003 [P] Create backup of all Kafka-related files before refactoring

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core service layer implementation that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Implement KafkaService struct and NewKafkaService constructor in `alicloud/service_alicloud_alikafka.go`
- [x] T005 [P] Implement ID encoding/decoding functions for all Kafka resources in `alicloud/service_alicloud_alikafka.go`
- [x] T006 [P] Implement Kafka instance service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T007 [P] Implement Kafka topic service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T008 [P] Implement Kafka consumer group service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T009 [P] Implement Kafka SASL user service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T010 [P] Implement Kafka SASL ACL service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T011 [P] Implement Kafka allowed IP service methods (Describe/Create/Delete/WaitFor) in `alicloud/service_alicloud_alikafka.go`
- [x] T012 Implement proper error handling with standard predicates in service layer
- [x] T013 Verify and use pagination already encapsulated by cws-lib-go; do not add pagination loops in Provider Service layer (align with constitution: pagination in *_api.go)
- [x] T014 Validate service layer implementation compiles with `make` command

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Modernized Kafka Provider Implementation (Priority: P1) 🎯 MVP

**Goal**: Refactor all Kafka resources to use cws-lib-go API layer instead of direct SDK calls

**Independent Test**: All existing Kafka provider acceptance tests pass with 100% success rate

### Implementation for User Story 1

- [ ] T015 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_instance.go` to use KafkaService
- [ ] T016 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_topic.go` to use KafkaService
- [ ] T017 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_consumer_group.go` to use KafkaService
- [ ] T018 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_sasl_user.go` to use KafkaService
- [ ] T019 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_sasl_acl.go` to use KafkaService
- [ ] T020 [P] [US1] Refactor `alicloud/resource_alicloud_alikafka_instance_allowed_ip_attachment.go` to use KafkaService
- [ ] T021 [P] [US1] Refactor `alicloud/data_source_alicloud_alikafka_instances.go` to use KafkaService
- [ ] T022 [P] [US1] Refactor `alicloud/data_source_alicloud_alikafka_topics.go` to use KafkaService
- [ ] T023 [P] [US1] Refactor `alicloud/data_source_alicloud_alikafka_consumer_groups.go` to use KafkaService
- [ ] T024 [P] [US1] Refactor `alicloud/data_source_alicloud_alikafka_sasl_users.go` to use KafkaService
- [ ] T025 [P] [US1] Refactor `alicloud/data_source_alicloud_alikafka_sasl_acls.go` to use KafkaService
- [ ] T026 [US1] Update all resource schemas to use strong typing from cws-lib-go, and maintain a "Provider schema field ↔ cws-lib-go type field" mapping table in `data-model.md`
- [ ] T027 [US1] Remove all direct alikafka SDK calls and replace with KafkaService calls
- [ ] T028 [US1] Ensure all composite IDs use proper Encode/Decode functions
- [ ] T029 [US1] Verify all resources use proper timeout configurations

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Consistent Error Handling and State Management (Priority: P2)

**Goal**: Ensure consistent error handling and state management across all Kafka resources

**Independent Test**: Error messages are consistent with other provider resources and provide actionable information; state transitions are reliable

### Implementation for User Story 2

- [ ] T030 [P] [US2] Standardize error handling in all Kafka resources using IsNotFoundError/IsAlreadyExistError/NeedRetry predicates
- [ ] T031 [P] [US2] Ensure all Create operations use WaitFor* functions instead of direct Read calls
- [ ] T032 [P] [US2] Ensure all Delete operations properly wait for resource deletion completion
- [ ] T033 [P] [US2] Implement proper retry logic for transient errors in all operations
- [ ] T034 [P] [US2] Standardize error wrapping with WrapError/WrapErrorf patterns
- [ ] T035 [US2] Verify state refresh functions handle all edge cases correctly
- [ ] T036 [US2] Test error scenarios (quota limits, invalid configurations) and verify appropriate error messages

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Proper Layered Architecture Compliance (Priority: P3)

**Goal**: Ensure 100% compliance with layered architecture guidelines

**Independent Test**: Code review shows clear separation between Resource/DataSource, Service, and API layers

### Implementation for User Story 3

- [ ] T037 [P] [US3] Verify Resources/DataSources only call Service layer functions (no direct API/SDK calls)
- [ ] T038 [P] [US3] Verify Service layer only calls cws-lib-go API functions (no direct HTTP/SDK calls)
- [ ] T039 [P] [US3] Ensure no `map[string]interface{}` usage in new code (strong typing only)
- [ ] T040 [P] [US3] Verify all files are under 1000 lines (split if necessary)
- [ ] T041 [US3] Conduct architectural compliance review across all refactored files
- [ ] T042 [US3] Document the layered architecture implementation for future maintainers

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and quality assurance

- [ ] T043 [P] Run all existing Kafka acceptance tests and verify 100% pass rate
- [ ] T044 Run `make` command and fix any compilation or linting issues
- [ ] T045 [P] Perform manual testing of all Kafka resource types (create/read/update/delete)
- [ ] T046 [P] Verify resource creation times are within 10% of current performance baseline
- [ ] T047 Conduct final code review ensuring constitutional compliance:
  - [ ] T048 Layering: Resources/DataSources → Service → API
  - [ ] T049 State management: Proper WaitFor functions with timeouts
  - [ ] T050 Error handling: Standard predicates and wrapping
  - [ ] T051 Strong typing: cws-lib-go types only, no weak typing
  - [ ] T052 Pagination: Encapsulated in API layer (cws-lib-go *_api.go); Provider must not re-implement loops
  - [ ] T053 ID encoding: Proper Encode/Decode functions used consistently
  - [ ] T054 Build verification: `make` succeeds, files under 1000 LOC
- [ ] T055 Update documentation if needed (produce Chinese documentation per constitution)
- [ ] T056 Final validation using quickstart.md guidance
- [ ] T057 [P] Validate edge case: error code differences between old SDK and cws-lib-go; establish mapping and ensure consistent handling
- [ ] T058 [P] Validate rate limiting behavior and retry policy: exponential backoff with jitter as per plan; confirm effective under load
- [ ] T059 [P] Validate field/structure differences: regression check for schema-field mapping and state synchronization
- [ ] T060 [P] Verify strong typing compliance: ensure all new code uses cws-lib-go strong types exclusively
- [ ] T061 [P] Check file sizes and split any files exceeding 1000 lines if necessary

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All resource refactoring tasks within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all resource refactoring tasks for User Story 1 together:
Task: "Refactor alicloud/resource_alicloud_alikafka_instance.go to use KafkaService"
Task: "Refactor alicloud/resource_alicloud_alikafka_topic.go to use KafkaService"
Task: "Refactor alicloud/resource_alicloud_alikafka_consumer_group.go to use KafkaService"
Task: "Refactor alicloud/resource_alicloud_alikafka_sasl_user.go to use KafkaService"
Task: "Refactor alicloud/resource_alicloud_alikafka_sasl_acl.go to use KafkaService"
Task: "Refactor alicloud/resource_alicloud_alikafka_instance_allowed_ip_attachment.go to use KafkaService"
Task: "Refactor alicloud/data_source_alicloud_alikafka_instances.go to use KafkaService"
Task: "Refactor alicloud/data_source_alicloud_alikafka_topics.go to use KafkaService"
Task: "Refactor alicloud/data_source_alicloud_alikafka_consumer_groups.go to use KafkaService"
Task: "Refactor alicloud/data_source_alicloud_alikafka_sasl_users.go to use KafkaService"
Task: "Refactor alicloud/data_source_alicloud_alikafka_sasl_acls.go to use KafkaService"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Run all Kafka acceptance tests to verify identical behavior
5. Deploy/demonstrate if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Validate (MVP!)
3. Add User Story 2 → Test independently → Validate
4. Add User Story 3 → Test independently → Validate
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (resource refactoring)
   - Developer B: User Story 2 (error handling and state management)
   - Developer C: User Story 3 (architectural compliance)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- **Critical Success Factor**: All existing functionality MUST be preserved with identical behavior