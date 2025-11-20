# Data Model: Kafka Provider Refactoring

## Entities

### KafkaInstance
**Description**: Represents an Alibaba Cloud Kafka instance
**Fields**:
- `InstanceId` (string, required): Unique identifier for the Kafka instance
- `RegionId` (string, required): Region where the instance is located
- `Name` (string, optional): Instance name
- `DiskType` (string, optional): Disk type (e.g., "cloud_ssd", "cloud_efficiency")
- `DiskSize` (int, optional): Disk size in GB
- `DeployType` (int, optional): Deployment type (0=classic, 1=vpc, 2=serverless)
- `PaidType` (int, optional): Billing type (0=postpaid, 1=prepaid, 3=serverless, 4=confluent)
- `IoMax` (int, optional): Maximum I/O capacity
- `IoMaxSpec` (string, optional): I/O specification
- `ServiceStatus` (int, optional): Service status (5=running, 10=released)
- `VpcId` (string, optional): VPC ID for VPC deployment
- `VSwitchId` (string, optional): VSwitch ID for VPC deployment
- `SecurityGroupId` (string, optional): Security group ID
- `Tags` (map[string]string, optional): Resource tags

**State Transitions**:
- Creating → Running (ServiceStatus transitions from creating states to 5)
- Running → Deleting → Deleted (ServiceStatus transitions to 10)

**Validation Rules**:
- InstanceId must be non-empty
- RegionId must be valid Alibaba Cloud region
- DiskSize must be within allowed range for instance type
- IoMax must be compatible with instance specifications

### KafkaTopic
**Description**: Represents a Kafka topic within an instance
**Fields**:
- `InstanceId` (string, required): Parent instance ID
- `Topic` (string, required): Topic name
- `PartitionNum` (int, optional): Number of partitions
- `Remark` (string, optional): Topic description
- `CompactTopic` (bool, optional): Whether topic is compacted

**State Transitions**:
- Creating → Existing (topic appears in list)
- Existing → Deleting → Deleted (topic disappears from list)

**Validation Rules**:
- Topic name must follow Kafka naming conventions
- PartitionNum must be positive integer
- InstanceId must reference valid existing instance

### KafkaConsumerGroup
**Description**: Represents a Kafka consumer group
**Fields**:
- `InstanceId` (string, required): Parent instance ID  
- `ConsumerId` (string, required): Consumer group ID
- `Remark` (string, optional): Consumer group description

**State Transitions**:
- Creating → Existing (consumer group appears in list)
- Existing → Deleting → Deleted (consumer group disappears from list)

**Validation Rules**:
- ConsumerId must be non-empty string
- InstanceId must reference valid existing instance

### KafkaSaslUser
**Description**: Represents a SASL user for authentication
**Fields**:
- `InstanceId` (string, required): Parent instance ID
- `Username` (string, required): SASL username
- `Mechanism` (string, optional): Authentication mechanism

**State Transitions**:
- Creating → Existing (user appears in SASL user list)
- Existing → Deleting → Deleted (user disappears from list)

**Validation Rules**:
- Username must be non-empty and follow SASL naming rules
- InstanceId must reference valid existing instance

### KafkaSaslAcl
**Description**: Represents SASL ACL (Access Control List) rules
**Fields**:
- `InstanceId` (string, required): Parent instance ID
- `Username` (string, required): SASL username
- `AclResourceType` (string, required): Resource type (TOPIC, GROUP, CLUSTER, TRANSACTIONAL_ID)
- `AclResourceName` (string, required): Resource name (topic name, group name, etc.)
- `AclResourcePatternType` (string, required): Pattern type (LITERAL, PREFIXED)
- `AclOperationType` (string, required): Operation type (READ, WRITE, CREATE, DELETE, ALTER, DESCRIBE, ALL)

**State Transitions**:
- Creating → Existing (ACL rule appears in list)
- Existing → Deleting → Deleted (ACL rule disappears from list)

**Validation Rules**:
- All fields must be non-empty and valid
- Resource type must be one of supported types
- Operation type must be compatible with resource type
- InstanceId must reference valid existing instance

### KafkaAllowedIp
**Description**: Represents IP allowlist configuration for Kafka instances
**Fields**:
- `InstanceId` (string, required): Parent instance ID
- `AllowedType` (string, required): Type of allowed access (vpc, internet)
- `PortRange` (string, required): Port range (e.g., "9092/9092")
- `AllowedIpList` ([]string, required): List of allowed IP addresses/CIDR blocks

**State Transitions**:
- Creating → Existing (IP rule appears in allowlist)
- Existing → Deleting → Deleted (IP rule disappears from allowlist)

**Validation Rules**:
- AllowedType must be "vpc" or "internet"
- PortRange must be valid port specification
- AllowedIpList must contain valid IP addresses or CIDR blocks
- InstanceId must reference valid existing instance

## Relationships

- **KafkaInstance** (1) → (N) **KafkaTopic**: Each instance can have multiple topics
- **KafkaInstance** (1) → (N) **KafkaConsumerGroup**: Each instance can have multiple consumer groups  
- **KafkaInstance** (1) → (N) **KafkaSaslUser**: Each instance can have multiple SASL users
- **KafkaInstance** (1) → (N) **KafkaSaslAcl**: Each instance can have multiple SASL ACLs
- **KafkaInstance** (1) → (N) **KafkaAllowedIp**: Each instance can have multiple IP allowlist entries

## Composite IDs

The following composite ID patterns will be used for resource identification:

- **ConsumerGroupID**: `InstanceId:ConsumerId`
- **TopicID**: `InstanceId:Topic`  
- **SaslUserID**: `InstanceId:Username`
- **SaslAclID**: `InstanceId:Username:AclResourceType:AclResourceName:AclResourcePatternType:AclOperationType`
- **AllowedIpID**: `InstanceId:AllowedType:PortRange:IpAddress`

## Service Layer Interface

The `KafkaService` will provide the following methods:

### Instance Methods
- `DescribeKafkaInstance(id string) (*kafka.Instance, error)`
- `CreateKafkaInstance(request *kafka.CreateInstanceRequest) (*kafka.Instance, error)`  
- `DeleteKafkaInstance(id string) error`
- `KafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc`
- `WaitForKafkaInstanceCreating(id string, timeout time.Duration) error`
- `WaitForKafkaInstanceDeleting(id string, timeout time.Duration) error`

### Topic Methods
- `DescribeKafkaTopic(id string) (*kafka.Topic, error)`
- `CreateKafkaTopic(request *kafka.CreateTopicRequest) error`
- `DeleteKafkaTopic(id string) error`
- `KafkaTopicStateRefreshFunc(id string) resource.StateRefreshFunc`
- `WaitForKafkaTopicCreating(id string, timeout time.Duration) error`
- `WaitForKafkaTopicDeleting(id string, timeout time.Duration) error`

### Consumer Group Methods
- `DescribeKafkaConsumerGroup(id string) (*kafka.ConsumerGroup, error)`
- `CreateKafkaConsumerGroup(request *kafka.CreateConsumerGroupRequest) error`
- `DeleteKafkaConsumerGroup(id string) error`
- `KafkaConsumerGroupStateRefreshFunc(id string) resource.StateRefreshFunc`
- `WaitForKafkaConsumerGroupCreating(id string, timeout time.Duration) error`
- `WaitForKafkaConsumerGroupDeleting(id string, timeout time.Duration) error`

### SASL User Methods
- `DescribeKafkaSaslUser(id string) (*kafka.SaslUser, error)`
- `CreateKafkaSaslUser(request *kafka.CreateSaslUserRequest) error`
- `DeleteKafkaSaslUser(id string) error`
- `KafkaSaslUserStateRefreshFunc(id string) resource.StateRefreshFunc`
- `WaitForKafkaSaslUserCreating(id string, timeout time.Duration) error`
- `WaitForKafkaSaslUserDeleting(id string, timeout time.Duration) error`

### SASL ACL Methods
- `DescribeKafkaSaslAcl(id string) (*kafka.SaslAcl, error)`
- `CreateKafkaSaslAcl(request *kafka.CreateSaslAclRequest) error`
- `DeleteKafkaSaslAcl(id string) error`
- `KafkaSaslAclStateRefreshFunc(id string) resource.StateRefreshFunc`
- `WaitForKafkaSaslAclCreating(id string, timeout time.Duration) error`
- `WaitForKafkaSaslAclDeleting(id string, timeout time.Duration) error`

### Allowed IP Methods
- `DescribeKafkaAllowedIp(id string) (*kafka.AllowedIp, error)`
- `CreateKafkaAllowedIp(request *kafka.CreateAllowedIpRequest) error`
- `DeleteKafkaAllowedIp(id string) error`
- `KafkaAllowedIpStateRefreshFunc(id string) resource.StateRefreshFunc`
- `WaitForKafkaAllowedIpCreating(id string, timeout time.Duration) error`
- `WaitForKafkaAllowedIpDeleting(id string, timeout time.Duration) error`

## Error Handling

Standard error predicates will be used:
- `IsNotFoundError(err)`: Resource does not exist
- `IsAlreadyExistError(err)`: Resource already exists  
- `NeedRetry(err)`: Operation should be retried
- Predefined error lists for service-specific errors

All errors will be wrapped using `WrapError(err)` or `WrapErrorf(err, format, args...)` with appropriate context.

## Updated Data Model Based on cws-lib-go Types

### KafkaInstance (based on cws-lib-go)
**Description**: Represents an Alibaba Cloud Kafka instance resource
**Fields**:
- `InstanceId` (string, required): Unique identifier for the instance
- `Name` (string, required): Name of the instance (max 64 characters)
- `Description` (string, optional): Description of the instance
- `Status` (string, computed): Status of the instance (Active, Inactive, Creating, Deleting, etc.)
- `RegionId` (string, required): Region ID
- `ZoneId` (string, required): Availability zone ID
- `SpecType` (string, required): Instance specification type
- `DeployType` (int32, required): Deployment type
- `DiskSize` (int32, required): Disk size in GB (must be > 0)
- `DiskType` (string, required): Disk type
- `IoMax` (int32, required): Maximum IO specification
- `IoMaxSpec` (string, optional): IO specification
- `Version` (string, required): Kafka version
- `EndPoint` (string, computed): Access endpoint
- `CreateTime` (string, computed): Creation time
- `ExpireTime` (string, computed): Expiration time
- `SslEndPoint` (string, computed): SSL access endpoint
- `SaslEndPoint` (string, computed): SASL access endpoint

**Validation Rules**:
- Name must be between 1 and 64 characters
- DiskSize must be greater than 0
- RegionId and ZoneId are required

**State Transitions**:
- Creating → Running (Status 5)
- Running → Deleting
- Deleting → Deleted/Released (Status 10)

### KafkaTopic (based on cws-lib-go)
**Description**: Represents a Kafka topic resource within an instance
**Fields**:
- `Topic` (string, required): Topic name
- `InstanceId` (string, required): ID of the instance to which the topic belongs
- `PartitionNum` (int32, required): Number of partitions
- `Remark` (string, optional): Remark
- `CreateTime` (int64, computed): Creation timestamp
- `Status` (int32, computed): Status
- `CompactTopic` (bool, optional): Indicates whether it is a Compact topic
- `LocalTopic` (bool, optional): Indicates whether it is a local topic
- `ReplicaNum` (int32, required): Number of replicas (default: 3)
- `Tags` (map[string]string, optional): Tags

**Validation Rules**:
- Topic name is required
- PartitionNum must be positive
- ReplicaNum should be at least 1

**State Transitions**:
- Existence-based (topic exists = ready, doesn't exist = deleted)

### ConsumerGroup (based on cws-lib-go)
**Description**: Represents a Kafka consumer group
**Fields**:
- `ConsumerGroupId` (string, required): Consumer group ID
- `InstanceId` (string, required): ID of the instance to which the consumer group belongs
- `ConsumerGroupName` (string, optional): Consumer group name
- `ConsumerGroupType` (string, optional): Consumer group type
- `ConsumerGroupStatus` (string, computed): Consumer group status
- `ConsumerGroupDescription` (string, optional): Consumer group description
- `ConsumerGroupTags` ([]Tag, optional): Consumer group tags
- `ConsumerGroupCreateTime` (string, computed): Creation time
- `ConsumerGroupUpdateTime` (string, computed): Update time

**Validation Rules**:
- ConsumerGroupId is required
- InstanceId is required

**State Transitions**:
- Existence-based (consumer group exists = ready, doesn't exist = deleted)

### AclRule (based on cws-lib-go)
**Description**: Represents a Kafka access control rule
**Fields**:
- `AclResourceType` (string, required): ACL resource type
- `AclResourceName` (string, required): ACL resource name
- `AclResourcePatternType` (string, required): ACL resource pattern type
- `AclOperation` (string, required): ACL operation type
- `AclPermissionType` (string, required): ACL permission type
- `AclHost` (string, required): ACL host
- `AclPrincipal` (string, required): ACL principal
- `InstanceId` (string, required): ID of the instance to which the ACL rule belongs

**Validation Rules**:
- All fields are required
- InstanceId is required

**State Transitions**:
- Existence-based (ACL exists = ready, doesn't exist = deleted)

### Tag (based on cws-lib-go)
**Description**: Represents a tag
**Fields**:
- `Key` (string, required): Tag key
- `Value` (string, required): Tag value

## ID Encoding/Decoding Functions

The service layer will implement the following ID encoding/decoding functions:

### Consumer Groups
- `EncodeConsumerGroupId(instanceId, consumerId string) string`
- `DecodeConsumerGroupId(id string) (string, string, error)`

### Topics
- `EncodeTopicId(instanceId, topic string) string`
- `DecodeTopicId(id string) (string, string, error)`

### SASL Users
- `EncodeSaslUserId(instanceId, username string) string`
- `DecodeSaslUserId(id string) (string, string, error)`

### SASL ACLs
- `EncodeSaslAclId(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string) string`
- `DecodeSaslAclId(id string) (string, string, string, string, string, string, error)`

### IP Attachments
- `EncodeAllowedIpId(instanceId, allowedType, portRange, ipAddress string) string`
- `DecodeAllowedIpId(id string) (string, string, string, string, error)`