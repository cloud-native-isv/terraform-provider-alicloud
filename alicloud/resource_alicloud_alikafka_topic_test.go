package alicloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAliKafkaTopicFieldMappingRegression(t *testing.T) {
	t.Skip("AliKafka topic regression test requires real infrastructure")
}

func TestAliKafkaTopicPartitionDecreaseForcesNew(t *testing.T) {
	resource := resourceAliCloudAlikafkaTopic()
	diff, err := resource.Diff(&terraform.InstanceState{
		ID: "instance-1:topic-a",
		Attributes: map[string]string{
			"instance_id":   "instance-1",
			"topic":         "topic-a",
			"partition_num": "12",
			"remark":        "topic-a",
		},
	}, terraform.NewResourceConfigRaw(map[string]interface{}{
		"instance_id":   "instance-1",
		"topic":         "topic-a",
		"partition_num": 6,
		"remark":        "topic-a",
	}), nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if diff == nil || diff.Attributes == nil {
		t.Fatal("expected diff attributes")
	}
	attr, ok := diff.Attributes["partition_num"]
	if !ok {
		t.Fatal("expected partition_num diff")
	}
	if !attr.RequiresNew {
		t.Fatal("expected partition_num decrease to force new")
	}
}

func TestAliKafkaTopicPartitionIncreaseDoesNotForceNew(t *testing.T) {
	resource := resourceAliCloudAlikafkaTopic()
	diff, err := resource.Diff(&terraform.InstanceState{
		ID: "instance-1:topic-a",
		Attributes: map[string]string{
			"instance_id":   "instance-1",
			"topic":         "topic-a",
			"partition_num": "12",
			"remark":        "topic-a",
		},
	}, terraform.NewResourceConfigRaw(map[string]interface{}{
		"instance_id":   "instance-1",
		"topic":         "topic-a",
		"partition_num": 18,
		"remark":        "topic-a",
	}), nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if diff == nil || diff.Attributes == nil {
		t.Fatal("expected diff attributes")
	}
	attr, ok := diff.Attributes["partition_num"]
	if !ok {
		t.Fatal("expected partition_num diff")
	}
	if attr.RequiresNew {
		t.Fatal("expected partition_num increase not to force new")
	}
}
