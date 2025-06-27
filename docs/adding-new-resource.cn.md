# 添加新的Resource开发指导文档

## 概述

本文档基于现有阿里云资源实现的分析，提供添加新resource的完整开发指导。遵循本指导可以确保新resource与现有代码风格保持一致，符合项目架构设计原则。

## 1. Import包和API对象调用规范

### 1.1 标准Import包结构

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

### 1.2 API对象调用规范

**✅ 推荐的调用方式 - 使用CWS-Lib-Go封装的API：**

```go
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

// 3. 使用CWS-Lib-Go类型定义
request := &serviceAPI.CreateRequest{
    Name:            d.Get("name").(string),
    ResourceGroupId: d.Get("resource_group_id").(string),
    // ...
}
```

**❌ 避免的调用方式：**

```go
// 不要直接调用RPC方法
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")

// 不要使用第三方SDK
import "github.com/aliyun/aliyun-log-go-sdk"
```

## 2. Resource Schema定义规范

### 2.1 基本Schema结构

```go
func resourceAliCloudServiceResource() *schema.Resource {
    return &schema.Resource{
        Create: resourceAliCloudServiceResourceCreate,
        Read:   resourceAliCloudServiceResourceRead,
        Update: resourceAliCloudServiceResourceUpdate, // 可选
        Delete: resourceAliCloudServiceResourceDelete,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            // Schema定义
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
        },
    }
}
```

### 2.2 字段属性规范

#### 2.2.1 Required字段
```go
"instance_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The ID of the instance.",
},
```

#### 2.2.2 Optional字段
```go
"security_group_id": {
    Type:        schema.TypeString,
    Optional:    true,
    Description: "The ID of the security group.",
    ForceNew:    true,
},
```

#### 2.2.3 Computed字段
```go
"status": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "The status of the resource.",
},
```

#### 2.2.4 Optional + Computed字段
```go
"instance_name": {
    Type:        schema.TypeString,
    Optional:    true,
    Computed:    true,
    Description: "The name of the instance. Defaults to the instance ID if not specified.",
},
```

#### 2.2.5 嵌套对象定义
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

#### 2.2.6 验证规则
```go
"instance_type": {
    Type:         schema.TypeString,
    Optional:     true,
    Default:      "ecs.t5-lc1m1.small",
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
    ForceNew:     true,
},
```

### 2.3 ID字段命名约定

所有ID字段必须使用`Id`而不是`ID`：

```go
// ✅ 正确
"InstanceId": instance.InstanceId,
"UserId": user.UserId,

// ❌ 错误
"InstanceID": instance.InstanceID,
"UserID": user.UserID,
```

## 3. Service对象创建模式

### 3.1 标准Service创建模式

所有resource的CRUD操作都必须通过Service层：

```go
func resourceAliCloudServiceResourceCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    
    // 创建Service对象
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }
    
    // 业务逻辑...
}
```

### 3.2 Service层方法调用

```go
// 创建资源
resource, err := service.CreateResource(request)
if err != nil {
    return WrapError(err)
}

// 读取资源
resource, err := service.DescribeResource(id)
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}

// 删除资源
err = service.DeleteResource(id)
if err != nil {
    return WrapError(err)
}
```

## 4. 资源创建的resource.Retry逻辑

### 4.1 基本Retry模式

```go
// 创建资源时使用Retry机制处理临时错误
var result *serviceAPI.Resource
err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
    resp, err := service.CreateResource(request)
    if err != nil {
        // 可重试的错误
        if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException"}) {
            time.Sleep(5 * time.Second)
            return resource.RetryableError(err)
        }
        // 不可重试的错误
        return resource.NonRetryableError(err)
    }
    result = resp
    return nil
})

if err != nil {
    return WrapErrorf(err, DefaultErrorMsg, "alicloud_service_resource", "CreateResource", AlibabaCloudSdkGoERROR)
}
```

### 4.2 常见可重试错误类型

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

## 5. 创建资源后的状态刷新逻辑

### 5.1 使用BuildStateConf和StateRefreshFunc

**✅ 推荐模式：使用BuildStateConf**

```go
// 等待资源创建完成
stateConf := BuildStateConf(
    []string{"Creating", "Pending"},      // pending states
    []string{"Running", "Available"},     // target states  
    d.Timeout(schema.TimeoutCreate),      // timeout
    5*time.Second,                        // poll interval
    service.ResourceStateRefreshFunc(d.Id(), []string{"Failed", "Error"}), // refresh func with fail states
)
if _, err := stateConf.WaitForState(); err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}
```

### 5.2 StateRefreshFunc实现示例

Service层需要提供StateRefreshFunc方法：

```go
func (s *ServiceService) ResourceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
    return func() (interface{}, string, error) {
        object, err := s.DescribeResource(id)
        if err != nil {
            if NotFoundError(err) {
                return nil, "", nil
            }
            return nil, "", WrapError(err)
        }
        
        for _, failState := range failStates {
            if object.Status == failState {
                return object, object.Status, WrapError(Error(FailedToReachTargetStatus, object.Status))
            }
        }
        
        return object, object.Status, nil
    }
}
```

### 5.3 不同类型资源的状态等待示例

**ECS实例创建：**
```go
stateConf := BuildStateConf(
    []string{"Pending", "Starting"}, 
    []string{"Running"}, 
    d.Timeout(schema.TimeoutCreate), 
    10*time.Second, 
    ecsService.InstanceStateRefreshFunc(d.Id(), []string{"Stopped", "Failed"})
)
```

**RDS实例创建：**
```go
stateConf := BuildStateConf(
    []string{"Creating"}, 
    []string{"Running"}, 
    d.Timeout(schema.TimeoutCreate), 
    30*time.Second, 
    rdsService.DBInstanceStateRefreshFunc(d.Id(), []string{"Failed"})
)
```

**负载均衡器创建：**
```go
stateConf := BuildStateConf(
    []string{"inactive"}, 
    []string{"active"}, 
    d.Timeout(schema.TimeoutCreate), 
    5*time.Second, 
    slbService.LoadBalancerStateRefreshFunc(d.Id(), []string{"failed"})
)
```

## 6. 删除资源后的状态同步逻辑

### 6.1 使用StateChangeConf检查NotFoundError

**标准删除等待模式：**

```go
func resourceAliCloudServiceResourceDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }

    // 执行删除操作
    err = service.DeleteResource(d.Id())
    if err != nil {
        if NotFoundError(err) {
            return nil
        }
        return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteResource", AlibabaCloudSdkGoERROR)
    }

    // 等待资源被完全删除
    stateConf := &resource.StateChangeConf{
        Pending: []string{"Deleting"},
        Target:  []string{""},
        Refresh: func() (interface{}, string, error) {
            resource, err := service.DescribeResource(d.Id())
            if err != nil {
                if NotFoundError(err) {
                    // 资源已被删除，这是我们期望的结果
                    return nil, "", nil
                }
                return nil, "", WrapError(err)
            }
            // 如果仍能获取到资源，说明还在删除中
            return resource, resource.Status, nil
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

### 6.2 ECS实例删除示例

```go
stateConf := &resource.StateChangeConf{
    Pending: []string{"Stopping", "Deleting"},
    Target:  []string{""},
    Refresh: func() (interface{}, string, error) {
        instance, err := ecsService.DescribeInstance(d.Id())
        if err != nil {
            if NotFoundError(err) {
                return nil, "", nil
            }
            return nil, "", WrapError(err)
        }
        return instance, instance.Status, nil
    },
    Timeout:    d.Timeout(schema.TimeoutDelete),
    Delay:      10 * time.Second,
    MinTimeout: 3 * time.Second,
}
```

### 6.3 简单删除模式

对于一些简单的资源，可以直接删除无需等待：

```go
func resourceAliCloudSimpleResourceDelete(d *schema.ResourceData, meta interface{}) error {
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
        return WrapError(err)
    }
    
    return nil
}
```

## 7. 完整的Resource实现示例

### 7.1 基础Resource结构

```go
func resourceAliCloudServiceResource() *schema.Resource {
    return &schema.Resource{
        Create: resourceAliCloudServiceResourceCreate,
        Read:   resourceAliCloudServiceResourceRead,
        Update: resourceAliCloudServiceResourceUpdate,
        Delete: resourceAliCloudServiceResourceDelete,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
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
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
        },
    }
}
```

### 7.2 Create方法实现

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
    stateConf := BuildStateConf(
        []string{"Creating"}, 
        []string{"Available"}, 
        d.Timeout(schema.TimeoutCreate), 
        5*time.Second, 
        service.ResourceStateRefreshFunc(d.Id(), []string{"Failed"})
    )
    if _, err := stateConf.WaitForState(); err != nil {
        return WrapErrorf(err, IdMsg, d.Id())
    }

    // 最后调用Read同步状态
    return resourceAliCloudServiceResourceRead(d, meta)
}
```

### 7.3 Read方法实现

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

    // 设置所有必要字段，防止forces replacement
    d.Set("instance_id", object.InstanceId)
    d.Set("name", object.Name)
    d.Set("status", object.Status)

    return nil
}
```

### 7.4 Delete方法实现

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

## 8. 开发检查清单

### 8.1 基础要求
- [ ] 使用标准的import包结构
- [ ] 通过Service层调用CWS-Lib-Go API
- [ ] 正确定义Schema（Required/Optional/Computed）
- [ ] 实现所有必要的CRUD方法
- [ ] 添加适当的Timeout配置

### 8.2 错误处理
- [ ] 正确使用WrapError和WrapErrorf
- [ ] 适当处理NotFoundError
- [ ] 在Create中使用resource.Retry处理临时错误
- [ ] 正确的错误消息格式

### 8.3 状态管理
- [ ] Create后使用BuildStateConf等待资源就绪
- [ ] Delete后使用StateChangeConf等待资源删除
- [ ] Read方法中正确设置所有字段
- [ ] 最后调用Read同步状态

### 8.4 代码质量
- [ ] 遵循ID字段命名约定（使用Id而不是ID）
- [ ] 添加适当的Description
- [ ] 正确处理复杂对象的类型转换
- [ ] 合理的日志记录

## 9. 常见问题和解决方案

### 9.1 Forces Replacement问题
**问题：** Terraform plan显示资源需要重新创建
**解决：** 确保Read方法中设置了所有必要字段，特别是Required和Optional字段

### 9.2 状态刷新超时
**问题：** StateRefreshFunc超时
**解决：** 检查状态值是否正确，调整轮询间隔和超时时间

### 9.3 ID格式问题
**问题：** 复合ID解析错误
**解决：** 使用一致的ID分隔符，提供解析函数

```go
// ID格式: workspace_id:namespace_name:resource_name
func parseResourceId(id string) (string, string, string, error) {
    parts := strings.Split(id, ":")
    if len(parts) != 3 {
        return "", "", "", fmt.Errorf("invalid resource ID format: %s", id)
    }
    return parts[0], parts[1], parts[2], nil
}
```

## 10. 总结

遵循本指导文档可以确保新resource：
1. 与现有代码风格保持一致
2. 正确使用架构分层设计
3. 提供良好的用户体验
4. 具备稳定的错误处理能力
5. 支持完整的生命周期管理

在开发过程中，建议参考现有的阿里云资源实现，如ECS、RDS、SLB等服务的resource文件，它们提供了完整的最佳实践示例。

### 常见阿里云服务资源参考
- **ECS服务**: `resource_alicloud_instance.go`
- **RDS服务**: `resource_alicloud_db_instance.go`  
- **SLB服务**: `resource_alicloud_slb.go`
- **VPC服务**: `resource_alicloud_vpc.go`
- **OSS服务**: `resource_alicloud_oss_bucket.go`

这些实现文件展示了不同复杂度资源的开发模式，可以作为新resource开发的参考模板。