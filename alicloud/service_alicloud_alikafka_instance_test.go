package alicloud

import "testing"

func TestAliKafkaInstanceServiceLifecycle(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka instance lifecycle test requires real infrastructure")
}
