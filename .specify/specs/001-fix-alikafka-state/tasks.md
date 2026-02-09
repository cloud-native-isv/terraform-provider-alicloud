# Task Checklist - Fix Alikafka State

**Feature**: 009-AliKafka Instance Management
**Review**: [requirements.md](requirements.md), [plan.md](plan.md)

## Phase 1: Investigation (US1)

Goal: Understand the root cause of missing fields and prepare the fix strategy.

- [ ] T001 [US1] Analyze `alicloud/service_alicloud_alikafka.go` to determine if `DescribeInstance` returns the full instance details (including endpoints and network info).

## Phase 2: Implementation (US1)

Goal: Fix the state population logic in Service and Resource layers.

- [ ] T002 [US1] Update `alicloud/service_alicloud_alikafka.go` (DescribeInstance) to ensure it returns all required fields (endpoints, configs, tags, etc.) if they are missing from the current response.
- [ ] T003 [US1] Update `alicloud/resource_alicloud_alikafka_instance.go` in the `Read` function to map all missing fields from the Service response to Terraform state (`d.Set`). Attributes to populate include `vpc_id`, `zone_id`, `security_group`, `server_version`, `config`, and all `*_endpoint` fields.

## Phase 3: Verification (US1)

Goal: Verify that the state is correctly populated.

- [ ] T004 [US1] Run acceptance test `TestAccAlicloudAlikafkaInstance_basic` to ensure no regression and check if state population improves.
- [ ] T005 [US1] Perform manual verification using `terraform show` (refer to `quickstart.md`) to confirm that fields like `domain_endpoint`, `ssl_domain_endpoint`, `vpc_id`, etc., are no longer null/empty.

## Phase 4: Polish

Goal: Code cleanup and final quality check.

- [ ] T006 Ensure `make build` passes and code is compliant with project standards.
- [ ] T007 Run `make lint` (if available) or check for static analysis warnings.

## Dependencies

- US1: All tasks are sequential. T003 depends on T002 findings (or implementation). T004/T005 depend on T003.

## Implementation Strategy

- Focus on **accuracy**: Ensure we don't break existing `id` or `status` fields.
- Use **defensive programming**: Handle potential nil pointers in the response from SDK/Service layer.
