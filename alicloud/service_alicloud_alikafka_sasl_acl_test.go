package alicloud

import "testing"

func TestAliKafkaSaslAclServiceLifecycle(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka SASL ACL lifecycle test requires real infrastructure")
}
