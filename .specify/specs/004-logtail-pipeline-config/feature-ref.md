# Feature Reference: SLS Logtail Pipeline Config (Feature 010)

**Feature**: [.specify/memory/features/010.md](../../memory/features/010.md)  
**Index**: [.specify/memory/features.md](../../memory/features.md)  
**Spec**: [.specify/specs/004-logtail-pipeline-config/requirements.md](requirements.md)

## What this spec delivers

- 将 `alicloud_logtail_config` 的 Service 层从旧版 LogtailConfig API 迁移到新版 Logtail Pipeline Config API。
- 对资源 schema 进行不兼容重设计，使其更贴合新版“插件流水线”模型，并减少永久 diff。

## Why it belongs to Feature 010

Feature 010 的定义即为“基于新版 Logtail Pipeline Config 的资源与服务层升级（含 schema 重设计）”。本 spec 是该 Feature 的第一次实现性切片。

## Related code areas

- `alicloud/resource_alicloud_logtail_config.go`
- `alicloud/service_alicloud_sls_logtail_config.go`
- `pkg/cws-lib-go/lib/cloud/aliyun/api/sls/alicloud_sls_logtail_pipeline_api.go`
