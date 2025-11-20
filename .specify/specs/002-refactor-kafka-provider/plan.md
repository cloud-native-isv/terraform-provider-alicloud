# Implementation Plan: Kafka Provider Refactoring

**Branch**: `002-refactor-kafka-provider` | **Date**: 2025-11-19 | **Spec**: /cws_data/terraform-provider-alicloud/.specify/specs/002-refactor-kafka-provider/spec.md
**Input**: Feature specification from `/.specify/specs/002-refactor-kafka-provider/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.github/prompts/speckit.plan.prompt.md` for the execution workflow.

## Summary

The primary requirement is to refactor the Kafka provider implementation to use the modern cws-lib-go API layer instead of direct SDK calls. This involves updating 11 files (5 data sources, 5 resources, and 1 service file) to follow the Provider → Resource/DataSource → Service → API layered architecture. The technical approach is to follow the established Flink implementation pattern, using strong typing, proper error handling, and state management best practices.

## Technical Context

**Language/Version**: Go 1.18+  
**Primary Dependencies**: 
- github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity
- github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka
- github.com/hashicorp/terraform-plugin-sdk
**Storage**: N/A  
**Testing**: github.com/hashicorp/terraform-plugin-testing  
**Target Platform**: Linux, macOS, Windows
**Project Type**: Single project (Terraform provider)  
**Performance Goals**: Standard Terraform provider performance  
**Constraints**: 
- Must use cws-lib-go API layer
- Must maintain backward compatibility
- Must pass existing acceptance tests
**Scale/Scope**: 
- 11 files to refactor
- Existing functionality must be preserved

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Confirm the following gates derived from the project constitution:

- Layering: Resources/DataSources call Service layer only; Service uses CWS-Lib-Go API layer. No direct SDK/HTTP. ✅
- State management: Create/Delete use Service-layer WaitFor funcs; no Read in Create; proper timeouts configured. ✅
- Error handling: Use wrapped errors and helper predicates (IsNotFoundError/IsAlreadyExistError/NeedRetry). Avoid raw IsExpectedErrors. ✅
- Strong typing: Prefer CWS-Lib-Go strong types; avoid `map[string]interface{}` except legacy code with justification. ✅
- Pagination: Encapsulated in `*_api.go`; callers do not handle pagination. ✅
- ID encoding: Encode/Decode helpers implemented and used consistently. ✅
- Build verification: `make` passes locally before merge; code split if file >1000 LOC. ✅

All constitutional requirements are met by the planned implementation approach.

## Project Structure

### Documentation (this feature)

```
.specify/specs/002-refactor-kafka-provider/
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
├── data_source_alicloud_alikafka_consumer_groups.go
├── data_source_alicloud_alikafka_instances.go
├── data_source_alicloud_alikafka_sasl_acls.go
├── data_source_alicloud_alikafka_sasl_users.go
├── data_source_alicloud_alikafka_topics.go
├── resource_alicloud_alikafka_consumer_group.go
├── resource_alicloud_alikafka_instance_allowed_ip_attachment.go
├── resource_alicloud_alikafka_instance.go
├── resource_alicloud_alikafka_sasl_acl.go
├── resource_alicloud_alikafka_sasl_user.go
├── resource_alicloud_alikafka_topic.go
└── service_alicloud_alikafka.go
```

**Structure Decision**: The refactoring focuses on the existing Kafka-related files in the alicloud/ directory. No new directories or major structural changes are needed.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | All constitutional requirements can be met without violations |

No constitutional violations are anticipated in this refactoring.
