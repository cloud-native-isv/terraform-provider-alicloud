package alicloud

import "testing"

func TestAliKafkaInstanceAcceptanceSkeleton(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka instance acceptance test requires real infrastructure")
}

func TestAccAliKafkaInstance(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka instance testacc requires real infrastructure")
}
