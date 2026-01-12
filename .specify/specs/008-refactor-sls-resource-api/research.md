# Phase 0: Research and Discovery

## Goals
1. Identify all legacy SLS resources/data sources using `github.com/aliyun/aliyun-log-go-sdk` directly.
2. Verify target `SlsService` capabilities.
3. Determine refactoring strategy.

## Key Findings

### Legacy Files Identified
- `alicloud/resource_alicloud_log_alert.go`: Uses `client.WithLogClient` and direct SDK structs.
- `alicloud/data_source_alicloud_sls_logtail_config.go`: Uses `client.WithLogClient`.
- `alicloud/data_source_alicloud_sls_machine_groups.go`: Uses `client.WithLogClient`.

### Target Service Capabilities
- `alicloud/service_alicloud_sls_alert.go`: Fully implements `CreateSlsAlert`, `GetSlsAlert`, `UpdateSlsAlert`, `DeleteSlsAlert` using `cws-lib-go`.
- `alicloud/service_alicloud_sls_logtail_config.go`: Implements `ListSlsLogtailConfigs`, `DescribeSlsLogtailConfig`, etc.
- `alicloud/service_alicloud_sls_machine_group.go`: Implements `ListSlsMachineGroups`, `DescribeSlsMachineGroup`, etc.

### Schema Mapping Analysis
- **Alert Resource**: The legacy resource `alicloud_log_alert` has deprecated fields (`notification_list`, `dashboard`) that must be mapped to the new `sls.Alert` struct (likely via `Configuration` or compatible fields in `cws-lib-go`).
- `cws-lib-go`'s `AlertConfig` struct supports `NotificationList`, ensuring backward compatibility for legacy fields is possible.

## Decisions
1. **Scope**: Refactor `resource_alicloud_log_alert.go`, `data_source_alicloud_sls_logtail_config.go`, and `data_source_alicloud_sls_machine_groups.go`.
2. **Strategy**: Direct substitution of `client.WithLogClient` with `SlsService` methods.
3. **Compatibility**: Maintain all existing schema fields in `alicloud_log_alert`. Map deprecated fields to the underlying `sls.Alert` struct where supported.
4. **Validation**: Rely on `make build` availability.

## Next Steps
- Execute Phase 2: Refactor Data Sources (Low risk).
- Execute Phase 3: Refactor Resource (Medium risk due to schema mapping).
