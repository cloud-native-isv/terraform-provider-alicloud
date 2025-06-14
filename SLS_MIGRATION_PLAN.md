# SLS 服务迁移计划 - 替换为 aliyunSlsAPI 实现

## 迁移概述

将 `/cws_data/terraform-provider-alicloud/alicloud/service_alicloud_sls.go` 中的方法替换为使用 `aliyunSlsAPI` 中提供的方法。如果 `aliyunSlsAPI` 没有提供对应的方法，需要先在 `cws-lib-go` 中增加相关实现。

## 迁移步骤

### 第一步：调研现有 service_alicloud_sls.go 中的方法

- [x] 分析 `service_alicloud_sls.go` 中所有需要迁移的方法
- [x] 列出每个方法的功能和参数
- [x] 确定哪些方法在 `aliyunSlsAPI` 中已存在
- [x] 确定哪些方法需要新增

#### 第一步完成 - 现有方法分析结果

**已迁移的方法（使用 aliyunSlsAPI）：**
- ✅ `CreateSlsLogging` -> 已使用 `aliyunSlsAPI.CreateLogProjectLogging`
- ✅ `UpdateSlsLogging` -> 已使用 `aliyunSlsAPI.UpdateLogProjectLogging`
- ✅ `DeleteSlsLogging` -> 已使用 `aliyunSlsAPI.DeleteLogProjectLogging`
- ✅ `GetSlsLogging` -> 已使用 `aliyunSlsAPI.GetLogProjectLogging`

**需要迁移的核心描述方法：**
- [ ] `DescribeSlsProject` -> 需要使用 `aliyunSlsAPI` 的 LogProject 相关方法
- [ ] `DescribeSlsLogStore` -> 需要使用 `aliyunSlsAPI` 的 LogStore 相关方法
- [ ] `DescribeSlsAlert` -> 需要添加 Alert 类型和 CRUD 方法
- [ ] `DescribeSlsScheduledSQL` -> 需要添加 ScheduledSQL 类型和 CRUD 方法
- [ ] `DescribeSlsCollectionPolicy` -> 需要添加 CollectionPolicy 类型和 CRUD 方法
- [ ] `DescribeSlsOssExportSink` -> 需要添加 OssExportSink 类型和 CRUD 方法
- [ ] `DescribeSlsEtl` -> 需要添加 ETL 类型和 CRUD 方法

**辅助方法：**
- [ ] `DescribeListTagResources` -> 需要添加标签管理方法
- [ ] `SetResourceTags` -> 需要添加标签管理方法
- [ ] `DescribeGetLogStoreMeteringMode` -> 需要添加 LogStore 计费模式相关方法

**状态刷新方法（保持现有实现）：**
- `SlsProjectStateRefreshFunc`
- `SlsLogStoreStateRefreshFunc`
- `SlsAlertStateRefreshFunc`
- `SlsScheduledSQLStateRefreshFunc`
- `SlsCollectionPolicyStateRefreshFunc`
- `SlsOssExportSinkStateRefreshFunc`
- `SlsEtlStateRefreshFunc`

### 第二步：调研 SDK 中的类型定义和函数

- [x] 查找 SLS SDK 中可用的 Alert 相关方法
- [x] 查找 SLS SDK 中可用的 ScheduledSQL 相关方法  
- [x] 查找 SLS SDK 中可用的 Collection Policy 相关方法
- [x] 查找 SLS SDK 中可用的 OSS Export 相关方法
- [x] 查找 SLS SDK 中可用的 ETL 相关方法
- [x] 查找 SLS SDK 中可用的 Tag 相关方法

#### 第二步完成 - SDK 可用方法调研结果

**Alert 相关方法（✅ SDK 已支持）：**
- `CreateAlert` - 创建告警规则
- `GetAlert` - 获取告警规则详情
- `UpdateAlert` - 更新告警规则
- `DeleteAlert` - 删除告警规则
- `ListAlerts` - 列出告警规则
- `EnableAlert` - 启用告警规则
- `DisableAlert` - 禁用告警规则

**ScheduledSQL 相关方法（✅ SDK 已支持）：**
- `CreateScheduledSQL` - 创建定时 SQL 任务
- `GetScheduledSQL` - 获取定时 SQL 任务详情
- `UpdateScheduledSQL` - 更新定时 SQL 任务
- `DeleteScheduledSQL` - 删除定时 SQL 任务
- `ListScheduledSQLs` - 列出定时 SQL 任务
- `EnableScheduledSQL` - 启用定时 SQL 任务
- `DisableScheduledSQL` - 禁用定时 SQL 任务

**Collection Policy 相关方法（✅ SDK 已支持）：**
- `UpsertCollectionPolicy` - 创建或更新收集策略
- `GetCollectionPolicy` - 获取收集策略详情
- `DeleteCollectionPolicy` - 删除收集策略
- `ListCollectionPolicies` - 列出收集策略

**OSS Export 相关方法（✅ SDK 已支持）：**
- `CreateOSSExport` - 创建 OSS 导出任务
- `GetOSSExport` - 获取 OSS 导出任务详情
- `UpdateOSSExport` - 更新 OSS 导出任务
- `DeleteOSSExport` - 删除 OSS 导出任务
- `ListOSSExports` - 列出 OSS 导出任务
- `StartOSSExport` - 启动 OSS 导出任务
- `StopOSSExport` - 停止 OSS 导出任务

**ETL 相关方法（✅ SDK 已支持）：**
- `CreateETL` - 创建 ETL 任务
- `GetETL` - 获取 ETL 任务详情
- `UpdateETL` - 更新 ETL 任务
- `DeleteETL` - 删除 ETL 任务
- `ListETLs` - 列出 ETL 任务
- `StartETL` - 启动 ETL 任务
- `StopETL` - 停止 ETL 任务

**Tag 相关方法（✅ SDK 已支持）：**
- `TagResources` - 为资源添加标签
- `UntagResources` - 移除资源标签
- `ListTagResources` - 列出资源标签

**LogStore 计费模式相关方法：**
- `GetLogStoreMeteringMode` - 需要在 SDK 中查找对应的方法
- `UpdateLogStoreMeteringMode` - 需要在 SDK 中查找对应的方法

### 第三步：在 alicloud_sls_types.go 中定义相关类型

- [x] 分析 SDK 中的请求/响应结构
- [x] 在 `/cws_data/cws-lib-go/lib/cloud/aliyun/api/sls/alicloud_sls_types.go` 中定义对应的类型
- [x] 确保类型定义与 SDK 兼容并符合命名规范

#### 第三步完成 - 新增类型定义

**已添加的类型：**
- ✅ `ScheduledSQL` 及相关配置类型（`ScheduledSQLConfig`、`ScheduledSQLSchedule`）
- ✅ `CollectionPolicy` 及相关配置类型（`CollectionPolicyConfig`、`CollectionPolicyDataConfig`、`CollectionPolicyCentralizeConfig`、`CollectionPolicyResourceDir`）
- ✅ `OSSExport` 及相关配置类型（`OSSExportConfig`、`OSSExportSink`）
- ✅ `ETL` 及相关配置类型（`ETLConfig`、`ETLSink`）
- ✅ `LogStoreMeteringMode` - LogStore 计费模式配置
- ✅ `TagResource`、`TagResourceRequest`、`TagResourceRequestTag` - 标签管理相关类型

**已添加的转换函数（在 alicloud_sls_utils.go 中）：**
- ✅ Alert 相关转换函数（`convertAlertToSDKCreateRequest`、`convertSDKAlertToAlert` 等）
- ✅ ScheduledSQL 相关转换函数（`convertScheduledSQLToSDKCreateRequest`、`convertSDKScheduledSQLToScheduledSQL` 等）
- ✅ CollectionPolicy 相关转换函数（`convertCollectionPolicyToSDKUpsertRequest`、`convertSDKCollectionPolicyToCollectionPolicy` 等）
- ✅ OSSExport 相关转换函数（`convertOSSExportToSDKCreateRequest`、`convertSDKOSSExportToOSSExport` 等）
- ✅ ETL 相关转换函数（`convertETLToSDKCreateRequest`、`convertSDKETLToETL` 等）
- ✅ 标签管理转换函数（`convertTagResourceRequestToSDK`、`convertSDKTagResourceToTagResource` 等）

### 第四步：在 alicloud_sls_api.go 中添加 CRUD 方法

- [x] 为每个新类型在 `/cws_data/cws-lib-go/lib/cloud/aliyun/api/sls/alicloud_sls_api.go` 中添加：
  - Create 方法
  - Get 方法  
  - Update 方法
  - Delete 方法
  - List 方法（如果适用）

#### 第四步完成 - 新增 CRUD 方法

**Alert 相关方法（✅ 已完成）：**
- ✅ `CreateAlert` - 创建告警规则
- ✅ `GetAlert` - 获取告警规则详情
- ✅ `UpdateAlert` - 更新告警规则
- ✅ `DeleteAlert` - 删除告警规则
- ✅ `ListAlerts` - 列出告警规则
- ✅ `EnableAlert` - 启用告警规则
- ✅ `DisableAlert` - 禁用告警规则

**ScheduledSQL 相关方法（✅ 已完成）：**
- ✅ `CreateScheduledSQL` - 创建定时 SQL 任务
- ✅ `GetScheduledSQL` - 获取定时 SQL 任务详情
- ✅ `UpdateScheduledSQL` - 更新定时 SQL 任务
- ✅ `DeleteScheduledSQL` - 删除定时 SQL 任务
- ✅ `ListScheduledSQLs` - 列出定时 SQL 任务
- ✅ `EnableScheduledSQL` - 启用定时 SQL 任务
- ✅ `DisableScheduledSQL` - 禁用定时 SQL 任务

**CollectionPolicy 相关方法（✅ 已完成）：**
- ✅ `UpsertCollectionPolicy` - 创建或更新收集策略
- ✅ `GetCollectionPolicy` - 获取收集策略详情
- ✅ `DeleteCollectionPolicy` - 删除收集策略
- ✅ `ListCollectionPolicies` - 列出收集策略

**OSSExport 相关方法（✅ 已完成）：**
- ✅ `CreateOSSExport` - 创建 OSS 导出任务
- ✅ `GetOSSExport` - 获取 OSS 导出任务详情
- ✅ `UpdateOSSExport` - 更新 OSS 导出任务
- ✅ `DeleteOSSExport` - 删除 OSS 导出任务
- ✅ `ListOSSExports` - 列出 OSS 导出任务
- ✅ `StartOSSExport` - 启动 OSS 导出任务
- ✅ `StopOSSExport` - 停止 OSS 导出任务

**ETL 相关方法（✅ 已完成）：**
- ✅ `CreateETL` - 创建 ETL 任务
- ✅ `GetETL` - 获取 ETL 任务详情
- ✅ `UpdateETL` - 更新 ETL 任务
- ✅ `DeleteETL` - 删除 ETL 任务
- ✅ `ListETLs` - 列出 ETL 任务
- ✅ `StartETL` - 启动 ETL 任务
- ✅ `StopETL` - 停止 ETL 任务

**标签管理方法（✅ 已完成）：**
- ✅ `TagResources` - 为资源添加标签
- ✅ `UntagResources` - 移除资源标签
- ✅ `ListTagResources` - 列出资源标签

**LogStore 计费模式方法（✅ 已完成）：**
- ✅ `GetLogStoreMeteringMode` - 获取 LogStore 计费模式
- ✅ `UpdateLogStoreMeteringMode` - 更新 LogStore 计费模式

**方法特性说明：**
- 所有方法都实现了完整的错误处理和状态码检查
- List 方法都支持分页处理，自动处理多页结果聚合
- 使用统一的日志记录模式（Debug 和 Info 级别）
- 通过转换函数与 SDK 类型进行数据转换
- 遵循 API 设计最佳实践，封装分页逻辑

### 第五步：更新 service_alicloud_sls.go 中的封装方法

- [ ] 修改 `SlsService` 结构，确保包含 `aliyunSlsAPI` 客户端
- [ ] 更新构造函数以初始化 `aliyunSlsAPI` 客户端
- [ ] 将现有方法替换为调用 `aliyunSlsAPI` 的实现

### 第六步：更新资源定义文件

- [ ] 更新 `/cws_data/terraform-provider-alicloud/alicloud/resource_alicloud_log_*.go` 中的资源定义
- [ ] 确保所有资源文件使用新的服务方法
- [ ] 验证参数类型匹配和错误处理

## 需要迁移的方法列表

基于当前 `service_alicloud_sls.go` 文件，需要迁移的方法包括：

### 已存在的方法（LogProjectLogging）
- [x] `CreateSlsLogging` -> `CreateLogProjectLogging`
- [x] `UpdateSlsLogging` -> `UpdateLogProjectLogging` 
- [x] `DeleteSlsLogging` -> `DeleteLogProjectLogging`
- [x] `GetSlsLogging` -> `GetLogProjectLogging`

### 需要新增的方法

#### Alert 相关
- [ ] `DescribeSlsAlert` -> 需要添加 Alert 类型和 CRUD 方法

#### ScheduledSQL 相关  
- [ ] `DescribeSlsScheduledSQL` -> 需要添加 ScheduledSQL 类型和 CRUD 方法

#### Collection Policy 相关
- [ ] `DescribeSlsCollectionPolicy` -> 需要添加 CollectionPolicy 类型和 CRUD 方法

#### OSS Export Sink 相关
- [ ] `DescribeSlsOssExportSink` -> 需要添加 OssExportSink 类型和 CRUD 方法

#### ETL 相关
- [ ] `DescribeSlsEtl` -> 需要添加 ETL 类型和 CRUD 方法

### 其他功能方法
- [ ] `DescribeSlsProject` -> 使用现有 LogProject 方法
- [ ] `DescribeSlsLogStore` -> 使用现有 LogStore 方法
- [ ] `SetResourceTags` -> 需要添加标签管理方法

## 执行检查点

每完成一个步骤，更新此文档中的复选框状态。

## 注意事项

1. **错误处理**：确保新实现保持与原有代码相同的错误处理模式
2. **类型转换**：注意 SDK 类型和 Terraform 类型之间的转换
3. **向后兼容**：确保新实现不破坏现有的 Terraform 配置
4. **测试验证**：每个步骤完成后进行基本的编译和功能测试

## 完成状态

- [ ] 第一步：调研现有方法 
- [ ] 第二步：调研 SDK 类型和函数
- [ ] 第三步：定义新类型
- [ ] 第四步：添加 CRUD 方法
- [ ] 第五步：更新服务封装方法
- [ ] 第六步：更新资源定义文件
- [ ] 验证和测试