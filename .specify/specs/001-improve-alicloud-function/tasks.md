# Tasks: Improve AliCloud Function Compute Support

**Input**: Design documents from `/.specify/specs/001-improve-alicloud-function/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: The examples below include test tasks. Tests are OPTIONAL - only include them if explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `alicloud/`, `tests/` at repository root
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project structure per implementation plan
- [x] T002 Initialize Go project with Terraform provider dependencies
- [x] T003 [P] Configure linting and formatting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Setup FC service base structure in alicloud/service_alicloud_fc_base.go
- [x] T005 [P] Implement FC API integration with cws-lib-go
- [x] T006 [P] Setup error handling and logging infrastructure for FC services
- [x] T007 Create base models/entities that all stories depend on
- [x] T008 Configure environment configuration management for FC

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enhanced Function Management (Priority: P1) 🎯 MVP

**Goal**: Complete and consistent support for all FC resources so users can reliably provision and manage serverless applications

**Independent Test**: Create, update, and delete various FC resources (functions, layers, triggers, etc.) and verify that all operations work correctly with proper state management

### Tests for User Story 1 (OPTIONAL - only if tests requested) ⚠️

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T009 [P] [US1] Unit test for FC function service in tests/unit/service_alicloud_fc_function_test.go
- [ ] T010 [P] [US1] Unit test for FC layer service in tests/unit/service_alicloud_fc_layer_test.go
- [ ] T011 [P] [US1] Unit test for FC trigger service in tests/unit/service_alicloud_fc_trigger_test.go
- [ ] T012 [P] [US1] Acceptance test for FC function resource in tests/acceptance/resource_alicloud_fc_function_test.go
- [ ] T013 [P] [US1] Acceptance test for FC layer version resource in tests/acceptance/resource_alicloud_fc_layer_version_test.go
- [ ] T014 [P] [US1] Acceptance test for FC trigger resource in tests/acceptance/resource_alicloud_fc_trigger_test.go

### Implementation for User Story 1

- [x] T015 [P] [US1] Enhance FC function service in alicloud/service_alicloud_fc_function.go
- [x] T016 [P] [US1] Enhance FC layer service in alicloud/service_alicloud_fc_layer.go
- [x] T017 [P] [US1] Complete FC trigger service in alicloud/service_alicloud_fc_trigger.go
- [x] T018 [P] [US1] Enhance FC function resource in alicloud/resource_alicloud_fc_function.go
- [x] T019 [P] [US1] Enhance FC layer version resource in alicloud/resource_alicloud_fc_layer_version.go
- [x] T020 [P] [US1] Enhance FC trigger resource in alicloud/resource_alicloud_fc_trigger.go
- [x] T021 [US1] Add validation and error handling for all FC resources
- [x] T022 [US1] Add logging for FC resource operations

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Consistent API Integration (Priority: P2)

**Goal**: All FC resources follow consistent patterns for API integration so the codebase is maintainable and new features can be added easily

**Independent Test**: Review the code structure of different FC service implementations and verify they follow the same patterns for error handling, state management, and API calls

### Tests for User Story 2 (OPTIONAL - only if tests requested) ⚠️

- [ ] T023 [P] [US2] Unit test for FC custom domain service in tests/unit/service_alicloud_fc_custom_domain_test.go
- [ ] T024 [P] [US2] Unit test for FC alias service in tests/unit/service_alicloud_fc_alias_test.go
- [ ] T025 [P] [US2] Acceptance test for FC custom domain resource in tests/acceptance/resource_alicloud_fc_custom_domain_test.go
- [ ] T026 [P] [US2] Acceptance test for FC alias resource in tests/acceptance/resource_alicloud_fc_alias_test.go

### Implementation for User Story 2

- [x] T027 [P] [US2] Complete FC custom domain service in alicloud/service_alicloud_fc_custom_domain.go
- [x] T028 [P] [US2] Complete FC alias service in alicloud/service_alicloud_fc_alias.go
- [x] T029 [P] [US2] Enhance FC custom domain resource in alicloud/resource_alicloud_fc_custom_domain.go
- [x] T030 [P] [US2] Enhance FC alias resource in alicloud/resource_alicloud_fc_alias.go
- [x] T031 [US2] Ensure consistent API integration patterns across all FC services
- [x] T032 [US2] Verify Encode/Decode functions follow consistent patterns
- [x] T033 [US2] Verify Describe methods follow consistent patterns
- [x] T034 [US2] Verify StateRefreshFunc implementations follow consistent patterns

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Improved Error Handling (Priority: P3)

**Goal**: Clear and actionable error messages when FC operations fail so users can quickly diagnose and fix configuration issues

**Independent Test**: Intentionally trigger various error conditions and verify that the error messages are clear and helpful

### Tests for User Story 3 (OPTIONAL - only if tests requested) ⚠️

- [ ] T035 [P] [US3] Unit test for FC async invoke config service in tests/unit/service_alicloud_fc_async_invoke_config_test.go
- [ ] T036 [P] [US3] Unit test for FC concurrency config service in tests/unit/service_alicloud_fc_concurrency_config_test.go
- [ ] T037 [P] [US3] Acceptance test for FC async invoke config resource in tests/acceptance/resource_alicloud_fc_async_invoke_config_test.go
- [ ] T038 [P] [US3] Acceptance test for FC concurrency config resource in tests/acceptance/resource_alicloud_fc_concurrency_config_test.go

### Implementation for User Story 3

- [x] T039 [P] [US3] Complete FC async invoke config service in alicloud/service_alicloud_fc_base.go
- [x] T040 [P] [US3] Complete FC concurrency config service in alicloud/service_alicloud_fc_concurrency_config.go
- [x] T041 [P] [US3] Complete FC provision config service in alicloud/service_alicloud_fc_provision_config.go
- [x] T042 [P] [US3] Complete FC VPC binding service in alicloud/service_alicloud_fc_vpc_binding.go
- [x] T043 [P] [US3] Enhance FC async invoke config resource in alicloud/resource_alicloud_fc_async_invoke_config.go
- [x] T044 [P] [US3] Enhance FC concurrency config resource in alicloud/resource_alicloud_fc_concurrency_config.go
- [x] T045 [P] [US3] Enhance FC provision config resource in alicloud/resource_alicloud_fc_provision_config.go
- [x] T046 [P] [US3] Enhance FC VPC binding resource in alicloud/resource_alicloud_fc_vpc_binding.go
- [x] T047 [US3] Implement improved error handling with actionable messages
- [x] T048 [US3] Add appropriate retry logic for transient FC API errors

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Data Source Implementation

**Goal**: Complete implementation of FC data sources for all resources

**Independent Test**: Use data sources to retrieve FC resource information and verify correct data is returned

### Tests for Data Sources (OPTIONAL - only if tests requested) ⚠️

- [ ] T049 [P] [US1] Acceptance test for FC functions data source in tests/acceptance/data_source_alicloud_fc_functions_test.go
- [ ] T050 [P] [US2] Acceptance test for FC triggers data source in tests/acceptance/data_source_alicloud_fc_triggers_test.go
- [ ] T051 [P] [US2] Acceptance test for FC custom domains data source in tests/acceptance/data_source_alicloud_fc_custom_domains_test.go
- [ ] T052 [P] [US1] Acceptance test for FC zones data source in tests/acceptance/data_source_alicloud_fc_zones_test.go

### Implementation for Data Sources

- [ ] T053 [P] [US1] Enhance FC functions data source in alicloud/data_source_alicloud_fc_functions.go
- [ ] T054 [P] [US2] Enhance FC triggers data source in alicloud/data_source_alicloud_fc_triggers.go
- [ ] T055 [P] [US2] Enhance FC custom domains data source in alicloud/data_source_alicloud_fc_custom_domains.go
- [ ] T056 [P] [US1] Enhance FC zones data source in alicloud/data_source_alicloud_fc_zones.go

**Checkpoint**: All data sources should be functional and testable

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T057 [P] Documentation updates for all FC resources
- [ ] T058 Code cleanup and refactoring across all FC services
- [ ] T059 Performance optimization across all FC services
- [ ] T060 [P] Additional unit tests for all FC services
- [ ] T061 Security hardening for FC API calls
- [ ] T062 Run quickstart.md validation to ensure examples work correctly
- [ ] T063 Validate all code changes with 'make' command

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

- Tests (if included) MUST be written and FAIL before implementation
- Services before resources
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Services and resources within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Unit test for FC function service in tests/unit/service_alicloud_fc_function_test.go"
Task: "Unit test for FC layer service in tests/unit/service_alicloud_fc_layer_test.go"
Task: "Unit test for FC trigger service in tests/unit/service_alicloud_fc_trigger_test.go"
Task: "Acceptance test for FC function resource in tests/acceptance/resource_alicloud_fc_function_test.go"
Task: "Acceptance test for FC layer version resource in tests/acceptance/resource_alicloud_fc_layer_version_test.go"
Task: "Acceptance test for FC trigger resource in tests/acceptance/resource_alicloud_fc_trigger_test.go"

# Launch all services for User Story 1 together:
Task: "Enhance FC function service in alicloud/service_alicloud_fc_function.go"
Task: "Enhance FC layer service in alicloud/service_alicloud_fc_layer.go"
Task: "Complete FC trigger service in alicloud/service_alicloud_fc_trigger.go"

# Launch all resources for User Story 1 together:
Task: "Enhance FC function resource in alicloud/resource_alicloud_fc_function.go"
Task: "Enhance FC layer version resource in alicloud/resource_alicloud_fc_layer_version.go"
Task: "Enhance FC trigger resource in alicloud/resource_alicloud_fc_trigger.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence