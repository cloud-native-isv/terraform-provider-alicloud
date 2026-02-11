package alicloud

import "testing"

func TestAliKafkaIdEncodeDecode(t *testing.T) {
	topicId := EncodeTopicId("instance-1", "topic-a")
	if topicId == "" {
		t.Fatalf("expected topic id")
	}
	instanceId, topic, err := DecodeTopicId(topicId)
	if err != nil || instanceId != "instance-1" || topic != "topic-a" {
		t.Fatalf("unexpected topic decode: %v, %s, %s", err, instanceId, topic)
	}

	saslUserId := EncodeSaslUserId("instance-1", "user-a")
	inst, user, err := DecodeSaslUserId(saslUserId)
	if err != nil || inst != "instance-1" || user != "user-a" {
		t.Fatalf("unexpected sasl user decode: %v, %s, %s", err, inst, user)
	}

	aclId := EncodeSaslAclId("instance-1", "user-a", "Topic", "t1", "LITERAL", "Read")
	inst, user, resType, resName, patternType, opType, err := DecodeSaslAclId(aclId)
	if err != nil || inst != "instance-1" || user != "user-a" || resType != "Topic" || resName != "t1" || patternType != "LITERAL" || opType != "Read" {
		t.Fatalf("unexpected sasl acl decode: %v", err)
	}

	consumerId := EncodeConsumerGroupId("instance-1", "cg-1")
	inst, cg, err := DecodeConsumerGroupId(consumerId)
	if err != nil || inst != "instance-1" || cg != "cg-1" {
		t.Fatalf("unexpected consumer group decode: %v, %s, %s", err, inst, cg)
	}

	allowedId := EncodeAllowedIpId("instance-1", "vpc", "9092/9092", "10.0.0.1")
	inst, allowedType, portRange, ip, err := DecodeAllowedIpId(allowedId)
	if err != nil || inst != "instance-1" || allowedType != "vpc" || portRange != "9092/9092" || ip != "10.0.0.1" {
		t.Fatalf("unexpected allowed ip decode: %v", err)
	}

	if id, err := DecodeInstanceId(EncodeInstanceId("instance-1")); err != nil || id != "instance-1" {
		t.Fatalf("unexpected instance id decode: %v, %s", err, id)
	}

	if id, err := DecodeDeploymentId(EncodeDeploymentId("instance-1")); err != nil || id != "instance-1" {
		t.Fatalf("unexpected deployment id decode: %v, %s", err, id)
	}
}

func TestAliKafkaBillingAndInstanceTypeResolve(t *testing.T) {
	instanceType, err := ResolveAliKafkaInstanceType(AliKafkaInstanceTypeReserved)
	if err != nil || instanceType != "Reserved" {
		t.Fatalf("unexpected instance type: %v, %v", instanceType, err)
	}

	instanceType, err = ResolveAliKafkaInstanceType(AliKafkaInstanceTypeServerless)
	if err != nil || instanceType != "Serverless" {
		t.Fatalf("unexpected instance type: %v, %v", instanceType, err)
	}

	instanceType, err = ResolveAliKafkaInstanceType("")
	if err != nil || instanceType != "Reserved" {
		t.Fatalf("unexpected default instance type: %v, %v", instanceType, err)
	}

	if _, err = ResolveAliKafkaInstanceType("invalid"); err == nil {
		t.Fatalf("expected error for invalid instance type")
	}

	billingType, err := ResolveAliKafkaPaidType(AliKafkaBillingTypePostPaid)
	if err != nil || billingType != "PostPay" {
		t.Fatalf("unexpected billing type: %v, %v", billingType, err)
	}

	billingType, err = ResolveAliKafkaPaidType(AliKafkaBillingTypePrePaid)
	if err != nil || billingType != "PrePay" {
		t.Fatalf("unexpected billing type: %v, %v", billingType, err)
	}

	billingType, err = ResolveAliKafkaPaidType("")
	if err != nil || billingType != "PostPay" {
		t.Fatalf("unexpected default billing type: %v, %v", billingType, err)
	}

	if _, err = ResolveAliKafkaPaidType("invalid"); err == nil {
		t.Fatalf("expected error for invalid billing type")
	}
}
