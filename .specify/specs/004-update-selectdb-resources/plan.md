# Implementation Plan: Update SelectDB Resources

**Branch**: `004-update-selectdb-resources` | **Date**: 2025-12-14 | **Spec**: [.specify/specs/004-update-selectdb-resources/spec.md](spec.md)
**Input**: Feature specification from `.specify/specs/004-update-selectdb-resources/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Update `alicloud_selectdb_instance` and `alicloud_selectdb_cluster` resources and their corresponding data sources to support the latest Alibaba Cloud SelectDB API fields defined in `cws-lib-go`. This includes adding support for new configuration options, computed fields, and ensuring full coverage of the API capabilities.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.22+
**Primary Dependencies**: `cws-lib-go` (SelectDB API), Terraform Plugin SDK v2
**Storage**: Terraform State
**Testing**: Go testing framework, Terraform acceptance tests
**Target Platform**: Terraform Provider Alicloud
**Project Type**: Terraform Provider Resource/DataSource
**Performance Goals**: Standard Terraform provider performance
**Constraints**: Must use `cws-lib-go` for API calls, must follow provider architecture.
**Scale/Scope**: Update 2 resources and 2 data sources.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Library-First**: N/A (Provider implementation)
- **CLI Interface**: N/A (Provider implementation)
- **Test-First**: Acceptance tests will be updated/added.
- **Integration Testing**: Acceptance tests cover integration with Alibaba Cloud.
- **Observability**: Logging is standard in provider.
- **Simplicity**: Following standard provider patterns.
- **Architecture Layering**: Adhering to Provider -> Resource -> Service -> API.
- **State Management**: Using `StateRefreshFunc` and `WaitFor*`.
- **Error Handling**: Using `alicloud/errors.go`.
- **Strong Typing**: Using `cws-lib-go` types.

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/004-update-selectdb-resources/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
alicloud/
├── resource_alicloud_selectdb_instance.go
├── resource_alicloud_selectdb_cluster.go
├── data_source_alicloud_selectdb_instances.go
└── data_source_alicloud_selectdb_clusters.go
```

**Structure Decision**: Updating existing files in `alicloud/` package.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | - | - |
