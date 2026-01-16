# Feature Detail: Use CreateInstance API for Instance Resource

**Feature ID**: 010  
**Name**: Use CreateInstance API for Instance Resource  
**Description**: Switch the backend API of alicloud_instance resource create operation from RunInstances to CreateInstance.  
**Status**: Planned  
**Created**: 2026-01-15  
**Last Updated**: 2026-01-15

## Overview

Historically, the `alicloud_instance` resource used the `RunInstances` API. This feature aims to refactor the creation logic to use the `CreateInstance` API instead. This change is driven by alignment with specific API definitions provided (`CreateInstance.json`) while ensuring backward compatibility for the Terraform schema.

## Latest Review

Initial planning phase. Specification created.

## Key Changes

1. Refactor `resourceAliCloudInstanceCreate` to call `CreateInstance` API.
2. Implement polling logic to wait for instance status (Stopped -> Running) as `CreateInstance` does not start instances automatically.
3. Map existing schema attributes to `CreateInstance` parameters.
4. Update error handling and response processing.

## Implementation Notes

- `CreateInstance` API creates instances in `Stopped` status. Explicit `StartInstance` call is required.
- Backward compatibility for all existing schema arguments is mandatory.
- Must handle asynchronous state transitions robustly.

## Future Evolution Suggestions

- Re-evaluate if `RunInstances` features (like bulk creation) are needed in future and how they map to `CreateInstance` (which is single instance).

## Related Files

- Specification: .specify/specs/010-use-create-instance-api/spec.md
- Feature Index: memory/feature-index.md
- Feature Detail: memory/features/010-use-create-instance-api.md
- Quality Checklist: .specify/specs/010-use-create-instance-api/checklists/requirements.md

## Status Tracking

- **Draft**: Feature defined but spec incomplete.
- **Planned**: Spec approved, implementation pending.
- **Implemented**: Code changes merged.
- **Ready for Review**: PR submitted.
- **Completed**: Deployed and verified.
