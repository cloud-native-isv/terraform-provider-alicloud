## Code Generation Guidelines
- Refer to the contents of the README.md file in the project directory to understand the basic information of the project. If there is a docs directory in the project root directory, refer to the contents within it.
- For complex tasks, create a TODO.md file first to list the plan and steps, then execute step by step. Update the corresponding records in the TODO.md document each time a step is completed, and check whether all items in TODO.md are completed after the task is finished.
- For complex file operations, first generate a Python or shell script, then perform the operations by executing the script.
- Generate documentation in English, and use English for code comments and logs.
- When programming language code files (*.go, *.java, *.py, *.ts, *.js, *.c, *.cpp, *.cs, *.php, *.rb, *.rs, etc.) exceed 1500 lines, they should be split to improve code maintainability and readability. Data files (*.json, *.yaml, *.csv, *.xml, etc.) are not subject to this limitation.

# Terraform Provider Alicloud Development Guide

## Service Layer API Calling Standards

The service layer in the current project contains three ways of calling underlying APIs:

### 1. client.RpcPost Direct HTTP Requests (❌ Not Recommended, Should Be Deprecated)
```go
// Not recommended approach
response, err := client.RpcPost("ecs", "2014-05-26", "DescribeInstances", parameters, "")
```
**Issues:**
- High maintenance cost
- Cloud service API updates cause extensive cascading modifications in code
- Lack of type safety and automated code generation
- Complex error handling

### 2. Third-party SDK Calls (⚠️ Not Recommended)
```go
// Not recommended approach
import "github.com/aliyun/aliyun-log-go-sdk"
import "github.com/aliyun/aliyun-oss-go-sdk"
```
**Issues:**
- Third-party SDKs lack maintenance and updates
- Version compatibility problems
- Security and stability cannot be guaranteed

### 3. CWS-Lib-Go Wrapper Calls (✅ Recommended)
```go
// Recommended approach
import "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
```
**Advantages:**
- Secondary encapsulation using official SDK
- Provides unified API interfaces and error handling
- Regular maintenance and updates
- Type safety and code generation support
- Better test coverage

**Usage Example:**
```go
func (s *EcsService) DescribeInstances(request *api.DescribeInstancesRequest) (*api.DescribeInstancesResponse, error) {
    client := s.client.WithApiInfo("ecs", "2014-05-26", "DescribeInstances")
    response := &api.DescribeInstancesResponse{}
    err := client.DoAction(request, response)
    return response, err
}
```

## Code Generation Rules
- Reference README.md file contents in the directory
- Create TODO.md for complex tasks first, then generate code
- No error fixes needed after large file generation
- Use English for documentation, code comments and logs

## Naming Conventions
- Resources: `alicloud_<service>_<resource>`
- Data Sources: `alicloud_<service>_<resource>s` (plural)
- Functions: camelCase (`resourceAlicloudEcs`)
- Variables: snake_case (`access_key`)
- ID variables should be resourceID
- Name variables should be resourceName
- Variable identifiers must clearly indicate whether they are ID or Name

## Resource Structure
Must include: `Create`, `Read`, `Update`, `Delete`, `Schema`

## Error Handling Pattern
```go
if err != nil {
    if NotFoundError(err) {
        d.SetId("")
        return nil
    }
    return WrapError(err)
}
```

## State Management
- **Prohibit direct Read function calls in Create function**: Use StateRefreshFunc mechanism to wait for resource creation completion
- Use `d.SetId("")` when resource doesn't exist
- Set all computed attributes in `Read` method

### State Refresh Best Practices
```go
// Correct approach: Use StateRefreshFunc in Create function to wait for resource readiness
stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, service.ResourceStateRefreshFunc(id, []string{}))
if _, err := stateConf.WaitForState(); err != nil {
    return WrapErrorf(err, IdMsg, d.Id())
}

// Finally call Read to sync state
return resourceAlicloudServiceResourceRead(d, meta)
```

```go
// Wrong approach: Direct Read function call in Create function
func resourceCreate(d *schema.ResourceData, meta interface{}) error {
    // ... create resource ...
    d.SetId(id)
    
    // ❌ Error: Should not call Read function directly
    return resourceRead(d, meta)
}
```

## Documentation Requirements
Include: resource description, parameter explanation, attribute description, import guide, usage examples

## Coding Standards
- Use `gofmt` formatting
- Follow Go conventions
- Avoid duplicate code
- Add meaningful comments
