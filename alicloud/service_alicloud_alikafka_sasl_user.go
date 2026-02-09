package alicloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// CreateAlikafkaSaslUser creates a SASL user using CWS-Lib-Go
func (s *KafkaService) CreateAlikafkaSaslUser(instanceId, username, password string, options map[string]interface{}) error {
	if options == nil {
		options = map[string]interface{}{}
	}
	if err := s.kafkaApi.CreateSaslUser(context.Background(), s.client.RegionId, instanceId, username, password, options); err != nil {
		return WrapError(err)
	}
	return nil
}

// DeleteAlikafkaSaslUser deletes a SASL user using CWS-Lib-Go
func (s *KafkaService) DeleteAlikafkaSaslUser(instanceId, username string, options map[string]interface{}) error {
	if options == nil {
		options = map[string]interface{}{}
	}
	if err := s.kafkaApi.DeleteSaslUser(context.Background(), s.client.RegionId, instanceId, username, options); err != nil {
		return WrapError(err)
	}
	return nil
}

// ListAlikafkaSaslUsers lists SASL users using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaSaslUsers(instanceId string) ([]kafka.SaslUser, error) {
	resp, err := s.kafkaApi.DescribeSaslUsers(context.Background(), s.client.RegionId, instanceId)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.SaslUserList, nil
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

// DescribeSaslUser retrieves a Kafka SASL user using CWS-Lib-Go
func (s *KafkaService) DescribeSaslUser(instanceId, username string) (*kafka.SaslUser, error) {
	var object kafka.SaslUser
	if err := s.retryWithCommonErrors(5*time.Minute, func() error {
		resp, e := s.kafkaApi.DescribeSaslUsers(context.Background(), s.client.RegionId, instanceId)
		if e != nil {
			return e
		}

		for _, u := range resp.SaslUserList {
			if u.Username == username {
				object = u
				return nil
			}
		}
		return WrapErrorf(NotFoundErr("AlikafkaSaslUser", username), NotFoundMsg, ProviderERROR)
	}); err != nil {
		return nil, err
	}
	return &object, nil
}
