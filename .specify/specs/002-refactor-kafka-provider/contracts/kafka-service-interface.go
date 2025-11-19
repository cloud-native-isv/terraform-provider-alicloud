package contracts

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// KafkaServiceInterface defines the contract for Kafka service operations
// This interface ensures consistent implementation across all Kafka resources
type KafkaServiceInterface interface {
	// Instance operations
	DescribeKafkaInstance(id string) (*KafkaInstance, error)
	CreateKafkaInstance(request *CreateKafkaInstanceRequest) (*KafkaInstance, error)
	DeleteKafkaInstance(id string) error
	KafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc
	WaitForKafkaInstanceCreating(id string, timeout time.Duration) error
	WaitForKafkaInstanceDeleting(id string, timeout time.Duration) error

	// Topic operations
	DescribeKafkaTopic(id string) (*KafkaTopic, error)
	CreateKafkaTopic(request *CreateKafkaTopicRequest) error
	DeleteKafkaTopic(id string) error
	KafkaTopicStateRefreshFunc(id string) resource.StateRefreshFunc
	WaitForKafkaTopicCreating(id string, timeout time.Duration) error
	WaitForKafkaTopicDeleting(id string, timeout time.Duration) error

	// Consumer group operations
	DescribeKafkaConsumerGroup(id string) (*KafkaConsumerGroup, error)
	CreateKafkaConsumerGroup(request *CreateKafkaConsumerGroupRequest) error
	DeleteKafkaConsumerGroup(id string) error
	KafkaConsumerGroupStateRefreshFunc(id string) resource.StateRefreshFunc
	WaitForKafkaConsumerGroupCreating(id string, timeout time.Duration) error
	WaitForKafkaConsumerGroupDeleting(id string, timeout time.Duration) error

	// SASL user operations
	DescribeKafkaSaslUser(id string) (*KafkaSaslUser, error)
	CreateKafkaSaslUser(request *CreateKafkaSaslUserRequest) error
	DeleteKafkaSaslUser(id string) error
	KafkaSaslUserStateRefreshFunc(id string) resource.StateRefreshFunc
	WaitForKafkaSaslUserCreating(id string, timeout time.Duration) error
	WaitForKafkaSaslUserDeleting(id string, timeout time.Duration) error

	// SASL ACL operations
	DescribeKafkaSaslAcl(id string) (*KafkaSaslAcl, error)
	CreateKafkaSaslAcl(request *CreateKafkaSaslAclRequest) error
	DeleteKafkaSaslAcl(id string) error
	KafkaSaslAclStateRefreshFunc(id string) resource.StateRefreshFunc
	WaitForKafkaSaslAclCreating(id string, timeout time.Duration) error
	WaitForKafkaSaslAclDeleting(id string, timeout time.Duration) error

	// Allowed IP operations
	DescribeKafkaAllowedIp(id string) (*KafkaAllowedIp, error)
	CreateKafkaAllowedIp(request *CreateKafkaAllowedIpRequest) error
	DeleteKafkaAllowedIp(id string) error
	KafkaAllowedIpStateRefreshFunc(id string) resource.StateRefreshFunc
	WaitForKafkaAllowedIpCreating(id string, timeout time.Duration) error
	WaitForKafkaAllowedIpDeleting(id string, timeout time.Duration) error
}

// Request and response types for Kafka operations
// These types should map directly to cws-lib-go strong types

type KafkaInstance struct {
	InstanceId      string
	RegionId        string
	Name            string
	DiskType        string
	DiskSize        int
	DeployType      int
	PaidType        int
	IoMax           int
	IoMaxSpec       string
	ServiceStatus   int
	VpcId           string
	VSwitchId       string
	SecurityGroupId string
	Tags            map[string]string
}

type CreateKafkaInstanceRequest struct {
	RegionId        string
	Name            string
	DiskType        string
	DiskSize        int
	DeployType      int
	PaidType        int
	IoMax           int
	VpcId           string
	VSwitchId       string
	SecurityGroupId string
	Tags            map[string]string
}

type KafkaTopic struct {
	InstanceId   string
	Topic        string
	PartitionNum int
	Remark       string
	CompactTopic bool
}

type CreateKafkaTopicRequest struct {
	InstanceId   string
	Topic        string
	PartitionNum int
	Remark       string
	CompactTopic bool
}

type KafkaConsumerGroup struct {
	InstanceId string
	ConsumerId string
	Remark     string
}

type CreateKafkaConsumerGroupRequest struct {
	InstanceId string
	ConsumerId string
	Remark     string
}

type KafkaSaslUser struct {
	InstanceId string
	Username   string
	Mechanism  string
}

type CreateKafkaSaslUserRequest struct {
	InstanceId string
	Username   string
	Mechanism  string
}

type KafkaSaslAcl struct {
	InstanceId             string
	Username               string
	AclResourceType        string
	AclResourceName        string
	AclResourcePatternType string
	AclOperationType       string
}

type CreateKafkaSaslAclRequest struct {
	InstanceId             string
	Username               string
	AclResourceType        string
	AclResourceName        string
	AclResourcePatternType string
	AclOperationType       string
}

type KafkaAllowedIp struct {
	InstanceId    string
	AllowedType   string
	PortRange     string
	AllowedIpList []string
}

type CreateKafkaAllowedIpRequest struct {
	InstanceId    string
	AllowedType   string
	PortRange     string
	AllowedIpList []string
}
