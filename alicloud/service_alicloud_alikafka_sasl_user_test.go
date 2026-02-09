package alicloud

import "testing"

func TestAliKafkaSaslUserServiceLifecycle(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka SASL user lifecycle test requires real infrastructure")
}
