# Implementation Plan: Fix AliKafka Resource Implementation

**Branch**: `004-fix-alikafka-resource` | **Date**: 2026-01-26 | **Spec**: [spec.md](spec.md)
**Input**: Specification from `.specify/specs/004-fix-alikafka-resource/spec.md`

## Summary

This plan addresses critical issues in the `alicloud_alikafka_instance` and `alicloud_alikafka_deployment` resources. It involves modernizing the implementation to use the `CWS-Lib-Go` API wrapper, ensuring comprehensive state persistence for all computed fields, and effectively managing the complex lifecycle states (Order -> Deployment -> Running) of AliKafka instances.

## Technical Context

**Language/Version**: Go 1.20+
**Primary Dependencies**: 
- `github.com/cloud-native-tools/cws-lib-go` (v2 API wrapper)
- `github.com/hashicorp/terraform-plugin-sdk` (v1/v2)
- `github.com/aliyun/alibaba-cloud-sdk-go` (Legacy SDK, being replaced)
**Storage**: Terraform State (local or remote backend)
**Testing**: Go testing (`go test`, `make` for build validation)
**Target Platform**: Terraform Provider (Binaries for Linux/Darwin/Windows)
**Project Type**: Terraform Provider (Single Go module)
**Performance Goals**: Create < 60 mins, Update < 120 mins (cloud dependent)
**Constraints**: Must maintain backward compatibility for existing resources. Must handle "Order" vs "Instance" duality transparently.
**Scale/Scope**: Refactoring 2 resources (`alikafka_instance`, `alikafka_deployment`) and 1 service (`service_alicloud_alikafka.go`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Spec-Driven**: Spec created (`spec.md` in `004`), Plan created.
- **Agent-First**: Plan uses markdown, clearly structured.
- **Library/CLI-First**: Leveraging `CWS-Lib-Go` library as mandated.
- **Test-First**: Validation steps included in tasks (T041, T042).
- **Context Preservation**: Feature memory updated.
- **Architecture Layering**: Strictly enforcing Service Layer pattern.
- **State Management**: Using `WaitFor` patterns preventing Create->Read direct calls.
- **Strong Typing**: Using `kafka.KafkaInstance` struct from CWS-Lib-Go.
- **Feature-Centric**: Working within Feature 004 context.

**Gates Status**: [✅ All gates pass]

## Project Structure

### Documentation (this spec)

```text
.specify/specs/004-fix-alikafka-resource/
├── plan.md              # This file
├── spec.md              # Feature specification
├── tasks.md             # Task breakdown
├── data-model.md        # Entity definitions
├── quickstart.md        # Usage guide
└── contracts/           # API/Schema contracts
    └── resource_schema.md
```

### Source Code

```text
alicloud/
├── service_alicloud_alikafka.go          # Service layer (Updated)
├── resource_alicloud_alikafka_instance.go   # Resource layer (Refactored)
└── resource_alicloud_alikafka_deployment.go # Lifecycle resource (Refactored)
pkg/cws-lib-go/lib/cloud/aliyun/api/kafka/
├── alicloud_kafka_instance_types.go      # API Types (Reference)
└── alicloud_kafka_instance_api.go        # API Methods (Reference)
```

## Phases

### Phase 1: Foundation (Service Layer)

1.  **Enhance CWS-Lib-Go**: Ensure `KafkaInstance` struct supports all required fields (Tags, PaidType, etc.) and API methods (`UpgradeInstance`).
2.  **Service Layer Update**: Implement `NewKafkaService` and wrapper methods in `service_alicloud_alikafka.go`.
3.  **State Management**: Implement `WaitFor` logic for Creating, Updating, Stopping states.

### Phase 2: Implementation (Resource Layer)

1.  **State Persistence**: Refactor `Read` method in `resource_alicloud_alikafka_instance.go` to map ALL API fields to schema.
2.  **API Integration**: Replace legacy SDK calls with `kafkaService` methods (CreateInstance, UpgradeInstance).
3.  **Lifecycle Management**: Ensure `Create` waits for "Running" state. Ensure `Update` waits for completion.
4.  **Deployment Resource**: Align `deployment` resource to use shared service logic.

### Phase 3: Documentation & Polish

1.  Update resource documentation (`alicloud_alikafka_instance.html.markdown`).
2.  Create comprehensive example (`examples/alicloud_alikafka/`).
3.  Verify compilation and basic functionality.
