# Terraform Provider Alicloud Constitution

<!--
  Sync Impact Report: (2026-07-06)
  - Version: 1.0.2 -> 1.1.0.1
  - Bump type: MINOR (adding 3 new principles + materially expanding Principle I)
  - Modified principles:
    - Principle I: "Layered Architecture & Library-First Design" — strengthened
      with explicit MUST NOT prohibition and layer violation detection mandate
  - Added principles:
    - VIII. Toolchain & Version Baseline
    - IX. Error Handling & Resilience
    - X. Security & Credential Governance
  - Updated sections:
    - Engineering Standards: added Go version baseline (1.20 -> 1.26),
      added Error Handling and Security subsections
    - Governance: added development_guide.md alignment declaration and
      version bump policy documentation
  - Templates requiring updates:
    - /.specify/templates/plan-template.md         (no change — dynamic rendering)
    - /.specify/templates/requirements-template.md (no change — no stale refs)
    - /.specify/templates/tasks-template.md        (no change — dynamic reference)
  - Follow-up TODOs: None
-->

## Core Principles

### I. Layered Architecture & Library-First Design

Every resource or data source MUST be built upon a robust Service Layer:

- **Resource/DataSource** -> **Service** -> **API (CWS-Lib-Go)** -> **SDK**.
- Service Layer MUST be cohesive, reusable, and independently testable.
- Resource/DataSource layers MUST NOT directly call SDK or RPC functions.
  Direct SDK calls in Resource/DataSource code are FORBIDDEN in new code.
- Layer violations MUST be detectable via code review and SHOULD be
  enforceable through automated checks (e.g., lint rules, import analysis).

Rationale: separates concerns, encourages reuse, and simplifies maintenance.
Aligns with `docs/development_guide.md` §1.1-1.2.

### II. Standardized Interfaces

The provider MUST expose consistent interfaces:

- Use standard Terraform Resource/DataSource schemas.
- Implement strictly typed Service methods (no `map[string]interface{}`).
- Pagination logic MUST be encapsulated in the API/Service layer, not
  exposed to callers.

Rationale: ensures type safety, reliability, and uniform behavior.

### III. Test-First Development

Implementation MUST follow a Test-Driven Development style for core logic:

- Write or update tests BEFORE implementing new behavior.
- Ensure integration/acceptance tests (Terraform Acceptance Tests) cover
  critical flows.

Rationale: reduces regressions and clarifies intent.

### IV. Integration & Contract Testing

Service layer interactions SHOULD be verified against the real API or mocks:

- Usage of CWS-Lib-Go implies API contract adherence.
- State inconsistencies MUST be handled gracefully (e.g., `WaitFor` logic).

Rationale: validates real-world behavior and cloud consistency.

### V. Observability, Versioning & Simplicity

All components MUST be observable and maintainable:

- Use structured logs (standard Terraform logging) for important events.
- Adhere to Semantic Versioning for the provider releases.
- Documentation MUST be clear, with API docs in English and internal
  docs/guides in Chinese.

Rationale: makes systems debuggable, upgradable, and maintainable.

### VI. Continuous Integration & Quality Gates

Changes MUST be safe to merge:

- `make` MUST pass (compilation, linting).
- New behavior MUST be reflected in specs/plan/tasks/docs.

Rationale: ensures consistent quality and predictable releases.

### VII. Feature-Centric Development

Feature 是项目的长期核心框架：

- Feature 列表必须保持为项目的"单一事实来源"。
- 在 spec -> plan -> tasks -> implement 的每个阶段都必须复核 Feature
  的新增/合并/拆分/删除。
- Feature 变更必须可追溯到相应的 spec/plan 依据，并记录在 Feature 详情中。

Rationale: 让项目演进以 Feature 为中心，确保长期一致性与可维护性。

### VIII. Toolchain & Version Baseline

The project MUST enforce a consistent toolchain across all development
environments:

- **Go Version**: Minimum Go 1.26 (as declared in `go.mod`). The previous
  baseline of Go 1.20 is superseded.
- **Code Formatting**: `gofmt` and `goimports` MUST be applied to all
  committed code. CI MUST reject non-formatted submissions.
- **Dependency Management**: All external dependencies MUST be declared
  in `go.mod`. Local SDK replacements (`replace` directives) MUST be
  used for Alibaba Cloud SDKs vendored under `pkg/cws-lib-go/`.
- **Build Reproducibility**: `make build` MUST produce a deterministic
  binary given the same `go.mod` / `go.sum` state.

Rationale: eliminates environment drift, ensures reproducible builds,
and keeps the codebase maintainable as the Go ecosystem evolves.

### IX. Error Handling & Resilience

All error paths MUST follow the project's established error handling patterns:

- **Error Wrapping**: Every error returned to Terraform MUST be wrapped
  using `WrapError(err)` or `WrapErrorf(err, msg, args...)` with
  meaningful context (resource ID, action name).
- **Error Classification**: Use the canonical helpers from
  `alicloud/errors.go`: `IsNotFoundError(err)`, `IsAlreadyExistError(err)`,
  `NeedRetry(err)`. Inline `IsExpectedErrors` with ad-hoc code lists is
  discouraged in new code; prefer pre-defined error code constants.
- **Retry Logic**: Retryable errors (`ThrottlingException`,
  `ServiceUnavailable`, `InternalError`, `SystemBusy`,
  `OperationConflict`) MUST be handled via `resource.Retry` at the
  Resource layer or via Service-layer `WaitFor*` methods. Resource
  `Create` methods MUST NOT contain raw polling loops.
- **No Silent Swallowing**: Errors MUST NOT be silently ignored. If an
  error is intentionally discarded, a `log.Printf` with `[DEBUG]` or
  `[WARN]` level MUST record the reason.

Rationale: ensures consistent, debuggable error behaviour and aligns
with `docs/development_guide.md` §2.3.

### X. Security & Credential Governance

Security-sensitive operations MUST adhere to the following safeguards:

- **Credential Sources**: Alibaba Cloud credentials
  (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION`,
  `ALICLOUD_ACCOUNT_ID`) MUST be sourced from environment variables or
  Terraform provider configuration. Hardcoding secrets in source code,
  tests, or documentation is FORBIDDEN.
- **Sensitive Fields**: Schema attributes containing secrets, tokens, or
  passwords MUST set `Sensitive: true` to prevent value leakage in
  Terraform plan/apply output.
- **Log Hygiene**: Logs MUST NOT print raw credentials, tokens, or
  user-supplied secrets. When diagnostic context is needed, redact or
  hash the value before logging.
- **Audit Trail**: State-changing operations (Create, Update, Delete)
  SHOULD emit a structured log entry capturing the resource type, ID,
  and operation outcome for traceability.

Rationale: protects user secrets, enables compliance auditing, and
prevents accidental credential exposure through logs or state files.

## Engineering Standards

### Architecture & API

- **Layering**: Strictly follow `Resource -> Service -> API -> SDK`.
  Resource/DataSource layers MUST NOT bypass the Service layer.
- **Strong Typing**: MUST use `cws-lib-go` strong types.
  `map[string]interface{}` is FORBIDDEN in new code.
- **Pagination**: MUST be handled in `*_api.go` or Service methods;
  return full slices to callers.
- **Go Version**: Minimum Go 1.26 (`go.mod` baseline). Previous
  Go 1.20 baseline is superseded.

### State Management

- **Refresh & Wait**: Service layer MUST implement `StateRefreshFunc`
  and `WaitFor*` methods.
- **No Polling in Resource**: Resource `Create` methods MUST NOT contain
  polling loops; use Service `WaitFor`.

### Error Handling

- **Wrapping**: Use `WrapError` / `WrapErrorf` for all returned errors
  with meaningful context (resource ID, action).
- **Classification**: Prefer `IsNotFoundError`, `IsAlreadyExistError`,
  `NeedRetry` over ad-hoc `IsExpectedErrors` checks.
- **Retry**: Encapsulate retry logic in Service layer or
  `resource.Retry`; do not implement custom polling in Resource CRUD.

### Security

- **Credentials**: Environment variables only; no hardcoded secrets.
- **Sensitive Schema**: Set `Sensitive: true` on secret-bearing fields.
- **Log Redaction**: No raw secrets in logs; redact or hash.

## Development Workflow

### Verification & Operations

- **Syntax Check**: Execute `make` to verify code before committing.
- **Safety**: Generate scripts for complex file ops; backup before batch
  changes.
- **Binaries**: Output binaries to `bin/`, never root.

### Code Organization

- **Splitting**: Split files exceeding 1000 lines.
- **Language**: Generate documents in Chinese. Code comments, logs, API
  docs, and error messages in English.

## Governance

### Authority

This Constitution and the `docs/development_guide.md` are authoritative.

- **Precedence**: This Constitution defines high-level principles.
  `docs/development_guide.md` defines specific engineering constraints
  and code patterns. Where they overlap, the Constitution's principles
  take precedence; the development guide provides implementation detail.
- **Alignment**: The development guide (current version 1.3.0) MUST be
  updated whenever a Constitution amendment affects its scope.
- **Compliance**: All PRs MUST check compliance with these principles.

### Amendments

- Changes to these principles require a PR, review, and a version bump
  of this Constitution.
- Version bumps follow the `x.y.z.ddd` format:
  - **MAJOR (x)**: Rewrite-level changes — complete restructuring of
    principles or backward-incompatible governance removals.
  - **MINOR (y)**: Adding/removing/renaming principles or materially
    expanding or contracting principle scope.
  - **PATCH (z)**: Clarifications, wording improvements, typo fixes,
    non-semantic adjustments to existing principles.
  - **DAILY (ddd)**: Every update regardless of change magnitude.

**Version**: 1.1.0.1 | **Ratified**: 2026-02-04 | **Last Amended**: 2026-07-06
