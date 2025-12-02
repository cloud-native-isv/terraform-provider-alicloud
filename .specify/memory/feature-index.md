# [PROJECT_NAME] Feature Index

**Last Updated**: [LAST_UPDATED_DATE]
**Total Features**: [FEATURE_COUNT]

## Features

[FEATURE_ENTRIES]

## Feature Entry Format

Each feature entry should follow this format in the table:

| ID | Name | Description | Status | Spec Path | Last Updated |
|----|------|-------------|--------|-----------|--------------|
| [FEATURE_ID] | [FEATURE_NAME] | [FEATURE_DESCRIPTION] | [FEATURE_STATUS] | [SPEC_PATH] | [FEATURE_LAST_UPDATED] |

### Column Definitions

| Column | Description |
|--------|-------------|
| ID | Sequential three-digit feature identifier (001, 002, etc.) |
| Name | Short feature name (2-4 words) describing the feature |
| Description | Brief summary of the feature's purpose and scope |
| Status | Current implementation status (Draft, Planned, Implemented, Ready for Review, Completed) |
| Spec Path | Path to specification file or "(Not yet created)" if not yet created |
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
