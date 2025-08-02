# Terraform Provider Alicloud Development Guide

## Overview

Development guide for Terraform Provider Alicloud project with AI code generation specifications, architectural design principles, coding standards, and best practices. Ensures code quality, consistency, and maintainability.

**Note**: In all examples, `Xxx` represents the resource name (e.g., Flink, Ecs, Rds, Slb).

## 1. Code Generation Specifications

### 1.1 Information Reference
- Reference README.md for basic project information
- Check docs directory for detailed documentation
- Review examples directory for usage patterns

### 1.2 Task Workflow
- Create TODO.md for complex tasks with step-by-step plans
- Update TODO.md after completing each item
- Verify completion before task closure
- Use phased approach for large refactoring

### 1.3 File Operations
- Generate scripts for complex file operations
- Always backup before batch operations
- Track changes with version control

### 1.4 Language Standards
- Use English for all documentation and comments
- Maintain internationalization compatibility

### 1.5 Code Splitting
- Split code files exceeding 1000 lines
- Data files (.json, .yaml, .csv, .xml) exempt
- Ensure single responsibility per file

## 2. Architecture Design

### 2.1 Layered Architecture

Resource/DataSource layer calls Service layer functions, not SDK directly.

**Hierarchy:**
```
Provider Layer (alicloud/)
├── Resource Layer (resource_alicloud_*.go)
├── DataSource Layer (data_source_alicloud_*.go)
└── Service Layer (service_alicloud_*.go)
    └── API Layer (CWS-Lib-Go)
        └── SDK Layer (Alibaba Cloud Official SDK)
```

Service layer methods:
1. CRUD and state refresh methods
2. EncodeXxxId/DecodeXxxId for ID handling
3. WaitForXxx for state synchronization

### 2.2 Service Layer API Calls

#### 2.2.1 ✅ Recommended: CWS-Lib-Go Wrapper

```go
// Create Service object
service, err := NewXxxService(client)
if err != nil {
    return WrapError(err)
}

// Call Service methods
resource, err := service.CreateXxx(request)
if err != nil {
    return WrapError(err)
}
```

**Benefits:** Unified interfaces, error handling, type safety, retry mechanisms

#### 2.2.2 ❌ Avoid: Direct SDK/HTTP Calls

```go
// Not recommended
response, err := client.RpcPost("service", "version", "Action", params, "")
```

### 2.3 Pagination Encapsulation

Encapsulate pagination in Service layer with MAX_RESOURCE_COUNT limit:

```go
func (s *XxxService) DescribeXxxs(request *XxxRequest) ([]*Xxx, error) {
    var allResources []*Xxx
    // Pagination logic with MAX_RESOURCE_COUNT limit
    for pageNumber := 1; ; pageNumber++ {
        // Limit check, pagination request, response handling...
        response, err := s.client.DescribeXxxs(request)
        if err != nil {
            return nil, err
        }
        allResources = append(allResources, response.Resources...)
        // Break condition...
    }
    return allResources, nil
}
```

### 2.4 ID Encoding/Decoding

Define EncodeXxxId/DecodeXxxId functions:

```go
// DecodeXxxId parses combined ID string
func DecodeXxxId(id string) (string, string, string, error) {
    parts := strings.Split(id, ":")
    if len(parts) != 3 {
        return "", "", "", fmt.Errorf("invalid ID format: %s", id)
    }
    return parts[0], parts[1], parts[2], nil
}
```

### 2.5 State Management

Implement StateRefreshFunc and WaitForXxx functions:

```go
// State refresh function
func (s *XxxService) XxxStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
    return func() (interface{}, string, error) {
        object, err := s.DescribeXxx(id)
        if err != nil {
            if IsNotFoundError(err) {
                return nil, "", nil
            }
            return nil, "", WrapError(err)
        }

        currentStatus := object.Status
        for _, failState := range failStates {
            if currentStatus == failState {
                return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
            }
        }
        return object, currentStatus, nil
    }
}

// Wait for resource state
func (s *XxxService) WaitForXxxCreating(id string, timeout time.Duration) error {
    stateConf := BuildStateConf(
        []string{"CREATING", "STARTING"}, // pending states
        []string{"RUNNING", "ACTIVE"},    // target states
        timeout,
        5*time.Second,
        s.XxxStateRefreshFunc(id, []string{"FAILED", "ERROR"}),
    )
    
    _, err := stateConf.WaitForState()
    return WrapErrorf(err, IdMsg, id)
}
```

### 2.6 Service Base Files

Create `service_alicloud_<service>_base.go` for new services:

#### Structure Standards
- Package declaration
- Imports (stdlib + third-party + project)
- Service struct definition
- NewXxxService constructor
- Utility methods

#### Service Struct Pattern
```go
type XxxService struct {
    client *connectivity.AliyunClient
    xxxAPI *aliyunXxxAPI.XxxAPI
}
```

#### Constructor Pattern
```go
func NewXxxService(client *connectivity.AliyunClient) (*XxxService, error) {
    credentials := &aliyunCommonAPI.Credentials{
        AccessKey:     client.AccessKey,
        SecretKey:     client.SecretKey,
        RegionId:      client.RegionId,
        SecurityToken: client.SecurityToken,
    }

    xxxAPI, err := aliyunXxxAPI.NewXxxAPI(credentials)
    if err != nil {
        return nil, fmt.Errorf("failed to create cws-lib-go XxxAPI: %w", err)
    }

    return &XxxService{
        client: client,
        xxxAPI: xxxAPI,
    }, nil
}
```

#### Utility Functions
```go
func (service *XxxService) GetAPI() (*aliyunXxxAPI.XxxAPI, error) {
    // add some customize logic for this API object
    return service.xxxAPI, nil
}
```

#### Import Order
```go
import (
    "fmt"  // Standard library

    "github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"  // Project
    aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"  // CWS-Lib-Go common
    aliyunXxxAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/<service>"  // CWS-Lib-Go service
)
```

## 3. Naming Conventions

### 3.1 Resource/DataSource Naming
- Resources: `alicloud_<service>_<resource>`
- Data sources: `alicloud_<service>_<resource>s` (plural)
- Service names: lowercase with underscores (`ecs`, `rds`, `slb`)

### 3.2 Function/Variable Naming
- Functions: camelCase (`resourceAlicloudXxx`)
- Variables: snake_case (`access_key`)
- ID fields: `resourceId`, name fields: `resourceName`
- Constants: UPPER_SNAKE_CASE (`DEFAULT_TIMEOUT`)

### 3.3 ID Field Convention
Use `Id` instead of `ID`:

```go
// ✅ Correct
"WorkspaceId": workspace.WorkspaceId,
"UserId": user.UserId,

// ❌ Incorrect  
"WorkspaceID": workspace.WorkspaceID,
"UserID": user.UserID,
```

### 3.4 Resource Structure
All resources must include:

```go
func resourceAlicloudXxxResource() *schema.Resource {
    return &schema.Resource{
        Create: resourceAlicloudXxxResourceCreate,
        Read:   resourceAlicloudXxxResourceRead,
        Update: resourceAlicloudXxxResourceUpdate, // optional
        Delete: resourceAlicloudXxxResourceDelete,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            // Resource attributes...
        },
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(5 * time.Minute),
        },
    }
}
```

## 4. Error Handling

### 4.1 Basic Error Handling
Use error judgment functions from `alicloud/errors.go`:

```go
if err != nil {
    if IsNotFoundError(err) {
        log.Printf("[WARN] Resource (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

### 4.2 Common Error Functions

```go
// Resource not found
if IsNotFoundError(err) {
    d.SetId("")
    return nil
}

// Resource already exists
if IsAlreadyExistError(err) {
    return resourceAlicloudXxxResourceRead(d, meta)
}

// Retry needed
if NeedRetry(err) {
    time.Sleep(5 * time.Second)
    return resource.RetryableError(err)
}
```

### 4.3 Service-Specific Errors
Use predefined error code lists:

```go
// Service-specific errors
if IsExpectedErrors(err, XxxNotFound) {
    d.SetId("")
    return nil
}

if IsExpectedErrors(err, XxxBusy) {
    return resource.RetryableError(err)
}
```

### 4.4 Avoid Direct IsExpectedErrors
```go
// ❌ Not recommended
if IsExpectedErrors(err, []string{"InvalidInstance.NotFound", "Forbidden.InstanceNotFound"}) {
    d.SetId("")
    return nil
}

// ✅ Recommended
if IsNotFoundError(err) {
    d.SetId("")
    return nil
}
```

### 4.5 Composite Error Pattern

```go
if err != nil {
    // Check resource not found
    if IsNotFoundError(err) {
        if !d.IsNewResource() {
            log.Printf("[DEBUG] Resource alicloud_xxx_resource DescribeResource Failed!!! %s", err)
            d.SetId("")
            return nil
        }
        return WrapError(err)
    }
    
    // Check already exists (Create operations)
    if IsAlreadyExistError(err) {
        return resourceAlicloudXxxResourceRead(d, meta)
    }
    
    // Check retry needed
    if NeedRetry(err) {
        return resource.RetryableError(err)
    }
    
    // Return other errors
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

## 5. State Management

### 5.1 Basic Rules
- **No direct Read calls in Create**: Use StateRefreshFunc
- Use `d.SetId("")` when resource doesn't exist
- Set all computed attributes in Read method
- Implement idempotent operations

### 5.2 Correct State Pattern

```go
// ✅ Use Service WaitFor function in Create
err = service.WaitForXxxResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
if err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// Finally call Read to sync state
return resourceAlicloudXxxResourceRead(d, meta)
```

## 6. Schema Definition

### 6.1 Field Types

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

// Optional + Computed
"instance_name": {
    Type:        schema.TypeString,
    Optional:    true,
    Computed:    true,
    Description: "The name of the instance.",
},
```

### 6.2 Nested Objects

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
            // ... other fields
        },
    },
    Description: "Configuration parameters.",
},
```

### 6.3 Input Validation
```go
"instance_type": {
    Type:         schema.TypeString,
    Required:     true,
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
},
```

### 6.4 Type Conversion
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

## 7. CRUD Implementation

### 7.1 Standard Imports

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

### 7.2 Create Method

```go
func resourceAliCloudXxxResourceCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewXxxService(client)
    if err != nil {
        return WrapError(err)
    }

    // Build request
    request := &xxxAPI.CreateResourceRequest{
        InstanceId: d.Get("instance_id").(string),
        Name:       d.Get("name").(string),
    }

    // Use Retry for creation
    var result *xxxAPI.Resource
    err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
        resp, err := service.CreateResource(request)
        if err != nil {
            // Error handling logic...
            return resource.RetryableError(err)
        }
        result = resp
        return nil
    })

    if err != nil {
        return WrapErrorf(err, DefaultErrorMsg, "alicloud_xxx_resource", "CreateResource", AlibabaCloudSdkGoERROR)
    }

    d.SetId(result.ResourceId)

    // Wait for readiness 
    err = service.WaitForXxxResourceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
    if err != nil {
        return WrapErrorf(err, IdMsg, d.Id())
    }

    // Sync state
    return resourceAlicloudXxxResourceRead(d, meta)
}
```

### 7.3 Read Method

```go
func resourceAliCloudXxxResourceRead(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewXxxService(client)
    if err != nil {
        return WrapError(err)
    }

    object, err := service.DescribeResource(d.Id())
    if err != nil {
        if !d.IsNewResource() && IsNotFoundError(err) {
            log.Printf("[DEBUG] Resource alicloud_xxx_resource DescribeResource Failed!!! %s", err)
            d.SetId("")
            return nil
        }
        return WrapError(err)
    }

    // Set all fields
    d.Set("instance_id", object.InstanceId)
    d.Set("name", object.Name)
    d.Set("status", object.Status)

    return nil
}
```

### 7.4 Delete Method

```go
func resourceAliCloudXxxResourceDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*connectivity.AliyunClient)
    service, err := NewXxxService(client)
    if err != nil {
        return WrapError(err)
    }

    err = service.DeleteResource(d.Id())
    if err != nil {
        if IsNotFoundError(err) {
            return nil
        }
        return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteResource", AlibabaCloudSdkGoERROR)
    }

    // Wait for deletion
    stateConf := &resource.StateChangeConf{
        Pending: []string{"Deleting"},
        Target:  []string{""},
        Refresh: func() (interface{}, string, error) {
            obj, err := service.DescribeResource(d.Id())
            if err != nil {
                if IsNotFoundError(err) {
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

### 7.5 Retry Logic

```go
// Common retryable errors
var retryableErrors = []string{
    "ServiceUnavailable",
    "ThrottlingException", 
    "InternalError",
    // ... other errors
}

err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
    _, err := service.CreateResource(request)
    if err != nil {
        if IsExpectedErrors(err, retryableErrors) {
            // Retry with backoff...
            return resource.RetryableError(err)
        }
        return resource.NonRetryableError(err)
    }
    return nil
})
```

## 8. Development Checklist

### 8.1 Basic Requirements
- [ ] Use standard import package structure
- [ ] Call CWS-Lib-Go API through Service layer
- [ ] Correctly define Schema (Required/Optional/Computed)
- [ ] Implement all necessary CRUD methods
- [ ] Add appropriate Timeout configuration

### 8.2 State Management
- [ ] Use Service layer's WaitFor function after Create to wait for resource readiness
- [ ] Use Service layer's WaitFor function after Delete to wait for resource deletion
- [ ] Correctly set all fields in Read method
- [ ] Finally call Read to synchronize state

### 8.3 Code Quality
- [ ] Follow ID field naming convention (use Id instead of ID)
- [ ] Add appropriate Description
- [ ] Correctly handle type conversion for complex objects
- [ ] Reasonable logging

## 9. Summary

This guide serves as the core development specification for the Terraform Provider Alicloud project. During development, it should be strictly followed, especially:

1. **Architectural Layering Principles**: Strictly follow Provider → Resource/DataSource → Service → API hierarchy
2. **State Management Best Practices**: Correctly use StateRefreshFunc, avoid direct Read function calls
3. **Error Handling Specifications**: Unified error handling patterns and appropriate error classification
4. **Test Coverage**: Ensure adequate unit testing and integration testing
5. **Documentation Completeness**: Provide clear, complete usage documentation and examples

Following these specifications will help improve code quality, reduce maintenance costs, and provide users with a better experience.