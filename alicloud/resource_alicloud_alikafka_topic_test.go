package alicloud

import "testing"

func TestAliKafkaTopicFieldMappingRegression(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka topic regression test requires real infrastructure")
}
