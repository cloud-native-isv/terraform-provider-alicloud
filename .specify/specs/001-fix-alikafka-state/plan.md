# Implementation Plan - Fix Alikafka State

**Goal**: Populate missing attributes in `alicloud_alikafka_instance` state by updating the reading logic.
**Feature**: 009-AliKafka Instance Management
**Spec**: [requirements.md](requirements.md)

## Technical Context

### Background
The current `alicloud_alikafka_instance` resource fails to persist several fields to state after creation/read. This includes network details (`vpc_id`, `zone_id`) and endpoints (`domain_endpoint`, `ssl_endpoint`). These fields are crucial for downstream dependencies.

### Architecture Analysis
- **Resource Layer**: `alicloud/resource_alicloud_alikafka_instance.go`
  - Defines schema and handles Create/Read/Update/Delete.
  - Currently maps fields from `kafkaService.DescribeInstance` result.
- **Service Layer**: `alicloud/service_alicloud_alikafka.go`
  - `DescribeInstance` wraps `s.kafkaApi.GetInstance`.
  - There is also `DescribeAlikafkaInstance` which uses `GetInstanceList`.
- **API Internal**: `cws-lib-go/lib/cloud/aliyun/api/kafka` (Internal library)
  - Provides `GetInstance` method.

### Root Cause Hypothesis
The `DescribeInstance` method creates a `GetInstanceRequest` or similar. If it's using an underlying API that only returns summary info (like `GetInstanceList` with a filter but without full details), fields like endpoints might be missing. Alternatively, the Struct Definition in `cws-lib-go` might be missing JSON tags or fields, or standard mapping in `resource_alicloud_alikafka_instance.go` is incomplete.

Since we cannot easily modify `cws-lib-go` if it's an external module (unless it's in vendor), we assume we must ensure:
1. We are calling the correct API that returns these fields.
2. If `cws-lib-go` is incomplete, we might need to fix it or fallback to the official Alibaba Cloud SDK for this specific call if strictly necessary (though Constitution says prefer cws-lib-go). *Correction*: Constitution II & IV implies using Service Layer and CWS-Lib-Go. If CWS-Lib-Go is broken, we should fix the integration or logic in the Service layer.

### Strategy
1. **Investigate**: Check `alicloud/service_alicloud_alikafka.go`. The method `DescribeInstance` (singular) seems to use `s.kafkaApi.GetInstance`. We need to verify if `s.kafkaApi.GetInstance` returns a populated object.
2. **Fix**:
   - Ensure `DescribeInstance` returns the full set of data.
   - Update `resource_alicloud_alikafka_instance.go` `Read` method to map ALL missing fields.
   - Verify if type conversions (e.g., Integer to String) are handled correctly.

## Constitution Check

| Principle | Status | Note |
|---|---|---|
| **I. Layered Architecture** | ✅ | Maintains Resource -> Service separation. |
| **II. Standardized Interfaces** | ✅ | Standard Terraform Schema used. Strong typing enforced. |
| **III. Test-First** | ✅ | Tasks include verifying with acceptance tests. |
| **IV. Integration** | ✅ | Using Service layer for API interaction. |
| **V. Observability** | ✅ | Logs in Read method are preserved/improved. |

## Implementation Phases

### Phase 1: Investigation & Service Layer Fix
**Objective**: Ensure the Service layer returns complete data.

- [ ] **Task 1**: Analyze `alicloud/service_alicloud_alikafka.go`. Verify `DescribeInstance` logic. Update it to return full instance details if currently returning summary.
- [ ] **Task 2**: Verify field availability in `kafka.KafkaInstance` struct (via code inspection or test).

### Phase 2: Resource Layer Update
**Objective**: Populate Terraform state.

- [ ] **Task 3**: Update `resource_alicloud_alikafka_instance.go`.
  - In `resourceAliCloudAlikafkaInstanceRead`, ensure all fields listed in Requirements (Network Info, Endpoints, Config, etc.) are set into `d.Set()`.
  - Handle null pointers safely.

### Phase 3: Verification
**Objective**: Ensure fix works.

- [ ] **Task 4**: Run Acceptance Test `TestAccAlicloudAlikafkaInstance_basic`.
- [ ] **Task 5**: Verify `terraform show` output contains previously missing fields.

## Artifacts

- `alicloud/resource_alicloud_alikafka_instance.go` (Modified)
- `alicloud/service_alicloud_alikafka.go` (Modified if needed)
