# Implementation Plan: OSS Bucket Prune Before Delete

**Branch**: `003-oss-prune-bucket` | **Date**: 2026-01-23 | **Spec**: [spec.md](spec.md)
**Input**: Specification from `.specify/specs/003-oss-prune-bucket/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command.

## Summary

Implement OSS bucket deletion flow so that when `force_destroy` is enabled, the provider
prunes the bucket using the new `PruneBucket` API, retries on retryable errors, and waits
for cleanup completion (or timeout) before deleting the bucket. This replaces the previous
`EmptyBucket` flow and ensures cleanup is idempotent, consistent, and safe.

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**: terraform-plugin-sdk v1.17.x, cws-lib-go OSS API, aliyun-oss-go-sdk (indirect)  
**Storage**: N/A (remote OSS service)  
**Testing**: `make test`, `make testacc`, `go test ./alicloud`  
**Target Platform**: Terraform provider plugin (Linux/macOS/Windows)  
**Project Type**: single (Go module)  
**Performance Goals**: Cleanup of up to 100,000 objects within 10 minutes under normal conditions  
**Constraints**: Must use Service layer and CWS-Lib-Go API, avoid weak typing, wait via Service-layer state handling, validate with `make`  
**Scale/Scope**: OSS bucket delete flow only; no new resources or external endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Architecture Layering**: Resource/DataSource → Service → API → SDK enforced
- **State Management**: WaitFor/StateRefreshFunc used; Create avoids direct Read
- **Error Handling**: WrapError patterns and standard helpers used
- **Strong Typing**: CWS-Lib-Go types preferred; no new weak typing
- **Testing & Validation**: `make` run required; timeouts configured
- **Feature-Centric Development**: Feature list reviewed/updated for this spec

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/003-oss-prune-bucket/
├── plan.md              # This file (/speckit.plan command output)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this spec. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
alicloud/
├── resource_alicloud_oss_bucket.go
├── service_alicloud_oss_bucket.go
└── service_alicloud_oss_object.go

pkg/cws-lib-go/lib/cloud/aliyun/api/oss/

docs/
```

**Structure Decision**: Use the existing Terraform provider layout under `alicloud/` for
resource/service logic and rely on `pkg/cws-lib-go/.../oss` for API integration.

## Complexity Tracking

No Constitution violations identified.
