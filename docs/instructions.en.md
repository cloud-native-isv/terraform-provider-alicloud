# Terraform Provider Alicloud Development Guide

## Overview

This document serves as the development guide for the Terraform Provider Alicloud project, including large language model code generation specifications, architectural design principles, coding standards, and best practices. This guide aims to ensure code quality, consistency, and maintainability, providing developers with comprehensive technical specifications.

## 1. Code Generation Specifications

### 1.1 Project Information Reference
- Refer to the contents of the README.md file in the project directory to understand the basic information of the project
- If there is a docs directory in the project root directory, refer to the contents within it
- Check the examples directory to understand actual usage scenarios and patterns

### 1.2 Task Execution Process
- For complex tasks, create a TODO.md file first to list the plan and steps, then execute step by step
- Update the corresponding records in the TODO.md document each time a step is completed
- Check whether all items in TODO.md are completed after the task is finished
- For large refactoring tasks, it is recommended to proceed in phases and record verification points for each phase

### 1.3 File Operation Specifications
- For complex file operations, first generate a Python or shell script, then perform the operations by executing the script
- Ensure backup before batch operations
- Use version control to track all changes

### 1.4 Language Usage Specifications
- Generate documentation in Chinese
- Use English for code comments and logs
- Use English for API documentation and error messages to maintain internationalization compatibility

### 1.5 Code Splitting Specifications
- When programming language code files (*.go, *.java, *.py, *.ts, *.js, *.c, *.cpp, *.cs, *.php, *.rb, *.rs, etc.) exceed 1000 lines, they should be split to improve code maintainability and readability
- Data files (*.json, *.yaml, *.csv, *.xml, etc.) are not subject to this limitation
- Split by functional modules, ensuring each file has a single and clear responsibility

## 2. Architectural Design Principles

### 2.1 Layered Architecture Design

Resource or datasource layers should call functions provided by the service layer, rather than directly calling underlying SDK or API functions.

**Architecture Hierarchy:**
```
Provider Layer (alicloud/)
├── Resource Layer (resource_alicloud_*.go)
├── DataSource Layer (data_source_alicloud_*.go)
└── Service Layer (service_alicloud_*.go)
    └── API Layer (CWS-Lib-Go/Alibaba Cloud Official SDK)
```

For adding or refactoring an existing resource or datasource, you need to first define its service layer, and the resource or datasource should only depend on functions provided by the service layer.

The service layer may contain one or more Go files, including CRUD methods and state refresh methods for resource objects.

### 2.2 Service Layer API Calling Specifications

The current project's service layer contains three ways to call underlying APIs:

#### 2.2.1 client.RpcPost Direct HTTP Requests (❌ Not Recommended, Should Be Deprecated)

```go
// Not recommended approach
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")
```

**Problems:**
- High maintenance cost
- When cloud service APIs are updated, it will cause a large number of cascading changes
- Lack of type safety and automated code generation
- Complex error handling
- Lack of request retry and circuit breaker mechanisms

#### 2.2.2 Third-party SDK Calls (⚠️ Not Recommended)

```go
// Not recommended approach
import "github.com/aliyun/aliyun-log-go-sdk"
import "github.com/aliyun/aliyun-oss-go-sdk"
```

**Problems:**
- Third-party SDKs lack maintenance and updates
- Version compatibility issues
- Security and stability cannot be guaranteed
- May contain security vulnerabilities

#### 2.2.3 CWS-Lib-Go Encapsulated Calls (✅ Recommended)

```go
// Recommended approach
import "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
```

**Advantages:**
- Uses official SDK for secondary encapsulation
- Provides unified API interfaces and error handling
- Regular maintenance and updates
- Type safety and code generation support
- Better test coverage
- Built-in retry mechanisms and error recovery
- Unified authentication and configuration management

### 2.3 API Pagination Logic Encapsulation

All pagination/batching logic should be encapsulated in `*_api.go` files, and external callers should not need to handle pagination details. The API layer should provide simple methods to abstract pagination complexity, page size management, and result aggregation.

```go
// Recommended approach: Encapsulate pagination logic
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

## 3. Coding Specifications

### 3.1 Naming Conventions

#### 3.1.1 Resource and Data Source Naming
- Resources: `alicloud_<service>_<resource>`
- Data sources: `alicloud_<service>_<resource>s` (plural)
- Service names use lowercase with underscores: `ecs`, `rds`, `slb`
- Resource names should be concise and clear, avoiding redundancy

**Examples:**
```
alicloud_ecs_instance
alicloud_rds_instance
alicloud_slb_load_balancer
alicloud_ecs_instances (data source)
```

#### 3.1.2 Function and Variable Naming
- Functions: camelCase (`resourceAlicloudEcs`)
- Variables: snake_case (`access_key`)
- Variables representing IDs should be `resourceId`
- Variables representing Names should be `resourceName`
- Variable identifiers must clearly indicate whether they are ID or Name
- Constants use uppercase with underscores: `DEFAULT_TIMEOUT`

#### 3.1.3 ID Field Naming Convention
All fields representing IDs should consistently use the variable name `Id` instead of `ID` to maintain compatibility with many automated generation tools.

**Examples:**
- `WorkspaceId`
- `UserId`
- `ResourceId`
- `InstanceId`

### 3.2 Resource Structure Requirements

All resources must include the following methods:
- `Create` - Create resource
- `Read` - Read resource state
- `Update` - Update resource (if supported)
- `Delete` - Delete resource
- `Schema` - Resource schema definition
- `Importer` - Resource import support (recommended)

**Resource File Structure Example:**
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

```go
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}
```

**Enhanced Error Handling:**
```go
// Recommended error handling pattern
if err != nil {
    if IsExpectedErrors(err, []string{"InvalidInstance.NotFound", "Forbidden.InstanceNotFound"}) {
        log.Printf("[WARN] Resource (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }
    return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
}
```

### 3.4 State Management Specifications

#### 3.4.1 Basic Rules
- **Prohibit directly calling Read function in Create function**: Use StateRefreshFunc mechanism to wait for resource creation completion
- Use `d.SetId("")` when resource does not exist
- Set all computed attributes in `Read` method
- Use appropriate timeout values to avoid infinite waiting
- Implement idempotent operations

#### 3.4.2 State Refresh Best Practices

**Correct Approach:**

```go
// Use StateRefreshFunc in Create function to wait for resource readiness
stateConf := BuildStateConf([]string{"Pending", "Starting"}, []string{"Running", "Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.ResourceStateRefreshFunc(id, []string{"Failed", "Error"}))
if _, err := stateConf.WaitForState(); err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// Finally call Read to sync state
return resourceAlicloudServiceResourceRead(d, meta)
```

**Wrong Approach:**

```go
// Wrong approach: Directly calling Read in Create function
func resourceCreate(d *schema.ResourceData, meta interface{}) error {
    // ... create resource ...
    d.SetId(id)
    
    // ❌ Wrong: Should not directly call Read function
    return resourceRead(d, meta)
}
```

### 3.5 Data Validation and Conversion

#### 3.5.1 Input Validation
```go
// Validation in Schema
"instance_type": {
    Type:         schema.TypeString,
    Required:     true,
    ValidateFunc: validation.StringMatch(regexp.MustCompile(`^ecs\..+`), "instance_type must start with 'ecs.'"),
},
```

#### 3.5.2 Type Conversion
```go
// Safe type conversion
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

### 3.6 Code Formatting Specifications
- Use `gofmt` for formatting
- Follow Go conventions
- Avoid duplicate code
- Add meaningful comments
- Functions should not exceed 50 lines; consider splitting if they do
- Use consistent error handling patterns

## 4. Testing Specifications

### 4.1 Unit Testing Requirements
- Each service function should have corresponding unit tests
- Test coverage should reach 80% or higher
- Use Mock objects to test external dependencies

### 4.2 Integration Testing Specifications
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

## 5. Documentation Requirements

### 5.1 Resource Documentation Structure
Each resource and data source documentation must include:
- **Resource Description**: Brief explanation of the resource's purpose and functionality
- **Parameter Description**: Detailed description of all input parameters, including type, whether required, default values
- **Attribute Description**: Description of all output attributes
- **Import Guide**: How to import existing resources
- **Usage Examples**: Complete, runnable example code
- **Notes**: Important usage limitations and considerations

### 5.2 Example Code Specifications
```hcl
# Basic example
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

## 6. Performance Optimization Guide

### 6.1 API Call Optimization
- Reasonably use batch APIs to reduce network requests
- Implement client caching mechanisms
- Use connection pools to manage HTTP connections
- Set reasonable timeout values

### 6.2 State Refresh Optimization
- Use exponential backoff algorithm for retries
- Avoid overly frequent state checks
- Set reasonable state check intervals

## 7. Security Best Practices

### 7.1 Credential Management
- Do not hard-code access keys in code
- Support multiple authentication methods: environment variables, configuration files, RAM roles
- Implement access key rotation mechanisms

### 7.2 Access Control
- Follow the principle of least privilege
- Clearly mark sensitive fields
- Implement resource access control

## 8. Troubleshooting Guide

### 8.1 Common Issues
- **API Rate Limiting**: Implement retry mechanisms and backoff strategies
- **Network Timeout**: Check network connections and firewall settings
- **Permission Errors**: Verify RAM role and policy configurations
- **Resource Not Found**: Properly handle NotFound errors

### 8.2 Debugging Techniques
- Enable detailed logging
- Use environment variables to control debug levels
- Log key API calls and responses

## 9. Version Compatibility

### 9.1 Backward Compatibility
- Mark new fields as Optional
- Maintain backward compatibility for deprecated fields
- Provide migration guides

### 9.2 Version Management
- Follow semantic versioning
- Maintain CHANGELOG documentation
- Provide upgrade guidance

## 10. Summary

This guide serves as the core development specifications for the Terraform Provider Alicloud project, aiming to ensure code quality, consistency, and maintainability. During development, these specifications should be strictly followed, particularly:

1. **Architectural Layering Principles**: Strictly follow the Provider → Resource/DataSource → Service → API hierarchy
2. **State Management Best Practices**: Properly use StateRefreshFunc, avoid directly calling Read functions
3. **Error Handling Specifications**: Unified error handling patterns and appropriate error classification
4. **Test Coverage**: Ensure adequate unit testing and integration testing
5. **Documentation Completeness**: Provide clear, complete usage documentation and examples

Following these specifications will help improve code quality, reduce maintenance costs, and provide users with a better experience.