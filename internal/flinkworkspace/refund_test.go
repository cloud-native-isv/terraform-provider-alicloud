package flinkworkspace

import (
	"strings"
	"testing"
)

func TestRefundProductType(t *testing.T) {
	if got := RefundProductType(false); got != "sc_flinkserverless_public_cn" {
		t.Fatalf("domestic product type = %q", got)
	}
	if got := RefundProductType(true); got != "sc_flinkserverless_public_intl" {
		t.Fatalf("international product type = %q", got)
	}
}

func TestBuildRefundRequest(t *testing.T) {
	request := BuildRefundRequest("f-test", true, "token-test")
	for key, want := range map[string]interface{}{
		"InstanceId":         "f-test",
		"ClientToken":        "token-test",
		"ImmediatelyRelease": "1",
		"ProductCode":        "sc",
		"ProductType":        "sc_flinkserverless_public_intl",
	} {
		if got := request[key]; got != want {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}
	}
}

func TestRefundClientTokenIsStableForAnInstance(t *testing.T) {
	first := RefundClientToken("cn-beijing", "f-test")
	second := RefundClientToken("cn-beijing", "f-test")
	other := RefundClientToken("cn-beijing", "f-other")
	otherRegion := RefundClientToken("ap-southeast-1", "f-test")

	if first != second {
		t.Fatalf("same instance produced different tokens: %q != %q", first, second)
	}
	if first == other {
		t.Fatalf("different instances produced the same token %q", first)
	}
	if first == otherRegion {
		t.Fatalf("different regions produced the same token %q", first)
	}
	if len(first) > 64 || !strings.HasPrefix(first, "TF-RefundInstance-") {
		t.Fatalf("invalid refund token %q", first)
	}
}
