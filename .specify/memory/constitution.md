# Terraform Provider Alicloud Constitution
<!--
Synced: 2026-02-06
-->

<!--
  Sync Impact Report: (2026-02-06)
  - Version: 1.0.0 -> 1.0.1
  - Updates:
    - Refreshed Constitution Check in plan-template.md
    - Corrected Principle references in tasks-template.md
  - Templates requiring updates:
    - /.specify/templates/plan-template.md (✅ updated)
    - /.specify/templates/tasks-template.md (✅ updated)
-->

## Core Principles

### I. Layered Architecture & Library-First Design
Every resource or data source MUST be built upon a robust Service Layer:
- **Resource/DataSource** -> **Service** -> **API (CWS-Lib-Go)** -> **SDK**.
- Service Layer MUST be cohesive, reusable, and independently testable.
- Avoid direct SDK calls in Resource/DataSource layers.

Rationale: separates concerns, encourages reuse, and simplifies maintenance.

### II. Standardized Interfaces
The provider MUST expose consistent interfaces:
- Use standard Terraform Resource/DataSource schemas.
- Implement strictly typed Service methods (no `map[string]interface{}`).
- Pagination logic MUST be encapsulated in the API/Service layer, not exposed to callers.

Rationale: ensures type safety, reliability, and uniform behavior.

### III. Test-First Development
Implementation MUST follow a Test-Driven Development style for core logic:
- Write or update tests BEFORE implementing new behavior.
- Ensure integration/acceptance tests (Terraform Acceptance Tests) cover critical flows.

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
- Documentation MUST be clear, with API docs in English and internal docs/guides in Chinese.

Rationale: makes systems debuggable, upgradable, and maintainable.

### VI. Continuous Integration & Quality Gates
Changes MUST be safe to merge:
- `make` MUST pass (compilation, linting).
- New behavior MUST be reflected in specs/plan/tasks/docs.

Rationale: ensures consistent quality and predictable releases.

### VII. Feature-Centric Development
Feature 是项目的长期核心框架：
- Feature 列表必须保持为项目的“单一事实来源”。
- 在 spec → plan → tasks → implement 的每个阶段都必须复核 Feature 的新增/合并/拆分/删除。
- Feature 变更必须可追溯到相应的 spec/plan 依据，并记录在 Feature 详情中。

Rationale: 让项目演进以 Feature 为中心，确保长期一致性与可维护性。

## Engineering Standards

### Architecture & API
- **Layering**: Strictly follow `Resource -> Service -> API -> SDK`.
- **Strong Typing**: MUST use `cws-lib-go` strong types. `map[string]interface{}` is FORBIDDEN in new code.
- **Pagination**: MUST be handled in `*_api.go` or Service methods; return full slices to callers.

### State Management
- **Refresh & Wait**: Service layer MUST implement `StateRefreshFunc` and `WaitFor*` methods.
- **No Polling in Resource**: Resource `Create` methods MUST NOT contain polling loops; use Service `WaitFor`.

## Development Workflow

### Verification & Operations
- **Syntax Check**: Execute `make` to verify code before committing.
- **Safety**: Generate scripts for complex file ops; backup before batch changes.
- **Binaries**: Output binaries to `bin/`, never root.

### Code Organization
- **Splitting**: Split files exceeding 1000 lines.
- **Language**: Generate documents in Chinese. Code comments, logs, API docs, and error messages in English.

## Governance

### Authority
This Constitution and the `docs/development_guide.md` are authoritative.
- **Precedence**: This Constitution defines high-level principles. `docs/development_guide.md` defines specific engineering constraints.
- **Compliance**: All PRs MUST check compliance with these principles.

### Amendments
- Changes to these principles require a PR, review, and a version bump of this Constitution.

**Version**: 1.0.1 | **Ratified**: 2026-02-04 | **Last Amended**: 2026-02-06
