# Implementation Plan - Refactor SLS Resources to Service Layer

## Problem
Legacy SLS resources and data sources use `github.com/aliyun/aliyun-log-go-sdk` directly via `client.WithLogClient` or raw `client.Do`. This bypasses the standardized `SlsService` layer and `cws-lib-go` strong types.

## Proposed Changes
Refactor the following files to use `alicloud/service_alicloud_sls_*.go` which leverages `cws-lib-go`:
1. `alicloud/data_source_alicloud_sls_machine_groups.go`
2. `alicloud/data_source_alicloud_sls_logtail_config.go`
3. `alicloud/resource_alicloud_log_alert.go`

## Verification Plan
### Automated Tests
- Run `make build` after each file modification to ensure type safety and compilation.
- (Optional) Run unit tests if available for these specific resources.

## Phased Execution

### Phase 1: Preparation
- [x] Identify legacy usage patterns.
- [x] Verify `SlsService` capabilities (List/Get/Create/Update).
- [x] Create Data Model mapping.

### Phase 2: Refactor Data Sources
- [ ] Refactor `alicloud/data_source_alicloud_sls_machine_groups.go`
  - Replace `client.WithLogClient` with `SlsService.ListSlsMachineGroups`.
  - Use `SlsService.DescribeSlsMachineGroup` for details.
  - Remove direct `aliyun-log-go-sdk` imports.
- [ ] Refactor `alicloud/data_source_alicloud_sls_logtail_config.go`
  - Replace `client.WithLogClient` with `SlsService.ListSlsLogtailConfigs`.
  - Use `SlsService.DescribeSlsLogtailConfig` for details.
  - Remove direct `aliyun-log-go-sdk` imports.

### Phase 3: Refactor Resource `alicloud_log_alert`
- [ ] Create `buildSlsAlertFromSchemaLegacy` helper (or adapt existing) to map legacy schema to `sls.Alert`.
- [ ] Refactor `Create`: Use `SlsService.CreateSlsAlert` + `WaitForSlsAlert`.
- [ ] Refactor `Read`: Use `SlsService.DescribeSlsAlert`.
- [ ] Refactor `Update`: Use `SlsService.UpdateSlsAlert` + `WaitForSlsAlert`.
- [ ] Refactor `Delete`: Use `SlsService.DeleteSlsAlert` + `WaitForSlsAlert`.
- [ ] Ensure backward compatibility for `notification_list` and other deprecated fields.

### Phase 4: Verification
- [ ] Run `make build` to verify compilation.
