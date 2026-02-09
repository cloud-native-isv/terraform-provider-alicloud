package alicloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// ListAlikafkaConsumerGroups lists Kafka consumer groups using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaConsumerGroups(instanceId string) ([]*kafka.ConsumerGroup, error) {
	groups, err := s.kafkaApi.ListConsumerGroups(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return groups, nil
}

// CreateAlikafkaConsumerGroup creates a consumer group using CWS-Lib-Go
func (s *KafkaService) CreateAlikafkaConsumerGroup(instanceId, consumerGroupId, description string, tags map[string]string) (*kafka.ConsumerGroup, error) {
	group, err := s.kafkaApi.CreateConsumerGroup(instanceId, s.client.RegionId, consumerGroupId, description, tags)
	if err != nil {
		return nil, WrapError(err)
	}
	return group, nil
}

// DeleteAlikafkaConsumerGroup deletes a consumer group using CWS-Lib-Go
func (s *KafkaService) DeleteAlikafkaConsumerGroup(instanceId, consumerGroupId string) error {
	if err := s.kafkaApi.DeleteConsumerGroup(instanceId, s.client.RegionId, consumerGroupId); err != nil {
		return WrapError(err)
	}
	return nil
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
	if err := s.retryWithCommonErrors(10*time.Minute, func() error {
		var err error
		object, err = s.kafkaApi.GetConsumerGroup(instanceId, consumerId)
		return err
	}); err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, consumerId, "GetConsumerGroup", AlibabaCloudSdkGoERROR)
	}
	return object, nil
}
