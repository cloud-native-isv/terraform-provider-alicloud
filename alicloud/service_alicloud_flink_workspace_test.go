package alicloud

import (
	"errors"
	"strings"
	"testing"
	"time"

	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type fakeFlinkRefundClient struct {
	international bool
	product       string
	version       string
	action        string
	request       map[string]interface{}
	autoRetry     bool
	endpoint      string
	err           error
}

func (f *fakeFlinkRefundClient) IsInternationalAccount() bool { return f.international }

func (f *fakeFlinkRefundClient) RpcPostWithEndpoint(product, version, action string, _ map[string]interface{}, request map[string]interface{}, autoRetry bool, endpoint string) (map[string]interface{}, error) {
	f.product = product
	f.version = version
	f.action = action
	f.request = request
	f.autoRetry = autoRetry
	f.endpoint = endpoint
	return map[string]interface{}{"RequestId": "request-test"}, f.err
}

func TestRefundFlinkWorkspaceUsesPublicBSSAPI(t *testing.T) {
	client := &fakeFlinkRefundClient{international: true}
	if err := refundFlinkWorkspace(client, "ap-southeast-1", "f-test", "token-test"); err != nil {
		t.Fatal(err)
	}
	if client.product != "BssOpenApi" || client.version != "2017-12-14" || client.action != "RefundInstance" || !client.autoRetry {
		t.Fatalf("RPC = %s/%s/%s autoRetry=%v", client.product, client.version, client.action, client.autoRetry)
	}
	for key, want := range map[string]interface{}{
		"InstanceId":         "f-test",
		"ClientToken":        "token-test",
		"ImmediatelyRelease": "1",
		"ProductCode":        "sc",
		"ProductType":        "sc_flinkserverless_public_intl",
	} {
		if got := client.request[key]; got != want {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}
	}
}

func TestRefundFlinkWorkspacePreservesCommodityErrorContext(t *testing.T) {
	client := &fakeFlinkRefundClient{err: errors.New("CommodityNotSupported")}
	err := refundFlinkWorkspace(client, "cn-beijing", "f-test", "token-test")
	if err == nil {
		t.Fatal("expected refund failure")
	}
	for _, fragment := range []string{"cn-beijing", "f-test", "ProductCode=sc", "ProductType=sc_flinkserverless_public_cn", "CommodityNotSupported"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q does not contain %q", err, fragment)
		}
	}
}

type fakeFlinkWorkspaceDeleteService struct {
	workspace   *flink.Workspace
	describeErr error
	deleteErr   error
	refundErr   error
	waitErr     error
	deletes     int
	refunds     int
	waits       int
}

func (f *fakeFlinkWorkspaceDeleteService) DescribeFlinkWorkspace(string) (*flink.Workspace, error) {
	return f.workspace, f.describeErr
}

func (f *fakeFlinkWorkspaceDeleteService) DeleteInstance(string) error {
	f.deletes++
	return f.deleteErr
}

func (f *fakeFlinkWorkspaceDeleteService) RefundInstance(string) error {
	f.refunds++
	return f.refundErr
}

func (f *fakeFlinkWorkspaceDeleteService) WaitForWorkspaceDeleting(string, time.Duration) error {
	f.waits++
	return f.waitErr
}

func TestDeleteFlinkWorkspaceBranchesByActualChargeType(t *testing.T) {
	for _, test := range []struct {
		chargeType  string
		wantDeletes int
		wantRefunds int
	}{
		{chargeType: "POST", wantDeletes: 1},
		{chargeType: "PRE", wantRefunds: 1},
	} {
		t.Run(test.chargeType, func(t *testing.T) {
			service := &fakeFlinkWorkspaceDeleteService{workspace: &flink.Workspace{Id: "f-test", ChargeType: test.chargeType}}
			if err := deleteFlinkWorkspace(service, "f-test", time.Minute); err != nil {
				t.Fatal(err)
			}
			if service.deletes != test.wantDeletes || service.refunds != test.wantRefunds || service.waits != 1 {
				t.Fatalf("deletes=%d refunds=%d waits=%d", service.deletes, service.refunds, service.waits)
			}
		})
	}
}

func TestDeleteFlinkWorkspaceRefundFailureStopsAndReturnsError(t *testing.T) {
	service := &fakeFlinkWorkspaceDeleteService{
		workspace: &flink.Workspace{Id: "f-test", ChargeType: "PRE"},
		refundErr: errors.New("CommodityNotSupported"),
	}
	err := deleteFlinkWorkspace(service, "f-test", time.Minute)
	if err == nil || !strings.Contains(err.Error(), "CommodityNotSupported") {
		t.Fatalf("error = %v", err)
	}
	if service.deletes != 0 || service.refunds != 1 || service.waits != 0 {
		t.Fatalf("deletes=%d refunds=%d waits=%d", service.deletes, service.refunds, service.waits)
	}
}
