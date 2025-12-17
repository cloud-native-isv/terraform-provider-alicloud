# Implementation Plan: Implement KafkaService methods using cws-lib-go

**Branch**: `005-implement-kafka-service` | **Date**: 2025-12-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `.specify/specs/005-implement-kafka-service/spec.md`

## Summary

Update `KafkaService` in `alicloud/service_alicloud_alikafka.go` to implement missing methods (`CreatePostPayOrder`, `CreatePrePayOrder`, `StartInstance`, `ModifyInstanceName`, `UpgradeInstanceVersion`) using the `cws-lib-go` library. Refactor the service methods to use strong types instead of `map[string]interface{}` to comply with the project constitution.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: `github.com/cloud-native-tools/cws-lib-go` (aliyun/api/kafka)
**Project Type**: Terraform Provider
**Constraints**: Must maintain backward compatibility where possible, but refactor internal service signatures to strong types as per Constitution.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Library-First**: Yes, using `cws-lib-go`.
- **CLI Interface**: N/A (Provider).
- **Test-First**: Will verify with `make`.
- **Integration Testing**: N/A for this refactor (relying on existing tests/manual verification).
- **Observability**: Using `WrapError`.
- **Simplicity**: Direct mapping to API.
- **Strong Typing**: **YES**, this is the main driver. Replacing `map[string]interface{}` with structs.

**Gates Status**: [✅ All gates pass]

## Project Structure

### Documentation (this feature)

```text
.specify/specs/005-implement-kafka-service/
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
├── service_alicloud_alikafka.go        # Service implementation (Modified)
├── service_alicloud_alikafka_types.go  # New file for request structs
└── resource_alicloud_alikafka_instance.go # Resource implementation (Modified)
```

**Structure Decision**: Add a new types file to keep `service_alicloud_alikafka.go` clean and define strong types for requests.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |
