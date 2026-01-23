# Research Findings: Fix OSS Bucket Force Destroy

## Project Context Analysis

The project is a Terraform Provider for Alibaba Cloud. The codebase follows a layered architecture (Resource -> Service -> API). `resource_alicloud_oss_bucket` is the resource in question. It currently uses `EmptyBucket` in a reactive manner (only after `DeleteBucket` fails). This has proven unreliable ("BucketNotEmpty" errors).

We need to make `force_destroy` proactive: if set, empty the bucket first.

The `EmptyBucket` function is located in `alicloud/service_alicloud_oss_bucket.go`. It already implements robust logic (pagination, object versions, delete markers, multipart uploads).

There is a mandate to use `cws-lib-go` where possible, but `cws-lib-go`'s OSS support is limited (lacks versioning/multipart logic). Thus, `EmptyBucket` uses `aliyun-oss-go-sdk` directly, which is acceptable as legacy/bridge code.

## References

- `alicloud/resource_alicloud_oss_bucket.go`: Current resource implementation.
- `alicloud/service_alicloud_oss_bucket.go`: Current service implementation containing `EmptyBucket`.
- `alicloud/errors.go`: Error handling logic.
- `docs/development_guide.md`: Development standards.

## Decisions

### Proactive Bucket Emptying
- **Decision**: Refactor `resourceAliCloudOssBucketDelete` to call `EmptyBucket` *before* `DeleteBucket` if `force_destroy` is true.
- **Rationale**: Ensures deterministic behavior. Relying on `DeleteBucket` failure is flaky due to distributed system eventual consistency or error code mismatches.
- **Alternatives considered**: Improving `IsExpectedErrors` mapping. Rejected because it doesn't solve the core issue of needing to empty the bucket anyway.

### Use of Existing EmptyBucket
- **Decision**: Reuse the existing `EmptyBucket` implementation in `alicloud/service_alicloud_oss_bucket.go`.
- **Rationale**: It already contains the necessary logic for cleaning up versions and multipart uploads. It uses paginated lists which is correct for large buckets.

### Error Handling
- **Decision**: Fail fast if `EmptyBucket` encounters an error during `force_destroy`.
- **Rationale**: If we can't empty the bucket, we can't destroy it. Retrying blindly won't help if it's a permission issue or API error.
