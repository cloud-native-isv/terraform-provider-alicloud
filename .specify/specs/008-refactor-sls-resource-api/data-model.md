# Data Model and Mapping

## Overview

The refactoring maps Terraform Schema data to `cws-lib-go` strong types.

## Resource: `alicloud_log_alert`

**Source Schema** -> **Target Struct**: `sls.Alert`

| Terraform Field | CWS-Lib-Go Field (`*sls.Alert`) | Type | Notes |
| :--- | :--- | :--- | :--- |
| `project_name` | (Method Arg) | `string` | Passed to Create/Update method |
| `alert_name` | `Name` | `string` | |
| `alert_displayname` | `DisplayName` | `string` | |
| `alert_description` | `Description` | `string` | |
| `state` | `State` | `string` | Default "Enabled" |
| `schedule` | `Schedule` | `*sls.Schedule` | Complex object |
| `configuration` | `Configuration` | `*sls.AlertConfiguration` | Complex object (V2) |

### `sls.Schedule` Mapping

| Terraform Field | CWS-Lib-Go Field | Type |
| :--- | :--- | :--- |
| `type` | `Type` | `string` |
| `interval` | `Interval` | `string` |
| `cron_expression` | `CronExpression` | `string` |
| `day_of_week` | `DayOfWeek` | `int32` |
| `hour` | `Hour` | `int32` |
| `delay` | `Delay` | `int32` |
| `run_immediately` | `RunImmediately` | `bool` |
| `time_zone` | `TimeZone` | `string` |

### `sls.AlertConfiguration` Mapping

| Terraform Field | CWS-Lib-Go Field | Notes |
| :--- | :--- | :--- |
| `version` | `Version` | |
| `type` | `Type` | |
| `threshold` | `Threshold` | |
| `condition` | `Condition` | |
| `dashboard` | `Dashboard` | |
| `query_list` | `QueryList` | `[]*sls.AlertQuery` |
| `notification_list` | `NotificationList` | `[]*sls.Notification` |
| `severity_configurations` | `SeverityConfigurations` | `[]*sls.SeverityConfiguration` |
| `join_configurations` | `JoinConfigurations` | `[]*sls.JoinConfiguration` |
| `group_configuration` | `GroupConfiguration` | `sls.GroupConfiguration` |
| `policy_configuration` | `PolicyConfiguration` | `sls.PolicyConfiguration` |
| `auto_annotation` | `AutoAnnotation` | |
| `send_resolved` | `SendResolved` | |
| `no_data_fire` | `NoDataFire` | |
| `no_data_severity` | `NoDataSeverity` | |

## Data Source: `alicloud_sls_machine_groups`

**Target Struct**: `sls.MachineGroup` from `ListMachineGroups`

| Terraform Field | CWS-Lib-Go Field | Notes |
| :--- | :--- | :--- |
| `project` | (Arg) | |
| `machine_group_type` | `MachineGroupType` | Filter locally if API doesn't support |
| `names` | `[]string` | Filter output |

## Data Source: `alicloud_sls_logtail_config`

**Target Struct**: `sls.LogConfig` from `ListConfig`

| Terraform Field | CWS-Lib-Go Field | Notes |
| :--- | :--- | :--- |
| `project` | (Arg) | |
| `logstore` | (Arg) | |
| `names` | `[]string` | Filter output |
