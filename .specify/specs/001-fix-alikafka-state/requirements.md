# Requirements Specification: Fix Missing State Fields for Alikafka Instance

**Requirement Branch**: `001-fix-alikafka-state`
**Created**: 2026-02-09
**Status**: Draft

## Feature Overview

The `alicloud_alikafka_instance` resource currently fails to populate several attributes in the Terraform state after creation or refresh. This results in a state file with many `null` or empty values for fields that should contain valid configuration or runtime data (e.g., endpoints, network IDs, configuration settings).

This specification outlines the requirements to ensure that all relevant fields in the `alicloud_alikafka_instance` resource are correctly read from the Alibaba Cloud API and persisted in the Terraform state.

## User Scenarios

### Scenario 1: Resource Creation and Verification
1.  User defines an `alicloud_alikafka_instance` resource in their Terraform configuration.
2.  User runs `terraform apply` to create the instance.
3.  After successful creation, the User runs `terraform show` or checks the `terraform.tfstate`.
4.  **Expectation**: All computed attributes (e.g., `domain_endpoint`, `ssl_domain_endpoint`, `vpc_id`, `zone_id`, `security_group`) are populated with values returned from the API, rather than being null or empty strings.

### Scenario 2: Importing Existing Resources
1.  User has an existing Alikafka instance created outside of Terraform.
2.  User runs `terraform import alicloud_alikafka_instance.example <instance_id>`.
3.  **Expectation**: The resulting state contains full details of the instance, matching the actual cloud resource configuration.

## Functional Requirements

### 1. State Population Implementation
The `resourceAliCloudAlikafkaInstance` implementation (specifically the `Read` function, and `Create`/`Update` where applicable) must be updated to map API response fields to the following Terraform schema attributes:

*   **Network Info**: `vpc_id`, `vswitch_id`, `zone_id`, `security_group`
*   **Endpoints**: `domain_endpoint`, `ssl_domain_endpoint`, `sasl_domain_endpoint`, `end_point`, `ssl_endpoint`
*   **Configuration**: `config`, `default_topic_partition_num`, `enable_auto_group`, `enable_auto_topic`, `service_version`
*   **Status/Usage (if applicable as Computed fields)**: `topic_left`, `topic_used`, `partition_left`, `partition_used`, `group_left`, `group_used`
*   **Other Metadata**: `kms_key_id`, `cross_zone`

### 2. Data Consistency
*   Ensure that type conversions between the Go SDK response and Terraform schema are handled correctly (e.g., integer to string if required, or matching types).
*   Ensure that fields that are optional or conditionally returned are handled gracefully (i.e., set to appropriate zero-values or null if not present).

## Success Criteria

1.  **State Completeness**: For a standard Alikafka Professional/Basic instance, the following fields must not be null/empty in the state after a refresh:
    *   `vpc_id`, `vswitch_id`, `zone_id`
    *   `domain_endpoint` (and other endpoints if SSL/SASL enabled)
    *   `security_group`
2.  **No Regression**: Existing state management for critical fields (id, name, deploy_type) must remain unchanged.
3.  **Testability**: Acceptance tests should verify that these attributes are non-empty after resource creation.

## Assumptions

*   The fields listed in the "missing" set are already defined in the Terraform Schema for `alicloud_alikafka_instance`.
*   The Alibaba Cloud SDK for Go provides these fields in the `DescribeInstances` or similar API response.

## Technical Notes (Optional)
*   The provided state dump indicates Schema version 0.
*   The fix likely resides in `alicloud/resource_alicloud_alikafka_instance.go`.
