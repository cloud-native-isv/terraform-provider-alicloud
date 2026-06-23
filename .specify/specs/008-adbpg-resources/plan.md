# Implementation Plan: ADBPG Terraform Resources & Data Sources

**Branch**: `008-adbpg-resources` | **Date**: 2026-06-22 | **Spec**: [requirements.md](requirements.md)
**Input**: Specification from `.specify/specs/008-adbpg-resources/requirements.md`

## Summary

Implement 7 Terraform resources and 3 data sources for Alibaba Cloud AnalyticDB PostgreSQL (ADBPG) using the `alicloud_adbpg_*` naming convention. The implementation leverages the strongly-typed API layer in `pkg/cws-lib-go/lib/cloud/aliyun/api/adbpg/` and follows the strict Resource -> Service -> API -> SDK layered architecture. The service layer (`service_alicloud_adbpg.go`) creates an `AdbpgAPI` client via `cws-lib-go` credentials (same pattern as `KafkaService` and `FCService`) and exposes domain methods with `StateRefreshFunc` and `WaitFor*` patterns for async instance lifecycle operations.

## Technical Context

**Language/Version**: Go 1.26  
**Primary Dependencies**: `hashicorp/terraform-plugin-sdk v1.17.2`, `github.com/cloud-native-tools/cws-lib-go` (local submodule at `pkg/cws-lib-go`), `github.com/alibabacloud-go/gpdb-20160503/v5` (underlying SDK used by cws-lib-go)  
**Storage**: N/A (cloud API state managed by Terraform state file)  
**Testing**: `go test` with Terraform acceptance test framework (`TF_ACC=1`), `make testacc`  
**Target Platform**: Linux amd64, macOS amd64, macOS ARM64  
**Project Type**: Single Go module — Terraform provider plugin  
**Performance Goals**: Instance creation wait timeout up to 60 minutes (ADBPG clusters take 10-30 minutes to provision); polling interval 30 seconds for instance status, 5 seconds for sub-resource status  
**Constraints**: Must coexist with existing `gpdb` resources; must not modify any `gpdb` files; must pass `make build`, `make fmtcheck`, `make vet`  
**Scale/Scope**: 7 resources + 3 data sources + 1 service layer file = ~10 new Go source files + corresponding test files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance** (rendered from `.specify/memory/constitution.md`):

| # | Principle | Compliance | Evidence |
|---|-----------|------------|----------|
| I | Layered Architecture & Library-First Design | ✅ Pass | All resources call `AdbpgService` -> `AdbpgAPI` (cws-lib-go) -> SDK. No direct SDK calls in resource layer. See data-model.md § Architecture Layer Mapping. |
| II | Standardized Interfaces | ✅ Pass | All types use `cws-lib-go` strong types (`AdbpgInstanceCreate`, `AdbpgInstanceDetail`, etc.). No `map[string]interface{}` in new code. Pagination encapsulated in API layer (`ListInstances`). |
| III | Test-First Development | ✅ Pass | Acceptance tests planned for each resource and data source. Test files precede implementation in task ordering. |
| IV | Integration & Contract Testing | ✅ Pass | Acceptance tests exercise real API via `TF_ACC=1`. Service layer uses `WaitFor*` for state consistency. See contracts/ for API contracts. |
| V | Observability, Versioning & Simplicity | ✅ Pass | Standard Terraform `log.Printf` for debug/warn events. English error messages and API docs. |
| VI | Continuous Integration & Quality Gates | ✅ Pass | `make build` + `make fmtcheck` + `make vet` required. SC-005 and SC-006 in requirements enforce this. |
| VII | Feature-Centric Development | ✅ Pass | Bound to Feature 012 (ADBPG Resource Management). Feature index updated. |

**Gates Status**: ✅ All gates pass

**Re-check after Phase 1**: 2026-06-22 — All principles verified against data-model.md, contracts/, and quickstart.md. No violations.

## Project Structure

### Documentation (this spec)

```text
.specify/specs/008-adbpg-resources/
├── plan.md              # This file
├── requirements.md      # Feature specification
├── data-model.md        # Phase 1: entity model
├── quickstart.md        # Phase 1: example Terraform configurations
├── contracts/           # Phase 1: service layer API contracts
│   ├── adbpg-service.md # Service layer method contracts
│   └── adbpg-provider-registration.md # Provider resource registration
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

No standalone research.md — findings inlined below.

### Source Code (repository root)

```text
alicloud/                              # Provider source code
├── service_alicloud_adbpg.go          # Service layer: AdbpgService struct, NewAdbpgService, all CRUD + WaitFor methods
├── resource_alicloud_adbpg_instance.go        # Resource: alicloud_adbpg_instance
├── resource_alicloud_adbpg_account.go         # Resource: alicloud_adbpg_account
├── resource_alicloud_adbpg_database.go        # Resource: alicloud_adbpg_database
├── resource_alicloud_adbpg_connection.go      # Resource: alicloud_adbpg_connection
├── resource_alicloud_adbpg_security_ip_array.go  # Resource: alicloud_adbpg_security_ip_array
├── resource_alicloud_adbpg_backup_policy.go   # Resource: alicloud_adbpg_backup_policy
├── resource_alicloud_adbpg_ssl.go             # Resource: alicloud_adbpg_ssl
├── data_source_alicloud_adbpg_instances.go    # Data source: alicloud_adbpg_instances
├── data_source_alicloud_adbpg_accounts.go     # Data source: alicloud_adbpg_accounts
├── data_source_alicloud_adbpg_zones.go        # Data source: alicloud_adbpg_zones
└── provider.go                                # Registration of new resources/data sources (append to existing maps)
```

**Structure Decision**: Extends the existing provider package by adding 11 new Go files under `alicloud/` following the established naming convention (`resource_alicloud_<service>_<resource>.go`, `service_alicloud_<service>.go`). No new top-level directories. The service layer follows the same pattern as `service_alicloud_alikafka.go` (KafkaService) and `service_alicloud_fc_base.go` (FCService) for cws-lib-go integration.

## Complexity Tracking

N/A — No constitution violations.

## Phase 0: Research Review

### Internal Investigation Findings

1. **cws-lib-go API Layer**: Commits `aa4d774` and `0b9fa3a` provide complete ADBPG domain coverage across 6 API files:
   - `alicloud_adbpg_instance_api.go`: Instance CRUD, discovery (ListAvailableResources, ListRdsVpcs/VSwitchs), modification (description, maintain time, SSL), actions (restart, pause, resume, allocate/release public connection, clone, upgrade), information (SSL, net info, health, features, versions, roles, encryption keys)
   - `alicloud_adbpg_database_api.go`: Database CRUD, extension management, account CRUD, security IPs, parameter management, SQL execution
   - `alicloud_adbpg_datasource_api.go`: External/streaming data sources, streaming jobs, backups, backup policy, instance plans, rebalance status
   - `alicloud_adbpg_diagnostics_api.go`: Active/slow SQL, error logs, performance metrics, diagnosis records
   - `alicloud_adbpg_governance_api.go`: Tags, resource groups, Supabase projects
   - `alicloud_adbpg_vector_api.go`: Vector DB init, collections, documents, AI/model services, namespaces, secrets, knowledge base chat

2. **Service Layer Pattern**: Established by `KafkaService` (service_alicloud_alikafka.go) and `FCService` (service_alicloud_fc_base.go):
   - Struct holds `*connectivity.AliyunClient` + `*adbpg.AdbpgAPI`
   - Constructor `NewAdbpgService` converts AliyunClient credentials to `common.Credentials` and calls `adbpg.NewAdbpgAPI`
   - Methods delegate to API layer; no direct SDK calls

3. **State Management**: Existing `GpdbServiceV2` uses `StateRefreshFunc` with `field` parameter and `failStates`. The new `AdbpgService` will use the same pattern but with strongly-typed return values from `cws-lib-go` (no `map[string]interface{}`).

4. **Existing gpdb Resources**: 17 resource files + 6 data source files + 2 service files. All use the old `map[string]interface{}` pattern via `client.RpcPost()`. The new `adbpg` resources will NOT touch any `gpdb` files and will coexist alongside them.

5. **Provider Registration**: Resources and data sources are registered in `provider.go` via `ResourcesMap` and `DataSourcesMap` maps.
