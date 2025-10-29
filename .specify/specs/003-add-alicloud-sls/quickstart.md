# Quickstart: alicloud_sls_consumer_group

## Prerequisites
- Configured Alicloud credentials in Terraform provider
- Existing SLS Project and Logstore

## Minimal Example

```hcl
resource "alicloud_sls_consumer_group" "cg" {
  project         = "my-project"
  logstore        = "app-logs"
  consumer_group  = "etl-workers"

  # Behavioral parameters
  timeout = 60   # seconds
  order   = false

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
```

## Import

```
terraform import alicloud_sls_consumer_group.cg project:logstore:consumer_group
```

- Import ID format: `project:logstore:consumer_group`
- After import, `terraform plan` should show no drift for supported fields.

## Mutability
- ForceNew (replacement on change): `project`, `logstore`, `consumer_group`
- In-place update: `timeout`, `order`

## Create Idempotency
- If a group with the same identifiers exists, the resource will adopt it and converge `timeout`/`order` to match HCL.

## Notes
- Validation errors are surfaced during `plan` for missing/invalid fields
- Provider uses robust retry for transient errors (throttling/system busy)
