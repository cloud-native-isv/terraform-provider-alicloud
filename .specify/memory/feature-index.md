# If your account belongs to domestic site Feature Index

**Last Updated**: 2026-01-26
**Total Features**: 4

## Features

| 001 | Implement Ecs Instance | Specification completed for feature 001 | Planned | .specify/specs/001-implement-ecs-instance/spec.md | 2026-01-22 |
| 002 | Fix Oss Force Destroy | Specification completed for feature 002 | Planned | .specify/specs/002-fix-oss-force-destroy/spec.md | 2026-01-23 |
| 003 | Oss Prune Bucket | Specification completed for feature 003 | Planned | .specify/specs/003-oss-prune-bucket/spec.md | 2026-01-23 |
| 004 | Fix Alikafka Resource | Fix AliKafka resource state persistence and API integration | Implemented | .specify/specs/004-fix-alikafka-resource/spec.md | 2026-01-26 |

## Feature Entry Format

Each feature entry should follow this format in the table:

| ID | Name | Description | Status | Feature Details | Last Updated |
|----|------|-------------|--------|----------------|--------------|
| 001 | Feature Name | Brief description of the feature | Draft | .specify/memory/features/001.md | 2025-11-21 |

### Column Definitions

| Column | Description |
|--------|-------------|
| ID | Sequential three-digit feature identifier (001, 002, etc.) |
| Name | Short feature name (2-4 words) describing the feature |
| Description | Brief summary of the feature's purpose and scope |
| Status | Current implementation status (Draft, Planned, Implemented, Ready for Review, Completed) |
| Feature Details | Path to feature detail file in .specify/memory/features/[FEATURE_ID].md |
| Last Updated | When the feature entry was last modified (YYYY-MM-DD format) |

## Template Usage Instructions

This template contains placeholder tokens in square brackets (e.g., `[PROJECT_NAME]`, `[FEATURE_COUNT]`). 
When generating the actual feature index:

1. Replace `[PROJECT_NAME]` with the actual project name
2. Replace `[LAST_UPDATED_DATE]` with current date in YYYY-MM-DD format
3. Replace `[FEATURE_COUNT]` with the actual number of features
4. Replace `[FEATURE_ENTRIES]` with the complete Markdown table containing all feature entries
5. Each individual feature entry should have its placeholders replaced accordingly:
   - `[FEATURE_ID]`: Sequential three-digit ID
   - `[FEATURE_NAME]`: Short descriptive name (2-4 words)
   - `[FEATURE_DESCRIPTION]`: Brief feature description
   - `[FEATURE_STATUS]`: Current status (Draft, Planned, etc.)
   - `[SPEC_PATH]`: Path to spec file or "(Not yet created)"
   - `[FEATURE_LAST_UPDATED]`: Feature-specific last updated date

Ensure all placeholder tokens are replaced before finalizing the feature index.
