# Implementation Plan - Implement alicloud_ecs_instance Resource

## Technical Context

**Unknowns & Risks**:
- **UNKNOWN**: Specific parameters required by `CreateInstance` that might differ from `RunInstances`.
- **UNKNOWN**: How to effectively wait for network interface readiness during creation (API signal or polling method).
- **RISK**: Ensuring clean subset schema doesn't miss critical fields users expect from `alicloud_instance` (VPC focus).

**Technical Decisions**:
- **Pattern**: Layered Architecture (Resource -> Service -> API/SDK).
- **Library**: `CWS-Lib-Go` for API interactions.
- **Strategy**: "Create then Attach" for secondary resources (disks, ENIs).
- **State**: Explicit wait for `Running` state and Network readiness.
- **Defaults**: Rely on API defaults where possible.
- **Type Safety**: Use CWS-Lib-Go strong types; no `map[string]interface{}`.

## Constitution Check

| Principle | Check | Context |
|---|---|---|
| **I. Layering** | ✅ | Resource calls Service layer, Service calls CWS-Lib-Go API. |
| **II. State Mgmt** | ✅ | Create waits for `Running` & Network; Read used to sync state. |
| **III. Errors** | ✅ | Use `alicloud/errors.go` (IsNotFoundError, etc.). |
| **IV. Quality** | ✅ | Naming conventions (`alicloud_ecs_instance`), `Id` fields. |
| **V. Strong Types** | ✅ | Strict use of CWS-Lib-Go structs. |
| **VI. Testing** | ✅ | Integration tests, `make` verification mandatory. |

## Phased Execution

### Phase 1: Research & Design

**Goal**: Validate API requirements and design schema.

- [ ] **Task 1.1**: Research `CreateInstance` vs `RunInstances` API parameters.
  - *Context*: Identify mandatory fields for VPC instances in `CreateInstance`.
  - *Output*: List of fields for schema.
- [ ] **Task 1.2**: Research Network Readiness check mechanism.
  - *Context*: Determine best way to confirm primary ENI is attached and ready.
  - *Output*: Strategy for `WaitForNetwork`.
- [ ] **Task 1.3**: Define Resource Schema (Clean Subset).
  - *Context*: Map `alicloud_instance` fields to `CreateInstance` params, excluding legacy.
  - *Output*: `data-model.md` with schema definition.
- [ ] **Task 1.4**: Define Service Layer Interface.
  - *Context*: `CreateContext`, `DescribeContext`, `DeleteContext`, `WaitFor`.
  - *Output*: Service method signatures in `data-model.md`.

### Phase 2: Foundation & Service Layer

**Goal**: Implement Service layer and foundational types.

- [ ] **Task 2.1**: Implement `EcsService.CreateInstance` using CWS-Lib-Go.
  - *File*: `alicloud/service_alicloud_ecs.go`
  - *Details*: Wrap SDK call, standard error handling.
- [ ] **Task 2.2**: Implement `EcsService.DescribeInstance` & Pagination.
  - *File*: `alicloud/service_alicloud_ecs.go`
  - *Details*: Handle pagination, return `[]*ecs.Instance`.
- [ ] **Task 2.3**: Implement State Refresh & Wait Functions.
  - *File*: `alicloud/service_alicloud_ecs.go`
  - *Details*: `WaitForInstanceRunning`, `WaitForNetworkReady`.

### Phase 3: Resource Implementation

**Goal**: Implement the Terraform resource.

- [ ] **Task 3.1**: Create Resource Skeleton `resource_alicloud_ecs_instance.go`.
  - *File*: `alicloud/resource_alicloud_ecs_instance.go`
  - *Details*: Define Schema, basic CRUD function stubs.
- [ ] **Task 3.2**: Implement `Create` Logic.
  - *Details*: Input validation -> Service.Create -> Wait Network -> Wait Running -> Attach Disks/ENIs.
- [ ] **Task 3.3**: Implement `Read` Logic.
  - *Details*: Service.Describe -> d.Set.
- [ ] **Task 3.4**: Implement `Update` & `Delete` Logic.
  - *Details*: Basic updates, graceful deletion with wait.
- [ ] **Task 3.5**: Register Resource in `alicloud/provider.go`.

### Phase 4: Testing & Verification

**Goal**: Verify functionality.

- [ ] **Task 4.1**: Create Acceptance Test.
  - *File*: `alicloud/resource_alicloud_ecs_instance_test.go`
  - *Details*: Basic Create-Read-Delete lifecycle test.
- [ ] **Task 4.2**: Verify User Stories.
  - *Details*: Manual verification with TF config (Start from Stopped, Updates).
- [ ] **Task 4.3**: Code Cleanup & Review.
  - *Details*: `make` check, linting, formatting.
