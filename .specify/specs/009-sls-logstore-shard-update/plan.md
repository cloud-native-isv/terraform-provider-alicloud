# Implementation Plan - Update LogStore Shard Logic

**Feature**: Update LogStore Shard Logic
**Goal**: Implement dynamic shard scaling (split/merge) for SLS LogStores while adhering to strict synchronous and error-handling requirements.

## Technical Context

Supporting dynamic shard updates requires careful orchestration of the underlying SLS APIs because the generic Update API does not support `shard_count` changes.

### Unknowns & Risks

- **Shard Iteration Reliability**: The `Sequential/Arbitrary` strategy assumes we can reliably list and target shards. We need to verify that `DescribeLogStore` or `GetLogStoreShards` returns stable IDs immediately after operations.
- **Timing**: How fast does `SplitLogStoreShard` reflect in `DescribeLogStore`?
- **Concurrent Operations**: Can we split multiple shards in rapid succession, or must we serialize strictly? (Assuming serialized per requirements).

### Technical Choices

- **API Layer**: Use established `SlsService` pattern in `alicloud/service_alicloud_sls.go` (or similar).
- **Polling**: Implement explicit polling loop for state synchronization.
- **Error Handling**: Fail-fast approach; any error in the chain terminates the update.

## Constitution Check

### I. Architecture Layering
- [x] Logic will be implemented in `alicloud/service_alicloud_sls.go` (Service Layer) and called by `alicloud/resource_alicloud_log_store.go`.
- [x] No direct SDK calls in Resource layer.

### II. State Management
- [x] Implement `WaitForLogStoreShardCount` in Service layer.
- [x] Resource Read/Create will synchronize state properly.

### III. Error Handling
- [x] Use `alicloud/errors.go` (IsNotFoundError, etc.)
- [x] Fail-fast on partial update failures.

### IV. Code Quality
- [x] Use `sls_log_store` naming conventions.
- [x] Split files if `service_alicloud_sls.go` exceeds 1000 lines.

### V. Strong Typing
- [x] Use `cws-lib-go` structs for shard operations if available.

## Design

### Phase 0: Research

- [ ] Verify `SplitLogStoreShard` and `MergeLogStoreShards` signatures in `cws-lib-go` or current SDK wrapper.
- [ ] Confirm exact behavior of Shard IDs (active vs historical).

### Phase 1: Service Layer Implementation

- [ ] Implement `SplitLogStoreShards(project, logstore, count)` helper in Service layer.
  - Logic: List shards -> filter active -> split iteratively.
- [ ] Implement `MergeLogStoreShards(project, logstore, count)` helper.
  - Logic: List shards -> filter active -> merge iteratively.
- [ ] Implement `WaitForLogStoreShardCount(project, logstore, targetCount)` poller.

### Phase 2: Resource Integration & Test

- [ ] Update `resourceAlicloudLogStoreUpdate` to detect `shard_count` change.
- [ ] Call Service layer split/merge functions if change detected.
- [ ] Add acceptance test `TestAccAlicloudLogStoreDataSource_shardUpdate`.

## Verification Plan

### Automated Tests
- Run `TestAccAlicloudLogStoreDataSource_shardUpdate`: verify 2 -> 4 -> 2 scaling.

### Manual Verification
- None required if acceptance tests pass.
