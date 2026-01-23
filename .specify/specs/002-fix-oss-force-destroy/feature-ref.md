# Feature Reference: 002-fix-oss-force-destroy

## Core Functionality
- **Logic**: `OssService.EmptyBucket(bucketName)`
- **Trigger**: `resource_alicloud_oss_bucket` `Delete` method when `force_destroy` is true.

## API Changes
- **New Method**: `OssService.EmptyBucket`
- **Modified Resource**: `resource_alicloud_oss_bucket`

## Configuration
- No new Terraform configuration parameters.
- Relies on existing `force_destroy`.

## Error Handling
- **Fail Fast**: Any error during listing or deletion aborts the operation.
- **Errors**: Propagated from SDK.
