# Tasks: Implement KafkaService methods using cws-lib-go

**Feature Branch**: `005-implement-kafka-service`
**Status**: In Progress

## Phase 1: Setup
*Goal: Initialize project structure and dependencies*

- [ ] T001 Create `alicloud/service_alicloud_alikafka_types.go` with request structs (`StartInstanceRequest`, `ModifyInstanceNameRequest`, `UpgradeInstanceVersionRequest`)
- [ ] T002 Update `KafkaService` struct in `alicloud/service_alicloud_alikafka.go` to include `cws-lib-go` `KafkaAPI` client
- [ ] T003 Update `NewKafkaService` in `alicloud/service_alicloud_alikafka.go` to initialize `KafkaAPI` using credentials

## Phase 2: Foundational
*Goal: Core infrastructure and shared components*

- [ ] T004 Ensure `github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka` is imported in `alicloud/service_alicloud_alikafka.go`

## Phase 3: User Story 1 - Create Kafka Instance
*Goal: Enable creation of PrePaid and PostPaid instances using strong types*

**Independent Test**: Verify `CreatePostPayOrder` and `CreatePrePayOrder` invoke the API correctly.

- [ ] T005 [US1] Implement `CreatePostPayOrder` in `alicloud/service_alicloud_alikafka.go` accepting `*kafka.KafkaOrder`
- [ ] T006 [US1] Implement `CreatePrePayOrder` in `alicloud/service_alicloud_alikafka.go` accepting `*kafka.KafkaOrder`
- [ ] T007 [US1] Update `resourceAliCloudAlikafkaInstanceCreate` in `alicloud/resource_alicloud_alikafka_instance.go` to construct `kafka.KafkaOrder` and call `CreatePostPayOrder`
- [ ] T008 [US1] Update `resourceAliCloudAlikafkaInstanceCreate` in `alicloud/resource_alicloud_alikafka_instance.go` to construct `kafka.KafkaOrder` and call `CreatePrePayOrder`

## Phase 4: User Story 2 - Manage Kafka Instance Lifecycle
*Goal: Enable Start, Modify, and Upgrade operations using strong types*

**Independent Test**: Verify lifecycle methods invoke the API correctly.

- [ ] T009 [US2] Implement `StartInstance` in `alicloud/service_alicloud_alikafka.go` accepting `*StartInstanceRequest`
- [ ] T010 [US2] Update `resourceAliCloudAlikafkaInstanceCreate` in `alicloud/resource_alicloud_alikafka_instance.go` to construct `StartInstanceRequest` and call `StartInstance`
- [ ] T011 [US2] Implement `ModifyInstanceName` in `alicloud/service_alicloud_alikafka.go` accepting `*ModifyInstanceNameRequest`
- [ ] T012 [US2] Update `resourceAliCloudAlikafkaInstanceUpdate` in `alicloud/resource_alicloud_alikafka_instance.go` to construct `ModifyInstanceNameRequest` and call `ModifyInstanceName`
- [ ] T013 [US2] Implement `UpgradeInstanceVersion` in `alicloud/service_alicloud_alikafka.go` accepting `*UpgradeInstanceVersionRequest`
- [ ] T014 [US2] Update `resourceAliCloudAlikafkaInstanceUpdate` in `alicloud/resource_alicloud_alikafka_instance.go` to construct `UpgradeInstanceVersionRequest` and call `UpgradeInstanceVersion`
- [ ] T015 [US2] Implement stubs for `UpgradePostPayOrder` and `UpgradePrePayOrder` in `alicloud/service_alicloud_alikafka.go` returning "not implemented" error (as per spec FR-007)

## Final Phase: Polish
*Goal: Cleanup and verification*

- [ ] T016 Run `make` to verify compilation and fix any syntax errors
- [ ] T017 Verify error wrapping consistency across all new methods in `alicloud/service_alicloud_alikafka.go`

## Dependencies

1. **Setup** (T001-T003) must be completed first.
2. **User Story 1** (T005-T008) depends on Setup.
3. **User Story 2** (T009-T015) depends on Setup.

## Parallel Execution Examples

- **User Story 1**: T005 and T006 can be implemented in parallel.
- **User Story 2**: T009, T011, and T013 can be implemented in parallel.

## Implementation Strategy

1. **Setup**: Establish the types and API client connection.
2. **Create**: Enable instance creation first as it's the entry point.
3. **Lifecycle**: Enable management operations.
4. **Verify**: Ensure everything compiles and errors are handled correctly.
