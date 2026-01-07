# Data Model: Tablestore Instance VCU Extensions

## Entities

### `alicloud_ots_instance`

Current Schema Extensions (already present or to be confirmed):

| Field | Type | Required | Computed | Description |
|---|---|---|---|---|
| `instance_specification` | String | Yes | No | updated ENUM to include "VCU" |
| `elastic_vcu_upper_limit` | Float | No | Yes | Upper limit for VCU scaling |

## States

- `instance_specification` "VCU" -> implies `elastic_vcu_upper_limit` is relevant.
- Non-VCU types -> `elastic_vcu_upper_limit` ignored or error.

## Validation

- `instance_specification`: OneOf ["SSD", "HYBRID", "VCU"]
- `elastic_vcu_upper_limit`: > 0 (float)
