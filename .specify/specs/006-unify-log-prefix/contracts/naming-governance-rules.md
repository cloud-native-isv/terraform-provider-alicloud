# Naming Governance Rules: SLS to Log Prefix Unification

## Governance Catalog

### Catalog ID: `006-unify-log-prefix-2026-03-19`

### Scope Rule
All `alicloud_sls_*` entries are included per requirement FR-009.

### Included Items (Unified)

#### Resources
```yaml
included_items:
  # Resources that have been unified from alicloud_sls_* to alicloud_log_*
  - capability_key: "alert"
    legacy_name: "alicloud_sls_alert"
    canonical_name: "alicloud_log_alert"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_alert instead of alicloud_sls_alert"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "collection_policy"
    legacy_name: "alicloud_sls_collection_policy"
    canonical_name: "alicloud_log_collection_policy"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_collection_policy instead of alicloud_sls_collection_policy"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "etl"
    legacy_name: "alicloud_sls_etl"
    canonical_name: "alicloud_log_etl"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_etl instead of alicloud_sls_etl"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "scheduled_sql"
    legacy_name: "alicloud_sls_scheduled_sql"
    canonical_name: "alicloud_log_scheduled_sql"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_scheduled_sql instead of alicloud_sls_scheduled_sql"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "consumer_group"
    legacy_name: "alicloud_sls_consumer_group"
    canonical_name: "alicloud_log_consumer_group"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_consumer_group instead of alicloud_sls_consumer_group"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "oss_ingestion"
    legacy_name: "alicloud_sls_oss_ingestion"
    canonical_name: "alicloud_log_oss_ingestion"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_oss_ingestion instead of alicloud_sls_oss_ingestion"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "oss_export"
    legacy_name: "alicloud_sls_oss_export"
    canonical_name: "alicloud_log_oss_export"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_oss_export instead of alicloud_sls_oss_export"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "dashboard"
    legacy_name: "alicloud_sls_dashboard"
    canonical_name: "alicloud_log_dashboard"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_dashboard instead of alicloud_sls_dashboard"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "store_index"
    legacy_name: "alicloud_sls_store_index"
    canonical_name: "alicloud_log_store_index"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_store_index instead of alicloud_sls_store_index"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "machine_group"
    legacy_name: "alicloud_sls_machine_group"
    canonical_name: "alicloud_log_machine_group"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_machine_group instead of alicloud_sls_machine_group"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "store"
    legacy_name: "alicloud_sls_store"
    canonical_name: "alicloud_log_store"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_store instead of alicloud_sls_store"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "project_logging"
    legacy_name: "alicloud_sls_project_logging"
    canonical_name: "alicloud_log_project_logging"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_project_logging instead of alicloud_sls_project_logging"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "project"
    legacy_name: "alicloud_sls_project"
    canonical_name: "alicloud_log_project"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_project instead of alicloud_sls_project"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
```

#### Data Sources
```yaml
  - capability_key: "alerts"
    legacy_name: "alicloud_sls_alerts"
    canonical_name: "alicloud_log_alerts"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_alerts instead of alicloud_sls_alerts"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "projects"
    legacy_name: "alicloud_sls_projects"
    canonical_name: "alicloud_log_projects"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_projects instead of alicloud_sls_projects"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "service"
    legacy_name: "alicloud_sls_service"
    canonical_name: "alicloud_log_service"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_service instead of alicloud_sls_service"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "stores"
    legacy_name: "alicloud_sls_stores"
    canonical_name: "alicloud_log_stores"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_stores instead of alicloud_sls_stores"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "query"
    legacy_name: "alicloud_sls_query"
    canonical_name: "alicloud_log_query"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_query instead of alicloud_sls_query"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "machine_groups"
    legacy_name: "alicloud_sls_machine_groups"
    canonical_name: "alicloud_log_machine_groups"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_machine_groups instead of alicloud_sls_machine_groups"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "alert_resource"
    legacy_name: "alicloud_sls_alert_resource"
    canonical_name: "alicloud_log_alert_resource"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_alert_resource instead of alicloud_sls_alert_resource"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "logtail_config"
    legacy_name: "alicloud_sls_logtail_config"
    canonical_name: "alicloud_log_logtail_configs"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_logtail_configs instead of alicloud_sls_logtail_config"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
    
  - capability_key: "store_indexes"
    legacy_name: "alicloud_sls_store_indexes"
    canonical_name: "alicloud_log_store_indexes"
    mapping_type: "one_to_one"
    status: "unified"
    replacement_hint: "Use alicloud_log_store_indexes instead of alicloud_sls_store_indexes"
    effective_version: "1.0.0"
    updated_at: "2026-03-19"
```

### Excluded Items
```yaml
excluded_items: []
```

### Pending Items
```yaml
pending_items: []
```

### Release Note Reference
```yaml
release_note_ref: ".specify/specs/006-unify-log-prefix/contracts/release-note-template.md"
```

### Last Reviewed At
```yaml
last_reviewed_at: "2026-03-19T00:00:00Z"
```

## State Machine Rules

### NamingMappingItem.status

- `pending -> unified`
  - Condition: Canonical name has been registered and tested, documentation updated.
- `pending -> retired_with_note`
  - Condition: Confirmed no direct replacement, notes provided.
- `unified -> retired_with_note`
  - Condition: Capability is completely removed, with release notes indicating deprecation.

### Governance Catalog Validation

Each release must execute consistency check:
- Count of `alicloud_sls_*` entries in provider code = Count of items in `included_items` + Count of items in `pending_items`

## Additional Governance Notes

- All future SLS-related resources must use the `alicloud_log_*` naming convention
- Any new capabilities that would have used `alicloud_sls_*` naming must instead use `alicloud_log_*`
- Regular audits should ensure no new `alicloud_sls_*` entries are accidentally added
- Documentation must consistently reference unified `alicloud_log_*` names
- Error messages for deprecated prefixes must include clear migration instructions