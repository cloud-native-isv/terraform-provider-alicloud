# Research Findings: Implement alicloud_ecs_instance Resource

## Project Context Analysis
The project follows a strict layered architecture (Resource -> Service -> API/SDK) and mandates the use of CWS-Lib-Go for all API interactions. State management must use explicit wait functions. The goal is to implement a new resource `alicloud_ecs_instance` based on `CreateInstance`, avoiding legacy complexities of `alicloud_instance`.

## References
- `alicloud/resource_alicloud_instance.go` (Legacy implementation for reference)
- `pkg/cws-lib-go/lib/cloud/aliyun/api/ecs` (API definitions)
- `docs/development_guide.md` (Architecture and coding standards)

## Decisions

### Schema Subset
- **Decision**: Include only VPC-related fields: `image_id`, `instance_type`, `security_groups`, `vswitch_id`, `instance_name`, `description`, `tags`.
- **Rationale**: Minimal clean start to verify `CreateInstance` workflow.
- **Alternatives considered**: Full copy of `alicloud_instance` schema (rejected due to legacy debt).

### Network Readiness
- **Decision**: Wait for `Running` state first, then verify public IP (if allocated) or primary ENI status.
- **Rationale**: `CreateInstance` returns before networking is fully established. `WaitForInstanceRunning` is a good proxy, but checking specific network attributes adds robustness.

### API Parameters
- **Decision**: `CreateInstance` requires consistent `RegionId`, `ImageId`, `InstanceType`, `SecurityGroupId`, `VSwitchId`.
- **Rationale**: Standard mandatory parameters for VPC instances.

### Defaults
- **Decision**: Do not generate default passwords or names provider-side.
- **Rationale**: Simplify provider logic, rely on cloud defaults.
