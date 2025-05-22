# Terraform Provider Alibaba Cloud 开发指南

## 环境准备

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x 或更高版本
- [Go](https://golang.org/doc/install) 1.20 或更高版本
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)：
  ```
  go get golang.org/x/tools/cmd/goimports
  ```

## 代码结构

- `alicloud/`: 主要的提供商代码，包含资源定义和数据源
- `website/`: 文档目录
- `examples/`: 示例代码
- `sdk/`: SDK 相关代码
- `ci/`: CI/CD 配置

## 开发工作流程

1. 从主分支创建新的功能分支
2. 实现新功能或修复
3. 编写测试（单元测试和验收测试）
4. 本地测试验证
5. 提交 PR

## 代码生成
- 如果目录中包含README.md文件需要参考其内容
- 如果需要完成的动作较多，就先把需要做的动作写入一个TODO.md文件中，然后再进行代码生成
- 如果文件比较大，生成代码之后不需要进行错误修复
- 生成文档用中文，生成代码注释和日志用英文

## 资源开发最佳实践

### 命名规范

- 资源名称：使用 `alicloud_<service>_<resource>` 格式
- 数据源名称：使用 `alicloud_<service>_<resource>s` 格式，注意复数形式
- 函数名称：使用驼峰命名法（如 `resourceAlicloudEcs`）
- 变量名称：使用下划线命名法（如 `access_key`）

### 资源结构

资源定义应包含以下方法：
- `Create`: 创建资源
- `Read`: 读取资源状态
- `Update`: 更新资源
- `Delete`: 删除资源
- `Schema`: 定义资源属性

### 错误处理

- 使用 SDK 提供的错误处理机制
- 必须处理资源不存在的情况
- 对于可重试的错误，使用重试机制
- 提供有意义的错误消息

```go
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}
```

### 状态管理

- 资源创建和更新后，必须调用 `Read` 方法同步状态
- 使用 `d.SetId("")` 表示资源不存在
- 在 `Read` 方法中设置所有计算属性

### 验收测试

每个资源和数据源必须有验收测试：
- 基本测试：验证基础功能
- 更新测试：验证属性更新
- 导入测试：验证资源导入
- 数据源测试：验证数据源读取

运行验收测试：
```bash
export ALICLOUD_ACCESS_KEY=xxx
export ALICLOUD_SECRET_KEY=xxx
export ALICLOUD_REGION=xxx
export ALICLOUD_ACCOUNT_ID=xxx
TF_ACC=1 go test ./alicloud -v -run=TestAccAlicloud<ResourceName>_basic
```

### 文档标准

所有资源和数据源必须有完整文档：
- 资源描述
- 参数列表及说明
- 属性列表及说明
- 如何导入指南
- 使用示例

## 编码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 编码规范
- 避免重复代码，使用公共函数
- 正确处理并发和锁
- 添加有意义的注释

## 提交和审查

- 提交信息应简洁明了，描述变更内容
- 每个 PR 应解决单一问题或实现单一功能
- PR 应包含相关的测试
- 遵循代码审查流程
