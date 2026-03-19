# Feature Index

This index tracks all functional and non-functional features managed within the project. It serves as the central directory for specifications, plans, and implementation status.

**Total Features**: 11

| ID | Name | Description | Status | Feature Details | Spec Path | Last Updated |
|---|---|---|---|---|---|---|
| 001 | ECS Instance Management | ECS 实例与其生命周期管理 (create, read, update, delete) | Implemented | [Details](features/001.md) | - | 2026-02-06 |
| 002 | OSS Bucket Management | OSS 存储桶生命周期与清理策略管理 | Implemented | [Details](features/002.md) | - | 2026-02-06 |
| 003 | Go Development Environment | 统一 Go 1.20+ 与工具链标准 | Implemented | [Details](features/003.md) | - | 2026-02-06 |
| 004 | Layered Architecture | Resource -> Service -> API -> SDK 严格分层架构 | Implemented | [Details](features/004.md) | - | 2026-02-06 |
| 005 | CWS-Lib-Go Integration | 统一 API 调用封装与强类型接口 (cws-lib-go) | Implemented | [Details](features/005.md) | - | 2026-02-06 |
| 006 | Local SDK Management | 使用 sdk/ 目录固定关键 SDK 版本 | Implemented | [Details](features/006.md) | - | 2026-02-06 |
| 007 | Automated Testing Suite | 单元测试与验收测试覆盖关键流程 | Implemented | [Details](features/007.md) | - | 2026-02-06 |
| 008 | Strong Typing Constraints | 禁止新增弱类型 map 作为请求/响应载体 | Implemented | [Details](features/008.md) | - | 2026-02-06 |
| 009 | AliKafka Resource Layering | AliKafka 相关资源的分层架构与生命周期操作一致性 | Implemented | [Details](features/009.md) | .specify/specs/003-alikafka-instance-billing/requirements.md | 2026-02-10 |
| 010 | SLS Logtail Pipeline Config | Logtail 新旧资源并存与命名边界治理（旧版兼容 + 新版独立命名） | Implemented | [Details](features/010.md) | .specify/specs/005-restore-logtail-config/requirements.md | 2026-02-13 |
| 011 | SLS Log Prefix Unification | 日志服务 resource/data source 命名统一到 alicloud_log_* 前缀并提供迁移治理 | Ready for Review | [Details](features/011.md) | .specify/specs/006-unify-log-prefix/requirements.md | 2026-03-19 |
