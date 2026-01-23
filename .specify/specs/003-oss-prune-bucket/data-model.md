# Data Model: OSS Bucket Prune Before Delete

## Entities

### Bucket

- **Purpose**: Represents an OSS bucket managed by the provider lifecycle.
- **Key Attributes**:
  - `name` (string, unique within region)
  - `region` (string)
  - `force_destroy` (boolean)
  - `versioning_enabled` (boolean, derived)

### Stored Item

- **Purpose**: Represents any object or object version within a bucket.
- **Key Attributes**:
  - `key` (string)
  - `version_id` (string, optional)
  - `size_bytes` (integer)
  - `last_modified` (timestamp)

### Pending Upload

- **Purpose**: Represents a multipart upload that must be cleared before deletion.
- **Key Attributes**:
  - `upload_id` (string)
  - `key` (string)
  - `initiated_at` (timestamp)

## Relationships

- A **Bucket** contains many **Stored Items**.
- A **Bucket** contains many **Pending Uploads**.

## Identity & Uniqueness Rules

- `Bucket.name` + `region` uniquely identifies a bucket.
- `Stored Item` is uniquely identified by `key` + `version_id` within a bucket.
- `Pending Upload` is uniquely identified by `upload_id` within a bucket.

## Lifecycle Notes

- Cleanup removes all **Stored Items** (including versions) and **Pending Uploads**.
- Cleanup is idempotent; re-running on empty buckets produces no residual changes.
