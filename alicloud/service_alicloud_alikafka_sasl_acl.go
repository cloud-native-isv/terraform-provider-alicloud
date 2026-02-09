package alicloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alikafka"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

// CreateAlikafkaSaslAcl creates a SASL ACL using CWS-Lib-Go
func (s *KafkaService) CreateAlikafkaSaslAcl(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string, options map[string]interface{}) error {
	if options == nil {
		options = map[string]interface{}{}
	}
	if err := s.kafkaApi.CreateAcl(context.Background(), s.client.RegionId, instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType, options); err != nil {
		return WrapError(err)
	}
	return nil
}

// DeleteAlikafkaSaslAcl deletes a SASL ACL using CWS-Lib-Go
func (s *KafkaService) DeleteAlikafkaSaslAcl(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string, options map[string]interface{}) error {
	if options == nil {
		options = map[string]interface{}{}
	}
	if err := s.kafkaApi.DeleteAcl(context.Background(), s.client.RegionId, instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType, options); err != nil {
		return WrapError(err)
	}
	return nil
}

// ListAlikafkaSaslAcls lists SASL ACLs using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaSaslAcls(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType string) ([]kafka.AclRule, error) {
	options := map[string]interface{}{}
	if aclResourcePatternType != "" {
		options["aclResourcePatternType"] = aclResourcePatternType
	}
	if aclOperationType != "" {
		options["aclOperationType"] = aclOperationType
	}
	resp, err := s.kafkaApi.DescribeAcls(context.Background(), s.client.RegionId, instanceId, username, aclResourceType, aclResourceName, options)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.AclList, nil
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
	if err := s.retryWithCommonErrors(5*time.Minute, func() error {
		options := map[string]interface{}{
			"aclResourcePatternType": aclResourcePatternType,
			"aclOperationType":       aclOperationType,
		}
		resp, e := s.kafkaApi.DescribeAcls(context.Background(), s.client.RegionId, instanceId, username, aclResourceType, aclResourceName, options)
		if e != nil {
			return e
		}

		// Exact match check if multiple returned
		for _, acl := range resp.AclList {
			if acl.Username == username && acl.AclResourceType == aclResourceType && acl.AclResourceName == aclResourceName {
				object = acl
				return nil
			}
		}
		return WrapErrorf(NotFoundErr("AlikafkaSaslAcl", username), NotFoundMsg, ProviderERROR)
	}); err != nil {
		return nil, err
	}
	return &object, nil
}
