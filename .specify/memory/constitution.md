<!--
## Sync Impact Report

**Version change**: 1.3.1 → 1.4.0
**Modified principles**:
- V. Strong Typing with CWS-Lib-Go: Wording tightened, formatting normalized
- VI. Testing and Validation Requirements: Clarified validation scope and pagination rule placement
**Added sections**:
- VII. Feature-Centric Development
- Governance: Amendment procedure, versioning policy, compliance review expectations

**Removed sections**: None

**Templates requiring updates**:
✅ .specify/templates/plan-template.md – Constitution Check updated to match current principles
✅ .specify/templates/tasks-template.md – Constitution reference updated
✅ .specify/templates/spec-template.md – No change required (aligned with Constitution)

-->

# Terraform Provider Alicloud Constitution

## Core Principles

### I. Architecture Layering Principle
Resource or DataSource layers MUST call functions provided by the Service layer, NOT directly call underlying SDK or API functions. The architecture hierarchy is strictly: Provider Layer → Resource/DataSource Layer → Service Layer → API Layer (CWS-Lib-Go) → SDK Layer (Alibaba Cloud official SDK). Service layers contain Go files with CRUD methods and state refresh methods for resource objects.

Service layer API calls MUST use CWS-Lib-Go encapsulation:
- ✅ RECOMMENDED: Use `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api` imports
- ❌ AVOID: Direct HTTP requests or third-party SDKs like `github.com/aliyun/aliyun-log-go-sdk`

### II. State Management Best Practices
State management MUST follow proper patterns: NEVER call Read functions directly in Create functions; use StateRefreshFunc mechanisms to wait for resource creation completion; use d.SetId("") when resources don't exist; set all computed properties in Read methods; implement idempotent operations; use Service layer WaitFor functions to wait for resource readiness after Create/Delete operations.

Service layer MUST implement proper state refresh and wait functions:
- `*StateRefreshFunc` for state polling with fail state handling
- `WaitFor*` functions using `BuildStateConf` with proper pending/target states
- Timeout configurations aligned with resource timeouts

### III. Error Handling Standardization
Error handling MUST use encapsulated error judgment functions from alicloud/errors.go rather than IsExpectedErrors directly. Priority order: IsNotFoundError(err) for resource not found, IsAlreadyExistError(err) for resource already exists, NeedRetry(err) for retryable errors. Use predefined error code lists (EcsNotFound, SlbIsBusy, OperationDeniedDBStatus) for service-specific errors. Always wrap errors using WrapError(err) or WrapErrorf(err, msg, args...) with detailed context.

Retry logic MUST handle common retryable errors:
- `ServiceUnavailable`, `ThrottlingException`, `InternalError`
- `Throttling`, `SystemBusy`, `OperationConflict`
- Use `resource.Retry` with proper timeout handling

### IV. Code Quality and Consistency
All code MUST follow strict naming conventions: Resources use alicloud_<service>_<resource> format, Data sources use plural form alicloud_<service>_<resource>s, service names use lowercase underscore (ecs, rds, slb). Functions use camelCase, variables use snake_case, ID fields use resourceId format, constants use uppercase underscore. All ID fields MUST use Id not ID (e.g., WorkspaceId not WorkspaceID). All schema fields MUST include appropriate Description.

Service layer MUST implement proper ID encoding/decoding:
- `Encode*Id` functions format: `workspaceId:namespace:jobId`
- `Decode*Id` functions with proper error handling for invalid formats
- Consistent ID handling across all service operations

### V. Strong Typing with CWS-Lib-Go
Implementations MUST prefer strong types provided by CWS-Lib-Go over weakly typed
structures such as `map[string]interface{}` or untyped `interface{}` payloads. This
requirement applies across Service and API layers to ensure type safety, maintainability,
and clearer contracts.

- MUST use generated/defined structs and enums from `github.com/cloud-native-tools/cws-lib-go`
- wherever applicable.
- MUST NOT introduce new usages of `map[string]interface{}` for request/response shapes,
  except when interacting with legacy code paths.
- Legacy code is exempt (read-only, minimal-touch). Any refactoring SHOULD migrate to
- strong types opportunistically while maintaining backward compatibility.
- Code reviews MUST flag weak typing in new/modified code unless explicitly justified
- (e.g., bridging adapters to third-party libs not yet modeled in cws-lib-go).

### VI. Testing and Validation Requirements
Every code change MUST be validated by executing 'cd /cws_data/terraform-provider-alicloud && make' to ensure syntax correctness and successful compilation. Comprehensive unit tests and integration tests are mandatory. All resources MUST include proper Timeout configurations. Code files exceeding 1000 lines MUST be split by functional modules to ensure single responsibility.

Binary generation MUST NOT occur in the root directory. All binary files MUST be output to the `bin` directory and ignored by `.gitignore`.

API pagination logic MUST be encapsulated in `*_api.go` files:
- External callers MUST NOT handle pagination details
- Use page number/page size iteration until all results are collected
- Return complete result sets to callers

### VII. Feature-Centric Development
Feature 是项目的长期核心框架：
- Feature 列表必须保持为项目的“单一事实来源”。
- 在 spec → plan → tasks → implement 的每个阶段都必须复核 Feature 的新增/合并/拆分/删除。
- Feature 变更必须可追溯到相应的 spec/plan 依据，并记录在 Feature 详情中。

## Development Workflow Standards

All complex tasks MUST create a TODO.md file listing plans and steps, then execute step by step with updates to the TODO.md after each completion. Large refactoring tasks SHOULD be performed in phases with validation checkpoints recorded. Complex file operations SHOULD generate Python or Shell scripts first, then execute scripts. Batch operations MUST be backed up before execution. All changes MUST be tracked using version control.

## Quality Assurance Requirements

Documentation MUST be generated in Chinese, while code comments and logs MUST use English to maintain international compatibility for API documentation and error messages. Programming language code files (*.go, *.java, *.py, *.ts, *.js, *.c, *.cpp, *.cs, *.php, *.rb, *.rs, etc.) exceeding 1000 lines MUST be split by functional modules. Data files (*.json, *.yaml, *.csv, *.xml, etc.) are exempt from this restriction. All schema definitions MUST properly use Required/Optional/Computed fields with appropriate validation functions.

CRUD operations MUST follow standardized patterns:

**Create Pattern:**
- Build request objects from Terraform schema data
- Use `resource.Retry` for creation with proper error handling
- Set resource ID from creation response
- Wait for resource readiness using Service layer WaitFor functions
- Call Read function to synchronize final state

**Read Pattern:**
- Call Service layer Describe function
- Handle `IsNotFoundError` for non-new resources by clearing ID
- Set all schema fields including computed properties
- Return proper error wrapping for unexpected errors

**Delete Pattern:**
- Call Service layer Delete function
- Handle `IsNotFoundError` as successful deletion
- Use StateChangeConf to wait for actual deletion completion
- Proper timeout and delay configuration

Data validation and conversion MUST be properly implemented:
- Use `validation.StringMatch` for string validation with regex
- Implement proper type conversion functions (e.g., `convertToStringSlice`)
- Handle nil values appropriately in conversion functions
- Validate nested object structures with proper Elem definitions

## Governance
This Constitution supersedes all other development practices and guidelines. All pull
requests and code reviews MUST verify compliance with these principles. Any complexity
or deviation from these standards MUST be explicitly justified. Use the development
guide at .github/copilot-instructions.md for runtime development guidance.

**Amendment Procedure**:
- Proposals MUST document rationale, scope, and migration impact.
- Changes MUST be reviewed and approved by the project maintainers.
- Amendments MUST update dependent templates and guidance artifacts.

**Versioning Policy**:
- MAJOR: backward-incompatible governance changes or principle removals.
- MINOR: new principles or material expansions of requirements.
- PATCH: clarifications, wording fixes, or non-semantic refinements.

**Compliance Review**:
- Every spec/plan/tasks artifact MUST include a Constitution check.
- Reviews MUST document any deviations and their approved justification.

**Version**: 1.4.0 | **Ratified**: 2026-01-23 | **Last Amended**: 2026-01-23
