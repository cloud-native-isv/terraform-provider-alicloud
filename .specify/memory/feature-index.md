# Terraform Provider Alicloud Feature Index

**Last Updated**: 2026-01-23
**Total Features**: 8

## Features

| ID | Name | Description | Status | Feature Details | Last Updated |
|----|------|-------------|--------|-----------------|--------------|
| 001 | Go Development Environment | Standardization on Go 1.20+ and related toolchain for provider development. | Implemented | .specify/memory/features/001.md | 2026-01-22 |
| 002 | Layered Architecture Provider | Strict separation of concerns (Resource -> Service -> API -> SDK) for maintainability. | Implemented | .specify/memory/features/002.md | 2026-01-22 |
| 003 | CWS-Lib-Go Integration | Standardization of API interactions via the shared cws-lib-go wrapper library. | Implemented | .specify/memory/features/003.md | 2026-01-22 |
| 004 | Local SDK Vendor Management | Management of local SDK copies in `sdk/` directory for specific service versions. | Implemented | .specify/memory/features/004.md | 2026-01-22 |
| 005 | Automated Testing Suite | Comprehensive unit and acceptance testing framework (testacc) application. | Implemented | .specify/memory/features/005.md | 2026-01-22 |
| 006 | Strong Typing Enforcement | Mandate strict usage of defined structs over weakly typed maps. | Implemented | .specify/memory/features/006.md | 2026-01-22 |
| 007 | Implement alicloud_ecs_instance Resource | Implementation of new resource `alicloud_ecs_instance` using `CreateInstance` API. | Planned | .specify/memory/features/007.md | 2026-01-22 |
| 008 | OSS Bucket Management | Consolidated management of OSS Bucket resources and configurations, including lifecycle and destruction policies. | Implemented | .specify/memory/features/008.md | 2026-01-23 |
