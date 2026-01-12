# LogStore 资源升级指南

本指南详细说明了 `alicloud_log_store` 资源的变更以及如何升级您的 Terraform 代码。

## 变更摘要

针对 LogStore 的存储分层功能，我们对 `alicloud_log_store` 资源进行了以下更新：
1.  **废弃并移除了 `retention_period` 字段**：统一使用 `ttl` 字段配置数据保存时间。
2.  **调整了默认值**：更新了 `ttl`、`hot_ttl` 和 `infrequent_access_ttl` 的默认值以适应通用的分层存储策略。

## 详细变更

### 1. 字段替换：`retention_period` -> `ttl`

`retention_period` 字段已被移除。请在您的 `.tf` 文件中将其替换为 `ttl`。

*   **旧配置**：
    ```hcl
    resource "alicloud_log_store" "example" {
      project_name      = "my-project"
      logstore_name     = "my-logstore"
      retention_period  = 30
      # ...
    }
    ```

*   **新配置**：
    ```hcl
    resource "alicloud_log_store" "example" {
      project_name      = "my-project"
      logstore_name     = "my-logstore"
      ttl               = 30
      # ...
    }
    ```

### 2. 参数默认值调整

为了简化配置，以下参数有了新的默认值。如果您的配置中未显式指定这些参数，Terraform 将在下次应用时尝试将资源更新为新的默认值。

| 参数名称 | 旧默认值 | 新默认值 | 说明 |
|----------|----------|----------|------|
| `ttl` | 30 | **360** | 数据总保留时间（天）。 |
| `hot_ttl` | 无 (Optional) | **30** | 热存储数据保留时间（天）。 |
| `infrequent_access_ttl` | 无 (Optional) | **3650** | 低频存储数据保留时间（天）。 |

### 升级建议

1.  **检查 `retention_period`**：全局搜索 `.tf` 代码，将所有 `retention_period` 替换为 `ttl`。
2.  **确认保留策略**：
    *   如果您希望保持之前的 **30天** 数据保留期，请务必显式设置 `ttl = 30`，否则它将默认为 360 天。
    *   如果您不希望启用分层存储或希望使用特定的热存储/低频存储周期，请显式配置 `hot_ttl` 和 `infrequent_access_ttl`，以避免使用新的默认值覆盖现有状态。

## 示例

### 场景：保持原有 30 天保留期

```hcl
resource "alicloud_log_store" "keep_30_days" {
  project_name  = "tf-project"
  logstore_name = "tf-logstore"
  
  # 必须显式指定 ttl 为 30，否则默认为 360
  ttl = 30
}
```

### 场景：使用默认分层策略

如果不指定相关参数，LogStore 将配置为：
- 数据总保留时间：360 天 (`ttl`)
- 热存储保留时间：30 天 (`hot_ttl`)
- 低频存储保留时间：3650 天 (`infrequent_access_ttl`)

> 注意：请在执行 `terraform apply`前仔细检查 `terraform plan` 的输出，确认变更符合预期。
