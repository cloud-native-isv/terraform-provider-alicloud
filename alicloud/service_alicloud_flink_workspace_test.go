package alicloud

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type fakeFlinkRefundClient struct {
	international bool
	product       string
	version       string
	action        string
	request       map[string]interface{}
	requests      []map[string]interface{}
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
	f.requests = append(f.requests, request)
	f.autoRetry = autoRetry
	f.endpoint = endpoint
	return map[string]interface{}{"RequestId": "request-test"}, f.err
}

func TestRefundFlinkWorkspaceForInstanceReusesClientToken(t *testing.T) {
	client := &fakeFlinkRefundClient{}
	for i := 0; i < 2; i++ {
		if err := refundFlinkWorkspaceForInstance(client, "cn-beijing", "f-test"); err != nil {
			t.Fatal(err)
		}
	}
	if len(client.requests) != 2 || client.requests[0]["ClientToken"] != client.requests[1]["ClientToken"] {
		t.Fatalf("refund tokens are not stable across retries: %#v", client.requests)
	}
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
	workspace             *flink.Workspace
	describeErr           error
	describeAfterFirstErr error
	deleteErr             error
	refundErr             error
	waitErr               error
	describes             int
	deletes               int
	refunds               int
	waits                 int
}

type fakeFlinkWorkspaceCreateService struct {
	listResponses  [][]flink.Workspace
	listErr        error
	createResponse *flink.Workspace
	createErr      error
	listErrs       []error
	createCalls    int
	listCalls      int
	options        flinkworkspace.CreateOptions
	request        *flink.Workspace
}

func (f *fakeFlinkWorkspaceCreateService) ListInstances() ([]flink.Workspace, error) {
	f.listCalls++
	if index := f.listCalls - 1; index < len(f.listErrs) && f.listErrs[index] != nil {
		return nil, f.listErrs[index]
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	if len(f.listResponses) == 0 {
		return nil, nil
	}
	index := f.listCalls - 1
	if index >= len(f.listResponses) {
		index = len(f.listResponses) - 1
	}
	return f.listResponses[index], nil
}

func TestCreateFlinkWorkspaceRetriesTransientDiscoveryAfterAmbiguousCreate(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	token := flinkworkspace.WorkspaceCreateToken(request)
	service := &fakeFlinkWorkspaceCreateService{
		listResponses: [][]flink.Workspace{
			nil,
			nil,
			nil,
			{createTokenWorkspace("f-recovered", request.Name, request.Region, token)},
		},
		listErrs:  []error{nil, io.EOF},
		createErr: &ambiguousFlinkWorkspaceCreateError{cause: errors.New("response lost")},
	}

	workspace, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.Id != "f-recovered" || service.createCalls != 1 || service.listCalls != 4 {
		t.Fatalf("workspace=%#v createCalls=%d listCalls=%d", workspace, service.createCalls, service.listCalls)
	}
}

func TestCreateFlinkWorkspacePersistsPendingMarkerAfterRecoveryTimeout(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	service := &fakeFlinkWorkspaceCreateService{
		listResponses: [][]flink.Workspace{nil},
		createErr:     &ambiguousFlinkWorkspaceCreateError{cause: errors.New("response lost")},
	}

	_, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, time.Millisecond)
	var pending *pendingFlinkWorkspaceCreateError
	if !errors.As(err, &pending) {
		t.Fatalf("error = %T %v, want pending marker error", err, err)
	}
	wantToken := flinkworkspace.WorkspaceCreateToken(request)
	if pending.token != wantToken || pendingFlinkWorkspaceCreateID(pending.token) == "" {
		t.Fatalf("pending = %#v", pending)
	}
	if got, ok := flinkWorkspaceCreateTokenFromPendingID(pendingFlinkWorkspaceCreateID(pending.token)); !ok || got != wantToken {
		t.Fatalf("pending ID token = %q, %v", got, ok)
	}
}

func TestSDKPersistsPendingCreateIDWhenCreateReturnsError(t *testing.T) {
	wantID := pendingFlinkWorkspaceCreateID("token")
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
		},
		Create: func(d *schema.ResourceData, _ interface{}) error {
			d.SetId(wantID)
			return errors.New("ambiguous create")
		},
	}
	diff, err := resource.Diff(nil, terraform.NewResourceConfigRaw(map[string]interface{}{"name": "workspace"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	state, err := resource.Apply(nil, diff, nil)
	if err == nil {
		t.Fatal("expected create error")
	}
	if state == nil || state.ID != wantID {
		t.Fatalf("state = %#v, want pending ID %q", state, wantID)
	}
}

func (f *fakeFlinkWorkspaceCreateService) CreateInstance(request *flink.Workspace, options flinkworkspace.CreateOptions) (*flink.Workspace, error) {
	f.createCalls++
	f.options = options
	f.request = request
	return f.createResponse, f.createErr
}

func createTokenWorkspace(id, name, region, token string) flink.Workspace {
	return flink.Workspace{
		Id:     id,
		Name:   name,
		Region: region,
		Tags:   []flink.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}},
	}
}

func TestCreateFlinkWorkspaceRecoversBeforePurchasing(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	token := flinkworkspace.WorkspaceCreateToken(request)
	service := &fakeFlinkWorkspaceCreateService{listResponses: [][]flink.Workspace{{
		createTokenWorkspace("f-recovered", request.Name, request.Region, token),
	}}}

	workspace, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.Id != "f-recovered" || service.createCalls != 0 {
		t.Fatalf("workspace=%#v createCalls=%d", workspace, service.createCalls)
	}
}

func TestCreateFlinkWorkspaceRecoversAmbiguousResponse(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	token := flinkworkspace.WorkspaceCreateToken(request)
	service := &fakeFlinkWorkspaceCreateService{
		listResponses: [][]flink.Workspace{
			nil,
			{createTokenWorkspace("f-recovered", request.Name, request.Region, token)},
		},
		createErr: &ambiguousFlinkWorkspaceCreateError{cause: errors.New("response lost")},
	}

	workspace, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.Id != "f-recovered" || service.createCalls != 1 || service.listCalls != 2 {
		t.Fatalf("workspace=%#v createCalls=%d listCalls=%d", workspace, service.createCalls, service.listCalls)
	}
}

func TestCreateFlinkWorkspaceRefusesAmbiguousRecovery(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	token := flinkworkspace.WorkspaceCreateToken(request)
	service := &fakeFlinkWorkspaceCreateService{listResponses: [][]flink.Workspace{{
		createTokenWorkspace("f-one", request.Name, request.Region, token),
		createTokenWorkspace("f-two", request.Name, request.Region, token),
	}}}

	_, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, time.Minute)
	if err == nil || !strings.Contains(err.Error(), "multiple") || service.createCalls != 0 {
		t.Fatalf("error=%v createCalls=%d", err, service.createCalls)
	}
}

func TestCreateFlinkWorkspaceRefusesTaggedImmutableMismatch(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing", VpcId: "vpc-expected", ChargeType: "PRE"}
	token := flinkworkspace.WorkspaceCreateToken(request)
	mismatch := createTokenWorkspace("f-existing", request.Name, request.Region, token)
	mismatch.VpcId = "vpc-other"
	mismatch.ChargeType = request.ChargeType
	service := &fakeFlinkWorkspaceCreateService{listResponses: [][]flink.Workspace{{mismatch}}}

	_, err := createFlinkWorkspace(service, request, flinkworkspace.CreateOptions{}, time.Minute)
	if err == nil || !strings.Contains(err.Error(), "VPC") || service.createCalls != 0 {
		t.Fatalf("error=%v createCalls=%d", err, service.createCalls)
	}
}

func TestCreateFlinkWorkspaceAddsStableToken(t *testing.T) {
	request := &flink.Workspace{Name: "workspace", Region: "cn-beijing"}
	service := &fakeFlinkWorkspaceCreateService{
		listResponses:  [][]flink.Workspace{nil},
		createResponse: &flink.Workspace{Id: "f-created"},
	}
	autoRenew := true
	duration := int32(1)
	options := flinkworkspace.CreateOptions{AutoRenew: &autoRenew, Duration: &duration, PricingCycle: "Month"}
	if _, err := createFlinkWorkspace(service, request, options, time.Minute); err != nil {
		t.Fatal(err)
	}
	if service.options.AutoRenew == nil || !*service.options.AutoRenew || service.options.Duration == nil || *service.options.Duration != 1 || service.options.PricingCycle != "Month" {
		t.Fatalf("options=%#v", service.options)
	}
	token := flinkworkspace.WorkspaceCreateToken(request)
	if service.request == nil || !flinkWorkspaceHasTag(service.request.Tags, flinkworkspace.CreateTokenTagKey, token) {
		t.Fatalf("request tags = %#v", service.request)
	}
}

func TestClassifyFlinkWorkspaceCreateErrorTreatsTransportAsAmbiguous(t *testing.T) {
	classified := classifyFlinkWorkspaceCreateError(io.EOF)
	if !isAmbiguousFlinkWorkspaceCreateError(classified) {
		t.Fatalf("bare transport error was not ambiguous: %T %v", classified, classified)
	}

	rejected := tea.NewSDKError(map[string]interface{}{
		"statusCode": 400,
		"code":       "InvalidParameter",
		"message":    "invalid request",
	})
	classified = classifyFlinkWorkspaceCreateError(rejected)
	if isAmbiguousFlinkWorkspaceCreateError(classified) {
		t.Fatalf("definitive service rejection was marked ambiguous: %T %v", classified, classified)
	}
}

func (f *fakeFlinkWorkspaceDeleteService) DescribeFlinkWorkspace(string) (*flink.Workspace, error) {
	f.describes++
	if f.describes > 1 && f.describeAfterFirstErr != nil {
		return nil, f.describeAfterFirstErr
	}
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

func TestDeleteFlinkWorkspaceRefundNotFoundDoesNotLoseExistingWorkspace(t *testing.T) {
	service := &fakeFlinkWorkspaceDeleteService{
		workspace: &flink.Workspace{Id: "f-test", ChargeType: "PRE"},
		refundErr: errors.New("HTTP 404 order not found"),
	}
	err := deleteFlinkWorkspace(service, "f-test", time.Minute)
	if err == nil || !strings.Contains(err.Error(), "order not found") {
		t.Fatalf("error = %v", err)
	}
	if service.describes != 2 || service.waits != 0 {
		t.Fatalf("describes=%d waits=%d", service.describes, service.waits)
	}
}

func TestDeleteFlinkWorkspaceOperationErrorSucceedsOnlyAfterConfirmedAbsent(t *testing.T) {
	service := &fakeFlinkWorkspaceDeleteService{
		workspace:             &flink.Workspace{Id: "f-test", ChargeType: "PRE"},
		refundErr:             errors.New("response lost"),
		describeAfterFirstErr: errors.New("workspace not found"),
	}
	if err := deleteFlinkWorkspace(service, "f-test", time.Minute); err != nil {
		t.Fatal(err)
	}
	if service.describes != 2 || service.waits != 0 {
		t.Fatalf("describes=%d waits=%d", service.describes, service.waits)
	}
}
