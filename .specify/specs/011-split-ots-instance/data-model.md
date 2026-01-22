# Data Model: Split OTS Instance Resources

## Entities

### `alicloud_ots_instance` (Legacy/Standard)

Represents standard Tablestore instances (SSD, HYBRID).

**Schema Fields (Terraform)**:
- `name`: String, Required, ForceNew (Validate: OTS Name rules)
- `instance_specification`: String, Required, ForceNew (Validate: "SSD", "HYBRID" ONLY)
- `description`: String, Optional
- `resource_group_id`: String, Optional, Computed
- `tags`: Map(String), Optional
- `network_source_acl`: Set(String), Optional, Computed
- `network_type_acl`: Set(String), Optional, Computed
- `status`: String, Computed
- `table_quota`: Int, Computed
- `create_time`: String, Computed
- `user_id`: String, Computed

**Removed Fields**:
- `elastic_vcu_upper_limit`
- `vcu_quota`

### `alicloud_ots_instance_vcu` (New)

Represents VCU-model Tablestore instances.

**Schema Fields (Terraform)**:
- `name`: String, Required, ForceNew (Validate: OTS Name rules)
- `description`: String, Optional
- `resource_group_id`: String, Optional, Computed
- `elastic_vcu_upper_limit`: Float, Optional, Computed
- `tags`: Map(String), Optional
- `network_source_acl`: Set(String), Optional, Computed
- `network_type_acl`: Set(String), Optional, Computed
- `status`: String, Computed
- `table_quota`: Int, Computed
- `vcu_quota`: Int, Computed (Read-only output)
- `create_time`: String, Computed
- `user_id`: String, Computed

**Hidden/Hardcoded**:
- `instance_specification`: Always set to "VCU" internally. Not exposed in Schema.

## Validations

- **Naming**: Length 3-16, letters/digits/hyphens, start with letter.
- **Specification**: `alicloud_ots_instance_vcu` implies "VCU". `alicloud_ots_instance` explicitly prohibits "VCU".
- **VCU Limit**: `elastic_vcu_upper_limit` > 0 if set.

## State Transitions

- **Create**: 
  - Call `CreateOtsInstance`.
  - Wait for status "Running" (or equivalent ready state).
- **Update**:
  - `description`, `elastic_vcu_upper_limit`, `tags`, `acl` can be updated.
  - Updates to `elastic_vcu_upper_limit` call specialized API or update param.
- **Delete**:
  - Call `DeleteOtsInstance`.
  - Wait for NotFound.

## API Contracts (Internal CWS-Lib-Go Structs)

We use `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore`.

**Struct**: `tablestore.TablestoreInstance`

```go
type TablestoreInstance struct {
    InstanceName             string
    InstanceDescription      string
    InstanceSpecification    string // Set to "VCU" for new resource
    ElasticVCUUpperLimit     float32
    // ... tags, resourceGroupId etc.
}
```
