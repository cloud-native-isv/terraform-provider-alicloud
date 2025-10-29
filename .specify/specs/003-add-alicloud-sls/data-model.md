# Data Model: SLS Consumer Group

## Entities

### ConsumerGroup
- Keys (immutable → ForceNew):
  - project: string (SLS project name)
  - logstore: string (SLS logstore name)
  - consumer_group: string (consumer group name)
- Attributes (updatable):
  - timeout: integer (seconds), reasonable bounds enforced
  - order: boolean (ordered consumption)
- Derived/Computed:
  - id: string = `project:logstore:consumer_group`

## Relationships
- Project 1..N Logstore
- Logstore 1..N ConsumerGroup

## Validation Rules
- Required: project, logstore, consumer_group
- consumer_group naming: non-empty; provider-specific allowed charset; enforce length bounds (e.g., 1–128)
- timeout: positive integer; min 1, max reasonable (e.g., 86400) — exact limits may follow SLS constraints
- order: boolean
- id import format: must match `^([^:]+):([^:]+):([^:]+)$`

## Lifecycle & States (conceptual)
- Creating → Active (on success) | Failed (on error)
- Updating (timeout/order) → Active
- Deleting → Deleted (Read treats not found as cleared state)

## Notes
- Any change to keys triggers replacement (Terraform ForceNew)
- Adopt on create if exists, then converge timeout/order to desired
