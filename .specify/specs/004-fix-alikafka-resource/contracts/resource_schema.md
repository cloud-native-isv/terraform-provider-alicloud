# Resource Schema Contracts

## alicloud_alikafka_instance

### Input Schema (Create/Update)

**Required Fields**:
- `deploy_type` (int): Deployment type, must be 4 or 5
- `disk_size` (int): Disk size in GB
- `disk_type` (string): Disk type
- `paid_type` (string): Payment type, must be "PrePaid" or "PostPaid"

**Optional Fields**:
- `io_max_spec` (string): IO specification
- `spec_type` (string): Instance specification type, defaults to "normal"
- `partition_num` (int): Number of partitions
- `eip_max` (int): Maximum EIP count (validation suppressed when deploy_type == 5)
- `duration` (int): Duration for prepaid instances
- `resource_group_id` (string): Resource group ID
- `tags` (map[string]string): Resource tags
- `config` (string): Configuration as JSON string

### Output Schema (Read)

**Computed Fields**:
- `name` (string): Instance name
- `security_group` (string): Security group ID
- `service_version` (string): Service version
- `config` (string): Configuration as JSON string
- `kms_key_id` (string): KMS key ID
- `vpc_id` (string): VPC ID
- `zone_id` (string): Zone ID
- `vswitch_id` (string): VSwitch ID
- `enable_auto_group` (bool): Auto group creation enabled
- `enable_auto_topic` (string): Auto topic creation setting
- `default_topic_partition_num` (int): Default topic partition number
- `vswitch_ids` (list of string): Derived from vswitch_id
- `selected_zones` (list of list of string): Derived from zone_id
- `cross_zone` (bool): Cross-zone deployment status
- `end_point` (string): Access endpoint
- `ssl_endpoint` (string): SSL access endpoint
- `domain_endpoint` (string): Domain endpoint
- `ssl_domain_endpoint` (string): SSL domain endpoint
- `sasl_domain_endpoint` (string): SASL domain endpoint
- `topic_num_of_buy` (int): Topic quota purchased
- `topic_used` (int): Topics used
- `topic_left` (int): Topics remaining
- `partition_used` (int): Partitions used
- `partition_left` (int): Partitions remaining
- `group_used` (int): Consumer groups used
- `group_left` (int): Consumer groups remaining
- `is_partition_buy` (int): Partition purchase status
- `status` (int): Instance status

## alicloud_alikafka_deployment

### Input Schema (Create)

**Required Fields**:
- `instance_id` (string): Reference to existing AliKafka instance
- `vswitch_id` (string): VSwitch ID for deployment

**Optional Fields**:
- `vpc_id` (string): VPC ID (auto-detected if not provided)
- `zone_id` (string): Zone ID
- `name` (string): Deployment name
- `cross_zone` (bool): Cross-zone deployment, defaults to true
- `security_group` (string): Security group ID
- `service_version` (string): Service version
- `config` (string): Configuration as JSON string
- `kms_key_id` (string): KMS key ID
- `selected_zones` (list of list of string): Selected zones for deployment
- `vswitch_ids` (list of string): VSwitch IDs

### Output Schema (Read)

**Computed Fields**:
- `instance_id` (string): Instance ID
- `name` (string): Deployment name
- `vpc_id` (string): VPC ID
- `vswitch_id` (string): VSwitch ID
- `zone_id` (string): Zone ID
- `security_group` (string): Security group ID
- `config` (string): Configuration as JSON string
- `kms_key_id` (string): KMS key ID
- `vswitch_ids` (list of string): VSwitch IDs
- `eip_max` (int): Maximum EIP count

## Service Layer Interface Contracts

### KafkaService Interface

**Methods**:
- `CreateInstance(instance *kafka.KafkaInstance) (*kafka.KafkaInstance, error)`
- `GetInstance(instanceId string) (*kafka.KafkaInstance, error)`
- `UpdateInstance(instance *kafka.KafkaInstance) error`
- `StartInstance(request *StartInstanceRequest) error`
- `StopInstance(request *StopInstanceRequest) error`
- `UpdateInstanceConfig(instanceId string, config map[string]*string) error`
- `UpgradeInstanceVersion(instanceId string, targetVersion string) error`
- `DescribeTags(resourceId string, resourceType string) ([]*Tag, error)`
- `setInstanceTags(d *schema.ResourceData, resourceType string) error`

**State Management Methods**:
- `AliKafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc`
- `WaitForAliKafkaInstanceCreating(id string, timeout time.Duration) error`
- `WaitForAliKafkaInstanceUpdating(id string, timeout time.Duration) error`

### API Layer Interface

**CWS-Lib-Go Kafka API Methods**:
- `CreateInstance(instance *KafkaInstance) (*KafkaInstance, error)`
- `GetInstance(instanceId string) (*KafkaInstance, error)`
- `UpdateInstance(instance *KafkaInstance) error`
- `StartInstance(instanceId, regionId, vpcId, vswitchId string, options map[string]interface{}) error`
- `StopInstance(instanceId, regionId string) error`
- `UpdateInstanceConfig(instanceId, regionId string, config map[string]*string) error`
- `UpgradeInstanceVersion(instanceId, regionId, targetVersion string) error`

## Error Handling Contracts

**Standard Error Patterns**:
- `IsNotFoundError(err)`: Resource not found
- `IsAlreadyExistError(err)`: Resource already exists
- `NeedRetry(err)`: Operation should be retried
- `WrapError(err)`: Wrap errors with context
- `WrapErrorf(err, format, args...)`: Wrap errors with formatted context

**Retryable Errors**:
- `ServiceUnavailable`
- `ThrottlingException`
- `InternalError`
- `Throttling`
- `SystemBusy`
- `OperationConflict`