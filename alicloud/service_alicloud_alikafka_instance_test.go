package alicloud

import (
	"fmt"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

func TestAliKafkaInstanceServiceLifecycle(t *testing.T) {
	service := testAccAliKafkaService(t)
	if service == nil {
		t.Fatal("expected AliKafka service")
	}
	t.Skip("AliKafka instance lifecycle test requires real infrastructure")
}

func TestAliKafkaInstanceBillingCombinationValidation(t *testing.T) {
	if err := validateAliKafkaBillingCombination(kafka.InstanceSeriesReserved, kafka.BillingTypePostPay); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateAliKafkaBillingCombination(kafka.InstanceSeriesReserved, kafka.BillingTypePrePay); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateAliKafkaBillingCombination(kafka.InstanceSeriesServerless, kafka.BillingTypePostPay); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateAliKafkaBillingCombination(kafka.InstanceSeriesServerless, kafka.BillingTypePrePay); err == nil {
		t.Fatalf("expected error for serverless + prepay")
	}
}

func TestBuildAliKafkaInstanceCreationConfigReserved(t *testing.T) {
	diskType := kafka.KafkaDiskTypeSSD
	deployType := kafka.KafkaDeployTypeV5
	instance := &kafka.KafkaInstance{
		RegionId:   "cn-hangzhou",
		DiskSize:   tea.Int(300),
		DiskType:   &diskType,
		DeployType: &deployType,
		Duration:   tea.Int(12),
	}

	config := buildAliKafkaInstanceCreationConfig(instance, kafka.InstanceSeriesReserved, kafka.BillingTypePostPay)
	if config.InstanceType != kafka.InstanceSeriesReserved {
		t.Fatalf("unexpected instance type: %v", config.InstanceType)
	}
	if config.BillingType != kafka.BillingTypePostPay {
		t.Fatalf("unexpected billing type: %v", config.BillingType)
	}
	if config.DiskSize != 300 {
		t.Fatalf("unexpected disk size: %v", config.DiskSize)
	}
	if config.DiskType == "" {
		t.Fatalf("expected disk type")
	}
	if config.DeployType != int(kafka.KafkaDeployTypeV5) {
		t.Fatalf("unexpected deploy type: %v", config.DeployType)
	}

	prepayConfig := buildAliKafkaInstanceCreationConfig(instance, kafka.InstanceSeriesReserved, kafka.BillingTypePrePay)
	if prepayConfig.BillingType != kafka.BillingTypePrePay {
		t.Fatalf("unexpected billing type: %v", prepayConfig.BillingType)
	}
	if prepayConfig.Duration != 12 {
		t.Fatalf("unexpected duration: %v", prepayConfig.Duration)
	}
}

func TestBuildAliKafkaInstanceCreationConfigServerless(t *testing.T) {
	instance := &kafka.KafkaInstance{
		RegionId: "cn-hangzhou",
		SpecType: tea.String("normal"),
	}
	config := buildAliKafkaInstanceCreationConfig(instance, kafka.InstanceSeriesServerless, kafka.BillingTypePostPay)
	if config.InstanceType != kafka.InstanceSeriesServerless {
		t.Fatalf("unexpected instance type: %v", config.InstanceType)
	}
	if config.BillingType != kafka.BillingTypePostPay {
		t.Fatalf("unexpected billing type: %v", config.BillingType)
	}
	if config.SpecType != "normal" {
		t.Fatalf("unexpected spec type: %v", config.SpecType)
	}
}

func TestAliKafkaInstanceUpdateTargetStates(t *testing.T) {
	states := aliKafkaInstanceUpdateTargetStates()

	hasRunning := false
	hasCreated := false
	hasChanging := false
	for _, state := range states {
		if state == fmt.Sprint(kafka.KafkaViewInstanceStatusRunning) {
			hasRunning = true
		}
		if state == fmt.Sprint(kafka.KafkaViewInstanceStatusCreated) {
			hasCreated = true
		}
		if state == fmt.Sprint(kafka.KafkaViewInstanceStatusChanging) {
			hasChanging = true
		}
	}

	if !hasRunning {
		t.Fatalf("expected update target states to include running")
	}
	if !hasCreated {
		t.Fatalf("expected update target states to include created for compatibility")
	}
	if hasChanging {
		t.Fatalf("expected update target states not to include changing")
	}
}
