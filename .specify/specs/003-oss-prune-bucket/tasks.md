---
description: "Task list for OSS bucket prune-before-delete"
---

# Tasks: OSS Bucket Prune Before Delete

**Input**: Design documents from `.specify/specs/003-oss-prune-bucket/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), data-model.md, contracts/, quickstart.md

**Tests**: Not required by the spec; focus on implementation and validation via `make`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Review current OSS bucket delete flow in alicloud/resource_alicloud_oss_bucket.go
- [x] T002 Review OSS service helpers in alicloud/service_alicloud_oss_bucket.go and alicloud/service_alicloud_oss_object.go
- [x] T003 Confirm PruneBucket API availability in pkg/cws-lib-go/lib/cloud/aliyun/api/oss/

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T004 Rename EmptyBucket to PruneBucket in alicloud/service_alicloud_oss_bucket.go and update any references
- [x] T005 Update OSS service cleanup logic to call PruneBucket API via CWS-Lib-Go in alicloud/service_alicloud_oss_bucket.go

---

## Phase 3: User Story 1 - Force-destroy non-empty bucket (Priority: P1) 🎯 MVP

**Goal**: When `force_destroy` is enabled, cleanup runs before delete and deletion waits for completion/timeout.

**Independent Test**: Delete a non-empty bucket with `force_destroy = true` and verify cleanup then deletion succeeds.

### Implementation for User Story 1

- [x] T006 [US1] Update delete flow to call PruneBucket before bucket deletion in alicloud/resource_alicloud_oss_bucket.go
- [x] T007 [US1] Ensure delete flow waits for cleanup completion/timeout in alicloud/resource_alicloud_oss_bucket.go
- [x] T008 [US1] Add retryable vs non-retryable error handling around cleanup in alicloud/service_alicloud_oss_bucket.go
- [x] T009 [US1] Ensure cleanup is idempotent and safe for empty buckets in alicloud/service_alicloud_oss_bucket.go

---

## Phase 4: User Story 2 - Comprehensive cleanup coverage (Priority: P2)

**Goal**: Cleanup removes versions and pending uploads to guarantee complete deletion.

**Independent Test**: Create versions and multipart uploads, delete with `force_destroy = true`, and confirm bucket removal.

### Implementation for User Story 2

- [x] T010 [US2] Verify PruneBucket invocation covers versions and pending uploads in alicloud/service_alicloud_oss_bucket.go
- [x] T011 [US2] Ensure force_destroy cleanup retries when new objects appear during cleanup in alicloud/service_alicloud_oss_bucket.go

---

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T012 Run `make` to validate build and formatting (repo root)
- [x] T013 Update any related OSS deletion comments or inline documentation in alicloud/resource_alicloud_oss_bucket.go

---

## Dependencies & Execution Order

- **Setup (Phase 1)** → **Foundational (Phase 2)** → **User Story 1 (Phase 3)** → **User Story 2 (Phase 4)** → **Polish (Phase 5)**

## Parallel Opportunities

- T001–T003 can run in parallel (analysis-only tasks).
- T006–T009 can be sequenced but parts may be parallel if edits are in different files.

## Parallel Example: User Story 1

- T006 [US1] Update delete flow in alicloud/resource_alicloud_oss_bucket.go
- T008 [US1] Add retryable error handling in alicloud/service_alicloud_oss_bucket.go

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2
2. Complete Phase 3 (User Story 1)
3. Validate with `make`
4. Stop and verify deletion flow works end-to-end

### Incremental Delivery

1. Implement User Story 1 (MVP)
2. Add User Story 2 cleanup coverage
3. Re-validate with `make`
