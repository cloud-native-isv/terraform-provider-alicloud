# Quickstart: 004-logtail-pipeline-config

**Date**: 2026-02-12  
**Branch**: 004-logtail-pipeline-config

## Goal

快速验证 `alicloud_logtail_config` 资源已切换到新版 Logtail Pipeline Config 能力，并且 schema 映射与 JSON 规范化不会产生永久 diff。

## Preconditions

- Go 1.20+
- 可用的阿里云凭证与区域配置（参考项目 README 中的 Acceptance Testing 部分）
- 已存在 SLS Project（以及业务上需要的相关资源）

## Local Build & Unit Tests

1. 在仓库根目录执行：
   - `make test`
2. （可选）只跑与本资源映射相关的单元测试：
   - `go test ./alicloud -run TestAccAlicloudLogtailConfig -timeout=120m`（若已新增测试）

## Acceptance Test (Terraform)

1. 设置必要环境变量（示例见 README）：
   - `TF_ACC=1`
   - `ALICLOUD_ACCESS_KEY` / `ALICLOUD_SECRET_KEY` / `ALICLOUD_REGION` 等
2. 运行验收测试：
   - `make testacc`（或按项目约定只运行单个测试用例）

## Manual Terraform Check (Smoke)

- 创建资源后重复执行两次 `terraform plan`，应无差异。
- 更新任意可更新字段后 `apply`，应能读回一致状态。
- Import 后再次 `plan`，应无差异。
