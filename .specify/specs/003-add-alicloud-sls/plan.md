# Implementation Plan: Add Terraform resource alicloud_sls_consumer_group

**Branch**: `003-add-alicloud-sls` | **Date**: 2025-10-28 | **Spec**: /.specify/specs/003-add-alicloud-sls/spec.md  
**Input**: Feature specification from `/.specify/specs/003-add-alicloud-sls/spec.md`

**Note**: This plan is produced by the /speckit.plan workflow and aligns with the Constitution.

## Summary

Add a new Terraform Resource `alicloud_sls_consumer_group` to declaratively manage SLS consumer groups. Immutable identifiers (project, logstore, consumer_group) are ForceNew; behavioral parameters (timeout, order) are updatable in-place. Import/ID encoding uses `project:logstore:consumer_group`. Create is idempotent: if already exists, adopt and converge timeout/order to HCL. Implementation follows provider layering: Resource/DataSource → Service → CWS-Lib-Go API → SDK. Proper state refresh/wait, retry/error handling and timeouts per Constitution.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: hashicorp/terraform-plugin-sdk v1.17, cloud-native-tools/cws-lib-go (aliyun/api), standard Go stdlib  
**Storage**: N/A (remote SLS service)  
**Testing**: go test; terraform-plugin-sdk acceptance tests (env-dependent); compile validation via `make`  
**Target Platform**: Linux server (Terraform provider plugin)  
**Project Type**: Single provider repository (Terraform Provider in Go)  
**Performance Goals**: CRUD within configured timeout windows; import/plan produce no unexpected drift  
**Constraints**: Must comply with Constitution on layering, state mgmt, error handling, naming (Id not ID), and timeouts; ensure retry on throttling/system busy; no direct SDK calls in Resource layer  
**Scale/Scope**: Single resource with standard CRUD and import; no large data volumes expected

## Constitution Check

Gate assessment against /.specify/memory/constitution.md:

- Architecture Layering: PASS — Resource will call Service; Service uses CWS-Lib-Go API wrappers.
- State Management: PASS — Create/Delete use WaitFor*; StateRefreshFunc implemented; no direct Read in Create.
- Error Handling: PASS — Use WrapError/WrapErrorf, IsNotFoundError/IsAlreadyExistError/NeedRetry; retry common transient errors.
- Code Quality/Naming: PASS — Resource name `alicloud_sls_consumer_group`; Id naming convention; Schema with Description and validation.
- Testing/Validation: PASS — Plan includes `make` build check; artifacts for quickstart and contracts; file sizes kept < 1000 lines.

Re-check after Phase 1: expected PASS.

## Project Structure

### Documentation (this feature)

```
.specify/specs/003-add-alicloud-sls/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (later by /speckit.tasks)
```

### Source Code (repository root)

```
alicloud/
├── resource_alicloud_sls_consumer_group.go   # Resource layer (new)
└── service_alicloud_sls_consumer_group.go    # Service layer (new)

docs/
└── r/
    └── sls_consumer_group.md                 # Resource docs (new)

examples/
└── sls/consumer_group/
    └── main.tf                               # Usage example (new)

tests/
└── resource_alicloud_sls_consumer_group_test.go  # (optional acceptance tests)
```

**Structure Decision**: Extend existing provider layout by adding one Resource file and one Service file under `alicloud/`, with accompanying docs and example for user adoption.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| None | — | — |
