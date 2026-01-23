# Feature Specification: Fix OSS Bucket Force Destroy

**Feature Branch**: `002-fix-oss-force-destroy`
**Created**: 2026-01-22
**Status**: Draft
**Input**: User description: "修改resource_alicloud_oss_bucket.go中的实现逻辑，当前force_destroy为true的时候直接删除bucket中的文件，避免出现如下错误 [ERROR]terraform-provider-alicloud/alicloud/resource_alicloud_oss_bucket.go:374: Resource kangaroo-xuanji-cn-hangzhou-data-7d75470c DeleteBucket Failed!!! [SDK aliyun-oss-go-sdk ERROR]: oss: service returned error: StatusCode=409, ErrorCode=BucketNotEmpty, ErrorMessage='The bucket has objects. Please delete them first.'"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Force Destroy Non-Empty Bucket (Priority: P1)

As a Terraform user, I want to be able to destroy an OSS bucket resource even if it contains objects or versions by setting `force_destroy = true`, so that I can easily clean up environments without manual intervention.

**Why this priority**: It addresses a direct error that blocks workflow automation and requires manual intervention to fix.

**Independent Test**: Create a bucket, add objects (and versions if enabled), run `terraform destroy`.
**Acceptance Scenarios**:

1. **Given** a bucket with `force_destroy = true` containing objects, **When** `terraform destroy` is run, **Then** all objects are deleted and the bucket is deleted successfully.
2. **Given** a bucket with `force_destroy = true` containing object versions and delete markers, **When** `terraform destroy` is run, **Then** all versions/markers are deleted and the bucket is deleted successfully.
3. **Given** a bucket with `force_destroy = true` containing multipart uploads, **When** `terraform destroy` is run, **Then** all uploads are aborted and the bucket is deleted successfully.
4. **Given** a bucket with `force_destroy = false` containing objects, **When** `terraform destroy` is run, **Then** the operation fails with a "BucketNotEmpty" error (or similar).

### Edge Cases

- **Large Bucket**: Bucket contains thousands of objects (needs pagination).
- **Versioning**: Bucket has versioning enabled and has deleted objects (delete markers).
- **Multipart Uploads**: Bucket has incomplete multipart uploads.
- **Permissions**: User has permission to delete bucket but not list/delete objects (should fail gracefully).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The resource deletion logic MUST check for `force_destroy` flag.
- **FR-002**: If `force_destroy` is true, the system MUST use `ListObjectVersions` (handling pagination) to identify and delete all items in the bucket. This consolidated approach MUST cover current object versions, historical versions, and delete markers in a single logical pass.
- **FR-004**: If `force_destroy` is true, the system MUST list and abort all incomplete multipart uploads, handling pagination to ensure all uploads are processed.
- **FR-005**: If deletion of any object/version fails, the bucket deletion MUST fail immediately and report the error (Fail Fast).
- **FR-006**: The system MUST retry deletion operations if transient errors occur, consistent with provider retry logic.

### Key Entities

- **OSS Bucket**: The resource being managed.
- **OSS Object**: Files stored in the bucket.
- **OSS Object Version**: Versions of files in versioned buckets.
- **OSS Multipart Upload**: In-progress uploads.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `terraform destroy` completes successfully for a bucket with 100+ objects when `force_destroy=true`.
- **SC-002**: `terraform destroy` completes successfully for a bucket with versioning and delete markers when `force_destroy=true`.
- **SC-003**: No `BucketNotEmpty` errors are observed during destruction of `force_destroy=true` resources.

## Clarifications

### Session 2026-01-23

- Q: Object Version Cleaning Scope → A: Always perform version check - Attempt to list/delete versions for all buckets, ensuring any residual versions in Suspended buckets are removed.
- Q: Error Handling Strategy → A: Stop on first error (Fail Fast) - Immediately abort the destruction process and return the error. Ensures issues are reported quickly.
- Q: Multipart Upload Pagination → A: Yes, use pagination - Implement pagination for `ListMultipartUploads` to ensure all incomplete uploads are found and aborted.
- Q: Redundancy of Object Deletion Requirements → A: Consolidate - Use `ListObjectVersions` (and `DeleteObjectVersions`) exclusively. It covers all objects, versions, and delete markers in one pass logic.

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
