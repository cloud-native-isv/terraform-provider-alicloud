package alicloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

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

var alikafkaRetryableErrors = []string{ThrottlingUser, "ONS_SYSTEM_FLOW_CONTROL"}

func (s *KafkaService) retryWithCommonErrors(timeout time.Duration, fn func() error) error {
	wait := incrementalWait(2*time.Second, 1*time.Second)
	return resource.Retry(timeout, func() *resource.RetryError {
		if err := fn(); err != nil {
			if IsExpectedErrors(err, alikafkaRetryableErrors) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
}

func (s *KafkaService) pageByNumber(pageSize int, fetch func(pageNumber int) (int, error)) error {
	if pageSize <= 0 {
		pageSize = PageSizeLarge
	}

	for pageNumber := 1; ; pageNumber++ {
		count, err := fetch(pageNumber)
		if err != nil {
			return err
		}
		if count < pageSize {
			return nil
		}
	}
}

// ListAlikafkaDeployments lists deployments (instances) using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaDeployments(regionId string) ([]*kafka.KafkaInstance, error) {
	instances, err := s.kafkaApi.ListInstances(regionId)
	if err != nil {
		return nil, WrapError(err)
	}
	return instances, nil
}

// ListAlikafkaAllowedIps lists allowed IPs for an instance using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaAllowedIps(instanceId string) (*kafka.AllowedList, error) {
	allowedList, err := s.kafkaApi.GetAllowedIpList(instanceId, s.client.RegionId)
	if err != nil {
		return nil, WrapError(err)
	}
	return allowedList, nil
}

// AttachAlikafkaAllowedIp attaches an allowed IP to an instance using CWS-Lib-Go
func (s *KafkaService) AttachAlikafkaAllowedIp(instanceId, allowedType, portRange, allowedIp, description string) error {
	if err := s.kafkaApi.UpdateAllowedIp(s.client.RegionId, "add", portRange, allowedType, allowedIp, instanceId, description); err != nil {
		return WrapError(err)
	}
	return nil
}

// DetachAlikafkaAllowedIp detaches an allowed IP from an instance using CWS-Lib-Go
func (s *KafkaService) DetachAlikafkaAllowedIp(instanceId, allowedType, portRange, allowedIp, description string) error {
	if err := s.kafkaApi.UpdateAllowedIp(s.client.RegionId, "delete", portRange, allowedType, allowedIp, instanceId, description); err != nil {
		return WrapError(err)
	}
	return nil
}
