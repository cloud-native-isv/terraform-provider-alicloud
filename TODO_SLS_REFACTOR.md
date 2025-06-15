# SLS Legacy Service 重构任务

## 目标
将 `service_alicloud_sls_legacy.go` 中的 LogService 功能分解并合并到各自归属的文件中，同时更新所有调用方。

## 分析现状
1. `service_alicloud_sls_legacy.go` 包含：
   - LogService 结构体（包装 SlsService）
   - 多个 Describe 函数（委托给 SlsService）
   - 多个 Wait 函数（用于状态等待）
   - SlsOssShipperStateRefreshFunc 函数

2. 发现的调用方：
   - `resource_alicloud_log_ingestion.go` 使用 LogService

## 重构计划

### 第一步：分析所有 LogService 的使用位置
- [x] 搜索所有引用 LogService 的文件
- [ ] 分析每个使用场景的具体需求

### 第二步：将函数迁移到合适的文件
- [ ] 将 SlsOssShipperStateRefreshFunc 移动到 service_alicloud_sls_base.go
- [ ] 在 SlsService 中添加 Wait 函数（如果还没有的话）
- [ ] 确保所有 Describe 函数在 SlsService 中都有对应实现

### 第三步：更新调用方
- [ ] 更新 resource_alicloud_log_ingestion.go 使用 SlsService 而不是 LogService
- [ ] 检查其他可能的调用方并更新

### 第四步：清理
- [ ] 删除 service_alicloud_sls_legacy.go 文件
- [ ] 验证所有测试通过

### 第五步：验证
- [ ] 运行相关测试
- [ ] 确保没有编译错误