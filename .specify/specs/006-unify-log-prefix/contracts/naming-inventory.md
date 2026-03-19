# Naming Inventory Baseline: alicloud_sls_* → alicloud_log_*

## Generated: 2026-03-19

This document contains the complete baseline inventory of all current SLS-related resources and data sources, serving as the foundation for mapping to unified `alicloud_log_*` naming convention.

## Resource Inventory

### Current alicloud_sls_* Resources

| Capability Key | Legacy Name | Unified Name | Status |
|----------------|-------------|--------------|---------|
| alert | alicloud_sls_alert | alicloud_log_alert | unified |
| collection_policy | alicloud_sls_collection_policy | alicloud_log_collection_policy | unified |
| etl | alicloud_sls_etl | alicloud_log_etl | unified |
| scheduled_sql | alicloud_sls_scheduled_sql | alicloud_log_scheduled_sql | unified |
| consumer_group | alicloud_sls_consumer_group | alicloud_log_consumer_group | unified |
| oss_ingestion | alicloud_sls_oss_ingestion | alicloud_log_oss_ingestion | unified |
| oss_export | alicloud_sls_oss_export | alicloud_log_oss_export | unified |
| dashboard | alicloud_sls_dashboard | alicloud_log_dashboard | unified |
| store_index | alicloud_sls_store_index | alicloud_log_store_index | unified |
| machine_group | alicloud_sls_machine_group | alicloud_log_machine_group | unified |
| store | alicloud_sls_store | alicloud_log_store | unified |
| project_logging | alicloud_sls_project_logging | alicloud_log_project_logging | unified |
| project | alicloud_sls_project | alicloud_log_project | unified |
| helper | alicloud_sls_alert_helper | NA | retired_with_note (helper function, not resource) |

## Data Source Inventory

### Current alicloud_sls_* Data Sources

| Capability Key | Legacy Name | Unified Name | Status |
|----------------|-------------|--------------|---------|
| alerts | alicloud_sls_alerts | alicloud_log_alerts | unified |
| projects | alicloud_sls_projects | alicloud_log_projects | unified |
| service | alicloud_sls_service | alicloud_log_service | unified |
| stores | alicloud_sls_stores | alicloud_log_stores | unified |
| query | alicloud_sls_query | alicloud_log_query | unified |
| machine_groups | alicloud_sls_machine_groups | alicloud_log_machine_groups | unified |
| alert_resources | alicloud_sls_alert_resource | alicloud_log_alert_resource | unified |
| logtail_configs | alicloud_sls_logtail_config | alicloud_log_logtail_configs | unified |
| store_indexes | alicloud_sls_store_indexes | alicloud_log_store_indexes | unified |

## Additional Notes

- Some `alicloud_log_*` resources already exist (e.g., `alicloud_log_alert`) and should be verified as the correct unified counterparts
- Helper files (like `alicloud_sls_alert_helper.go`) are utility functions and not directly mapped to user-facing resources
- The existing `alicloud_log_*` resources may need updates to ensure consistency with unified naming conventions
- All SLS-related resources in the provider need to be audited for completeness