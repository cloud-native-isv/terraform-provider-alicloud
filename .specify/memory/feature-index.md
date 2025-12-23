# Terraform Provider Alicloud Feature Index

**Last Updated**: 2025-12-02
**Total Features**: 5

## Features

| ID | Name | Description | Status | Spec Path | Last Updated |
|----|------|-------------|--------|-----------|--------------|
| 001 | 基于 Go 语言的 Terraform Provider 实现 | 使用 Go 语言实现阿里云 Terraform Provider，提供基础设施即代码能力 | Implemented | .specify/specs/001-cws-data-cws/spec.md | 2025-12-02 |
| 002 | 分层架构设计 | 实现 Provider -> Resource/DataSource -> Service -> API 的四层架构设计 | Implemented | .specify/specs/002-refactor-kafka-provider/spec.md | 2025-12-02 |
| 003 | Kafka 实例管理 | 提供阿里云 Kafka 实例的全生命周期管理能力 | Implemented | .specify/specs/006-split-alikafka-instance/spec.md | 2025-12-23 |
| 004 | Makefile 构建系统 | 基于 Makefile 的自动化构建和开发工具链 | Implemented | (Not yet created) | 2025-12-02 |
| 005 | 强类型 API 调用 | 使用 cws-lib-go 库提供的强类型结构体进行 API 调用 | Implemented | (Not yet created) | 2025-12-02 |
| 006 | SelectDB Management | Provide lifecycle management for Alibaba Cloud SelectDB instances and clusters. | Implemented | .specify/specs/004-update-selectdb-resources/spec.md | 2025-12-14 |

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
