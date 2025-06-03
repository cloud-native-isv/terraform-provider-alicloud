# Terraform Provider Alicloud 开发指南

## 代码生成规则
- 参考目录中的README.md文件内容
- 复杂任务先写TODO.md，再生成代码
- 大文件生成后无需错误修复
- 文档用中文，代码注释和日志用英文

## 命名规范
- 资源：`alicloud_<service>_<resource>`
- 数据源：`alicloud_<service>_<resource>s` (复数)
- 函数：驼峰命名 (`resourceAlicloudEcs`)
- 变量：下划线命名 (`access_key`)
- 代表ID的变量应该是resourceID
- 代表Name的变量应该是resourceName
- 变量标识一定要明确是ID还是Name

## 资源结构
必须包含：`Create`, `Read`, `Update`, `Delete`, `Schema`

## 错误处理模式
```go
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}
```

## 状态管理
- **禁止在Create函数中直接调用Read函数**：应使用StateRefreshFunc机制等待资源创建完成
- 资源不存在时使用`d.SetId("")`
- `Read`方法中设置所有计算属性

### State Refresh 最佳实践
```go
// 正确做法：在Create函数中使用StateRefreshFunc等待资源就绪
stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.ResourceStateRefreshFunc(id, []string{}))
if _, err := stateConf.WaitForState(); err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// 最后调用Read同步状态
return resourceAlicloudServiceResourceRead(d, meta)
```

```go
// 错误做法：在Create函数中直接调用Read
func resourceCreate(d *schema.ResourceData, meta interface{}) error {
    // ... 创建资源 ...
    d.SetId(id)
    
    // ❌ 错误：不应直接调用Read函数
    return resourceRead(d, meta)
}
```

## 文档要求
包含：资源描述、参数说明、属性说明、导入指南、使用示例

## 编码规范
- 使用`gofmt`格式化
- 遵循Go规范
- 避免重复代码
- 添加有意义注释

