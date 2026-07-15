package flinkworkspace

import "testing"

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
