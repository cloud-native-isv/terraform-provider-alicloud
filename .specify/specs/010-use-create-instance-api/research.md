# Phase 0: Research

**Feature**: Use CreateInstance API
**Date**: 2026-01-16

## API Analysis: CreateInstance vs RunInstances

### Key Differences

| Feature | `RunInstances` (Legacy) | `CreateInstance` (Target) | Impact |
| :--- | :--- | :--- | :--- |
| **Initial State** | `Running` (configurable) | `Stopped` | Must explicitly call `StartInstance` after creation. |
| **Security Groups** | Multiple (`SecurityGroupIds`) | Single (`SecurityGroupId`) | Must iterate `JoinSecurityGroup` for 2nd+ groups. |
| **Public IP** | Allocate if Bandwidth > 0 | Not supported | Must block or warn if Public IP implied/requested. |
| **Return Value** | InstanceId (List) | InstanceId (Single) | Refactor logic consuming the ID. |
| **Response Type** | Sync/Async | Async (Wait for Stopped) | Need robust polling `WaitForEcsInstance`. |

### Parameter Mapping

*Source: CreateInstance.json*

-   `RegionId`: Required.
-   `ImageId`: Required.
-   `InstanceType`: Required.
-   `SecurityGroupId`: Required (Single).
-   `InstanceName`: Optional.
-   `InternetChargeType`: Optional.
-   `AutoRenew`: Optional.
-   `AutoRenewPeriod`: Optional.
-   `InternetMaxBandwidthIn`: Optional.
-   `InternetMaxBandwidthOut`: Optional. (Caution: sets bandwidth but NO IP).
-   `HostName`: Optional.
-   `Password`: Optional.
-   `ZoneId`: Optional.
-   `SystemDisk.*`: Mapped from `system_disk_...`.
-   `DataDisk`: List, Mapped from `data_disks`.
-   `Tag`: List, Mapped from `tags`.

### Error Handling

Common `CreateInstance` errors to handle:
-   `InvalidInstanceType.NotFound`: Instance type not available.
-   `InvalidSecurityGroupId.NotFound`: Security group issues.
-   `OperationDenied`: Account restrictions.

## Architecture Decisions

### Service Layer Implementation

We will add the following methods to `EcsService` in `alicloud/service_alicloud_ecs.go`:

1.  `CreateInstance(request *ecs.CreateInstanceRequest) (*ecs.CreateInstanceResponse, error)`
    -   Wraps strict type `CreateInstanceRequest`.
    -   Handles generic retry logic.

2.  `StartInstance(id string) error`
    -   Calls `ecs.StartInstance`.
    -   Handles waiting for `Starting`/`Running` is handled by separate Wait function or caller.

### Implementation Strategy

1.  **Preparation**: Create request object from schema data.
2.  **Creation**: Call `service.CreateInstance`.
3.  **Wait**: Call `service.WaitForEcsInstance` for `Stopped` status.
4.  **Security Groups**: Check `security_groups` set. If > 1, loop `service.JoinSecurityGroups`.
5.  **Start**: Call `service.StartInstance`.
6.  **Wait**: Call `service.WaitForEcsInstance` for `Running` status.
7.  **Finalize**: Set ID, call Read.

## Conclusion

The refactor is feasible but requires careful orchestration of the Create-Wait-Start-Wait flow to preserve the synchronous user experience of Terraform.
