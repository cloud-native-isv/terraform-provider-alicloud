# Research: Update SelectDB Resources

**Feature**: Update SelectDB Resources
**Status**: Completed

## Decisions

### 1. API Mapping for SelectDB Instance

**Decision**: Map new API fields to Terraform schema in `alicloud_selectdb_instance`.

**Rationale**: The `cws-lib-go` SDK provides a comprehensive `Instance` struct. We need to expose these fields to users.

**Mapping**:
- `EngineMinorVersion` -> `engine_minor_version` (Computed)
- `DeployScheme` -> `deploy_scheme` (Optional, ForceNew)
- `InstanceUsedType` -> `instance_used_type` (Computed)
- `MultiZone` -> `multi_zone` (Computed, List)
- `ResourceCpu` -> `resource_cpu` (Computed)
- `ResourceMemory` -> `resource_memory` (Computed)
- `StorageSize` -> `storage_size` (Computed)
- `StorageType` -> `storage_type` (Computed)
- `ObjectStoreSize` -> `object_store_size` (Computed)
- `ClusterCount` -> `cluster_count` (Computed)
- `ScaleMin` -> `scale_min` (Computed)
- `ScaleMax` -> `scale_max` (Computed)
- `ScaleReplica` -> `scale_replica` (Computed)
- `LockMode` -> `lock_mode` (Computed)
- `LockReason` -> `lock_reason` (Computed)
- `ResourceGroupId` -> `resource_group_id` (Optional, Computed)
- `Tags` -> `tags` (Optional)
- `CanUpgradeVersions` -> `can_upgrade_versions` (Computed, List)
- `DBClusterList` -> `cluster_list` (Computed, List)

### 2. API Mapping for SelectDB Cluster

**Decision**: Map new API fields to Terraform schema in `alicloud_selectdb_cluster`.

**Rationale**: The `cws-lib-go` SDK provides a comprehensive `Cluster` struct.

**Mapping**:
- `ClusterId` -> `cluster_id` (Computed)
- `ClusterName` -> `cluster_name` (Computed)
- `ClusterDescription` -> `description` (Required)
- `InstanceId` -> `instance_id` (Required, ForceNew)
- `Status` -> `status` (Computed)
- `ClusterClass` -> `cluster_class` (Required)
- `ChargeType` -> `charge_type` (Optional)
- `Engine` -> `engine` (Optional)
- `EngineVersion` -> `engine_version` (Optional)
- `CpuCores` -> `cpu_cores` (Computed)
- `Memory` -> `memory` (Computed)
- `CacheSize` -> `cache_size` (Required)
- `CacheStorageType` -> `cache_storage_type` (Computed)
- `PerformanceLevel` -> `performance_level` (Computed)
- `ScalingRulesEnable` -> `scaling_rules_enable` (Computed)
- `VpcId` -> `vpc_id` (Required, ForceNew)
- `VSwitchId` -> `vswitch_id` (Required, ForceNew)
- `ZoneId` -> `zone_id` (Required, ForceNew)
- `SubDomain` -> `sub_domain` (Computed)
- `CreatedTime` -> `create_time` (Computed)
- `Params` -> `params` (Optional, Set)

### 3. Data Source Updates

**Decision**: Update `alicloud_selectdb_instances` and `alicloud_selectdb_clusters` to reflect the new schema fields.

**Rationale**: Data sources should provide a read-only view of all available resource attributes.

### 4. SDK Usage

**Decision**: Use `cws-lib-go` for all API interactions.

**Rationale**: Mandated by the Constitution and project requirements.

## Alternatives Considered

- **Using generic map for params**: Rejected. Strong typing is preferred and `cws-lib-go` provides structured types for parameters.
- **Ignoring computed fields**: Rejected. Users need visibility into the actual state of the resources, especially for fields like `status`, `lock_mode`, etc.
