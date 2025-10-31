# Tasks: ConcurrentUpdateBucketFailed error support

Feature: 004-concurrentupdatebucketfailed-error-support  
Spec: /.specify/specs/004-concurrentupdatebucketfailed-error-support/spec.md  
Plan: /.specify/specs/004-concurrentupdatebucketfailed-error-support/plan.md

Notes:
- Tests are optional; spec does not explicitly require TDD. We provide independent test criteria per story but do not add test-implementation tasks by default.
- Tasks are grouped by user story with phases and explicit ordering. [P] marks tasks that can run in parallel.

---

## Phase 1: Setup (shared infrastructure)

- [X] T001: Verify feature directory and design docs present [P]
  - Path: /cws_data/terraform-provider-alicloud/.specify/specs/004-concurrentupdatebucketfailed-error-support
  - Check files: plan.md, spec.md; optional: research.md, data-model.md, contracts/, quickstart.md
  - Acceptance: All required docs readable; optional docs status recorded

- [X] T002: Baseline build to ensure clean starting point
  - Run: make (repo root)
  - Acceptance: Build succeeds with no errors

- [X] T003: Locate target implementation file and insertion points [P]
  - Path: alicloud/resource_alicloud_oss_public_access_block.go
  - Identify: client.Do("Oss", ...) calls in Create and Update; confirm timeouts used by BuildStateConf afterward
  - Acceptance: Exact line ranges noted for patching

---

## Phase 2: Foundational (blocking prerequisites)

- [X] T004: Define retry gating for OSS concurrency conflict [P]
  - Pattern: treat Code=ConcurrentUpdateBucketFailed (HTTP 409) as retryable
  - Implementation note: use IsExpectedErrors(err, []string{"ConcurrentUpdateBucketFailed"}) OR NeedRetry(err)
  - Acceptance: Helper predicate or inline check drafted for reuse in Create/Update

- [X] T005: Establish exponential backoff with jitter helper [P]
  - Strategy: initialDelay=2s, factor≈1.6–2.0, jitter±25%, maxDelay=30s
  - Scope: local to resource function(s); respect Terraform d.Timeout boundaries
  - Acceptance: Helper function or inline logic ready for integration

---

## Phase 3: User Story 1 (P1) – 自动重试并成功应用

Story Goal: 当发生 409 ConcurrentUpdateBucketFailed 时，自动重试并在超时窗口内尽量完成。

Independent Test Criteria:
- 人为制造并发对同一桶公共访问阻断设置，观察 409 发生后自动重试并最终成功；
- 无并发时，路径不引入额外等待。

Implementation Tasks:

- [X] T006 [US1]: Add retry around Create path (PUT) in resource_alicloud_oss_public_access_block.go (sequential with T007)
  - File: alicloud/resource_alicloud_oss_public_access_block.go
  - Wrap client.Do(...) in resource.Retry(d.Timeout(schema.TimeoutCreate), ...)
  - Retry when IsExpectedErrors(err, {"ConcurrentUpdateBucketFailed"}) || NeedRetry(err);
    apply backoff+jitter; log attempt count and last error summary
  - Acceptance: Build passes; manual read-through shows correct gating and timeout use

- [X] T007 [US1]: Add retry around Update path (PUT) in resource_alicloud_oss_public_access_block.go (sequential with T006)
  - File: alicloud/resource_alicloud_oss_public_access_block.go
  - Wrap client.Do(...) in resource.Retry(d.Timeout(schema.TimeoutUpdate), ...)
  - Same retry gating and backoff parameters as Create
  - Acceptance: Build passes; update path mirrors create semantics

- [X] T008 [US1]: Build and smoke check [P]
  - Run: make; optional go test ./...
  - Acceptance: No compile errors; provider binary builds

- [X] T009 [US1]: Update/verify quickstart usage guidance [P]
  - File: /.specify/specs/004-concurrentupdatebucketfailed-error-support/quickstart.md
  - Ensure it reflects observed behavior and log hints
  - Acceptance: Doc describes how to trigger/observe retries

---

## Phase 4: User Story 2 (P2) – 重试后仍失败的清晰反馈

Story Goal: 当持续冲突超时，尽快失败并给出可操作的提示。

Independent Test Criteria:
- 持续制造冲突直至超时；错误信息包含“并发更新检测、请冷却后重试”。

Implementation Tasks:

- [X] T010 [US2]: Enrich terminal error message on retry timeout (sequential with T006/T007)
  - File: alicloud/resource_alicloud_oss_public_access_block.go
  - On resource.Retry returning error, WrapErrorf with DefaultErrorMsg context + hint about concurrency conflict
  - Acceptance: Error output includes clear guidance without leaking sensitive details

- [X] T011 [US2]: Ensure retry metrics in logs [P]
  - File: alicloud/resource_alicloud_oss_public_access_block.go
  - Log total attempts and last error summary at WARN/DEBUG level
  - Acceptance: Logs show attempt count and last error summary

---

## Phase 5: User Story 3 (P3) – 向后兼容且无额外配置

Story Goal: 无需新增配置；无并发场景时延不显著增加。

Independent Test Criteria:
- 无并发干扰 apply，执行时延与变更前等效（或 P90 增量 ≤ 5%）。

Implementation Tasks:

- [X] T012 [US3]: Confirm no added latency on non-conflict path [P]
  - File: alicloud/resource_alicloud_oss_public_access_block.go
  - Verify backoff only occurs when retry branch taken; no sleep on success path
  - Acceptance: Code review checklist confirms zero added waits on happy path

- [X] T013 [US3]: Build baseline verification [P]
  - Run: make
  - Acceptance: Build stable; no additional warnings/errors

---

## Final Phase: Polish & Cross-Cutting

- [X] T014: Optional unit test scaffolding (only if tests requested) [P]
  - Scope: add minimal unit test for backoff helper and error-gating
  - Acceptance: Backoff unit test added under build tag `expbackoff_unit` to avoid impacting default builds; full `go test` across repo currently blocked by unrelated package compile issues.

- [X] T015: Consider extracting retry/backoff helper to shared util [P]
  - Scope: reduce duplication across similar OSS resources
  - Acceptance: Shared helper `ExpBackoffWait` and `IsOssConcurrentUpdateError` added in `alicloud/retry.go`; resource updated to use Service wrapper.

- [X] T016: Long-term: Service-layer write API for PublicAccessBlock [P]
  - Scope: move PUT calls from resource to Service层，统一分层与错误处理
  - Acceptance: Implemented `PutOssBucketPublicAccessBlock` in `service_alicloud_oss_public_access_block.go`; resource now calls Service for writes with centralized retry/backoff.

---

## Dependencies & Ordering

- Phase order: Setup → Foundational → US1 (P1) → US2 (P2) → US3 (P3) → Polish
- Task dependencies (selected):
  - T006 → requires T004, T005, T003
  - T007 → requires T006
  - T010 → requires T006, T007
  - Other [P] tasks can run in parallel with unrelated-file edits or builds

User Story completion order:
- US1 → US2 → US3

---

## Parallel Execution Examples

- In US1: T008 (build) and T009 (doc verification) can run in parallel with code review, but code edits T006 and T007 are sequential (same file).
- Across phases: T001 and T003 can run in parallel; T002 (build) can run while drafting T004/T005.

---

## Implementation Strategy

- MVP: Complete US1 (T006–T009). This delivers the core auto-retry capability under contention.
- Then: US2 for improved failure messaging, followed by US3 compatibility checks.
- Defer refactors/tests unless explicitly requested or required by review.

---

## Counts & Coverage

- Total tasks: 16
- By story:
  - US1: 4 tasks (T006–T009)
  - US2: 2 tasks (T010–T011)
  - US3: 2 tasks (T012–T013)
- Setup: 3 tasks (T001–T003)
- Foundational: 2 tasks (T004–T005)
- Polish: 3 tasks (T014–T016)

Parallel opportunities identified:
- [P] T001, T003, T004, T005, T008, T009, T011, T012, T013, T014, T015, T016

Independent test criteria (per story):
- US1: 并发冲突后自动重试直至成功；无并发时无额外等待
- US2: 超时失败时提供并发冲突提示
- US3: 无并发场景的时延等效

Suggested MVP scope:
- 完成 US1（T006–T009）