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
)

// NewKafkaService creates a new KafkaService using cws-lib-go implementation
func NewKafkaService(client *connectivity.AliyunClient) (*KafkaService, error) {
	return &KafkaService{
		client: client,
	}, nil
}

// KafkaService provides Kafka instance management operations
type KafkaService struct {
	client *connectivity.AliyunClient
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

func (s *KafkaService) WaitForAlikafkaConsumerGroup(id string, status Status, timeout int) error {
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

// CreatePostPayOrder creates a post-paid Kafka instance order
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) CreatePostPayOrder(request map[string]interface{}) (map[string]interface{}, error) {
	action := "CreatePostPayOrder"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// CreatePrePayOrderRequest 创建预付费订单的请求结构
type CreatePrePayOrderRequest struct {
	RegionId        string
	PartitionNum    int32
	DiskType        int32
	DiskSize        int32
	DeployType      int32
	IoMax           int32
	IoMaxSpec       string
	SpecType        string
	EipMax          int32
	ResourceGroupId string
}

// CreatePrePayOrder creates a pre-paid Kafka instance order
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) CreatePrePayOrder(request map[string]interface{}) (map[string]interface{}, error) {
	action := "CreatePrePayOrder"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// StartInstanceRequest 启动Kafka实例的请求结构
type StartInstanceRequest struct {
	RegionId       string
	InstanceId     string
	VSwitchId      string
	VpcId          string
	ZoneId         string
	VSwitchIds     []string
	DeployModule   string
	IsEipInner     bool
	Name           string
	SecurityGroup  string
	ServiceVersion string
	Config         string
	KMSKeyId       string
	SelectedZones  string
	CrossZone      bool
}

// StartInstance 启动Kafka实例
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) StartInstance(request map[string]interface{}) (map[string]interface{}, error) {
	action := "StartInstance"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// ModifyInstanceNameRequest 修改实例名称的请求结构
type ModifyInstanceNameRequest struct {
	RegionId     string
	InstanceId   string
	InstanceName string
}

// ModifyInstanceName 修改Kafka实例名称
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) ModifyInstanceName(request map[string]interface{}) (map[string]interface{}, error) {
	action := "ModifyInstanceName"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// ConvertPostPayOrderRequest 转换后付费订单的请求结构
type ConvertPostPayOrderRequest struct {
	RegionId   string
	InstanceId string
}

// ConvertPostPayOrder 转换后付费订单为预付费
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) ConvertPostPayOrder(request map[string]interface{}) (map[string]interface{}, error) {
	action := "ConvertPostPayOrder"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// UpgradePostPayOrderRequest 升级后付费实例的请求结构
type UpgradePostPayOrderRequest struct {
	InstanceId   string
	RegionId     string
	PartitionNum int32
	DiskSize     int32
	IoMax        int32
	IoMaxSpec    string
	SpecType     string
	EipMax       int32
	EipModel     bool
}

// UpgradePostPayOrder 升级后付费Kafka实例
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) UpgradePostPayOrder(request map[string]interface{}) (map[string]interface{}, error) {
	action := "UpgradePostPayOrder"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// UpgradePrePayOrder 升级预付费Kafka实例
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) UpgradePrePayOrder(request map[string]interface{}) (map[string]interface{}, error) {
	action := "UpgradePrePayOrder"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// UpgradeInstanceVersionRequest 升级实例版本的请求结构
type UpgradeInstanceVersionRequest struct {
	InstanceId    string
	RegionId      string
	TargetVersion string
}

// UpgradeInstanceVersion 升级Kafka实例版本
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) UpgradeInstanceVersion(request map[string]interface{}) (map[string]interface{}, error) {
	action := "UpgradeInstanceVersion"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// UpdateInstanceConfigRequest 更新实例配置的请求结构
type UpdateInstanceConfigRequest struct {
	RegionId   string
	InstanceId string
	Config     string
}

// UpdateInstanceConfig 更新Kafka实例配置
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) UpdateInstanceConfig(request map[string]interface{}) (map[string]interface{}, error) {
	action := "UpdateInstanceConfig"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// ChangeResourceGroupRequest 变更资源组的请求结构
type ChangeResourceGroupRequest struct {
	RegionId           string
	ResourceId         string
	NewResourceGroupId string
}

// ChangeResourceGroup 变更Kafka实例资源组
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) ChangeResourceGroup(request map[string]interface{}) (map[string]interface{}, error) {
	action := "ChangeResourceGroup"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// EnableAutoGroupCreationRequest 启用自动创建消费者组的请求结构
type EnableAutoGroupCreationRequest struct {
	RegionId   string
	InstanceId string
	Enable     bool
}

// EnableAutoGroupCreation 启用/禁用Kafka实例自动创建消费者组
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) EnableAutoGroupCreation(request map[string]interface{}) (map[string]interface{}, error) {
	action := "EnableAutoGroupCreation"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// EnableAutoTopicCreationRequest 启用自动创建主题的请求结构
type EnableAutoTopicCreationRequest struct {
	RegionId        string
	InstanceId      string
	Operate         string
	PartitionNum    int32
	UpdatePartition bool
}

// EnableAutoTopicCreation 启用/禁用Kafka实例自动创建主题
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) EnableAutoTopicCreation(request map[string]interface{}) (map[string]interface{}, error) {
	action := "EnableAutoTopicCreation"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// ReleaseInstanceRequest 释放实例的请求结构
type ReleaseInstanceRequest struct {
	InstanceId          string
	RegionId            string
	ForceDeleteInstance bool
}

// ReleaseInstance 释放Kafka实例（仅适用于后付费实例）
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) ReleaseInstance(request map[string]interface{}) (map[string]interface{}, error) {
	action := "ReleaseInstance"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// DeleteInstanceRequest 删除实例的请求结构
type DeleteInstanceRequest struct {
	InstanceId string
	RegionId   string
}

// DeleteInstance 删除Kafka实例
// Note: This operation is not直接支持 by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) DeleteInstance(request map[string]interface{}) (map[string]interface{}, error) {
	action := "DeleteInstance"
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, false)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "kafka_instance", action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

// GetQuotaTip gets the Kafka instance quota information
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) GetQuotaTip(instanceId string) (map[string]interface{}, error) {
	action := "GetQuotaTip"
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"InstanceId": instanceId,
	}
	response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
	if err != nil {
		return nil, WrapError(err)
	}

	v, err := jsonpath.Get("$.QuotaData", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, instanceId, "$.QuotaData", response)
	}
	return v.(map[string]interface{}), nil
}

// DescribeKafkaInstanceByOrderId describes a Kafka instance by order ID
// Note: This operation is not directly supported by cws-lib-go yet,
// so we fall back to the underlying SDK while maintaining service layer abstraction
func (s *KafkaService) DescribeKafkaInstanceByOrderId(orderId string, timeout int) (map[string]interface{}, error) {
	action := "GetInstanceList"
	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"OrderId":  orderId,
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		response, err := s.client.RpcPost("alikafka", "2019-09-16", action, nil, request, true)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, orderId, action, AlibabaCloudSdkGoERROR)
		}

		v, err := jsonpath.Get("$.InstanceList.InstanceVO", response)
		if err != nil {
			return nil, WrapErrorf(err, FailedGetAttributeMsg, orderId, "$.InstanceList.InstanceVO", response)
		}

		for _, v := range v.([]interface{}) {
			return v.(map[string]interface{}), nil
		}

		if time.Now().After(deadline) {
			return nil, WrapErrorf(NotFoundErr("AlikafkaInstance", orderId), NotFoundMsg, ProviderERROR)
		}

		time.Sleep(5 * time.Second)
	}
}

// DescribeAliKafkaInstance gets Kafka instance details as map[string]interface{}
// This method wraps the existing DescribeAlikafkaInstance method and converts the result
func (s *KafkaService) DescribeAliKafkaInstance(id string) (object map[string]interface{}, err error) {
	instance, err := s.DescribeAlikafkaInstance(id)
	if err != nil {
		return nil, err
	}

	// Convert the InstanceVO to map[string]interface{}
	object = make(map[string]interface{})
	object["Name"] = instance.Name
	object["DiskType"] = instance.DiskType
	object["DiskSize"] = instance.DiskSize
	object["DeployType"] = instance.DeployType
	object["IoMax"] = instance.IoMax
	object["IoMaxSpec"] = instance.IoMaxSpec
	object["EipMax"] = instance.EipMax
	object["ResourceGroupId"] = instance.ResourceGroupId
	object["VpcId"] = instance.VpcId
	object["VSwitchId"] = instance.VSwitchId
	object["ZoneId"] = instance.ZoneId
	object["PaidType"] = instance.PaidType
	object["SpecType"] = instance.SpecType
	object["SecurityGroup"] = instance.SecurityGroup
	object["EndPoint"] = instance.EndPoint
	object["SslEndPoint"] = instance.SslEndPoint
	object["DomainEndpoint"] = instance.DomainEndpoint
	object["SslDomainEndpoint"] = instance.SslDomainEndpoint
	object["SaslDomainEndpoint"] = instance.SaslDomainEndpoint
	object["ServiceStatus"] = instance.ServiceStatus
	object["UpgradeServiceDetailInfo"] = instance.UpgradeServiceDetailInfo
	object["AllConfig"] = instance.AllConfig
	object["KmsKeyId"] = instance.KmsKeyId
	// Note: Some fields may not exist in the InstanceVO struct, so we skip them for now
	// object["AutoCreateGroupEnable"] = instance.AutoCreateGroupEnable
	// object["AutoCreateTopicEnable"] = instance.AutoCreateTopicEnable
	// object["DefaultPartitionNum"] = instance.DefaultPartitionNum
	// object["VSwitchIds"] = instance.VSwitchIds
	// object["CrossZone"] = instance.CrossZone

	return object, nil
}

// AliKafkaInstanceStateRefreshFunc Kafka instance state refresh function
func (s *KafkaService) AliKafkaInstanceStateRefreshFunc(id, attribute string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaInstance(id)
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
