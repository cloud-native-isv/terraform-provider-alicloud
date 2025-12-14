# Data Model: Update SelectDB Resources

## Resource: alicloud_selectdb_instance

### Schema

| Field | Type | Required | Optional | Computed | ForceNew | Description |
|-------|------|----------|----------|----------|----------|-------------|
| `instance_name` | String | Yes | - | - | - | The description of the SelectDB instance. |
| `resource_group_id` | String | - | Yes | Yes | - | The resource group ID. |
| `tags` | Map(String) | - | Yes | - | - | A mapping of tags to assign to the resource. |
| `engine` | String | - | Yes | - | Yes | The engine type. Default: `selectdb`. |
| `engine_version` | String | - | Yes | Yes | - | The engine version. |
| `engine_minor_version` | String | - | - | Yes | - | The engine minor version. |
| `zone_id` | String | Yes | - | - | Yes | The zone ID. |
| `vpc_id` | String | Yes | - | - | Yes | The VPC ID. |
| `vswitch_id` | String | Yes | - | - | Yes | The VSwitch ID. |
| `multi_zone` | List | - | - | Yes | - | Multi-zone configuration. |
| `instance_class` | String | Yes | - | - | Yes | The instance class. |
| `cache_size` | Int | Yes | - | - | Yes | The cache size in GB. |
| `deploy_scheme` | String | - | Yes | - | Yes | The deployment scheme. Default: `single_az`. |
| `charge_type` | String | Yes | - | - | Yes | The payment type. Valid values: `Prepaid`, `Postpaid`. |
| `period` | String | - | Yes | - | Yes | The billing period for Prepaid. |
| `period_time` | Int | - | Yes | - | Yes | The period time for Prepaid. |
| `maintain_start_time` | String | - | Yes | Yes | - | Maintenance start time. |
| `maintain_end_time` | String | - | Yes | Yes | - | Maintenance end time. |
| `username` | String | Yes | - | - | - | Admin username. |
| `password` | String | Yes | - | - | - | Admin password. |
| `security_ip_groups` | Set | - | Yes | - | - | Security IP groups. |
| `status` | String | - | - | Yes | - | Instance status. |
| `category` | String | - | - | Yes | - | Instance category. |
| `instance_used_type` | String | - | - | Yes | - | Instance used type. |
| `connection_string` | String | - | - | Yes | - | Connection string. |
| `sub_domain` | String | - | - | Yes | - | Sub domain. |
| `resource_cpu` | Int | - | - | Yes | - | CPU cores. |
| `resource_memory` | Int | - | - | Yes | - | Memory size in GB. |
| `storage_size` | Int | - | - | Yes | - | Storage size in GB. |
| `storage_type` | String | - | - | Yes | - | Storage type. |
| `object_store_size` | Int | - | - | Yes | - | Object store size in GB. |
| `cluster_count` | Int | - | - | Yes | - | Cluster count. |
| `scale_min` | Int | - | - | Yes | - | Minimum scale. |
| `scale_max` | Int | - | - | Yes | - | Maximum scale. |
| `scale_replica` | Int | - | - | Yes | - | Scale replica. |
| `lock_mode` | Int | - | - | Yes | - | Lock mode. |
| `lock_reason` | String | - | - | Yes | - | Lock reason. |
| `create_time` | String | - | - | Yes | - | Creation time. |
| `gmt_created` | String | - | - | Yes | - | GMT creation time. |
| `gmt_modified` | String | - | - | Yes | - | GMT modification time. |
| `expire_time` | String | - | - | Yes | - | Expiration time. |
| `can_upgrade_versions` | List(String) | - | - | Yes | - | Upgradeable versions. |
| `instance_net_infos` | List | - | - | Yes | - | Network information. |
| `security_ip_lists` | List | - | - | Yes | - | Security IP lists (computed). |
| `cluster_list` | List | - | - | Yes | - | Database cluster list. |

## Resource: alicloud_selectdb_cluster

### Schema

| Field | Type | Required | Optional | Computed | ForceNew | Description |
|-------|------|----------|----------|----------|----------|-------------|
| `instance_id` | String | Yes | - | - | Yes | The ID of the SelectDB instance. |
| `cluster_name` | String | - | - | Yes | - | The name of the SelectDB cluster. |
| `description` | String | Yes | - | - | - | The description of the SelectDB cluster. |
| `zone_id` | String | Yes | - | - | Yes | The zone ID. |
| `vpc_id` | String | Yes | - | - | Yes | The VPC ID. |
| `vswitch_id` | String | Yes | - | - | Yes | The VSwitch ID. |
| `cluster_class` | String | Yes | - | - | - | The cluster class. |
| `cache_size` | Int | Yes | - | - | - | The cache size in GB. |
| `engine` | String | - | Yes | - | - | The engine type. Default: `selectdb`. |
| `engine_version` | String | - | Yes | - | - | The engine version. Default: `4.0`. |
| `charge_type` | String | - | Yes | - | - | The billing method. Default: `PostPaid`. |
| `params` | Set | - | Yes | - | - | Configuration parameters. |
| `cluster_id` | String | - | - | Yes | - | The cluster ID. |
| `status` | String | - | - | Yes | - | The cluster status. |
| `create_time` | String | - | - | Yes | - | Creation time. |
| `cpu_cores` | Int | - | - | Yes | - | CPU cores. |
| `memory` | Int | - | - | Yes | - | Memory in GB. |
| `cache_storage_type` | String | - | - | Yes | - | Cache storage type. |
| `performance_level` | String | - | - | Yes | - | Performance level. |
| `scaling_rules_enable` | Bool | - | - | Yes | - | Whether scaling rules are enabled. |
| `sub_domain` | String | - | - | Yes | - | Sub domain. |

## Data Source: alicloud_selectdb_instances

### Schema

| Field | Type | Description |
|-------|------|-------------|
| `ids` | List(String) | A list of Instance IDs. |
| `tags` | Map(String) | A mapping of tags to assign to the resource. |
| `output_file` | String | File name where to save data source results. |
| `instances` | List | A list of SelectDB Instances. |

### Instance Object

Includes all fields from `alicloud_selectdb_instance` resource.

## Data Source: alicloud_selectdb_clusters

### Schema

| Field | Type | Description |
|-------|------|-------------|
| `ids` | List(String) | A list of Cluster IDs. |
| `output_file` | String | File name where to save data source results. |
| `clusters` | List | A list of SelectDB Clusters. |

### Cluster Object

Includes all fields from `alicloud_selectdb_cluster` resource.
