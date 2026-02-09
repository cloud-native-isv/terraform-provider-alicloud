package alicloud

import (
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
