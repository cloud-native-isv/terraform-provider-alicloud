# Naming Mapping: alicloud_sls_* → alicloud_log_*

## Mapping Rules

This document establishes the definitive mapping between legacy `alicloud_sls_*` names and their unified `alicloud_log_*` counterparts.

## Resource Mappings

| Legacy Name | Unified Name | Mapping Type | Status | Replacement Hint | Effective Version | Updated At |
|-------------|--------------|--------------|--------|------------------|-------------------|------------|
| alicloud_sls_alert | alicloud_log_alert | one_to_one | unified | Use `alicloud_log_alert` instead of `alicloud_sls_alert` | 1.0.0 | 2026-03-19 |
| alicloud_sls_collection_policy | alicloud_log_collection_policy | one_to_one | unified | Use `alicloud_log_collection_policy` instead of `alicloud_sls_collection_policy` | 1.0.0 | 2026-03-19 |
| alicloud_sls_etl | alicloud_log_etl | one_to_one | unified | Use `alicloud_log_etl` instead of `alicloud_sls_etl` | 1.0.0 | 2026-03-19 |
| alicloud_sls_scheduled_sql | alicloud_log_scheduled_sql | one_to_one | unified | Use `alicloud_log_scheduled_sql` instead of `alicloud_sls_scheduled_sql` | 1.0.0 | 2026-03-19 |
| alicloud_sls_consumer_group | alicloud_log_consumer_group | one_to_one | unified | Use `alicloud_log_consumer_group` instead of `alicloud_sls_consumer_group` | 1.0.0 | 2026-03-19 |
| alicloud_sls_oss_ingestion | alicloud_log_oss_ingestion | one_to_one | unified | Use `alicloud_log_oss_ingestion` instead of `alicloud_sls_oss_ingestion` | 1.0.0 | 2026-03-19 |
| alicloud_sls_oss_export | alicloud_log_oss_export | one_to_one | unified | Use `alicloud_log_oss_export` instead of `alicloud_sls_oss_export` | 1.0.0 | 2026-03-19 |
| alicloud_sls_dashboard | alicloud_log_dashboard | one_to_one | unified | Use `alicloud_log_dashboard` instead of `alicloud_sls_dashboard` | 1.0.0 | 2026-03-19 |
| alicloud_sls_store_index | alicloud_log_store_index | one_to_one | unified | Use `alicloud_log_store_index` instead of `alicloud_sls_store_index` | 1.0.0 | 2026-03-19 |
| alicloud_sls_machine_group | alicloud_log_machine_group | one_to_one | unified | Use `alicloud_log_machine_group` instead of `alicloud_sls_machine_group` | 1.0.0 | 2026-03-19 |
| alicloud_sls_store | alicloud_log_store | one_to_one | unified | Use `alicloud_log_store` instead of `alicloud_sls_store` | 1.0.0 | 2026-03-19 |
| alicloud_sls_project_logging | alicloud_log_project_logging | one_to_one | unified | Use `alicloud_log_project_logging` instead of `alicloud_sls_project_logging` | 1.0.0 | 2026-03-19 |
| alicloud_sls_project | alicloud_log_project | one_to_one | unified | Use `alicloud_log_project` instead of `alicloud_sls_project` | 1.0.0 | 2026-03-19 |

## Data Source Mappings

| Legacy Name | Unified Name | Mapping Type | Status | Replacement Hint | Effective Version | Updated At |
|-------------|--------------|--------------|--------|------------------|-------------------|------------|
| alicloud_sls_alerts | alicloud_log_alerts | one_to_one | unified | Use `alicloud_log_alerts` instead of `alicloud_sls_alerts` | 1.0.0 | 2026-03-19 |
| alicloud_sls_projects | alicloud_log_projects | one_to_one | unified | Use `alicloud_log_projects` instead of `alicloud_sls_projects` | 1.0.0 | 2026-03-19 |
| alicloud_sls_service | alicloud_log_service | one_to_one | unified | Use `alicloud_log_service` instead of `alicloud_sls_service` | 1.0.0 | 2026-03-19 |
| alicloud_sls_stores | alicloud_log_stores | one_to_one | unified | Use `alicloud_log_stores` instead of `alicloud_sls_stores` | 1.0.0 | 2026-03-19 |
| alicloud_sls_query | alicloud_log_query | one_to_one | unified | Use `alicloud_log_query` instead of `alicloud_sls_query` | 1.0.0 | 2026-03-19 |
| alicloud_sls_machine_groups | alicloud_log_machine_groups | one_to_one | unified | Use `alicloud_log_machine_groups` instead of `alicloud_sls_machine_groups` | 1.0.0 | 2026-03-19 |
| alicloud_sls_alert_resource | alicloud_log_alert_resource | one_to_one | unified | Use `alicloud_log_alert_resource` instead of `alicloud_sls_alert_resource` | 1.0.0 | 2026-03-19 |
| alicloud_sls_logtail_config | alicloud_log_logtail_configs | one_to_one | unified | Use `alicloud_log_logtail_configs` instead of `alicloud_sls_logtail_config` | 1.0.0 | 2026-03-19 |
| alicloud_sls_store_indexes | alicloud_log_store_indexes | one_to_one | unified | Use `alicloud_log_store_indexes` instead of `alicloud_sls_store_indexes` | 1.0.0 | 2026-03-19 |

## Special Cases

| Legacy Name | Status | Reason | Note |
|-------------|--------|--------|------|
| alicloud_sls_alert_helper | retired_with_note | Helper utility function, not a resource | Not user-facing, no direct replacement needed |
| alicloud_sls_*.go helper files | retired_with_note | Helper/utility functions, not user-facing resources | Should be refactored to support unified resources |

## Governance Notes

- This mapping serves as the single source of truth for all SLS-to-Log naming transitions
- Any new SLS-related resources must follow the `alicloud_log_*` naming convention
- The `status` field indicates the current state of the mapping implementation
- The `effective_version` indicates when the deprecation of legacy names takes effect