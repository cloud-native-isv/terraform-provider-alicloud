# Implementation Plan: Fix OSS Bucket Force Destroy

**Branch**: `002-fix-oss-force-destroy` | **Date**: 2026-01-23 | **Spec**: [spec.md](spec.md)
**Input**: Specification from `.specify/specs/002-fix-oss-force-destroy/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The goal is to fix the `force_destroy` behavior in `alicloud_oss_bucket`. Currently, it attempts to delete the bucket and only cleans up if checking for specific error codes, which proves unreliable. The new approach will actively clean up the bucket (objects, versions, delete markers, multipart uploads) *before* attempting deletion when `force_destroy` is enabled, using a robust `EmptyBucket` implementation in the Service layer.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: `github.com/aliyun/aliyun-oss-go-sdk/oss`
**Storage**: Alibaba Cloud OSS
**Testing**: Terraform Plugin SDK (Acceptance Tests)
**Target Platform**: Terraform Provider (Linux/Darwin/Windows)
**Project Type**: Terraform Provider
**Performance Goals**: N/A
**Constraints**: Must handle large number of objects (pagination), must handle versioning and delete markers.
**Scale/Scope**: Single resource modification.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Spec-Driven**: Spec flow is Spec -> Plan -> Task -> Implement
- **Agent-First**: Artifacts are structured for both Human and AI consumption
- **Library/CLI-First**: Feature implemented as reusable library with CLI (Service Layer pattern)
- **Test-First**: TDD flow followed, tests written before code
- **Context Preservation**: Update logs and decision records maintained

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/002-fix-oss-force-destroy/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
├── feature-ref.md       # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
alicloud/
├── resource_alicloud_oss_bucket.go  # Resource definition to be updated
└── service_alicloud_oss_bucket.go   # Service layer with EmptyBucket logic
```
