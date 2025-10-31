# Quickstart — 自动接管已存在数据库（alicloud_db_database）

本向导演示如何在数据库已存在的情况下，使用 Terraform 自动接管该数据库为受管对象。

## 前置条件
- 已安装 Terraform 与本 Provider。
- 目标 RDS 实例已存在且包含名称为 `example_db` 的数据库。
- 具备最小只读权限（Plan）或创建/删除权限（Apply）。

## 示例 HCL

```hcl
resource "alicloud_db_database" "example" {
  instance_id   = var.instance_id
  name          = "example_db"
  character_set = "utf8"     # PostgreSQL 可为 "UTF8,en_US.UTF-8,en_US.UTF-8"
  description   = "managed by terraform"

  # 由 Provider 维护的只读提示字段（不会触发实际变更）
  # adopt_existing 与 adoption_notice 为 Optional+Computed，供 Plan 透明性使用
  # adopt_existing 便于在 CI 中断言“将发生接管”
}
```

## 运行 Plan（展示接管提示）

```bash
terraform plan
```

预期在 Plan 输出中看到类似提示字段（根据现场情况可能带有更细致的说明）：

- `adopt_existing = true`（Optional+Computed）
- `adoption_notice = "Detected existing database and will adopt it on apply"`（只读提示）
- 当 `description` 与现状不同：
  `adoption_notice = "Detected existing database and will adopt it on apply; description differs and won't be aligned in this apply."`

如果权限不足，Plan 会返回友好的提示，说明需要的最小权限（Describe/List Databases）。
如果遇到限流/系统繁忙等暂时性错误，Plan 会给出非阻断性说明并继续（可能无法确认是否将接管）。

## 运行 Apply（执行接管或创建）

```bash
terraform apply
```

- 若数据库已存在：Apply 不会创建新对象，而是将现有数据库接入为受管。
- 若数据库不存在：正常创建并等待就绪，然后进入受管状态。

## 再次 Plan（幂等验证）

```bash
terraform plan
```

- 预期显示“无改动（No changes）”。

## 差异冲突处理
- 若不可变字段（如 `character_set`）与现状冲突：Apply 将失败并给出明确指引（保守策略）。
- 可变字段（如 `description`）在接管的同一轮不自动对齐；可在下一次变更中更新。

## 故障与重试
- 常见可重试错误（限流、系统繁忙等）由 Provider 自动重试并采用退避策略。
- 实例非 Running 时，删除/创建前会等待到 Running。

## 最小权限
- Plan：只读列举/查询数据库权限即可。
- Apply：创建/删除数据库权限（若需要创建/删除）。
