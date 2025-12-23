<!--
## Sync Impact Report

**Version change**: 1.2.0 → 1.3.0
**Modified principles**: 
- Updated: All principles refined and reorganized for clarity

**Added sections**: 
- Development Workflow Standards
- Quality Assurance Requirements

**Removed sections**: None

**Templates requiring updates**:
✅ .specify/templates/plan-template.md – Constitution Check updated to include strong typing and layering/state/error gates
✅ .specify/templates/spec-template.md – Added non-functional requirements placeholders aligned to constitution (incl. strong typing)
✅ .specify/templates/tasks-template.md – Added cross-cutting task to enforce strong typing and build verification
⚠ .specify/templates/commands/*.md – Not present in repository; N/A (no action)

**Follow-up TODOs**: 
None
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
	wherever applicable.
- MUST NOT introduce new usages of `map[string]interface{}` for request/response shapes,
	except when interacting with legacy code paths.
- Legacy code is exempt (read-only, minimal-touch). Any refactoring SHOULD migrate to
	strong types opportunistically while maintaining backward compatibility.
- Code reviews MUST flag weak typing in new/modified code unless explicitly justified
	(e.g., bridging adapters to third-party libs not yet modeled in cws-lib-go).

### VI. Testing and Validation Requirements
Every code change MUST be validated by executing 'cd /cws_data/terraform-provider-alicloud && make' to ensure syntax correctness and successful compilation. Comprehensive unit tests and integration tests are mandatory. All resources MUST include proper Timeout configurations. Code files exceeding 1000 lines MUST be split by functional modules to ensure single responsibility.

Binary generation MUST NOT occur in the root directory. All binary files MUST be output to the `bin` directory and ignored by `.gitignore`.

API pagination logic MUST be encapsulated in `*_api.go` files:
- External callers should not handle pagination details
- Use page number/page size iteration until all results are collected
- Return complete result sets to callers

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
This Constitution supersedes all other development practices and guidelines. All pull requests and code reviews MUST verify compliance with these principles. Any complexity or deviation from these standards MUST be explicitly justified. Use the development guide at .github/copilot-instructions.md for runtime development guidance. Amendments require documentation, team approval, and migration plans for existing code.

**Version**: 1.3.0 | **Ratified**: 2017-01-19 | **Last Amended**: 2025-12-02