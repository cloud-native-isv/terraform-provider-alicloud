package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// NewKafkaService creates a new KafkaService using cws-lib-go implementation

func NewKafkaService(client *connectivity.AliyunClient) (*KafkaService, error) {
	creds := &common.Credentials{
		AccessKey: client.AccessKey,
		SecretKey: client.SecretKey,
		RegionId:  client.RegionId,
	}
	kafkaApi, err := kafka.NewKafkaAPI(creds)
	if err != nil {
		return nil, WrapError(err)
	}
	return &KafkaService{client: client, kafkaApi: kafkaApi}, nil
}

// KafkaService provides Kafka instance management operations
type KafkaService struct {
	client   *connectivity.AliyunClient
	kafkaApi *kafka.KafkaAPI
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

func (s *KafkaService) DescribeAlikafkaInstance(instanceId string) (*alikafka.InstanceVO, error) {
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

func (s *KafkaService) DescribeAlikafkaInstanceByOrderId(orderId string, timeout int) (*alikafka.InstanceVO, error) {
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

func (s *KafkaService) DescribeAlikafkaConsumerGroup(id string) (*alikafka.ConsumerVO, error) {
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

func (s *KafkaService) DescribeAlikafkaTopicStatus(id string) (*alikafka.TopicStatus, error) {
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

func (s *KafkaService) DescribeAlikafkaTopic(id string) (*alikafka.TopicVO, error) {

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

func (s *KafkaService) DescribeAlikafkaSaslUser(id string) (*alikafka.SaslUserVO, error) {
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

func (s *KafkaService) DescribeAlikafkaSaslAcl(id string) (*alikafka.KafkaAclVO, error) {
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

func (s *KafkaService) WaitForAlikafkaInstanceUpdated(id string, topicQuota int, diskSize int, ioMax int,
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

func (s *KafkaService) WaitForAlikafkaInstance(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			if NotFoundError(err) {
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

func (s *KafkaService) WaitForAlikafkaConsumerGroup(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaConsumerGroup(id)
		if err != nil {
			if NotFoundError(err) {
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

func (s *KafkaService) KafkaTopicStatusRefreshFunc(id string) resource.StateRefreshFunc {
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

func (s *KafkaService) WaitForAlikafkaTopic(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaTopic(id)
		if err != nil {
			if NotFoundError(err) {
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

func (s *KafkaService) WaitForAlikafkaSaslUser(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	for {
		object, err := s.DescribeAlikafkaSaslUser(id)
		if err != nil {
			if NotFoundError(err) {
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

func (s *KafkaService) WaitForAlikafkaSaslAcl(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	parts, err := ParseResourceId(id, 6)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	for {
		object, err := s.DescribeAlikafkaSaslAcl(id)
		if err != nil {

			if NotFoundError(err) {
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

func (s *KafkaService) DescribeTags(resourceId string, resourceTags map[string]interface{}, resourceType TagResourceType) (tags []alikafka.TagResource, err error) {
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

func (s *KafkaService) setInstanceTags(d *schema.ResourceData, resourceType TagResourceType) error {
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

func (s *KafkaService) tagsToMap(tags []alikafka.TagResource) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		if !s.ignoreTag(t) {
			result[t.TagKey] = t.TagValue
		}
	}
	return result
}

func (s *KafkaService) ignoreTag(t alikafka.TagResource) bool {
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

func (s *KafkaService) tagVOTagsToMap(tags []alikafka.TagVO) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		if !s.tagVOIgnoreTag(t) {
			result[t.Key] = t.Value
		}
	}
	return result
}

func (s *KafkaService) tagVOIgnoreTag(t alikafka.TagVO) bool {
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

func (s *KafkaService) diffTags(oldTags, newTags []alikafka.TagResourcesTag) ([]alikafka.TagResourcesTag, []alikafka.TagResourcesTag) {
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

func (s *KafkaService) tagsFromMap(m map[string]interface{}) []alikafka.TagResourcesTag {
	result := make([]alikafka.TagResourcesTag, 0, len(m))
	for k, v := range m {
		result = append(result, alikafka.TagResourcesTag{
			Key:   k,
			Value: v.(string),
		})
	}

	return result
}

func (s *KafkaService) GetAllowedIpList(id string) (object map[string]interface{}, err error) {
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

func (s *KafkaService) SetResourceTags(d *schema.ResourceData, resourceType string) error {
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
			// 由于 CWS-Lib-Go 中已经通过 KafkaService.setInstanceTags 处理标签，这里的
			// 旧版基于 RPC 的标签创建逻辑保持为空，实现向后兼容但不实际调用。
		}
		d.SetPartial("tags")
	}
	return nil
}

// CreatePostPayOrder creates a post-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) CreatePostPayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePostPaid
	return s.kafkaApi.CreateOrder(order)
}

// CreatePrePayOrder creates a pre-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) CreatePrePayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePrePaid
	return s.kafkaApi.CreateOrder(order)
}

// StartInstance 启动Kafka实例
func (s *KafkaService) StartInstance(request *StartInstanceRequest) error {
	params := make(map[string]interface{})
	if request.ZoneId != "" {
		params["ZoneId"] = request.ZoneId
	}
	if request.DeployModule != "" {
		params["DeployModule"] = request.DeployModule
	}
	if request.IsEipInner {
		params["IsEipInner"] = request.IsEipInner
	}
	if request.IsSetUserAndPassword {
		params["IsSetUserAndPassword"] = request.IsSetUserAndPassword
	}
	if request.Username != "" {
		params["Username"] = request.Username
	}
	if request.Password != "" {
		params["Password"] = request.Password
	}
	if request.Name != "" {
		params["Name"] = request.Name
	}
	if request.CrossZone {
		params["CrossZone"] = request.CrossZone
	}
	if request.SecurityGroup != "" {
		params["SecurityGroup"] = request.SecurityGroup
	}
	if request.ServiceVersion != "" {
		params["ServiceVersion"] = request.ServiceVersion
	}
	if request.Config != "" {
		params["Config"] = request.Config
	}
	if request.KMSKeyId != "" {
		params["KMSKeyId"] = request.KMSKeyId
	}
	if request.Notifier != "" {
		params["Notifier"] = request.Notifier
	}
	if request.UserPhoneNum != "" {
		params["UserPhoneNum"] = request.UserPhoneNum
	}
	if request.SelectedZones != "" {
		params["SelectedZones"] = request.SelectedZones
	}
	if request.IsForceSelectedZones {
		params["IsForceSelectedZones"] = request.IsForceSelectedZones
	}
	if len(request.VSwitchIds) > 0 {
		params["VSwitchIds"] = request.VSwitchIds
	}

	return s.kafkaApi.StartInstance(request.RegionId, request.InstanceId, request.VSwitchId, request.VpcId, params)
}

// StopInstance stops a Kafka instance
func (s *KafkaService) StopInstance(request *StopInstanceRequest) error {
	return s.kafkaApi.StopInstance(request.RegionId, request.InstanceId)
}

// ModifyInstanceName 修改Kafka实例名称
func (s *KafkaService) ModifyInstanceName(request *ModifyInstanceNameRequest) error {
	return s.kafkaApi.ModifyInstanceName(request.RegionId, request.InstanceId, request.InstanceName)
}

// UpgradeInstanceVersion 升级Kafka实例版本
func (s *KafkaService) UpgradeInstanceVersion(request *UpgradeInstanceVersionRequest) error {
	return s.kafkaApi.UpgradeInstanceVersion(request.RegionId, request.InstanceId, request.TargetVersion)
}

// UpgradePostPayOrder upgrades a post-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) UpgradePostPayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePostPaid
	return s.kafkaApi.UpgradeOrder(order)
}

// UpgradePrePayOrder upgrades a pre-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) UpgradePrePayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePrePaid
	return s.kafkaApi.UpgradeOrder(order)
}

func (s *KafkaService) AliKafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		status := fmt.Sprintf("%d", object.ServiceStatus)

		for _, failState := range failStates {
			if status == failState {
				return object, status, fmt.Errorf("resource in failed state: %s", status)
			}
		}

		return object, status, nil
	}
}
