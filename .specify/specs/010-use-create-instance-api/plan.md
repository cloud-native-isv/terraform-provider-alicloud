# Implementation Plan: Use CreateInstance API for Instance Resource

**Branch**: `010-use-create-instance-api` | **Date**: 2026-01-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `.specify/specs/010-use-create-instance-api/spec.md`

## Summary

Refactor the `alicloud_instance` resource creation to use the `CreateInstance` API instead of the legacy `RunInstances`. This involves updating the backend logic to handle the `Stopped` state returned by `CreateInstance`, strictly implementing the Service Layer pattern, and orchestrating subsequent actions like starting the instance and attaching multiple security groups.

## Technical Context

**Language/Version**: Go 1.20+
**Primary Dependencies**: 
- `github.com/aliyun/alibaba-cloud-sdk-go/services/ecs` (SDK)
- `github.com/hashicorp/terraform-plugin-sdk` (Terraform SDK)
**Project Type**: Terraform Provider (Single)
**Architecture**: Provider -> Resource -> Service -> API (SDK)
**Legacy API**: `RunInstances` (Synchronous-like/Atomic start)
**Target API**: `CreateInstance` (Asynchronous/Stopped state)
**Key Challenges**:
- `CreateInstance` does not start the instance automatically.
- `CreateInstance` supports only one security group at creation.
- `CreateInstance` does not allocate public IP addresses.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Architecture Layering**: ✅ Resource layer will call `EcsService` methods (`CreateInstance`, `StartInstance`, `JoinSecurityGroups`).
- **State Management**: ✅ Will implement `WaitForEcsInstance` for transitions (Stopped -> Running).
- **Strong Typing**: ✅ Will use typed `ecs.CreateInstanceRequest` and `ecs.StartInstanceRequest` from SDK.
- **Error Handling**: ✅ Will use standardized error wrapping and checking from `alicloud/errors.go`.
- **Code Quality**: ✅ ID handling and validation will follow existing patterns.
- **Testing**: ✅ Existing acceptance tests must pass.

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/010-use-create-instance-api/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (Schema <-> API Mapping)
├── quickstart.md        # Phase 1 output (Migration Guide/Notes)
└── tasks.md             # Phase 2 output
```

### Source Code

```text
alicloud/
├── resource_alicloud_instance.go   # MODIFY: Update Create method
├── service_alicloud_ecs.go         # MODIFY: Add CreateInstance, StartInstance methods
└── errors.go                       # REFERENCE: Error handling helpers
```

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |
