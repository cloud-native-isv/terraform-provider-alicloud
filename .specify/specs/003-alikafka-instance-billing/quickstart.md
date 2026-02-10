# Quickstart: AliKafka 实例计费组合一致性

## 目的

验证 AliKafka 实例创建路径已支持合法计费组合，并对不支持组合给出清晰错误。

## 最小验证步骤

1. 检查资源与服务层入口文件：
   - 资源层：alicloud/resource_alicloud_alikafka_instance.go
   - 服务层：alicloud/service_alicloud_alikafka_instance.go
2. 验证创建请求路径统一使用 CreateInstance API，并在创建前完成组合校验。
3. 验证默认行为：未提供实例类型或计费方式时，沿用既有默认逻辑并回显最终选择。
4. 运行最小回归测试：
   - `make test`
   - 如具备条件，执行 AliKafka 实例相关验收测试（TF_ACC=1）。

## 验收要点

- Reserved + Prepaid、Reserved + Postpaid、Serverless + Postpaid 能创建成功。
- Serverless + Prepaid 在创建前返回明确错误信息。
- 成功创建返回 instanceId；如产生订单返回 orderId 或等价标识。
- 资源状态中回显最终选择的 instanceType 与 billingType。

## 手工验收记录

- 日期：2026-02-10
- 范围：计费组合一致性验收
- 结果：未执行（需要真实 Aliyun 账号与 AliKafka 实例）
- 备注：可执行 `go test ./alicloud -run TestAccAliKafkaInstance` 进行验证

### 最小回归检查记录

- 日期：2026-02-10
- 项目：最小回归检查
- 结果：未执行（需要真实环境与凭证）
- 备注：建议执行 `make test` 作为基础验证

## 实现更新记录

- 日期：2026-02-10
- 内容：完成 AliKafka 实例计费组合一致性实现与单测补齐；尚未执行实际验收测试。
