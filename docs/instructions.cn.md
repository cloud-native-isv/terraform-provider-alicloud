# Terraform Provider Alicloud 开发指南

## 概述

本文档为 Terraform Provider Alicloud 项目的完整开发指南，包含大语言模型代码生成规范、架构设计原则、编码标准和最佳实践。本指南旨在确保代码质量、一致性和可维护性，为开发者提供全面的技术规范。

## 1. 代码生成规范

### 1.1 项目信息参考
- 参考项目目录中的 README.md 文件内容，了解项目基本信息
- 如果项目根目录下存在 docs 目录，则参考其中的内容
- 查看 examples 目录了解实际使用场景和模式

### 1.2 任务执行流程
- 复杂任务先创建 TODO.md 文件列出计划和步骤，然后逐步执行
- 每完成一项更新 TODO.md 文档中对应的记录
- 任务结束后检查 TODO.md 中是否都完成
- 大型重构任务建议分阶段进行并记录每个阶段的验证点

### 1.3 文件操作规范
- 复杂文件操作时，先生成 Python 或 Shell 脚本，然后执行脚本
- 批量操作前务必备份
- 使用版本控制跟踪所有变更

### 1.4 语言使用规范
- 生成文档时使用中文
- 代码注释和日志使用英文
- API 文档和错误消息使用英文以保持国际化兼容性

### 1.5 代码拆分规范
- 编程语言代码文件（*.go、*.java、*.py、*.ts、*.js、*.c、*.cpp、*.cs、*.php、*.rb、*.rs等）超过1000行时需要拆分
- 数据文件（*.json、*.yaml、*.csv、*.xml等）不受此限制
- 按功能模块拆分，确保每个文件职责单一且清晰

## 2. 架构设计原则

### 2.1 分层架构设计

Resource 或 DataSource 层应调用 Service 层提供的函数，而不是直接调用底层 SDK 或 API 函数。

**架构层次：**
```
Provider 层 (alicloud/)
├── Resource 层 (resource_alicloud_*.go)
├── DataSource 层 (data_source_alicloud_*.go)
└── Service 层 (service_alicloud_*.go)
    └── API 层 (CWS-Lib-Go)
        └── SDK 层 (阿里云官方SDK)
```

Service 层包含一个或多个 Go 文件，包含针对资源对象的增删改查方法和状态刷新方法。

### 2.2 Service层API调用规范

#### 2.2.1 ✅ 推荐：CWS-Lib-Go封装调用

```go
// 推荐的方式
import "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"

// 1. 创建Service对象
service, err := NewServiceService(client)
if err != nil {
    return WrapError(err)
}

// 2. 调用Service层方法
instance, err := service.CreateInstance(request)
if err != nil {
    return WrapError(err)
}
```

**优势：**
- 使用官方SDK进行二次封装
- 提供统一的API接口和错误处理
- 定期维护和更新，类型安全和代码生成支持
- 内置重试机制和错误恢复

#### 2.2.2 ❌ 避免：直接HTTP请求和第三方SDK

```go
// 不推荐：直接HTTP请求
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")

// 不推荐：第三方SDK
import "github.com/aliyun/aliyun-log-go-sdk"
```

### 2.3 API分页逻辑封装

所有分页逻辑应封装在 `*_api.go` 文件中，外部调用者无需处理分页细节：

```go
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

### 2.4 Service层ID编码规范

每个 Service 文件都需要定义相应的 `Encode*Id` 和 `Decode*Id` 函数：

```go
// EncodeJobId 将工作空间ID、命名空间和作业ID编码为单一ID字符串
// 格式: workspaceId:namespace:jobId
func EncodeJobId(workspaceId, namespace, jobId string) string {
    return fmt.Sprintf("%s:%s:%s", workspaceId, namespace, jobId)
}

// DecodeJobId 解析作业ID字符串为工作空间ID、命名空间和作业ID组件
func DecodeJobId(id string) (string, string, string, error) {
    parts := strings.Split(id, ":")
    if len(parts) != 3 {
        return "", "", "", fmt.Errorf("invalid job ID format, expected workspaceId:namespace:jobId, got %s", id)
    }
    return parts[0], parts[1], parts[2], nil
}
```

### 2.5 Service层状态管理规范

针对每个 Resource，在对应的 Service 中需要添加 `*StateRefreshFunc` 状态刷新函数和 `WaitFor*` 状态同步函数：

```go
// StateRefreshFunc 状态刷新函数
func (s *FlinkService) FlinkJobStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
    return func() (interface{}, string, error) {
        object, err := s.DescribeFlinkJob(id)
        if err != nil {
            if NotFoundError(err) {
                return nil, "", nil
            }
            return nil, "", WrapError(err)
        }

        var currentStatus string
        if object.Status != nil {
            currentStatus = object.Status.CurrentJobStatus
        }

        for _, failState := range failStates {
            if currentStatus == failState {
                return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
            }
        }
        return object, currentStatus, nil
    }
}

// WaitFor 状态同步函数
func (s *FlinkService) WaitForFlinkJobCreating(id string, timeout time.Duration) error {
    stateConf := BuildStateConf(
        []string{"STARTING", "SUBMITTING"}, // pending states
        []string{"RUNNING", "FINISHED"},    // target states  
        timeout,
        5*time.Second,
        s.FlinkJobStateRefreshFunc(id, []string{"FAILED", "CANCELLED", "CANCELLING"}),
    )
    
    _, err := stateConf.WaitForState()
    return WrapErrorf(err, IdMsg, id)
}
```

## 3. 编码规范

### 3.1 命名约定

#### 3.1.1 资源和数据源命名
- Resources: `alicloud_<service>_<resource>`
- Data sources: `alicloud_<service>_<resource>s` (复数)
- 服务名使用小写下划线：`ecs`, `rds`, `slb`

#### 3.1.2 函数和变量命名
- 函数：camelCase (`resourceAlicloudEcs`)
- 变量：snake_case (`access_key`)
- ID字段：`resourceId`, 名称字段：`resourceName`
- 常量：大写下划线 (`DEFAULT_TIMEOUT`)

#### 3.1.3 ID字段命名约定
所有表示ID的字段统一使用 `Id` 而不是 `ID`：

```go
// ✅ 正确
"WorkspaceId": workspace.WorkspaceId,
"UserId": user.UserId,

// ❌ 错误  
"WorkspaceID": workspace.WorkspaceID,
"UserID": user.UserID,
```

### 3.2 Resource结构要求

所有资源必须包含以下方法：

```go
func resourceAlicloudEcsInstance() *schema.Resource {
    return &schema.Resource{
        Create: resourceAlicloudEcsInstanceCreate,
        Read:   resourceAlicloudEcsInstanceRead,
        Update: resourceAlicloudEcsInstanceUpdate, // 可选
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

#### 3.3.1 基本错误处理

推荐使用 `alicloud/errors.go` 中封装的错误类型判断函数，而不是 `IsExpectedErrors`：

```go
if err != nil {
    if NotFoundError(err) {
        log.Printf("[WARN] Resource (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

#### 3.3.2 常用错误判断函数

```go
// 资源不存在错误判断
if NotFoundError(err) {
    d.SetId("")
    return nil
}

// 资源已存在错误判断
if IsAlreadyExistError(err) {
    // 处理资源已存在的情况
    return resourceAlicloudServiceResourceRead(d, meta)
}

// 需要重试的错误判断
if NeedRetry(err) {
    time.Sleep(5 * time.Second)
    return resource.RetryableError(err)
}
```

#### 3.3.3 特定服务错误处理

对于特定服务的错误，可以使用预定义的错误码列表：

```go
// ECS 相关错误
if IsExpectedErrors(err, EcsNotFound) {
    d.SetId("")
    return nil
}

// SLB 繁忙错误
if IsExpectedErrors(err, SlbIsBusy) {
    return resource.RetryableError(err)
}

// 数据库状态错误
if IsExpectedErrors(err, OperationDeniedDBStatus) {
    return resource.RetryableError(err)
}
```

#### 3.3.4 避免使用的模式

不推荐直接使用 `IsExpectedErrors` 进行错误判断：

```go
// ❌ 不推荐的方式
if IsExpectedErrors(err, []string{"InvalidInstance.NotFound", "Forbidden.InstanceNotFound"}) {
    d.SetId("")
    return nil
}

// ✅ 推荐的方式
if NotFoundError(err) {
    d.SetId("")
    return nil
}
```

#### 3.3.5 复合错误处理模式

```go
if err != nil {
    // 首先检查是否为资源不存在错误
    if NotFoundError(err) {
        if !d.IsNewResource() {
            log.Printf("[DEBUG] Resource alicloud_service_resource DescribeResource Failed!!! %s", err)
            d.SetId("")
            return nil
        }
        return WrapError(err)
    }
    
    // 检查是否为资源已存在错误（通常在Create操作中）
    if IsAlreadyExistError(err) {
        // 如果资源已存在，读取现有资源状态
        return resourceAlicloudServiceResourceRead(d, meta)
    }
    
    // 检查是否需要重试
    if NeedRetry(err) {
        return resource.RetryableError(err)
    }
    
    // 其他错误直接返回
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

#### 3.3.6 错误处理最佳实践

1. **优先使用封装的错误判断函数**：
   - `NotFoundError(err)` - 检查资源不存在
   - `IsAlreadyExistError(err)` - 检查资源已存在  
   - `NeedRetry(err)` - 检查是否需要重试

2. **使用预定义错误码列表**：
   - `EcsNotFound`, `SlbIsBusy`, `OperationDeniedDBStatus` 等

3. **统一错误包装**：
   - 使用 `WrapError(err)` 或 `WrapErrorf(err, msg, args...)` 包装错误
   - 包含详细的上下文信息

4. **适当的日志记录**：
   - 记录关键错误信息用于调试
   - 区分不同级别的日志（DEBUG、WARN、ERROR）
```go

### 3.4 状态管理规范

#### 3.4.1 基本规则
- **禁止在Create函数中直接调用Read函数**：使用StateRefreshFunc机制等待资源创建完成
- 使用 `d.SetId("")` 当资源不存在时
- 在 `Read` 方法中设置所有computed属性
- 实现幂等操作

#### 3.4.2 正确的状态刷新模式

```go
// ✅ 正确：在Create函数中使用Service层的WaitFor函数等待资源就绪
err = service.WaitForServiceResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
if err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// 最后调用Read同步状态
return resourceAlicloudServiceResourceRead(d, meta)
```
### 3.5 数据验证和转换

#### 3.5.1 输入验证
```go
"instance_type": {
    Type:         schema.TypeString,
    Required:     true,
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
},
```

#### 3.5.2 类型转换
```go
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

## 4. Resource开发指南

### 4.1 标准Import包结构

```go
package alicloud

import (
    "fmt"
    "log"
    "time"
    "strings"

    "github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
    "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)
```

### 4.2 Schema定义规范

#### 4.2.1 字段类型定义

```go
// Required字段
"instance_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The ID of the instance.",
},

// Optional字段
"name": {
    Type:        schema.TypeString,
    Optional:    true,
    Description: "The name of the resource.",
},

// Computed字段
"status": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "The status of the resource.",
},

// Optional + Computed字段
"instance_name": {
    Type:        schema.TypeString,
    Optional:    true,
    Computed:    true,
    Description: "The name of the instance.",
},
```

#### 4.2.2 嵌套对象定义

```go
"config": {
    Type:     schema.TypeList,
    Required: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "cpu": {
                Type:        schema.TypeInt,
                Required:    true,
                Description: "CPU specifications.",
            },
            "memory": {
                Type:        schema.TypeInt,
                Required:    true,
                Description: "Memory specifications in GB.",
            },
        },
    },
    Description: "Configuration parameters.",
},
```

### 4.3 CRUD操作实现

#### 4.3.1 Create方法

```go
func resourceAliCloudServiceResourceCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }

    // 构建请求对象
    request := &serviceAPI.CreateResourceRequest{
        InstanceId: d.Get("instance_id").(string),
        Name:       d.Get("name").(string),
    }

    // 使用Retry创建资源
    var result *serviceAPI.Resource
    err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
        resp, err := service.CreateResource(request)
        if err != nil {
            if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
                time.Sleep(5 * time.Second)
                return resource.RetryableError(err)
            }
            return resource.NonRetryableError(err)
        }
        result = resp
        return nil
    })

    if err != nil {
        return WrapErrorf(err, DefaultErrorMsg, "alicloud_service_resource", "CreateResource", AlibabaCloudSdkGoERROR)
    }

    d.SetId(result.ResourceId)

    // 等待资源就绪
    err = service.WaitForServiceResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
    if err != nil {
        return WrapErrorf(err, IdMsg, d.Id())
    }

    // 最后调用Read同步状态
    return resourceAlicloudServiceResourceRead(d, meta)
}
```

#### 4.3.2 Read方法

```go
func resourceAliCloudServiceResourceRead(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }

    object, err := service.DescribeResource(d.Id())
    if err != nil {
        if !d.IsNewResource() && NotFoundError(err) {
            log.Printf("[DEBUG] Resource alicloud_service_resource DescribeResource Failed!!! %s", err)
            d.SetId("")
            return nil
        }
        return WrapError(err)
    }

    // 设置所有必要字段
    d.Set("instance_id", object.InstanceId)
    d.Set("name", object.Name)
    d.Set("status", object.Status)

    return nil
}
```

#### 4.3.3 Delete方法

```go
func resourceAliCloudServiceResourceDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }

    err = service.DeleteResource(d.Id())
    if err != nil {
        if NotFoundError(err) {
            return nil
        }
        return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteResource", AlibabaCloudSdkGoERROR)
    }

    // 等待资源被删除
    stateConf := &resource.StateChangeConf{
        Pending: []string{"Deleting"},
        Target:  []string{""},
        Refresh: func() (interface{}, string, error) {
            obj, err := service.DescribeResource(d.Id())
            if err != nil {
                if NotFoundError(err) {
                    return nil, "", nil
                }
                return nil, "", WrapError(err)
            }
            return obj, obj.Status, nil
        },
        Timeout:    d.Timeout(schema.TimeoutDelete),
        Delay:      5 * time.Second,
        MinTimeout: 3 * time.Second,
    }

    _, err = stateConf.WaitForState()
    if err != nil {
        return WrapErrorf(err, IdMsg, d.Id())
    }

    return nil
}
```

### 4.4 资源创建的resource.Retry逻辑

```go
// 常见的可重试错误
var retryableErrors = []string{
    "ServiceUnavailable",
    "ThrottlingException", 
    "InternalError",
    "Throttling",
    "SystemBusy",
    "OperationConflict",
}

err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
    _, err := service.CreateResource(request)
    if err != nil {
        if IsExpectedErrors(err, retryableErrors) {
            time.Sleep(5 * time.Second)
            return resource.RetryableError(err)
        }
        return resource.NonRetryableError(err)
    }
    return nil
})
```


## 11. 开发检查清单

### 11.1 基础要求
- [ ] 使用标准的import包结构
- [ ] 通过Service层调用CWS-Lib-Go API
- [ ] 正确定义Schema（Required/Optional/Computed）
- [ ] 实现所有必要的CRUD方法
- [ ] 添加适当的Timeout配置

### 11.2 状态管理
- [ ] Create后使用Service层的WaitFor函数等待资源就绪
- [ ] Delete后使用Service层的WaitFor函数等待资源删除
- [ ] Read方法中正确设置所有字段
- [ ] 最后调用Read同步状态

### 11.3 代码质量
- [ ] 遵循ID字段命名约定（使用Id而不是ID）
- [ ] 添加适当的Description
- [ ] 正确处理复杂对象的类型转换
- [ ] 合理的日志记录

## 12. 总结

本指南为Terraform Provider Alicloud项目的核心开发规范，开发过程中应严格遵循，特别是：

1. **架构分层原则**：严格遵循Provider → Resource/DataSource → Service → API层次
2. **状态管理最佳实践**：正确使用StateRefreshFunc，避免直接调用Read函数
3. **错误处理规范**：统一错误处理模式和适当错误分类
4. **测试覆盖**：确保充分的单元测试和集成测试
5. **文档完整性**：提供清晰、完整的使用文档和示例

遵循这些规范将有助于提高代码质量，降低维护成本，为用户提供更好的体验。
