package alicloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// CreateAlikafkaTopic creates a Kafka topic using CWS-Lib-Go
func (s *KafkaService) CreateAlikafkaTopic(topic *kafka.KafkaTopic) error {
	if err := s.retryWithCommonErrors(10*time.Minute, func() error {
		return s.kafkaApi.CreateTopic(topic)
	}); err != nil {
		return WrapError(err)
	}
	return nil
}

// DeleteAlikafkaTopic deletes a Kafka topic using CWS-Lib-Go
func (s *KafkaService) DeleteAlikafkaTopic(instanceId, topicName string) error {
	if err := s.retryWithCommonErrors(10*time.Minute, func() error {
		return s.kafkaApi.DeleteTopic(instanceId, topicName)
	}); err != nil {
		return WrapError(err)
	}
	return nil
}

// ListAlikafkaTopics lists Kafka topics using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaTopics(instanceId string) ([]*kafka.KafkaTopic, error) {
	topics, err := s.kafkaApi.ListTopics(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return topics, nil
}

// ModifyAlikafkaTopicRemark updates a topic remark using CWS-Lib-Go
func (s *KafkaService) ModifyAlikafkaTopicRemark(instanceId, topicName, remark string) error {
	if err := s.kafkaApi.ModifyTopicRemark(instanceId, topicName, remark); err != nil {
		return WrapError(err)
	}
	return nil
}

// ModifyAlikafkaTopicPartitions updates topic partition count using CWS-Lib-Go
func (s *KafkaService) ModifyAlikafkaTopicPartitions(instanceId, topicName string, addPartitionNum int32) error {
	if err := s.retryWithCommonErrors(10*time.Minute, func() error {
		return s.kafkaApi.ModifyPartitionNum(instanceId, topicName, s.client.RegionId, addPartitionNum)
	}); err != nil {
		return WrapError(err)
	}
	return nil
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
	instanceId, topicName, err := DecodeTopicId(id)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeAlikafkaTopic(instanceId, topicName)
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

// DescribeAlikafkaTopic retrieves a Kafka topic using CWS-Lib-Go
func (s *KafkaService) DescribeAlikafkaTopic(instanceId, topicName string) (*kafka.KafkaTopic, error) {
	var object *kafka.KafkaTopic
	if err := s.retryWithCommonErrors(5*time.Minute, func() error {
		var err error
		object, err = s.kafkaApi.GetTopic(instanceId, topicName)
		return err
	}); err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, topicName, "GetTopic", AlibabaCloudSdkGoERROR)
	}

	return object, nil
}

// DescribeTopicStatus retrieves a Kafka topic status using CWS-Lib-Go
func (s *KafkaService) DescribeTopicStatus(instanceId, topicName string) (*kafka.TopicStatus, error) {
	var object *kafka.TopicStatus
	if err := s.retryWithCommonErrors(5*time.Minute, func() error {
		var err error
		object, err = s.kafkaApi.GetTopicStatus(instanceId, topicName)
		return err
	}); err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, topicName, "GetTopicStatus", AlibabaCloudSdkGoERROR)
	}
	return object, nil
}
