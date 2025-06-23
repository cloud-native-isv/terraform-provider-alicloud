# Terraform Provider Alicloud 开发指南

## 概述

本文档为 Terraform Provider Alicloud 项目的开发指南，包含大语言模型代码生成规范、架构设计原则、编码标准和最佳实践。本指南旨在确保代码质量、一致性和可维护性，为开发者提供全面的技术规范。

## 一、代码生成规范

### 1.1 项目信息参考
- 参考项目目录中的 README.md 文件内容，以此了解项目的基本信息
- 如果项目根目录下存在 docs 目录，则参考其中的内容
- 查看 examples 目录了解实际使用场景和模式

### 1.2 任务执行流程
- 复杂任务先创建 TODO.md 文件列出计划和步骤，然后一步一步执行
- 每完成一项更新一次 TODO.md 文档中对应的记录
- 在任务结束之后再检查 TODO.md 中是否都完成
- 对于大型重构任务，建议分阶段进行并记录每个阶段的验证点

### 1.3 文件操作规范
- 在执行复杂的文件操作时，先生成一个 python 或者 shell 脚本，然后通过执行脚本来进行操作
- 批量操作前务必进行备份
- 使用版本控制跟踪所有变更

### 1.4 语言使用规范
- 生成文档时使用中文
- 生成代码中的注释和日志使用英文
- API 文档和错误消息使用英文以保持国际化兼容性

### 1.5 代码拆分规范
- 当编程语言代码文件（*.go、*.java、*.py、*.ts、*.js、*.c、*.cpp、*.cs、*.php、*.rb、*.rs等）超过1000行时，需要进行拆分以提高代码的可维护性和可读性
- 数据文件（*.json、*.yaml、*.csv、*.xml等）不受此限制
- 按功能模块进行拆分，确保每个文件职责单一且清晰

## 二、架构设计原则

### 2.1 分层架构设计

Resource 或 datasource 层应该调用 service 层提供的函数，而不是直接调用底层的 sdk 或者 api 函数。

**架构层次：**
```
Provider 层 (alicloud/)
├── Resource 层 (resource_alicloud_*.go)
├── DataSource 层 (data_source_alicloud_*.go)
└── Service 层 (service_alicloud_*.go)
    └── API 层 (CWS-Lib-Go)
        └── SDk 层 (阿里云官方SDK)
```

对于添加或者重构一个已有的 resource 或 datasource，需要先定义他的 service 层，resource 或 datasource 只依赖与 service 层提供的函数。

service 层可能包含一个或多个 go 文件，包含了针对资源对象的增删改查方法和状态刷新方法。

### 2.2 Service层API调用规范

当前项目的 service 层包含三种对底层 API 的调用方式：

#### 2.2.1 client.RpcPost直接HTTP请求（❌ 不推荐，应废弃）

```go
// 不推荐的方式
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")
```

**存在问题：**
- 维护成本高
- 当云服务API更新时会产生大量联动修改
- 缺乏类型安全和自动化代码生成
- 错误处理复杂
- 缺乏请求重试和熔断机制

#### 2.2.2 第三方SDK调用（⚠️ 不推荐）

```go
// 不推荐的方式
import "github.com/aliyun/aliyun-log-go-sdk"
import "github.com/aliyun/aliyun-oss-go-sdk"
```

**存在问题：**
- 第三方SDK缺乏维护和更新
- 版本兼容性问题
- 安全性和稳定性无法保证
- 可能存在安全漏洞

#### 2.2.3 CWS-Lib-Go封装调用（✅ 推荐）

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
- 内置重试机制和错误恢复
- 统一的认证和配置管理

### 2.3 API分页逻辑封装

所有分页/分批逻辑应封装在 `*_api.go` 文件中，外部调用者无需处理分页细节。API层应提供简单的方法来抽象分页复杂性、页面大小管理和结果聚合。

```go
// 推荐做法：封装分页逻辑
func (s *EcsService) DescribeInstances(request *ecs.DescribeInstancesRequest) ([]*ecs.Instance, error) {
    var allInstances []*ecs.Instance
    pageNumber := 1
    pageSize := 50
    
    for {
        request.PageNumber = pageNumber
        request.PageSize = pageSize
        
        response, err := s.client.DescribeInstances(request)
        if err != nil {
            return nil, err
        }
        
        allInstances = append(allInstances, response.Instances.Instance...)
        
        if len(response.Instances.Instance) < pageSize {
            break
        }
        pageNumber++
    }
    
    return allInstances, nil
}
```

## 三、编码规范

### 3.1 命名规范

#### 3.1.1 资源和数据源命名
- 资源：`alicloud_<service>_<resource>`
- 数据源：`alicloud_<service>_<resource>s` (复数)
- 服务名使用小写下划线分隔：`ecs`, `rds`, `slb`
- 资源名称应简洁明了，避免冗余

**示例：**
```
alicloud_ecs_instance
alicloud_rds_instance
alicloud_slb_load_balancer
alicloud_ecs_instances (数据源)
```

#### 3.1.2 函数和变量命名
- 函数：驼峰命名 (`resourceAlicloudEcs`)
- 变量：下划线命名 (`access_key`)
- 代表ID的变量应该是 `resourceId`
- 代表Name的变量应该是 `resourceName`
- 变量标识一定要明确是ID还是Name
- 常量使用全大写下划线分隔：`DEFAULT_TIMEOUT`

#### 3.1.3 ID 字段命名规范
所有表示 ID 的字段应一致使用变量名 `Id` 而不是 `ID`，以保持与许多自动生成工具的兼容性。

**示例：**
- `WorkspaceId`
- `UserId`
- `ResourceId`
- `InstanceId`

### 3.2 资源结构要求

所有资源必须包含以下方法：
- `Create` - 创建资源
- `Read` - 读取资源状态
- `Update` - 更新资源（如果支持）
- `Delete` - 删除资源
- `Schema` - 资源模式定义
- `Importer` - 资源导入支持（推荐）

**资源文件结构示例：**
```go
func resourceAlicloudEcsInstance() *schema.Resource {
    return &schema.Resource{
        Create: resourceAlicloudEcsInstanceCreate,
        Read:   resourceAlicloudEcsInstanceRead,
        Update: resourceAlicloudEcsInstanceUpdate,
        Delete: resourceAlicloudEcsInstanceDelete,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            // 资源属性定义
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
        },
    }
}
```

### 3.3 错误处理模式

```go
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}
```

**增强错误处理：**
```go
// 推荐的错误处理模式
if err != nil {
    if IsExpectedErrors(err, []string{"InvalidInstance.NotFound", "Forbidden.InstanceNotFound"}) {
        log.Printf("[WARN] Resource (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

### 3.4 状态管理规范

#### 3.4.1 基本规则
- **禁止在Create函数中直接调用Read函数**：应使用StateRefreshFunc机制等待资源创建完成
- 资源不存在时使用 `d.SetId("")`
- `Read` 方法中设置所有计算属性
- 使用适当的超时时间，避免无限等待
- 实现幂等性操作

#### 3.4.2 State Refresh 最佳实践

**正确做法：**

```go
// 在Create函数中使用StateRefreshFunc等待资源就绪
stateConf := BuildStateConf([]string{"Pending", "Starting"}, []string{"Running", "Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.ResourceStateRefreshFunc(id, []string{"Failed", "Error"}))
if _, err := stateConf.WaitForState(); err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// 最后调用Read同步状态
return resourceAlicloudServiceResourceRead(d, meta)
```

**错误做法：**

```go
// 错误做法：在Create函数中直接调用Read
func resourceCreate(d *schema.ResourceData, meta interface{}) error {
    // ... 创建资源 ...
    d.SetId(id)
    
    // ❌ 错误：不应直接调用Read函数
    return resourceRead(d, meta)
}
```

### 3.5 数据验证和转换

#### 3.5.1 输入验证
```go
// Schema中的验证
"instance_type": {
    Type:         schema.TypeString,
    Required:     true,
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
},
```

#### 3.5.2 类型转换
```go
// 安全的类型转换
func convertToStringSlice(v interface{}) []string {
    if v == nil {
        return []string{}
    }
    vList := v.([]interface{})
    result := make([]string, len(vList))
    for i, val := range vList {
        result[i] = val.(string)
    }
    return result
}
```

### 3.6 代码格式规范
- 使用 `gofmt` 格式化
- 遵循 Go 规范
- 避免重复代码
- 添加有意义注释
- 函数长度不超过50行，超过时考虑拆分
- 使用一致的错误处理模式

## 四、测试规范

### 4.1 单元测试要求
- 每个 service 函数都应有对应的单元测试
- 测试覆盖率应达到80%以上
- 使用 Mock 对象测试外部依赖

### 4.2 集成测试规范
```go
func TestAccAlicloudEcsInstance_basic(t *testing.T) {
    var instance ecs.Instance
    resourceId := "alicloud_ecs_instance.default"
    ra := resourceAttrInit(resourceId, testAccEcsInstanceBasicMap)
    testAccCheck := ra.resourceAttrMapUpdateSet()
    resource.Test(t, resource.TestCase{
        PreCheck: func() {
            testAccPreCheck(t)
        },
        IDRefreshName: resourceId,
        Providers:     testAccProviders,
        CheckDestroy:  testAccCheckEcsInstanceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccEcsInstanceConfigBasic(EcsInstanceCommonTestCase),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckEcsInstanceExists(resourceId, &instance),
                    testAccCheck(map[string]string{
                        "instance_name": "tf-testAccEcsInstanceConfigBasic",
                    }),
                ),
            },
        },
    })
}
```

## 五、文档要求

### 5.1 资源文档结构
每个资源和数据源的文档必须包含：
- **资源描述**：简要说明资源的用途和功能
- **参数说明**：所有输入参数的详细描述，包括类型、是否必需、默认值
- **属性说明**：所有输出属性的描述
- **导入指南**：如何导入现有资源
- **使用示例**：完整的、可运行的示例代码
- **注意事项**：重要的使用限制和注意事项

### 5.2 示例代码规范
```hcl
# 基础示例
resource "alicloud_ecs_instance" "example" {
  availability_zone = data.alicloud_zones.default.zones.0.id
  security_groups   = [alicloud_security_group.default.id]
  instance_type     = "ecs.n4.large"
  system_disk_category = "cloud_efficiency"
  image_id         = data.alicloud_images.default.images.0.id
  instance_name    = "terraform-example"
  vswitch_id       = alicloud_vswitch.default.id
}
```

## 六、性能优化指南

### 6.1 API调用优化
- 合理使用批量API减少网络请求
- 实现客户端缓存机制
- 使用连接池管理HTTP连接
- 设置合理的超时时间

### 6.2 状态刷新优化
- 使用指数退避算法进行重试
- 避免过于频繁的状态检查
- 合理设置状态检查间隔

## 七、安全最佳实践

### 7.1 凭证管理
- 不在代码中硬编码访问密钥
- 支持多种认证方式：环境变量、配置文件、RAM角色
- 实现访问密钥轮换机制

### 7.2 权限控制
- 遵循最小权限原则
- 明确标记敏感字段
- 实现资源访问控制

## 八、故障排除指南

### 8.1 常见问题
- **API限流**：实现重试机制和退避策略
- **网络超时**：检查网络连接和防火墙设置
- **权限错误**：验证RAM角色和策略配置
- **资源不存在**：正确处理NotFound错误

### 8.2 调试技巧
- 启用详细日志记录
- 使用环境变量控制调试级别
- 记录关键API调用和响应

## 九、版本兼容性

### 9.1 向后兼容
- 新增字段使用Optional标记
- 废弃字段保持向后兼容
- 提供迁移指南

### 9.2 版本管理
- 遵循语义化版本控制
- 维护CHANGELOG文档
- 提供升级指导

## 十、总结

本指南为 Terraform Provider Alicloud 项目的核心开发规范，旨在确保代码质量、一致性和可维护性。开发过程中应严格遵循这些规范，特别是：

1. **架构分层原则**：严格按照 Provider → Resource/DataSource → Service → API 的层次结构
2. **状态管理最佳实践**：正确使用StateRefreshFunc，避免直接调用Read函数
3. **错误处理规范**：统一的错误处理模式和适当的错误分类
4. **测试覆盖**：确保充分的单元测试和集成测试
5. **文档完整性**：提供清晰、完整的使用文档和示例

遵循这些规范将有助于提高代码质量，降低维护成本，并为用户提供更好的使用体验。

