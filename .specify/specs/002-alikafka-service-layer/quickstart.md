# Quickstart: AliKafka 资源分层改造

## 目的

验证 AliKafka 资源已完成服务层对齐且核心生命周期操作可用。

## 最小验证步骤

1. 识别 AliKafka 资源与服务层文件：
   - 资源层：alicloud/resource_alicloud_alikafka_*.go
   - 服务层：alicloud/service_alicloud_alikafka*.go
2. 检查每个资源的核心操作仅调用服务层方法。
3. 检查服务层已补齐资源所需的核心操作方法。
4. 运行最小回归测试：
   - `make test`（单元测试）
   - 选取至少 1 个 AliKafka 资源的验收测试进行验证（如可用）。

## 验收要点

- 资源层无直接 SDK/底层 API 调用。
- 服务层封装分页与重试逻辑。
- 状态等待与刷新逻辑遵循 `wait_for_state.md` 规范。
- 既有资源配置无破坏性变更。

## 兼容性注意事项

- 资源 Schema 未变更，字段名称与语义保持一致。
- Topic 创建默认使用副本数 3（与服务层默认一致），如业务有特殊副本策略需在后续扩展参数。
- 实例删除改为真实调用服务层删除接口，预付费实例仍保持不删除行为。

## 手工验收记录

- 日期：2026-02-09
- 范围：User Story 1（服务层与资源层对齐）
- 结果：未执行（需要真实 Aliyun 账号与 AliKafka 实例）
- 备注：可运行 `go test ./alicloud -run AliKafka` 与最小验收流程进行验证

### Testacc 记录

- 日期：2026-02-09
- 项目：AliKafka 实例 testacc
- 结果：未执行（需要真实 Aliyun 账号与 AliKafka 实例）
- 备注：如具备条件可执行 `go test ./alicloud -run TestAccAliKafkaInstance`

### 最小回归检查记录

- 日期：2026-02-09
- 项目：最小回归检查
- 结果：未执行（需要真实环境与凭证）
- 备注：建议执行 `go test ./alicloud -run AliKafka` 与相关资源最小流程验证
