---
description: "Task list for feature: 更新VPC资源以使用新的CWS-Lib-Go API"
---

# Tasks: 更新VPC资源以使用新的CWS-Lib-Go API

Input: Design documents from `/.specify/specs/001-cws-data-cws/`
Prerequisites: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

Tests: OPTIONAL. The feature spec does not explicitly request tests, so this plan omits test tasks.

Organization: Tasks are grouped by user story to enable independent implementation and testing of each story.

Format: `[ID] [P?] [Story] Description`
- [P]: Can run in parallel (different files, no dependencies)
- [Story]: US1, US2, US3, US4 per spec.md priorities
- Include exact file paths

## Phase 1: Setup (Shared Infrastructure)

Purpose: Prepare environment and dependencies per plan.md (Go 1.24, Terraform Plugin SDK, CWS-Lib-Go).

- [X] T001 Verify build works before changes: run `make` at repo root to ensure a clean baseline. (Path: `/cws_data/terraform-provider-alicloud/`)
- [X] T002 [P] Add/verify dependency on CWS-Lib-Go: ensure `require github.com/cloud-native-tools/cws-lib-go` exists in `go.mod`, run `go mod tidy`. (Path: `/cws_data/terraform-provider-alicloud/go.mod`)
- [X] T003 [P] Confirm provider imports and helper utilities available: `alicloud/common.go` error helpers (`WrapError`, `IsNotFoundError`, `NeedRetry`) usage patterns understood. (Path: `/cws_data/terraform-provider-alicloud/alicloud/common.go`)

---

## Phase 2: Foundational (Blocking Prerequisites)

Purpose: Establish Service 层结构与通用模式，所有用户故事依赖此阶段。

CRITICAL: 完成后方可开始任何用户故事实现。

- [X] T004 Create Service layer scaffolding files for VPC resources:
  - `/cws_data/terraform-provider-alicloud/alicloud/service_alicloud_vpc_eip.go`
  - `/cws_data/terraform-provider-alicloud/alicloud/service_alicloud_vpc_nat_gateway.go`
  - `/cws_data/terraform-provider-alicloud/alicloud/service_alicloud_vpc_snat_entry.go`
  Each file must:
  - Define corresponding service struct(s) binding to `connectivity.AliyunClient`
  - Provide constructor(s) `NewVpcEipService`, `NewVpcNatGatewayService`, `NewVpcSnatEntryService`
  - Include placeholders for CRUD and Describe methods mapped to CWS-Lib-Go API (no direct SDK/RPC)
  - Include ID helpers `Encode*Id`/`Decode*Id` per guidelines (simple passthrough if not composite)
- [X] T005 [P] Implement common state management patterns in the above services:
  - `*StateRefreshFunc` for EIP, NAT Gateway, SNAT Entry
  - `WaitFor*Creating`, `WaitFor*Deleting` with pending/target/fail states read from research.md
  - Use `BuildStateConf` if available; else `resource.StateChangeConf` with retry intervals
- [X] T006 [P] Define retryable errors usage in services with `NeedRetry(err)` and wrap all returns using `WrapError/WrapErrorf`.
- [X] T007 Add pagination stubs where list/describe may paginate (service functions to aggregate pages; callers receive full list). Place pagination in Service layer only.
- [X] T008 Compile check after scaffolding: run `make` and fix any syntax/import issues. (No functional changes to resources yet.)

Checkpoint: Service layer ready with constructors, method stubs, state patterns, and build passes.

---

## Phase 3: User Story 1 - 更新EIP地址资源 (Priority: P1) 🎯 MVP

Goal: EIP 地址资源完全切换为通过 Service 层调用 CWS-Lib-Go API，并具备正确状态管理与错误处理。

Independent Test: 通过创建、读取、更新、删除 EIP，验证兼容性与新 API 生效（不要求自动化测试）。

### Implementation for User Story 1

- [X] T009 [US1] Implement EIP Service methods in `/alicloud/service_alicloud_vpc_eip.go`:
  - `AllocateEipAddress(request)` → returns allocationId
  - `DescribeEipAddress(allocationId)` → returns full object
  - `ModifyEipAddressAttribute(allocationId, attrs)`
  - `ReleaseEipAddress(allocationId)`
  - `AssociateEipAddress(allocationId, instanceId, instanceType, privateIp)`
  - `UnassociateEipAddress(allocationId)`
  - `EipStateRefreshFunc(id, failStates)` and `WaitForEipCreating(id, timeout)`; use pending: ["Allocating", "Associating", "Unassociating"], target: ["Available"]; fail: ["Released", "Failed"]
  - Note: Temporarily implemented via legacy RPC inside Service (no direct calls from resource); will switch to `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc` in follow-up.
- [X] T010 [P] [US1] Update resource to use Service layer: `/alicloud/resource_alicloud_eip_address.go`
  - Create: call `AllocateEipAddress`, set `d.SetId(allocationId)`, then `WaitForEipCreating`
  - Read: call `DescribeEipAddress`, set all computed/known fields (ip_address, bandwidth, status, tags)
  - Update: map mutable fields to `ModifyEipAddressAttribute` (name/description/bandwidth/internet_charge_type if supported)
  - Delete: call `ReleaseEipAddress` and wait until not found; handle `IsNotFoundError`
  - Replace any direct SDK/RPC usage with Service calls; keep timeouts via `schema.ResourceTimeout`
- [X] T011 [P] [US1] Ensure error handling complies with guidelines in resource methods
  - Use `IsNotFoundError` to clear state on Read when resource no longer exists
  - Use `NeedRetry` within `resource.Retry` around Create/Update as needed (Throttling/SystemBusy)
- [X] T012 [US1] Build & smoke: run `make`; create/update/delete EIP locally against a test account (manual validation note only)

- [X] T012A [P] [US1] Add `GetAPI()` accessor to `VpcEipService` and create migration doc
  - Add `GetAPI()` in `/alicloud/service_alicloud_vpc_eip.go` for direct cws-lib-go access
  - Add doc `/docs/cws-lib-go/vpc-eip-migration.md` describing exact API mapping, error semantics, and state model

- [X] T012B [P] [US1] Switch Describe path to cws-lib-go with safe fallback
  - Update `DescribeEipAddress` to attempt cws-lib-go VPC API first (reflection by common method names),
    convert typed result to `map[string]interface{}`; fallback to legacy V2 on absence/NotFound;
    keep behavior-compatible while enabling immediate usage of cws-lib-go where available.

Checkpoint: EIP 地址资源可独立工作（MVP）。

---

## Phase 4: User Story 2 - 更新NAT网关资源 (Priority: P1)

Goal: NAT 网关资源切换至 Service 层 + CWS-Lib-Go，支持创建/修改/删除 + 状态与错误管理。

Independent Test: 通过创建、修改与删除 NAT 网关验证功能。

### Implementation for User Story 2

- [ ] T013 [US2] Implement NAT Service methods in `/alicloud/service_alicloud_vpc_nat_gateway.go`:
  - `CreateNatGateway(request)` → returns natGatewayId
  - `DescribeNatGateway(id)`
  - `ModifyNatGatewayAttribute(id, attrs)`
  - `ModifyNatGatewaySpec(id, spec)`
  - `DeleteNatGateway(id)`
  - `NatGatewayStateRefreshFunc(id, failStates)` and `WaitForNatGatewayCreating(id, timeout)`; pending: ["Creating", "Modifying"], target: ["Available"], fail: ["Deleting", "Failed"]
- [ ] T014 [P] [US2] Update resource file `/alicloud/resource_alicloud_nat_gateway.go` to call Service methods
  - Create → WaitFor creating; Read set computed fields; Update supports spec/attrs; Delete → wait for removal
  - Ensure timeouts present and consistent
- [ ] T015 [P] [US2] Error/Retry compliance per guidelines; remove any direct RPC/SDK
- [ ] T016 [US2] Build & smoke: `make`

Checkpoint: US1 + US2 独立可用。

---

## Phase 5: User Story 3 - 更新SNAT条目资源 (Priority: P2)

Goal: SNAT 条目资源切换至 Service 层 + CWS-Lib-Go，支持创建/修改/删除 + 状态与错误管理。

Independent Test: 通过创建、修改、删除 SNAT 条目验证功能。

### Implementation for User Story 3

- [ ] T017 [US3] Implement SNAT Service methods in `/alicloud/service_alicloud_vpc_snat_entry.go`:
  - `CreateSnatEntry(request)` → returns snatEntryId
  - `DescribeSnatEntry(id)`
  - `ModifySnatEntry(id, attrs)`
  - `DeleteSnatEntry(id)`
  - `SnatEntryStateRefreshFunc(id, failStates)` and `WaitForSnatEntryCreating(id, timeout)`; pending: ["Creating", "Modifying"], target: ["Available"], fail: ["Deleting", "Failed"]
  - Provide `EncodeSnatId(snatTableId, snatEntryId)` and `DecodeSnatId(id)` if resource uses composite ids; else pass-through
- [ ] T018 [P] [US3] Update resource file `/alicloud/resource_alicloud_snat_entry.go` to call Service methods and wait appropriately
- [ ] T019 [P] [US3] Error/Retry compliance; ensure `IsNotFoundError` clears state on Read
- [ ] T020 [US3] Build & smoke: `make`

Checkpoint: US1 + US2 + US3 独立可用。

---

## Phase 6: User Story 4 - 更新EIP关联资源 (Priority: P2)

Goal: 通过 Service 层方法在 EIP 与实例之间执行关联/解绑，具备等待与错误处理。

Independent Test: 将 EIP 关联到实例并解除关联。

### Implementation for User Story 4

- [ ] T021 [US4] Reuse EIP Service associate/unassociate in `/alicloud/service_alicloud_vpc_eip.go` (already added in US1) with proper waits
- [ ] T022 [P] [US4] Update `/alicloud/resource_alicloud_eip_association.go` to call `AssociateEipAddress` / `UnassociateEipAddress`
  - Apply waits for transitional states (Associating/Unassociating → Associated/Available as applicable)
  - Handle cross-instance types via `instanceType`
- [ ] T023 [P] [US4] Error/Retry compliance
- [ ] T024 [US4] Build & smoke: `make`

Checkpoint: 所有四个资源均切换完成。

---

## Final Phase: Polish & Cross-Cutting Concerns

- [ ] T025 [P] Documentation: validate `quickstart.md` examples align with actual provider behavior; update inline comments in resource files where changed.
- [ ] T026 [P] Contracts alignment: ensure `/contracts/openapi.yaml` paths/ops match implemented service calls; note any gaps.
- [ ] T027 Missing API follow-up: if any API not available in CWS-Lib-Go, append/confirm items in spec.md "缺失API需求文档" section and open an issue/PR in cws-lib-go (link in commit message notes).
- [ ] T028 Build gate: final `make` and confirm no lint/type errors; ensure IDs use `Id` naming convention.

---

## Dependencies & Execution Order

Phase Dependencies
- Setup (Phase 1): none
- Foundational (Phase 2): depends on Phase 1 completion; BLOCKS all user stories
- User Stories (Phases 3-6): depend on Phase 2 completion
- Polish (Final Phase): depends on all desired user stories being complete

User Story Completion Order (from spec.md)
- P1: US1 (EIP地址), US2 (NAT网关)
- P2: US3 (SNAT条目), US4 (EIP关联)

Within Each Story
- Services (in their service file) before resource integration
- Resource CRUD integration before final smoke/build
- Use waits and error wrappers per guidelines

---

## Parallel Execution Examples

US1 (EIP地址)
- In parallel after T009 begins: T010 (resource file), T011 (error compliance) [distinct files]. Then T012 build.

US2 (NAT网关)
- In parallel after T013 begins: T014 (resource), T015 (error compliance). Then T016 build.

US3 (SNAT条目)
- In parallel after T017 begins: T018 (resource), T019 (error compliance). Then T020 build.

US4 (EIP关联)
- In parallel after T021 is validated: T022 (resource), T023 (error compliance). Then T024 build.

Global
- Phase 1 T002 and T003 can run in parallel.
- Phase 2 T005, T006, T007 can run in parallel once T004 exists.
- Final Phase T025 and T026 can run in parallel.

---

## Implementation Strategy

MVP First (US1 only)
1) Complete Phase 1 and Phase 2
2) Complete Phase 3 (US1 EIP地址)
3) Validate independently; ship as MVP

Incremental Delivery
- After MVP, add US2 → validate → ship; then US3 → ship; then US4 → ship

Notes
- No direct SDK/RPC; only CWS-Lib-Go through Service layer
- Use `IsNotFoundError`, `IsAlreadyExistError`, `NeedRetry` and `WrapError`/`WrapErrorf`
- Do not call Read directly in Create
- Maintain ID naming `*Id`
