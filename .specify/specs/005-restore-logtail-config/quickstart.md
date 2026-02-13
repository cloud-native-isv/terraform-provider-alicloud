# Quickstart: 005-restore-logtail-config

**Date**: 2026-02-13  
**Branch**: 005-restore-logtail-config

## Goal

验证以下目标：
1. 存量 `alicloud_logtail_config` 升级后不被阻断。
2. 新资源 `alicloud_logtail_pipeline_config` 可独立完成生命周期。
3. 新旧资源并存时遵循“最后一次成功写入生效”。

## Preconditions

- Go 1.20+
- Terraform 与 Provider 本地可编译
- 已配置阿里云凭证与 SLS 相关环境

## Step 1: Build & Basic Tests

- 执行 `make`
- 执行 `make test`

预期：编译与基础单元测试通过。

## Step 2: Legacy Compatibility Smoke

- 使用仅包含 `alicloud_logtail_config` 的历史配置执行：
  - `terraform init`
  - `terraform plan`
  - `terraform apply`（如测试环境允许）

预期：
- 不出现“未知资源类型 `alicloud_logtail_config`”
- read/update 行为与旧版本保持连续

## Step 3: Pipeline Resource Lifecycle

- 新建仅使用 `alicloud_logtail_pipeline_config` 的最小配置
- 执行 create/read/update/delete 流程

预期：生命周期动作全部成功。

## Step 4: Coexistence Scenario

- 在同一 workspace 中声明：
  - 一个 `alicloud_logtail_config`
  - 一个 `alicloud_logtail_pipeline_config`
- 两者指向同一远端对象，并分别执行更新

预期：
- 两个资源类型可独立识别与执行
- 最终远端对象状态与最后一次成功写入一致

## Step 5: Documentation Checks

核对发布说明与资源文档是否包含：
- 旧资源弃用状态（仅文档提示）
- 兼容窗口（至少 2 个小版本）
- 新资源替代入口与迁移建议

预期：上述三项全部具备。
