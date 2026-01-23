# Quickstart: Force Destroy OSS Bucket

## Usage

To ensure an OSS bucket is destroyed even if it contains objects, versions, or multipart uploads, set `force_destroy` to `true`.

```hcl
resource "alicloud_oss_bucket" "bucket-force-destroy" {
  bucket        = "tf-test-bucket-force-destroy"
  acl           = "private"
  force_destroy = true
}
```

## Behavior

- When `terraform destroy` is run:
  - If `force_destroy` is `true`:
    1. The provider explicitly lists and deletes all multipart uploads.
    2. The provider explicitly lists and deletes all object versions and delete markers.
    3. The provider deletes the bucket.
    4. If any step fails, the destruction stops and reports the error.
  - If `force_destroy` is `false`:
    1. The provider attempts to delete the bucket.
    2. If the bucket is not empty, the destruction fails with a standard error.
