package alicloud

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const flinkWorkspaceUnknownConfigValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

func TestFlinkWorkspacePendingAdoptionRejectsEveryNonProtocolDiffBeforeRequiresNew(t *testing.T) {
	token := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		ha     bool
		setup  func(map[string]interface{}, *aliyunFlinkAPI.Workspace)
		mutate func(map[string]interface{})
	}{
		{name: "name", mutate: func(config map[string]interface{}) { config["name"] = "other-workspace" }},
		{name: "resource group", mutate: func(config map[string]interface{}) { config["resource_group_id"] = "rg-other" }},
		{name: "primary zone", mutate: func(config map[string]interface{}) { config["zone_id"] = "cn-test-c" }},
		{name: "VPC", mutate: func(config map[string]interface{}) { config["vpc_id"] = "vpc-other" }},
		{name: "primary vSwitch", mutate: func(config map[string]interface{}) { config["vswitch_ids"] = []interface{}{"vsw-other"} }},
		{
			name: "security group",
			setup: func(config map[string]interface{}, workspace *aliyunFlinkAPI.Workspace) {
				config["security_group_id"] = "sg-1"
				workspace.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{SecurityGroupId: "sg-1"}
			},
			mutate: func(config map[string]interface{}) { config["security_group_id"] = "sg-other" },
		},
		{
			name: "empty security group",
			setup: func(config map[string]interface{}, workspace *aliyunFlinkAPI.Workspace) {
				config["security_group_id"] = "sg-1"
				workspace.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{SecurityGroupId: "sg-1"}
			},
			mutate: func(config map[string]interface{}) { config["security_group_id"] = "" },
		},
		{name: "architecture", mutate: func(config map[string]interface{}) { config["architecture_type"] = "ARM" }},
		{name: "monitor", mutate: func(config map[string]interface{}) { config["monitor_type"] = "TAIHAO" }},
		{name: "storage", mutate: func(config map[string]interface{}) {
			config["storage"] = []interface{}{map[string]interface{}{"oss_bucket": "other-bucket"}}
		}},
		{name: "empty storage", mutate: func(config map[string]interface{}) { config["storage"] = []interface{}{} }},
		{name: "HA vSwitch", ha: true, mutate: func(config map[string]interface{}) {
			ha := firstTestBlock(config["ha"])
			ha["vswitch_ids"] = []interface{}{"vsw-other"}
		}},
		{name: "HA zone", ha: true, mutate: func(config map[string]interface{}) {
			ha := firstTestBlock(config["ha"])
			ha["zone_id"] = "cn-test-c"
		}},
		{name: "unknown name", mutate: func(config map[string]interface{}) { config["name"] = flinkWorkspaceUnknownConfigValue }},
		{name: "unknown nested storage", mutate: func(config map[string]interface{}) {
			config["storage"] = []interface{}{map[string]interface{}{"oss_bucket": flinkWorkspaceUnknownConfigValue}}
		}},
		{name: "unknown nested HA network", ha: true, mutate: func(config map[string]interface{}) {
			ha := firstTestBlock(config["ha"])
			ha["vswitch_ids"] = flinkWorkspaceUnknownConfigValue
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceInitialLifecycleConfig(test.ha)
			workspace := flinkWorkspaceInitialLifecycleFixture(test.ha)
			if test.setup != nil {
				test.setup(config, workspace)
			}
			workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
			test.mutate(config)

			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err == nil || !strings.Contains(err.Error(), "state-only") {
				t.Fatalf("pending adoption diff=%#v error=%v, want state-only fail-closed error", diff, err)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("pending adoption mismatch escaped as replacement: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspacePendingInitialAdoptionRejectsObservableDriftWithoutRecovery(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|initial", workspace)
	config := flinkWorkspaceInitialLifecycleConfig(false)
	config["vpc_id"] = "vpc-other"

	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "state-only") {
		t.Fatalf("initial adoption diff=%#v error=%v, want state-only fail-closed error", diff, err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("initial adoption mismatch escaped as replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceAdoptionApplyRevalidatesEveryObservableFieldWithStateOnlyCoreCalls(t *testing.T) {
	token := strings.Repeat("b", 64)
	tests := []struct {
		name   string
		ha     bool
		setup  func(map[string]interface{}, *aliyunFlinkAPI.Workspace)
		mutate func(*aliyunFlinkAPI.Workspace)
	}{
		{name: "name", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.Name = "other-workspace" }},
		{name: "resource group", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ResourceGroupId = "rg-other" }},
		{name: "primary zone", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ZoneId = "cn-test-c" }},
		{name: "VPC", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.VpcId = "vpc-other" }},
		{name: "primary vSwitch", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.VSwitchIds = []string{"vsw-other"} }},
		{
			name: "security group",
			setup: func(config map[string]interface{}, workspace *aliyunFlinkAPI.Workspace) {
				config["security_group_id"] = "sg-1"
				workspace.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{SecurityGroupId: "sg-1"}
			},
			mutate: func(workspace *aliyunFlinkAPI.Workspace) {
				workspace.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{SecurityGroupId: "sg-other"}
			},
		},
		{name: "architecture", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ArchitectureType = "ARM" }},
		{name: "monitor", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.MonitorType = "TAIHAO" }},
		{name: "storage", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.Storage.Oss.Bucket = "other-bucket" }},
		{name: "missing storage", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.Storage = nil }},
		{name: "HA vSwitch", ha: true, mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.HaVSwitchIds = []string{"vsw-other"} }},
		{name: "HA zone", ha: true, mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.HaZoneId = "cn-test-c" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceInitialLifecycleConfig(test.ha)
			workspace := flinkWorkspaceInitialLifecycleFixture(test.ha)
			if test.setup != nil {
				test.setup(config, workspace)
			}
			workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err != nil || diffRequiresNew(diff) {
				t.Fatalf("safe adoption plan diff=%#v error=%v", diff, err)
			}

			drifted := flinkWorkspaceInitialLifecycleFixture(test.ha)
			if test.setup != nil {
				test.setup(config, drifted)
			}
			drifted.Tags = append([]aliyunFlinkAPI.Tag(nil), workspace.Tags...)
			test.mutate(drifted)
			service := &fakeFlinkWorkspaceAdoptionService{workspace: drifted}
			var creates, updates, deletes int
			resource.Create = func(*schema.ResourceData, interface{}) error { creates++; return nil }
			resource.Update = func(d *schema.ResourceData, _ interface{}) error {
				updates++
				return updateFlinkWorkspaceWithService(d, service)
			}
			resource.Delete = func(*schema.ResourceData, interface{}) error { deletes++; return nil }

			state, err := resource.Apply(prior, diff, nil)
			if err == nil || !strings.Contains(err.Error(), "observable") {
				t.Fatalf("Apply state=%#v error=%v, want observable revalidation failure", state, err)
			}
			if creates != 0 || deletes != 0 || updates != 1 || service.describes != 1 {
				t.Fatalf("Core calls create/update/delete/describe=%d/%d/%d/%d, want 0/1/0/1", creates, updates, deletes, service.describes)
			}
			if state == nil || state.Attributes["purchase_options_state"] != flinkWorkspacePurchaseRecoveryPending || state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityInitialAdoptionPending {
				t.Fatalf("failed state-only adoption lost pending protocol state: %#v", state)
			}
		})
	}
}

type fakeFlinkWorkspaceVisibilityService struct {
	workspace         *aliyunFlinkAPI.Workspace
	describeResponses []*aliyunFlinkAPI.Workspace
	describeErrors    []error
	listResponses     [][]aliyunFlinkAPI.Workspace
	listErrors        []error
	describes         int
	lists             int
}

func (f *fakeFlinkWorkspaceVisibilityService) DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error) {
	call := f.describes
	f.describes++
	if call < len(f.describeErrors) && f.describeErrors[call] != nil {
		return nil, f.describeErrors[call]
	}
	if call < len(f.describeResponses) && f.describeResponses[call] != nil {
		return f.describeResponses[call], nil
	}
	return f.workspace, nil
}

func (f *fakeFlinkWorkspaceVisibilityService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	call := f.lists
	f.lists++
	if call < len(f.listErrors) && f.listErrors[call] != nil {
		return nil, f.listErrors[call]
	}
	if len(f.listResponses) == 0 {
		return nil, nil
	}
	if call >= len(f.listResponses) {
		call = len(f.listResponses) - 1
	}
	return f.listResponses[call], nil
}

func TestFlinkWorkspaceIdentityVisibilitySchemaAndInitialStates(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	field := resource.Schema["identity_visibility_state"]
	if field == nil || !field.Computed || field.Optional || field.Required {
		t.Fatalf("identity_visibility_state schema = %#v, want computed-only", field)
	}

	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	for _, test := range []struct {
		name     string
		importID string
		want     string
	}{
		{name: "bare import is removable", importID: "f-bare", want: "STABLE"},
		{name: "initial import is removable", importID: "f-initial|initial", want: "STABLE"},
		{name: "recovery awaits first authoritative read", importID: "f-recovery|recover=" + strings.Repeat("c", 64) + "|initial", want: "AWAITING_FIRST_READ"},
	} {
		t.Run(test.name, func(t *testing.T) {
			states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, test.importID)
			if err != nil {
				t.Fatal(err)
			}
			if got := resource.Data(states[0]).Get("identity_visibility_state"); got != test.want {
				t.Fatalf("identity_visibility_state = %#v, want %q", got, test.want)
			}
		})
	}

	data := schema.TestResourceDataRaw(t, resource.Schema, flinkWorkspaceInitialLifecycleConfig(false))
	request, options := flinkWorkspaceTestCreateIntent(flinkWorkspaceInitialLifecycleConfig(false), flinkworkspace.CapacityIntentInitial)
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
		t.Fatal(err)
	}
	if got := data.Get("identity_visibility_state"); got != "AWAITING_FIRST_READ" {
		t.Fatalf("fresh Create identity_visibility_state = %#v, want AWAITING_FIRST_READ", got)
	}
}

func TestFlinkWorkspaceRecoveryVisibilitySurvivesDescribeListInterleavingUntilSuccessfulRead(t *testing.T) {
	token := strings.Repeat("d", 64)
	config := flinkWorkspaceInitialLifecycleConfig(false)
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Id = "f-recovery"
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-recovery|recover="+token+"|initial")
	if err != nil {
		t.Fatal(err)
	}
	data := resource.Data(states[0])
	notFound := errors.New("workspace not found")
	service := &fakeFlinkWorkspaceVisibilityService{
		describeErrors:    []error{notFound, notFound, nil},
		describeResponses: []*aliyunFlinkAPI.Workspace{nil, nil, workspace},
		listResponses:     [][]aliyunFlinkAPI.Workspace{nil, {*workspace}},
	}

	for attempt := 1; attempt <= 2; attempt++ {
		err := readFlinkWorkspaceWithService(data, service)
		if err == nil || !strings.Contains(err.Error(), "retained") {
			t.Fatalf("attempt %d error = %v, want retryable fail-closed retained-ID error", attempt, err)
		}
		if data.Id() != "f-recovery" || data.Get("identity_visibility_state") != "AWAITING_FIRST_READ" {
			t.Fatalf("attempt %d lost protected identity: id=%q visibility=%#v", attempt, data.Id(), data.Get("identity_visibility_state"))
		}
	}
	if service.describes != 2 || service.lists != 2 {
		t.Fatalf("pre-visibility describe/list calls = %d/%d, want 2/2", service.describes, service.lists)
	}
	if err := readFlinkWorkspaceWithService(data, service); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "f-recovery" || data.Get("identity_visibility_state") != "STABLE" {
		t.Fatalf("successful Read did not stabilize identity: id=%q visibility=%#v", data.Id(), data.Get("identity_visibility_state"))
	}
	if service.describes != 3 || service.lists != 2 {
		t.Fatalf("total describe/list calls = %d/%d, want 3/2", service.describes, service.lists)
	}
}

func TestFlinkWorkspaceV0InitialMigrationSurvivesDescribeListInterleavingUntilTokenVerifiedRead(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "HA"}[ha], func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			legacy := initialWorkspaceState(ha)
			legacy.Meta = map[string]interface{}{"schema_version": "0"}
			meta := &connectivity.AliyunClient{RegionId: "cn-test"}
			token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: meta.RegionId, Name: "workspace"})
			workspace := flinkWorkspaceInitialLifecycleFixture(ha)
			workspace.Id = legacy.ID
			workspace.Region = meta.RegionId
			workspace.Tags = []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}}
			notFound := errors.New("workspace not found")
			service := &fakeFlinkWorkspaceVisibilityService{
				describeErrors:    []error{notFound, notFound, nil},
				describeResponses: []*aliyunFlinkAPI.Workspace{nil, nil, workspace},
				listResponses:     [][]aliyunFlinkAPI.Workspace{nil, {*workspace}},
			}
			resource.Read = func(d *schema.ResourceData, _ interface{}) error {
				return readFlinkWorkspaceWithService(d, service)
			}

			state := legacy
			for attempt := 1; attempt <= 2; attempt++ {
				refreshed, err := resource.Refresh(state, meta)
				if err == nil || !strings.Contains(err.Error(), "retained") {
					t.Fatalf("attempt %d error=%v, want fail-closed retained-ID error", attempt, err)
				}
				if refreshed == nil || refreshed.ID != legacy.ID {
					t.Fatalf("attempt %d lost migrated paid identity: %#v", attempt, refreshed)
				}
				for field, want := range map[string]string{
					"purchase_options_state":    "MIGRATED_INITIAL_PENDING",
					"capacity_intent_mode":      "INITIAL",
					"identity_visibility_state": "MIGRATED_AWAITING_FIRST_READ",
					"terraform_create_token":    token,
					"create_intent_fingerprint": "UNAVAILABLE",
				} {
					if got := refreshed.Attributes[field]; got != want {
						t.Fatalf("attempt %d %s=%q, want %q; state=%#v", attempt, field, got, want, refreshed.Attributes)
					}
				}
				state = refreshed
			}
			if service.describes != 2 || service.lists != 2 {
				t.Fatalf("pre-visibility describe/list calls=%d/%d, want 2/2", service.describes, service.lists)
			}

			stable, err := resource.Refresh(state, meta)
			if err != nil {
				t.Fatal(err)
			}
			if stable == nil || stable.ID != legacy.ID {
				t.Fatalf("successful Read lost migrated identity: %#v", stable)
			}
			for field, want := range map[string]string{
				"purchase_options_state":    "MANAGED",
				"capacity_intent_mode":      "INITIAL",
				"identity_visibility_state": "STABLE",
				"terraform_create_token":    "UNAVAILABLE",
				"create_intent_fingerprint": "UNAVAILABLE",
			} {
				if got := stable.Attributes[field]; got != want {
					t.Fatalf("successful Read %s=%q, want %q; state=%#v", field, got, want, stable.Attributes)
				}
			}
			if service.describes != 3 || service.lists != 2 {
				t.Fatalf("total describe/list calls=%d/%d, want 3/2", service.describes, service.lists)
			}

			var creates, updates, deletes, refunds int
			resource.Create = func(*schema.ResourceData, interface{}) error { creates++; return nil }
			resource.Update = func(*schema.ResourceData, interface{}) error { updates++; return nil }
			resource.Delete = func(d *schema.ResourceData, _ interface{}) error {
				if d.Get("charge_type").(string) == "PRE" {
					refunds++
				} else {
					deletes++
				}
				return nil
			}
			emptyDiff := terraform.NewInstanceDiff()
			emptyDiff.Attributes = map[string]*terraform.ResourceAttrDiff{}
			if _, err := resource.Apply(stable, emptyDiff, nil); err != nil {
				t.Fatal(err)
			}
			if creates != 0 || deletes != 0 || refunds != 0 || updates != 1 {
				t.Fatalf("post-migration Apply calls create/update/delete/refund=%d/%d/%d/%d, want 0/1/0/0", creates, updates, deletes, refunds)
			}
		})
	}
}

func TestFlinkWorkspaceV0InitialMigrationProtocolHasOneExactFailClosedTuple(t *testing.T) {
	token := strings.Repeat("a", 64)
	valid := flinkWorkspaceProtocolTuple{
		purchase:    "MIGRATED_INITIAL_PENDING",
		capacity:    "INITIAL",
		visibility:  "MIGRATED_AWAITING_FIRST_READ",
		token:       token,
		fingerprint: "UNAVAILABLE",
	}
	if err := validateFlinkWorkspaceProtocolTuple("f-migrated", valid, flinkWorkspaceProtocolExisting); err != nil {
		t.Fatalf("exact migration-first-read tuple rejected: %v", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*flinkWorkspaceProtocolTuple)
	}{
		{name: "managed purchase masquerade", mutate: func(tuple *flinkWorkspaceProtocolTuple) { tuple.purchase = "MANAGED" }},
		{name: "legacy capacity masquerade", mutate: func(tuple *flinkWorkspaceProtocolTuple) { tuple.capacity = "LEGACY" }},
		{name: "ordinary awaiting visibility masquerade", mutate: func(tuple *flinkWorkspaceProtocolTuple) { tuple.visibility = "AWAITING_FIRST_READ" }},
		{name: "missing migration token", mutate: func(tuple *flinkWorkspaceProtocolTuple) { tuple.token = "UNAVAILABLE" }},
		{name: "invented v1 fingerprint", mutate: func(tuple *flinkWorkspaceProtocolTuple) { tuple.fingerprint = strings.Repeat("b", 64) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			tuple := valid
			test.mutate(&tuple)
			if err := validateFlinkWorkspaceProtocolTuple("f-migrated", tuple, flinkWorkspaceProtocolExisting); err == nil {
				t.Fatalf("near-miss tuple accepted: %#v", tuple)
			}
		})
	}

	resource := resourceAliCloudFlinkWorkspace()
	data := schema.TestResourceDataRaw(t, resource.Schema, flinkWorkspaceInitialLifecycleConfig(false))
	data.SetId("f-migrated")
	for field, value := range map[string]string{
		"purchase_options_state":    valid.purchase,
		"capacity_intent_mode":      valid.capacity,
		"identity_visibility_state": valid.visibility,
		"terraform_create_token":    valid.token,
		"create_intent_fingerprint": valid.fingerprint,
	} {
		if err := data.Set(field, value); err != nil {
			t.Fatal(err)
		}
	}
	diff, err := resource.Diff(data.State(), terraform.NewResourceConfigRaw(flinkWorkspaceInitialLifecycleConfig(false)), nil)
	if err == nil || !strings.Contains(err.Error(), "authoritative Read") {
		t.Fatalf("migration-first-read diff=%#v error=%v, want fail-closed refresh requirement", diff, err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("migration-first-read protocol planned replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceV0InitialMigrationFailureMatrixAlwaysRetainsPaidIdentity(t *testing.T) {
	meta := &connectivity.AliyunClient{RegionId: "cn-test"}
	token := flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: meta.RegionId, Name: "workspace"})
	baseWorkspace := flinkWorkspaceInitialLifecycleFixture(false)
	baseWorkspace.Id = "f-test"
	baseWorkspace.Region = meta.RegionId
	baseWorkspace.Tags = []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}}
	notFound := errors.New("workspace not found")
	listFailure := errors.New("ListInstances unavailable")

	otherIdentity := *baseWorkspace
	otherIdentity.Id = "f-other"
	missingTag := *baseWorkspace
	missingTag.Tags = nil
	wrongTag := *baseWorkspace
	wrongTag.Tags = []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: strings.Repeat("f", 64)}}
	duplicateTag := *baseWorkspace
	duplicateTag.Tags = []aliyunFlinkAPI.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: token},
		{Key: flinkworkspace.CreateTokenTagKey, Value: token},
	}

	for _, test := range []struct {
		name      string
		service   *fakeFlinkWorkspaceVisibilityService
		wantError string
		wantLists int
	}{
		{name: "canceled Describe", service: &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{context.Canceled}}, wantError: context.Canceled.Error()},
		{name: "NotFound then List error", service: &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{notFound}, listErrors: []error{listFailure}}, wantError: listFailure.Error(), wantLists: 1},
		{name: "NotFound then token on another ID", service: &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{notFound}, listResponses: [][]aliyunFlinkAPI.Workspace{{otherIdentity}}}, wantError: "refusing automatic identity transfer", wantLists: 1},
		{name: "successful ID Read missing token", service: &fakeFlinkWorkspaceVisibilityService{workspace: &missingTag}, wantError: "cannot verify its create token"},
		{name: "successful ID Read wrong token", service: &fakeFlinkWorkspaceVisibilityService{workspace: &wrongTag}, wantError: "cannot verify its create token"},
		{name: "successful ID Read duplicate token", service: &fakeFlinkWorkspaceVisibilityService{workspace: &duplicateTag}, wantError: "cannot verify its create token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			resource.Read = func(d *schema.ResourceData, _ interface{}) error {
				return readFlinkWorkspaceWithService(d, test.service)
			}
			legacy := initialWorkspaceState(false)
			legacy.Meta = map[string]interface{}{"schema_version": "0"}
			refreshed, err := resource.Refresh(legacy, meta)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Refresh state=%#v error=%v, want %q", refreshed, err, test.wantError)
			}
			if refreshed == nil || refreshed.ID != legacy.ID {
				t.Fatalf("failure lost paid identity: %#v", refreshed)
			}
			for field, want := range map[string]string{
				"purchase_options_state":    "MIGRATED_INITIAL_PENDING",
				"capacity_intent_mode":      "INITIAL",
				"identity_visibility_state": "MIGRATED_AWAITING_FIRST_READ",
				"terraform_create_token":    token,
				"create_intent_fingerprint": "UNAVAILABLE",
			} {
				if got := refreshed.Attributes[field]; got != want {
					t.Fatalf("failure %s=%q, want %q; state=%#v", field, got, want, refreshed.Attributes)
				}
			}
			if test.service.describes != 1 || test.service.lists != test.wantLists {
				t.Fatalf("describe/list calls=%d/%d, want 1/%d", test.service.describes, test.service.lists, test.wantLists)
			}
		})
	}
}

func TestFlinkWorkspaceFreshManagedIdentitySurvivesNotFoundCancelAndRetry(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, resource.Schema, config)
	request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentInitial)
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
		t.Fatal(err)
	}
	data.SetId("f-new-paid")
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Id = data.Id()
	workspace.Tags = []aliyunFlinkAPI.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: data.Get("terraform_create_token").(string)},
		{Key: flinkworkspace.CreateIntentTagKey, Value: data.Get("create_intent_fingerprint").(string)},
	}
	notFound := errors.New("workspace not found")
	service := &fakeFlinkWorkspaceVisibilityService{
		describeErrors:    []error{notFound, context.Canceled, nil},
		describeResponses: []*aliyunFlinkAPI.Workspace{nil, nil, workspace},
		listResponses:     [][]aliyunFlinkAPI.Workspace{nil},
	}

	if err := readFlinkWorkspaceWithService(data, service); err == nil || data.Id() != "f-new-paid" {
		t.Fatalf("NotFound error=%v id=%q, want retained paid identity", err, data.Id())
	}
	if err := readFlinkWorkspaceWithService(data, service); err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) || data.Id() != "f-new-paid" {
		t.Fatalf("cancel error=%v id=%q, want retained paid identity", err, data.Id())
	}
	if err := readFlinkWorkspaceWithService(data, service); err != nil {
		t.Fatal(err)
	}
	if got := data.Get("identity_visibility_state"); got != "STABLE" {
		t.Fatalf("successful managed Read visibility = %#v, want STABLE", got)
	}
}

func TestFlinkWorkspaceLegacyCreateCompletesNormallyOnDeferredSuccessfulRefresh(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, resource.Schema, config)
	request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentLegacy)
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentLegacy); err != nil {
		t.Fatal(err)
	}
	if err := completeFlinkWorkspaceCreate(data, nil, nil, "f-legacy", false, 0, nil); err != nil {
		t.Fatal(err)
	}
	workspace := flinkWorkspaceLifecycleFixture(false)
	workspace.Id = data.Id()
	workspace.Tags = []aliyunFlinkAPI.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: data.Get("terraform_create_token").(string)},
		{Key: flinkworkspace.CreateIntentTagKey, Value: data.Get("create_intent_fingerprint").(string)},
	}
	service := &fakeFlinkWorkspaceVisibilityService{workspace: workspace}
	if err := readFlinkWorkspaceWithService(data, service); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "f-legacy" || data.Get("identity_visibility_state") != flinkWorkspaceIdentityStable {
		t.Fatalf("deferred legacy refresh did not stabilize paid identity: id=%q visibility=%#v", data.Id(), data.Get("identity_visibility_state"))
	}
	if service.describes != 1 || service.lists != 0 {
		t.Fatalf("legacy deferred refresh describe/list calls=%d/%d, want 1/0", service.describes, service.lists)
	}
	if !flinkListBlockConfigured(data.Get("resource")) {
		t.Fatalf("legacy deferred refresh did not materialize API-observable resource: %#v", data.Get("resource"))
	}
}

func TestFlinkWorkspaceStableManagedAndBareImportNotFoundKeepRemoveSemantics(t *testing.T) {
	notFound := errors.New("workspace not found")

	t.Run("bare import", func(t *testing.T) {
		resource := resourceAliCloudFlinkWorkspace()
		provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
		states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-bare")
		if err != nil {
			t.Fatal(err)
		}
		data := resource.Data(states[0])
		if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{notFound}}); err != nil {
			t.Fatal(err)
		}
		if data.Id() != "" {
			t.Fatalf("bare import retained deleted ID %q", data.Id())
		}
	})

	t.Run("stable managed", func(t *testing.T) {
		resource := resourceAliCloudFlinkWorkspace()
		config := flinkWorkspaceInitialLifecycleConfig(false)
		data := schema.TestResourceDataRaw(t, resource.Schema, config)
		request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentInitial)
		if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
			t.Fatal(err)
		}
		data.SetId("f-stable")
		if err := data.Set("identity_visibility_state", "STABLE"); err != nil {
			t.Fatal(err)
		}
		if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{notFound}}); err != nil {
			t.Fatal(err)
		}
		if data.Id() != "" {
			t.Fatalf("stable managed resource retained deleted ID %q", data.Id())
		}
	})
}

func TestFlinkWorkspaceFailedInitialCreateNotFoundCannotReachSecondCoreCreate(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentInitial)
	var creates, updates int
	resource.Create = func(d *schema.ResourceData, _ interface{}) error {
		creates++
		if err := setFlinkWorkspaceCreateProtocolState(d, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
			return err
		}
		d.SetId("f-paid-but-not-visible")
		return errors.New("response lost after paid Create")
	}
	resource.Update = func(*schema.ResourceData, interface{}) error { updates++; return nil }

	diff, err := resource.Diff(nil, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	failedState, err := resource.Apply(nil, diff, nil)
	if err == nil || failedState == nil || failedState.ID != "f-paid-but-not-visible" {
		t.Fatalf("first Apply state=%#v error=%v", failedState, err)
	}
	data := resource.Data(failedState)
	if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceVisibilityService{describeErrors: []error{errors.New("workspace not found")}}); err == nil {
		t.Fatal("post-create NotFound did not fail closed")
	}
	retained := data.State()
	if retained == nil || retained.ID != "f-paid-but-not-visible" {
		t.Fatalf("NotFound lost failed-create state: %#v", retained)
	}
	emptyDiff := terraform.NewInstanceDiff()
	emptyDiff.Attributes = map[string]*terraform.ResourceAttrDiff{}
	if _, err := resource.Apply(retained, emptyDiff, nil); err != nil {
		t.Fatal(err)
	}
	if creates != 1 || updates != 1 {
		t.Fatalf("Core calls create/update=%d/%d, want exactly one original Create and state-preserving Update", creates, updates)
	}
}

func TestFlinkWorkspaceProtocolTupleRejectsEveryPartialInvalidOrContradictoryV1StateWithoutRefresh(t *testing.T) {
	baseResource := resourceAliCloudFlinkWorkspace()
	baseConfig := flinkWorkspaceInitialLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, baseResource.Schema, baseConfig)
	request, options := flinkWorkspaceTestCreateIntent(baseConfig, flinkworkspace.CapacityIntentInitial)
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
		t.Fatal(err)
	}
	data.SetId("f-managed-stable")
	if err := data.Set("identity_visibility_state", "STABLE"); err != nil {
		t.Fatal(err)
	}
	valid := data.State()

	tests := []struct {
		name   string
		mutate func(map[string]string)
	}{
		{name: "missing purchase", mutate: func(a map[string]string) { delete(a, "purchase_options_state") }},
		{name: "empty purchase", mutate: func(a map[string]string) { a["purchase_options_state"] = "" }},
		{name: "invalid purchase", mutate: func(a map[string]string) { a["purchase_options_state"] = "CORRUPT" }},
		{name: "missing capacity", mutate: func(a map[string]string) { delete(a, "capacity_intent_mode") }},
		{name: "empty capacity", mutate: func(a map[string]string) { a["capacity_intent_mode"] = "" }},
		{name: "invalid capacity", mutate: func(a map[string]string) { a["capacity_intent_mode"] = "CORRUPT" }},
		{name: "missing visibility", mutate: func(a map[string]string) { delete(a, "identity_visibility_state") }},
		{name: "empty visibility", mutate: func(a map[string]string) { a["identity_visibility_state"] = "" }},
		{name: "invalid visibility", mutate: func(a map[string]string) { a["identity_visibility_state"] = "CORRUPT" }},
		{name: "missing token", mutate: func(a map[string]string) { delete(a, "terraform_create_token") }},
		{name: "empty token", mutate: func(a map[string]string) { a["terraform_create_token"] = "" }},
		{name: "invalid token", mutate: func(a map[string]string) { a["terraform_create_token"] = "short" }},
		{name: "missing fingerprint", mutate: func(a map[string]string) { delete(a, "create_intent_fingerprint") }},
		{name: "empty fingerprint", mutate: func(a map[string]string) { a["create_intent_fingerprint"] = "" }},
		{name: "invalid fingerprint", mutate: func(a map[string]string) { a["create_intent_fingerprint"] = "short" }},
		{name: "managed pending capacity", mutate: func(a map[string]string) { a["capacity_intent_mode"] = "INITIAL_ADOPTION_PENDING" }},
		{name: "managed awaiting without paid identity", mutate: func(a map[string]string) {
			a["identity_visibility_state"] = "AWAITING_FIRST_READ"
			a["terraform_create_token"] = "UNAVAILABLE"
			a["create_intent_fingerprint"] = "UNAVAILABLE"
		}},
		{name: "imported unknown claims managed identity", mutate: func(a map[string]string) {
			a["purchase_options_state"] = "IMPORTED_UNKNOWN"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := valid.DeepCopy()
			state.Meta = map[string]interface{}{"schema_version": "1"}
			test.mutate(state.Attributes)
			config := cloneFlinkWorkspaceLifecycleConfig(baseConfig)
			config["duration"] = 3
			resource := resourceAliCloudFlinkWorkspace()
			diff, err := resource.Diff(state, terraform.NewResourceConfigRaw(config), nil)
			if err == nil || !strings.Contains(err.Error(), "protocol") {
				t.Fatalf("partial v1 Diff=%#v error=%v, want protocol fail-closed error", diff, err)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("partial v1 state produced replacement before validation: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceInvalidImportedUnknownTupleCannotBeDiffSuppressedIntoSilentUpdate(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported")
	if err != nil {
		t.Fatal(err)
	}
	state := states[0].DeepCopy()
	delete(state.Attributes, "capacity_intent_mode")
	state.Meta = map[string]interface{}{"schema_version": "1"}
	config := flinkWorkspaceLifecycleConfig(false)
	config["duration"] = 3

	diff, err := resource.Diff(state, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("invalid imported tuple Diff=%#v error=%v, want protocol failure", diff, err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("invalid imported tuple produced paid replacement: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceProtocolTupleIsValidatedBeforeReadAndUpdateCloudCalls(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, resource.Schema, config)
	request, options := flinkWorkspaceTestCreateIntent(config, flinkworkspace.CapacityIntentInitial)
	if err := setFlinkWorkspaceCreateProtocolState(data, request, options, flinkworkspace.CapacityIntentInitial); err != nil {
		t.Fatal(err)
	}
	data.SetId("f-corrupt")
	if err := data.Set("identity_visibility_state", "STABLE"); err != nil {
		t.Fatal(err)
	}
	if err := data.Set("purchase_options_state", "CORRUPT"); err != nil {
		t.Fatal(err)
	}
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	service := &fakeFlinkWorkspaceVisibilityService{workspace: workspace}

	if err := readFlinkWorkspaceWithService(data, service); err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("Read error=%v, want protocol failure", err)
	}
	if service.describes != 0 || service.lists != 0 {
		t.Fatalf("invalid Read tuple reached cloud calls describe/list=%d/%d", service.describes, service.lists)
	}
	if err := updateFlinkWorkspaceWithService(data, service); err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("Update error=%v, want protocol failure", err)
	}
	if service.describes != 0 || service.lists != 0 {
		t.Fatalf("invalid Update tuple reached cloud calls describe/list=%d/%d", service.describes, service.lists)
	}
}

func TestFlinkWorkspaceCreateEntryRejectsCorruptProtocolBeforeClientAccess(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	data := schema.TestResourceDataRaw(t, resource.Schema, flinkWorkspaceInitialLifecycleConfig(false))
	if err := data.Set("purchase_options_state", "CORRUPT"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Create reached client access before protocol validation: %v", recovered)
		}
	}()
	if err := resourceAliCloudFlinkWorkspaceCreate(data, nil); err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("Create error=%v, want protocol failure", err)
	}
}

func TestFlinkWorkspaceFailedApplyRejectsCorruptProtocolTransitionBeforeDescribeOrWrites(t *testing.T) {
	token := strings.Repeat("e", 64)
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if marker := diff.Attributes["purchase_options_state"]; marker == nil {
		t.Fatal("recovery plan has no purchase marker transition")
	} else {
		marker.New = "CORRUPT"
	}
	// Keep the production Update entrypoint. nil meta is an intentionally
	// unreachable client factory: invalid planned protocol must be rolled back
	// before the entrypoint can type-assert or construct any cloud client.
	state, err := resource.Apply(prior, diff, nil)
	if err == nil || !strings.Contains(err.Error(), "protocol") {
		t.Fatalf("Apply state=%#v error=%v, want protocol failure", state, err)
	}
	if state == nil || state.Tainted {
		t.Fatalf("failed protocol Apply corrupted or tainted pending state: %#v", state)
	}
	for field, want := range map[string]string{
		"purchase_options_state":    flinkWorkspacePurchaseRecoveryPending,
		"capacity_intent_mode":      flinkWorkspaceCapacityInitialAdoptionPending,
		"identity_visibility_state": flinkWorkspaceIdentityStable,
		"terraform_create_token":    token,
		"create_intent_fingerprint": workspace.Tags[1].Value,
	} {
		if got := state.Attributes[field]; got != want {
			t.Fatalf("failed production Update %s=%q, want rolled-back %q; state=%#v", field, got, want, state.Attributes)
		}
	}
}

func TestFlinkWorkspaceProductionUpdateClientFailureRollsBackValidPlannedTransition(t *testing.T) {
	token := strings.Repeat("d", 64)
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}

	state, err := resource.Apply(prior, diff, nil)
	if err == nil || !strings.Contains(err.Error(), "configured Aliyun client") {
		t.Fatalf("Apply state=%#v error=%v, want client failure", state, err)
	}
	if state == nil || state.Tainted {
		t.Fatalf("client failure corrupted or tainted pending state: %#v", state)
	}
	for field, want := range map[string]string{
		"purchase_options_state":    flinkWorkspacePurchaseRecoveryPending,
		"capacity_intent_mode":      flinkWorkspaceCapacityInitialAdoptionPending,
		"identity_visibility_state": flinkWorkspaceIdentityStable,
		"terraform_create_token":    token,
		"create_intent_fingerprint": workspace.Tags[1].Value,
	} {
		if got := state.Attributes[field]; got != want {
			t.Fatalf("client failure %s=%q, want rolled-back %q; state=%#v", field, got, want, state.Attributes)
		}
	}
}

func TestFlinkWorkspaceImportAndV0UpgradeUseExplicitUnavailableProtocolValues(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"terraform_create_token", "create_intent_fingerprint"} {
		if got := resource.Data(states[0]).Get(field); got != "UNAVAILABLE" {
			t.Fatalf("bare import %s = %#v, want UNAVAILABLE", field, got)
		}
	}

	legacy := legacyWorkspaceState()
	legacy.Meta = map[string]interface{}{"schema_version": "0"}
	resource.Read = schema.Noop
	upgraded, err := resource.Refresh(legacy, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"terraform_create_token", "create_intent_fingerprint"} {
		if got := upgraded.Attributes[field]; got != "UNAVAILABLE" {
			t.Fatalf("v0 upgrade %s = %q, want UNAVAILABLE; state=%#v", field, got, upgraded.Attributes)
		}
	}
}
