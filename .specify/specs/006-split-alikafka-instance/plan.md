# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `.specify/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Split the existing `alicloud_alikafka_instance` resource into two resources: `alicloud_alikafka_instance` for instance creation (ordering) and `alicloud_alikafka_deployment` for instance deployment (starting/stopping). This decouples the billing/creation lifecycle from the runtime lifecycle.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: `github.com/cloud-native-tools/cws-lib-go`
**Storage**: Terraform State
**Testing**: Terraform Acceptance Tests
**Target Platform**: Terraform Provider
**Project Type**: Terraform Provider Resource
**Performance Goals**: Standard Terraform resource performance
**Constraints**: Backward compatibility for existing `alicloud_alikafka_instance` users (breaking change accepted per spec, but migration path needed).
**Scale/Scope**: 2 Resources

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Library-First**: Yes, using `cws-lib-go` for API calls.
- **CLI Interface**: N/A
- **Test-First**: Acceptance tests will be updated/added.
- **Integration Testing**: Yes.
- **Observability**: Logging included.
- **Simplicity**: Decoupling simplifies the lifecycle management.

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/006-split-alikafka-instance/
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
├── resource_alicloud_alikafka_instance.go      # Modified
├── resource_alicloud_alikafka_deployment.go    # New
├── service_alicloud_alikafka.go                # Modified (Add StopInstance)
└── service_alicloud_alikafka_types.go          # Modified (Add StopInstanceRequest)
```

**Structure Decision**: Standard Terraform Provider structure.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |

