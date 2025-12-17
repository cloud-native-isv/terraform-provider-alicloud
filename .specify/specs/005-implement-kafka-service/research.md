# Research: Implement KafkaService methods using cws-lib-go

## API Analysis

### cws-lib-go KafkaAPI

- **CreateOrder**: `func (api *KafkaAPI) CreateOrder(order *KafkaOrder) (string, error)`
  - Supports both PrePaid and PostPaid via `PaidType` field in `KafkaOrder`.
- **StartInstance**: `func (api *KafkaAPI) StartInstance(instanceId, regionId, vpcId, vswitchId string, options map[string]interface{}) error`
  - Requires explicit IDs and a map for optional parameters.
- **ModifyInstanceName**: `func (api *KafkaAPI) ModifyInstanceName(instanceId, regionId, instanceName string) error`
- **UpgradeInstanceVersion**: `func (api *KafkaAPI) UpgradeInstanceVersion(instanceId, regionId, targetVersion string) error`
- **UpgradePostPayOrder / UpgradePrePayOrder**: Not implemented in `cws-lib-go`.

### Current KafkaService

- Methods use `map[string]interface{}` for input and output.
- Methods are stubbed with error messages.

## Decisions

- **Strong Typing**: We will define local structs in `alicloud/service_alicloud_alikafka_types.go` to represent the requests.
  - `StartInstanceRequest`
  - `ModifyInstanceNameRequest`
  - `UpgradeInstanceVersionRequest`
- **Mapping**: `KafkaService` methods will accept these structs, validate them, and map them to the `cws-lib-go` API calls.
- **Unimplemented Methods**: `UpgradePostPayOrder` and `UpgradePrePayOrder` will remain unimplemented (returning error) as the underlying API does not support them yet.
- **Resource Update**: `resource_alicloud_alikafka_instance.go` will be updated to construct these structs.

## Rationale

- **Constitution Compliance**: Strong typing is required.
- **Maintainability**: Structs are easier to maintain and validate than maps.
- **Safety**: Compile-time checks for fields.
