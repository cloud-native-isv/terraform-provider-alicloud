# Data Model & Schema Mapping

**Feature**: Use CreateInstance API
**Date**: 2026-01-16

## Schema to API Mapping

The `alicloud_instance` resource schema maps to `ecs.CreateInstanceRequest` as follows:

| Terraform Schema Attribute | API Parameter (`CreateInstance`) | Notes |
| :--- | :--- | :--- |
| `image_id` | `ImageId` | Direct mapping. |
| `instance_type` | `InstanceType` | Direct mapping. |
| `security_groups` | `SecurityGroupId` | **Complex**: First ID used here. Rest used in `JoinSecurityGroup`. |
| `instance_name` | `InstanceName` | Direct mapping. |
| `description` | `Description` | Direct mapping. |
| `internet_charge_type` | `InternetChargeType` | Direct mapping. |
| `internet_max_bandwidth_out`| `InternetMaxBandwidthOut` | Direct mapping. **Warning**: No Public IP allocated. |
| `host_name` | `HostName` | Direct mapping. |
| `password` | `Password` | Direct mapping. |
| `kms_encrypted_password` | N/A | Handled via separate KMS logic if needed, or deciphered before `Password`. |
| `system_disk_category` | `SystemDisk.Category` | Direct mapping. |
| `system_disk_size` | `SystemDisk.Size` | Direct mapping. |
| `system_disk_name` | `SystemDisk.DiskName` | Direct mapping. |
| `system_disk_description` | `SystemDisk.Description` | Direct mapping. |
| `data_disks` | `DataDisk` (List) | Direct mapping of list items. |
| `tags` | `Tag` (List) | Direct mapping. |
| `user_data` | `UserData` | Direct mapping (Base64 encoded). |
| `vpc_id` | N/A | Implicit via `VSwitchId`. |
| `vswitch_id` | `VSwitchId` | Direct mapping. |
| `private_ip` | `PrivateIpAddress` | Direct mapping. |
| `instance_charge_type` | `InstanceChargeType` | `PrePaid` / `PostPaid`. |
| `period` | `Period` | For `PrePaid`. |
| `period_unit` | `PeriodUnit` | For `PrePaid`. |
| `auto_renew` | `AutoRenew` | For `PrePaid`. |
| `auto_renew_period` | `AutoRenewPeriod` | For `PrePaid`. |
| `spot_strategy` | `SpotStrategy` | Direct mapping. |
| `spot_price_limit` | `SpotPriceLimit` | Direct mapping. |
| `key_name` | `KeyPairName` | Direct mapping. |
| `role_name` | `RamRoleName` | Direct mapping. |
| `dry_run` | `DryRun` | Direct mapping. |

## Entity Relationships

-   **Instance** has many **Disks** (`SystemDisk` + `DataDisk`).
-   **Instance** belongs to one **VSwitch**.
-   **Instance** belongs to one **Security Group** (at creation) -> Many (post-creation).
-   **Instance** has one **Image**.

## State Transitions

1.  **Creation**: `CreateInstance` -> Instance State: `Stopped`.
2.  **Startup**: `StartInstance` -> Instance State: `Starting` -> `Running`.
