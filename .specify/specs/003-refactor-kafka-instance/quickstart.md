# Quick Start: Kafka Instance Refactor

## Overview

This document provides a quick start guide for the Kafka Instance API refactor feature. The refactor replaces direct `client.RpcPost` calls with `cws-lib-go` API methods in the `alicloud_alikafka_instance` resource.

## Prerequisites

- Go 1.18+
- Terraform 0.12+
- Alibaba Cloud account with Kafka service enabled
- Properly configured Alibaba Cloud credentials

## Feature Components

1. **Resource**: `alicloud_alikafka_instance` - The main Kafka instance resource
2. **Service**: `KafkaService` - Service layer implementation in `service_alicloud_alikafka.go`
3. **API**: `cws-lib-go` - The underlying Alibaba Cloud Kafka API wrapper

## Key Changes

### Before Refactor
- Direct `client.RpcPost` calls in `resource_alicloud_alikafka_instance.go`
- Violation of architecture layering principle

### After Refactor
- Service layer calls using `cws-lib-go` API methods
- Proper architecture layering (Resource → Service → API)
- Strong typing with `cws-lib-go` structures

## Implementation Steps

### 1. Extend KafkaService

Add all required methods to `KafkaService` in `service_alicloud_alikafka.go`:

- Order creation: `CreatePostPayOrder`, `CreatePrePayOrder`
- Instance management: `StartInstance`, `ModifyInstanceName`, `ConvertPostPayOrder`
- Instance upgrades: `UpgradePostPayOrder`, `UpgradePrePayOrder`, `UpgradeInstanceVersion`, `UpdateInstanceConfig`
- Resource management: `ChangeResourceGroup`, `EnableAutoGroupCreation`, `EnableAutoTopicCreation`
- Instance deletion: `ReleaseInstance`, `DeleteInstance`

### 2. Update Resource Layer

Modify `resource_alicloud_alikafka_instance.go` to call `KafkaService` methods instead of direct RPC calls.

### 3. Maintain Compatibility

Ensure all existing functionality is preserved:
- Schema definitions remain unchanged
- State management patterns are maintained
- Error handling follows standardized approaches
- All acceptance tests continue to pass

## Testing

### Unit Tests
Run unit tests to verify individual method implementations:
```bash
make test
```

### Acceptance Tests
Run acceptance tests to verify end-to-end functionality:
```bash
make testacc TEST=./alicloud TESTARGS='-run=TestAccAlicloudAlikafkaInstance'
```

## Verification

### Success Criteria
1. ✅ 100% of `client.RpcPost` calls in `resource_alicloud_alikafka_instance.go` are replaced
2. ✅ Existing acceptance tests pass without modification
3. ✅ Code compiles without errors
4. ✅ Architecture layering principle is followed
5. ✅ Strong typing is used throughout

### Validation Commands
```bash
# Compile check
make

# Run specific tests
make test TEST=./alicloud TESTARGS='-run=TestAccAlicloudAlikafkaInstance'

# Full acceptance test suite
make testacc TEST=./alicloud TESTARGS='-run=TestAccAlicloudAlikafka'
```

## Common Issues

### Compilation Errors
If you encounter compilation errors:
1. Verify all `cws-lib-go` dependencies are correctly imported
2. Check that method signatures match expected patterns
3. Ensure all required fields are properly mapped

### Test Failures
If tests fail:
1. Verify that external behavior is unchanged
2. Check error handling consistency
3. Confirm state management patterns are preserved

## Next Steps

After completing the refactor:
1. Run full test suite
2. Update documentation if needed
3. Submit pull request for review
4. Monitor CI/CD pipeline for any issues