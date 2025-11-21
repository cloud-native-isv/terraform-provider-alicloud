# Implementation Plan: Refactor Kafka Instance API

**Branch**: `003-refactor-kafka-instance` | **Date**: 2025-11-21 | **Spec**: [link](./spec.md)
**Input**: Feature specification from `/.specify/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The primary requirement is to refactor the `alicloud_alikafka_instance` resource to replace direct `client.RpcPost` calls with `cws-lib-go` API methods. The technical approach involves creating or updating service layer functions to use the `cws-lib-go` Kafka API, ensuring all existing functionality is preserved.

## Technical Context

**Language/Version**: Go 1.18+  
**Primary Dependencies**: github.com/cloud-native-tools/cws-lib-go, github.com/aliyun/terraform-provider-alicloud  
**Storage**: N/A (API calls to Alibaba Cloud)  
**Testing**: Go test, Terraform acceptance tests  
**Target Platform**: Linux, macOS, Windows (Terraform provider compatible)  
**Project Type**: Single project (Terraform provider)  
**Performance Goals**: N/A (API latency dependent)  
**Constraints**: Must not change external behavior of the resource  
**Scale/Scope**: Single resource refactor, affects all Kafka instance operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Architecture Layering Principle**: ✅ The refactor will ensure Resource layer calls Service layer, which uses cws-lib-go API.
- **State Management Best Practices**: ✅ The refactor will maintain existing state management patterns.
- **Error Handling Standardization**: ✅ The refactor will use encapsulated error handling functions.
- **Code Quality and Consistency**: ✅ The refactor will follow naming conventions and schema definitions.
- **Strong Typing with CWS-Lib-Go**: ✅ The refactor will use strong types from cws-lib-go.
- **Testing and Validation Requirements**: ✅ Existing tests will validate the refactor.

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/003-refactor-kafka-instance/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
alicloud/
├── resource_alicloud_alikafka_instance.go  # Resource layer (to be refactored)
├── service_alicloud_alikafka.go            # Service layer (to be updated)
└── ...                                     # Other files

/cws_data/cws-lib-go/lib/cloud/aliyun/api/kafka/  # cws-lib-go Kafka API
```

**Structure Decision**: The refactor will modify existing files in the alicloud directory and utilize the existing cws-lib-go library.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |
