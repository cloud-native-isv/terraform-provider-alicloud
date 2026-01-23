# Tasks: Fix OSS Bucket Force Destroy

Implementation of proactive bucket emptying logic for `force_destroy` in `alicloud_oss_bucket` resource.

## Phase 1: Setup

- [x] T001 Verify project structure and `plan.md` alignment

## Phase 2: User Story 1 - Force Destroy Non-Empty Bucket

**Goal**: Ensure `terraform destroy` works reliably for non-empty buckets when `force_destroy = true`.

**Impacted Files**:
- `alicloud/resource_alicloud_oss_bucket.go`
- `alicloud/service_alicloud_oss_bucket.go`

**Tests**:
- Acceptance test: `TestAccAlicloudOssBucketDataSource_basic` (or similar existing test)
- New test case: Bucket with objects/versions + `force_destroy=true`

- [x] T002 [US1] Create reproduction acceptance test case (bucket with objects + force_destroy=true) in `alicloud/resource_alicloud_oss_bucket_test.go`
- [x] T003 [US1] Refactor `resourceAliCloudOssBucketDelete` in `alicloud/resource_alicloud_oss_bucket.go` to proactively call `EmptyBucket` when `force_destroy` is true
- [x] T004 [US1] Implement Fail Fast logic: Stop destruction if `EmptyBucket` fails
- [x] T005 [US1] Clean up legacy reactive logic (EmptyBucket call inside Retry loop) in `alicloud/resource_alicloud_oss_bucket.go` to avoid redundancy
- [x] T006 [US1] Verify `EmptyBucket` in `alicloud/service_alicloud_oss_bucket.go` handles pagination and versioning correctly (Review only, code exists)

## Phase 3: Polish & Verification

- [x] T007 Run acceptance tests to verify fix
- [x] T008 Perform code review and ensure compliance with coding standards (checking error handling and logging)

## Dependencies

- All US1 tasks depend on T001
- T003 and T004/T005 are closely related and should be implemented together
- T007 depends on implementation completion

## Implementation Strategy

1. **Test First**: Write the acceptance test (T002) to confirm current flaky/reactive behavior (optional but good practice) or to ensure the new behavior works.
2. **Refactor Resource**: Modify `resourceAliCloudOssBucketDelete` to insert `EmptyBucket` call before the deletion retry loop.
3. **Clean up**: Remove the reactive logic within the retry loop.
4. **Verify**: Run tests.
