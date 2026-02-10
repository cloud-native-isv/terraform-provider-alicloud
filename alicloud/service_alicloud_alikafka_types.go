package alicloud

import (
	"fmt"
	"regexp"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

const (
	AliKafkaInstanceTypeReserved   = "Reserved"
	AliKafkaInstanceTypeServerless = "Serverless"
	AliKafkaBillingTypePrePaid     = "PrePaid"
	AliKafkaBillingTypePostPaid    = "PostPaid"
)

func ResolveAliKafkaInstanceType(value string) (kafka.KafkaInstanceSeries, error) {
	if value == "" {
		return kafka.InstanceSeriesReserved, nil
	}
	switch value {
	case AliKafkaInstanceTypeReserved:
		return kafka.InstanceSeriesReserved, nil
	case AliKafkaInstanceTypeServerless:
		return kafka.InstanceSeriesServerless, nil
	default:
		return "", fmt.Errorf("unsupported instance_type: %s", value)
	}
}

func ResolveAliKafkaBillingType(value string) (kafka.KafkaBillingType, error) {
	if value == "" {
		return kafka.BillingTypePostPay, nil
	}
	switch value {
	case AliKafkaBillingTypePostPaid:
		return kafka.BillingTypePostPay, nil
	case AliKafkaBillingTypePrePaid:
		return kafka.BillingTypePrePay, nil
	default:
		return "", fmt.Errorf("unsupported billing_type: %s", value)
	}
}

func FormatAliKafkaBillingType(paidType *kafka.KafkaPaidType) string {
	if paidType == nil {
		return ""
	}
	switch *paidType {
	case kafka.KafkaPaidTypePrePay:
		return AliKafkaBillingTypePrePaid
	case kafka.KafkaPaidTypePostPay:
		return AliKafkaBillingTypePostPaid
	default:
		return fmt.Sprintf("%d", *paidType)
	}
}

// EncodeInstanceId returns the instance ID as-is
func EncodeInstanceId(instanceId string) string {
	return instanceId
}

// DecodeInstanceId validates the instance ID format
func DecodeInstanceId(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid instance ID format: empty")
	}
	return id, nil
}

// EncodeDeploymentId returns the deployment ID as-is (instance ID)
func EncodeDeploymentId(instanceId string) string {
	return instanceId
}

// DecodeDeploymentId validates the deployment ID format
func DecodeDeploymentId(id string) (string, error) {
	return DecodeInstanceId(id)
}

// EncodeTopicId 将实例ID和主题名称编码为单一ID字符串
// 格式: instanceId:topic
func EncodeTopicId(instanceId, topic string) string {
	return fmt.Sprintf("%s:%s", instanceId, topic)
}

// DecodeTopicId 解析主题ID字符串为实例ID和主题名称组件
func DecodeTopicId(id string) (string, string, error) {
	parts := regexp.MustCompile(`^([^\:]+):(.+)$`).FindStringSubmatch(id)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid topic ID format, expected instanceId:topic, got %s", id)
	}
	return parts[1], parts[2], nil
}

// EncodeSaslUserId 将实例ID和用户名编码为单一ID字符串
// 格式: instanceId:username
func EncodeSaslUserId(instanceId, username string) string {
	return fmt.Sprintf("%s:%s", instanceId, username)
}

// DecodeSaslUserId 解析SASL用户ID字符串为实例ID和用户名组件
func DecodeSaslUserId(id string) (string, string, error) {
	parts := regexp.MustCompile(`^([^\:]+):(.+)$`).FindStringSubmatch(id)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid SASL user ID format, expected instanceId:username, got %s", id)
	}
	return parts[1], parts[2], nil
}

// EncodeSaslAclId 将SASL ACL的所有组件编码为单一ID字符串
// 格式: instanceId:username:aclResourceType:aclResourceName:aclResourcePatternType:aclOperationType
func EncodeSaslAclId(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType)
}

// DecodeSaslAclId 解析SASL ACL ID字符串为所有组件
func DecodeSaslAclId(id string) (string, string, string, string, string, string, error) {
	parts := regexp.MustCompile(`^([^\:]+):([^\:]+):([^\:]+):([^\:]+):([^\:]+):(.+)$`).FindStringSubmatch(id)
	if len(parts) != 7 {
		return "", "", "", "", "", "", fmt.Errorf("invalid SASL ACL ID format, expected instanceId:username:aclResourceType:aclResourceName:aclResourcePatternType:aclOperationType, got %s", id)
	}
	return parts[1], parts[2], parts[3], parts[4], parts[5], parts[6], nil
}

// EncodeConsumerGroupId 将实例ID和消费者组ID编码为单一ID字符串
// 格式: instanceId:consumerId
func EncodeConsumerGroupId(instanceId, consumerId string) string {
	return fmt.Sprintf("%s:%s", instanceId, consumerId)
}

// DecodeConsumerGroupId 解析消费者组ID字符串为实例ID和消费者组ID组件
func DecodeConsumerGroupId(id string) (string, string, error) {
	parts := regexp.MustCompile(`^([^\:]+):(.+)$`).FindStringSubmatch(id)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid consumer group ID format, expected instanceId:consumerId, got %s", id)
	}
	return parts[1], parts[2], nil
}

// EncodeAllowedIpId 将允许IP的所有组件编码为单一ID字符串
// 格式: instanceId:allowedType:portRange:ipAddress
func EncodeAllowedIpId(instanceId, allowedType, portRange, ipAddress string) string {
	return fmt.Sprintf("%s:%s:%s:%s", instanceId, allowedType, portRange, ipAddress)
}

// DecodeAllowedIpId 解析允许IP ID字符串为所有组件
func DecodeAllowedIpId(id string) (string, string, string, string, error) {
	parts := regexp.MustCompile(`^([^\:]+):([^\:]+):([^\:]+):(.+)$`).FindStringSubmatch(id)
	if len(parts) != 5 {
		return "", "", "", "", fmt.Errorf("invalid allowed IP ID format, expected instanceId:allowedType:portRange:ipAddress, got %s", id)
	}
	return parts[1], parts[2], parts[3], parts[4], nil
}

// StartInstanceRequest represents the request to start a Kafka instance
type StartInstanceRequest struct {
	InstanceId           string
	RegionId             string
	VpcId                string
	VSwitchId            string
	ZoneId               string
	DeployModule         string
	IsEipInner           bool
	IsSetUserAndPassword bool
	Username             string
	Password             string
	Name                 string
	CrossZone            bool
	SecurityGroup        string
	ServiceVersion       string
	Config               string
	KMSKeyId             string
	Notifier             string
	UserPhoneNum         string
	SelectedZones        string
	IsForceSelectedZones bool
	VSwitchIds           []string
}

// ModifyInstanceNameRequest represents the request to modify a Kafka instance name
type ModifyInstanceNameRequest struct {
	InstanceId   string
	RegionId     string
	InstanceName string
}

// UpgradeInstanceVersionRequest represents the request to upgrade a Kafka instance version
type UpgradeInstanceVersionRequest struct {
	InstanceId    string
	RegionId      string
	TargetVersion string
}

// StopInstanceRequest represents the request to stop a Kafka instance
type StopInstanceRequest struct {
	InstanceId string
	RegionId   string
}
