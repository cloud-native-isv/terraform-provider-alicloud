# Quickstart: OSS Bucket Prune Before Delete

## Goal

Delete a non-empty OSS bucket by enabling `force_destroy`, which triggers cleanup before deletion.

## Steps (User Flow)

1. Configure an `alicloud_oss_bucket` resource with `force_destroy = true`.
2. Apply the configuration to create the bucket and upload objects (optional for testing).
3. Run `terraform destroy` for the bucket resource.
4. Observe that cleanup runs first, then the bucket is deleted after cleanup completes or times out.

## Expected Results

- Non-empty buckets with `force_destroy` are cleaned and deleted without manual intervention.
- If cleanup hits retryable errors, it retries automatically until completion or timeout.
- If cleanup fails with non-retryable errors, deletion stops with a clear error.
