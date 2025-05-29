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
- 创建/更新后调用`Read`同步状态
- 资源不存在时使用`d.SetId("")`
- `Read`方法中设置所有计算属性

## 文档要求
包含：资源描述、参数说明、属性说明、导入指南、使用示例

## 编码规范
- 使用`gofmt`格式化
- 遵循Go规范
- 避免重复代码
- 添加有意义注释

