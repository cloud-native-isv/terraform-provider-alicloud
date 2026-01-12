# Tasks: Refactor SLS Resources to Service Layer

## Phase 1: Setup
- [ ] T001 Verify project compilation environment by running `make build` locally.

## Phase 2: Foundational
- [ ] T002 Verify `SlsService` capabilities in `alicloud/service_alicloud_sls_*.go` match requirements (done in research phase, just a check).

## Phase 3: User Story 1 - Maintain SLS Resource Behavior (Data Sources)
**Story Goal**: Refactor data sources to use `SlsService` while maintaining existing behavior.
**Independent Test**: `make build` passes; data sources compile and return correct types logic.

- [ ] T003 [P] [US1] Refactor `alicloud/data_source_alicloud_sls_machine_groups.go`: Replace `client.WithLogClient` with `SlsService.ListSlsMachineGroups` and `SlsService.DescribeSlsMachineGroup`. Remove `aliyun-log-go-sdk` direct imports.
- [ ] T004 [P] [US1] Refactor `alicloud/data_source_alicloud_sls_logtail_config.go`: Replace `client.WithLogClient` with `SlsService.ListSlsLogtailConfigs` and `SlsService.DescribeSlsLogtailConfig`. Remove `aliyun-log-go-sdk` direct imports.
- [ ] T005 [US1] Verify compilation of data sources.

## Phase 4: User Story 2 - Use Strong Types End-to-End (`log_alert` Resource)
**Story Goal**: Refactor `alicloud_log_alert` to use strong types via `SlsService`, handling complex schema mapping.
**Independent Test**: `make build` passes; resource CRUD operations compile and mapped fields align with `cws-lib-go` structs.

- [ ] T006 [US2] Create helper function `buildSlsAlertFromSchemaLegacy` in `alicloud/resource_alicloud_log_alert.go` (or adapt existing) to map Terraform schema (including deprecated fields) to `sls.Alert` struct from `cws-lib-go`.
- [ ] T007 [US2] Refactor `Create` method in `alicloud/resource_alicloud_log_alert.go` to use `SlsService.CreateSlsAlert` and `SlsService.WaitForSlsAlert`.
- [ ] T008 [US2] Refactor `Read` method in `alicloud/resource_alicloud_log_alert.go` to use `SlsService.DescribeSlsAlert`. Ensure response map handles deprecated fields correctly.
- [ ] T009 [US2] Refactor `Update` method in `alicloud/resource_alicloud_log_alert.go` to use `SlsService.UpdateSlsAlert` and `SlsService.WaitForSlsAlert`.
- [ ] T010 [US2] Refactor `Delete` method in `alicloud/resource_alicloud_log_alert.go` to use `SlsService.DeleteSlsAlert` and `SlsService.WaitForSlsAlert`.
- [ ] T011 [US2] Verify backward compatibility for deprecated fields (`notification_list`, `dashboard`) logic in `alicloud/resource_alicloud_log_alert.go`.
- [ ] T012 [US2] Verify compilation of `alicloud_log_alert` resource.

## Phase 5: Polish & Cross-Cutting Concerns
- [ ] T013 Check for and remove any remaining unused `aliyun-log-go-sdk` imports in modified files.
- [ ] T014 Run final `make build` to ensure project stability.

## Dependencies
- US1 (Data Sources) and US2 (Resources) can arguably be done in parallel, but US2 depends on `SlsService` being reliable (which US1 helps validate).
- Within US2, T007-T010 depend on T006 (Helper).

## Parallel Execution Examples
- T003 and T004 can be executed in parallel by different agents/developers.
- T008 (Read) and T010 (Delete) in US2 can be implemented in parallel once T006 is defined.

## Implementation Strategy
- MVP: Complete US1 (Data Sources) first as a low-risk validation of the pattern.
- Then tackle US2 (`log_alert`), starting with the mapping helper (T006), then `Read` (T008), then `Create/Update/Delete`.
