package alicloud

import (
	"reflect"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAliKafkaInstanceAcceptanceSkeleton(t *testing.T) {
	t.Skip("AliKafka instance acceptance test requires real infrastructure")
}

func TestAccAliKafkaInstance(t *testing.T) {
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
	testSchema := map[string]*schema.Schema{
		"disk_size":     {Type: schema.TypeInt, Optional: true},
		"partition_num": {Type: schema.TypeInt, Optional: true},
		"eip_max":       {Type: schema.TypeInt, Optional: true},
		"spec_type":     {Type: schema.TypeString, Optional: true},
		"io_max_spec":   {Type: schema.TypeString, Optional: true},
	}
	resourceData := schema.TestResourceDataRaw(t, testSchema, map[string]interface{}{})

	if hasAliKafkaInstanceOnlineUpgradeChanges(resourceData) {
		t.Fatalf("expected no online upgrade changes before mutation")
	}

	changedResourceData := schema.TestResourceDataRaw(t, testSchema, map[string]interface{}{
		"partition_num": 4000,
	})
	if !hasAliKafkaInstanceOnlineUpgradeChanges(changedResourceData) {
		t.Fatalf("expected partition_num change to trigger online upgrade")
	}
}

func TestBuildAliKafkaInstanceUpgradeRequest(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"disk_size":     {Type: schema.TypeInt, Optional: true},
		"partition_num": {Type: schema.TypeInt, Optional: true},
		"eip_max":       {Type: schema.TypeInt, Optional: true},
		"spec_type":     {Type: schema.TypeString, Optional: true},
		"io_max_spec":   {Type: schema.TypeString, Optional: true},
	}
	state := &terraform.InstanceState{
		ID: "alikafka_test_instance",
		Attributes: map[string]string{
			"disk_size":     "500",
			"partition_num": "2000",
			"io_max_spec":   "alikafka.hw.80xlarge",
			"spec_type":     "normal",
			"eip_max":       "20",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"disk_size":     800,
		"partition_num": 8000,
		"io_max_spec":   "alikafka.hw.120xlarge",
		"spec_type":     "normal",
		"eip_max":       0,
	})
	diff, err := schema.InternalMap(testSchema).Diff(state, config, nil, nil, true)
	if err != nil {
		t.Fatalf("build diff: %v", err)
	}
	resourceData, err := schema.InternalMap(testSchema).Data(state, diff)
	if err != nil {
		t.Fatalf("build resource data: %v", err)
	}
	if !resourceData.HasChange("eip_max") {
		t.Fatalf("expected eip_max change from 20 to 0")
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

func TestSetAliKafkaInstancePartitionNumState_UseRemoteValue(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"partition_num": 2000,
	})

	setAliKafkaInstancePartitionNumState(resourceData, &kafka.KafkaInstance{PartitionNum: tea.Int(8000)})

	if got := resourceData.Get("partition_num").(int); got != 8000 {
		t.Fatalf("expected partition_num from remote, got %d", got)
	}
}

func TestSetAliKafkaInstancePartitionNumState_FallbackToState(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"partition_num": 8000,
	})

	setAliKafkaInstancePartitionNumState(resourceData, &kafka.KafkaInstance{})

	if got := resourceData.Get("partition_num").(int); got != 8000 {
		t.Fatalf("expected partition_num to fallback to state, got %d", got)
	}
}

func TestSetAliKafkaInstanceEipMaxState_UseRemoteValue(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"eip_max": 20,
	})

	setAliKafkaInstanceEipMaxState(resourceData, &kafka.KafkaInstance{EipMax: tea.Int(0)})

	if got := resourceData.Get("eip_max").(int); got != 0 {
		t.Fatalf("expected eip_max from remote, got %d", got)
	}
}

func TestSetAliKafkaInstanceEipMaxState_FallbackToState(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceAliCloudAlikafkaInstance().Schema, map[string]interface{}{
		"eip_max": 15,
	})

	setAliKafkaInstanceEipMaxState(resourceData, &kafka.KafkaInstance{})

	if got := resourceData.Get("eip_max").(int); got != 15 {
		t.Fatalf("expected eip_max to fallback to state, got %d", got)
	}
}
