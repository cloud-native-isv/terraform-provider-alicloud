# Feature Specification: Update LogStore Shard Logic

**Feature Branch**: `009-sls-logstore-shard-update`
**Created**: 2026-01-14
**Status**: Draft
**Input**: User description: "当shard_count发生变动的时候底层API不支持直接针对logstore的shard进行update，需要使用(s *SlsAPI) SplitLogStoreShard和MergeLogStoreShards来单独进行修改"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Update Shard Count Upwards (Split) (Priority: P1)

As a user, I want to increase the `shard_count` of my LogStore so that I can handle higher data ingestion throughput. The provider should automatically handle this using the split operation.

**Why this priority**: Essential for scaling operations.

**Independent Test**:
1. Creates a LogStore with `shard_count = 2`.
2. Apply successfully.
3. Change Terraform config `shard_count = 4`.
4. Apply.
5. Verify that the LogStore now has 4 shards and no errors were reported.

**Acceptance Scenarios**:

1. **Given** a LogStore with 2 shards, **When** I update configuration to 4 shards, **Then** the provider calls `SplitLogStoreShard` and the state is updated to 4.

---

### User Story 2 - Update Shard Count Downwards (Merge) (Priority: P1)

As a user, I want to decrease the `shard_count` of my LogStore to reduce costs when high throughput is no longer needed. The provider should automatically handle this using the merge operation.

**Why this priority**: Essential for cost management.

**Independent Test**:
1. Creates a LogStore with `shard_count = 4`.
2. Apply successfully.
3. Change Terraform config `shard_count = 2`.
4. Apply.
5. Verify that the LogStore now has 2 shards.

**Acceptance Scenarios**:

1. **Given** a LogStore with 4 shards, **When** I update configuration to 2 shards, **Then** the provider calls `MergeLogStoreShards` and the state is updated to 2.

### Edge Cases

- **No Change**: If `shard_count` in config matches state, no API call should be made.
- **Unsupported operations**: If the underlying API fails (e.g., trying to merge below 1, or split beyond limit), the provider should return a clear error.
- **Partial Failure**: If a multi-step update fails (e.g., 2 -> 5 shards fails at step 2), the process MUST stop immediately and return the error. No rollback attempts. The user is expected to run `apply` again to retry or repair.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST detect changes to the `shard_count` attribute in `alicloud_log_store` resource.
    - **Refinement**: The system MUST count ALL shards (Active + ReadOnly) when determining current state against `shard_count`.
- **FR-002**: When `shard_count` increases, the system MUST use `(s *SlsAPI) SplitLogStoreShard` to increase the number of shards.
    - **Refinement**: The system MUST iterate through available `ReadWrite` shards and split them sequentially until the target count is reached.
- **FR-003**: When `shard_count` decreases, the system MUST use `(s *SlsAPI) MergeLogStoreShards` to decrease the number of shards.
    - **Refinement**: The system MUST iterate through available `ReadWrite` shards and merge them sequentially until the target count is reached.
- **FR-004**: The system MUST NOT attempt to update `shard_count` via the generic `UpdateLogStore` API as it is not supported.
- **FR-005**: The system MUST block until the shard operation is effectively reflected or return success if the operation is asynchronous but accepted.
    - **Refinement**: The system MUST poll the `DescribeLogStore` API (or equivalent) until the active shard count matches the target configuration.
    - **Timeout**: Use a default reasonable timeout (e.g., matching standard Create timeouts) or a specific configurable timeout if standard is insufficient.


### Key Entities

- **LogStore**: The SLS resource.
- **Shard**: A partition of the LogStore.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can update `shard_count` from N to M (where M > N) successfully via Terraform.
- **SC-002**: Users can update `shard_count` from N to M (where M < N) successfully via Terraform.
- **SC-003**: `terraform plan` following an apply shows no drift (clean state).

## Clarifications

### Session 2026-01-14

- Q: Shard Selection Strategy for Scaling → A: Sequential/Arbitrary (Iterate through available ReadWrite shards and split (midpoint) or merge them sequentially until the target count is reached).
- Q: State Synchronization Behavior (Wait vs Async) → A: Block until active shard count matches target (Synchronous).
- Q: Partial Failure Handling (Multi-step operations) → A: Stop and Error (Fail fast, leave state partially updated).
- Q: Drift Detection Strategy (ReadOnly Shards) → A: Strict Count (Count ALL shards including ReadOnly/Historical).

