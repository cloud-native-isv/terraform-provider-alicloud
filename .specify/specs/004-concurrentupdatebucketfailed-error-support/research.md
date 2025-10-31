# Research: ConcurrentUpdateBucketFailed error support

## Decisions

1. Decision: Treat `ConcurrentUpdateBucketFailed` (HTTP 409) as retryable during Create/Update of `alicloud_oss_bucket_public_access_block`.
   - Rationale: The error explicitly indicates concurrent update conflicts that are transient; retrying after cooldown typically succeeds.
   - Alternatives considered:
     - Rely solely on `NeedRetry(err)`: Rejected because generic retry detection may not consistently identify this OSS-specific 409 conflict.
     - Static sleep with fixed interval: Rejected due to potential herd effects and longer convergence time under contention.

2. Decision: Implement exponential backoff with jitter within `resource.Retry` loop.
   - Rationale: Industry-standard approach to mitigate contention and avoid synchronized retries; respects Terraform timeouts.
   - Strategy: Start at 2s delay, multiplier ~1.6–2.0, apply ±25% jitter, cap single-sleep at 30s, and always respect `d.Timeout(schema.TimeoutCreate/Update)`.
   - Alternatives considered: Linear backoff; immediate aggressive retries without sleeps.

3. Decision: Scope change to Bucket-level Public Access Block only (Create/Update), keep Read/Delete unchanged.
   - Rationale: Spec scope is to address failures observed at `resource_alicloud_oss_public_access_block.go:79` in Put path; Read/Delete not impacted.
   - Alternatives considered: Extend to Account/Access Point variants now; deferred to follow-up to keep change minimal and low-risk.

4. Decision: Error handling uses existing helpers and wrapping
   - Rationale: Align with constitution: use `IsExpectedErrors(err, []string{"ConcurrentUpdateBucketFailed"})` OR `NeedRetry(err)`; wrap via `WrapErrorf`.
   - Alternatives considered: Ad-hoc string matching only; rejected in favor of helpers that already parse tea/ServerError/common errors.

## Implementation Notes

- In Create/Update flows around `client.Do("Oss", ...)` invocations, add `resource.Retry` with retryable branch on `ConcurrentUpdateBucketFailed`.
- Add incremental sleep with jitter inside the retry closure. Keep a local `attempt` counter to compute delay.
- Respect timeouts via `d.Timeout(schema.TimeoutCreate/Update)` which bounds `resource.Retry`.
- Log retry attempts and last error summary without leaking sensitive info.

## Risks & Mitigations

- Risk: Persistent contention causing timeouts.
  - Mitigation: Clear error message indicating concurrent update detected; suggest cooldown and retry later.
- Risk: Unintended retries on non-concurrency errors.
  - Mitigation: Strictly gate retries on `ConcurrentUpdateBucketFailed` or `NeedRetry(err)`; otherwise non-retryable.
- Risk: Latency regression in non-contention path.
  - Mitigation: No sleeps unless retry branch is taken; normal path unaffected.

## References

- Terraform retry pattern in repository (`resource.Retry`, `IsExpectedErrors`, `NeedRetry`).
- OSS error code semantics: 409 concurrent updates require cooldown.
