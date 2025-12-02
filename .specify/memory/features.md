# Terraform Provider For Alibaba Cloud Feature Index

**Last Updated**: 2025-12-02
**Total Features**: 3

## Features

| ID | Name | Description | Status | Spec Path | Last Updated |
|----|------|-------------|--------|-----------|--------------|
| 001 | VPC Resources Update | 更新VPC相关资源以使用新的CWS-Lib-Go API | Draft | .specify/specs/001-cws-data-cws/spec.md | 2025-11-04 |
| 002 | Kafka Provider Refactoring | Kafka Provider重构 | Draft | .specify/specs/002-refactor-kafka-provider/spec.md | 2025-11-17 |
| 003 | Kafka Instance API Refactor | 重构Kafka实例API，将client.RpcPost调用替换为cws-lib-go API的方法 | Implemented | .specify/specs/003-refactor-kafka-instance/spec.md | 2025-12-02 |

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
