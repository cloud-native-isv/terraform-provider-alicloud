# Provider Registration Baseline: alicloud_sls_* Entries

## Generated: 2026-03-19

This document captures the current state of SLS-related resource and data source registrations in the provider, serving as a baseline for the transition to `alicloud_log_*` naming.

## Current Resource Registrations in alicloud/provider.go

### SLS Resources
```
"alicloud_sls_alert":                                        resourceAliCloudSlsAlert(),
"alicloud_sls_collection_policy":                           resourceAliCloudSlsCollectionPolicy(),
"alicloud_sls_etl":                                         resourceAliCloudLogETL(),  // Note: This one already uses LogETL function
"alicloud_sls_scheduled_sql":                              resourceAliCloudSlsScheduledSQL(),
"alicloud_sls_consumer_group":                             resourceAliCloudSlsConsumerGroup(),
"alicloud_sls_oss_ingestion":                             resourceAliCloudSlsOssIngestion(),
"alicloud_sls_oss_export":                                resourceAliCloudSlsOssExport(),
"alicloud_sls_dashboard":                                 resourceAliCloudSlsDashboard(),
"alicloud_sls_store_index":                               resourceAliCloudSlsStoreIndex(),
"alicloud_sls_machine_group":                             resourceAliCloudSlsMachineGroup(),
"alicloud_sls_store":                                     resourceAliCloudSlsStore(),
"alicloud_sls_project_logging":                           resourceAliCloudSlsProjectLogging(),
"alicloud_sls_project":                                   resourceAliCloudSlsProject(),
```

## Current Data Source Registrations in alicloud/provider.go

### SLS Data Sources
```
"alicloud_sls_alerts":                                    dataSourceAliCloudSlsAlerts(),
"alicloud_sls_projects":                                  dataSourceAliCloudSlsProjects(),
"alicloud_sls_service":                                   dataSourceAliCloudSlsService(),
"alicloud_sls_stores":                                    dataSourceAliCloudSlsStores(),
"alicloud_sls_query":                                     dataSourceAliCloudSlsQuery(),
"alicloud_sls_machine_groups":                            dataSourceAliCloudSlsMachineGroups(),
"alicloud_sls_alert_resource":                            dataSourceAliCloudLogAlertResource(),  // Note: Uses LogAlertResource function
"alicloud_sls_logtail_config":                            dataSourceAliCloudSlsLogtailConfig(),
"alicloud_sls_store_indexes":                             dataSourceAliCloudSlsStoreIndexes(),
```

## Key Observations

1. Some resources already use functions with "Log" naming rather than "Sls" (e.g., `resourceAliCloudLogETL()` for `alicloud_sls_etl`).

2. The `alicloud_sls_etl` resource points to `resourceAliCloudLogETL()` function.

3. The `alicloud_sls_alert_resource` data source points to `dataSourceAliCloudLogAlertResource()` function.

4. This baseline will help track the transition from legacy `alicloud_sls_*` keys to `alicloud_log_*` keys while maintaining or updating the underlying function implementations.

## Expected Transition Targets

Based on our naming mapping, these registrations should transition to:

### Resource Transitions
- `alicloud_sls_alert` → `alicloud_log_alert`
- `alicloud_sls_collection_policy` → `alicloud_log_collection_policy`
- `alicloud_sls_etl` → `alicloud_log_etl`
- `alicloud_sls_scheduled_sql` → `alicloud_log_scheduled_sql`
- `alicloud_sls_consumer_group` → `alicloud_log_consumer_group`
- `alicloud_sls_oss_ingestion` → `alicloud_log_oss_ingestion`
- `alicloud_sls_oss_export` → `alicloud_log_oss_export`
- `alicloud_sls_dashboard` → `alicloud_log_dashboard`
- `alicloud_sls_store_index` → `alicloud_log_store_index`
- `alicloud_sls_machine_group` → `alicloud_log_machine_group`
- `alicloud_sls_store` → `alicloud_log_store`
- `alicloud_sls_project_logging` → `alicloud_log_project_logging`
- `alicloud_sls_project` → `alicloud_log_project`

### Data Source Transitions
- `alicloud_sls_alerts` → `alicloud_log_alerts`
- `alicloud_sls_projects` → `alicloud_log_projects`
- `alicloud_sls_service` → `alicloud_log_service`
- `alicloud_sls_stores` → `alicloud_log_stores`
- `alicloud_sls_query` → `alicloud_log_query`
- `alicloud_sls_machine_groups` → `alicloud_log_machine_groups`
- `alicloud_sls_alert_resource` → `alicloud_log_alert_resource`
- `alicloud_sls_logtail_config` → `alicloud_log_logtail_configs`
- `alicloud_sls_store_indexes` → `alicloud_log_store_indexes`