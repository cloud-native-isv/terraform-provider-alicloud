# Quickstart: Adapting to CreateInstance API

**Feature**: Use CreateInstance API
**Date**: 2026-01-16

## Overview

This feature switches the underlying API for `alicloud_instance` creation from `RunInstances` to `CreateInstance`. While the Terraform user interface (HCL) remains largely backward compatible, there are key behavioral changes developers must be aware of during maintenance.

## Key Changes for Developers

1.  **Stop-Start Cycle**:
    -   Instances are now created in `Stopped` state.
    -   The provider automatically calls `StartInstance`.
    -   **Debugging**: If creation hangs, check if `StartInstance` failed or if the instance is stuck in a transition state.

2.  **Public IP Addresses**:
    -   `CreateInstance` **DOES NOT** allocate Public IPs.
    -   Previous behavior for `internet_max_bandwidth_out > 0` allocated an IP.
    -   **New Behavior**: This configuration setting will set bandwidth limits but users **MUST** verify if an IP is actually assigned (it likely won't be unless creating in Classic network or specific scenarios, but generally for VPC, use `alicloud_eip`).

3.  **Security Groups**:
    -   Creation is atomic only for the *first* security group.
    -   Subsequent groups are joined via API calls.
    -   **Failure Scenario**: If `JoinSecurityGroup` fails for the 2nd group, the instance exists but has incomplete security configuration. Terraform taint mechanism (partial state) handles this, but be aware of "half-configured" instances in console during debugging.

## Verification

To verify the implementation:
1.  Run `make testacc TEST=TestAccAlicloudInstance_basic`
2.  Ensure instance goes to `Running` state.
3.  Ensure all parameters (Disks, Tags) are correctly set.
