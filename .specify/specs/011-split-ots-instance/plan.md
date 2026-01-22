# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `.specify/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature splits the existing `alicloud_ots_instance` resource into two separate resources: `alicloud_ots_instance` (for SSD and HYBRID types) and `alicloud_ots_instance_vcu` (for VCU type only). This separation clarifies usage, enforces strict parameter validation, and removes VCU-specific logic from the legacy resource, opting for a "hard break" approach where VCU support is completely removed from the legacy resource, requiring manual state migration for existing users.

## Technical Context

**Language/Version**: Go 1.22 (Current project version)
**Primary Dependencies**: 
- `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore` (for API calls)
- `github.com/hashicorp/terraform-plugin-sdk/helper/schema` (for resource schema)
- `github.com/hashicorp/terraform-plugin-sdk/helper/resource` (for state refresh/retry)
**Storage**: Terraform State
**Testing**: Terraform Acceptance Tests (`TestAcc...`)
**Target Platform**: Terraform Provider (Linux/macOS/Windows)
**Project Type**: Terraform Provider Resource
**Performance Goals**: Standard Terraform provider performance (CRUD operations within API timeout limits)
**Constraints**: 
- **Hard Break**: VCU support removed from `alicloud_ots_instance`.
- **Schema Cleanup**: VCU fields removed from `alicloud_ots_instance`.
- **New Resource**: `alicloud_ots_instance_vcu` handles VCU logic exclusively.
**Scale/Scope**: 2 Resource definitions, ~500 LOC refactoring/new code.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Spec-Driven**: ✅ Feature flow is Spec -> Plan -> Task -> Implement
- **Agent-First**: ✅ Artifacts are structured for AI consumption
- **Library/CLI-First**: ✅ N/A (Provider architecture dictates library style)
- **Test-First**: ✅ TDD flow required, Acceptance tests must be updated/added
- **Context Preservation**: ✅ Feature index and specs updated
- **Architecture Layering**: ✅ Will use Service layer + CWS-Lib-Go API
- **Strong Typing**: ✅ Will use `tablestoreAPI` structs from CWS-Lib-Go
- **State Management**: ✅ Will use `StateRefreshFunc` and `WaitFor` patterns
- **Error Handling**: ✅ Will use `alicloud/errors.go` and `WrapError`

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/011-split-ots-instance/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
alicloud/
├── resource_alicloud_ots_instance.go         # Modified: Remove VCU logic/fields
├── resource_alicloud_ots_instance_vcu.go     # New: VCU implementation
├── service_alicloud_ots.go                   # Modified: Ensure service methods support separation if needed
└── extension_utils.go (or similar)           # Register new resource
```

**Structure Decision**: Standard Terraform Provider structure. New resource file for VCU instance to separate concerns physically.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
