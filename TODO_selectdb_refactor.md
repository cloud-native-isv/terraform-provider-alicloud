# SelectDB Type System Refactoring TODO

## 目标
将SelectDB相关代码从使用多个option结构体改为使用统一的Instance struct类型。

## 完成的任务
- [x] 更新Terraform resource schema定义
- [x] 更新resourceAliCloudSelectDBInstance函数的CRUD operations
- [x] 更新API层的CreateInstance、ListInstances、ModifyInstance、CheckCreateInstance函数
- [x] 更新service层的DescribeSelectDBInstances、ModifySelectDBInstance函数
- [x] 修复resource层的ModifySelectDBInstance调用
- [x] 添加UpgradeInstanceEngineVersion API函数
- [x] 更新service层的UpgradeSelectDBInstanceEngineVersion函数
- [x] 更新WaitForSelectDBInstanceStatus函数
- [x] 更新resetSelectDBInstancePassword函数
- [x] 更新DescribeSelectDBRegions函数
- [x] 删除ConvertToCreateInstanceOptions函数
- [x] 删除ConvertToModifyInstanceOptions函数
- [x] 修复所有字段名引用(DBInstanceId -> Id)
- [x] 修复错误处理函数引用(IsNotFoundError -> NotFoundError)

## 待完成的任务

### 类型定义需要清理
- [ ] 从types文件中删除InstanceQueryOptions
- [ ] 从types文件中删除InstanceCreateOptions
- [ ] 从types文件中删除InstanceModifyOptions
- [ ] 从types文件中删除InstanceOperationResult
- [ ] 从types文件中删除InstancePaginationInfo
- [ ] 从types文件中删除InstanceUpgradeOptions
- [ ] 从types文件中删除ResetPasswordOptions
- [ ] 从types文件中删除DescribeRegionsOptions

### 待实现的功能
- [ ] 实现SetResourceTags函数用于tag管理
- [ ] 验证所有功能正常工作

## 当前状态
主要的重构工作已经完成！所有编译错误都已修复，代码成功使用统一的Instance struct而不是分离的option types。

## 下一步
1. 清理不再使用的类型定义
2. 测试验证功能正常
3. 实现剩余的功能（如tag管理）
