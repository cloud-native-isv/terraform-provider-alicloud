# Quickstart: 使用新 VPC API 的资源

本向导展示如何在 Terraform 中使用升级后的 EIP、EIP 关联、NAT 网关与 SNAT 条目资源（通过 CWS-Lib-Go Service 层）。

> 说明：以下示例为文档用途。实际资源名称以 provider 中对应 `resource_alicloud_*` 实现为准；本变更保持向后兼容，已有 HCL 配置无需修改字段名。

## 1) EIP 地址

```hcl
resource "alicloud_eip_address" "example" {
  bandwidth           = 10
  internet_charge_type = "PayByTraffic"
  description         = "example-eip"
  tags = {
    env = "dev"
  }
}
```

## 2) EIP 关联

```hcl
resource "alicloud_eip_association" "example" {
  allocation_id = alicloud_eip_address.example.id
  instance_id   = alicloud_instance.example.id
  instance_type = "EcsInstance"
}
```

## 3) NAT 网关

```hcl
resource "alicloud_nat_gateway" "example" {
  vpc_id      = alicloud_vpc.example.id
  spec        = "Small"
  description = "nat for egress"
  tags = {
    owner = "network"
  }
}
```

## 4) SNAT 条目

```hcl
resource "alicloud_snat_entry" "example" {
  snat_table_id = alicloud_nat_gateway.example.snat_table_ids[0]
  source_cidr   = "10.0.0.0/24"
  snat_ip       = alicloud_eip_address.example.ip_address
}
```

## 迁移与兼容性
- 本次升级将资源层的底层调用切换为 Service 层 + CWS-Lib-Go API，但字段与语义保持不变；
- 不建议在模块中依赖内部实现细节（例如具体 API 名称），以保持模块的前向兼容性；
- 若使用到增强能力（如删除保护、转发模式等），请关注后续 Provider 版本与发布说明。

## 验证与故障排查
- Plan/Apply 过程中若出现限流/系统繁忙，Provider 会按统一重试策略处理；
- 若资源不存在（例如被控制台删除），Read 将清空状态，后续 Apply 将按期望进行漂移修复；
- 如遇长期 Pending，可在 TF 日志中查看状态刷新信息以定位问题。
