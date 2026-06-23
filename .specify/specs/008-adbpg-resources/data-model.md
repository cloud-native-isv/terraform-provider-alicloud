# Data Model: ADBPG Resources

**Spec**: [requirements.md](requirements.md) | **Plan**: [plan.md](plan.md)

## Architecture Layer Mapping

```
Resource Layer (alicloud/)                    Service Layer (alicloud/)               API Layer (pkg/cws-lib-go)
─────────────────────────────────────         ──────────────────────                  ─────────────────────────
resource_alicloud_adbpg_instance.go      →    AdbpgService.CreateInstance()      →    AdbpgAPI.CreateInstance()
resource_alicloud_adbpg_account.go       →    AdbpgService.CreateAccount()       →    AdbpgAPI.CreateAccount()
resource_alicloud_adbpg_database.go      →    AdbpgService.CreateDatabase()      →    AdbpgAPI.CreateDatabase()
resource_alicloud_adbpg_connection.go    →    AdbpgService.AllocateConnection()  →    AdbpgAPI.AllocatePublicConnection()
resource_alicloud_adbpg_security_ip.go   →    AdbpgService.ModifySecurityIps()   →    AdbpgAPI.ModifySecurityIps()
resource_alicloud_adbpg_backup_policy.go →    AdbpgService.DescribeBackupPolicy()→    AdbpgAPI.DescribeBackupPolicy()
resource_alicloud_adbpg_ssl.go           →    AdbpgService.ModifyInstanceSSL()   →    AdbpgAPI.ModifyInstanceSSL()
data_source_alicloud_adbpg_instances.go  →    AdbpgService.ListInstances()       →    AdbpgAPI.ListInstances()
data_source_alicloud_adbpg_accounts.go   →    AdbpgService.ListAccounts()        →    AdbpgAPI.ListAccounts()
data_source_alicloud_adbpg_zones.go      →    AdbpgService.ListAvailableResources()→  AdbpgAPI.ListAvailableResources()
```

## Entities

### E1: AdbpgService (Service Layer Struct)

Service layer mediator between Terraform resources and the cws-lib-go API.

| Field | Type | Source |
|-------|------|--------|
| client | `*connectivity.AliyunClient` | Provider connectivity |
| adbpgAPI | `*adbpg.AdbpgAPI` | cws-lib-go API client |

**Constructor**: `NewAdbpgService(client *connectivity.AliyunClient) (*AdbpgService, error)` — converts `client.AccessKey`, `client.SecretKey`, `client.RegionId`, `client.SecurityToken` to `common.Credentials` and calls `adbpg.NewAdbpgAPI()`.

### E2: ADBPG Instance

**Terraform Resource**: `alicloud_adbpg_instance`  
**API Create Type**: `adbpg.AdbpgInstanceCreate`  
**API Detail Type**: `adbpg.AdbpgInstanceDetail`  
**ID Format**: `{db_instance_id}` (single value, returned by CreateInstance)

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Computed | - | DBInstanceId |
| description | TypeString | Optional | No | DBInstanceDescription |
| engine_version | TypeString | Required | Yes | EngineVersion |
| db_instance_class | TypeString | Required | No | DBInstanceClass |
| db_instance_mode | TypeString | Optional, Computed | Yes | DBInstanceMode |
| instance_network_type | TypeString | Optional, Computed | Yes | InstanceNetworkType |
| vpc_id | TypeString | Optional | Yes | VpcId |
| vswitch_id | TypeString | Optional | Yes | VSwitchId |
| zone_id | TypeString | Required | Yes | ZoneId |
| pay_type | TypeString | Optional, Computed | Yes | PayType |
| security_ip_list | TypeString | Optional, Computed | No | SecurityIPList |
| storage_size | TypeInt | Optional, Computed | No | StorageSize |
| storage_type | TypeString | Optional, Computed | Yes | StorageType |
| master_node_num | TypeInt | Optional, Computed | Yes | MasterNodeNum |
| seg_node_num | TypeInt | Required | No | SegNodeNum |
| resource_group_id | TypeString | Optional, Computed | No | ResourceGroupId |
| serverless_mode | TypeString | Optional, Computed | Yes | ServerlessMode |
| encryption_key | TypeString | Optional | Yes | EncryptionKey |
| encryption_type | TypeString | Optional | Yes | EncryptionType |
| tags | TypeMap | Optional | No | Tags |
| status | TypeString | Computed | - | DBInstanceStatus |
| connection_string | TypeString | Computed | - | ConnectionString |
| port | TypeString | Computed | - | Port |
| maintain_start_time | TypeString | Optional, Computed | No | MaintainStartTime |
| maintain_end_time | TypeString | Optional, Computed | No | MaintainEndTime |
| creation_time | TypeString | Computed | - | CreationTime |

**State Transitions**:
- Create: `→ Creating → Running`
- Delete: `Running → Deleting → (gone)`
- Modify (description/maintain): `Running → Running` (synchronous)
- Upgrade: `Running → ClassChanging → Running`

**Pending States**: `Creating`, `ClassChanging`, `NetAddressCreating`, `NetAddressDeleting`, `Restarting`  
**Target States**: `Running`  
**Fail States**: `DBInstanceStatus == "Locked"` (when lock_mode != "Unlock")

### E3: ADBPG Account

**Terraform Resource**: `alicloud_adbpg_account`  
**API Create Type**: `adbpg.AdbpgAccountCreate`  
**API List Type**: `adbpg.AdbpgAccount`  
**ID Format**: `{db_instance_id}:{account_name}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | (path param) |
| account_name | TypeString | Required | Yes | AccountName |
| account_password | TypeString | Required, Sensitive | No | AccountPassword |
| account_type | TypeString | Optional, Computed | Yes | AccountType |
| account_description | TypeString | Optional | No | AccountDescription |
| status | TypeString | Computed | - | AccountStatus |

**Update Logic**:
- `account_description` change → `ModifyAccountDescription` (not in current API; append to description field)
- `account_password` change → `ResetAccountPassword`
- `account_type` → ForceNew (cannot change type after creation)

### E4: ADBPG Database

**Terraform Resource**: `alicloud_adbpg_database`  
**API Create Type**: `adbpg.AdbpgDatabaseCreate`  
**API List Type**: `adbpg.AdbpgDatabase`  
**ID Format**: `{db_instance_id}:{db_name}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | (path param) |
| db_name | TypeString | Required | Yes | DBName |
| db_description | TypeString | Optional | Yes | DBDescription |
| character_name | TypeString | Optional | Yes | CharacterName |

**Lifecycle**: Create and Delete only. No update API available — all fields are ForceNew.

### E5: ADBPG Connection

**Terraform Resource**: `alicloud_adbpg_connection`  
**API Methods**: `AllocatePublicConnection` / `ReleasePublicConnection` / `DescribeInstanceNetInfo`  
**ID Format**: `{db_instance_id}:{connection_string_prefix}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | (path param) |
| connection_string_prefix | TypeString | Required | Yes | ConnectionStringPrefix |
| connection_string | TypeString | Computed | - | ConnectionString |
| ip_address | TypeString | Computed | - | IPAddress |
| port | TypeString | Computed | - | Port |

**Lifecycle**: Create (allocate) and Delete (release) only. Read via `DescribeInstanceNetInfo` filtering for public (IPType == "Public") entries.

### E6: ADBPG Security IP Array

**Terraform Resource**: `alicloud_adbpg_security_ip_array`  
**API Method**: `ModifySecurityIps`  
**ID Format**: `{db_instance_id}:{db_instance_ip_array_name}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | (path param) |
| db_instance_ip_array_name | TypeString | Optional (default: "default") | Yes | DBInstanceIPArrayName |
| security_ip_list | TypeString | Required | No | SecurityIPList |

**Lifecycle**: Create and Update both use `ModifySecurityIps`. Delete sets security_ip_list to "127.0.0.1" (default/empty state). Read via `DescribeInstance` and inspecting `SecurityIPList`.

### E7: ADBPG Backup Policy

**Terraform Resource**: `alicloud_adbpg_backup_policy`  
**API Methods**: `DescribeBackupPolicy`  
**API Type**: `adbpg.AdbpgBackupPolicy`  
**ID Format**: `{db_instance_id}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | (path param) |
| backup_retention_period | TypeInt | Optional, Computed | No | BackupRetentionPeriod |
| preferred_backup_period | TypeString | Optional, Computed | No | PreferredBackupPeriod |
| preferred_backup_time | TypeString | Optional, Computed | No | PreferredBackupTime |
| enable_recovery_point | TypeBool | Optional, Computed | No | EnableRecoveryPoint |

**Lifecycle**: This resource represents instance-level configuration. Create and Update both configure the backup policy. Delete resets to default values. There is no separate create/delete API — the policy always exists on the instance.

### E8: ADBPG SSL

**Terraform Resource**: `alicloud_adbpg_ssl`  
**API Methods**: `DescribeInstanceSSL` / `ModifyInstanceSSL`  
**API Type**: `adbpg.AdbpgSSLConfig`  
**ID Format**: `{db_instance_id}`

| Terraform Attribute | Schema Type | Required/Optional/Computed | ForceNew | API Field |
|---------------------|-------------|---------------------------|----------|-----------|
| db_instance_id | TypeString | Required | Yes | DBInstanceId |
| ssl_enabled | TypeBool | Required | No | SSLEnabled |
| ssl_expired | TypeString | Computed | - | SSLExpired |

**Lifecycle**: Create and Update both use `ModifyInstanceSSL`. Delete disables SSL. Read via `DescribeInstanceSSL`.

## Data Source Entities

### DS1: ADBPG Instances

**Terraform Data Source**: `data.alicloud_adbpg_instances`  
**API Method**: `ListInstances`

| Filter Attribute | Schema Type | Description |
|-----------------|-------------|-------------|
| ids | TypeList of TypeString | Filter by instance IDs |
| description_regex | TypeString | Regex to match description |
| status | TypeString | Filter by status |
| resource_group_id | TypeString | Filter by resource group |
| tags | TypeMap | Filter by tags |
| output_file | TypeString | Write results to file |

**Output**: List of instances with all `AdbpgInstance` attributes.

### DS2: ADBPG Accounts

**Terraform Data Source**: `data.alicloud_adbpg_accounts`  
**API Method**: `ListAccounts`

| Filter Attribute | Schema Type | Description |
|-----------------|-------------|-------------|
| db_instance_id | TypeString, Required | Instance to query |
| name_regex | TypeString | Regex to match account name |
| output_file | TypeString | Write results to file |

**Output**: List of accounts with `account_name`, `account_type`, `account_status`, `account_description`.

### DS3: ADBPG Zones

**Terraform Data Source**: `data.alicloud_adbpg_zones`  
**API Method**: `ListAvailableResources`

| Filter Attribute | Schema Type | Description |
|-----------------|-------------|-------------|
| multi | TypeBool | Multi-zone filter |
| output_file | TypeString | Write results to file |

**Output**: List of zones with supported engine versions and instance classes.

## ID Encoding/Decoding

Composite IDs use `:` separator, consistent with existing provider patterns.

```
EncodeAdbpgAccountId(instanceId, accountName) → "instanceId:accountName"
DecodeAdbpgAccountId(id) → (instanceId, accountName, error)

EncodeAdbpgDatabaseId(instanceId, dbName) → "instanceId:dbName"
DecodeAdbpgDatabaseId(id) → (instanceId, dbName, error)

EncodeAdbpgConnectionId(instanceId, prefix) → "instanceId:prefix"
DecodeAdbpgConnectionId(id) → (instanceId, prefix, error)

EncodeAdbpgSecurityIPArrayId(instanceId, arrayName) → "instanceId:arrayName"
DecodeAdbpgSecurityIPArrayId(id) → (instanceId, arrayName, error)
```

Single-value IDs (instance, backup_policy, ssl) use `db_instance_id` directly.
