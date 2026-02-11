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

func TestAliKafkaInstanceBillingValidation(t *testing.T) {
	if _, _, err := resolveAliKafkaInstanceBilling(AliKafkaInstanceTypeServerless, AliKafkaBillingTypePrePaid); err == nil {
		t.Fatalf("expected error for serverless + prepaid")
	}
	if _, _, err := resolveAliKafkaInstanceBilling(AliKafkaInstanceTypeReserved, AliKafkaBillingTypePrePaid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := resolveAliKafkaInstanceBilling(AliKafkaInstanceTypeReserved, AliKafkaBillingTypePostPaid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
