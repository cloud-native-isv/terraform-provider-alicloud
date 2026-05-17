# Implementation Plan: CMS Provider 封装能力改造

**Branch**: `007-refactor-cms-api` | **Date**: 2026-05-15 | **Spec**: [requirements.md](requirements.md)
**Input**: Specification from `.specify/specs/007-refactor-cms-api/requirements.md`

## Summary

本计划覆盖用户脚本枚举出的 51 个 CMS Provider 文件，将 CMS 资源、数据源与服务层统一收敛到 `Resource/DataSource -> Service -> CWS-Lib-Go CMS API -> Alibaba Cloud SDK` 调用链。当前范围内 50 个文件仅含 `package alicloud` 占位，`data_source_alicloud_cms_service.go` 已有实现但直接调用 `client.RpcPost` 且存在重复 `package alicloud` 语法风险；计划将其迁移到 CMS 服务层，并以服务能力矩阵判定每个 CMS 对象为“已启用改造、无需改造占位、待启用补齐”。

## Technical Context

**Language/Version**: Go 1.24 as declared in `go.mod`; project documentation still states Go 1.20+ as minimum build guidance.  
**Primary Dependencies**: Terraform Plugin SDK v1.17.2; `github.com/cloud-native-tools/cws-lib-go/lib` and `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api` replaced to local `pkg/cws-lib-go`; Alibaba Cloud CMS SDK through the CWS-Lib-Go CMS wrapper.  
**Storage**: N/A. Provider manages remote Alibaba Cloud CMS state and Terraform state only.  
**Testing**: Go tests via `make test`; targeted package tests via `go test ./alicloud`; CWS-Lib-Go CMS package tests under `pkg/cws-lib-go/lib/cloud/aliyun/api/cms`; acceptance tests via `TF_ACC=1 go test ./alicloud -run=TestAccAlicloud...` when credentials are available.  
**Target Platform**: Terraform Provider plugin for Linux/macOS/Windows; current workspace OS is Linux.  
**Project Type**: Single Go provider repository with main implementation under `alicloud/` and local API wrapper under `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/`.  
**Performance Goals**: CMS list/query operations return complete results without exposing pagination to resource/data source callers; no regression in Terraform plan/apply/read latency beyond CMS API behavior.  
**Constraints**: Maintain user-visible Terraform schema/resource/data source compatibility; no direct SDK/RPC calls in CMS resource/data source layers; no new `map[string]interface{}` as primary request/response carrier in Provider service code; pagination and retry/not-found normalization belong in API or service layer; documents in Chinese, code comments/logs/errors in English.  
**Scale/Scope**: 17 CMS object groups × 3 file kinds = 51 files, with 17 CWS-Lib-Go CMS API object areas available locally (`addon`, `agg_task_group`, `alert`, `cloud_resource`, `context`, `context_store`, `dataset`, `delivery_task`, `entity_store`, `integration_policy`, `memory`, `memory_store`, `pipeline`, `prometheus`, `service`, `umodel`, `workspace`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **I. Layered Architecture & Library-First Design**: PASS. Plan mandates `Resource/DataSource -> Service -> API (CWS-Lib-Go) -> SDK`; direct `RpcPost` in `data_source_alicloud_cms_service.go` is explicitly tracked as a migration target.
- **II. Standardized Interfaces**: PASS. Service methods must use typed CWS-Lib-Go CMS structs where available; pagination remains in API/service layer; legacy dynamic query payloads are isolated to API wrapper methods that already expose dynamic CMS query semantics.
- **III. Test-First Development**: PASS. Tasks must add/adjust targeted tests before each object migration, including service conversion tests and provider read/CRUD compatibility checks.
- **IV. Integration & Contract Testing**: PASS. Existing CWS-Lib-Go CMS API tests are part of the baseline; Provider service wrappers must preserve error, not-found, wait and status behavior.
- **V. Observability, Versioning & Simplicity**: PASS. Existing Terraform logging/error wrapping patterns remain; no new user-visible resource/data source is introduced without explicit enablement evidence.
- **VI. Continuous Integration & Quality Gates**: PASS. Required validation includes `gofmt/goimports`, targeted `go test`, and `make` where practical.
- **VII. Feature-Centric Development**: PASS. The plan remains bound to Feature 005 / CWS-Lib-Go Integration and cross-references Feature 004, 007 and 008; no new Feature is introduced or merged.

**Gates Status**: ✅ All gates pass. No unresolved clarification remains; absence of a prior `research.md` was resolved by repository documentation, feature memory and source/API inspection during planning.

## Project Structure

### Documentation (this spec)

```text
.specify/specs/007-refactor-cms-api/
├── requirements.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── feature-ref.md
└── contracts/
    └── cms-provider-contract.yaml
```

### Source Code (repository root)

```text
alicloud/
├── service_alicloud_cms_*.go          # Provider CMS service layer, one object group per file
├── resource_alicloud_cms_*.go         # Terraform resource definitions and CRUD orchestration
├── data_source_alicloud_cms_*.go      # Terraform data source definitions and read orchestration
├── provider.go                        # Resource/data source registration, only for enabled user-visible capabilities
└── *_test.go                          # Provider unit/acceptance tests

pkg/cws-lib-go/lib/cloud/aliyun/api/cms/
├── alicloud_cms_*_api.go              # Local CMS API wrapper methods and pagination handling
├── alicloud_cms_*_types.go            # Typed CMS API request/response domain types
├── alicloud_cms_*_utils.go            # SDK conversion helpers
└── alicloud_cms_*_test.go             # API wrapper contract/unit tests

docs/
├── development_guide.md               # Authoritative engineering constraints
└── wait_for_state.md                  # WaitFor/StateRefreshFunc conventions
```

**Structure Decision**: Use the existing single Go provider layout. The plan does not create a new module or package; it fills current CMS placeholder files and migrates the one active CMS service data source through Provider service methods backed by local CWS-Lib-Go CMS API wrappers.

## Phase 0: Research Review & Context

### Findings

- User input did not provide additional `$ARGUMENTS`; scope is therefore exactly the requirement specification and the original shell enumeration.
- `research.md` did not exist before planning; repository docs and source inspection were sufficient to resolve technical context.
- `README.md` and `.specify/memory/features.md` identify CWS-Lib-Go Integration as an existing non-functional Feature, already bound to this spec path.
- `docs/development_guide.md` and the Constitution require strict layering, strong typing, service-level pagination and service-level wait/state functions.
- All 51 enumerated CMS files exist. 50 files are one-line placeholders; `alicloud/data_source_alicloud_cms_service.go` is implemented but bypasses service/API layering with `client.RpcPost` and must be migrated.
- Local CWS-Lib-Go CMS API wrappers exist for all object groups in scope, with tests and typed files. Some wrapper methods still expose dynamic CMS query payloads (`map[string]any`) where the remote API semantics are inherently dynamic; Provider-facing service methods should avoid adding new untyped request/response surfaces unless documented as compatibility exceptions.

### Decisions

1. Treat placeholder resource/data source files as “scope-controlled but not user-visible” until provider registration and schema/tests are deliberately added.
2. Start migration with enabled/current behavior (`alicloud_cms_service` data source) to remove direct RPC usage and fix syntax hygiene before enabling broader CMS objects.
3. For each CMS object group, create a service-layer capability matrix before adding user-visible resource/data source code.
4. Preserve existing Terraform schema names, IDs, import behavior and status semantics for any already enabled CMS capability.
5. Prefer Provider service methods that wrap CWS-Lib-Go typed structs and return full result slices; resource/data source code must not implement pagination.

## Phase 1: Design & Contracts

Design artifacts generated by this plan:

- [data-model.md](data-model.md): CMS Provider file, object, service capability and coverage evidence entities.
- [contracts/cms-provider-contract.yaml](contracts/cms-provider-contract.yaml): Provider-internal contract for scope audit, service capability and Terraform operation flows.
- [quickstart.md](quickstart.md): Planning and validation workflow for implementing the CMS migration.
- [feature-ref.md](feature-ref.md): Feature review record confirming Feature 005 remains the binding feature and no Feature split/merge is required.

## Post-Design Constitution Check

**Gates Status**: ✅ All gates still pass after Phase 1 design.

- Data model encodes the required scope matrix and evidence trail, satisfying Feature-Centric Development and traceability.
- Contract file formalizes the resource/data source to service to API boundary and prevents direct SDK/RPC bypass.
- Quickstart validation commands align with the repository quality gates and test-first expectations.
- No justified complexity exception is required.

## Complexity Tracking

N/A. No Constitution violation or extra architecture layer is introduced.
