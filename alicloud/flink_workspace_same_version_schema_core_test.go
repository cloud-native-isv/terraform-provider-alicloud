package alicloud

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestFlinkWorkspaceCoreSchemaGateAtExactProviderConstraintRejectsDowngrade
// exercises Terraform Core from Plugin SDK v1.17.2. The fixture declares and
// verifies the exact = 1.247.0 provider requirement before returning either
// in-process factory. It proves the Core schema gate, not real binary selection
// for that coordinate; the latter belongs to the OpenTofu mirror/suite gate.
func TestFlinkWorkspaceCoreSchemaGateAtExactProviderConstraintRejectsDowngrade(t *testing.T) {
	moduleCache, err := flinkAllocationCoreGoEnv("GOMODCACHE")
	if err != nil {
		t.Fatal(err)
	}
	sdkDir := filepath.Join(moduleCache, "github.com/hashicorp/terraform-plugin-sdk@v1.17.2")
	if _, err := os.Stat(sdkDir); err != nil {
		t.Fatalf("Plugin SDK v1.17.2 module directory %q: %v", sdkDir, err)
	}

	fixtureDir := t.TempDir()
	goMod := fmt.Sprintf("module github.com/hashicorp/terraform-plugin-sdk/flink-workspace-core-schema-gate\n\ngo 1.20\n\nrequire github.com/hashicorp/terraform-plugin-sdk v1.17.2\n\nreplace github.com/hashicorp/terraform-plugin-sdk => %s\n", sdkDir)
	if err := os.WriteFile(filepath.Join(fixtureDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "main.go"), []byte(flinkWorkspaceCoreSchemaGateProgram), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "race_enabled.go"), []byte(flinkWorkspaceCoreRaceEnabledProgram), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "race_disabled.go"), []byte(flinkWorkspaceCoreRaceDisabledProgram), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "-race", "-mod=mod", ".")
	cmd.Dir = fixtureDir
	cmd.Env = append(os.Environ(), "GOPROXY=off")
	output, err := cmd.CombinedOutput()
	t.Logf("Terraform Core exact-constraint schema downgrade fixture:\n%s", output)
	if err != nil {
		t.Fatalf("Terraform Core exact-constraint schema downgrade contract failed: %v\n%s", err, output)
	}
}

// TestFlinkWorkspaceCurrentResourceKeepsLegacyModuleShapesNoOp is the current
// Resource forward-compatibility gate. It covers the cws-lib module's single
// and HA legacy shapes through the production Read implementation, but does
// not claim that a deployment layer selected a particular provider binary.
func TestFlinkWorkspaceCurrentResourceKeepsLegacyModuleShapesNoOp(t *testing.T) {
	for _, test := range []struct {
		name string
		ha   bool
	}{
		{name: "single legacy resource block"},
		{name: "HA legacy resource blocks", ha: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspace()
			config := flinkWorkspaceLegacyModuleConfig(test.ha)
			service := &fakeFlinkWorkspaceReadService{workspace: flinkWorkspaceLegacyModuleFixture(test.ha)}
			data := schema.TestResourceDataRaw(t, resource.Schema, config)
			data.SetId("f-schema-v0")
			state := data.State()
			for _, marker := range []string{
				"identity_visibility_state",
				"purchase_options_state",
				"capacity_intent_mode",
				"terraform_create_token",
				"create_intent_fingerprint",
			} {
				delete(state.Attributes, marker)
			}
			state.Meta = map[string]interface{}{"schema_version": "0"}

			writes := struct{ creates, updates, deletes int }{}
			reads := 0
			resource.Create = func(*schema.ResourceData, interface{}) error {
				writes.creates++
				return nil
			}
			resource.Read = func(data *schema.ResourceData, _ interface{}) error {
				reads++
				return readFlinkWorkspaceWithService(data, service)
			}
			resource.Update = func(*schema.ResourceData, interface{}) error {
				writes.updates++
				return nil
			}
			resource.Delete = func(*schema.ResourceData, interface{}) error {
				writes.deletes++
				return nil
			}

			migrated, err := resource.Refresh(state, nil)
			if err != nil {
				t.Fatal(err)
			}
			if reads != 1 || service.calls != 1 || writes != (struct{ creates, updates, deletes int }{}) {
				t.Fatalf("legacy migration callbacks reads=%d describes=%d writes=%+v, want one production Read/Describe and no cloud write", reads, service.calls, writes)
			}
			if migrated.Meta["schema_version"] != "1" || migrated.Attributes["resource.#"] != "1" {
				t.Fatalf("legacy primary resource did not survive v0->v1 migration: meta=%#v state=%#v", migrated.Meta, migrated.Attributes)
			}
			if test.ha && (migrated.Attributes["ha.#"] != "1" || migrated.Attributes["ha.0.resource.#"] != "1") {
				t.Fatalf("legacy HA resource did not survive v0->v1 migration: %#v", migrated.Attributes)
			}
			wantChargeType := "POST"
			if test.ha {
				wantChargeType = "PRE"
			}
			for key, want := range map[string]string{
				"charge_type":          wantChargeType,
				"security_group_id":    "sg-workspace",
				"zone_id":              "cn-test-a",
				"vpc_id":               "vpc-1",
				"vswitch_ids.0":        "vsw-a",
				"storage.0.oss_bucket": "bucket",
				"resource.0.cpu":       "2",
				"resource.0.memory":    "8",
				"architecture_type":    "X86",
				"auto_renew":           "true",
				"duration":             "1",
				"pricing_cycle":        "Month",
				"extra":                "",
				"monitor_type":         "ARMS",
			} {
				if got := migrated.Attributes[key]; got != want {
					t.Fatalf("authoritative Read state %s=%q, want %q: %#v", key, got, want, migrated.Attributes)
				}
			}
			if test.ha {
				for key, want := range map[string]string{
					"ha.0.zone_id":           "cn-test-b",
					"ha.0.vswitch_ids.0":     "vsw-b",
					"ha.0.resource.0.cpu":    "2",
					"ha.0.resource.0.memory": "8",
				} {
					if got := migrated.Attributes[key]; got != want {
						t.Fatalf("authoritative HA Read state %s=%q, want %q: %#v", key, got, want, migrated.Attributes)
					}
				}
			}

			diff, err := resource.Diff(migrated, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			if diff != nil && !diff.Empty() {
				t.Fatalf("legacy config changed after v0->v1 migration: %#v", diff)
			}
			if diffRequiresNew(diff) {
				t.Fatalf("legacy config planned replacement after v0->v1 migration: %#v", diff.Attributes)
			}
			if writes != (struct{ creates, updates, deletes int }{}) {
				t.Fatalf("legacy no-op diff reached cloud write callback: %+v", writes)
			}
		})
	}
}

func flinkWorkspaceLegacyModuleConfig(ha bool) map[string]interface{} {
	chargeType := "POST"
	if ha {
		// The cws-lib module defaults single-zone workspaces to POST. Its HA
		// shape is valid only with the supported explicit PRE override.
		chargeType = "PRE"
	}
	config := map[string]interface{}{
		"name":              "workspace",
		"resource_group_id": "rg-1",
		"zone_id":           "cn-test-a",
		"vpc_id":            "vpc-1",
		"vswitch_ids":       []interface{}{"vsw-a"},
		"security_group_id": "sg-workspace",
		"architecture_type": "X86",
		"auto_renew":        true,
		"charge_type":       chargeType,
		"duration":          1,
		"extra":             "",
		"monitor_type":      "ARMS",
		"resource":          []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
		"storage":           []interface{}{map[string]interface{}{"oss_bucket": "bucket"}},
	}
	if ha {
		config["ha"] = []interface{}{map[string]interface{}{
			"zone_id":     "cn-test-b",
			"vswitch_ids": []interface{}{"vsw-b"},
			"resource":    []interface{}{map[string]interface{}{"cpu": 2, "memory": 8}},
		}}
	}
	return config
}

func flinkWorkspaceLegacyModuleFixture(ha bool) *aliyunFlinkAPI.Workspace {
	workspace := flinkWorkspaceLifecycleFixture(ha)
	workspace.Id = "f-schema-v0"
	workspace.ChargeType = "POST"
	if ha {
		workspace.ChargeType = "PRE"
	}
	workspace.SecurityGroupInfo = &aliyunFlinkAPI.SecurityGroupInfo{SecurityGroupId: "sg-workspace"}
	return workspace
}

const flinkWorkspaceCoreSchemaGateProgram = `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/helper/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/internal/plugin/discovery"
	"github.com/hashicorp/terraform-plugin-sdk/internal/providers"
	"github.com/hashicorp/terraform-plugin-sdk/internal/states"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	tfplugin "github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const (
	resourceType             = "alicloud_flink_workspace"
	resourceID               = "f-schema-v1"
	exactProviderRequirement = "terraform {\n  required_providers {\n    alicloud = \"= 1.247.0\"\n  }\n}\n\n"
)

var markerValues = map[string]string{
	"identity_visibility_state": "STABLE",
	"purchase_options_state":    "MANAGED",
	"capacity_intent_mode":      "LEGACY",
	"terraform_create_token":    strings.Repeat("a", 64),
	"create_intent_fingerprint": strings.Repeat("b", 64),
}

var privateStateSentinel = []byte("{\"fixture\":\"flink-workspace-schema-v1-private-sentinel\",\"schema_version\":\"1\"}")

func productionShapedDependencies() []addrs.Referenceable {
	return []addrs.Referenceable{
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_resource_manager_resource_group", Name: "workspace"},
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_vpc", Name: "workspace"},
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_vswitch", Name: "primary"},
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_vswitch", Name: "standby"},
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_security_group", Name: "workspace"},
		addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_oss_bucket", Name: "workspace"},
		addrs.Resource{Mode: addrs.DataResourceMode, Type: "alicloud_zones", Name: "primary"},
		addrs.Resource{Mode: addrs.DataResourceMode, Type: "alicloud_zones", Name: "standby"},
	}
}

type callbackCounts struct {
	creates int
	reads   int
	updates int
	deletes int
}

var currentCallbacks callbackCounts
var oldCallbacks callbackCounts

func main() {
	log.SetOutput(io.Discard)
	if !raceInstrumented {
		fmt.Fprintln(os.Stderr, "inner Terraform Core schema gate is not race-instrumented")
		os.Exit(1)
	}
	if err := run("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dir string) error {
	stableConfig, err := loadConfig(filepath.Join(dir, "stable"), exactProviderRequirement+"resource \"alicloud_flink_workspace\" \"this\" {\n  name = \"core-workspace\"\n  description = \"before\"\n  resource {\n    cpu = 2\n    memory = 8\n  }\n}\n")
	if err != nil {
		return err
	}
	changedConfig, err := loadConfig(filepath.Join(dir, "changed"), exactProviderRequirement+"resource \"alicloud_flink_workspace\" \"this\" {\n  name = \"core-workspace\"\n  description = \"after\"\n  resource {\n    cpu = 2\n    memory = 8\n  }\n}\n")
	if err != nil {
		return err
	}
	removedConfig, err := loadConfig(filepath.Join(dir, "removed"), exactProviderRequirement)
	if err != nil {
		return err
	}

	currentProvider := workspaceProvider(1, &currentCallbacks)
	currentContext, err := newContext(stableConfig, nil, currentProvider)
	if err != nil {
		return err
	}
	if _, diags := currentContext.Plan(); diags.HasErrors() {
		return fmt.Errorf("schema-v1 initial plan: %s", diags.Err())
	}
	state, diags := currentContext.Apply()
	if diags.HasErrors() {
		return fmt.Errorf("schema-v1 initial apply: %s", diags.Err())
	}
	if currentCallbacks.creates != 1 {
		return fmt.Errorf("schema-v1 callbacks=%+v, want exactly one Create", currentCallbacks)
	}

	addr := resourceAddress()
	instance := state.ResourceInstance(addr)
	if instance == nil || instance.Current == nil {
		return fmt.Errorf("schema-v1 state injection target missing: %#v", instance)
	}
	// Private and Dependencies are opaque persisted-state metadata. Inject
	// deterministic production-shaped sentinels so every failed operation and
	// retry proves that Core preserves the whole object, not only AttrsJSON.
	instance.Current.Private = append([]byte(nil), privateStateSentinel...)
	instance.Current.Dependencies = append([]addrs.Referenceable(nil), productionShapedDependencies()...)
	snapshot, err := stateSnapshot(state, addr)
	if err != nil {
		return fmt.Errorf("schema-v1 state snapshot: %w", err)
	}
	if err := assertSchemaV1Snapshot(snapshot, "schema-v1 state"); err != nil {
		return err
	}

	oldProvider := workspaceProvider(0, &oldCallbacks)
	attempts := []struct {
		name    string
		config  *configs.Config
		refresh bool
	}{
		{name: "refresh/read", config: stableConfig, refresh: true},
		{name: "changed/update", config: changedConfig},
		{name: "removed/delete", config: removedConfig},
		{name: "retry refresh/read", config: stableConfig, refresh: true},
	}
	for _, attempt := range attempts {
		oldContext, err := newContext(attempt.config, state, oldProvider)
		if err != nil {
			return fmt.Errorf("%s context: %w", attempt.name, err)
		}
		var failedState *states.State
		if attempt.refresh {
			failedState, diags = oldContext.Refresh()
		} else {
			_, diags = oldContext.Plan()
		}
		if !diags.HasErrors() {
			return fmt.Errorf("%s accepted schema-v1 state with schema-v0 provider", attempt.name)
		}
		errText := diags.Err().Error()
		if !strings.Contains(errText, "Resource instance managed by newer provider version") {
			return fmt.Errorf("%s error=%q, want newer-provider-version gate", attempt.name, errText)
		}
		if err := assertStateSnapshot(state, addr, snapshot, attempt.name+" input state"); err != nil {
			return err
		}
		// Context.Refresh returns nil when the graph walk has errors. The
		// caller-owned input is the persisted state boundary; if Core ever
		// returns a partial candidate, it must also retain the exact snapshot.
		if failedState != nil {
			if err := assertStateSnapshot(failedState, addr, snapshot, attempt.name+" returned state"); err != nil {
				return err
			}
		}
		if oldCallbacks != (callbackCounts{}) {
			return fmt.Errorf("%s reached old cloud callback before schema gate: %+v", attempt.name, oldCallbacks)
		}
	}

	fmt.Printf("exact = 1.247.0 requirement accepted; schema-v1 state rejected by schema-v0 Core before callbacks; attempts=%d; complete object/private/%d dependencies preserved; old callbacks=%+v\n", len(attempts), len(snapshot.Dependencies), oldCallbacks)
	return nil
}

func workspaceProvider(version int, callbacks *callbackCounts) *schema.Provider {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":        {Type: schema.TypeString, Required: true, ForceNew: true},
			"description": {Type: schema.TypeString, Optional: true},
			"resource": {
				Type: schema.TypeList, Optional: true, ForceNew: true, MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"cpu":    {Type: schema.TypeInt, Required: true, ForceNew: true},
					"memory": {Type: schema.TypeInt, Required: true, ForceNew: true},
				}},
			},
		},
		Create: func(d *schema.ResourceData, _ interface{}) error {
			callbacks.creates++
			d.SetId(resourceID)
			if version == 1 {
				for name, value := range markerValues {
					if err := d.Set(name, value); err != nil {
						return fmt.Errorf("set %s: %w", name, err)
					}
				}
			}
			return nil
		},
		Read: func(*schema.ResourceData, interface{}) error {
			callbacks.reads++
			return nil
		},
		Update: func(*schema.ResourceData, interface{}) error {
			callbacks.updates++
			return nil
		},
		Delete: func(d *schema.ResourceData, _ interface{}) error {
			callbacks.deletes++
			d.SetId("")
			return nil
		},
	}
	if version == 1 {
		resource.SchemaVersion = 1
		for name := range markerValues {
			resource.Schema[name] = &schema.Schema{Type: schema.TypeString, Computed: true}
		}
	}
	return &schema.Provider{ResourcesMap: map[string]*schema.Resource{resourceType: resource}}
}

func resourceAddress() addrs.AbsResourceInstance {
	return addrs.Resource{Mode: addrs.ManagedResourceMode, Type: resourceType, Name: "this"}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance)
}

func stateSnapshot(state *states.State, addr addrs.AbsResourceInstance) (*states.ResourceInstanceObjectSrc, error) {
	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}
	instance := state.ResourceInstance(addr)
	if instance == nil || instance.Current == nil {
		return nil, fmt.Errorf("workspace object missing: %#v", instance)
	}
	return instance.Current.DeepCopy(), nil
}

func assertStateSnapshot(state *states.State, addr addrs.AbsResourceInstance, want *states.ResourceInstanceObjectSrc, label string) error {
	got, err := stateSnapshot(state, addr)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("%s was rewritten or discarded: got=%#v want=%#v", label, got, want)
	}
	return assertSchemaV1Snapshot(got, label)
}

func assertSchemaV1Snapshot(snapshot *states.ResourceInstanceObjectSrc, label string) error {
	if snapshot.SchemaVersion != 1 {
		return fmt.Errorf("%s schema version=%d, want 1", label, snapshot.SchemaVersion)
	}
	if snapshot.Status != states.ObjectReady {
		return fmt.Errorf("%s object status=%s, want ready", label, snapshot.Status)
	}
	if len(snapshot.AttrsJSON) == 0 || snapshot.AttrsFlat != nil {
		return fmt.Errorf("%s encoding JSON/flatmap=%d/%#v, want schema-v1 JSON only", label, len(snapshot.AttrsJSON), snapshot.AttrsFlat)
	}
	if len(snapshot.Private) == 0 {
		return fmt.Errorf("%s private state is empty, want deterministic sentinel", label)
	}
	if !bytes.Equal(snapshot.Private, privateStateSentinel) {
		return fmt.Errorf("%s private state=%q, want %q", label, snapshot.Private, privateStateSentinel)
	}
	wantDependencies := productionShapedDependencies()
	if len(snapshot.Dependencies) < 8 || len(snapshot.Dependencies) != len(wantDependencies) {
		return fmt.Errorf("%s dependencies=%d, want %d production-shaped addresses", label, len(snapshot.Dependencies), len(wantDependencies))
	}
	for i, want := range wantDependencies {
		if got := snapshot.Dependencies[i].String(); got != want.String() {
			return fmt.Errorf("%s dependency[%d]=%q, want %q", label, i, got, want.String())
		}
	}
	var attrs map[string]interface{}
	if err := json.Unmarshal(snapshot.AttrsJSON, &attrs); err != nil {
		return fmt.Errorf("%s decode attrs: %w", label, err)
	}
	if attrs["id"] != resourceID {
		return fmt.Errorf("%s id=%#v, want %q", label, attrs["id"], resourceID)
	}
	for name, value := range markerValues {
		if attrs[name] != value {
			return fmt.Errorf("%s marker %s=%#v, want %q", label, name, attrs[name], value)
		}
		encodedName, _ := json.Marshal(name)
		encodedValue, _ := json.Marshal(value)
		fragment := append(append(encodedName, ':'), encodedValue...)
		if !bytes.Contains(snapshot.AttrsJSON, fragment) {
			return fmt.Errorf("%s marker bytes %q are absent from state snapshot", label, fragment)
		}
	}
	return nil
}

func loadConfig(dir, content string) (*configs.Config, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o600); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}
	module, diags := configs.NewParser(afero.NewOsFs()).LoadConfigDir(dir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("load config: %s", diags.Error())
	}
	config, diags := configs.BuildConfig(module, nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("build config: %s", diags.Error())
	}
	return config, nil
}

func newContext(config *configs.Config, state *states.State, provider *schema.Provider) (*terraform.Context, error) {
	ctx, diags := terraform.NewContext(&terraform.ContextOpts{
		Config: config,
		State:  state,
		ProviderResolver: providers.ResolverFunc(func(reqd discovery.PluginRequirements) (map[string]providers.Factory, []error) {
			if len(reqd) != 1 || reqd["alicloud"] == nil {
				return nil, []error{fmt.Errorf("provider requirements=%#v, want exactly alicloud", reqd)}
			}
			got := reqd["alicloud"].Versions
			want := discovery.ConstraintStr("= 1.247.0").MustParse()
			if got.String() != want.String() || !got.Allows(discovery.VersionStr("1.247.0").MustParse()) {
				return nil, []error{fmt.Errorf("alicloud provider requirement=%q, want exact %q", got.String(), want.String())}
			}
			return map[string]providers.Factory{
				"alicloud": func() (providers.Interface, error) { return grpcProvider(provider) },
			}, nil
		}),
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("new context: %s", diags.Err())
	}
	return ctx, nil
}

func grpcProvider(provider *schema.Provider) (providers.Interface, error) {
	listener := bufconn.Listen(256 * 1024)
	server := grpc.NewServer()
	proto.RegisterProviderServer(server, plugin.NewGRPCProviderServerShim(provider))
	go server.Serve(listener)
	conn, err := grpc.Dial("", grpc.WithDialer(func(string, time.Duration) (net.Conn, error) { return listener.Dial() }), grpc.WithInsecure())
	if err != nil {
		server.Stop()
		return nil, fmt.Errorf("dial provider: %w", err)
	}
	var grpcPlugin tfplugin.GRPCProviderPlugin
	client, err := grpcPlugin.GRPCClient(context.Background(), nil, conn)
	if err != nil {
		server.Stop()
		return nil, fmt.Errorf("create provider client: %w", err)
	}
	grpcClient := client.(*tfplugin.GRPCProvider)
	grpcClient.TestServer = server
	return grpcClient, nil
}
`

const flinkWorkspaceCoreRaceEnabledProgram = `//go:build race

package main

const raceInstrumented = true
`

const flinkWorkspaceCoreRaceDisabledProgram = `//go:build !race

package main

const raceInstrumented = false
`
