# Data Model: OSS Public Access Block

## Entities

- Entity: OssBucketPublicAccessBlock
  - Description: Public access blocking configuration attached to a specific OSS bucket.
  - Fields:
    - bucket_name (string, identifier): OSS bucket name (resource ID in this context)
    - block_public_access (bool): Consolidated flag reflecting effective public access block configuration
  - Relationships:
    - 1:1 with OSS bucket
  - Validation Rules:
    - `bucket_name` must be non-empty and valid per OSS naming
    - `block_public_access` is boolean; computed in Read from underlying configuration

## State & Transitions

- No complex multi-state transitions. The configuration is set (PUT) and then read back until consistent.
- Create/Update: perform PUT, then wait using StateRefresh until `block_public_access` matches desired value.
- Delete: perform DELETE (not in scope for this change).

## Notes

- The internal representation maps multiple low-level flags
  (BlockPublicAcls, IgnorePublicAcls, BlockPublicPolicy, RestrictPublicBuckets) to a single
  `block_public_access` boolean in resource schema.
