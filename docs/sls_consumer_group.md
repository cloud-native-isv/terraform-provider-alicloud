# alicloud_sls_consumer_group

管理阿里云日志服务（SLS）消费组的 Terraform 资源。

- 资源名：`alicloud_sls_consumer_group`
- 导入 ID 格式：`project:logstore:consumer_group`
- ForceNew 字段：`project`、`logstore`、`consumer_group`
- 可更新字段：`timeout`、`order`

## 示例

```hcl
resource "alicloud_sls_consumer_group" "cg" {
  project        = "my-project"
  logstore       = "app-logs"
  consumer_group = "etl-workers"

  # 行为参数
  timeout = 60   # 秒
  order   = false

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
```

## 参数说明

- `project` (必填，ForceNew)：SLS Project 名称。
- `logstore` (必填，ForceNew)：SLS Logstore 名称。
- `consumer_group` (必填，ForceNew)：消费组名称。
- `timeout` (可选，默认 60)：消费组心跳超时（秒）。
- `order` (可选，默认 false)：是否按序消费。

校验规则：

- `project`/`logstore`/`consumer_group` 名称需匹配正则 `^[a-zA-Z0-9][a-zA-Z0-9_-]{1,127}$`
- `timeout` 必须在 `1..86400` 秒范围内

## 只读属性

- `checkpoints`：消费组的分片检查点列表，包含：
  - `shard_id`：分片 ID
  - `checkpoint`：游标
  - `update_time`：更新时间（Unix 秒）
  - `consumer`：更新该检查点的消费者名

## 导入

```bash
terraform import alicloud_sls_consumer_group.cg project:logstore:consumer_group
```

导入成功后执行 `terraform plan` 不应产生非预期变更；若 ID 解析失败或目标不存在，将给出明确错误提示。

## 备注

- 若消费组已存在，创建时会进行 adopt 并对齐 `timeout`/`order` 与 HCL 配置。
- 输入校验与错误处理遵循 Provider 统一规范，包含对临时错误的重试（如 `ServiceUnavailable`/`Throttling`/`SystemBusy`/`OperationConflict`）。
