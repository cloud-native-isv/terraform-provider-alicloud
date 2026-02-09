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

// DescribeSaslAcl retrieves a Kafka SASL ACL using CWS-Lib-Go
func (s *KafkaService) DescribeSaslAcl(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string) (*kafka.AclRule, error) {
	var object kafka.AclRule
	var err error

	wait := incrementalWait(2*time.Second, 1*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		options := map[string]interface{}{
			"aclResourcePatternType": aclResourcePatternType,
			"aclOperationType":       aclOperationType,
		}
		resp, e := s.kafkaApi.DescribeAcls(context.Background(), s.client.RegionId, instanceId, username, aclResourceType, aclResourceName, options)
		if e != nil {
			if IsExpectedErrors(e, []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}) {
				wait()
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}

		if len(resp.AclList) == 0 {
			return resource.NonRetryableError(WrapErrorf(NotFoundErr("AlikafkaSaslAcl", username), NotFoundMsg, ProviderERROR))
		}

		// Exact match check if multiple returned
		for _, acl := range resp.AclList {
			if acl.Username == username && acl.AclResourceType == aclResourceType && acl.AclResourceName == aclResourceName {
				object = acl
				return nil
			}
		}
		return resource.NonRetryableError(WrapErrorf(NotFoundErr("AlikafkaSaslAcl", username), NotFoundMsg, ProviderERROR))
	})

	if err != nil {
		return nil, err
	}
	return &object, nil
}
