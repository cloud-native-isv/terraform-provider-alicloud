package alicloud

import "testing"

func TestAliKafkaTopicServiceLifecycle(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka topic lifecycle test requires real infrastructure")
}
