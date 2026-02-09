package alicloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

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


// DescribeConsumerGroup retrieves a Kafka consumer group using CWS-Lib-Go
func (s *KafkaService) DescribeConsumerGroup(instanceId, consumerId string) (*kafka.ConsumerGroup, error) {
	var object *kafka.ConsumerGroup
	var err error

	wait := incrementalWait(2*time.Second, 1*time.Second)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		object, err = s.kafkaApi.GetConsumerGroup(instanceId, consumerId)
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, consumerId, "GetConsumerGroup", AlibabaCloudSdkGoERROR)
	}
	return object, nil
}

