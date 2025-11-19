# Implementation Plan: Kafka Provider Refactoring

**Branch**: `002-refactor-kafka-provider` | **Date**: 2025-11-17 | **Spec**: [link](spec.md)
**Input**: Feature specification from `/.specify/specs/002-refactor-kafka-provider/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.github/prompts/speckit.plan.prompt.md` for the execution workflow.

## Summary

Refactor Kafka-related provider implementation to use modern cws-lib-go API layer instead of direct SDK calls, following the layered architecture pattern established by the Flink provider implementation. This includes 11 files covering instances, topics, consumer groups, SASL users, SASL ACLs, and IP attachments.

## Technical Context

**Language**: Go (version managed by toolchain/CI; go.mod currently has no explicit `go` directive)  
**Primary Dependencies**: cws-lib-go (kafka API), terraform-plugin-sdk v1.17.2, alibaba-cloud-sdk-go (replaced)  
**Storage**: N/A (Terraform provider, cloud resource management only)  
**Testing**: Go test framework with Terraform acceptance tests  
**Target Platform**: Cross-platform (Linux/Windows/macOS - Terraform provider standard)  
**Project Type**: Single project (Terraform provider)  
**Performance Goals**: Maintain existing performance baseline (within 10% of current)  
**Constraints**: Must maintain 100% backward compatibility, follow Terraform provider architecture guidelines  
**Scale/Scope**: Refactoring 11 existing files with complete functional equivalence

## Operational Policies

**Retry & Backoff (Constitution-aligned)**
- Use exponential backoff with jitter for retryable errors (e.g., ServiceUnavailable, Throttling, InternalError):
	- Initial backoff: 1s
	- Multiplier: x2
	- Jitter: full jitter (randomized delay up to current backoff)
	- Max backoff per attempt: 30s
	- Overall operation deadline bounded by Terraform timeouts (Create/Update/Delete)

**Timeout Baselines (Testable)**
- Default timeouts (can be overridden per resource if justified):
	- Create: 10 minutes
	- Update: 10 minutes
	- Delete: 5 minutes
	- Read: rely on provider defaults; avoid long-blocking reads

These baselines must be reflected in resource Timeouts and respected by Service-layer WaitFor* functions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Confirm the following gates derived from the project constitution:

- ✅ Layering: Resources/DataSources will call Service layer only; Service will use CWS-Lib-Go API layer. No direct SDK/HTTP.
- ✅ State management: Create/Delete will use Service-layer WaitFor funcs; no Read in Create; proper timeouts configured.
- ✅ Error handling: Will use wrapped errors and helper predicates (IsNotFoundError/IsAlreadyExistError/NeedRetry). Avoid raw IsExpectedErrors.
- ✅ Strong typing: Will use CWS-Lib-Go strong types exclusively; avoid `map[string]interface{}` completely.
- ✅ Pagination: Will be encapsulated in service layer; callers will not handle pagination.
- ✅ ID encoding: Encode/Decode helpers will be implemented and used consistently.
- ✅ Build verification: `make` will pass locally before merge; code will be split if file >1000 LOC.

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

**Structure Decision**: The existing file structure will be maintained with refactored implementations following the layered architecture pattern. Each resource/data source will call the service layer, which will use cws-lib-go API functions.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No violations expected. The refactoring strictly follows all constitutional requirements.
