# Feature Specification: OSS Bucket Prune Before Delete

**Feature Branch**: `003-oss-prune-bucket`  
**Created**: 2026-01-23  
**Status**: Draft  
**Input**: User description: "将ossService.EmptyBucket重命名为ossService.PruneBucket，然后调用ossapi新添加的PruneBucket这个API来完成bucket的清理。当resourceAliCloudOssBucket实现的resource中force_destroy为true的时候，需要先调用ossService.PruneBucket将bucket清空，然后再进行bucket的删除。"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Force-destroy non-empty bucket (Priority: P1)

As an infrastructure operator, I want a non-empty bucket to be cleaned automatically when I enable force destroy, so that deletion succeeds without manual cleanup.

**Why this priority**: This is the core deletion path that currently blocks users and causes operational friction.

**Independent Test**: Can be fully tested by deleting a non-empty bucket with force destroy enabled and verifying that the bucket no longer exists.

**Acceptance Scenarios**:

1. **Given** a bucket that contains stored items and force destroy is enabled, **When** the user deletes the bucket, **Then** the system cleans the bucket and deletes it successfully.
2. **Given** an empty bucket and force destroy is enabled, **When** the user deletes the bucket, **Then** the deletion succeeds without extra steps.
3. **Given** a non-empty bucket and force destroy is disabled, **When** the user deletes the bucket, **Then** the deletion fails and the bucket remains intact.

---

### User Story 2 - Comprehensive cleanup coverage (Priority: P2)

As an infrastructure operator, I want force destroy to remove all stored content associated with a bucket (including versions and pending uploads), so that deletion is complete and predictable.

**Why this priority**: Incomplete cleanup leads to failed deletes and confusing residual data.

**Independent Test**: Can be fully tested by creating objects, versions, and pending uploads, enabling force destroy, and confirming the bucket is removed.

**Acceptance Scenarios**:

1. **Given** a bucket with versions and pending uploads, **When** force destroy is enabled and deletion is triggered, **Then** all stored content is removed and the bucket is deleted.

---

### Edge Cases

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right edge cases.
-->

- What happens when the cleanup succeeds but deletion is blocked by a permissions error?
- How does the system handle very large buckets that take a long time to clean?
- What happens if new objects are added while cleanup is in progress? Cleanup retries until the bucket is empty before deletion proceeds.
- How does the system behave when cleanup encounters transient errors (e.g., throttling) versus permanent errors (e.g., permission denied)?
- How does deletion behave when cleanup runs longer than expected? Wait until cleanup completes or times out.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: System MUST perform a bucket cleanup step before deletion when force destroy is enabled.
- **FR-002**: System MUST remove all stored content associated with the bucket, including versions and pending uploads, during cleanup.
- **FR-003**: System MUST NOT perform cleanup when force destroy is disabled.
- **FR-004**: Cleanup MUST be idempotent; re-running it on an empty bucket MUST succeed without error.
- **FR-005**: If cleanup fails, the deletion MUST stop and return a clear, actionable error.
- **FR-006**: Deletion MUST proceed only after cleanup completes successfully.
- **FR-007**: If new objects appear during cleanup, the system MUST retry cleanup until the bucket is empty before deletion proceeds.
- **FR-008**: Cleanup failures MUST distinguish retryable vs non-retryable errors; retryable errors MUST trigger automatic retry.
- **FR-009**: Deletion MUST wait for cleanup to finish or reach timeout before returning.

### Assumptions

- Users enable force destroy intentionally and understand it will remove all stored content.
- Standard account permissions allow listing and removing stored content for buckets managed by Terraform.
- Buckets may contain versions and pending uploads that must be included in cleanup scope.

### Key Entities *(include if feature involves data)*

- **Bucket**: A named storage container subject to creation, cleanup, and deletion.
- **Stored Item**: Any object or version stored in a bucket.
- **Pending Upload**: An in-progress upload that must be cleared before deletion.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: Users can delete a non-empty bucket with force destroy enabled in under 10 minutes for buckets up to 100,000 stored items.
- **SC-002**: At least 95% of force-destroy deletions of non-empty buckets complete without manual cleanup.
- **SC-003**: 90% of users successfully complete bucket deletion on the first attempt when force destroy is enabled.
- **SC-004**: Support tickets related to failed bucket deletion due to leftover content drop by 50% within one release cycle.

## Clarifications

### Session 2026-01-23

- Q: 当 force_destroy 触发清理期间又新增对象时，系统应如何处理？ → A: 重试清理直到桶为空后继续删除。
- Q: 当清理失败时，是否需要区分“可重试错误”与“不可重试错误”来决定是否自动重试？ → A: 区分可重试与不可重试错误，可重试则自动重试。
- Q: 当清理耗时较长时，删除应如何等待？ → A: 等待直到清理完成或超时。
