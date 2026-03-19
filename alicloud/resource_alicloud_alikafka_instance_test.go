package alicloud

import (
	"reflect"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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

func TestAliKafkaInstanceHasOnlineUpgradeChanges(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"instance_type":     AliKafkaInstanceTypeReserved,
		"paid_type":         AliKafkaBillingTypePostPaid,
		"disk_size":         500,
		"disk_type":         "1",
		"deploy_type":       5,
		"partition_num":     2000,
		"io_max_spec":       "alikafka.hw.80xlarge",
		"spec_type":         "normal",
		"resource_group_id": "rg-test",
	})
	resourceData.SetId("alikafka_test_instance")

	if hasAliKafkaInstanceOnlineUpgradeChanges(resourceData) {
		t.Fatalf("expected no online upgrade changes before mutation")
	}

	if err := resourceData.Set("partition_num", 4000); err != nil {
		t.Fatalf("failed to set partition_num: %v", err)
	}

	if !hasAliKafkaInstanceOnlineUpgradeChanges(resourceData) {
		t.Fatalf("expected partition_num change to trigger online upgrade")
	}
}

func TestBuildAliKafkaInstanceUpgradeRequest(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"instance_type": AliKafkaInstanceTypeReserved,
		"paid_type":     AliKafkaBillingTypePrePaid,
		"disk_size":     500,
		"disk_type":     "1",
		"deploy_type":   5,
		"partition_num": 2000,
		"io_max_spec":   "alikafka.hw.80xlarge",
		"spec_type":     "normal",
		"eip_max":       0,
	})
	resourceData.SetId("alikafka_test_instance")

	for field, value := range map[string]interface{}{
		"disk_size":     800,
		"partition_num": 8000,
		"io_max_spec":   "alikafka.hw.120xlarge",
		"eip_max":       0,
	} {
		if err := resourceData.Set(field, value); err != nil {
			t.Fatalf("failed to set %s: %v", field, err)
		}
	}

	paidType := kafka.KafkaPaidTypePrePay
	request := buildAliKafkaInstanceUpgradeRequest(resourceData, &kafka.KafkaInstance{
		InstanceId:   "alikafka_test_instance",
		RegionId:     "cn-shanghai",
		PaidType:     &paidType,
		SpecType:     tea.String("normal"),
		DiskSize:     tea.Int(500),
		PartitionNum: tea.Int(2000),
		IoMaxSpec:    tea.String("alikafka.hw.80xlarge"),
		EipMax:       tea.Int(20),
	}, "cn-hangzhou")

	if request.InstanceId != "alikafka_test_instance" {
		t.Fatalf("unexpected instance id: %s", request.InstanceId)
	}
	if request.RegionId != "cn-shanghai" {
		t.Fatalf("expected remote region to be preserved, got %s", request.RegionId)
	}
	if request.PaidType == nil || *request.PaidType != kafka.KafkaPaidTypePrePay {
		t.Fatalf("expected paid type to be preserved")
	}
	if got := tea.IntValue(request.DiskSize); got != 800 {
		t.Fatalf("unexpected disk_size: %d", got)
	}
	if got := tea.IntValue(request.PartitionNum); got != 8000 {
		t.Fatalf("unexpected partition_num: %d", got)
	}
	if got := tea.StringValue(request.IoMaxSpec); got != "alikafka.hw.120xlarge" {
		t.Fatalf("unexpected io_max_spec: %s", got)
	}
	if got := tea.IntValue(request.EipMax); got != 0 {
		t.Fatalf("unexpected eip_max: %d", got)
	}
	if got := tea.StringValue(request.SpecType); got != "normal" {
		t.Fatalf("unexpected spec_type: %s", got)
	}
}

func TestAliKafkaInstanceOnlineUpgradeableFields(t *testing.T) {
	expected := []string{"disk_size", "spec_type", "partition_num", "io_max_spec", "eip_max"}
	if !reflect.DeepEqual(expected, aliKafkaInstanceOnlineUpgradeableFields) {
		t.Fatalf("unexpected online upgradeable fields: %#v", aliKafkaInstanceOnlineUpgradeableFields)
	}
}

func TestAliKafkaInstanceNameSchemaAllowsConfiguration(t *testing.T) {
	nameSchema := resourceAliCloudAlikafkaInstance().Schema["name"]
	if nameSchema == nil {
		t.Fatal("expected name schema")
	}
	if !nameSchema.Optional {
		t.Fatal("expected name schema to be optional")
	}
	if !nameSchema.Computed {
		t.Fatal("expected name schema to remain computed")
	}
}

func TestBuildAliKafkaInstanceCreationConfigWithName(t *testing.T) {
	config := buildAliKafkaInstanceCreationConfig(&kafka.KafkaInstance{
		RegionId: "cn-hangzhou",
		Name:     tea.String("tf-test-kafka"),
	}, kafka.InstanceSeriesReserved, kafka.BillingTypePostPay)

	if config.Name != "tf-test-kafka" {
		t.Fatalf("expected creation config name to be preserved, got %q", config.Name)
	}
}
