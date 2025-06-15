## 代码生成规范
- 参考项目目录中的README.md文件内容，以此了解项目的基本信息，如果项目根目录下存在docs目录，则参考其中的内容。
- 复杂任务先创建TODO.md文件列出计划和步骤，然后一步一步执行，每完成一项更新一次TODO.md文档中对应的记录，在任务结束之后再检查TODO.md中是否都完成。
- 在执行复杂的文件操作时，先生成一个python或者shell脚本，然后通过执行脚本来进行操作。
- 生成文档时使用中文，生成代码中的注释和日志使用英文。
- 当编程语言代码文件（*.go、*.java、*.py、*.ts、*.js、*.c、*.cpp、*.cs、*.php、*.rb、*.rs等）超过1500行时，需要进行拆分以提高代码的可维护性和可读性。数据文件（*.json、*.yaml、*.csv、*.xml等）不受此限制。

# Terraform Provider Alicloud 开发指南

## Service层API调用规范

当前项目的service层包含三种对底层API的调用方式：

### 1. client.RpcPost直接HTTP请求（❌ 不推荐，应废弃）
```go
// 不推荐的方式
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")
```
**存在问题：**
- 维护成本高
- 当云服务API更新时会产生大量联动修改
- 缺乏类型安全和自动化代码生成
- 错误处理复杂

### 2. 第三方SDK调用（⚠️ 不推荐）
```go
// 不推荐的方式
import "github.com/aliyun/aliyun-log-go-sdk"
import "github.com/aliyun/aliyun-oss-go-sdk"
```
**存在问题：**
- 第三方SDK缺乏维护和更新
- 版本兼容性问题
- 安全性和稳定性无法保证

### 3. CWS-Lib-Go封装调用（✅ 推荐）
```go
// 推荐的方式
import "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
```
**优势：**
- 使用官方SDK进行二次封装
- 提供统一的API接口和错误处理
- 定期维护和更新
- 类型安全和代码生成支持
- 更好的测试覆盖率

**使用示例：**
```go
func (s *EcsService) DescribeInstances(request *api.DescribeInstancesRequest) (*api.DescribeInstancesResponse, error) {
    client := s.client.WithApiInfo("ecs", "2014-05-26", "DescribeInstances")
    response := &api.DescribeInstancesResponse{}
    err := client.DoAction(request, response)
    return response, err
}
```

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

