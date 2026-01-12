# Feature Specification: Refactor SLS Resource API Calls

**Feature Branch**: `008-refactor-sls-resource-api`  
**Created**: 2026-01-12  
**Status**: Draft  
**Input**: User description: "将sls相关resource中基于client.Do(\"Sls\"的调用，改成基于SlsService和cws-lib-go中实现的API调用。如更新logstore的调用client.Do(\"Sls\", roaParam(\"PUT\", \"2020-12-30\", \"UpdateLogStore\", action), query, body, nil, hostMap, false)应该更新成SlsService中的UpdateLogStore调用，参数传递都是用API层定义的强类型。"

**Feature (Parent)**: 005 - 强类型 API 调用

## User Scenarios & Testing *(mandatory)*



### User Story 1 - Maintain SLS Resource Behavior While Refactoring (Priority: P1)

As a Terraform Provider maintainer, I want SLS resources to use the Service/API layered architecture and strong types, so that the codebase is safer, more maintainable, and consistent with project standards.

**Why this priority**: This directly affects correctness and maintainability across SLS resources and reduces long-term maintenance cost.

**Independent Test**: Can be tested by running provider compilation and ensuring the updated resources compile and preserve the same schema/state semantics.

**Acceptance Scenarios**:

1. **Given** an SLS resource implementation that currently constructs low-level remote requests in the resource layer, **When** the resource is refactored, **Then** it delegates CRUD operations to the provider's service and API abstraction layers.
2. **Given** the updated provider code, **When** the provider is built, **Then** the build completes successfully without new compilation errors.

---

### User Story 2 - Use Strong Types End-to-End (Priority: P2)

As a maintainer, I want request/response handling to prefer cws-lib-go strong types, so that new SLS changes do not reintroduce weakly-typed maps.

**Why this priority**: Prevents regressions back to `map[string]interface{}` payloads and improves static safety.

**Independent Test**: Can be tested by code review checks (no new `map[string]interface{}` request bodies for SLS API calls) and compilation.

**Acceptance Scenarios**:

1. **Given** an SLS resource update path (e.g., logstore update), **When** the request payload is built, **Then** it uses typed request structures rather than generic key/value maps.

---

### User Story 3 - Improve Consistency for Future SLS Enhancements (Priority: P3)

As a maintainer, I want SLS resources to share a consistent service-driven pattern, so that adding new fields and behaviors is predictable and localized.

**Why this priority**: Reduces onboarding and avoids duplicated Roa request logic.

**Independent Test**: Can be tested by verifying new calls route through `service_alicloud_sls_*.go` and API methods.

**Acceptance Scenarios**:

1. **Given** a new SLS resource capability, **When** it is added, **Then** it can be implemented by extending or reusing the provider's service/API abstraction layers without introducing duplicated low-level request construction in resources.

---

### Edge Cases

- When a resource is newly created and the backend is eventually consistent, reads should tolerate transient NotFound errors using existing retry patterns.
- When the API returns a NotFound for an existing state object during Read, the resource should be removed from state (set ID to empty) following provider conventions.
- When a field is not supported by a specific SLS store type (e.g., metric store vs log store), behavior should remain consistent with current implementation (no accidental breaking changes).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: SLS-related resources MUST delegate logstore CRUD operations to the provider's service/API abstraction layers where an equivalent capability exists.
- **FR-002**: SLS-related resources MUST delegate shard-related operations to the provider's service/API abstraction layers where supported.
- **FR-003**: Request payloads sent through the service/API abstraction layers MUST use typed request structures and MUST NOT introduce new generic key/value request payloads for those operations.
- **FR-004**: Resource schema behavior (ForceNew/Optional/Computed/Deprecated fields and DiffSuppress behavior) MUST remain unchanged.
- **FR-005**: Resource Read behavior MUST continue to correctly populate all computed fields and handle NotFound by removing from state per provider conventions.
- **FR-006**: Any new or adjusted abstraction methods introduced for this refactor MUST keep resource/data-source layers free of direct low-level remote request construction.

### Key Entities *(include if feature involves data)*

- **Logstore**: An SLS storage entity identified by (projectName, logstoreName) with configuration such as TTL, shard count, telemetry type, encryption configuration, and metering mode.
- **Shard**: A partition entity within a logstore identified by shard Id, with begin/end key ranges and status.

## Assumptions

- This work targets SLS resources that currently implement logstore CRUD and shard operations using direct, low-level request construction in the resource layer.
- This refactor is behavior-preserving: no intentional changes to schema, drift detection, or state fields.
- Equivalent capabilities already exist (or can be added) in the provider's service/API abstraction layers.

## Dependencies

- Provider build must remain healthy so compilation can be used as a verification gate.
- Where an equivalent capability does not exist yet in the abstraction layers, it will be added there rather than reintroducing low-level logic into resource code.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of targeted SLS resource code paths that previously constructed low-level remote requests in resources are refactored to use the provider's abstraction layers.
- **SC-002**: Provider builds successfully with zero new compilation errors.
- **SC-003**: No new weakly-typed request payloads are introduced for SLS operations covered by this spec.
- **SC-004**: Resource schema definitions for the targeted resources remain unchanged (no diff in schema keys and flags).

## Clarifications

### Session 2026-01-12

- Q: Does the refactoring scope include ALL `sls` resources (e.g. Project, MachineGroup, Config, etc.) that use `client.Do`, or is it strictly limited to `alicloud_log_store` and its shard operations? → A: All SLS resources found using `client.Do("Sls"...)` (Comprehensive).
- Q: I found extensive usage of the `client.WithLogClient` wrapper (which allows direct usage of the legacy `aliyun-log-go-sdk`) to be the pattern used in older resources (e.g., `alicloud_log_alert`, `alicloud_sls_machine_group`), rather than `client.Do("Sls"...)`. Should I assume the goal is to refactor these `client.WithLogClient` usages to `SlsService`? → A: Yes - Refactor `client.WithLogClient` usages to `SlsService`.

```
