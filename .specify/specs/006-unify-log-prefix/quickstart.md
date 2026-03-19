# Quickstart: 006-unify-log-prefix

## 目标

在一次升级窗口内，将 Terraform 配置中的 `alicloud_sls_*` 全量替换为 `alicloud_log_*`，并通过销毁重建完成迁移验证。

## 前置条件

- 已升级到包含统一策略的 provider 版本。
- 已获得目标环境的阿里云访问凭证。
- 已备份 Terraform state 与关键资源清单。

## Step 1: 识别历史命名

在代码仓库扫描历史前缀：

```bash
grep -R "alicloud_sls_" -n .
```

将结果与命名映射清单逐项对应，形成迁移列表。

## Step 2: 替换为统一命名

- 按映射关系将 `alicloud_sls_*` 替换为 `alicloud_log_*`。
- 对每个替换项记录：旧名称、替代名称、涉及模块、负责人。

## Step 3: 执行销毁重建

按模块分批执行（建议先低风险环境）：

1. 移除旧资源块或改名后执行 `terraform plan`。
2. 根据计划确认需要销毁重建的对象。
3. 执行 `terraform apply` 完成替换。

> 本需求不提供 `terraform state mv` 官方承诺，统一采用销毁重建策略。

## Step 4: 验证与回归

- 执行 `terraform plan`，确认无残留 `alicloud_sls_*`。
- 关键业务路径做一次端到端验证。
- 核对 provider 报错信息：若误用旧前缀，应包含明确替代项提示。

## Step 5: 发布治理同步

- 在发布说明中明确“本次直接生效”的停用规则。
- 更新治理清单状态（`pending -> unified` 或 `retired_with_note`）。
- 将迁移结果写入团队变更记录。

## 常见风险与建议

- **风险**: 销毁重建导致资源标识变化。
  - **建议**: 先在预发环境演练并做回滚方案。
- **风险**: 文档、映射、代码替换不同步。
  - **建议**: 以单一映射清单为准，发布前做一致性审计。
