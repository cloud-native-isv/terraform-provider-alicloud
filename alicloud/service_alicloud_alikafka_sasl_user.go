package alicloud

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

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
	var err error

	wait := incrementalWait(2*time.Second, 1*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, e := s.kafkaApi.DescribeSaslUsers(context.Background(), s.client.RegionId, instanceId)
		if e != nil {
			if IsExpectedErrors(e, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}

		found := false
		for _, u := range resp.SaslUserList {
			if u.Username == username {
				object = u
				found = true
				break
			}
		}

		if !found {
			return resource.NonRetryableError(WrapErrorf(NotFoundErr("AlikafkaSaslUser", username), NotFoundMsg, ProviderERROR))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &object, nil
}

