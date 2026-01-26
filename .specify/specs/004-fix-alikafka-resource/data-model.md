# Data Model: AliKafka Resource Implementation

## Entities

### AliKafkaInstance

Represents a Kafka instance in Alibaba Cloud with attributes including deployment type, disk configuration, network settings, security configuration, and operational status.

**Fields**:
- `InstanceId` (string): Unique identifier for the instance
- `Name` (string): Name of the instance
- `Description` (string): Description of the instance  
- `Status` (string): Status of the instance (Active, Inactive, Creating, Deleting, etc.)
- `RegionId` (string): Region ID
- `ZoneId` (string): Availability zone ID
- `SpecType` (string): Instance specification type
- `DeployType` (int32): Deployment type
- `DiskSize` (int32): Disk size in GB
- `DiskType` (string): Disk type
- `IoMax` (int32): Maximum IO specification
- `IoMaxSpec` (string): IO specification
- `Version` (string): Kafka version
- `EndPoint` (string): Access endpoint
- `SslEndPoint` (string): SSL access endpoint
- `SaslEndPoint` (string): SASL access endpoint
- `CreateTime` (string): Creation time
- `ExpireTime` (string): Expiration time

**Derived Fields** (from basic fields):
- `vswitch_ids` (list of string): Derived from VSwitchId
- `selected_zones` (list of list of string): Derived from ZoneId
- `cross_zone` (bool): Determined from deployment configuration

**Quota Fields** (if available in API response):
- `topic_num_of_buy` (int)
- `topic_used` (int) 
- `topic_left` (int)
- `partition_used` (int)
- `partition_left` (int)
- `group_used` (int)
- `group_left` (int)
- `is_partition_buy` (int)

### AliKafkaDeployment

Represents the deployment configuration of a Kafka instance, including VPC/VSwitch settings, zone configuration, and runtime parameters.

**Fields**:
- `InstanceId` (string): Reference to the AliKafkaInstance
- `VpcId` (string): VPC ID
- `VSwitchId` (string): VSwitch ID  
- `ZoneId` (string): Zone ID
- `Name` (string): Deployment name
- `CrossZone` (bool): Whether to deploy across zones
- `SecurityGroup` (string): Security group ID
- `ServiceVersion` (string): Service version
- `Config` (string): Configuration as JSON string
- `KMSKeyId` (string): KMS key ID
- `SelectedZones` (list of list of string): Selected zones for deployment
- `VSwitchIds` (list of string): VSwitch IDs
- `EipMax` (int): Maximum EIP count

## State Transitions

### KafkaInstanceState

Represents the lifecycle state of a Kafka instance, transitioning through order processing, creation, configuration, startup, and running states.

**Valid States**:
- `0` = Order Processing (waiting for instance ID assignment)
- `2` = Creating (infrastructure provisioning)  
- `3` = Configuring (applying instance settings)
- `4` = Starting (service initialization)
- `5` = Running (fully operational)

**State Transitions**:
- Initial: No instance exists
- → State 0: Order created/instance creation initiated
- → State 2: Infrastructure provisioning started
- → State 3: Configuration being applied
- → State 4: Service initialization in progress  
- → State 5: Fully operational
- → Deleted: Instance terminated

**Error States**:
- Failed states during any transition should be handled with appropriate error classification (IsNotFoundError, IsAlreadyExistError, NeedRetry)

## Validation Rules

- **InstanceId**: Required, non-empty string
- **DeployType**: Must be in [4, 5] (as per current schema validation)
- **DiskSize**: Required, positive integer
- **DiskType**: Required, non-empty string
- **PaidType**: Must be "PrePaid" or "PostPaid"
- **PartitionNum**: Optional, but if provided must be positive integer
- **EipMax**: Optional, but validation depends on DeployType (suppressed when DeployType == 5)
- **ZoneId**: Must be valid availability zone in the specified region
- **VSwitchId**: Must be valid VSwitch in the specified VPC and zone

## Relationships

- **AliKafkaInstance** ↔ **AliKafkaDeployment**: One-to-one relationship
  - Each instance has exactly one deployment configuration
  - Deployment operations (StartInstance/StopInstance) operate on the instance
- **AliKafkaInstance** → **Tags**: One-to-many relationship
  - Each instance can have multiple tags
  - Tags are managed through the standard tagging interface