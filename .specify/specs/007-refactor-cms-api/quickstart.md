# Quickstart: CMS Provider 封装能力改造

## 1. 确认分支与范围

```sh
git branch --show-current
for f in {service,resource,data_source}_alicloud_cms_{addon,agg_task,alert,cloud_resource,context,context_store,dataset,delivery_task,entity_store,integration_policy,memory,memory_store,pipeline,prometheus,service,umodel,workspace}.go; do
  test -e "alicloud/$f" && echo "alicloud/$f" || echo "MISSING: alicloud/$f"
done
```

期望：当前分支为 `007-refactor-cms-api`，并输出 51 个存在的文件。

## 2. 建立范围矩阵

对每个对象组记录：

- Provider 文件：service/resource/data_source 三个路径。
- 当前状态：占位、已实现但直连、已通过服务层、是否注册。
- CWS-Lib-Go API wrapper：对应 `pkg/cws-lib-go/lib/cloud/aliyun/api/cms/alicloud_cms_*` 方法。
- 决策：`migrate_now`、`no_change_placeholder`、`enable_later`、`needs_api_gap_resolution`。
- 验证证据：测试名、代码审查项、兼容性备注。

首个必须迁移对象：`alicloud_cms_service` 数据源。

## 3. 实施顺序建议

1. 添加或更新 `alicloud/data_source_alicloud_cms_service_test.go`，先覆盖现有 schema、`enable=Off` 行为、`enable=On` 状态语义和服务层调用边界。
2. 补齐 `alicloud/service_alicloud_cms_service.go`：提供打开/检查 CMS 服务状态的强类型服务方法，并调用本地 CWS-Lib-Go CMS service API。
3. 修改 `alicloud/data_source_alicloud_cms_service.go`：移除直接 `client.RpcPost`，通过服务层完成操作；清理重复 `package alicloud`。
4. 对 16 个其它对象组逐项建立 `MigrationEvidence`：占位对象保持未注册且记录 `placeholder_no_user_visible_change`；若启用，必须先补齐服务方法、schema、测试和 provider 注册。
5. 每实现一个对象组后更新范围矩阵和任务记录，保证 17 个对象组均有结论。

## 4. 分层检查

实现后运行：

```sh
grep -R "RpcPost\|WithCmsClient" -n \
  alicloud/{service,resource,data_source}_alicloud_cms_{addon,agg_task,alert,cloud_resource,context,context_store,dataset,delivery_task,entity_store,integration_policy,memory,memory_store,pipeline,prometheus,service,umodel,workspace}.go
```

期望：资源层和数据源层不出现直接 SDK/RPC 调用；如服务层出现底层调用，必须有无法使用 CWS-Lib-Go wrapper 的明确缺口记录。优先目标是通过 `pkg/cws-lib-go/lib/cloud/aliyun/api/cms` 完成云服务交互。

## 5. 格式化与单元验证

```sh
gofmt -w alicloud/*alicloud_cms*.go
go test ./pkg/cws-lib-go/lib/cloud/aliyun/api/cms
go test ./alicloud -run 'Test.*Cms|Test.*CMS' -count=1
```

如目标测试名称尚未存在，任务阶段必须先创建针对迁移对象的测试。

## 6. Provider 编译/质量门禁

```sh
make test
make
```

如果本地环境无法运行完整 `make`，至少记录阻塞原因，并运行可用的 targeted `go test`。

## 7. 兼容性验收

对已启用能力验证：

- Terraform 名称不变。
- schema 字段、默认值和 computed 字段语义不变。
- ID 与 import 行为不变。
- 对象不存在时状态清理或空结果行为不变。
- 错误信息仍可诊断，并通过 `WrapError`/`WrapErrorf` 保持项目风格。

## 8. 后续命令

计划完成后运行：

```text
/speckit.tasks
```

任务拆分时必须按用户故事优先级组织，并将 `alicloud_cms_service` 数据源迁移放入可独立交付的首个切片。

## 9. 可选验收测试（US3）

```sh
cd /cws_data/terraform-provider-alicloud
TF_ACC=1 go test ./alicloud -run TestAccAlicloudCms -count=1
```

当前已知阻塞：

- 主仓库 `go test ./alicloud` 在当前工作区存在模块路径/替换关联问题（`.../pkg/cws-lib-go/lib/cloud/aliyun/api/alicloud: directory not found`）。
- `pkg/cws-lib-go/lib/cloud/aliyun/api/cms` 存在与本次实现无关的预存编译错误（dataset/delivery/entity_store utils），会影响 typed integration 的全链路验证。
