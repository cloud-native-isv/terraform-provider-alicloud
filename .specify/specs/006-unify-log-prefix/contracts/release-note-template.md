# Release Note: SLS to LOG Prefix Unification

## Version: 1.0.0

### Breaking Changes

#### Deprecation of alicloud_sls_* Resources and Data Sources
As of version 1.0.0, all `alicloud_sls_*` prefixed resources and data sources have been deprecated in favor of the unified `alicloud_log_*` naming convention. The legacy `alicloud_sls_*` names will continue to function for a limited time but will generate deprecation warnings and will eventually be removed entirely.

**Affected Resources:**
- `alicloud_sls_alert` → Use `alicloud_log_alert`
- `alicloud_sls_collection_policy` → Use `alicloud_log_collection_policy`
- `alicloud_sls_etl` → Use `alicloud_log_etl`
- `alicloud_sls_scheduled_sql` → Use `alicloud_log_scheduled_sql`
- `alicloud_sls_consumer_group` → Use `alicloud_log_consumer_group`
- `alicloud_sls_oss_ingestion` → Use `alicloud_log_oss_ingestion`
- `alicloud_sls_oss_export` → Use `alicloud_log_oss_export`
- `alicloud_sls_dashboard` → Use `alicloud_log_dashboard`
- `alicloud_sls_store_index` → Use `alicloud_log_store_index`
- `alicloud_sls_machine_group` → Use `alicloud_log_machine_group`
- `alicloud_sls_store` → Use `alicloud_log_store`
- `alicloud_sls_project_logging` → Use `alicloud_log_project_logging`
- `alicloud_sls_project` → Use `alicloud_log_project`

**Affected Data Sources:**
- `alicloud_sls_alerts` → Use `alicloud_log_alerts`
- `alicloud_sls_projects` → Use `alicloud_log_projects`
- `alicloud_sls_service` → Use `alicloud_log_service`
- `alicloud_sls_stores` → Use `alicloud_log_stores`
- `alicloud_sls_query` → Use `alicloud_log_query`
- `alicloud_sls_machine_groups` → Use `alicloud_log_machine_groups`
- `alicloud_sls_alert_resource` → Use `alicloud_log_alert_resource`
- `alicloud_sls_logtail_config` → Use `alicloud_log_logtail_configs`
- `alicloud_sls_store_indexes` → Use `alicloud_log_store_indexes`

### Migration Instructions

#### Immediate Action Required
All users of the `alicloud_sls_*` resources and data sources must update their Terraform configurations to use the new `alicloud_log_*` equivalents.

**Before:**
```hcl
resource "alicloud_sls_project" "example" {
  name = "example-project"
  # ... configuration
}
```

**After:**
```hcl
resource "alicloud_log_project" "example" {
  name = "example-project"
  # ... configuration
}
```

#### State Migration
To avoid recreating resources, use the following commands to migrate your state:

```bash
# Migrate resources
terraform state mv 'alicloud_sls_project.old_name' 'alicloud_log_project.new_name'
terraform state mv 'alicloud_sls_store.old_name' 'alicloud_log_store.new_name'
# Continue for all affected resources...

# Migrate data sources (no state migration needed for data sources)
```

### Motivation

This change implements a consistent naming convention across all log-related services in the Alibaba Cloud Terraform provider. The `alicloud_log_*` naming is clearer and more consistent with industry terminology.

### Compatibility

- **Backward Compatibility**: Legacy `alicloud_sls_*` resources and data sources will continue to work in this version but will show deprecation warnings
- **Forward Compatibility**: New configurations should use `alicloud_log_*` only
- **Future Removal**: `alicloud_sls_*` resources will be completely removed in a future major version

### Support

- Migration Guide: Available in the provider documentation
- Support: Contact Alibaba Cloud support for complex migration scenarios
- Issues: Report any migration issues on the provider's GitHub repository

### Additional Notes

For detailed migration examples and troubleshooting tips, please refer to the migration guide in the documentation.