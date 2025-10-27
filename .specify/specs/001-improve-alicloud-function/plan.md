# Implementation Plan: Improve AliCloud Function Compute Support

**Branch**: `001-improve-alicloud-function` | **Date**: October 24, 2025 | **Spec**: [/cws_data/terraform-provider-alicloud/.specify/specs/001-improve-alicloud-function/spec.md](file:///cws_data/terraform-provider-alicloud/.specify/specs/001-improve-alicloud-function/spec.md)
**Input**: Feature specification from `/.specify/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature aims to improve the AliCloud Function Compute support in the Terraform provider by completing and standardizing the implementation of all FC-related resources. The technical approach involves following the established patterns in service_alicloud_fc_function.go and service_alicloud_fc_layer.go to ensure consistency across all FC service implementations.

## Technical Context

**Language/Version**: Go 1.18+  
**Primary Dependencies**: github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3, github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity  
**Storage**: N/A (stateless API interactions with AliCloud FC service)  
**Testing**: Go testing package with Terraform acceptance tests  
**Target Platform**: Cross-platform (Linux, macOS, Windows)  
**Project Type**: Single project (Terraform provider)  
**Performance Goals**: Standard Terraform provider performance expectations  
**Constraints**: Must follow Terraform Provider AliCloud Constitution and coding standards  
**Scale/Scope**: All FC-related resources in the Terraform provider

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on the Terraform Provider Alicloud Constitution (Version 1.1.0), this implementation plan must ensure compliance with the following principles:

1. **Architecture Layering Principle**: 
   - ✅ All FC resources will follow the required architecture: Provider Layer → Resource/DataSource Layer → Service Layer → API Layer (CWS-Lib-Go)
   - ✅ Service layer API calls will use CWS-Lib-Go encapsulation
   - ✅ Direct HTTP requests or third-party SDKs will be avoided

2. **State Management Best Practices**:
   - ✅ State management will follow proper patterns with StateRefreshFunc mechanisms
   - ✅ Read functions will not be called directly in Create functions
   - ✅ d.SetId("") will be used when resources don't exist
   - ✅ All computed properties will be set in Read methods
   - ✅ Idempotent operations will be implemented
   - ✅ Service layer WaitFor functions will be used for resource readiness

3. **Error Handling Standardization**:
   - ✅ Error handling will use encapsulated error judgment functions from alicloud/errors.go
   - ✅ Priority order: IsNotFoundError, IsAlreadyExistError, NeedRetry
   - ✅ Predefined error code lists will be used for service-specific errors
   - ✅ Errors will be wrapped using WrapError/WrapErrorf with detailed context
   - ✅ Retry logic will handle common retryable errors

4. **Code Quality and Consistency**:
   - ✅ Naming conventions will be followed strictly
   - ✅ ID fields will use "Id" not "ID"
   - ✅ All schema fields will include appropriate Description
   - ✅ Service layer will implement proper ID encoding/decoding functions

5. **Testing and Validation Requirements**:
   - ✅ Every code change will be validated with 'make' command
   - ✅ Comprehensive unit tests and integration tests will be included
   - ✅ All resources will include proper Timeout configurations
   - ✅ API pagination logic will be encapsulated in service files
   - ✅ Code files will be split if exceeding 1000 lines

6. **CRUD Operations Standardization**:
   - ✅ Create pattern will follow: build request, retry with error handling, set ID, wait for readiness, call Read
   - ✅ Read pattern will follow: call Describe, handle IsNotFoundError, set schema fields, proper error wrapping
   - ✅ Delete pattern will follow: call Delete, handle IsNotFoundError, wait for deletion, proper timeout config

7. **Data Validation and Conversion**:
   - ✅ Proper validation functions will be used
   - ✅ Type conversion functions will handle nil values appropriately
   - ✅ Nested object structures will have proper Elem definitions

## Project Structure

### Documentation (this feature)

```
.specify/specs/001-improve-alicloud-function/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
alicloud/
├── service_alicloud_fc_base.go
├── service_alicloud_fc_function.go
├── service_alicloud_fc_layer.go
├── service_alicloud_fc_trigger.go
├── service_alicloud_fc_custom_domain.go
├── service_alicloud_fc_alias.go
├── resource_alicloud_fc_function.go
├── resource_alicloud_fc_function_version.go
├── resource_alicloud_fc_layer_version.go
├── resource_alicloud_fc_trigger.go
├── resource_alicloud_fc_custom_domain.go
├── resource_alicloud_fc_alias.go
├── resource_alicloud_fc_async_invoke_config.go
├── resource_alicloud_fc_concurrency_config.go
├── resource_alicloud_fc_provision_config.go
├── resource_alicloud_fc_vpc_binding.go
├── data_source_alicloud_fc_functions.go
├── data_source_alicloud_fc_triggers.go
├── data_source_alicloud_fc_custom_domains.go
└── data_source_alicloud_fc_zones.go
```

**Structure Decision**: This feature focuses on improving the existing FC-related service implementations in the alicloud/ directory. The structure follows the existing Terraform provider pattern with service files containing the business logic and resource files containing the Terraform integration.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
