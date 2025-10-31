# Quickstart: ConcurrentUpdateBucketFailed error support

本特性默认开启，无需额外配置。以下示例展示如何使用 `alicloud_oss_bucket_public_access_block` 并验证在并发更新时的自动重试行为。

## 前置条件
- 已配置阿里云凭证（AK/SK/或RAM Role）
- 已存在目标 OSS 桶（或在 Terraform 中先创建）
- 已安装 Terraform 和本 Provider

## 示例 HCL

```hcl
resource "alicloud_oss_bucket" "example" {
  bucket = "example-concurrency-bucket"
}

resource "alicloud_oss_bucket_public_access_block" "example" {
  bucket              = alicloud_oss_bucket.example.bucket
  block_public_access = true
}
```

## 如何触发并观察并发冲突
- 在 `terraform apply` 同时，通过控制台或其他自动化对同一桶的公共访问阻断进行修改，可能触发 409 并发错误。
- Provider 会自动重试，并在日志中打印重试次数与最后一次错误摘要（不包含敏感信息）。

## 超时与失败信息
- 若并发冲突持续到超过资源操作超时，`apply` 将失败并返回包含“并发更新检测、请冷却后重试”的清晰提示，便于排障。

## 构建与编译校验

```bash
# 从仓库根目录执行
make
```

如需运行 Go 测试：

```bash
# 可选
go test ./...
```

## 兼容性
- 无并发冲突路径下，执行逻辑与延迟与变更前保持等效；不引入新配置项。
