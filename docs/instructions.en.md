# Terraform Provider Alicloud Development Guide

## Overview

This document serves as a comprehensive development guide for the Terraform Provider Alicloud project, including Large Language Model code generation specifications, architectural design principles, coding standards, and best practices. This guide aims to ensure code quality, consistency, and maintainability, providing developers with comprehensive technical specifications.

## 1. Code Generation Specifications

### 1.1 Project Information Reference
- Reference the README.md file content in the project directory to understand basic project information
- If there is a docs directory in the project root, reference the content within it
- Check the examples directory to understand actual usage scenarios and patterns

### 1.2 Task Execution Workflow
- For complex tasks, first create a TODO.md file listing plans and steps, then execute step by step
- Update the corresponding records in the TODO.md document after completing each item
- Check if everything in TODO.md is completed after task completion
- Large refactoring tasks are recommended to be done in phases with verification points recorded for each phase

### 1.3 File Operation Specifications
- For complex file operations, first generate Python or Shell scripts, then execute the scripts
- Always backup before batch operations
- Use version control to track all changes

### 1.4 Language Usage Specifications
- Use Chinese when generating documentation
- Use English for code comments and logs
- Use English for API documentation and error messages to maintain internationalization compatibility

### 1.5 Code Splitting Specifications
- Programming language code files (*.go, *.java, *.py, *.ts, *.js, *.c, *.cpp, *.cs, *.php, *.rb, *.rs, etc.) should be split when exceeding 1000 lines
- Data files (*.json, *.yaml, *.csv, *.xml, etc.) are not subject to this limitation
- Split by functional modules, ensuring each file has a single and clear responsibility

## 2. Architectural Design Principles

### 2.1 Layered Architecture Design

The Resource or DataSource layer should call functions provided by the Service layer, rather than directly calling underlying SDK or API functions.

**Architecture Hierarchy:**
```
Provider Layer (alicloud/)
├── Resource Layer (resource_alicloud_*.go)
├── DataSource Layer (data_source_alicloud_*.go)
└── Service Layer (service_alicloud_*.go)
    └── API Layer (CWS-Lib-Go)
        └── SDK Layer (Alibaba Cloud Official SDK)
```

The Service layer contains one or more Go files with CRUD methods and state refresh methods for resource objects.

### 2.2 Service Layer API Call Specifications

#### 2.2.1 ✅ Recommended: CWS-Lib-Go Wrapper Calls

```go
// Recommended approach
import "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"

// 1. Create Service object
service, err := NewServiceService(client)
if err != nil {
    return WrapError(err)
}

// 2. Call Service layer methods
instance, err := service.CreateInstance(request)
if err != nil {
    return WrapError(err)
}
```

**Advantages:**
- Secondary encapsulation using official SDK
- Provides unified API interfaces and error handling
- Regular maintenance and updates, type safety and code generation support
- Built-in retry mechanisms and error recovery

#### 2.2.2 ❌ Avoid: Direct HTTP Requests and Third-party SDKs

```go
// Not recommended: Direct HTTP requests
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")

// Not recommended: Third-party SDKs
import "github.com/aliyun/aliyun-log-go-sdk"
```

### 2.3 API Pagination Logic Encapsulation

All pagination logic should be encapsulated in `*_api.go` files, with external callers not needing to handle pagination details:

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

### 2.4 Service Layer ID Encoding Specifications

Each Service file needs to define corresponding `Encode*Id` and `Decode*Id` functions:

```go
// EncodeJobId encodes workspace ID, namespace, and job ID into a single ID string
// Format: workspaceId:namespace:jobId
func EncodeJobId(workspaceId, namespace, jobId string) string {
    return fmt.Sprintf("%s:%s:%s", workspaceId, namespace, jobId)
}

// DecodeJobId parses job ID string into workspace ID, namespace, and job ID components
func DecodeJobId(id string) (string, string, string, error) {
    parts := strings.Split(id, ":")
    if len(parts) != 3 {
        return "", "", "", fmt.Errorf("invalid job ID format, expected workspaceId:namespace:jobId, got %s", id)
    }
    return parts[0], parts[1], parts[2], nil
}
```

### 2.5 Service Layer State Management Specifications

For each Resource, the corresponding Service needs to add `*StateRefreshFunc` state refresh functions and `WaitFor*` state synchronization functions:

```go
// StateRefreshFunc state refresh function
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

// WaitFor state synchronization function
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

## 3. Coding Specifications

### 3.1 Naming Conventions

#### 3.1.1 Resource and Data Source Naming
- Resources: `alicloud_<service>_<resource>`
- Data sources: `alicloud_<service>_<resource>s` (plural)
- Service names use lowercase underscores: `ecs`, `rds`, `slb`

#### 3.1.2 Function and Variable Naming
- Functions: camelCase (`resourceAlicloudEcs`)
- Variables: snake_case (`access_key`)
- ID fields: `resourceId`, name fields: `resourceName`
- Constants: uppercase underscores (`DEFAULT_TIMEOUT`)

#### 3.1.3 ID Field Naming Convention
All fields representing IDs should uniformly use `Id` instead of `ID`:

```go
// ✅ Correct
"WorkspaceId": workspace.WorkspaceId,
"UserId": user.UserId,

// ❌ Incorrect  
"WorkspaceID": workspace.WorkspaceID,
"UserID": user.UserID,
```

### 3.2 Resource Structure Requirements

All resources must include the following methods:

```go
func resourceAlicloudEcsInstance() *schema.Resource {
    return &schema.Resource{
        Create: resourceAlicloudEcsInstanceCreate,
        Read:   resourceAlicloudEcsInstanceRead,
        Update: resourceAlicloudEcsInstanceUpdate, // optional
        Delete: resourceAlicloudEcsInstanceDelete,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            // Resource attribute definitions
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
        },
    }
}
```

### 3.3 Error Handling Patterns

#### 3.3.1 Basic Error Handling

Recommend using error type judgment functions encapsulated in `alicloud/errors.go`, rather than `IsExpectedErrors`:

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

#### 3.3.2 Common Error Judgment Functions

```go
// Resource not found error judgment
if NotFoundError(err) {
    d.SetId("")
    return nil
}

// Resource already exists error judgment
if IsAlreadyExistError(err) {
    // Handle resource already exists case
    return resourceAlicloudServiceResourceRead(d, meta)
}

// Need retry error judgment
if NeedRetry(err) {
    time.Sleep(5 * time.Second)
    return resource.RetryableError(err)
}
```

#### 3.3.3 Specific Service Error Handling

For specific service errors, you can use predefined error code lists:

```go
// ECS related errors
if IsExpectedErrors(err, EcsNotFound) {
    d.SetId("")
    return nil
}

// SLB busy errors
if IsExpectedErrors(err, SlbIsBusy) {
    return resource.RetryableError(err)
}

// Database status errors
if IsExpectedErrors(err, OperationDeniedDBStatus) {
    return resource.RetryableError(err)
}
```

#### 3.3.4 Patterns to Avoid

Not recommended to use `IsExpectedErrors` directly for error judgment:

```go
// ❌ Not recommended approach
if IsExpectedErrors(err, []string{"InvalidInstance.NotFound", "Forbidden.InstanceNotFound"}) {
    d.SetId("")
    return nil
}

// ✅ Recommended approach
if NotFoundError(err) {
    d.SetId("")
    return nil
}
```

#### 3.3.5 Composite Error Handling Pattern

```go
if err != nil {
    // First check if it's a resource not found error
    if NotFoundError(err) {
        if !d.IsNewResource() {
            log.Printf("[DEBUG] Resource alicloud_service_resource DescribeResource Failed!!! %s", err)
            d.SetId("")
            return nil
        }
        return WrapError(err)
    }
    
    // Check if it's a resource already exists error (usually in Create operations)
    if IsAlreadyExistError(err) {
        // If resource already exists, read existing resource state
        return resourceAlicloudServiceResourceRead(d, meta)
    }
    
    // Check if retry is needed
    if NeedRetry(err) {
        return resource.RetryableError(err)
    }
    
    // Return other errors directly
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

#### 3.3.6 Error Handling Best Practices

1. **Prioritize using encapsulated error judgment functions**:
   - `NotFoundError(err)` - Check resource not found
   - `IsAlreadyExistError(err)` - Check resource already exists  
   - `NeedRetry(err)` - Check if retry is needed

2. **Use predefined error code lists**:
   - `EcsNotFound`, `SlbIsBusy`, `OperationDeniedDBStatus`, etc.

3. **Unified error wrapping**:
   - Use `WrapError(err)` or `WrapErrorf(err, msg, args...)` to wrap errors
   - Include detailed context information

4. **Appropriate logging**:
   - Record key error information for debugging
   - Distinguish different log levels (DEBUG, WARN, ERROR)

### 3.4 State Management Specifications

#### 3.4.1 Basic Rules
- **Prohibit direct Read function calls in Create functions**: Use StateRefreshFunc mechanism to wait for resource creation completion
- Use `d.SetId("")` when resource doesn't exist
- Set all computed attributes in the `Read` method
- Implement idempotent operations

#### 3.4.2 Correct State Refresh Pattern

```go
// ✅ Correct: Use Service layer's WaitFor function in Create function to wait for resource readiness
err = service.WaitForServiceResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
if err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// Finally call Read to synchronize state
return resourceAlicloudServiceResourceRead(d, meta)
```

### 3.5 Data Validation and Conversion

#### 3.5.1 Input Validation
```go
"instance_type": {
    Type:         schema.TypeString,
    Required:     true,
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
},
```

#### 3.5.2 Type Conversion
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

## 4. Resource Development Guide

### 4.1 Standard Import Package Structure

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

### 4.2 Schema Definition Specifications

#### 4.2.1 Field Type Definitions

```go
// Required field
"instance_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The ID of the instance.",
},

// Optional field
"name": {
    Type:        schema.TypeString,
    Optional:    true,
    Description: "The name of the resource.",
},

// Computed field
"status": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "The status of the resource.",
},

// Optional + Computed field
"instance_name": {
    Type:        schema.TypeString,
    Optional:    true,
    Computed:    true,
    Description: "The name of the instance.",
},
```

#### 4.2.2 Nested Object Definition

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

### 4.3 CRUD Operation Implementation

#### 4.3.1 Create Method

```go
func resourceAliCloudServiceResourceCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewServiceService(client)
    if err != nil {
        return WrapError(err)
    }

    // Build request object
    request := &serviceAPI.CreateResourceRequest{
        InstanceId: d.Get("instance_id").(string),
        Name:       d.Get("name").(string),
    }

    // Use Retry to create resource
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

    // Wait for resource readiness
    err = service.WaitForServiceResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
    if err != nil {
        return WrapErrorf(err, IdMsg, d.Id())
    }

    // Finally call Read to synchronize state
    return resourceAlicloudServiceResourceRead(d, meta)
}
```

#### 4.3.2 Read Method

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

    // Set all necessary fields
    d.Set("instance_id", object.InstanceId)
    d.Set("name", object.Name)
    d.Set("status", object.Status)

    return nil
}
```

#### 4.3.3 Delete Method

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

    // Wait for resource deletion
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

### 4.4 Resource Creation resource.Retry Logic

```go
// Common retryable errors
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

## 11. Development Checklist

### 11.1 Basic Requirements
- [ ] Use standard import package structure
- [ ] Call CWS-Lib-Go API through Service layer
- [ ] Correctly define Schema (Required/Optional/Computed)
- [ ] Implement all necessary CRUD methods
- [ ] Add appropriate Timeout configuration

### 11.2 State Management
- [ ] Use Service layer's WaitFor function after Create to wait for resource readiness
- [ ] Use Service layer's WaitFor function after Delete to wait for resource deletion
- [ ] Correctly set all fields in Read method
- [ ] Finally call Read to synchronize state

### 11.3 Code Quality
- [ ] Follow ID field naming convention (use Id instead of ID)
- [ ] Add appropriate Description
- [ ] Correctly handle type conversion for complex objects
- [ ] Reasonable logging

## 12. Summary

This guide serves as the core development specification for the Terraform Provider Alicloud project. During development, it should be strictly followed, especially:

1. **Architectural Layering Principles**: Strictly follow Provider → Resource/DataSource → Service → API hierarchy
2. **State Management Best Practices**: Correctly use StateRefreshFunc, avoid direct Read function calls
3. **Error Handling Specifications**: Unified error handling patterns and appropriate error classification
4. **Test Coverage**: Ensure adequate unit testing and integration testing
5. **Documentation Completeness**: Provide clear, complete usage documentation and examples

Following these specifications will help improve code quality, reduce maintenance costs, and provide users with a better experience.