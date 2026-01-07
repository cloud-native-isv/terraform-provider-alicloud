# Research Findings: Tablestore VCU Support

## Unknown 1: `instance_specification` casing
**Decision**: Use "VCU" (all caps).
**Rationale**: Existing schema validation checks for "VCU" explicitly. Most Tablestore specs ("SSD", "HYBRID") are uppercase.

## Unknown 2: `elastic_vcu_upper_limit` read handling
**Decision**: Code inspection is confirmed.
**Rationale**: `convertTablestoreInstanceToSchema` in `alicloud/service_alicloud_ots_instance.go` sets `d.Set("elastic_vcu_upper_limit", instance.ElasticVCUUpperLimit)`. And `alicloud_tablestore_instance_api.go`'s `convertToTablestoreInstance` (inferred) populates this field. I will assume this is correct given the context provided in attachments.

## Unknown 3: API Layer Capabilities
**Decision**: `CreateVCUInstance` exists.
**Rationale**: The attached `alicloud_tablestore_instance_api.go` shows explicit support. `CreateInstance` method specifically branches for "VCU" type.

## Dependencies
- `cws-lib-go`: Already updated with VCU support (implied by attachments).
