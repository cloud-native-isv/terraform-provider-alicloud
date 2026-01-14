# Terraform Provider Alicloud Feature Index

**Last Updated**: 2026-01-12
**Total Features**: 8

## Features

| ID | Name | Description | Status | Spec Path | Last Updated |
|----|------|-------------|--------|-----------|--------------|
| 001 | 基于 Go 语言的 Terraform Provider 实现 | 使用 Go 语言实现阿里云 Terraform Provider，提供基础设施即代码能力 | Implemented | .specify/specs/001-cws-data-cws/spec.md | 2025-12-02 |
| 002 | 分层架构设计 | 实现 Provider -> Resource/DataSource -> Service -> API 的四层架构设计 | Implemented | .specify/specs/002-refactor-kafka-provider/spec.md | 2025-12-02 |
| 003 | Kafka 实例管理 | 提供阿里云 Kafka 实例的全生命周期管理能力 | Implemented | .specify/specs/006-split-alikafka-instance/spec.md | 2025-12-23 |
| 004 | Makefile 构建系统 | 基于 Makefile 的自动化构建和开发工具链 | Implemented | (Not yet created) | 2025-12-02 |
| 005 | 强类型 API 调用 | 使用 cws-lib-go 库提供的强类型结构体进行 API 调用 | Completed | .specify/specs/008-refactor-sls-resource-api/spec.md | 2026-01-12 |
| 006 | SelectDB Management | Provide lifecycle management for Alibaba Cloud SelectDB instances and clusters. | Implemented | .specify/specs/004-update-selectdb-resources/spec.md | 2025-12-14 |
| 007 | Tablestore VCU Instance Support | Support configuration of Tablestore instances using the VCU sizing model and elastic limits. | Completed | .specify/specs/007-ots-vcu-support/spec.md | 2026-01-07 |
| 009 | SLS LogStore Shard Management | Enhance alicloud_log_store resource to support dynamic shard count updates (Split/Merge). | Planned | .specify/specs/009-sls-logstore-shard-update/spec.md | 2026-01-14 |
