# Quick Start Guide: AliKafka Resource Implementation

## Overview

This guide provides a quick overview of the AliKafka resource implementation for the Terraform Provider Alicloud. The implementation fixes state persistence issues, integrates with the CWS-Lib-Go Kafka API layer, and properly handles Kafka instance state lifecycle management.

## Key Features

1. **Complete State Persistence**: All computed attributes are properly synchronized to Terraform state
2. **Modern API Integration**: Uses only CWS-Lib-Go Kafka API functions (CreateInstance, GetInstance, etc.)
3. **Proper State Lifecycle**: Handles all Kafka instance states (0, 2, 3, 4, 5) correctly
4. **Backward Compatibility**: Maintains compatibility with existing Terraform configurations

## Basic Usage

### Create an AliKafka Instance

```hcl
resource "alicloud_alikafka_instance" "example" {
  deploy_type = 4
  disk_size   = 500
  disk_type   = "cloud_essd"
  paid_type   = "PostPaid"
  
  # Optional configuration
  io_max_spec = "alikafka.hw.2xlarge"
  spec_type   = "normal"
  partition_num = 10
  tags = {
    Environment = "production"
    Team = "data-platform"
  }
}
```

### Deploy an AliKafka Instance

```hcl
resource "alicloud_alikafka_deployment" "example" {
  instance_id = alicloud_alikafka_instance.example.id
  vswitch_id  = "vsw-xxxxxx"
  
  # Optional deployment configuration
  cross_zone = true
  security_group = "sg-xxxxxx"
  service_version = "2.8.0"
}
```

## State Management

The implementation properly handles the following Kafka instance states:

- **State 0**: Order Processing (waiting for instance ID assignment)
- **State 2**: Creating (infrastructure provisioning)
- **State 3**: Configuring (applying instance settings)
- **State 4**: Starting (service initialization)
- **State 5**: Running (fully operational)

Terraform will wait for the appropriate state transitions before proceeding with dependent resources.

## API Integration

The implementation uses the following CWS-Lib-Go Kafka API functions:

- `CreateInstance()`: For initial instance creation
- `GetInstance()`: For reading instance state
- `UpdateInstance()`: For updating instance properties
- `StartInstance()`: For deployment operations
- `StopInstance()`: For undeployment operations
- `UpdateInstanceConfig()`: For configuration updates
- `UpgradeInstanceVersion()`: For version upgrades

## Development Workflow

1. **Testing**: Use `make testacc` to run acceptance tests
2. **Validation**: Run `make` to validate Go code syntax
3. **Build**: Use `make build` to compile the provider
4. **Documentation**: Update resource documentation in `docs/resources/`

## Troubleshooting

- **State Drift**: If you experience state drift, ensure all computed fields are being set in the Read function
- **API Errors**: Check that the correct CWS-Lib-Go API functions are being used
- **State Transitions**: Verify that state refresh functions handle all valid states (0, 2, 3, 4, 5)

## Next Steps

1. Implement the service layer functions in `service_alicloud_kafka.go`
2. Update the resource implementations in `resource_alicloud_alikafka_instance.go` and `resource_alicloud_alikafka_deployment.go`
3. Add comprehensive acceptance tests
4. Update resource documentation
5. Test with real Alibaba Cloud infrastructure