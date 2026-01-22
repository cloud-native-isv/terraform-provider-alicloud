# Research: Split OTS Instance Resources

**Feature**: Split OTS Instance Resources
**Status**: Completed

## 1. Schema Definition for `alicloud_ots_instance_vcu`

**Decision**: 
- Create a new resource `alicloud_ots_instance_vcu`.
- **Remove** `instance_specification` from user input in Schema.
- **Hardcode** `InstanceSpecification: "VCU"` in the `CreateOtsInstance` API call within the resource implementation.
- Include `elastic_vcu_upper_limit` as an Optional/Computed field (specific to VCU).
- Include standard fields: `name`, `description`, `resource_group_id`, `tags`, `network_source_acl`, `network_type_acl`.

**Rationale**: 
- Clarification Q2 explicitly decided to remove/hardcode specification to avoid redundancy.
- Constitution requires strong typing; we will populate the CWS-Lib-Go struct directly.

**Alternatives Considered**:
- *Keep `instance_specification` as Computed/Default "VCU"*: Rejected per clarification to simplify user experience and strictly enforce the resource type's purpose.

## 2. Refactoring `alicloud_ots_instance`

**Decision**:
- **Remove** `instance_specification` validation for "VCU" (only allow "SSD", "HYBRID").
- **Remove** `elastic_vcu_upper_limit`, `vcu_quota` fields from Schema.
- **Fail Hard** on Create if user tries to somehow force VCU (though schema validation should catch this).
- **Hard Break** for existing state: If `terraform refresh` encounters a VCU instance managed by this resource, it might succeed reading (if fields match), but `Update` operations impacting removed fields would fail or be impossible. To support the "Hard Break" decision effectively, we will rely on removing the fields from schema; Terraform will simply ignore them or error depending on if they are in config. The explicit instruction is "Remove VCU support entirely".

**Rationale**:
- Clarification Q1 and Q3 mandated a hard break and removal of fields. This ensures code cleanliness and prevents "zombie" logic.

## 3. Service Layer Impact

**Decision**:
- The existing `OtsService` and `CreateOtsInstance` API in CWS-Lib-Go likely supports all parameters.
- We do NOT need to change the Service layer (`service_alicloud_ots.go`) unless it has validation logic preventing separate calls.
- We will reuse `OtsService` methods (`CreateOtsInstance`, `DescribeOtsInstance`, etc.) for both resources.

**Rationale**:
- Layering principle: Resources call Service. Service encapsulates API. The API is common; the parameters differentiate the instance type.

## 4. Migration Strategy

**Decision**:
- Documentation will instruct users to use `terraform state mv alicloud_ots_instance.old alicloud_ots_instance_vcu.new` if they have VCU instances.
- No automatic migration logic in the provider code (as it's a hard break/manual migration).

## 5. CWS-Lib-Go Integration

**Decision**:
- Verified `tablestoreAPI` imports in current file `resource_alicloud_ots_instance.go`.
- Will use `tablestoreAPI.TablestoreInstance` struct for all data passing.

**Rationale**:
- Strict compliance with Constitution Article V (Strong Typing).
