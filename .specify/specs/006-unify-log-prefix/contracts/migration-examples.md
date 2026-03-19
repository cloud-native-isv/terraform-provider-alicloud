# Migration Examples: SLS to LOG Prefix Unification

## Purpose

This document provides concrete examples of how to migrate from `alicloud_sls_*` to `alicloud_log_*` naming in Terraform configurations.

## Example 1: Migrate SLS Project Configuration

### Before Migration (Legacy)
```hcl
resource "alicloud_sls_project" "example_project" {
  name        = "example-project"
  description = "Example SLS project"
}
```

### After Migration (Unified)
```hcl
resource "alicloud_log_project" "example_project" {
  name        = "example-project"
  description = "Example SLS project"
}
```

## Example 2: Migrate SLS Store Configuration

### Before Migration (Legacy)
```hcl
resource "alicloud_sls_store" "example_store" {
  project  = alicloud_log_project.example_project.name
  name     = "example-store"
  ttl      = 30
  shard_count = 2
}
```

### After Migration (Unified)
```hcl
resource "alicloud_log_store" "example_store" {
  project  = alicloud_log_project.example_project.name
  name     = "example-store"
  ttl      = 30
  shard_count = 2
}
```

## Example 3: Migrate SLS Alert Configuration

### Before Migration (Legacy)
```hcl
resource "alicloud_sls_alert" "example_alert" {
  project = alicloud_log_project.example_project.name
  name    = "example-alert"
  # ... other configuration
}
```

### After Migration (Unified)
```hcl
resource "alicloud_log_alert" "example_alert" {
  project = alicloud_log_project.example_project.name
  name    = "example-alert"
  # ... other configuration
}
```

## Example 4: Migrate SLS Projects Data Source

### Before Migration (Legacy)
```hcl
data "alicloud_sls_projects" "example_projects" {
  name_regex = "^example-"
  output_file = "sls_projects.txt"
}
```

### After Migration (Unified)
```hcl
data "alicloud_log_projects" "example_projects" {
  name_regex = "^example-"
  output_file = "log_projects.txt"
}
```

## Example 5: Complex Migration Scenario

### Before Migration (Legacy)
```hcl
# SLS Project
resource "alicloud_sls_project" "app_project" {
  name        = "app-log-project"
  description = "Application logging project"
}

# SLS Store
resource "alicloud_sls_store" "app_store" {
  project     = alicloud_sls_project.app_project.name
  name        = "app-log-store"
  description = "Application logs store"
  ttl         = 90
  shard_count = 3
}

# SLS Dashboard
resource "alicloud_sls_dashboard" "app_dashboard" {
  project = alicloud_sls_project.app_project.name
  name    = "app-dashboard"
  detail  = jsonencode({
    title = "Application Logs"
    # ... dashboard configuration
  })
}

# SLS data source
data "alicloud_sls_projects" "all_projects" {
  name_regex = ".*"
}
```

### After Migration (Unified)
```hcl
# LOG Project
resource "alicloud_log_project" "app_project" {
  name        = "app-log-project"
  description = "Application logging project"
}

# LOG Store
resource "alicloud_log_store" "app_store" {
  project     = alicloud_log_project.app_project.name
  name        = "app-log-store"
  description = "Application logs store"
  ttl         = 90
  shard_count = 3
}

# LOG Dashboard
resource "alicloud_log_dashboard" "app_dashboard" {
  project = alicloud_log_project.app_project.name
  name    = "app-dashboard"
  detail  = jsonencode({
    title = "Application Logs"
    # ... dashboard configuration
  })
}

# LOG data source
data "alicloud_log_projects" "all_projects" {
  name_regex = ".*"
}
```

## Migration Steps

1. **Identify**: Find all instances of `alicloud_sls_*` in your Terraform configuration
2. **Replace**: Replace with the corresponding `alicloud_log_*` names using the mapping table
3. **Verify**: Check that the resource/data source schemas are compatible
4. **Test**: Apply the changes in a test environment first
5. **Execute**: Apply to production after verification

## Common Issues and Solutions

### Issue: State Migration Required
**Problem**: Resource names changed, previous state won't match
**Solution**: Use `terraform state mv` for each migrated resource:
```bash
terraform state mv 'alicloud_sls_project.app_project' 'alicloud_log_project.app_project'
terraform state mv 'alicloud_sls_store.app_store' 'alicloud_log_store.app_store'
```

### Issue: Schema Differences
**Problem**: Some unified resources might have slightly different schemas
**Solution**: Consult the documentation for the new `alicloud_log_*` resources to adjust your configuration accordingly

## Risk Notices

- **Downtime Risk**: Some resources may require recreation when migrating
- **State Management**: Manual state migration may be required
- **Dependency Chain**: Update all dependent resources to reference new names
- **Documentation**: Verify that new `alicloud_log_*` resources have the same functionality

## Rollback Guidance

If migration causes issues:

1. Revert Terraform configuration back to `alicloud_sls_*` naming (if still supported)
2. Or, manually migrate state back using `terraform state mv` command
3. Test thoroughly before applying to production