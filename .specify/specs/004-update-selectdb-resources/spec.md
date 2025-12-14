# Feature Specification: Update SelectDB Resources

**Feature Branch**: `004-update-selectdb-resources`  
**Created**: 2025-12-14  
**Status**: Draft  
**Input**: User description: "根据aliyun/api/selectdb/（cws-lib-go）目录中最新的selectdb api实现来更新，alicloud/*selectdb*中的terraform resource和datasource实现。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage SelectDB Instance (Priority: P1)

As a Terraform user, I want to create and update SelectDB instances using all available configuration options provided by the latest Alibaba Cloud API, so that I can fully utilize the service capabilities.

**Why this priority**: Core functionality for using SelectDB.

**Independent Test**: Create a SelectDB instance with new parameters, update it, and verify the changes are reflected in the state and on the cloud.

**Acceptance Scenarios**:

1. **Given** a valid SelectDB instance configuration with new fields, **When** I run `terraform apply`, **Then** the instance is created with the specified configuration.
2. **Given** an existing SelectDB instance, **When** I update the configuration with new fields, **Then** the instance is updated successfully.

---

### User Story 2 - Manage SelectDB Cluster (Priority: P1)

As a Terraform user, I want to create and update SelectDB clusters within an instance, using the latest API parameters.

**Why this priority**: Essential for scaling and managing compute resources within SelectDB.

**Independent Test**: Create a cluster with new parameters, update it, and verify.

**Acceptance Scenarios**:

1. **Given** a SelectDB instance, **When** I define a cluster with new fields, **Then** the cluster is created.
2. **Given** an existing cluster, **When** I update it, **Then** the changes are applied.

---

### User Story 3 - Query SelectDB Information (Priority: P2)

As a Terraform user, I want to query SelectDB instances and clusters using data sources and see all available attributes.

**Why this priority**: Needed for referencing existing resources and auditing.

**Independent Test**: Run `terraform plan` with data sources and check the output.

**Acceptance Scenarios**:

1. **Given** existing SelectDB resources, **When** I use `alicloud_selectdb_instances` or `alicloud_selectdb_clusters`, **Then** the output includes all new attributes.

### Edge Cases

- What happens when new fields are used with an older provider version? (Terraform handles this via schema validation).
- How does the system handle API errors for new fields? (Should return clear error messages).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `alicloud_selectdb_instance` resource MUST support all create/update parameters defined in the latest `cws-lib-go` SelectDB Instance API.
- **FR-002**: The `alicloud_selectdb_cluster` resource MUST support all create/update parameters defined in the latest `cws-lib-go` SelectDB Cluster API.
- **FR-003**: The `alicloud_selectdb_instances` data source MUST expose all attributes returned by the `DescribeDBInstances` API.
- **FR-004**: The `alicloud_selectdb_clusters` data source MUST expose all attributes returned by the `DescribeDBClusters` API.
- **FR-005**: The implementation MUST use the `cws-lib-go` SDK for all API interactions.

### Key Entities *(include if feature involves data)*

- **SelectDB Instance**: Represents a SelectDB instance.
- **SelectDB Cluster**: Represents a compute cluster within an instance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of new API fields in `CreateDBInstance` and `ModifyDBInstanceAttribute` are supported in `alicloud_selectdb_instance`.
- **SC-002**: 100% of new API fields in `CreateDBCluster` and `ModifyDBCluster` are supported in `alicloud_selectdb_cluster`.
- **SC-003**: `terraform plan` executes without errors for configurations using new fields.
- **SC-004**: `make build` passes successfully.

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
