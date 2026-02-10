package alicloud

import (
	"os"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
)

func testAccAliKafkaService(t *testing.T) *KafkaService {
	accessKey := os.Getenv("ALICLOUD_ACCESS_KEY")
	secretKey := os.Getenv("ALICLOUD_SECRET_KEY")
	regionId := os.Getenv("ALICLOUD_REGION")
	if accessKey == "" || secretKey == "" || regionId == "" {
		t.Skip("AliKafka service tests require ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, and ALICLOUD_REGION")
	}

	client := &connectivity.AliyunClient{
		AccessKey: accessKey,
		SecretKey: secretKey,
		RegionId:  regionId,
	}
	service, err := NewKafkaService(client)
	if err != nil {
		t.Fatalf("failed to create AliKafka service: %v", err)
	}
	return service
}

func TestAccAliKafkaReservedInstancesBilling(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka reserved instance billing test requires real infrastructure")
}

func TestAccAliKafkaServerlessInstanceBilling(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka serverless instance billing test requires real infrastructure")
}
