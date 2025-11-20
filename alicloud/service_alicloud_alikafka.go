package alicloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunKafkaAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// NewKafkaService creates a new KafkaService using cws-lib-go implementation
func NewKafkaService(client *connectivity.AliyunClient) (*KafkaService, error) {
	// Create the Kafka config
	config := &aliyunKafkaAPI.KafkaConfig{
		AccessKeyId:     client.AccessKey,
		AccessKeySecret: client.SecretKey,
		RegionId:        client.RegionId,
	}

	// Create the cws-lib-go KafkaAPI
	kafkaAPI, err := aliyunKafkaAPI.NewKafkaAPI(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go KafkaAPI: %w", err)
	}

	return &KafkaService{
		client:   client,
		kafkaAPI: kafkaAPI,
	}, nil
}

// GetAPI returns the KafkaAPI instance for direct API access
func (service *KafkaService) GetAPI() *aliyunKafkaAPI.KafkaAPI {
	return service.kafkaAPI
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

type AlikafkaService struct {
	client *connectivity.AliyunClient
}

func (s *AlikafkaService) DescribeAlikafkaInstance(instanceId string) (*alikafka.InstanceVO, error) {
	alikafkaInstance := &alikafka.InstanceVO{}
	instanceListReq := alikafka.CreateGetInstanceListRequest()
	instanceListReq.RegionId = s.client.RegionId

	wait := incrementalWait(2*time.Second, 1*time.Second)
	var raw interface{}
	var err error
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(client *alikafka.Client) (interface{}, error) {
			return client.GetInstanceList(instanceListReq)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(instanceListReq.GetActionName(), raw, instanceListReq.RpcRequest, instanceListReq)
		return nil
	})

	if err != nil {
		return alikafkaInstance, WrapErrorf(err, DefaultErrorMsg, instanceId, instanceListReq.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	instanceListResp, _ := raw.(*alikafka.GetInstanceListResponse)
	addDebug(instanceListReq.GetActionName(), raw, instanceListReq.RpcRequest, instanceListReq)

	for _, v := range instanceListResp.InstanceList.InstanceVO {

		// ServiceStatus equals 10 means the instance is released, do not return the instance.
		if v.InstanceId == instanceId && v.ServiceStatus != 10 {
			return &v, nil
		}
	}
	return alikafkaInstance, WrapErrorf(NotFoundErr("AlikafkaInstance", instanceId), NotFoundMsg, ProviderERROR)
}

func (s *AlikafkaService) DescribeAlikafkaInstanceByOrderId(orderId string, timeout int) (*alikafka.InstanceVO, error) {
	alikafkaInstance := &alikafka.InstanceVO{}
	instanceListReq := alikafka.CreateGetInstanceListRequest()
	instanceListReq.RegionId = s.client.RegionId
	instanceListReq.OrderId = orderId

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {

		wait := incrementalWait(2*time.Second, 1*time.Second)
		var raw interface{}
		var err error
		err = resource.Retry(10*time.Minute, func() *resource.RetryError {
			raw, err = s.client.WithAlikafkaClient(func(client *alikafka.Client) (interface{}, error) {
				return client.GetInstanceList(instanceListReq)
			})
			if err != nil {
				if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(instanceListReq.GetActionName(), raw, instanceListReq.RpcRequest, instanceListReq)
			return nil
		})

		if err != nil {
			return alikafkaInstance, WrapErrorf(err, DefaultErrorMsg, orderId, instanceListReq.GetActionName(), AlibabaCloudSdkGoERROR)
		}

		instanceListResp, _ := raw.(*alikafka.GetInstanceListResponse)
		addDebug(instanceListReq.GetActionName(), raw, instanceListReq.RpcRequest, instanceListReq)
		for _, v := range instanceListResp.InstanceList.InstanceVO {
			return &v, nil
		}
		if time.Now().After(deadline) {
			return alikafkaInstance, WrapErrorf(NotFoundErr("AlikafkaInstance", orderId), NotFoundMsg, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) DescribeAlikafkaConsumerGroup(id string) (*alikafka.ConsumerVO, error) {
	alikafkaConsumerGroup := &alikafka.ConsumerVO{}

	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return alikafkaConsumerGroup, WrapError(err)
	}
	instanceId := parts[0]
	consumerId := parts[1]

	request := alikafka.CreateGetConsumerListRequest()
	request.InstanceId = instanceId
	request.RegionId = s.client.RegionId

	wait := incrementalWait(2*time.Second, 1*time.Second)
	var raw interface{}
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(client *alikafka.Client) (interface{}, error) {
			return client.GetConsumerList(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		return nil
	})

	if err != nil {
		return alikafkaConsumerGroup, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	consumerListResp, _ := raw.(*alikafka.GetConsumerListResponse)
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	for _, v := range consumerListResp.ConsumerList.ConsumerVO {
		if v.ConsumerId == consumerId {
			return &v, nil
		}
	}
	return alikafkaConsumerGroup, WrapErrorf(NotFoundErr("AlikafkaConsumerGroup", id), NotFoundMsg, ProviderERROR)
}

func (s *AlikafkaService) DescribeAlikafkaTopicStatus(id string) (*alikafka.TopicStatus, error) {
	alikafkaTopicStatus := &alikafka.TopicStatus{}
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return alikafkaTopicStatus, WrapError(err)
	}
	instanceId := parts[0]
	topic := parts[1]

	request := alikafka.CreateGetTopicStatusRequest()
	request.InstanceId = instanceId
	request.RegionId = s.client.RegionId
	request.Topic = topic

	wait := incrementalWait(3*time.Second, 5*time.Second)
	var raw interface{}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(alikafkaClient *alikafka.Client) (interface{}, error) {
			return alikafkaClient.GetTopicStatus(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)

		return nil
	})

	if err != nil {
		return alikafkaTopicStatus, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	topicStatusResp, _ := raw.(*alikafka.GetTopicStatusResponse)

	if topicStatusResp.TopicStatus.OffsetTable.OffsetTableItem != nil {
		return &topicStatusResp.TopicStatus, nil
	}

	return alikafkaTopicStatus, WrapErrorf(NotFoundErr("AlikafkaTopicStatus "+ResourceNotfound, id), ResourceNotfound)
}

func (s *AlikafkaService) DescribeAlikafkaTopic(id string) (*alikafka.TopicVO, error) {

	alikafkaTopic := &alikafka.TopicVO{}
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return alikafkaTopic, WrapError(err)
	}
	instanceId := parts[0]
	topic := parts[1]

	request := alikafka.CreateGetTopicListRequest()
	request.InstanceId = instanceId
	request.RegionId = s.client.RegionId

	wait := incrementalWait(3*time.Second, 5*time.Second)
	var raw interface{}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(alikafkaClient *alikafka.Client) (interface{}, error) {
			return alikafkaClient.GetTopicList(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		return nil
	})

	if err != nil {
		return alikafkaTopic, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	topicListResp, _ := raw.(*alikafka.GetTopicListResponse)

	for _, v := range topicListResp.TopicList.TopicVO {
		if v.Topic == topic {
			return &v, nil
		}
	}
	return alikafkaTopic, WrapErrorf(NotFoundErr("AlikafkaTopic", id), NotFoundMsg, ProviderERROR)
}

func (s *AlikafkaService) DescribeAlikafkaSaslUser(id string) (*alikafka.SaslUserVO, error) {
	alikafkaSaslUser := &alikafka.SaslUserVO{}

	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return alikafkaSaslUser, WrapError(err)
	}
	instanceId := parts[0]
	username := parts[1]

	request := alikafka.CreateDescribeSaslUsersRequest()
	request.InstanceId = instanceId
	request.RegionId = s.client.RegionId

	wait := incrementalWait(3*time.Second, 5*time.Second)
	var raw interface{}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(alikafkaClient *alikafka.Client) (interface{}, error) {
			return alikafkaClient.DescribeSaslUsers(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		return nil
	})

	if err != nil {
		return alikafkaSaslUser, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	userListResp, _ := raw.(*alikafka.DescribeSaslUsersResponse)
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	for _, v := range userListResp.SaslUserList.SaslUserVO {
		if v.Username == username {
			return &v, nil
		}
	}
	return alikafkaSaslUser, WrapErrorf(NotFoundErr("AlikafkaSaslUser", id), NotFoundMsg, ProviderERROR)
}

func (s *AlikafkaService) DescribeAlikafkaSaslAcl(id string) (*alikafka.KafkaAclVO, error) {
	alikafkaSaslAcl := &alikafka.KafkaAclVO{}

	parts, err := ParseResourceId(id, 6)
	if err != nil {
		return alikafkaSaslAcl, WrapError(err)
	}
	instanceId := parts[0]
	username := parts[1]
	aclResourceType := parts[2]
	aclResourceName := parts[3]
	aclResourcePatternType := parts[4]
	aclOperationType := parts[5]

	request := alikafka.CreateDescribeAclsRequest()
	request.InstanceId = instanceId
	request.RegionId = s.client.RegionId
	request.Username = username
	request.AclResourceType = aclResourceType
	request.AclResourceName = aclResourceName

	var raw interface{}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(alikafkaClient *alikafka.Client) (interface{}, error) {
			return alikafkaClient.DescribeAcls(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{"BIZ_SUBSCRIPTION_NOT_FOUND", "BIZ_TOPIC_NOT_FOUND", "BIZ.INSTANCE.STATUS.ERROR"}) {
				return resource.NonRetryableError(WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR))
			}
			return resource.NonRetryableError(WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR))
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		return nil
	})

	if err != nil {
		return alikafkaSaslAcl, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}

	aclListResp, _ := raw.(*alikafka.DescribeAclsResponse)
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	for _, v := range aclListResp.KafkaAclList.KafkaAclVO {
		if v.Username == username && v.AclResourceType == aclResourceType && v.AclResourceName == aclResourceName && v.AclResourcePatternType == aclResourcePatternType && v.AclOperationType == aclOperationType {
			return &v, nil
		}
	}
	return alikafkaSaslAcl, WrapErrorf(NotFoundErr("AlikafkaSaslAcl", id), NotFoundMsg, ProviderERROR)
}

func (s *AlikafkaService) WaitForAlikafkaInstanceUpdated(id string, topicQuota int, diskSize int, ioMax int,
	eipMax int, paidType int, specType string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			return WrapError(err)
		}

		// Wait for all variables be equal.
		if object.InstanceId == id && object.TopicNumLimit == topicQuota && object.DiskSize == diskSize && object.IoMax == ioMax && object.EipMax == eipMax && object.PaidType == paidType && object.SpecType == specType {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object.InstanceId, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) WaitForAlikafkaInstance(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		// Process wait for running.
		if object.InstanceId == id && status == Running {

			// ServiceStatus equals 5, means the server is in service.
			if object.ServiceStatus == 5 {
				return nil
			}

		} else if object.InstanceId == id {

			// If target status is not deleted and found a instance, return.
			if status != Deleted {
				return nil
			} else {
				// ServiceStatus equals 10, means the server is in released.
				if object.ServiceStatus == 10 {
					return nil
				}
			}
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object.InstanceId, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) WaitForAlikafkaConsumerGroup(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaConsumerGroup(id)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if object.InstanceId+":"+object.ConsumerId == id && status != Deleted {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object.InstanceId+":"+object.ConsumerId, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) KafkaTopicStatusRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaTopicStatus(id)
		if err != nil {
			if !IsExpectedErrors(err, []string{ResourceNotfound}) {
				return nil, "", WrapError(err)
			}
		}

		if object.OffsetTable.OffsetTableItem != nil && len(object.OffsetTable.OffsetTableItem) > 0 {
			return object, "Running", WrapError(err)
		}

		return object, "Creating", nil
	}
}

func (s *AlikafkaService) WaitForAlikafkaTopic(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaTopic(id)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if object.InstanceId+":"+object.Topic == id && status != Deleted {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object.InstanceId+":"+object.Topic, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) WaitForAlikafkaSaslUser(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	for {
		object, err := s.DescribeAlikafkaSaslUser(id)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if instanceId+":"+object.Username == id && status != Deleted {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, instanceId+":"+object.Username, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) WaitForAlikafkaSaslAcl(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	parts, err := ParseResourceId(id, 6)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	for {
		object, err := s.DescribeAlikafkaSaslAcl(id)
		if err != nil {

			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if instanceId+":"+object.Username+":"+object.AclResourceType+":"+object.AclResourceName+":"+object.AclResourcePatternType+":"+object.AclOperationType == id && status != Deleted {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, instanceId+":"+object.Username, id, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *AlikafkaService) DescribeTags(resourceId string, resourceTags map[string]interface{}, resourceType TagResourceType) (tags []alikafka.TagResource, err error) {
	request := alikafka.CreateListTagResourcesRequest()
	request.RegionId = s.client.RegionId
	request.ResourceType = string(resourceType)
	request.ResourceId = &[]string{resourceId}
	if resourceTags != nil && len(resourceTags) > 0 {
		var reqTags []alikafka.ListTagResourcesTag
		for key, value := range resourceTags {
			reqTags = append(reqTags, alikafka.ListTagResourcesTag{
				Key:   key,
				Value: value.(string),
			})
		}
		request.Tag = &reqTags
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	var raw interface{}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err = s.client.WithAlikafkaClient(func(alikafkaClient *alikafka.Client) (interface{}, error) {
			return alikafkaClient.ListTagResources(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{Throttling, ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		return nil
	})
	if err != nil {
		err = WrapErrorf(err, DefaultErrorMsg, resourceId, request.GetActionName(), AlibabaCloudSdkGoERROR)
		return nil, err
	}
	response, _ := raw.(*alikafka.ListTagResourcesResponse)

	return response.TagResources.TagResource, nil
}

func (s *AlikafkaService) setInstanceTags(d *schema.ResourceData, resourceType TagResourceType) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := s.diffTags(s.tagsFromMap(o), s.tagsFromMap(n))

		if len(remove) > 0 {
			var tagKey []string
			for _, v := range remove {
				tagKey = append(tagKey, v.Key)
			}
			request := alikafka.CreateUntagResourcesRequest()
			request.ResourceId = &[]string{d.Id()}
			request.ResourceType = string(resourceType)
			request.TagKey = &tagKey
			request.RegionId = s.client.RegionId

			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				raw, err := s.client.WithAlikafkaClient(func(client *alikafka.Client) (interface{}, error) {
					return client.UntagResources(request)
				})
				if err != nil {
					if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(request.GetActionName(), raw, request.RpcRequest, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), request.GetActionName(), AlibabaCloudSdkGoERROR)
			}
		}

		if len(create) > 0 {
			request := alikafka.CreateTagResourcesRequest()
			request.ResourceId = &[]string{d.Id()}
			request.Tag = &create
			request.ResourceType = string(resourceType)
			request.RegionId = s.client.RegionId

			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				raw, err := s.client.WithAlikafkaClient(func(client *alikafka.Client) (interface{}, error) {
					return client.TagResources(request)
				})
				if err != nil {
					if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(request.GetActionName(), raw, request.RpcRequest, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), request.GetActionName(), AlibabaCloudSdkGoERROR)
			}
		}

		d.SetPartial("tags")
	}

	return nil
}

func (s *AlikafkaService) tagsToMap(tags []alikafka.TagResource) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		if !s.ignoreTag(t) {
			result[t.TagKey] = t.TagValue
		}
	}
	return result
}

func (s *AlikafkaService) ignoreTag(t alikafka.TagResource) bool {
	filter := []string{"^aliyun", "^acs:", "^http://", "^https://"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching prefix %v with %v\n", v, t.TagKey)
		ok, _ := regexp.MatchString(v, t.TagKey)
		if ok {
			log.Printf("[DEBUG] Found Alibaba Cloud specific t %s (val: %s), ignoring.\n", t.TagKey, t.TagValue)
			return true
		}
	}
	return false
}

func (s *AlikafkaService) tagVOTagsToMap(tags []alikafka.TagVO) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		if !s.tagVOIgnoreTag(t) {
			result[t.Key] = t.Value
		}
	}
	return result
}

func (s *AlikafkaService) tagVOIgnoreTag(t alikafka.TagVO) bool {
	filter := []string{"^aliyun", "^acs:", "^http://", "^https://"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching prefix %v with %v\n", v, t.Key)
		ok, _ := regexp.MatchString(v, t.Key)
		if ok {
			log.Printf("[DEBUG] Found Alibaba Cloud specific t %s (val: %s), ignoring.\n", t.Key, t.Value)
			return true
		}
	}
	return false
}

func (s *AlikafkaService) diffTags(oldTags, newTags []alikafka.TagResourcesTag) ([]alikafka.TagResourcesTag, []alikafka.TagResourcesTag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[t.Key] = t.Value
	}

	// Build the list of what to remove
	var remove []alikafka.TagResourcesTag
	for _, t := range oldTags {
		old, ok := create[t.Key]
		if !ok || old != t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return s.tagsFromMap(create), remove
}

func (s *AlikafkaService) tagsFromMap(m map[string]interface{}) []alikafka.TagResourcesTag {
	result := make([]alikafka.TagResourcesTag, 0, len(m))
	for k, v := range m {
		result = append(result, alikafka.TagResourcesTag{
			Key:   k,
			Value: v.(string),
		})
	}

	return result
}

func (s *AlikafkaService) GetAllowedIpList(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetAllowedIpList"
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": id,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *AlikafkaService) SetResourceTags(d *schema.ResourceData, resourceType string) error {
	if d.HasChange("tags") {
		added, removed := parsingTags(d)
		client := s.client

		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}
		if len(removedTagKeys) > 0 {
			action := "UntagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": resourceType,
				"ResourceId.1": d.Id(),
			}
			for i, key := range removedTagKeys {
				request[fmt.Sprintf("TagKey.%d", i+1)] = key
			}
			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) || IsExpectedErrors(err, []string{"ONS_SYSTEM_FLOW_CONTROL"}) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}
		if len(added) > 0 {
			action := "TagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": resourceType,
				"ResourceId.1": d.Id(),
			}
			count := 1
			for key, value := range added {
				request[fmt.Sprintf("Tag.%d.Key", count)] = key
				request[fmt.Sprintf("Tag.%d.Value", count)] = value
				count++
			}

			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) || IsExpectedErrors(err, []string{"ONS_SYSTEM_FLOW_CONTROL"}) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}
		d.SetPartial("tags")
	}
	return nil
}

func (s *AlikafkaService) DescribeAliKafkaConsumerGroup(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetConsumerList"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": parts[0],
	}
	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.ConsumerList.ConsumerVO", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ConsumerList.ConsumerVO", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("AliKafka", id), NotFoundWithResponse, response)
	}
	for _, v := range v.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["ConsumerId"]) == parts[1] {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}
	if !idExist {
		return object, WrapErrorf(NotFoundErr("AliKafka", id), NotFoundWithResponse, response)
	}
	return object, nil
}

func (s *AlikafkaService) DescribeAliKafkaSaslUser(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	action := "DescribeSaslUsers"
	client := s.client
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return object, WrapError(err)
	}

	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": parts[0],
	}

	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsExpectedErrors(err, []string{"BIZ_INSTANCE_STATUS_ERROR", "BIZ.INSTANCE.STATUS.ERROR"}) {
			return object, WrapErrorf(NotFoundErr("AliKafka:SaslUser", id), NotFoundWithResponse, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	resp, err := jsonpath.Get("$.SaslUserList.SaslUserVO", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.SaslUserList.SaslUserVO", response)
	}

	if v, ok := resp.([]interface{}); !ok || len(v) < 1 {
		return object, WrapErrorf(NotFoundErr("AliKafka:SaslUser", id), NotFoundWithResponse, response)
	}

	for _, v := range resp.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["Username"]) == parts[1] {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}

	if !idExist {
		return object, WrapErrorf(NotFoundErr("AliKafka:SaslUser", id), NotFoundWithResponse, response)
	}

	return object, nil
}

func (s *AlikafkaService) DescribeAliKafkaInstanceAllowedIpAttachment(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	action := "GetAllowedIpList"

	client := s.client

	parts, err := ParseResourceId(id, 4)
	if err != nil {
		err = WrapError(err)
		return
	}

	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": parts[0],
	}

	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsExpectedErrors(err, []string{"BIZ_INSTANCE_STATUS_ERROR", "BIZ.INSTANCE.STATUS.ERROR"}) {
			return object, WrapErrorf(NotFoundErr("AliKafka:InstanceAllowedIpAttachment", id), NotFoundWithResponse, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	var resp interface{}
	allowedType := parts[1]

	switch allowedType {
	case "vpc":
		resp, err = jsonpath.Get("$.AllowedList.VpcList", response)
		if err != nil {
			return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AllowedList.VpcList", response)
		}
	case "internet":
		resp, err = jsonpath.Get("$.AllowedList.InternetList", response)
		if err != nil {
			return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AllowedList.InternetList", response)
		}
	}

	if v, ok := resp.([]interface{}); !ok || len(v) < 1 {
		return object, WrapErrorf(NotFoundErr("AliKafka:InstanceAllowedIpAttachment", id), NotFoundWithResponse, response)
	}

	for _, v := range resp.([]interface{}) {
		ipList := v.(map[string]interface{})
		if fmt.Sprint(ipList["PortRange"]) == parts[2] {
			for _, ip := range ipList["AllowedIpList"].([]interface{}) {
				if fmt.Sprint(ip) == parts[3] {
					idExist = true
					return v.(map[string]interface{}), nil
				}
			}
		}
	}

	if !idExist {
		return object, WrapErrorf(NotFoundErr("AliKafka:InstanceAllowedIpAttachment", id), NotFoundWithResponse, response)
	}

	return object, nil
}

func (s *AlikafkaService) DescribeAliKafkaInstance(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	action := "GetInstanceList"

	client := s.client

	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": []string{id},
	}

	idExist := false
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	resp, err := jsonpath.Get("$.InstanceList.InstanceVO", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.InstanceList.InstanceVO", response)
	}

	if v, ok := resp.([]interface{}); !ok || len(v) < 1 {
		return object, WrapErrorf(NotFoundErr("AliKafka:Instance", id), NotFoundWithResponse, response)
	}

	for _, v := range resp.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["InstanceId"]) == id && fmt.Sprint(v.(map[string]interface{})["ServiceStatus"]) != "10" {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}

	if !idExist {
		return object, WrapErrorf(NotFoundErr("AliKafka:Instance", id), NotFoundWithResponse, response)
	}

	return object, nil
}

func (s *AlikafkaService) GetQuotaTip(instanceId string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetQuotaTip"
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": instanceId,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapError(err)
	}
	v, err := jsonpath.Get("$.QuotaData", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, instanceId, "$.QuotaData", response)
	}
	return v.(map[string]interface{}), nil
}

func (s *AlikafkaService) DescribeAliKafkaInstanceByOrderId(orderId string, timeout int) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetInstanceList"
	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"OrderId":  orderId,
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(10*time.Minute, func() *resource.RetryError {
			response, err = client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
			if err != nil {
				if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) || NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		addDebug(action, response, request)
		if err != nil {
			return object, WrapErrorf(err, DefaultErrorMsg, orderId, action, AlibabaCloudSdkGoERROR)
		}
		v, err := jsonpath.Get("$.InstanceList.InstanceVO", response)
		if err != nil {
			return object, WrapErrorf(err, FailedGetAttributeMsg, orderId, "$.InstanceList.InstanceVO", response)
		}
		for _, v := range v.([]interface{}) {
			return v.(map[string]interface{}), nil
		}
		if time.Now().After(deadline) {
			return object, WrapErrorf(NotFoundErr("AlikafkaInstance", orderId), NotFoundMsg, ProviderERROR)
		}
	}
}

func (s *AlikafkaService) AliKafkaInstanceStateRefreshFunc(id, attribute string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAliKafkaInstance(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {

			if fmt.Sprint(object[attribute]) == failState {
				return object, fmt.Sprint(object[attribute]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object[attribute])))
			}
		}
		return object, fmt.Sprint(object[attribute]), nil
	}
}

func (s *AlikafkaService) AliKafkaConsumerStateRefreshFunc(id, attribute string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAliKafkaConsumerGroup(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {

			if fmt.Sprint(object[attribute]) == failState {
				return object, fmt.Sprint(object[attribute]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object[attribute])))
			}
		}
		return object, fmt.Sprint(object[attribute]), nil
	}
}

func (s *AlikafkaService) AliKafkaTopicStateRefreshFunc(id, attribute string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaTopic(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		return object, "existing", nil
	}
}

// DescribeKafkaInstance 获取Kafka实例详细信息
func (s *KafkaService) DescribeKafkaInstance(id string) (*aliyunKafkaAPI.KafkaInstance, error) {
	kafkaInstance, err := s.GetAPI().GetInstance(context.TODO(), id)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, id)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	// ServiceStatus等于10表示实例已释放，不返回实例
	if kafkaInstance.Status == "10" {
		return nil, WrapErrorf(NotFoundErr("KafkaInstance", id), NotFoundMsg, ProviderERROR)
	}

	return kafkaInstance, nil
}

// CreateKafkaInstanceRequest 创建Kafka实例的请求结构
type CreateKafkaInstanceRequest struct {
	RegionId        string
	ZoneId          string
	Name            string
	DiskType        string
	DiskSize        int32
	IoMax           int32
	SpecType        string
	PaidType        int32
	VpcId           string
	VSwitchId       string
	SecurityGroupId string
}

// CreateKafkaInstance 创建Kafka实例
func (s *KafkaService) CreateKafkaInstance(request *CreateKafkaInstanceRequest) (*aliyunKafkaAPI.KafkaInstance, error) {
	kafkaInstance := &aliyunKafkaAPI.KafkaInstance{
		Name:     request.Name,
		RegionId: request.RegionId,
		ZoneId:   request.ZoneId,
		DiskType: request.DiskType,
		DiskSize: request.DiskSize,
		IoMax:    request.IoMax,
		SpecType: request.SpecType,
		Version:  "2.2.0", // 默认版本
	}

	result, err := s.GetAPI().CreateInstance(context.TODO(), kafkaInstance)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", "CreateKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

// DeleteKafkaInstance 删除Kafka实例
func (s *KafkaService) DeleteKafkaInstance(id string) error {
	err := s.GetAPI().DeleteInstance(context.TODO(), id)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "DeleteKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// KafkaInstanceStateRefreshFunc Kafka实例状态刷新函数
func (s *KafkaService) KafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeKafkaInstance(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.Status == failState {
				return object, object.Status, WrapError(Error(FailedToReachTargetStatus, object.Status))
			}
		}
		return object, object.Status, nil
	}
}

// WaitForKafkaInstanceCreating 等待Kafka实例创建完成
func (s *KafkaService) WaitForKafkaInstanceCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating"},
		[]string{"5"}, // 5表示运行中
		timeout,
		5*time.Second,
		s.KafkaInstanceStateRefreshFunc(id, []string{"10"}), // 10表示已删除
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// WaitForKafkaInstanceDeleting 等待Kafka实例删除完成
func (s *KafkaService) WaitForKafkaInstanceDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting"},
		[]string{}, // 空字符串表示已删除
		timeout,
		5*time.Second,
		s.KafkaInstanceStateRefreshFunc(id, []string{}),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// DescribeKafkaTopic 获取Kafka主题详细信息
func (s *KafkaService) DescribeKafkaTopic(id string) (*aliyunKafkaAPI.KafkaTopic, error) {
	instanceId, topicName, err := DecodeTopicId(id)
	if err != nil {
		return nil, WrapError(err)
	}

	kafkaTopic, err := s.GetAPI().GetTopic(context.TODO(), instanceId, topicName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, id)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeKafkaTopic", AlibabaCloudSdkGoERROR)
	}

	return kafkaTopic, nil
}

// CreateKafkaTopicRequest 创建Kafka主题的请求结构
type CreateKafkaTopicRequest struct {
	InstanceId   string
	Topic        string
	PartitionNum int32
	Remark       string
}

// CreateKafkaTopic 创建Kafka主题
func (s *KafkaService) CreateKafkaTopic(request *CreateKafkaTopicRequest) error {
	kafkaTopic := &aliyunKafkaAPI.KafkaTopic{
		InstanceId:   request.InstanceId,
		Topic:        request.Topic,
		PartitionNum: request.PartitionNum,
		Remark:       request.Remark,
		ReplicaNum:   3, // 默认副本数
	}

	err := s.GetAPI().CreateTopic(context.TODO(), kafkaTopic)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "kafka_topic", "CreateKafkaTopic", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteKafkaTopic 删除Kafka主题
func (s *KafkaService) DeleteKafkaTopic(id string) error {
	instanceId, topicName, err := DecodeTopicId(id)
	if err != nil {
		return WrapError(err)
	}

	err = s.GetAPI().DeleteTopic(context.TODO(), instanceId, topicName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "DeleteKafkaTopic", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// KafkaTopicStateRefreshFunc Kafka主题状态刷新函数
func (s *KafkaService) KafkaTopicStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeKafkaTopic(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For topics, we consider them "existing" if we can describe them
		return object, "existing", nil
	}
}

// WaitForKafkaTopicCreating 等待Kafka主题创建完成
func (s *KafkaService) WaitForKafkaTopicCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{},
		[]string{"existing"},
		timeout,
		5*time.Second,
		s.KafkaTopicStateRefreshFunc(id),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// WaitForKafkaTopicDeleting 等待Kafka主题删除完成
func (s *KafkaService) WaitForKafkaTopicDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"existing"},
		[]string{}, // 空字符串表示已删除
		timeout,
		5*time.Second,
		s.KafkaTopicStateRefreshFunc(id),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// DescribeKafkaConsumerGroup 获取Kafka消费者组详细信息
func (s *KafkaService) DescribeKafkaConsumerGroup(id string) (*aliyunKafkaAPI.ConsumerGroup, error) {
	instanceId, consumerId, err := DecodeConsumerGroupId(id)
	if err != nil {
		return nil, WrapError(err)
	}

	consumerGroup, err := s.GetAPI().GetConsumerGroup(context.TODO(), instanceId, consumerId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, id)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeKafkaConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	return consumerGroup, nil
}

// CreateKafkaConsumerGroupRequest 创建Kafka消费者组的请求结构
type CreateKafkaConsumerGroupRequest struct {
	InstanceId string
	ConsumerId string
	Remark     string
}

// CreateKafkaConsumerGroup 创建Kafka消费者组
func (s *KafkaService) CreateKafkaConsumerGroup(request *CreateKafkaConsumerGroupRequest) error {
	_, err := s.GetAPI().CreateConsumerGroup(context.TODO(), request.InstanceId, s.client.RegionId, request.ConsumerId, request.Remark, nil)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "kafka_consumer_group", "CreateKafkaConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteKafkaConsumerGroup 删除Kafka消费者组
func (s *KafkaService) DeleteKafkaConsumerGroup(id string) error {
	instanceId, consumerId, err := DecodeConsumerGroupId(id)
	if err != nil {
		return WrapError(err)
	}

	err = s.GetAPI().DeleteConsumerGroup(context.TODO(), instanceId, s.client.RegionId, consumerId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "DeleteKafkaConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// KafkaConsumerGroupStateRefreshFunc Kafka消费者组状态刷新函数
func (s *KafkaService) KafkaConsumerGroupStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeKafkaConsumerGroup(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For consumer groups, we consider them "existing" if we can describe them
		return object, "existing", nil
	}
}

// WaitForKafkaConsumerGroupCreating 等待Kafka消费者组创建完成
func (s *KafkaService) WaitForKafkaConsumerGroupCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{},
		[]string{"existing"},
		timeout,
		5*time.Second,
		s.KafkaConsumerGroupStateRefreshFunc(id),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// WaitForKafkaConsumerGroupDeleting 等待Kafka消费者组删除完成
func (s *KafkaService) WaitForKafkaConsumerGroupDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"existing"},
		[]string{}, // 空字符串表示已删除
		timeout,
		5*time.Second,
		s.KafkaConsumerGroupStateRefreshFunc(id),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// DescribeKafkaSaslUser 获取Kafka SASL用户详细信息
// 注意：当前 cws-lib-go 版本中未实现 SASL 用户管理功能，暂时返回未找到错误
func (s *KafkaService) DescribeKafkaSaslUser(id string) (*aliyunKafkaAPI.ConsumerGroup, error) {
	// 暂时返回未找到错误，因为 cws-lib-go 中未实现 SASL 用户管理功能
	return nil, WrapErrorf(Error(NotFoundMsg, ProviderERROR), "[ERROR] KafkaSaslUser not found: %s", id)
}

// CreateKafkaSaslUserRequest 创建Kafka SASL用户的请求结构
// 注意：当前 cws-lib-go 版本中未实现 SASL 用户管理功能
type CreateKafkaSaslUserRequest struct {
	InstanceId string
	Username   string
	Password   string
	Mechanism  string
}

// CreateKafkaSaslUser 创建Kafka SASL用户
// 注意：当前 cws-lib-go 版本中未实现 SASL 用户管理功能
func (s *KafkaService) CreateKafkaSaslUser(request *CreateKafkaSaslUserRequest) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 SASL 用户管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslUser management not implemented in current cws-lib-go version")
}

// DeleteKafkaSaslUser 删除Kafka SASL用户
// 注意：当前 cws-lib-go 版本中未实现 SASL 用户管理功能
func (s *KafkaService) DeleteKafkaSaslUser(id string) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 SASL 用户管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslUser management not implemented in current cws-lib-go version")
}

// DescribeKafkaSaslAcl 获取Kafka SASL ACL详细信息
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能，暂时返回未找到错误
func (s *KafkaService) DescribeKafkaSaslAcl(id string) (*aliyunKafkaAPI.ConsumerGroup, error) {
	// 暂时返回未找到错误，因为 cws-lib-go 中未实现 ACL 管理功能
	return nil, WrapErrorf(Error(NotFoundMsg, ProviderERROR), "[ERROR] KafkaSaslAcl not found: %s", id)
}

// CreateKafkaSaslAclRequest 创建Kafka SASL ACL的请求结构
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
type CreateKafkaSaslAclRequest struct {
	InstanceId             string
	Username               string
	AclResourceType        string
	AclResourceName        string
	AclResourcePatternType string
	AclOperationType       string
}

// CreateKafkaSaslAcl 创建Kafka SASL ACL
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
func (s *KafkaService) CreateKafkaSaslAcl(request *CreateKafkaSaslAclRequest) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 ACL 管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslAcl management not implemented in current cws-lib-go version")
}

// DeleteKafkaSaslAcl 删除Kafka SASL ACL
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
func (s *KafkaService) DeleteKafkaSaslAcl(id string) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 ACL 管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslAcl management not implemented in current cws-lib-go version")
}

// KafkaSaslAclStateRefreshFunc Kafka SASL ACL状态刷新函数
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
func (s *KafkaService) KafkaSaslAclStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// 暂时返回未找到错误，因为 cws-lib-go 中未实现 ACL 管理功能
		return nil, "", WrapErrorf(Error(NotFoundMsg, ProviderERROR), "[ERROR] KafkaSaslAcl not found: %s", id)
	}
}

// WaitForKafkaSaslAclCreating 等待Kafka SASL ACL创建完成
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
func (s *KafkaService) WaitForKafkaSaslAclCreating(id string, timeout time.Duration) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 ACL 管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslAcl management not implemented in current cws-lib-go version")
}

// WaitForKafkaSaslAclDeleting 等待Kafka SASL ACL删除完成
// 注意：当前 cws-lib-go 版本中未实现 ACL 管理功能
func (s *KafkaService) WaitForKafkaSaslAclDeleting(id string, timeout time.Duration) error {
	// 暂时不实现，因为 cws-lib-go 中未实现 ACL 管理功能
	return WrapErrorf(Error("NotImplemented"), "SaslAcl management not implemented in current cws-lib-go version")
}

// DescribeKafkaAllowedIp 获取Kafka 允许 IP 详细信息
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能，暂时返回未找到错误
func (s *KafkaService) DescribeKafkaAllowedIp(id string) (*aliyunKafkaAPI.ConsumerGroup, error) {
	// 暂时返回未找到错误，因为 cws-lib-go 中未实现允许 IP 管理功能
	return nil, WrapErrorf(Error(NotFoundMsg, ProviderERROR), "[ERROR] KafkaAllowedIp not found: %s", id)
}

// CreateKafkaAllowedIpRequest 创建Kafka 允许 IP 的请求结构
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
type CreateKafkaAllowedIpRequest struct {
	InstanceId  string
	AllowedType string
	PortRange   string
	IpAddress   string
}

// CreateKafkaAllowedIp 创建Kafka 允许 IP
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
func (s *KafkaService) CreateKafkaAllowedIp(request *CreateKafkaAllowedIpRequest) error {
	// 暂时不实现，因为 cws-lib-go 中未实现允许 IP 管理功能
	return WrapErrorf(Error("NotImplemented"), "AllowedIp management not implemented in current cws-lib-go version")
}

// DeleteKafkaAllowedIp 删除Kafka 允许 IP
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
func (s *KafkaService) DeleteKafkaAllowedIp(id string) error {
	// 暂时不实现，因为 cws-lib-go 中未实现允许 IP 管理功能
	return WrapErrorf(Error("NotImplemented"), "AllowedIp management not implemented in current cws-lib-go version")
}

// KafkaAllowedIpStateRefreshFunc Kafka 允许 IP 状态刷新函数
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
func (s *KafkaService) KafkaAllowedIpStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// 暂时返回未找到错误，因为 cws-lib-go 中未实现允许 IP 管理功能
		return nil, "", WrapErrorf(Error(NotFoundMsg, ProviderERROR), "[ERROR] KafkaAllowedIp not found: %s", id)
	}
}

// WaitForKafkaAllowedIpCreating 等待Kafka 允许 IP 创建完成
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
func (s *KafkaService) WaitForKafkaAllowedIpCreating(id string, timeout time.Duration) error {
	// 暂时不实现，因为 cws-lib-go 中未实现允许 IP 管理功能
	return WrapErrorf(Error("NotImplemented"), "AllowedIp management not implemented in current cws-lib-go version")
}

// WaitForKafkaAllowedIpDeleting 等待Kafka 允许 IP 删除完成
// 注意：当前 cws-lib-go 版本中未实现允许 IP 管理功能
func (s *KafkaService) WaitForKafkaAllowedIpDeleting(id string, timeout time.Duration) error {
	// 暂时不实现，因为 cws-lib-go 中未实现允许 IP 管理功能
	return WrapErrorf(Error("NotImplemented"), "AllowedIp management not implemented in current cws-lib-go version")
}

// KafkaService wraps the cws-lib-go KafkaAPI for Terraform provider usage
type KafkaService struct {
	client   *connectivity.AliyunClient
	kafkaAPI *aliyunKafkaAPI.KafkaAPI
}
