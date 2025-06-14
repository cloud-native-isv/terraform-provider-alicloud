# 阿里云Terraform Provider SLS日志服务重构计划

## 项目概述

本重构计划旨在重新整理阿里云Terraform Provider中所有与SLS（Simple Log Service）日志服务相关的代码，统一API调用方式，提高代码质量和可维护性。

## 重构目标

1. **代码结构统一化**：将`service_alicloud_log.go`合并到`service_alicloud_sls.go`中
2. **API统一调用**：所有resource和datasource统一使用`aliyunSlsAPI`库
3. **代码规范化**：遵循Terraform Provider开发最佳实践
4. **功能完整性**：确保所有现有功能在重构后正常工作

## 当前代码结构分析

### 文件清单
```
数据源文件 (Data Sources):
- data_source_alicloud_log_alert_resource.go
- data_source_alicloud_log_projects.go  
- data_source_alicloud_log_service.go
- data_source_alicloud_log_stores.go

资源文件 (Resources):
- resource_alicloud_log_alert.go
- resource_alicloud_log_alert_resource.go
- resource_alicloud_log_audit.go
- resource_alicloud_log_dashboard.go
- resource_alicloud_log_etl.go
- resource_alicloud_log_ingestion.go
- resource_alicloud_log_machine_group.go
- resource_alicloud_log_oss_export.go
- resource_alicloud_log_oss_shipper.go
- resource_alicloud_log_project.go
- resource_alicloud_log_project_logging.go
- resource_alicloud_log_resource.go
- resource_alicloud_log_resource_record.go
- resource_alicloud_log_store.go
- resource_alicloud_log_store_index.go
- resource_alicloud_logtail_attachment.go
- resource_alicloud_logtail_config.go

SLS新版资源文件:
- resource_alicloud_sls_alert.go
- resource_alicloud_sls_collection_policy.go
- resource_alicloud_sls_etl.go
- resource_alicloud_sls_oss_export_sink.go
- resource_alicloud_sls_scheduled_sql.go

服务文件:
- service_alicloud_log.go (37KB - 待合并)
- service_alicloud_sls.go (21KB - 主要服务文件)
```

### 当前API调用方式分析

1. **旧版资源** (`resource_alicloud_log_*`): 使用`github.com/aliyun/aliyun-log-go-sdk`
2. **新版资源** (`resource_alicloud_sls_*`): 使用官方REST API + `aliyunSlsAPI`库
3. **混合状态**: 部分功能重复，API调用不统一

## 重构实施计划

### 第一阶段：服务层重构 (1-2天)

#### 1.1 合并服务文件
- [x] 分析`service_alicloud_log.go`中的所有函数
- [x] 将`service_alicloud_log.go`中的函数迁移到`service_alicloud_sls.go`
- [x] 统一使用`aliyunSlsAPI`库
- [x] 删除`service_alicloud_log.go`文件

#### 1.2 服务函数标准化
已完成重构的主要函数：

**项目管理类**：
```go
✅ DescribeLogProject() -> 使用 SlsService.DescribeSlsProject()
✅ WaitForLogProject() -> 使用 SlsService.SlsProjectStateRefreshFunc()
✅ DescribeLogProjectPolicy() -> 合并到 DescribeSlsProject()
✅ DescribeLogProjectTags() -> 使用 SlsService.DescribeListTagResources()
```

**LogStore管理类**：
```go
✅ DescribeLogStore() -> 使用 SlsService.DescribeSlsLogStore()
✅ WaitForLogStore() -> 使用 SlsService.SlsLogStoreStateRefreshFunc() 
✅ DescribeLogStoreIndex() -> 新增 SlsService.DescribeSlsLogStoreIndex()
```

**告警管理类**：
```go
✅ DescribeLogAlert() -> 使用 SlsService.DescribeSlsAlert()
✅ WaitForLogstoreAlert() -> 使用 SlsService.SlsAlertStateRefreshFunc()
✅ DescribeLogAlertResource() -> 重构为 SlsService.DescribeSlsLogAlertResource()
```

**其他服务类**：
```go
✅ DescribeLogEtl() -> 使用 SlsService.DescribeSlsEtl()
✅ DescribeLogOssExport() -> 使用 SlsService.DescribeSlsOssExportSink()
✅ DescribeLogMachineGroup() -> 新增 SlsService.DescribeSlsMachineGroup()
✅ DescribeLogtailConfig() -> 新增 SlsService.DescribeSlsLogtailConfig()
✅ DescribeLogDashboard() -> 新增 SlsService.DescribeSlsDashboard()
✅ DescribeLogIngestion() -> 新增 SlsService.DescribeSlsIngestion()
✅ DescribeLogResource() -> 新增 SlsService.DescribeSlsResource()
✅ DescribeLogResourceRecord() -> 新增 SlsService.DescribeSlsResourceRecord()
✅ DescribeLogOssShipper() -> 新增 SlsService.DescribeSlsOssShipper()
```

#### 1.3 向后兼容性
- [x] 创建 LogService 包装器提供向后兼容
- [x] 所有原有函数签名保持不变
- [x] 新增统一的状态刷新机制
- [x] 确保现有资源文件无需修改即可工作

### 第二阶段：资源层重构 (3-4天)

#### 2.1 Project资源重构
**文件**: `resource_alicloud_log_project.go`
- [ ] 更新为使用`SlsService`
- [ ] 保持与`resource_alicloud_sls_project`兼容
- [ ] 统一状态管理机制

#### 2.2 LogStore资源重构  
**文件**: `resource_alicloud_log_store.go`
- [ ] 更新为使用`SlsService`
- [ ] 合并计量模式管理功能
- [ ] 统一索引配置管理

#### 2.3 Alert资源统一
**目标**: 统一`resource_alicloud_log_alert.go`和`resource_alicloud_sls_alert.go`
- [ ] 分析两个文件的功能差异
- [ ] 制定兼容性策略
- [ ] 统一Schema定义
- [ ] 迁移到统一的`SlsService.DescribeSlsAlert()`

#### 2.4 其他资源重构
按优先级重构以下资源：
1. **高优先级** (核心功能):
   - `resource_alicloud_log_audit.go`
   - `resource_alicloud_log_dashboard.go` 
   - `resource_alicloud_log_machine_group.go`

2. **中优先级** (常用功能):
   - `resource_alicloud_logtail_config.go`
   - `resource_alicloud_logtail_attachment.go`
   - `resource_alicloud_log_store_index.go`

3. **低优先级** (高级功能):
   - `resource_alicloud_log_oss_shipper.go`
   - `resource_alicloud_log_ingestion.go`
   - `resource_alicloud_log_resource.go`
   - `resource_alicloud_log_resource_record.go`

### 第三阶段：数据源重构 (1天)

#### 3.1 数据源统一
- [ ] `data_source_alicloud_log_projects.go` -> 使用`SlsService`
- [ ] `data_source_alicloud_log_stores.go` -> 使用`SlsService`  
- [ ] `data_source_alicloud_log_service.go` -> 使用`SlsService`
- [ ] `data_source_alicloud_log_alert_resource.go` -> 使用`SlsService`

### 第四阶段：测试和验证 (2-3天)

#### 4.1 单元测试更新
- [ ] 更新所有相关的测试文件
- [ ] 确保测试覆盖率不降低
- [ ] 验证API兼容性

#### 4.2 集成测试
- [ ] 运行完整的测试套件
- [ ] 验证资源创建、更新、删除功能
- [ ] 确认数据源查询功能

## 技术实施细节

### API调用统一标准

#### 统一的凭证管理
```go
type SlsService struct {
    client       *connectivity.AliyunClient
    aliyunSlsAPI *aliyunSlsAPI.SlsAPI
}

func NewSlsService(client *connectivity.AliyunClient) (*SlsService, error) {
    credentials := &aliyunSlsAPI.SlsCredentials{
        AccessKey:     client.AccessKey,
        SecretKey:     client.SecretKey,
        RegionId:      client.RegionId,
        SecurityToken: client.SecurityToken,
    }
    
    slsAPI, err := aliyunSlsAPI.NewSlsAPI(credentials)
    if err != nil {
        return nil, fmt.Errorf("failed to create SLS API client: %w", err)
    }
    
    return &SlsService{
        client:       client,
        aliyunSlsAPI: slsAPI,
    }, nil
}
```

#### 统一的错误处理
```go
func (s *SlsService) handleSlsError(err error, resourceType string, resourceId string) error {
    if err != nil {
        if strings.Contains(err.Error(), "NotExist") {
            return WrapErrorf(NotFoundErr(resourceType, resourceId), NotFoundMsg, "")
        }
        return WrapErrorf(err, DefaultErrorMsg, resourceId, "SLS Operation", AlibabaCloudSdkGoERROR)
    }
    return nil
}
```

#### 统一的状态刷新机制
```go
func (s *SlsService) createStateRefreshFunc(
    resourceId string, 
    describeFunc func(string) (map[string]interface{}, error),
    field string, 
    failStates []string,
) resource.StateRefreshFunc {
    return func() (interface{}, string, error) {
        object, err := describeFunc(resourceId)
        if err != nil {
            if NotFoundError(err) {
                return object, "", nil
            }
            return nil, "", WrapError(err)
        }
        
        v, err := jsonpath.Get(field, object)
        currentStatus := fmt.Sprint(v)
        
        for _, failState := range failStates {
            if currentStatus == failState {
                return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
            }
        }
        return object, currentStatus, nil
    }
}
```

### 数据转换标准

#### API响应转换
```go
// 将aliyunSlsAPI返回的结构体转换为Terraform兼容的map
func convertSlsObjectToMap(obj interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    // 使用反射或手动映射进行转换
    // 确保字段命名一致性
    return result
}
```

#### Schema字段映射
```go
// 定义统一的字段映射规则
var fieldMappings = map[string]string{
    "projectName":      "project_name", 
    "logstoreName":     "logstore_name",
    "createTime":       "create_time",
    "lastModifyTime":   "last_modify_time",
    // ... 更多映射规则
}
```

## 兼容性保证

### 向后兼容策略
1. **保留现有资源名称**: 不改变资源和数据源的名称
2. **Schema兼容**: 确保现有Schema字段不被破坏性修改
3. **行为一致**: 保证CRUD操作的行为与现有版本一致

### 弃用策略
1. **废弃旧API调用**: 逐步移除对旧SDK的依赖
2. **文档更新**: 更新所有相关文档
3. **迁移指南**: 提供从旧版本到新版本的迁移指南

## 质量保证

### 代码审查检查点
- [ ] 所有函数都使用统一的`SlsService`
- [ ] 错误处理符合Terraform Provider标准
- [ ] 状态管理使用StateRefreshFunc机制
- [ ] 单元测试覆盖率≥80%
- [ ] 集成测试通过率100%

### 性能考虑
- [ ] API调用次数优化
- [ ] 缓存机制实现
- [ ] 批量操作支持
- [ ] 超时和重试机制

## 风险评估

### 高风险项
1. **API兼容性变更**: 可能影响现有用户
2. **状态迁移**: 可能导致资源状态不一致
3. **测试覆盖**: 重构可能引入新的bug

### 风险缓解措施
1. **分阶段实施**: 逐步重构，确保每个阶段都经过充分测试
2. **功能对比**: 对比重构前后的功能，确保完全一致
3. **回滚计划**: 准备回滚策略以应对重大问题

## 时间计划

| 阶段 | 预计时间 | 关键里程碑 |
|------|----------|------------|
| 第一阶段 | 2天 | 服务层重构完成 |
| 第二阶段 | 4天 | 核心资源重构完成 |
| 第三阶段 | 1天 | 数据源重构完成 |
| 第四阶段 | 3天 | 测试验证完成 |
| **总计** | **10天** | **重构项目完成** |

## 成功标准

1. **功能完整性**: 所有现有功能在重构后正常工作
2. **代码质量**: 代码复用率提高，维护成本降低
3. **API统一**: 所有SLS相关操作使用统一的API
4. **测试通过**: 所有单元测试和集成测试通过
5. **文档完善**: 重构后的代码有完整的文档

## 后续优化

1. **性能优化**: 根据使用情况优化API调用
2. **功能增强**: 基于统一架构添加新功能
3. **监控改进**: 添加更好的错误监控和日志
4. **用户体验**: 改进错误消息和使用体验

---

*本重构计划将确保阿里云Terraform Provider中SLS服务的代码质量和维护性得到显著提升，为用户提供更好的使用体验。*