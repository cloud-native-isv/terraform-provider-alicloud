---
description: "Task list for SLS LogStore Shard Management feature"
---

# Tasks: Update LogStore Shard Logic

**Input**: Design documents from `.specify/specs/009-sls-logstore-shard-update/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md
**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Verify project structure and dependencies
- [ ] T002 [P] Configure environment for testing

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented
**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Update `alicloud/service_alicloud_sls.go` with API imports (cws-lib-go)
- [ ] T004 Define `SplitLogStoreShard` and `MergeLogStoreShards` function signatures in `alicloud/service_alicloud_sls.go` (if not present in SDK wrapper)
- [ ] T005 Implement `WaitForLogStoreShardCount` poller in `alicloud/service_alicloud_sls.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Update Shard Count Upwards (Split) (Priority: P1) 🎯 MVP

**Goal**: Enable users to increase shard count via Terraform `shard_count` variable.

**Independent Test**: Create LogStore with `shard_count=2`, apply, update to `4`, apply, verify state.

### Implementation for User Story 1

#### Service Layer (Split Logic)
- [ ] T006 [US1] Implement `FilterActiveShards` helper in `alicloud/service_alicloud_sls.go`
- [ ] T007 [US1] Implement `SplitLogStoreShards` iteration logic in `alicloud/service_alicloud_sls.go`
      - Lists shards, filters active, splits iteratively until target count reached.
      - Uses synchronous waiting between steps or at end (per spec).
      - Error handling: Fail fast.

#### Resource Integration
- [ ] T008 [US1] Update `resourceAlicloudLogStoreUpdate` in `alicloud/resource_alicloud_log_store.go`
      - Detect `shard_count` change.
      - Call `SplitLogStoreShards` if target > current.
      - Ensure standard Update API is NOT called for shard count.
- [ ] T013 [US1] Update `resourceAlicloudLogStoreRead` in `alicloud/resource_alicloud_log_store.go`
      - Ensure `shard_count` state is set by counting ALL shards (Active + ReadOnly) to match Update logic.

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Update Shard Count Downwards (Merge) (Priority: P1)

**Goal**: Enable users to decrease shard count via Terraform `shard_count` variable.

**Independent Test**: Create LogStore with `shard_count=4`, apply, update to `2`, apply, verify state.

### Implementation for User Story 2

#### Service Layer (Merge Logic)
- [ ] T009 [US2] Implement `MergeLogStoreShards` iteration logic in `alicloud/service_alicloud_sls.go`
      - Lists shards, filters active, merges iteratively until target count reached.
      - Uses synchronous waiting.
      - Error handling: Fail fast.

#### Resource Integration
- [ ] T010 [US2] Update `resourceAlicloudLogStoreUpdate` in `alicloud/resource_alicloud_log_store.go`
      - Call `MergeLogStoreShards` if target < current.

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: Testing & Verification (Cross-Cutting)

**Purpose**: Validation and hardening

- [ ] T011 [P] Create acceptance test `TestAccAlicloudLogStoreDataSource_shardUpdate` in `alicloud/resource_alicloud_log_store_test.go`
      - Test case: 2 -> 4 (Split)
      - Test case: 4 -> 2 (Merge)
- [ ] T012 Run acceptance tests and verify pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup
- **User Story 1 (Split)**: Depends on Foundational
- **User Story 2 (Merge)**: Depends on Foundational (Split/Merge logic share helpers but can be implemented in any order relative to each other, P1 priority for both)

### Parallel Opportunities

- T006 (Filter helper) can be done in parallel with T003/T004 foundations.
- T011 (Test writing) can be done in parallel with implementation (TDD style).

## Implementation Strategy

### Incremental Delivery

1. Implement Foundation (Service helpers).
2. Implement Split logic & Resource integration. Verify US1.
3. Implement Merge logic & Resource integration. Verify US2.
4. Finalize Acceptance Tests.
