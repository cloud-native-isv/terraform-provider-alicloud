# Research Findings: Kafka Instance API Refactoring

## Decision: Implement comprehensive KafkaService methods using cws-lib-go

### Rationale:
The current `resource_alicloud_alikafka_instance.go` uses direct `client.RpcPost` calls which violates the Architecture Layering Principle. The `service_alicloud_alikafka.go` file already contains a `KafkaService` struct that uses `cws-lib-go`, but it only implements basic CRUD operations. To properly refactor the resource, we need to implement all required API operations in the `KafkaService` using the `cws-lib-go` library.

### Key Findings:

1. **Current Implementation Pattern**: The Kafka instance resource uses a two-step creation process:
   - Step 1: Create order (`CreatePostPayOrder` or `CreatePrePayOrder`)
   - Step 2: Start instance (`StartInstance`)

2. **Required Operations**: The resource needs the following operations to be implemented in `KafkaService`:
   - Order creation: `CreatePostPayOrder`, `CreatePrePayOrder`
   - Instance management: `StartInstance`, `ModifyInstanceName`, `ConvertPostPayOrder`
   - Instance upgrades: `UpgradePostPayOrder`, `UpgradePrePayOrder`, `UpgradeInstanceVersion`, `UpdateInstanceConfig`
   - Resource management: `ChangeResourceGroup`, `EnableAutoGroupCreation`, `EnableAutoTopicCreation`
   - Instance deletion: `ReleaseInstance`, `DeleteInstance`
   - Quota management: `GetQuotaTip`

3. **cws-lib-go Integration**: The `KafkaService` already has a working integration with `cws-lib-go` through the `KafkaAPI` struct. We need to extend this to support all the required operations.

4. **State Management**: The current resource uses `AliKafkaInstanceStateRefreshFunc` for state management. We should maintain this pattern but ensure it works with the new `KafkaService` methods.

5. **Error Handling**: All new methods should follow the standardized error handling patterns using `IsNotFoundError`, `IsAlreadyExistError`, `NeedRetry`, etc.

### Alternatives Considered:

1. **Direct cws-lib-go calls in resource**: This would violate the Architecture Layering Principle by bypassing the Service layer.

2. **Partial refactoring**: Only refactoring some operations would leave the code inconsistent and still violate the principle.

3. **Complete rewrite**: While thorough, this would be unnecessarily complex given that the Service layer structure is already in place.

### Implementation Strategy:

1. **Extend KafkaService**: Add all required methods to the existing `KafkaService` struct in `service_alicloud_alikafka.go`
2. **Maintain backward compatibility**: Ensure all method signatures and return values match the current expectations
3. **Follow strong typing**: Use the strong types provided by `cws-lib-go` instead of `map[string]interface{}`
4. **Preserve existing behavior**: The external behavior of the resource should remain unchanged

### Technical Dependencies:

- The `cws-lib-go` library must support all the required Kafka API operations
- If certain operations are not available in `cws-lib-go`, we may need to implement them using the underlying SDK patterns
- All methods must handle pagination and error recovery as per the API guidelines

### Risk Assessment:

- **High**: If `cws-lib-go` doesn't support all required operations, we may need to fall back to direct SDK calls
- **Medium**: Ensuring all edge cases are handled correctly during the refactoring
- **Low**: Breaking existing functionality (mitigated by maintaining existing tests)

This research resolves all "NEEDS CLARIFICATION" items by providing a clear implementation path for the refactoring.