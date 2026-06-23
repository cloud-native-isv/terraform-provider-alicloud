# Requirements Specification: ADBPG Terraform Resources & Data Sources

**Requirement Branch**: `008-adbpg-resources`  
**Created**: 2026-06-22  
**Status**: Draft  
**Input**: User description: "需要基于pkg/cws-lib-go子模块最新代码(commit id 0b9fa3aaa7730925d06eb9003301b63111e7fc67 和commit aa4d7742d2f33491fc1392d04bc098e2fa27fa10)实现新的adbpg相关的resource和datasource. 旧的名称叫做gpdb(来源于greenplum database),新实现的resource使用adbpg作为名称. 可以在项目根目录使用`ls -l alicloud/*gpdb*`查看旧的resource和datasource都有哪些,然后结合pkg/cws-lib-go中最新的API层的定义,设计新的resource和datasource."

## Related Feature *(mandatory)*

**Feature ID**: 012  
**Feature Name**: ADBPG Resource Management

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage ADBPG Instance Lifecycle (Priority: P1)

As a cloud infrastructure operator, I want to use Terraform to create, configure, modify, and destroy AnalyticDB PostgreSQL instances so that I can manage my analytical database infrastructure as code.

**Why this priority**: Instance management is the foundational capability. All other ADBPG resources (accounts, databases, connections) depend on having an instance. Without this, no other user stories can function.

**Independent Test**: Can be fully tested by running `terraform plan` and `terraform apply` with an instance resource definition, verifying the instance is created, attributes are readable, modifications take effect, and destroy removes it.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with `alicloud_adbpg_instance`, **When** `terraform apply` is run, **Then** an ADBPG instance is created with the specified engine version, zone, VPC, instance class, and segment node count, and all attributes are populated in state.
2. **Given** an existing ADBPG instance managed by Terraform, **When** the description or maintenance window is changed and `terraform apply` is run, **Then** the instance is updated in-place without recreation.
3. **Given** an existing ADBPG instance managed by Terraform, **When** `terraform destroy` is run, **Then** the instance is deleted and removed from state.
4. **Given** an ADBPG instance in "Creating" status, **When** Terraform reads the resource, **Then** it waits for the instance to reach "Running" status before completing.

---

### User Story 2 - Manage ADBPG Account Lifecycle (Priority: P1)

As a database administrator, I want to manage database accounts (superuser and normal) via Terraform so that access control is codified and auditable.

**Why this priority**: Accounts are required for any database interaction and are a prerequisite for connecting to the instance. This is a core lifecycle resource.

**Independent Test**: Can be fully tested by creating an instance, then creating/modifying/deleting an account resource, verifying account appears in account listing.

**Acceptance Scenarios**:

1. **Given** a running ADBPG instance, **When** a `alicloud_adbpg_account` resource is applied, **Then** the account is created with the specified name, password, type, and description.
2. **Given** an existing account, **When** the password or description is changed, **Then** the account is updated accordingly.
3. **Given** an existing account, **When** `terraform destroy` is run, **Then** the account is deleted.

---

### User Story 3 - Manage ADBPG Database (Priority: P2)

As a database administrator, I want to create and manage databases within an ADBPG instance via Terraform so that database provisioning is automated.

**Why this priority**: Databases are the logical containers for data within an instance. Managing them via Terraform enables repeatable environment setup.

**Independent Test**: Can be tested by creating an instance, then applying a database resource and verifying it appears in the database listing.

**Acceptance Scenarios**:

1. **Given** a running ADBPG instance, **When** a `alicloud_adbpg_database` resource is applied, **Then** the database is created with the specified name, character set, and description.
2. **Given** an existing database, **When** `terraform destroy` is run, **Then** the database is deleted.

---

### User Story 4 - Manage ADBPG Connection (Public Endpoint) (Priority: P2)

As a network engineer, I want to allocate and release public connection endpoints for ADBPG instances via Terraform so that external access is managed as code.

**Why this priority**: Public connections enable external access to instances and are a common operational requirement for hybrid connectivity.

**Independent Test**: Can be tested by creating an instance, allocating a public connection, verifying the connection string is populated, then releasing it on destroy.

**Acceptance Scenarios**:

1. **Given** a running ADBPG instance, **When** a `alicloud_adbpg_connection` resource is applied, **Then** a public connection endpoint is allocated with the specified prefix, and the full connection string and port are readable.
2. **Given** an existing public connection, **When** `terraform destroy` is run, **Then** the public connection is released.

---

### User Story 5 - Manage ADBPG Security IP Whitelist (Priority: P2)

As a security engineer, I want to manage IP whitelists for ADBPG instances via Terraform so that network access control is version-controlled.

**Why this priority**: Security IP management is critical for controlling who can connect to database instances.

**Independent Test**: Can be tested by creating an instance, applying an IP whitelist resource, verifying the security IPs are set, then modifying and verifying the update.

**Acceptance Scenarios**:

1. **Given** a running ADBPG instance, **When** a `alicloud_adbpg_security_ip_array` resource is applied, **Then** the security IP list is configured for the specified IP array group.
2. **Given** an existing security IP configuration, **When** the IP list is modified, **Then** the update takes effect.

---

### User Story 6 - Query ADBPG Instances (Data Source) (Priority: P2)

As a Terraform user, I want to query existing ADBPG instances by various filters (status, description, tags) so that I can reference them in other resource configurations.

**Why this priority**: Data sources enable read-only discovery and cross-referencing, which is essential for modular Terraform configurations.

**Independent Test**: Can be tested by creating instances and then using the data source to query and verify results match the filters.

**Acceptance Scenarios**:

1. **Given** existing ADBPG instances, **When** a `data.alicloud_adbpg_instances` data source is applied with a filter, **Then** it returns matching instances with their attributes (ID, description, status, engine version, VPC, etc.).
2. **Given** no matching instances, **When** the data source is queried, **Then** an empty result set is returned without error.

---

### User Story 7 - Query ADBPG Accounts (Data Source) (Priority: P3)

As a Terraform user, I want to list accounts on an ADBPG instance so that I can reference them in configurations.

**Why this priority**: Supports cross-referencing and auditing of existing accounts.

**Independent Test**: Can be tested by creating accounts on an instance and querying the data source.

**Acceptance Scenarios**:

1. **Given** an ADBPG instance with accounts, **When** `data.alicloud_adbpg_accounts` is applied, **Then** it returns all accounts with name, type, status, and description.

---

### User Story 8 - Manage ADBPG Backup Policy (Priority: P3)

As an infrastructure operator, I want to configure the backup policy for an ADBPG instance via Terraform so that backup schedules and retention are codified.

**Why this priority**: Backup policies are important for data protection but not required for core instance operation.

**Independent Test**: Can be tested by creating an instance and applying a backup policy resource, verifying the backup period and time are set.

**Acceptance Scenarios**:

1. **Given** a running ADBPG instance, **When** a `alicloud_adbpg_backup_policy` resource is applied, **Then** the backup retention period, preferred backup period, and preferred backup time are configured.
2. **Given** an existing backup policy, **When** settings are changed, **Then** the update takes effect.

---

### Edge Cases

- What happens when an instance is in a transitional state (e.g., "Creating", "Restarting") during a Terraform operation? The provider must wait for the instance to reach a stable state before proceeding or reporting an error.
- What happens when a user tries to delete an account that is currently in use? The API should return an appropriate error, and the provider should surface it to the user.
- What happens when `terraform import` is used to import an existing ADBPG resource? The provider must support importing by resource ID and correctly populate all state attributes.
- What happens when the ADBPG instance is locked (e.g., overdue payment)? The provider should detect the lock mode and report it clearly rather than failing silently.

## Requirements *(mandatory)*

### Functional Requirements

#### Resources

- **FR-001**: Provider MUST register a `alicloud_adbpg_instance` resource supporting full CRUD lifecycle (create, read, update, delete) with all attributes defined in the `AdbpgInstanceCreate` and `AdbpgInstanceDetail` API types, including: region_id, zone_id, engine_version, db_instance_class, db_instance_mode, instance_network_type, vpc_id, vswitch_id, pay_type, description, security_ip_list, storage_size, storage_type, master_node_num, seg_node_num, resource_group_id, serverless_mode, encryption_key, encryption_type, and tags.
- **FR-002**: Provider MUST register a `alicloud_adbpg_account` resource supporting CRUD for database accounts, with attributes: account_name, account_password, account_type, account_description. Password updates MUST use the ResetAccountPassword API.
- **FR-003**: Provider MUST register a `alicloud_adbpg_database` resource supporting create and delete for databases within an instance, with attributes: db_name, db_description, character_name.
- **FR-004**: Provider MUST register a `alicloud_adbpg_connection` resource supporting allocate (create) and release (delete) of public connection endpoints, with attributes: connection_string_prefix and computed outputs: connection_string, port, ip_address.
- **FR-005**: Provider MUST register a `alicloud_adbpg_security_ip_array` resource for managing IP whitelist groups, with attributes: security_ip_list, db_instance_ip_array_name.
- **FR-006**: Provider MUST register a `alicloud_adbpg_backup_policy` resource for managing backup configuration, with attributes: backup_retention_period, preferred_backup_period, preferred_backup_time, enable_recovery_point.
- **FR-007**: Provider MUST register a `alicloud_adbpg_ssl` resource for managing SSL configuration, with attributes: ssl_enabled (boolean). Read via DescribeInstanceSSL, modify via ModifyInstanceSSL.

#### Data Sources

- **FR-008**: Provider MUST register a `data.alicloud_adbpg_instances` data source returning instances matching filters (description, ids, status, tags, resource_group_id), with all attributes from `AdbpgInstance`.
- **FR-009**: Provider MUST register a `data.alicloud_adbpg_accounts` data source returning accounts for a given instance, with attributes from `AdbpgAccount`.
- **FR-010**: Provider MUST register a `data.alicloud_adbpg_zones` data source returning available zones and supported engine/instance-class combinations from ListAvailableResources.

#### Architecture Constraints

- **FR-011**: All new resource and data source implementations MUST follow the layered architecture: Resource/DataSource -> Service (`service_alicloud_adbpg.go`) -> API (`pkg/cws-lib-go/lib/cloud/aliyun/api/adbpg`). Direct SDK calls from the resource layer are forbidden.
- **FR-012**: The service layer (`service_alicloud_adbpg.go`) MUST implement `StateRefreshFunc` and `WaitFor*` methods for instance status polling. Resource `Create` methods MUST NOT contain inline polling loops.
- **FR-013**: All request and response types MUST use strongly-typed structs from `cws-lib-go`. The `map[string]interface{}` pattern is forbidden in new code.
- **FR-014**: Pagination MUST be fully encapsulated in the API/Service layer and never exposed to resource callers.
- **FR-015**: The naming convention MUST use `adbpg` (not `gpdb`) for all new file names, resource names, and function names: `resource_alicloud_adbpg_*.go`, `data_source_alicloud_adbpg_*.go`, `service_alicloud_adbpg.go`.

#### Connectivity

- **FR-016**: Provider connectivity layer (`alicloud/connectivity/`) MUST be extended to support creating `AdbpgAPI` client instances via `cws-lib-go`, using existing credential and endpoint management.

### Key Entities

- **ADBPG Instance**: Central resource representing an AnalyticDB PostgreSQL database cluster. Attributes include instance class, engine version, network configuration (VPC/VSwitch), payment type, storage configuration, and segment topology.
- **ADBPG Account**: Database user account within an instance. Identified by account_name + db_instance_id. Has type (Super/Normal), status, and password.
- **ADBPG Database**: Logical database within an instance. Identified by db_name + db_instance_id. Has character set encoding and description.
- **ADBPG Connection**: Public network endpoint for an instance. Provides connection_string, IP address, and port for external access.
- **ADBPG Security IP Array**: Named group of IP addresses/CIDRs allowed to access an instance.
- **ADBPG Backup Policy**: Per-instance backup schedule and retention configuration.
- **ADBPG SSL Config**: Per-instance SSL encryption enablement state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 7 new resources (`alicloud_adbpg_instance`, `alicloud_adbpg_account`, `alicloud_adbpg_database`, `alicloud_adbpg_connection`, `alicloud_adbpg_security_ip_array`, `alicloud_adbpg_backup_policy`, `alicloud_adbpg_ssl`) pass full CRUD acceptance tests.
- **SC-002**: All 3 new data sources (`data.alicloud_adbpg_instances`, `data.alicloud_adbpg_accounts`, `data.alicloud_adbpg_zones`) return correct results in acceptance tests.
- **SC-003**: No resource or data source implementation contains direct SDK calls; all go through the service layer and `cws-lib-go` API.
- **SC-004**: `terraform import` works for all resources, correctly populating state from the remote API.
- **SC-005**: Provider builds successfully on all target platforms (Linux amd64, macOS amd64, macOS ARM64) with `make build`.
- **SC-006**: All new code passes `make fmt`, `make fmtcheck`, and `make vet` without warnings or errors.

### Measurement Sources & Collection Methods

- **SC-001 Source**: Acceptance test results (`make testacc` with `TF_ACC=1`, filtered to `TestAccAlicloudAdbpg*`). Measured per PR.
- **SC-002 Source**: Acceptance test results for data source tests (`TestAccAlicloudAdbpg*DataSource*`). Measured per PR.
- **SC-003 Source**: Code review and static analysis (`grep` for direct SDK imports in resource files). Verified during review.
- **SC-004 Source**: Import test cases within acceptance tests. Measured per PR.
- **SC-005 Source**: CI build output from `make build`. Measured per PR.
- **SC-006 Source**: CI lint output from `make fmtcheck` and `make vet`. Measured per PR.

## Shared Strings

| String ID | Value (verbatim) | Consumed by |
|-----------|------------------|-------------|
| `STR-001` | "alicloud_adbpg_instance" | FR-001, SC-001 |
| `STR-002` | "alicloud_adbpg_account" | FR-002, SC-001 |
| `STR-003` | "alicloud_adbpg_database" | FR-003, SC-001 |
| `STR-004` | "alicloud_adbpg_connection" | FR-004, SC-001 |
| `STR-005` | "alicloud_adbpg_security_ip_array" | FR-005, SC-001 |
| `STR-006` | "alicloud_adbpg_backup_policy" | FR-006, SC-001 |
| `STR-007` | "alicloud_adbpg_ssl" | FR-007, SC-001 |
| `STR-008` | "alicloud_adbpg_instances" | FR-008, SC-002 |
| `STR-009` | "alicloud_adbpg_accounts" | FR-009, SC-002 |
| `STR-010` | "alicloud_adbpg_zones" | FR-010, SC-002 |

## Clarifications

### Session 2026-06-22

- Q: Which Feature should this spec be bound to? → A: Feature 012 — ADBPG Resource Management
