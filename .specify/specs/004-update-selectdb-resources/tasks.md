# Tasks: Update SelectDB Resources

**Feature Branch**: `004-update-selectdb-resources`
**Status**: In Progress

## Phase 1: Setup
**Goal**: Initialize project environment and verify prerequisites.

- [ ] T001 Verify `cws-lib-go` dependency is available and up to date in `go.mod`
- [ ] T002 Verify existing SelectDB acceptance tests pass locally in `alicloud/resource_alicloud_selectdb_instance_test.go` (if exists)

## Phase 2: Foundational
**Goal**: Ensure shared components are ready.

- [ ] T003 [P] Review `alicloud/service_alicloud_selectdb_instance.go` to ensure it exposes necessary `cws-lib-go` methods
- [ ] T004 [P] Review `alicloud/service_alicloud_selectdb_cluster.go` to ensure it exposes necessary `cws-lib-go` methods

## Phase 3: User Story 1 - Manage SelectDB Instance (P1)
**Goal**: Enable creation and update of SelectDB instances with all new API fields.
**Independent Test**: Create instance with new fields (e.g., `deploy_scheme`, `tags`), update it, and verify state.

- [ ] T005 [US1] Update schema in `alicloud/resource_alicloud_selectdb_instance.go` to include new fields (e.g., `deploy_scheme`, `resource_group_id`, `tags`)
- [ ] T006 [US1] Update `resourceAliCloudSelectDBInstanceCreate` in `alicloud/resource_alicloud_selectdb_instance.go` to handle new fields
- [ ] T007 [US1] Update `resourceAliCloudSelectDBInstanceRead` in `alicloud/resource_alicloud_selectdb_instance.go` to map response fields to schema
- [ ] T008 [US1] Update `resourceAliCloudSelectDBInstanceUpdate` in `alicloud/resource_alicloud_selectdb_instance.go` to handle updates for new fields
- [ ] T009 [US1] Add acceptance test for SelectDB instance with new fields in `alicloud/resource_alicloud_selectdb_instance_test.go`

## Phase 4: User Story 2 - Manage SelectDB Cluster (P1)
**Goal**: Enable creation and update of SelectDB clusters with all new API fields.
**Independent Test**: Create cluster with new fields (e.g., `params`), update it, and verify state.

- [ ] T010 [US2] Update schema in `alicloud/resource_alicloud_selectdb_cluster.go` to include new fields (e.g., `params`, `charge_type`)
- [ ] T011 [US2] Update `resourceAliCloudSelectDBClusterCreate` in `alicloud/resource_alicloud_selectdb_cluster.go` to handle new fields
- [ ] T012 [US2] Update `resourceAliCloudSelectDBClusterRead` in `alicloud/resource_alicloud_selectdb_cluster.go` to map response fields to schema
- [ ] T013 [US2] Update `resourceAliCloudSelectDBClusterUpdate` in `alicloud/resource_alicloud_selectdb_cluster.go` to handle updates for new fields
- [ ] T014 [US2] Add acceptance test for SelectDB cluster with new fields in `alicloud/resource_alicloud_selectdb_cluster_test.go`

## Phase 5: User Story 3 - Query SelectDB Information (P2)
**Goal**: Expose all new attributes in data sources.
**Independent Test**: Run `terraform plan` using data sources and verify output contains new attributes.

- [ ] T015 [US3] Update schema and read function in `alicloud/data_source_alicloud_selectdb_instances.go` to include new instance attributes
- [ ] T016 [US3] Update schema and read function in `alicloud/data_source_alicloud_selectdb_clusters.go` to include new cluster attributes
- [ ] T017 [US3] Add acceptance test for SelectDB instances data source in `alicloud/data_source_alicloud_selectdb_instances_test.go`
- [ ] T018 [US3] Add acceptance test for SelectDB clusters data source in `alicloud/data_source_alicloud_selectdb_clusters_test.go`

## Final Phase: Polish
**Goal**: Final cleanup and documentation.

- [ ] T019 Run `make build` to ensure no compilation errors
- [ ] T020 Run full acceptance test suite for SelectDB resources
- [ ] T021 Update documentation for `alicloud_selectdb_instance` and `alicloud_selectdb_cluster` (if applicable)

## Dependencies

1. **Phase 1 & 2** must be completed first.
2. **Phase 3 (Instance)** and **Phase 4 (Cluster)** can be executed in parallel, but Cluster tests might depend on an Instance.
3. **Phase 5 (Data Sources)** depends on the API availability but can be implemented in parallel with resources if using mocks or existing resources.

## Parallel Execution Examples

- **Developer A**: Works on **Phase 3 (Instance)** (T005-T009).
- **Developer B**: Works on **Phase 4 (Cluster)** (T010-T014).
- **Developer C**: Works on **Phase 5 (Data Sources)** (T015-T018).

## Implementation Strategy

1.  **MVP**: Complete Phase 3 (Instance) first as it's the parent resource.
2.  **Incremental**: Add Cluster support (Phase 4).
3.  **Finalize**: Add Data Sources (Phase 5) and Polish.
