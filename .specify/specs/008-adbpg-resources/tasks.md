# Tasks: ADBPG Terraform Resources & Data Sources

**Requirement ID**: 008
**Requirement Key**: 008-adbpg-resources
**Related Feature**: 012 ADBPG Resource Management
**Input**: Design documents from `.specify/specs/008-adbpg-resources/`
**Prerequisites**: plan.md (required), requirements.md (required), data-model.md, contracts/adbpg-service.md, contracts/adbpg-provider-registration.md, quickstart.md

**Tests Mode**: ON (Constitution Principle III "Test-First Development" — implementation MUST follow TDD; Principle IV "Integration & Contract Testing" — acceptance tests cover critical flows)

## Clarifications

### Session 2026-06-23

- Q: T036 says "Create and Update → configure backup policy via API" but cws-lib-go only has `DescribeBackupPolicy` — no `ModifyBackupPolicy` wrapper exists. Which API method should be used? → A: Add a prerequisite task (T045) before T035 to implement `ModifyBackupPolicy` in cws-lib-go and add `ModifyAdbpgBackupPolicy` service method, then reference it in T036. The underlying SDK supports ModifyBackupPolicy (existing gpdb resource uses it via raw RPC).
- Q: `account_description` is Optional (not ForceNew) in data-model E3, but no `ModifyAccountDescription` API exists in cws-lib-go. How should description updates be handled? → A: Add `ModifyAccountDescription` wrapper to cws-lib-go (type `AdbpgAccountModify` already exists) and add `ModifyAdbpgAccountDescription` service method. Update T019 to handle description changes.
- Q: SC-004 requires `terraform import` support for all resources, but no task explicitly covers adding `Importer` to resource schemas. Should this be tracked? → A: Add explicit note to each schema task (T012, T018, T021, T024, T027, T035, T038) requiring `Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough}` per SC-004.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Definition of Done (DoD)

- DoD-1: All 7 resources and 3 data sources are implemented per data-model.md and contracts/
- DoD-2: Service layer (`service_alicloud_adbpg.go`) implements all 31 methods from contracts/adbpg-service.md (29 original + ModifyAdbpgBackupPolicy + ModifyAdbpgAccountDescription)
- DoD-3: All resources and data sources are registered in `provider.go`
- DoD-4: Acceptance tests exist for every resource and data source (`resource_alicloud_adbpg_*_test.go`, `data_source_alicloud_adbpg_*_test.go`)
- DoD-5: No direct SDK calls in resource or data source files — all go through `AdbpgService`
- DoD-6: No `map[string]interface{}` used in any new code — all types from `cws-lib-go`
- DoD-7: `make build` succeeds on Linux amd64
- DoD-8: `make fmtcheck` and `make vet` pass with no warnings or errors

**DoD Status**: pass (deferred: T045, T046)

## Phase 1: Setup (Provider Registration)

**Purpose**: Register all new ADBPG resources and data sources in `provider.go` so they are recognized by Terraform.

- [X] T001 Register 7 new resources in `ResourcesMap` in `alicloud/provider.go`: `alicloud_adbpg_instance` → `resourceAliCloudAdbpgInstance()`, `alicloud_adbpg_account` → `resourceAliCloudAdbpgAccount()`, `alicloud_adbpg_database` → `resourceAliCloudAdbpgDatabase()`, `alicloud_adbpg_connection` → `resourceAliCloudAdbpgConnection()`, `alicloud_adbpg_security_ip_array` → `resourceAliCloudAdbpgSecurityIpArray()`, `alicloud_adbpg_backup_policy` → `resourceAliCloudAdbpgBackupPolicy()`, `alicloud_adbpg_ssl` → `resourceAliCloudAdbpgSsl()`
- [X] T002 Register 3 new data sources in `DataSourcesMap` in `alicloud/provider.go`: `alicloud_adbpg_instances` → `dataSourceAliCloudAdbpgInstances()`, `alicloud_adbpg_accounts` → `dataSourceAliCloudAdbpgAccounts()`, `alicloud_adbpg_zones` → `dataSourceAliCloudAdbpgZones()`

---

## Phase 2: Foundational (Service Layer — Blocking Prerequisite)

**Purpose**: Implement the `AdbpgService` struct and all service methods. Every resource and data source depends on this layer.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 Create `alicloud/service_alicloud_adbpg.go` with `AdbpgService` struct (fields: `client *connectivity.AliyunClient`, `adbpgAPI *adbpg.AdbpgAPI`) and constructor `NewAdbpgService(client *connectivity.AliyunClient) (*AdbpgService, error)` — converts credentials to `common.Credentials` per contracts/adbpg-service.md C-1 (follow `KafkaService`/`FCService` pattern from `service_alicloud_alikafka.go`)
- [X] T004 Implement instance methods in `alicloud/service_alicloud_adbpg.go`: `DescribeAdbpgInstance` (C-2), `CreateAdbpgInstance` (C-3), `DeleteAdbpgInstance` (C-4), `ModifyAdbpgInstanceDescription` (C-5), `ModifyAdbpgInstanceMaintainTime` (C-6)
- [X] T005 Implement instance state methods in `alicloud/service_alicloud_adbpg.go`: `AdbpgInstanceStateRefreshFunc` (C-7), `WaitForAdbpgInstanceRunning` (C-8 — pending: Creating/ClassChanging/NetAddressCreating/Restarting, target: Running, delay: 30s, minTimeout: 10s), `WaitForAdbpgInstanceDeleted` (C-9 — pending: Deleting, target: empty, delay: 30s)
- [X] T006 Implement account methods in `alicloud/service_alicloud_adbpg.go`: `ListAdbpgAccounts` (C-10), `DescribeAdbpgAccount` (C-11 — decodes composite ID `instanceId:accountName`), `CreateAdbpgAccount` (C-12), `DeleteAdbpgAccount` (C-13), `ResetAdbpgAccountPassword` (C-14)
- [X] T007 Implement database methods in `alicloud/service_alicloud_adbpg.go`: `CreateAdbpgDatabase` (C-15), `DeleteAdbpgDatabase` (C-16), `ListAdbpgDatabases` (C-17), `DescribeAdbpgDatabase` (C-18 — decodes composite ID `instanceId:dbName`)
- [X] T008 Implement connection methods in `alicloud/service_alicloud_adbpg.go`: `AllocateAdbpgPublicConnection` (C-19), `ReleaseAdbpgPublicConnection` (C-20), `DescribeAdbpgPublicConnection` (C-21 — filters for IPType == "Public")
- [X] T009 Implement security IP, backup, SSL, discovery, and tag methods in `alicloud/service_alicloud_adbpg.go`: `ModifyAdbpgSecurityIps` (C-22), `DescribeAdbpgBackupPolicy` (C-23), `DescribeAdbpgSSL` (C-24), `ModifyAdbpgSSL` (C-25), `ListAdbpgAvailableResources` (C-26), `ListAdbpgInstances` (C-27), `TagAdbpgResources` (C-28), `UntagAdbpgResources` (C-29)

**Checkpoint**: Service layer ready — all 31 methods implemented (29 via cws-lib-go + 2 RPC fallbacks). User story implementation can now begin.

---

## Phase 3: User Story 1 — Manage ADBPG Instance Lifecycle (Priority: P1) MVP

**Goal**: Full CRUD for `alicloud_adbpg_instance` — create, read, update (description, maintain time, tags, seg_node_num, db_instance_class), and delete with async state waiting.

**Independent Test**: `terraform plan/apply/destroy` with an instance resource. Verify instance creation with specified attributes, in-place update of description/maintenance window, tag management, and clean destroy.

### Tests for User Story 1 (MANDATORY)

- [X] T010 [US1] Write acceptance test in `alicloud/resource_alicloud_adbpg_instance_test.go`: `TestAccAlicloudAdbpgInstance_basic` — create instance, verify all attributes, modify description and maintain time, verify updates, destroy
- [X] T011 [US1] Write acceptance test in `alicloud/resource_alicloud_adbpg_instance_test.go`: `TestAccAlicloudAdbpgInstance_tags` — create instance with tags, add/remove/modify tags, verify diff and updates

### Implementation for User Story 1

- [X] T012 [US1] Implement `resourceAliCloudAdbpgInstance()` schema in `alicloud/resource_alicloud_adbpg_instance.go` — define all attributes per data-model.md E2 (engine_version, db_instance_class, db_instance_mode, zone_id, vpc_id, vswitch_id, pay_type, seg_node_num, storage_size, storage_type, description, security_ip_list, master_node_num, resource_group_id, serverless_mode, encryption_key, encryption_type, tags, maintain_start_time, maintain_end_time; computed: status, connection_string, port, creation_time, db_instance_id). Set ForceNew per data-model.md. Include `Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough}` (SC-004). Timeouts: Create 60m, Update 30m, Delete 30m.
- [X] T013 [US1] Implement `resourceAliCloudAdbpgInstanceCreate` in `alicloud/resource_alicloud_adbpg_instance.go` — build `AdbpgInstanceCreate` struct from schema, call `AdbpgService.CreateAdbpgInstance`, set ID, call `WaitForAdbpgInstanceRunning`, apply tags via `TagAdbpgResources` if specified
- [X] T014 [US1] Implement `resourceAliCloudAdbpgInstanceRead` in `alicloud/resource_alicloud_adbpg_instance.go` — call `AdbpgService.DescribeAdbpgInstance`, set all attributes in state. Handle `IsNotFoundError` → `d.SetId("")`. Report lock_mode if not "Unlock".
- [X] T015 [US1] Implement `resourceAliCloudAdbpgInstanceUpdate` in `alicloud/resource_alicloud_adbpg_instance.go` — handle `d.HasChange` for description (→ `ModifyAdbpgInstanceDescription`), maintain_start_time/maintain_end_time (→ `ModifyAdbpgInstanceMaintainTime`), tags (→ `TagAdbpgResources`/`UntagAdbpgResources`), seg_node_num/db_instance_class (→ wait for Running after change)
- [X] T016 [US1] Implement `resourceAliCloudAdbpgInstanceDelete` in `alicloud/resource_alicloud_adbpg_instance.go` — call `AdbpgService.DeleteAdbpgInstance`, then `WaitForAdbpgInstanceDeleted`

**Checkpoint**: `alicloud_adbpg_instance` is fully functional — can create, read, update, and delete instances.

---

## Phase 4: User Story 2 — Manage ADBPG Account Lifecycle (Priority: P1)

**Goal**: CRUD for `alicloud_adbpg_account` — create accounts with type/password, update password and description, delete.

**Independent Test**: Create an instance, then create/modify/delete account resources. Verify account appears in listing.

### Tests for User Story 2 (MANDATORY)

- [X] T017 [US2] Write acceptance test in `alicloud/resource_alicloud_adbpg_account_test.go`: `TestAccAlicloudAdbpgAccount_basic` — create Super account, verify attributes, update password, destroy

### Implementation for User Story 2

- [~] T046 [US2] Implement `ModifyAccountDescription` API wrapper <!-- deferred: submodule read-only — RPC fallback implemented in service layer --> in `pkg/cws-lib-go/lib/cloud/aliyun/api/adbpg/alicloud_adbpg_database_api.go` (type `AdbpgAccountModify` already exists) and add corresponding `ModifyAdbpgAccountDescription` service method in `alicloud/service_alicloud_adbpg.go` — prerequisite for T019 description update logic
- [X] T018 [US2] Implement `resourceAliCloudAdbpgAccount()` schema in `alicloud/resource_alicloud_adbpg_account.go` — attributes per data-model.md E3 (db_instance_id ForceNew, account_name ForceNew, account_password Sensitive, account_type ForceNew Optional default "Normal", account_description Optional; computed: status). ID format: `instanceId:accountName`. Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 5m, Update 5m, Delete 5m.
- [X] T019 [US2] Implement CRUD functions in `alicloud/resource_alicloud_adbpg_account.go`: Create (→ `AdbpgService.CreateAdbpgAccount`), Read (→ `DescribeAdbpgAccount`, handle NotFound), Update (password → `ResetAdbpgAccountPassword`, description → `ModifyAdbpgAccountDescription`), Delete (→ `DeleteAdbpgAccount`)

**Checkpoint**: `alicloud_adbpg_account` is fully functional.

---

## Phase 5: User Story 3 — Manage ADBPG Database (Priority: P2)

**Goal**: Create and delete databases within an ADBPG instance via Terraform.

**Independent Test**: Create an instance, then apply a database resource and verify it appears. Destroy to verify deletion.

### Tests for User Story 3 (MANDATORY)

- [X] T020 [US3] Write acceptance test in `alicloud/resource_alicloud_adbpg_database_test.go`: `TestAccAlicloudAdbpgDatabase_basic` — create database with UTF8, verify, destroy

### Implementation for User Story 3

- [X] T021 [US3] Implement `resourceAliCloudAdbpgDatabase()` schema in `alicloud/resource_alicloud_adbpg_database.go` — attributes per data-model.md E4 (db_instance_id ForceNew, db_name ForceNew, db_description ForceNew Optional, character_name ForceNew Optional). ID format: `instanceId:dbName`. All fields ForceNew (no update API). Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 5m, Delete 5m.
- [X] T022 [US3] Implement Create, Read, Delete functions in `alicloud/resource_alicloud_adbpg_database.go`: Create (→ `AdbpgService.CreateAdbpgDatabase`), Read (→ `DescribeAdbpgDatabase`, handle NotFound), Delete (→ `DeleteAdbpgDatabase`)

**Checkpoint**: `alicloud_adbpg_database` is fully functional.

---

## Phase 6: User Story 4 — Manage ADBPG Connection (Priority: P2)

**Goal**: Allocate and release public connection endpoints for ADBPG instances.

**Independent Test**: Create an instance, allocate a public connection, verify connection_string and port are populated, destroy to release.

### Tests for User Story 4 (MANDATORY)

- [X] T023 [US4] Write acceptance test in `alicloud/resource_alicloud_adbpg_connection_test.go`: `TestAccAlicloudAdbpgConnection_basic` — allocate public connection, verify connection_string/port, destroy

### Implementation for User Story 4

- [X] T024 [US4] Implement `resourceAliCloudAdbpgConnection()` schema in `alicloud/resource_alicloud_adbpg_connection.go` — attributes per data-model.md E5 (db_instance_id ForceNew, connection_string_prefix ForceNew; computed: connection_string, ip_address, port). ID format: `instanceId:prefix`. Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 10m, Delete 10m.
- [X] T025 [US4] Implement Create, Read, Delete functions in `alicloud/resource_alicloud_adbpg_connection.go`: Create (→ `AllocateAdbpgPublicConnection`, wait for Running), Read (→ `DescribeAdbpgPublicConnection`, handle NotFound), Delete (→ `ReleaseAdbpgPublicConnection`)

**Checkpoint**: `alicloud_adbpg_connection` is fully functional.

---

## Phase 7: User Story 5 — Manage ADBPG Security IP Whitelist (Priority: P2)

**Goal**: Manage named IP whitelist groups for ADBPG instances.

**Independent Test**: Create an instance, apply a security IP array, verify IPs are set, modify and verify update.

### Tests for User Story 5 (MANDATORY)

- [X] T026 [US5] Write acceptance test in `alicloud/resource_alicloud_adbpg_security_ip_array_test.go`: `TestAccAlicloudAdbpgSecurityIpArray_basic` — create IP array, verify, modify IP list, verify update, destroy

### Implementation for User Story 5

- [X] T027 [US5] Implement `resourceAliCloudAdbpgSecurityIpArray()` schema in `alicloud/resource_alicloud_adbpg_security_ip_array.go` — attributes per data-model.md E6 (db_instance_id ForceNew, db_instance_ip_array_name ForceNew Optional default "default", security_ip_list Required). ID format: `instanceId:arrayName`. Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 5m, Update 5m, Delete 5m.
- [X] T028 [US5] Implement CRUD functions in `alicloud/resource_alicloud_adbpg_security_ip_array.go`: Create and Update both → `ModifyAdbpgSecurityIps`. Read → `DescribeAdbpgInstance` and inspect SecurityIPList. Delete → `ModifyAdbpgSecurityIps` with "127.0.0.1".

**Checkpoint**: `alicloud_adbpg_security_ip_array` is fully functional.

---

## Phase 8: User Story 6 — Query ADBPG Instances Data Source (Priority: P2)

**Goal**: Data source to query existing ADBPG instances by filters (IDs, description_regex, status, tags, resource_group_id).

**Independent Test**: Create instances, query via data source with filters, verify matching results.

### Tests for User Story 6 (MANDATORY)

- [X] T029 [US6] Write acceptance test in `alicloud/data_source_alicloud_adbpg_instances_test.go`: `TestAccAlicloudAdbpgInstancesDataSource_basic` — create instance, query by description_regex, verify result attributes

### Implementation for User Story 6

- [X] T030 [US6] Implement `dataSourceAliCloudAdbpgInstances()` schema in `alicloud/data_source_alicloud_adbpg_instances.go` — filters per data-model.md DS1 (ids, description_regex, status, resource_group_id, tags, output_file). Output: list of instances with all `AdbpgInstance` attributes.
- [X] T031 [US6] Implement `dataSourceAliCloudAdbpgInstancesRead` in `alicloud/data_source_alicloud_adbpg_instances.go` — call `AdbpgService.ListAdbpgInstances`, apply client-side filters (description_regex, ids), set results in state, write to output_file if specified

**Checkpoint**: `data.alicloud_adbpg_instances` is fully functional.

---

## Phase 9: User Story 7 — Query ADBPG Accounts Data Source (Priority: P3)

**Goal**: Data source to list accounts on an ADBPG instance.

**Independent Test**: Create accounts on an instance, query via data source, verify results.

### Tests for User Story 7 (MANDATORY)

- [X] T032 [US7] Write acceptance test in `alicloud/data_source_alicloud_adbpg_accounts_test.go`: `TestAccAlicloudAdbpgAccountsDataSource_basic` — create account, query data source, verify attributes

### Implementation for User Story 7

- [X] T033 [US7] Implement `dataSourceAliCloudAdbpgAccounts()` schema and read function in `alicloud/data_source_alicloud_adbpg_accounts.go` — filters per data-model.md DS2 (db_instance_id Required, name_regex, output_file). Output: list of accounts with account_name, account_type, account_status, account_description.

**Checkpoint**: `data.alicloud_adbpg_accounts` is fully functional.

---

## Phase 10: User Story 8 — Manage ADBPG Backup Policy (Priority: P3)

**Goal**: Configure backup retention, schedule, and recovery point settings for an ADBPG instance.

**Independent Test**: Create an instance, apply a backup policy, verify settings, modify and verify update.

### Tests for User Story 8 (MANDATORY)

- [X] T034 [US8] Write acceptance test in `alicloud/resource_alicloud_adbpg_backup_policy_test.go`: `TestAccAlicloudAdbpgBackupPolicy_basic` — configure backup policy, verify, modify retention, verify update, destroy

### Implementation for User Story 8

- [~] T045 [US8] Implement `ModifyBackupPolicy` API wrapper <!-- deferred: submodule read-only — RPC fallback implemented in service layer --> in `pkg/cws-lib-go/lib/cloud/aliyun/api/adbpg/alicloud_adbpg_datasource_api.go` and add corresponding `ModifyAdbpgBackupPolicy` service method in `alicloud/service_alicloud_adbpg.go` — prerequisite for T035/T036 (cws-lib-go currently only has `DescribeBackupPolicy`; underlying SDK supports `ModifyBackupPolicy` via RPC)
- [X] T035 [US8] Implement `resourceAliCloudAdbpgBackupPolicy()` schema in `alicloud/resource_alicloud_adbpg_backup_policy.go` — attributes per data-model.md E7 (db_instance_id ForceNew, backup_retention_period Optional Computed, preferred_backup_period Optional Computed, preferred_backup_time Optional Computed, enable_recovery_point Optional Computed). ID: `db_instance_id`. Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 5m, Update 5m, Delete 5m.
- [X] T036 [US8] Implement CRUD functions in `alicloud/resource_alicloud_adbpg_backup_policy.go`: Create and Update → `AdbpgService.ModifyAdbpgBackupPolicy`. Read → `DescribeAdbpgBackupPolicy`. Delete → reset to defaults via `ModifyAdbpgBackupPolicy`.

**Checkpoint**: `alicloud_adbpg_backup_policy` is fully functional.

---

## Phase 11: Additional Resources — SSL & Zones

**Purpose**: Implement remaining resources and data sources not directly covered by a primary user story.

### ADBPG SSL Resource

- [X] T037 [P] Write acceptance test in `alicloud/resource_alicloud_adbpg_ssl_test.go`: `TestAccAlicloudAdbpgSsl_basic` — enable SSL, verify, disable, verify
- [X] T038 [P] Implement `resourceAliCloudAdbpgSsl()` schema and CRUD in `alicloud/resource_alicloud_adbpg_ssl.go` — attributes per data-model.md E8 (db_instance_id ForceNew, ssl_enabled Required; computed: ssl_expired). ID: `db_instance_id`. Include `Importer` with `ImportStatePassthrough` (SC-004). Timeouts: Create 10m, Update 10m, Delete 10m. Create/Update → `ModifyAdbpgSSL`. Read → `DescribeAdbpgSSL`. Delete → `ModifyAdbpgSSL(false)`.

### ADBPG Zones Data Source

- [X] T039 [P] Write acceptance test in `alicloud/data_source_alicloud_adbpg_zones_test.go`: `TestAccAlicloudAdbpgZonesDataSource_basic` — query zones, verify non-empty result
- [X] T040 [P] Implement `dataSourceAliCloudAdbpgZones()` schema and read function in `alicloud/data_source_alicloud_adbpg_zones.go` — filters per data-model.md DS3 (multi, output_file). Output: list of zones with supported engine versions and instance classes. Calls `AdbpgService.ListAdbpgAvailableResources`.

**Checkpoint**: All 7 resources and 3 data sources are implemented.

---

## Phase 12: Polish & Cross-Cutting Concerns

**Purpose**: Build verification, code quality, and final validation.

- [X] T041 Run `make build` to verify compilation succeeds across all platforms in `alicloud/`
- [X] T042 Run `make fmtcheck` and `make vet` to verify code formatting and static analysis pass
- [X] T043 Verify no `map[string]interface{}` usage in any new `alicloud/*adbpg*.go` files (grep check)
- [X] T044 Verify no direct SDK calls in resource/data source files — all calls go through `AdbpgService` (grep check)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 (provider registration is forward declarations). BLOCKS all user stories.
- **User Stories (Phases 3–10)**: All depend on Phase 2 (service layer) completion
  - User stories can proceed sequentially in priority order (P1 → P2 → P3)
  - Within the same priority, stories can run in parallel (e.g., US3/US4/US5/US6 are all P2)
- **Additional (Phase 11)**: Depends on Phase 2. Can run in parallel with any user story.
- **Polish (Phase 12)**: Depends on ALL prior phases being complete

### User Story Dependencies

- **US1 (P1 — Instance)**: Depends only on Phase 2. No dependencies on other stories. **MVP target.**
- **US2 (P1 — Account)**: Depends on Phase 2. Acceptance tests require a running instance (created within test). Independent of US1 implementation.
- **US3 (P2 — Database)**: Depends on Phase 2. Tests require a running instance. Independent of US1/US2.
- **US4 (P2 — Connection)**: Depends on Phase 2. Tests require a running instance. Independent of other stories.
- **US5 (P2 — Security IP)**: Depends on Phase 2. Tests require a running instance. Independent of other stories.
- **US6 (P2 — Instances DS)**: Depends on Phase 2. Tests require running instances. Independent of other stories.
- **US7 (P3 — Accounts DS)**: Depends on Phase 2. Tests require running instance + accounts. Independent of US2 (creates own test accounts).
- **US8 (P3 — Backup Policy)**: Depends on Phase 2. Tests require a running instance. Independent of other stories.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Schema definition before CRUD functions
- Create before Read/Update/Delete
- Story complete before moving to next priority

### Parallel Opportunities

- T001 and T002 (Phase 1) can run in parallel
- T004–T009 (Phase 2 method groups) can be parallelized if service file is split, but since all write to the same file, they run sequentially
- All P2 user stories (Phases 5–8) can run in parallel after Phase 2 completes
- T037/T038 and T039/T040 (Phase 11) can run in parallel with each other and with any user story
- Phase 12 tasks T041–T044 can run in parallel with each other

---

## Parallel Example: P2 User Stories

```bash
# After Phase 2 completes, launch all P2 stories in parallel:
# Developer A: Phase 5 (US3 — Database)
# Developer B: Phase 6 (US4 — Connection)
# Developer C: Phase 7 (US5 — Security IP)
# Developer D: Phase 8 (US6 — Instances Data Source)
# Each story touches different files, no conflicts.
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (provider registration)
2. Complete Phase 2: Foundational (service layer — all 29 methods)
3. Complete Phase 3: User Story 1 (alicloud_adbpg_instance)
4. **STOP and VALIDATE**: Test instance CRUD independently with `TF_ACC=1`
5. Instance resource is the minimum viable product

### Incremental Delivery

1. Phase 1 + Phase 2 → Service layer ready
2. Phase 3 (US1 — Instance) → Test independently → MVP
3. Phase 4 (US2 — Account) → Test independently → Core account management
4. Phases 5–8 (US3–US6) → Test independently → Full P2 coverage
5. Phases 9–10 (US7–US8) → Test independently → Full P3 coverage
6. Phase 11 (SSL + Zones) → Test independently → Complete resource set
7. Phase 12 → Final validation → Ship ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable (each test creates its own instance fixture)
- All acceptance tests use `TF_ACC=1` and run via `make testacc`
- New files coexist with existing `gpdb` files — no `gpdb` files are modified
- Service layer file may exceed 1000 lines; consider splitting into `service_alicloud_adbpg_instance.go` and `service_alicloud_adbpg_sub.go` if needed (per Constitution "split files exceeding 1000 lines")
