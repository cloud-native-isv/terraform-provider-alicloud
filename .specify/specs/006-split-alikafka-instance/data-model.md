# Data Model: Split AliKafka Instance Resource

## Resources

### `alicloud_alikafka_instance` (Modified)

**Description**: Creates an AliKafka instance (order) but does not start it.

**Schema Changes**:
- **Removed**: `vswitch_id` (Moved to deployment, or kept for reference? `StartInstance` needs it. If we remove it from instance, we can't pass it to `StartInstance` unless deployment takes it. The current `StartInstance` takes `VSwitchId`. So `alicloud_alikafka_deployment` should probably take `vswitch_id`. However, `CreateOrder` might need it? Let's check `CreateOrder` in `service_alicloud_alikafka.go`. `CreateOrder` takes `DiskSize`, `DiskType`, `DeployType`, etc. It does NOT take `VSwitchId`. `StartInstance` takes `VSwitchId`. So `vswitch_id` should be moved to `alicloud_alikafka_deployment` or kept in `instance` and passed to `deployment`?
- **Decision**: `vswitch_id` is a deployment parameter. It should be in `alicloud_alikafka_deployment`.
- **However**: To maintain some backward compatibility or ease of use, maybe we keep it in `instance` and `deployment` reads it from `instance`? No, strict split means `deployment` handles deployment params.
- **Wait**: `alicloud_alikafka_instance` currently has `vswitch_id` as Required. If we remove it, it's a breaking change. The spec accepted breaking changes.
- **Refined Schema**:
    - `name`: Optional
    - `partition_num`: Optional
    - `disk_type`: Required
    - `disk_size`: Required
    - `deploy_type`: Required
    - `io_max`: Optional
    - `io_max_spec`: Optional
    - `spec_type`: Optional
    - `paid_type`: Optional
    - `eip_max`: Optional
    - `resource_group_id`: Optional
    - `kms_key_id`: Optional
    - `tags`: Optional
    - `vpc_id`: Optional (Computed?)
    - `zone_id`: Optional (Computed?)

### `alicloud_alikafka_deployment` (New)

**Description**: Deploys (starts) an AliKafka instance.

**Schema**:
- `instance_id`: Required, ForceNew (The ID of the `alicloud_alikafka_instance`)
- `vswitch_id`: Required, ForceNew (Moved from instance)
- `vpc_id`: Optional, ForceNew
- `zone_id`: Optional, ForceNew
- `security_group`: Optional, ForceNew
- `service_version`: Optional
- `config`: Optional
- `selected_zones`: Optional
- `cross_zone`: Optional

## Entities

### `StopInstanceRequest`

**Fields**:
- `InstanceId`: string
- `RegionId`: string

### `StartInstanceRequest` (Existing)

**Fields**:
- `InstanceId`: string
- `RegionId`: string
- `VSwitchId`: string
- ... (other fields)
