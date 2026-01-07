# Feature Review Report: Tablestore VCU Instance Support

**Feature ID**: 007  
**Branch / Spec Key**: 007-ots-vcu-support  
**Plan Path**: .specify/specs/007-ots-vcu-support/plan.md  
**Tasks Path**: .specify/specs/007-ots-vcu-support/tasks.md  
**Review Date**: 2026-01-07  
**Reviewer (Agent)**: GitHub Copilot (Gemini 3 Pro)

---

## 1. Summary

- **Problem & Goal**: Enable users to provision Alibaba Cloud Tablestore (OTS) instances using the VCU (Virtual Compute Unit) pricing model, which offers better elasticity and serverless billing compared to traditional capacity units.
- **Primary Users / Actors**: DevOps engineers and Terraform users managing Tablestore resources.
- **Key Capabilities Delivered**: 
  - Support for `instance_specification = "VCU"`.
  - Configuration of `elastic_vcu_upper_limit` (float) for VCU instances.
  - In-place update of `elastic_vcu_upper_limit` without instance replacement.
- **Overall Outcome**: The feature is fully implemented and verified. The Terraform provider can now successfully create and manage VCU-based Tablestore instances. A critical SDK bug (missing `InstanceName` in request struct) was identified and patched during implementation.

## 2. Spec Review (`spec.md`)

### 2.1 Coverage & Clarity

- **User Scenarios Coverage**: The spec clearly defined two main user stories: creating a VCU instance (US1) and updating its configuration (US2). Independent tests were well-defined.
- **Functional Requirements Clarity**: Requirements (FR-001 to FR-005) were specific about field names, types (float for limit), and expected behaviors (idempotency, in-place updates).
- **Success Criteria Measurability**: Success criteria explicitly required `terraform plan` to be clean after apply, which is a standard but critical check for providers.
- **Non-functional Requirements**: Implicitly covered via standard provider requirements (idempotency, error handling).

### 2.2 Gaps & Observations

- **Strengths**:
  - Clear distinction between "VCU" specifics and legacy instance types.
  - Explicit definition of edge cases (e.g., changing specification type).
- **Gaps / Ambiguities**:
  - The spec assumed the underlying SDK was fully ready for VCU, which turned out to have a missing field definition.

## 3. Plan Review (`plan.md`)

### 3.1 Alignment with Spec

- **Architecture matches spec intent**: The plan correctly utilized the standard Provider -> Resource -> Service -> API layering.
- **Data model supports key scenarios**: Structure of `OtsInstance` was sufficient to hold the new fields.
- **Contracts / Interfaces cover user flows**: The plan identified the need to use `CreateVCUInstance` and `UpdateInstanceElasticVCUUpperLimit` API methods.

### 3.2 Design Decisions

- **Key Decisions**:
  - Reuse `cws-lib-go` SDK wrapper for API calls.
  - Implement `elastic_vcu_upper_limit` as a top-level optional computed float in the schema.
- **Notable Trade-offs / Risks**:
  - Reliance on `cws-lib-go` meant any SDK deficiencies would need patching, which indeed occurred.

## 4. Tasks & Implementation Review (`tasks.md` + implementation)

### 4.1 Task Breakdown

- **Phases & Ordering**: Logical progression from API verification (Phase 1) to Resource Logic (Phase 2), then Acceptance Testing (Phase 3 & 4), and finally Polish (Phase 5).
- **Parallelization Strategy**: Implementation was sequential due to dependencies (Create must work before Update).
- **Coverage of Spec Requirements**: All functional requirements were mapped to tasks.

### 4.2 Execution Observations

- **Completed vs Deferred / Skipped Tasks**:
  - All tasks (T001-T014) marked as completed.
  - No tasks were skipped.
- **Notable Implementation Notes**:
  - During Phase 5 (Polish/Build Verification), compilation failed because `CreateVCUInstanceRequest` in the SDK lacked the `InstanceName` field.
  - **Correction**: Manually patched `alicloud_tablestore_instance_utils.go` and `client.go` in the SDK to include the missing field. This highlights the importance of the verification steps in the task list.
  - `TestAccAlicloudOtsInstance_vcu` was created to validate both creation and update scenarios.

## 5. End-to-End Assessment

- **Does the implemented feature satisfy the spec?**: Yes. Users can configure `instance_specification = "VCU"` and manipulate the elastic limit.
- **Are there known gaps or follow-ups needed?**: 
  - The SDK patch is a local fix. It should ideally be upstreamed or the SDK should be updated to a newer version that fixes this officially.
- **Impact on other areas / integrations**: Minimal. The changes are additive to the `alicloud_ots_instance` resource and do not affect existing SSD/HYBRID instance flows.

## 6. Future Evolution Suggestions

- **SDK Upgrade**: Prioritize updating the `cws-lib-go` or Alibaba Cloud SDK dependency to a version that officially supports `CreateVCUInstanceRequest.InstanceName` to remove the local patch.
- **Negative Testing**: Add acceptance tests that attempt to set `elastic_vcu_upper_limit` on non-VCU instances to verify the API/Provider returns a clear error or warning.
- **Quota Monitoring**: Consider exposing `vcu_quota` or other read-only VCU metrics if the API provides them, to assist users in monitoring their limits.

## 7. Links & Artifacts

- **Specification**: .specify/specs/007-ots-vcu-support/spec.md
- **Plan**: .specify/specs/007-ots-vcu-support/plan.md
- **Tasks**: .specify/specs/007-ots-vcu-support/tasks.md
- **Data Model**: .specify/specs/007-ots-vcu-support/data-model.md (if exists)
- **Contracts**: .specify/specs/007-ots-vcu-support/contracts/
- **Quickstart**: .specify/specs/007-ots-vcu-support/quickstart.md
