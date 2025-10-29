# Research: Add Terraform resource alicloud_sls_consumer_group

## Decisions & Rationale

### D1. Identifier Mutability & Updates
- Decision: `project`, `logstore`, `consumer_group` are immutable (ForceNew); `timeout`, `order` are updatable in-place.
- Rationale: Matches Terraform best practices for identity vs behavior; prevents ambiguous migrations; aligns with Constitution naming and state control.
- Alternatives: Allow renames/moves in-place (risk drift/migration complexity) → rejected.

### D2. Import/ID Encoding Format
- Decision: `project:logstore:consumer_group` (colon-separated triple).
- Rationale: Consistent with Service-layer ID encoding guidance; simple to parse and validate.
- Alternatives: Slash-delimited, region-prefixed, JSON → added complexity without clear benefit.

### D3. Create Idempotency on Existing Groups
- Decision: Adopt if exists and converge `timeout`/`order` to HCL.
- Rationale: Improves UX; avoids manual import; ensures post-apply no drift; ForceNew fields still enforce replacement when changed.
- Alternatives: Fail and force import; success without convergence (leads to drift) → rejected.

### D4. Architecture & Integration Path
- Decision: Resource → Service → CWS-Lib-Go API → SDK; no direct SDK calls in Resource.
- Rationale: Constitution I (Layering); reliability, retry & error uniformity.
- Alternatives: Direct aliyun-log-go-sdk usage in Resource → violates layering.

### D5. State Management
- Decision: Implement `*StateRefreshFunc` and `WaitFor*` in Service; Create/Delete use WaitFor; Read sets all computed fields; no Read in Create path.
- Rationale: Constitution II; predictable convergence; better error handling.

### D6. Error Handling & Retry
- Decision: Use `WrapError/WrapErrorf`, `IsNotFoundError/IsAlreadyExistError/NeedRetry` with standard retryable errors (Throttling, SystemBusy, etc.).
- Rationale: Constitution III; consistent operator feedback and resilience.

## Best Practices & Patterns
- Validation: Local schema validation for required fields and naming constraints to fail fast.
- Timeouts: Expose Create/Update/Delete timeouts with sensible defaults; honor in state waits and retries.
- Import: Validate triple format; produce actionable error messages when invalid.
- Docs: Clear examples incl. import ID and ForceNew behavior.

## Open Questions (Resolved by Spec Clarifications)
- None pending; all critical clarifications (mutability, import ID, adopt behavior) resolved.
