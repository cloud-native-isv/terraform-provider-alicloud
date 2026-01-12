<!--
  SOURCE TEMPLATE (development path): templates/feature-template.md
  INSTALLED TEMPLATE (runtime path): .specify/templates/feature-template.md
  Do NOT remove placeholder tokens. Each [TOKEN] must be replaced during feature instantiation.
  This template is derived from an actual feature detail file and generalized.
-->

# Feature Detail: Refactor SLS Resource API Calls

**Feature ID**: 008  
**Name**: Refactor SLS Resource API Calls  
**Description**: Refactor legacy SLS resources to use SlsService and cws-lib-go.  
**Status**: Completed  
**Created**: 2026-01-12  
**Last Updated**: 2026-01-12

## Overview

This feature addresses the technical debt in legacy SLS resources (`alicloud_log_alert`, `alicloud_sls_machine_groups`, etc.) by migrating them from direct `aliyun-log-go-sdk` usage to the standardized `SlsService` using `cws-lib-go` strong types. This ensures better maintainability, type safety, and consistency across the provider.

## Latest Review

- **Problem & Goal**: Legacy resources used raw `client.Do` or direct SDK usage, bypassing the service layer.
- **Outcome**: Successfully refactored key resources to use `SlsService`. Preserved V1 backward compatibility while enabling V2 features. All compilation checks passed.
- **Constraints**: Some fields in `cws-lib-go` were missing (e.g., `CompressType`), which were handled gracefully by omitting them from the mapped output.

## Key Changes

1. Replaced `client.WithLogClient` in `alicloud/data_source_alicloud_sls_machine_groups.go` with `SlsService` calls.
2. Replaced `client.WithLogClient` in `alicloud/data_source_alicloud_sls_logtail_config.go` with `SlsService` calls.
3. Refactored `alicloud/resource_alicloud_log_alert.go` to use `SlsService` for V2 alerts, with substantial CRUD logic updates.
4. Added `alicloud/resource_alicloud_sls_alert_helper.go` to handle complex schema mapping logic.

## Implementation Notes

- **Legacy V1 Compatibility**: V1 Alert logic is preserved in `resource_alicloud_log_alert.go` to ensure backward compatibility for old Terraform configurations invoking legacy arguments.
- **Library Limitations**: The `LogtailConfigOutputDetail` struct in `cws-lib-go` lacks `CompressType`; this field is currently ignored in the data source output.
- **Type Adjustments**: `MachineIdType` in `cws-lib-go` vs `MachineID` in legacy SDK required explicit casing conversion/handling.

## Future Evolution Suggestions

- **Dependency Update**: Update `cws-lib-go` to include `CompressType` if user demand arises.
- **Deprecation**: Deprecate V1 Alert logic in a future major version in favor of the strict V2 schema.
- **Testing**: Add more acceptance tests covering V2 specific configurations.

## Related Files

- Specification: .specify/specs/008-refactor-sls-resource-api/spec.md
- Plan: .specify/specs/008-refactor-sls-resource-api/plan.md
- Tasks: .specify/specs/008-refactor-sls-resource-api/tasks.md
- Review: .specify/specs/008-refactor-sls-resource-api/review.md
- Feature Index: .specify/memory/feature-index.md

## Status Tracking

- **Draft**: Feature defined but implementation pending.
- **Planned**: Plan and tasks created.
- **Implemented**: Code changes applied and verified locally.
- **Ready for Review**: Artifacts generated and ready for human/agent review.
- **Completed**: Review finalized and merged.
