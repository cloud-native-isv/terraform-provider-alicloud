# Contract: Provider Registration

**Spec**: [requirements.md](../requirements.md)

This document defines the provider registration contract for ADBPG resources and data sources.

## Resource Registration

The following entries MUST be added to the `ResourcesMap` in `alicloud/provider.go`:

| Resource Key | Function | Source File |
|-------------|----------|-------------|
| `alicloud_adbpg_instance` | `resourceAliCloudAdbpgInstance()` | `resource_alicloud_adbpg_instance.go` |
| `alicloud_adbpg_account` | `resourceAliCloudAdbpgAccount()` | `resource_alicloud_adbpg_account.go` |
| `alicloud_adbpg_database` | `resourceAliCloudAdbpgDatabase()` | `resource_alicloud_adbpg_database.go` |
| `alicloud_adbpg_connection` | `resourceAliCloudAdbpgConnection()` | `resource_alicloud_adbpg_connection.go` |
| `alicloud_adbpg_security_ip_array` | `resourceAliCloudAdbpgSecurityIpArray()` | `resource_alicloud_adbpg_security_ip_array.go` |
| `alicloud_adbpg_backup_policy` | `resourceAliCloudAdbpgBackupPolicy()` | `resource_alicloud_adbpg_backup_policy.go` |
| `alicloud_adbpg_ssl` | `resourceAliCloudAdbpgSsl()` | `resource_alicloud_adbpg_ssl.go` |

## Data Source Registration

The following entries MUST be added to the `DataSourcesMap` in `alicloud/provider.go`:

| Data Source Key | Function | Source File |
|----------------|----------|-------------|
| `alicloud_adbpg_instances` | `dataSourceAliCloudAdbpgInstances()` | `data_source_alicloud_adbpg_instances.go` |
| `alicloud_adbpg_accounts` | `dataSourceAliCloudAdbpgAccounts()` | `data_source_alicloud_adbpg_accounts.go` |
| `alicloud_adbpg_zones` | `dataSourceAliCloudAdbpgZones()` | `data_source_alicloud_adbpg_zones.go` |

## Function Naming Convention

- Resource schema functions: `resourceAliCloudAdbpg<Resource>()` returns `*schema.Resource`
- CRUD functions: `resourceAliCloudAdbpg<Resource><Operation>(d *schema.ResourceData, meta interface{}) error`
- Data source functions: `dataSourceAliCloudAdbpg<Resource>()` returns `*schema.Resource`
- Data source read: `dataSourceAliCloudAdbpg<Resource>Read(d *schema.ResourceData, meta interface{}) error`

## Timeout Defaults

| Resource | Create | Update | Delete |
|----------|--------|--------|--------|
| instance | 60 min | 30 min | 30 min |
| account | 5 min | 5 min | 5 min |
| database | 5 min | - | 5 min |
| connection | 10 min | - | 10 min |
| security_ip_array | 5 min | 5 min | 5 min |
| backup_policy | 5 min | 5 min | 5 min |
| ssl | 10 min | 10 min | 10 min |
