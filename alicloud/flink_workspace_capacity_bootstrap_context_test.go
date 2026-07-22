package alicloud

import (
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestFlinkWorkspaceCapacityBootstrapContextCanonicalRoundTrip(t *testing.T) {
	context := flinkWorkspaceCapacityBootstrapContext{
		Version:                          flinkWorkspaceCapacityBootstrapContextVersion,
		Origin:                           flinkWorkspaceCapacityBootstrapContextManagedInitialCreate,
		ExpectedInstanceID:               "f-cn-bootstrap",
		ExpectedResourceID:               "resource-bootstrap",
		TerraformCreateToken:             strings.Repeat("a", 64),
		CreateIntentFingerprint:          strings.Repeat("b", 64),
		IdentityAbsenceRetryNotAfterUnix: time.Unix(1_700_000_000, 0).Add(flinkWorkspaceCapacityBootstrapIdentityAbsenceRetryWindow).Unix(),
	}

	encoded, err := encodeFlinkWorkspaceCapacityBootstrapContext(context)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"version":1,"origin":"MANAGED_INITIAL_CREATE","expected_instance_id":"f-cn-bootstrap","expected_resource_id":"resource-bootstrap","terraform_create_token":"` + strings.Repeat("a", 64) + `","create_intent_fingerprint":"` + strings.Repeat("b", 64) + `","identity_absence_retry_not_after_unix":1700000900}`
	if encoded != want {
		t.Fatalf("encoded context = %q, want canonical %q", encoded, want)
	}

	decoded, err := parseFlinkWorkspaceCapacityBootstrapContext(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != context {
		t.Fatalf("decoded context = %#v, want %#v", decoded, context)
	}

	// Expiry removes only the identity-absence retry permission. It does not
	// make the persisted provenance context invalid.
	if _, err := parseFlinkWorkspaceCapacityBootstrapContext(encoded); err != nil {
		t.Fatalf("expired-but-canonical context must remain parseable: %v", err)
	}
}

func TestFlinkWorkspaceCapacityBootstrapContextRejectsNonCanonicalOrInvalidInput(t *testing.T) {
	valid := flinkWorkspaceCapacityBootstrapContext{
		Version:                          flinkWorkspaceCapacityBootstrapContextVersion,
		Origin:                           flinkWorkspaceCapacityBootstrapContextManagedInitialCreate,
		ExpectedInstanceID:               "f-cn-bootstrap",
		ExpectedResourceID:               "resource-bootstrap",
		TerraformCreateToken:             strings.Repeat("a", 64),
		CreateIntentFingerprint:          strings.Repeat("b", 64),
		IdentityAbsenceRetryNotAfterUnix: 1_700_000_900,
	}
	canonical, err := encodeFlinkWorkspaceCapacityBootstrapContext(valid)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "unknown field", value: strings.TrimSuffix(canonical, "}") + `,"unknown":true}`},
		{name: "duplicate field", value: strings.Replace(canonical, `"version":1`, `"version":1,"version":1`, 1)},
		{name: "trailing token", value: canonical + `{}`},
		{name: "whitespace is not canonical", value: " " + canonical},
		{name: "wrong version", value: strings.Replace(canonical, `"version":1`, `"version":2`, 1)},
		{name: "wrong origin", value: strings.Replace(canonical, `MANAGED_INITIAL_CREATE`, `IMPORTED`, 1)},
		{name: "empty instance", value: strings.Replace(canonical, `f-cn-bootstrap`, ``, 1)},
		{name: "uppercase token", value: strings.Replace(canonical, strings.Repeat("a", 64), strings.Repeat("A", 64), 1)},
		{name: "short fingerprint", value: strings.Replace(canonical, strings.Repeat("b", 64), strings.Repeat("b", 63), 1)},
		{name: "zero deadline", value: strings.Replace(canonical, `1700000900`, `0`, 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseFlinkWorkspaceCapacityBootstrapContext(test.value); err == nil {
				t.Fatalf("parse context %q succeeded, want fail closed", test.value)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityBootstrapContextEncoderRejectsInvalidStruct(t *testing.T) {
	valid := flinkWorkspaceCapacityBootstrapContext{
		Version:                          flinkWorkspaceCapacityBootstrapContextVersion,
		Origin:                           flinkWorkspaceCapacityBootstrapContextManagedInitialCreate,
		ExpectedInstanceID:               "f-cn-bootstrap",
		ExpectedResourceID:               "resource-bootstrap",
		TerraformCreateToken:             strings.Repeat("a", 64),
		CreateIntentFingerprint:          strings.Repeat("b", 64),
		IdentityAbsenceRetryNotAfterUnix: 1,
	}

	tests := []struct {
		name   string
		mutate func(*flinkWorkspaceCapacityBootstrapContext)
	}{
		{name: "version", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.Version = 0 }},
		{name: "origin", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.Origin = "" }},
		{name: "instance", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.ExpectedInstanceID = "" }},
		{name: "token", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.TerraformCreateToken = "invalid" }},
		{name: "fingerprint", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.CreateIntentFingerprint = "invalid" }},
		{name: "deadline", mutate: func(value *flinkWorkspaceCapacityBootstrapContext) { value.IdentityAbsenceRetryNotAfterUnix = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := valid
			test.mutate(&value)
			if _, err := encodeFlinkWorkspaceCapacityBootstrapContext(value); err == nil {
				t.Fatalf("encode %#v succeeded, want validation error", value)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityBootstrapContextSchemaAndFreshCreateLifecycle(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	field := resource.Schema["capacity_bootstrap_context"]
	if field == nil || field.Type != schema.TypeString || !field.Computed || !field.Sensitive || field.Optional || field.Required || field.ForceNew {
		t.Fatalf("capacity_bootstrap_context schema = %#v, want computed+sensitive state-only string", field)
	}
	if resource.SchemaVersion != 1 {
		t.Fatalf("workspace schema version = %d, want unchanged v1", resource.SchemaVersion)
	}

	createdAt := time.Unix(1_700_000_000, 987_654_321)
	data := schema.TestResourceDataRaw(t, resource.Schema, nil)
	if err := data.Set("terraform_create_token", strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	if err := data.Set("create_intent_fingerprint", strings.Repeat("b", 64)); err != nil {
		t.Fatal(err)
	}
	workspace := &flink.Workspace{Id: "f-cn-bootstrap", ResourceId: "resource-bootstrap"}
	if err := completeFlinkWorkspaceCreateWithCapacityBootstrapContext(data, workspace, true, createdAt); err != nil {
		t.Fatal(err)
	}
	if data.Id() != workspace.Id {
		t.Fatalf("workspace ID = %q, want %q", data.Id(), workspace.Id)
	}
	raw, ok := data.GetOk("capacity_bootstrap_context")
	if !ok {
		t.Fatal("fresh managed initial Create did not persist capacity bootstrap context")
	}
	context, err := parseFlinkWorkspaceCapacityBootstrapContext(raw.(string))
	if err != nil {
		t.Fatal(err)
	}
	want := flinkWorkspaceCapacityBootstrapContext{
		Version:                          flinkWorkspaceCapacityBootstrapContextVersion,
		Origin:                           flinkWorkspaceCapacityBootstrapContextManagedInitialCreate,
		ExpectedInstanceID:               workspace.Id,
		ExpectedResourceID:               workspace.ResourceId,
		TerraformCreateToken:             strings.Repeat("a", 64),
		CreateIntentFingerprint:          strings.Repeat("b", 64),
		IdentityAbsenceRetryNotAfterUnix: createdAt.Add(flinkWorkspaceCapacityBootstrapIdentityAbsenceRetryWindow).Unix(),
	}
	if context != want {
		t.Fatalf("persisted context = %#v, want %#v", context, want)
	}
}

func TestFlinkWorkspaceCapacityBootstrapContextUsesStrictSentinelWithoutGrantingGrace(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()

	for _, test := range []struct {
		name        string
		workspace   *flink.Workspace
		usesInitial bool
	}{
		{name: "legacy create", workspace: &flink.Workspace{Id: "f-legacy", ResourceId: "resource-legacy"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resource.Schema, nil)
			_ = data.Set("terraform_create_token", strings.Repeat("a", 64))
			_ = data.Set("create_intent_fingerprint", strings.Repeat("b", 64))
			if err := completeFlinkWorkspaceCreateWithCapacityBootstrapContext(data, test.workspace, test.usesInitial, time.Unix(1_700_000_000, 0)); err != nil {
				t.Fatal(err)
			}
			if data.Id() != test.workspace.Id {
				t.Fatalf("workspace ID = %q, want paid ID %q", data.Id(), test.workspace.Id)
			}
			if got := data.Get("capacity_bootstrap_context"); got != flinkWorkspaceCapacityBootstrapContextStrictExisting {
				t.Fatalf("%s context = %q, want strict sentinel", test.name, got)
			}
		})
	}

	for _, importID := range []string{
		"f-imported|initial",
		"f-recovered|recover=" + strings.Repeat("a", 64) + "|initial",
	} {
		data := schema.TestResourceDataRaw(t, resource.Schema, nil)
		data.SetId(importID)
		states, err := resource.Importer.State(data, nil)
		if err != nil {
			t.Fatalf("import %q: %v", importID, err)
		}
		if len(states) != 1 {
			t.Fatalf("import %q returned %d states, want one", importID, len(states))
		}
		if got := states[0].Get("capacity_bootstrap_context"); got != flinkWorkspaceCapacityBootstrapContextStrictExisting {
			t.Fatalf("import %q context = %q, want strict sentinel", importID, got)
		}
	}

	upgraded, err := upgradeFlinkWorkspaceStateV0(map[string]interface{}{
		"id":               "f-schema-v0",
		"name":             "schema-v0",
		"initial_capacity": []interface{}{map[string]interface{}{"fixed_cu": 0, "cross_zone_fixed_cu": 2}},
	}, &connectivity.AliyunClient{RegionId: "cn-test"})
	if err != nil {
		t.Fatal(err)
	}
	if got := upgraded["capacity_bootstrap_context"]; got != flinkWorkspaceCapacityBootstrapContextStrictExisting {
		t.Fatalf("schema-v0 context = %q, want strict sentinel", got)
	}
}

func TestFlinkWorkspaceCapacityBootstrapContextReturnsImmediatelyAfterIDOnlyCreate(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspace()
	data := schema.TestResourceDataRaw(t, resource.Schema, nil)
	token := strings.Repeat("a", 64)
	fingerprint := strings.Repeat("b", 64)
	if err := data.Set("terraform_create_token", token); err != nil {
		t.Fatal(err)
	}
	if err := data.Set("create_intent_fingerprint", fingerprint); err != nil {
		t.Fatal(err)
	}
	createdAt := time.Now()
	created := &flink.Workspace{Id: "f-id-only", Name: "id-only", Region: "cn-test"}
	if err := completeFlinkWorkspaceCreateWithCapacityBootstrapContext(data, created, true, createdAt); err != nil {
		t.Fatal(err)
	}
	if data.Id() != created.Id {
		t.Fatalf("ID-only Create ID = %q, want %q", data.Id(), created.Id)
	}
	raw, ok := data.GetOk("capacity_bootstrap_context")
	if !ok {
		t.Fatal("ID-only Create did not persist capacity bootstrap context before returning")
	}
	contextValue, err := parseFlinkWorkspaceCapacityBootstrapContext(raw.(string))
	if err != nil {
		t.Fatal(err)
	}
	if contextValue.ExpectedInstanceID != created.Id || contextValue.ExpectedResourceID != "" || contextValue.IdentityAbsenceRetryNotAfterUnix != createdAt.Add(flinkWorkspaceCapacityBootstrapIdentityAbsenceRetryWindow).Unix() {
		t.Fatalf("discovered context identity/deadline = %#v", contextValue)
	}
}
