package alicloud

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type fakeFlinkWorkspaceIdentityBoundaryService struct {
	workspace       *aliyunFlinkAPI.Workspace
	describedIDs    []string
	listCalls       int
	deleteCalls     int
	refundCalls     int
	waitDeleteCalls int
}

func (f *fakeFlinkWorkspaceIdentityBoundaryService) DescribeFlinkWorkspace(id string) (*aliyunFlinkAPI.Workspace, error) {
	f.describedIDs = append(f.describedIDs, id)
	return f.workspace, nil
}

func (f *fakeFlinkWorkspaceIdentityBoundaryService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	f.listCalls++
	return nil, nil
}

func (f *fakeFlinkWorkspaceIdentityBoundaryService) DeleteInstance(string) error {
	f.deleteCalls++
	return nil
}

func (f *fakeFlinkWorkspaceIdentityBoundaryService) RefundInstance(string) error {
	f.refundCalls++
	return nil
}

func (f *fakeFlinkWorkspaceIdentityBoundaryService) WaitForWorkspaceDeleting(string, time.Duration) error {
	f.waitDeleteCalls++
	return nil
}

func TestFlinkWorkspaceV0InitialDescribeIdentityMismatchPreservesSingleAndHAFlatmapCtyState(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "HA"}[ha], func(t *testing.T) {
			meta := &connectivity.AliyunClient{RegionId: "cn-test"}
			legacy := initialWorkspaceState(ha)
			legacy.Meta = map[string]interface{}{"schema_version": "0"}

			expectedResource := resourceAliCloudFlinkWorkspace()
			expectedResource.Read = schema.Noop
			expected, err := expectedResource.Refresh(legacy.DeepCopy(), meta)
			if err != nil {
				t.Fatal(err)
			}

			token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: meta.RegionId, Name: "workspace"})
			other := flinkWorkspaceInitialLifecycleFixture(ha)
			other.Id = "f-other"
			other.Name = "other-workspace"
			other.ResourceGroupId = "rg-other"
			other.VpcId = "vpc-other"
			other.Tags = []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}}
			service := &fakeFlinkWorkspaceIdentityBoundaryService{workspace: other}
			resource := resourceAliCloudFlinkWorkspace()
			resource.Read = func(d *schema.ResourceData, _ interface{}) error {
				return readFlinkWorkspaceWithService(d, service)
			}

			got, err := resource.Refresh(legacy.DeepCopy(), meta)
			if err == nil || !strings.Contains(err.Error(), "f-other") {
				t.Fatalf("Describe(%q) other-ID response error=%v state=%#v, want identity mismatch", legacy.ID, err, got)
			}
			assertFlinkWorkspaceInstanceStateUnchanged(t, expected, got)
			assertFlinkWorkspaceProtocolAttributes(t, got, map[string]string{
				"purchase_options_state":    flinkWorkspacePurchaseMigratedInitial,
				"capacity_intent_mode":      flinkWorkspaceCapacityInitial,
				"identity_visibility_state": flinkWorkspaceIdentityMigratedFirstRead,
				"terraform_create_token":    token,
				"create_intent_fingerprint": flinkWorkspaceProtocolUnavailable,
			})
			if !reflect.DeepEqual(service.describedIDs, []string{"f-test"}) || service.listCalls != 0 {
				t.Fatalf("Describe/List calls=%v/%d, want [f-test]/0", service.describedIDs, service.listCalls)
			}
		})
	}
}

func TestFlinkWorkspaceStableReadRejectsEmptyOrOtherDescribeIdentityBeforeStateMutation(t *testing.T) {
	for _, responseID := range []string{"", "f-other"} {
		name := "empty"
		if responseID != "" {
			name = "other"
		}
		t.Run(name, func(t *testing.T) {
			config := flinkWorkspaceInitialLifecycleConfig(false)
			resource := resourceAliCloudFlinkWorkspace()
			data := schema.TestResourceDataRaw(t, resource.Schema, config)
			request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentInitial)
			if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
				t.Fatal(err)
			}
			data.SetId("f-test")
			if err := data.Set("identity_visibility_state", flinkWorkspaceIdentityStable); err != nil {
				t.Fatal(err)
			}
			before := data.State()

			other := flinkWorkspaceInitialLifecycleFixture(false)
			other.Id = responseID
			other.Name = "other-workspace"
			other.ResourceGroupId = "rg-other"
			other.VpcId = "vpc-other"
			service := &fakeFlinkWorkspaceIdentityBoundaryService{workspace: other}
			err := readFlinkWorkspaceWithService(data, service)
			if err == nil || !strings.Contains(err.Error(), "f-test") {
				t.Fatalf("response ID %q error=%v, want fail-closed identity error", responseID, err)
			}
			assertFlinkWorkspaceInstanceStateUnchanged(t, before, data.State())
			if !reflect.DeepEqual(service.describedIDs, []string{"f-test"}) || service.listCalls != 0 {
				t.Fatalf("Describe/List calls=%v/%d, want [f-test]/0", service.describedIDs, service.listCalls)
			}
		})
	}
}

func TestFlinkWorkspaceRecoveryReadRejectsOtherDescribeIdentityWithCorrectTokenBeforeStateMutation(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: "cn-test", Name: "workspace"})
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-test|recover="+token+"|initial")
	if err != nil {
		t.Fatal(err)
	}
	data := resource.Data(states[0])
	before := data.State()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	other := flinkWorkspaceInitialLifecycleFixture(false)
	other.Id = "f-other"
	other.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	service := &fakeFlinkWorkspaceIdentityBoundaryService{workspace: other}

	err = readFlinkWorkspaceWithService(data, service)
	if err == nil || !strings.Contains(err.Error(), "f-other") {
		t.Fatalf("recovery Describe(f-test) response ID %q error=%v, want identity mismatch", other.Id, err)
	}
	assertFlinkWorkspaceInstanceStateUnchanged(t, before, data.State())
	if !reflect.DeepEqual(service.describedIDs, []string{"f-test"}) || service.listCalls != 0 {
		t.Fatalf("Describe/List calls=%v/%d, want [f-test]/0", service.describedIDs, service.listCalls)
	}
}

func TestFlinkWorkspaceAdoptionRejectsOtherDescribeIdentityBeforeProtocolOrCloudWrites(t *testing.T) {
	token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: "cn-test", Name: "workspace"})
	for _, test := range []struct {
		name     string
		ha       bool
		recovery bool
	}{
		{name: "initial single"},
		{name: "initial HA", ha: true},
		{name: "recovery single", recovery: true},
		{name: "recovery HA", ha: true, recovery: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceInitialLifecycleConfig(test.ha)
			exact := flinkWorkspaceInitialLifecycleFixture(test.ha)
			exact.Id = "f-test"
			importID := "f-test|initial"
			if test.recovery {
				importID = "f-test|recover=" + token + "|initial"
				exact.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
			}
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, importID, exact)
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			other := *exact
			other.Id = "f-other"
			other.ResourceId = "resource-from-f-other"
			service := &fakeFlinkWorkspaceIdentityBoundaryService{workspace: &other}
			resource.Update = func(d *schema.ResourceData, _ interface{}) error {
				return updateFlinkWorkspaceWithService(d, service)
			}

			state, err := resource.Apply(prior, diff, nil)
			if err == nil || !strings.Contains(err.Error(), "f-other") {
				t.Fatalf("adoption response ID %q state=%#v error=%v, want identity mismatch", other.Id, state, err)
			}
			assertFlinkWorkspaceIdentityAndProtocolUnchanged(t, prior, state)
			if got, want := state.Attributes["resource_id"], prior.Attributes["resource_id"]; got != want {
				t.Fatalf("other-ID response materialized resource_id=%q, want prior %q", got, want)
			}
			if !reflect.DeepEqual(service.describedIDs, []string{"f-test"}) {
				t.Fatalf("Describe IDs=%v, want [f-test]", service.describedIDs)
			}
			if service.deleteCalls != 0 || service.refundCalls != 0 || service.waitDeleteCalls != 0 {
				t.Fatalf("adoption identity mismatch reached cloud writes: delete/refund/wait=%d/%d/%d", service.deleteCalls, service.refundCalls, service.waitDeleteCalls)
			}
		})
	}
}

func TestDeleteFlinkWorkspaceRejectsEmptyOrOtherDescribeIdentityBeforeCloudWrite(t *testing.T) {
	for _, test := range []struct {
		name       string
		responseID string
		chargeType string
	}{
		{name: "empty prepaid", chargeType: "PRE"},
		{name: "other prepaid", responseID: "f-other", chargeType: "PRE"},
		{name: "other postpaid", responseID: "f-other", chargeType: "POST"},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeFlinkWorkspaceIdentityBoundaryService{workspace: &aliyunFlinkAPI.Workspace{Id: test.responseID, ChargeType: test.chargeType}}
			err := deleteFlinkWorkspace(service, "f-test", time.Minute)
			if err == nil || !strings.Contains(err.Error(), "f-test") {
				t.Fatalf("delete preflight response ID %q error=%v, want identity mismatch", test.responseID, err)
			}
			if !reflect.DeepEqual(service.describedIDs, []string{"f-test"}) {
				t.Fatalf("Describe IDs=%v, want [f-test]", service.describedIDs)
			}
			if service.deleteCalls != 0 || service.refundCalls != 0 || service.waitDeleteCalls != 0 {
				t.Fatalf("identity mismatch reached cloud writes: delete/refund/wait=%d/%d/%d", service.deleteCalls, service.refundCalls, service.waitDeleteCalls)
			}
		})
	}
}

func assertFlinkWorkspaceInstanceStateUnchanged(t *testing.T, before, after *terraform.InstanceState) {
	t.Helper()
	if before == nil || after == nil {
		t.Fatalf("state before/after=%#v/%#v, want both non-nil", before, after)
	}
	if before.ID != after.ID || before.Tainted != after.Tainted || !reflect.DeepEqual(before.Attributes, after.Attributes) || !reflect.DeepEqual(before.Meta, after.Meta) {
		t.Fatalf("Flink workspace state mutated across identity failure:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func assertFlinkWorkspaceIdentityAndProtocolUnchanged(t *testing.T, before, after *terraform.InstanceState) {
	t.Helper()
	if before == nil || after == nil {
		t.Fatalf("state before/after=%#v/%#v, want both non-nil", before, after)
	}
	if before.ID != after.ID || before.Tainted != after.Tainted {
		t.Fatalf("identity/taint changed across Describe identity failure: before=%#v after=%#v", before, after)
	}
	for _, field := range []string{
		"purchase_options_state",
		"capacity_intent_mode",
		"identity_visibility_state",
		"terraform_create_token",
		"create_intent_fingerprint",
	} {
		if got, want := after.Attributes[field], before.Attributes[field]; got != want {
			t.Fatalf("protocol field %s=%q, want prior %q; state=%#v", field, got, want, after.Attributes)
		}
	}
}

func assertFlinkWorkspaceProtocolAttributes(t *testing.T, state *terraform.InstanceState, want map[string]string) {
	t.Helper()
	if state == nil {
		t.Fatal("nil Flink workspace state")
	}
	for field, value := range want {
		if got := state.Attributes[field]; got != value {
			t.Fatalf("%s=%q, want %q; state=%#v", field, got, value, state.Attributes)
		}
	}
}
