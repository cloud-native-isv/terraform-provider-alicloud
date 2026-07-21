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

type fakeFlinkWorkspaceReadService struct {
	workspace *aliyunFlinkAPI.Workspace
	errors    []error
	calls     int
}

func (f *fakeFlinkWorkspaceReadService) DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error) {
	call := f.calls
	f.calls++
	if call < len(f.errors) && f.errors[call] != nil {
		return nil, f.errors[call]
	}
	return f.workspace, nil
}

func (f *fakeFlinkWorkspaceReadService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	return nil, nil
}

func TestFlinkWorkspaceImportGrammarPersistsProvenanceBeforeFirstRead(t *testing.T) {
	token := strings.Repeat("a", 64)
	tests := []struct {
		name         string
		importID     string
		wantID       string
		wantPurchase string
		wantCapacity string
		wantToken    string
	}{
		{name: "bare legacy", importID: "f-bare", wantID: "f-bare", wantPurchase: "IMPORTED_UNKNOWN", wantCapacity: "LEGACY", wantToken: flinkWorkspaceProtocolUnavailable},
		{name: "explicit legacy", importID: "f-legacy|legacy", wantID: "f-legacy", wantPurchase: "IMPORTED_UNKNOWN", wantCapacity: "LEGACY", wantToken: flinkWorkspaceProtocolUnavailable},
		{name: "explicit initial", importID: "f-initial|initial", wantID: "f-initial", wantPurchase: "IMPORTED_UNKNOWN", wantCapacity: "INITIAL_ADOPTION_PENDING", wantToken: flinkWorkspaceProtocolUnavailable},
		{name: "recover legacy", importID: "f-recover-legacy|recover=" + token + "|legacy", wantID: "f-recover-legacy", wantPurchase: "RECOVERY_PENDING", wantCapacity: "LEGACY", wantToken: token},
		{name: "recover initial", importID: "f-recover-initial|recover=" + token + "|initial", wantID: "f-recover-initial", wantPurchase: "RECOVERY_PENDING", wantCapacity: "INITIAL_ADOPTION_PENDING", wantToken: token},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
			states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, test.importID)
			if err != nil {
				t.Fatal(err)
			}
			if len(states) != 1 || states[0].ID != test.wantID {
				t.Fatalf("imported state = %#v, want one state with ID %q", states, test.wantID)
			}
			data := resource.Data(states[0])
			if got := data.Get("purchase_options_state"); got != test.wantPurchase {
				t.Fatalf("purchase_options_state = %#v, want %q", got, test.wantPurchase)
			}
			if got := data.Get("capacity_intent_mode"); got != test.wantCapacity {
				t.Fatalf("capacity_intent_mode = %#v, want %q", got, test.wantCapacity)
			}
			if got := data.Get("terraform_create_token"); got != test.wantToken {
				t.Fatalf("terraform_create_token = %#v, want %q", got, test.wantToken)
			}
		})
	}
}

func TestFlinkWorkspaceImportGrammarRejectsAmbiguousOrMalformedIDs(t *testing.T) {
	token := strings.Repeat("a", 64)
	for _, importID := range []string{
		"",
		"|legacy",
		"f-test|",
		"f-test|unknown",
		"f-test|legacy|legacy",
		"f-test|legacy|initial",
		"f-test|recover=" + token,
		"f-test|recover=short|legacy",
		"f-test|recover=" + token + "|recover=" + token + "|legacy",
		"f-test|recover=" + token + "|legacy|unknown",
	} {
		t.Run(strings.ReplaceAll(importID, "|", "_"), func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
			if states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, importID); err == nil {
				t.Fatalf("ImportState(%q) succeeded with states %#v", importID, states)
			}
		})
	}
}

func TestFlinkWorkspaceProvenanceSchemaIsComputedOnly(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	for _, name := range []string{"identity_visibility_state", "purchase_options_state", "capacity_intent_mode", "terraform_create_token", "create_intent_fingerprint"} {
		field := resource.Schema[name]
		if field == nil || !field.Computed || field.Optional || field.Required {
			t.Fatalf("%s schema = %#v, want computed-only", name, field)
		}
	}
	if resource.SchemaVersion != 1 || len(resource.StateUpgraders) != 1 || resource.StateUpgraders[0].Version != 0 {
		t.Fatalf("schema version/upgraders = %d/%#v, want v1 with one v0 upgrader", resource.SchemaVersion, resource.StateUpgraders)
	}
	for _, phrase := range []string{"|recover=<64-hex-create-token>|initial", "imported unknown", "state-only"} {
		if !strings.Contains(resource.Description, phrase) {
			t.Fatalf("resource Description does not document %q: %q", phrase, resource.Description)
		}
	}
}

func TestFlinkWorkspaceProviderInternalValidate(t *testing.T) {
	provider, ok := Provider().(*schema.Provider)
	if !ok {
		t.Fatalf("Provider() type = %T, want *schema.Provider", Provider())
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatal(err)
	}
}

func TestFlinkWorkspaceV0StateUpgraderUsesRealFlatmapAndCtyPath(t *testing.T) {
	tests := []struct {
		name           string
		state          *terraform.InstanceState
		wantPurchase   string
		wantCapacity   string
		wantVisibility string
		wantToken      string
	}{
		{name: "managed initial single", state: initialWorkspaceState(false), wantPurchase: "MIGRATED_INITIAL_PENDING", wantCapacity: "INITIAL", wantVisibility: "MIGRATED_AWAITING_FIRST_READ", wantToken: flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: "cn-test", Name: "workspace"})},
		{name: "managed initial HA", state: initialWorkspaceState(true), wantPurchase: "MIGRATED_INITIAL_PENDING", wantCapacity: "INITIAL", wantVisibility: "MIGRATED_AWAITING_FIRST_READ", wantToken: flinkworkspace.WorkspaceCreateToken(&aliyunFlinkAPI.Workspace{Region: "cn-test", Name: "workspace"})},
		{name: "default-looking legacy is unclassified", state: legacyWorkspaceState(), wantPurchase: "LEGACY_UNCLASSIFIED", wantCapacity: "LEGACY", wantVisibility: "STABLE", wantToken: "UNAVAILABLE"},
		{name: "pending legacy remains known", state: func() *terraform.InstanceState {
			state := legacyWorkspaceState()
			state.ID = pendingFlinkWorkspaceCreateID(strings.Repeat("b", 64))
			state.Attributes["auto_renew"] = "false"
			return state
		}(), wantPurchase: "MANAGED", wantCapacity: "LEGACY", wantVisibility: "AWAITING_FIRST_READ", wantToken: strings.Repeat("b", 64)},
		{name: "pending initial remains known", state: func() *terraform.InstanceState {
			state := initialWorkspaceState(true)
			state.ID = pendingFlinkWorkspaceCreateID(strings.Repeat("c", 64))
			return state
		}(), wantPurchase: "MANAGED", wantCapacity: "INITIAL", wantVisibility: "AWAITING_FIRST_READ", wantToken: strings.Repeat("c", 64)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			resource.Read = schema.Noop
			test.state.Meta = map[string]interface{}{"schema_version": "0"}
			refreshed, err := resource.Refresh(test.state, &connectivity.AliyunClient{RegionId: "cn-test"})
			if err != nil {
				t.Fatal(err)
			}
			if got := refreshed.Attributes["purchase_options_state"]; got != test.wantPurchase {
				t.Fatalf("purchase_options_state = %q, want %q; state=%#v", got, test.wantPurchase, refreshed.Attributes)
			}
			if got := refreshed.Attributes["capacity_intent_mode"]; got != test.wantCapacity {
				t.Fatalf("capacity_intent_mode = %q, want %q; state=%#v", got, test.wantCapacity, refreshed.Attributes)
			}
			if got := refreshed.Attributes["identity_visibility_state"]; got != test.wantVisibility {
				t.Fatalf("identity_visibility_state = %q, want %q; state=%#v", got, test.wantVisibility, refreshed.Attributes)
			}
			if got := refreshed.Attributes["terraform_create_token"]; got != test.wantToken {
				t.Fatalf("terraform_create_token = %q, want %q; state=%#v", got, test.wantToken, refreshed.Attributes)
			}
			if got := refreshed.Attributes["create_intent_fingerprint"]; got != flinkWorkspaceProtocolUnavailable {
				t.Fatalf("create_intent_fingerprint = %q, want UNAVAILABLE; state=%#v", got, refreshed.Attributes)
			}
			if got := refreshed.Meta["schema_version"]; got != "1" {
				t.Fatalf("schema_version = %#v, want 1", got)
			}
			if test.state.Attributes["auto_renew"] == "false" && refreshed.Attributes["auto_renew"] != "false" {
				t.Fatalf("bool false was not preserved: %#v", refreshed.Attributes)
			}
			if test.wantCapacity == "INITIAL" && refreshed.Attributes["initial_capacity.#"] != "1" {
				t.Fatalf("nested initial_capacity was not preserved through cty/flatmap: %#v", refreshed.Attributes)
			}
			if strings.Contains(test.name, "HA") && refreshed.Attributes["ha.0.vswitch_ids.0"] != "vsw-b" {
				t.Fatalf("nested HA flatmap was not preserved: %#v", refreshed.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceV0LegacyConsumerConfigsRemainNoOpAfterMigration(t *testing.T) {
	for _, test := range []struct {
		name   string
		config map[string]interface{}
	}{
		{name: "cws-lib legacy defaults", config: flinkWorkspaceLifecycleConfig(false)},
		{name: "xuanji explicit purchase options", config: func() map[string]interface{} {
			config := flinkWorkspaceLifecycleConfig(true)
			config["auto_renew"] = false
			config["duration"] = 3
			config["pricing_cycle"] = "Month"
			config["extra"] = "consumer-extra"
			return config
		}()},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			data := schema.TestResourceDataRaw(t, resource.Schema, test.config)
			data.SetId("f-v0")
			state := data.State()
			delete(state.Attributes, "purchase_options_state")
			delete(state.Attributes, "capacity_intent_mode")
			delete(state.Attributes, "terraform_create_token")
			delete(state.Attributes, "create_intent_fingerprint")
			state.Meta = map[string]interface{}{"schema_version": "0"}
			resource.Read = schema.Noop
			migrated, err := resource.Refresh(state, nil)
			if err != nil {
				t.Fatal(err)
			}
			if migrated.Attributes["purchase_options_state"] != flinkWorkspacePurchaseLegacyUnclassified {
				t.Fatalf("migrated state = %#v", migrated.Attributes)
			}
			diff, err := resource.Diff(migrated, terraform.NewResourceConfigRaw(test.config), nil)
			if err != nil || diffRequiresNew(diff) {
				t.Fatalf("legacy consumer post-migration diff=%#v error=%v", diff, err)
			}
		})
	}
}

func TestFlinkWorkspaceBareImportNeverInventsOrReplacesInputOnlyPurchaseOptions(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{name: "one month defaults", config: flinkWorkspaceLifecycleConfig(false)},
		{name: "three months auto renew false extra and promotion", config: func() map[string]interface{} {
			config := flinkWorkspaceLifecycleConfig(false)
			config["auto_renew"] = false
			config["duration"] = 3
			config["extra"] = "opaque-extra"
			config["promotion_code"] = "promotion"
			config["use_promotion_code"] = true
			return config
		}()},
		{name: "one year", config: func() map[string]interface{} {
			config := flinkWorkspaceLifecycleConfig(false)
			config["duration"] = 1
			config["pricing_cycle"] = "Year"
			return config
		}()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
			states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported")
			if err != nil {
				t.Fatal(err)
			}
			data := resource.Data(states[0])
			service := &fakeFlinkWorkspaceReadService{workspace: flinkWorkspaceLifecycleFixture(false)}
			if err := readFlinkWorkspaceWithService(data, service); err != nil {
				t.Fatal(err)
			}
			state := data.State()
			for _, field := range flinkWorkspacePurchaseOptionFields {
				if _, exists := state.Attributes[field]; exists {
					t.Fatalf("Read invented API-unobservable %s state: %#v", field, state.Attributes)
				}
			}
			if state.Attributes["resource.0.cpu"] != "2" || state.Attributes["resource.0.memory"] != "8" {
				t.Fatalf("bare legacy import did not materialize observable capacity: %#v", state.Attributes)
			}
			diff, err := resource.Diff(state, terraform.NewResourceConfigRaw(test.config), nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range flinkWorkspacePurchaseOptionFields {
				var attribute *terraform.ResourceAttrDiff
				if diff != nil {
					attribute = diff.Attributes[field]
				}
				if attribute != nil {
					t.Fatalf("imported unknown %s produced diff %#v", field, attribute)
				}
			}
			if diffRequiresNew(diff) {
				t.Fatalf("bare import planned paid replacement: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceManagedPurchaseOptionChangesRemainForceNew(t *testing.T) {
	base := flinkWorkspaceLifecycleConfig(false)
	base["extra"] = "extra"
	base["promotion_code"] = "promotion"
	base["use_promotion_code"] = false
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspace().Schema, base)
	data.SetId("f-managed")
	if err := data.Set("purchase_options_state", flinkWorkspacePurchaseManaged); err != nil {
		t.Fatal(err)
	}
	if err := data.Set("capacity_intent_mode", flinkWorkspaceCapacityLegacy); err != nil {
		t.Fatal(err)
	}
	setFlinkWorkspaceTestLegacyProtocolIdentity(t, data)
	state := data.State()

	for _, test := range []struct {
		field string
		value interface{}
	}{
		{field: "auto_renew", value: false},
		{field: "duration", value: 3},
		{field: "pricing_cycle", value: "Year"},
		{field: "extra", value: "changed"},
		{field: "promotion_code", value: "changed"},
		{field: "use_promotion_code", value: true},
	} {
		t.Run(test.field, func(t *testing.T) {
			config := cloneFlinkWorkspaceLifecycleConfig(base)
			config[test.field] = test.value
			diff, err := resourceAliCloudFlinkWorkspace().Diff(state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			attribute := diff.Attributes[test.field]
			if attribute == nil || !attribute.RequiresNew {
				t.Fatalf("managed %s change did not require replacement: %#v", test.field, diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceLegacyUnclassifiedPurchaseChangeFailsClosed(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	base := flinkWorkspaceLifecycleConfig(false)
	data := schema.TestResourceDataRaw(t, resource.Schema, base)
	data.SetId("f-legacy")
	if err := data.Set("purchase_options_state", flinkWorkspacePurchaseLegacyUnclassified); err != nil {
		t.Fatal(err)
	}
	if err := data.Set("capacity_intent_mode", flinkWorkspaceCapacityLegacy); err != nil {
		t.Fatal(err)
	}
	setFlinkWorkspaceTestLegacyProtocolIdentity(t, data)
	state := data.State()

	if diff, err := resource.Diff(state, terraform.NewResourceConfigRaw(base), nil); err != nil || diffRequiresNew(diff) {
		t.Fatalf("unchanged legacy-unclassified config diff=%#v err=%v", diff, err)
	}
	changed := cloneFlinkWorkspaceLifecycleConfig(base)
	changed["duration"] = 3
	diff, err := resource.Diff(state, terraform.NewResourceConfigRaw(changed), nil)
	if err == nil || !strings.Contains(err.Error(), "re-import") {
		t.Fatalf("legacy-unclassified change diff=%#v error=%v, want fail-closed re-import error", diff, err)
	}
	if diffRequiresNew(diff) {
		t.Fatalf("legacy-unclassified change returned a replacement plan: %#v", diff.Attributes)
	}
}

func TestFlinkWorkspaceExplicitInitialImportFirstReadDoesNotWriteLegacyCapacity(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "HA"}[ha], func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
			states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported|initial")
			if err != nil {
				t.Fatal(err)
			}
			data := resource.Data(states[0])
			if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceReadService{workspace: flinkWorkspaceLifecycleFixture(ha)}); err != nil {
				t.Fatal(err)
			}
			state := data.State()
			if state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityInitialAdoptionPending {
				t.Fatalf("capacity marker changed before HCL adoption: %#v", state.Attributes)
			}
			if primary := state.Attributes["resource.#"]; primary != "" && primary != "0" {
				t.Fatalf("initial import first Read materialized legacy capacity: %#v", state.Attributes)
			}
			if standby := state.Attributes["ha.0.resource.#"]; standby != "" && standby != "0" {
				t.Fatalf("initial import first Read materialized HA legacy capacity: %#v", state.Attributes)
			}
			if initial := state.Attributes["initial_capacity.#"]; initial != "" && initial != "0" {
				t.Fatalf("Read invented initial_capacity from runtime values: %#v", state.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceReadCancellationAndRetryPreserveMarkers(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported")
	if err != nil {
		t.Fatal(err)
	}
	data := resource.Data(states[0])
	service := &fakeFlinkWorkspaceReadService{workspace: flinkWorkspaceLifecycleFixture(false), errors: []error{context.Canceled}}
	if err := readFlinkWorkspaceWithService(data, service); err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("canceled Read error = %v, want context.Canceled", err)
	}
	if got := data.Get("purchase_options_state"); got != flinkWorkspacePurchaseImportedUnknown {
		t.Fatalf("canceled Read changed purchase marker to %#v", got)
	}
	if err := readFlinkWorkspaceWithService(data, service); err != nil {
		t.Fatal(err)
	}
	if got := data.Get("purchase_options_state"); got != flinkWorkspacePurchaseImportedUnknown {
		t.Fatalf("retry Read changed purchase marker to %#v", got)
	}
}

func TestFlinkWorkspaceImportedNotFoundRemovesStateWithoutChangingProtocol(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-missing")
	if err != nil {
		t.Fatal(err)
	}
	data := resource.Data(states[0])
	service := &fakeFlinkWorkspaceReadService{errors: []error{errors.New("workspace not found")}}
	if err := readFlinkWorkspaceWithService(data, service); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "" {
		t.Fatalf("NotFound Read retained ID %q", data.Id())
	}
	if service.calls != 1 {
		t.Fatalf("NotFound Read calls = %d, want 1", service.calls)
	}
}

func TestFlinkWorkspaceCreateProtocolStateIsWrittenBeforeCloudCall(t *testing.T) {
	for _, initial := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "initial"}[initial], func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceLifecycleConfig(false)
			mode := flinkworkspace.CapacityIntentLegacy
			if initial {
				config = workspaceConfig("PRE", true)
				mode = flinkworkspace.CapacityIntentInitial
			}
			data := schema.TestResourceDataRaw(t, resource.Schema, config)
			request := &aliyunFlinkAPI.Workspace{
				Name: "workspace", Region: "cn-test", ChargeType: "PRE",
				ResourceSpec: &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8},
			}
			autoRenew := data.Get("auto_renew").(bool)
			duration := int32(data.Get("duration").(int))
			usePromotion := data.Get("use_promotion_code").(bool)
			options := flinkworkspace.CreateOptions{
				AutoRenew: &autoRenew, Duration: &duration, PricingCycle: data.Get("pricing_cycle").(string),
				Extra: data.Get("extra").(string), PromotionCode: data.Get("promotion_code").(string), UsePromotionCode: &usePromotion,
			}
			if err := setFlinkWorkspaceCreateProtocolState(data, request, options, mode); err != nil {
				t.Fatal(err)
			}
			if got := data.Get("purchase_options_state"); got != flinkWorkspacePurchaseManaged {
				t.Fatalf("purchase_options_state = %#v", got)
			}
			wantCapacity := flinkWorkspaceCapacityLegacy
			if initial {
				wantCapacity = flinkWorkspaceCapacityInitial
			}
			if got := data.Get("capacity_intent_mode"); got != wantCapacity {
				t.Fatalf("capacity_intent_mode = %#v, want %q", got, wantCapacity)
			}
			if got := data.Get("terraform_create_token").(string); got != flinkworkspace.WorkspaceCreateToken(request) {
				t.Fatalf("terraform_create_token = %q", got)
			}
			if got := data.Get("create_intent_fingerprint").(string); got != flinkworkspace.WorkspaceCreateIntentFingerprint(request, options, mode) {
				t.Fatalf("create_intent_fingerprint = %q", got)
			}
		})
	}
}

type fakeFlinkWorkspaceAdoptionService struct {
	workspace *aliyunFlinkAPI.Workspace
	describes int
	creates   int
	modifies  int
	deletes   int
	refunds   int
}

func (f *fakeFlinkWorkspaceAdoptionService) DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error) {
	f.describes++
	return f.workspace, nil
}

func TestFlinkWorkspaceInitialAdoptionUsesStateOnlyUpdateAndConverges(t *testing.T) {
	token := strings.Repeat("d", 64)
	for _, test := range []struct {
		name     string
		ha       bool
		recovery bool
	}{
		{name: "explicit initial single"},
		{name: "explicit initial HA", ha: true},
		{name: "verified recovery single", recovery: true},
		{name: "verified recovery HA", ha: true, recovery: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceInitialLifecycleConfig(test.ha)
			workspace := flinkWorkspaceInitialLifecycleFixture(test.ha)
			importID := "f-imported|initial"
			if test.recovery {
				importID = "f-imported|recover=" + token + "|initial"
				workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
			}
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, importID, workspace)
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			capacityDiff := diff.Attributes["capacity_intent_mode"]
			if capacityDiff == nil || capacityDiff.New != flinkWorkspaceCapacityInitial || capacityDiff.RequiresNew {
				t.Fatalf("capacity marker diff = %#v", capacityDiff)
			}
			purchaseDiff := diff.Attributes["purchase_options_state"]
			if test.recovery {
				if purchaseDiff == nil || purchaseDiff.New != flinkWorkspacePurchaseManaged || purchaseDiff.RequiresNew {
					t.Fatalf("purchase marker diff = %#v", purchaseDiff)
				}
			} else if purchaseDiff != nil {
				t.Fatalf("explicit initial unexpectedly adopted unknown purchase inputs: %#v", purchaseDiff)
			}

			service := &fakeFlinkWorkspaceAdoptionService{workspace: workspace}
			resource.Update = func(d *schema.ResourceData, _ interface{}) error {
				return updateFlinkWorkspaceWithService(d, service)
			}
			state, err := resource.Apply(prior, diff, nil)
			if err != nil {
				t.Fatal(err)
			}
			if service.describes != 1 || service.creates+service.modifies+service.deletes+service.refunds != 0 {
				t.Fatalf("adoption calls describes=%d writes=%d/%d/%d/%d", service.describes, service.creates, service.modifies, service.deletes, service.refunds)
			}
			if state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityInitial {
				t.Fatalf("persisted capacity marker = %#v", state.Attributes)
			}
			wantPurchase := flinkWorkspacePurchaseImportedUnknown
			if test.recovery {
				wantPurchase = flinkWorkspacePurchaseManaged
			}
			if state.Attributes["purchase_options_state"] != wantPurchase {
				t.Fatalf("persisted purchase marker = %#v, want %s", state.Attributes, wantPurchase)
			}
			secondDiff, err := resource.Diff(state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, marker := range []string{"capacity_intent_mode", "purchase_options_state"} {
				if secondDiff != nil && secondDiff.Attributes[marker] != nil {
					t.Fatalf("second Diff changed %s: %#v", marker, secondDiff.Attributes[marker])
				}
			}
			if diffRequiresNew(secondDiff) {
				t.Fatalf("second Diff requires replacement: %#v", secondDiff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceVerifiedLegacyRecoveryUsesStateOnlyUpdate(t *testing.T) {
	token := strings.Repeat("e", 64)
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceLifecycleConfig(true)
	workspace := flinkWorkspaceLifecycleFixture(true)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentLegacy)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|legacy", workspace)
	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	if marker := diff.Attributes["purchase_options_state"]; marker == nil || marker.New != flinkWorkspacePurchaseManaged || marker.RequiresNew {
		t.Fatalf("legacy recovery marker diff = %#v", marker)
	}
	service := &fakeFlinkWorkspaceAdoptionService{workspace: workspace}
	resource.Update = func(d *schema.ResourceData, _ interface{}) error { return updateFlinkWorkspaceWithService(d, service) }
	state, err := resource.Apply(prior, diff, nil)
	if err != nil {
		t.Fatal(err)
	}
	if service.describes != 1 || service.creates+service.modifies+service.deletes+service.refunds != 0 {
		t.Fatalf("legacy recovery calls describes=%d writes=%d/%d/%d/%d", service.describes, service.creates, service.modifies, service.deletes, service.refunds)
	}
	if state.Attributes["purchase_options_state"] != flinkWorkspacePurchaseManaged || state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityLegacy {
		t.Fatalf("legacy recovery state = %#v", state.Attributes)
	}
}

func TestFlinkWorkspaceVerifiedRecoveryPersistsNonDefaultPurchaseInputs(t *testing.T) {
	token := strings.Repeat("8", 64)
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	config["auto_renew"] = false
	config["duration"] = 3
	config["pricing_cycle"] = "Month"
	config["extra"] = "opaque-extra"
	config["promotion_code"] = "promotion"
	config["use_promotion_code"] = true
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	service := &fakeFlinkWorkspaceAdoptionService{workspace: workspace}
	resource.Update = func(d *schema.ResourceData, _ interface{}) error { return updateFlinkWorkspaceWithService(d, service) }
	state, err := resource.Apply(prior, diff, nil)
	if err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{
		"auto_renew": "false", "duration": "3", "pricing_cycle": "Month", "extra": "opaque-extra",
		"promotion_code": "promotion", "use_promotion_code": "true",
	} {
		if got := state.Attributes[field]; got != want {
			t.Fatalf("persisted %s = %q, want %q; state=%#v", field, got, want, state.Attributes)
		}
	}
	secondDiff, err := resource.Diff(state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil || diffRequiresNew(secondDiff) {
		t.Fatalf("non-default recovery second diff=%#v error=%v", secondDiff, err)
	}
}

func TestFlinkWorkspaceRecoveryRejectsEveryPurchaseIntentMismatchWithoutReplacement(t *testing.T) {
	token := strings.Repeat("f", 64)
	base := flinkWorkspaceInitialLifecycleConfig(false)
	base["extra"] = "extra"
	base["promotion_code"] = "promotion"
	base["use_promotion_code"] = false
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, base, flinkworkspace.CapacityIntentInitial)

	for _, test := range []struct {
		field string
		value interface{}
	}{
		{field: "auto_renew", value: false},
		{field: "duration", value: 3},
		{field: "pricing_cycle", value: "Year"},
		{field: "extra", value: "changed"},
		{field: "promotion_code", value: "changed"},
		{field: "use_promotion_code", value: true},
	} {
		t.Run(test.field, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
			config := cloneFlinkWorkspaceLifecycleConfig(base)
			config[test.field] = test.value
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err == nil || !strings.Contains(err.Error(), "fingerprint") {
				t.Fatalf("%s mismatch diff=%#v error=%v, want fingerprint failure", test.field, diff, err)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("%s mismatch planned replacement: %#v", test.field, diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceInitialAdoptionRejectsObservableMismatch(t *testing.T) {
	for _, test := range []struct {
		name      string
		ha        bool
		mutate    func(*aliyunFlinkAPI.Workspace)
		wantError string
	}{
		{name: "POST", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ChargeType = "POST" }, wantError: "PRE"},
		{name: "fixed CU", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ResourceSpec.Cpu = 3 }, wantError: "capacity"},
		{name: "fractional CU", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ResourceSpec.Cpu = 2.5 }, wantError: "capacity"},
		{name: "missing spec", mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.ResourceSpec = nil }, wantError: "capacity"},
		{name: "unexpected HA", mutate: func(workspace *aliyunFlinkAPI.Workspace) {
			workspace.Ha = true
			workspace.HaVSwitchIds = []string{"vsw-b"}
			workspace.HaResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8}
		}, wantError: "HA"},
		{name: "missing HA", ha: true, mutate: func(workspace *aliyunFlinkAPI.Workspace) {
			workspace.Ha = false
			workspace.HaVSwitchIds = nil
			workspace.HaResourceSpec = nil
		}, wantError: "HA"},
		{name: "HA cross-zone CU", ha: true, mutate: func(workspace *aliyunFlinkAPI.Workspace) { workspace.HaResourceSpec.Cpu = 3 }, wantError: "capacity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			workspace := flinkWorkspaceInitialLifecycleFixture(test.ha)
			test.mutate(workspace)
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|initial", workspace)
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(flinkWorkspaceInitialLifecycleConfig(test.ha)), nil)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("mismatch diff=%#v error=%v, want %q", diff, err, test.wantError)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("mismatch planned replacement: %#v", diff.Attributes)
			}
		})
	}
}

func TestFlinkWorkspaceAdoptionRevalidatesAtApplyAndRetainsPendingState(t *testing.T) {
	token := strings.Repeat("1", 64)
	resource := resourceAliCloudFlinkWorkspace()
	config := flinkWorkspaceInitialLifecycleConfig(false)
	workspace := flinkWorkspaceInitialLifecycleFixture(false)
	workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
	prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
	diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatal(err)
	}
	drifted := flinkWorkspaceInitialLifecycleFixture(false)
	drifted.ResourceSpec.Cpu = 4
	drifted.Tags = workspace.Tags
	service := &fakeFlinkWorkspaceAdoptionService{workspace: drifted}
	resource.Update = func(d *schema.ResourceData, _ interface{}) error { return updateFlinkWorkspaceWithService(d, service) }
	state, err := resource.Apply(prior, diff, nil)
	if err == nil || !strings.Contains(err.Error(), "capacity") {
		t.Fatalf("Apply state=%#v error=%v, want capacity revalidation failure", state, err)
	}
	if service.describes != 1 || service.creates+service.modifies+service.deletes+service.refunds != 0 {
		t.Fatalf("failed adoption calls describes=%d writes=%d/%d/%d/%d", service.describes, service.creates, service.modifies, service.deletes, service.refunds)
	}
	if state == nil || state.Attributes["purchase_options_state"] != flinkWorkspacePurchaseRecoveryPending || state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityInitialAdoptionPending {
		t.Fatalf("failed adoption did not retain retryable pending state: %#v", state)
	}
}

func TestFlinkWorkspaceAdoptionRejectsProviderTagDriftAtApplyAndRetainsPendingState(t *testing.T) {
	token := strings.Repeat("5", 64)
	for _, test := range []struct {
		name      string
		token     string
		intent    string
		wantError string
	}{
		{name: "create token", token: strings.Repeat("9", 64), wantError: "create token mismatch"},
		{name: "intent fingerprint", token: token, intent: strings.Repeat("6", 64), wantError: "between plan and apply"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceInitialLifecycleConfig(false)
			workspace := flinkWorkspaceInitialLifecycleFixture(false)
			workspace.Tags = flinkWorkspaceRecoveryTags(token, config, flinkworkspace.CapacityIntentInitial)
			prior := importAndReadFlinkWorkspaceLifecycle(t, resource, "f-imported|recover="+token+"|initial", workspace)
			diff, err := resource.Diff(prior, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			drifted := flinkWorkspaceInitialLifecycleFixture(false)
			drifted.Tags = []aliyunFlinkAPI.Tag{
				{Key: flinkworkspace.CreateTokenTagKey, Value: test.token},
				{Key: flinkworkspace.CreateIntentTagKey, Value: test.intent},
			}
			if test.intent == "" {
				drifted.Tags[1].Value = workspace.Tags[1].Value
			}
			service := &fakeFlinkWorkspaceAdoptionService{workspace: drifted}
			resource.Update = func(d *schema.ResourceData, _ interface{}) error { return updateFlinkWorkspaceWithService(d, service) }
			state, err := resource.Apply(prior, diff, nil)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Apply state=%#v error=%v, want %q", state, err, test.wantError)
			}
			if service.describes != 1 || service.creates+service.modifies+service.deletes+service.refunds != 0 {
				t.Fatalf("tag drift calls describes=%d writes=%d/%d/%d/%d", service.describes, service.creates, service.modifies, service.deletes, service.refunds)
			}
			if state == nil || state.Attributes["purchase_options_state"] != flinkWorkspacePurchaseRecoveryPending || state.Attributes["capacity_intent_mode"] != flinkWorkspaceCapacityInitialAdoptionPending {
				t.Fatalf("tag drift did not retain pending state: %#v", state)
			}
		})
	}
}

func TestFlinkWorkspaceRecoveryReadRequiresExactUniqueProviderTags(t *testing.T) {
	token := strings.Repeat("2", 64)
	fingerprint := strings.Repeat("3", 64)
	for _, test := range []struct {
		name string
		tags []aliyunFlinkAPI.Tag
	}{
		{name: "missing token", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint}}},
		{name: "wrong token", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: strings.Repeat("4", 64)}, {Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint}}},
		{name: "duplicate token", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}, {Key: flinkworkspace.CreateTokenTagKey, Value: token}, {Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint}}},
		{name: "missing fingerprint", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}}},
		{name: "invalid fingerprint", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}, {Key: flinkworkspace.CreateIntentTagKey, Value: "short"}}},
		{name: "duplicate fingerprint", tags: []aliyunFlinkAPI.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: token}, {Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint}, {Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
			states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, "f-imported|recover="+token+"|initial")
			if err != nil {
				t.Fatal(err)
			}
			workspace := flinkWorkspaceInitialLifecycleFixture(false)
			workspace.Tags = test.tags
			data := resource.Data(states[0])
			if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceReadService{workspace: workspace}); err == nil {
				t.Fatalf("recovery Read accepted tags %#v", test.tags)
			}
		})
	}
}

var flinkWorkspacePurchaseOptionFields = []string{
	"auto_renew", "duration", "pricing_cycle", "extra", "promotion_code", "use_promotion_code",
}

func flinkWorkspaceLifecycleConfig(ha bool) map[string]interface{} {
	config := workspaceConfig("PRE", false)
	config["auto_renew"] = true
	config["duration"] = 1
	config["pricing_cycle"] = "Month"
	config["use_promotion_code"] = false
	config["resource"] = []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}}
	if ha {
		config["ha"] = []interface{}{map[string]interface{}{
			"vswitch_ids": []interface{}{"vsw-b"},
			"resource":    []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
		}}
	}
	return config
}

func flinkWorkspaceInitialLifecycleConfig(ha bool) map[string]interface{} {
	config := workspaceConfig("PRE", true)
	config["auto_renew"] = true
	config["duration"] = 1
	config["pricing_cycle"] = "Month"
	config["use_promotion_code"] = false
	if ha {
		config["ha"] = []interface{}{map[string]interface{}{"vswitch_ids": []interface{}{"vsw-b"}}}
		config["initial_capacity"] = []interface{}{map[string]interface{}{"fixed_cu": 0, "cross_zone_fixed_cu": 2}}
	}
	return config
}

func cloneFlinkWorkspaceLifecycleConfig(source map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{}, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func flinkWorkspaceLifecycleFixture(ha bool) *aliyunFlinkAPI.Workspace {
	workspace := &aliyunFlinkAPI.Workspace{
		Id:               "f-imported",
		Name:             "workspace",
		ResourceGroupId:  "rg-1",
		ZoneId:           "cn-test-a",
		VpcId:            "vpc-1",
		VSwitchIds:       []string{"vsw-a"},
		ArchitectureType: "X86",
		ChargeType:       "PRE",
		MonitorType:      "ARMS",
		ResourceSpec:     &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8},
		Storage:          &aliyunFlinkAPI.Storage{Oss: &aliyunFlinkAPI.OSSStorage{Bucket: "bucket"}},
	}
	if ha {
		workspace.Ha = true
		workspace.HaZoneId = "cn-test-b"
		workspace.HaVSwitchIds = []string{"vsw-b"}
		workspace.HaResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: 2, MemoryGB: 8}
	}
	return workspace
}

func flinkWorkspaceInitialLifecycleFixture(ha bool) *aliyunFlinkAPI.Workspace {
	workspace := flinkWorkspaceLifecycleFixture(ha)
	if ha {
		workspace.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: 0, MemoryGB: 0}
	}
	return workspace
}

func importAndReadFlinkWorkspaceLifecycle(t *testing.T, resource *schema.Resource, importID string, workspace *aliyunFlinkAPI.Workspace) *terraform.InstanceState {
	t.Helper()
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"alicloud_flink_workspace": resource}}
	states, err := provider.ImportState(&terraform.InstanceInfo{Type: "alicloud_flink_workspace"}, importID)
	if err != nil {
		t.Fatal(err)
	}
	data := resource.Data(states[0])
	if err := readFlinkWorkspaceWithService(data, &fakeFlinkWorkspaceReadService{workspace: workspace}); err != nil {
		t.Fatal(err)
	}
	return data.State()
}

func flinkWorkspaceRecoveryTags(token string, config map[string]interface{}, mode string) []aliyunFlinkAPI.Tag {
	request, options := flinkWorkspaceTestCreateIntent(config, mode)
	return []aliyunFlinkAPI.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: token},
		{Key: flinkworkspace.CreateIntentTagKey, Value: flinkworkspace.WorkspaceCreateIntentFingerprint(request, options, mode)},
	}
}

func flinkWorkspaceTestCreateIntent(config map[string]interface{}, mode string) (*aliyunFlinkAPI.Workspace, flinkworkspace.CreateOptions) {
	request := &aliyunFlinkAPI.Workspace{ChargeType: config["charge_type"].(string)}
	if mode == flinkworkspace.CapacityIntentInitial {
		capacity := firstTestBlock(config["initial_capacity"])
		fixed := capacity["fixed_cu"].(int)
		request.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(fixed), MemoryGB: float64(fixed * 4)}
	} else {
		resource := firstTestBlock(config["resource"])
		request.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(resource["cpu"].(int)), MemoryGB: float64(resource["memory"].(int))}
	}
	if ha := firstTestBlock(config["ha"]); ha != nil {
		request.HighAvailability = &aliyunFlinkAPI.HighAvailability{Enabled: true}
		if mode == flinkworkspace.CapacityIntentInitial {
			capacity := firstTestBlock(config["initial_capacity"])
			cross := capacity["cross_zone_fixed_cu"].(int)
			request.HighAvailability.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(cross), MemoryGB: float64(cross * 4)}
		} else {
			resource := firstTestBlock(ha["resource"])
			request.HighAvailability.ResourceSpec = &aliyunFlinkAPI.ResourceSpec{Cpu: float64(resource["cpu"].(int)), MemoryGB: float64(resource["memory"].(int))}
		}
	}
	autoRenew, _ := config["auto_renew"].(bool)
	duration := int32(1)
	if configured, ok := config["duration"].(int); ok {
		duration = int32(configured)
	}
	pricingCycle, _ := config["pricing_cycle"].(string)
	if pricingCycle == "" {
		pricingCycle = "Month"
	}
	usePromotion, _ := config["use_promotion_code"].(bool)
	return request, flinkworkspace.CreateOptions{
		AutoRenew: &autoRenew, Duration: &duration, PricingCycle: pricingCycle,
		Extra: configString(config, "extra"), PromotionCode: configString(config, "promotion_code"), UsePromotionCode: &usePromotion,
	}
}

func configString(config map[string]interface{}, key string) string {
	value, _ := config[key].(string)
	return value
}

func setFlinkWorkspaceTestLegacyProtocolIdentity(t *testing.T, data *schema.ResourceData) {
	t.Helper()
	for key, value := range map[string]string{
		"identity_visibility_state": flinkWorkspaceIdentityStable,
		"terraform_create_token":    flinkWorkspaceProtocolUnavailable,
		"create_intent_fingerprint": flinkWorkspaceProtocolUnavailable,
	} {
		if err := data.Set(key, value); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSDKV1ComputedMarkerSetNewTriggersUpdateAndPersists(t *testing.T) {
	const (
		pending = "RECOVERY_PENDING"
		managed = "MANAGED"
	)

	var (
		customizeSaw string
		updateCalls  int
	)
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"purchase_input": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(_ string, _, _ string, d *schema.ResourceData) bool {
					return d.Get("purchase_options_state").(string) == pending
				},
			},
			"purchase_options_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{State: func(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
			d.SetId("f-imported")
			if err := d.Set("purchase_options_state", pending); err != nil {
				return nil, err
			}
			return []*schema.ResourceData{d}, nil
		}},
		CustomizeDiff: func(d *schema.ResourceDiff, _ interface{}) error {
			customizeSaw = d.Get("purchase_input").(string)
			return d.SetNew("purchase_options_state", managed)
		},
		Read: func(*schema.ResourceData, interface{}) error { return nil },
		Update: func(d *schema.ResourceData, _ interface{}) error {
			updateCalls++
			if got := d.Get("purchase_options_state").(string); got != managed {
				t.Fatalf("Update marker = %q, want %q", got, managed)
			}
			return nil
		},
	}
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{"test_workspace": resource}}

	imported, err := provider.ImportState(&terraform.InstanceInfo{Type: "test_workspace"}, "f-imported")
	if err != nil {
		t.Fatal(err)
	}
	diff, err := resource.Diff(imported[0], terraform.NewResourceConfigRaw(map[string]interface{}{
		"purchase_input": "effective-hcl-value",
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if customizeSaw != "effective-hcl-value" {
		t.Fatalf("CustomizeDiff Get(purchase_input) = %q, want effective HCL value", customizeSaw)
	}
	if inputDiff := diff.Attributes["purchase_input"]; inputDiff != nil {
		t.Fatalf("suppressed purchase_input still produced a diff: %#v", inputDiff)
	}
	markerDiff := diff.Attributes["purchase_options_state"]
	if markerDiff == nil || markerDiff.New != managed || markerDiff.RequiresNew {
		t.Fatalf("computed marker diff = %#v, want in-place transition to %q", markerDiff, managed)
	}

	state, err := resource.Apply(imported[0], diff, nil)
	if err != nil {
		t.Fatal(err)
	}
	if updateCalls != 1 {
		t.Fatalf("Update calls = %d, want 1", updateCalls)
	}
	if got := state.Attributes["purchase_options_state"]; got != managed {
		t.Fatalf("persisted marker = %q, want %q", got, managed)
	}
}
