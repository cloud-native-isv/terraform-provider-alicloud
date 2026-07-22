package alicloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	flinkTofuRuntimeProviderVersion = "0.0.1"
	flinkTofuRuntimeProviderSource  = "registry.opentofu.org/aliyun/alicloud"
)

type flinkTofuRuntimeLedger struct {
	ParentExists             bool   `json:"parent_exists"`
	ParentID                 string `json:"parent_id"`
	Capacity                 int    `json:"capacity"`
	ParentCreates            int    `json:"parent_creates"`
	ParentDeletes            int    `json:"parent_deletes"`
	AllocationAttached       bool   `json:"allocation_attached"`
	AllocationCreates        int    `json:"allocation_creates"`
	AllocationDeletes        int    `json:"allocation_deletes"`
	CapacityWrites           int    `json:"capacity_writes"`
	FailNextAllocationCreate bool   `json:"fail_next_allocation_create"`
}

type flinkTofuRuntimePlan struct {
	ResourceChanges []struct {
		Address      string `json:"address"`
		ActionReason string `json:"action_reason"`
		Change       struct {
			Actions      []string               `json:"actions"`
			Before       map[string]interface{} `json:"before"`
			After        map[string]interface{} `json:"after"`
			AfterUnknown map[string]interface{} `json:"after_unknown"`
			ReplacePaths []interface{}          `json:"replace_paths"`
		} `json:"change"`
	} `json:"resource_changes"`
}

type flinkTofuRuntimeGate struct {
	t      *testing.T
	ctx    context.Context
	dir    string
	env    []string
	ledger string
	binary string
	tfData string
}

type flinkTofuRuntimeCommandResult struct {
	output string
	err    error
}

type flinkTofuRuntimeOwnedEnvironment struct {
	path          string
	home          string
	xdgCache      string
	xdgConfig     string
	xdgData       string
	xdgState      string
	temp          string
	emptyCLI      string
	tfData        string
	pluginCache   string
	goCache       string
	goPath        string
	goModuleCache string
}

type flinkTofuRuntimeFixtureControl struct {
	readDelay    string
	readStarted  string
	readFinished string
}

func TestFlinkTofuRuntimeEnvironmentDoesNotInheritAmbientCapabilities(t *testing.T) {
	contamination := map[string]string{
		"FLINK_TOFU_RUNTIME_READ_DELAY": "bogus",
		"HTTP_PROXY":                    "http://127.0.0.1:1",
		"http_proxy":                    "http://127.0.0.1:1",
		"HTTPS_PROXY":                   "http://127.0.0.1:2",
		"https_proxy":                   "http://127.0.0.1:2",
		"ALL_PROXY":                     "socks5://127.0.0.1:3",
		"all_proxy":                     "socks5://127.0.0.1:3",
		"NO_PROXY":                      "internal.invalid",
		"no_proxy":                      "internal.invalid",
		"SSH_AUTH_SOCK":                 "/tmp/forbidden-ssh-agent",
		"ALICLOUD_ACCESS_KEY":           "forbidden-alicloud-credential",
		"ALIBABA_CLOUD_ACCESS_KEY_ID":   "forbidden-alibaba-credential",
		"AWS_ACCESS_KEY_ID":             "forbidden-aws-credential",
		"GITHUB_TOKEN":                  "forbidden-token",
		"RUNTIME_GATE_TEST_CREDENTIAL":  "forbidden-generic-credential",
	}
	for name, value := range contamination {
		t.Setenv(name, value)
	}

	owned := flinkTofuRuntimeCreateOwnedEnvironment(t, t.TempDir())
	env := flinkTofuRuntimeEnvironment(owned, owned.emptyCLI, owned.tfData, owned.pluginCache, flinkTofuRuntimeFixtureControl{})
	got := make(map[string]string, len(env))
	for _, entry := range env {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("malformed environment entry %q", entry)
		}
		got[name] = value
	}
	for name := range contamination {
		if value, exists := got[name]; exists {
			t.Fatalf("runtime environment inherited forbidden %s=%q", name, value)
		}
	}
	allowed := map[string]struct{}{
		"CHECKPOINT_DISABLE": {}, "HOME": {}, "LANG": {}, "LC_ALL": {}, "NO_COLOR": {}, "PATH": {}, "TMPDIR": {},
		"TF_CLI_CONFIG_FILE": {}, "TF_DATA_DIR": {}, "TF_IN_AUTOMATION": {}, "TF_PLUGIN_CACHE_DIR": {},
		"XDG_CACHE_HOME": {}, "XDG_CONFIG_HOME": {}, "XDG_DATA_HOME": {}, "XDG_STATE_HOME": {},
	}
	for name := range got {
		if _, exists := allowed[name]; !exists {
			t.Errorf("runtime environment contains non-allowlisted variable %q", name)
		}
	}
	for name, want := range map[string]string{
		"CHECKPOINT_DISABLE": "1", "HOME": owned.home, "LANG": "C.UTF-8", "LC_ALL": "C.UTF-8", "NO_COLOR": "1", "PATH": owned.path,
		"TMPDIR": owned.temp, "TF_CLI_CONFIG_FILE": owned.emptyCLI, "TF_DATA_DIR": owned.tfData, "TF_IN_AUTOMATION": "1",
		"TF_PLUGIN_CACHE_DIR": owned.pluginCache, "XDG_CACHE_HOME": owned.xdgCache, "XDG_CONFIG_HOME": owned.xdgConfig,
		"XDG_DATA_HOME": owned.xdgData, "XDG_STATE_HOME": owned.xdgState,
	} {
		if got[name] != want {
			t.Errorf("runtime environment %s = %q, want %q", name, got[name], want)
		}
	}

	goEnv := flinkTofuRuntimeGoEnvironment(owned)
	goGot := make(map[string]string, len(goEnv))
	for _, entry := range goEnv {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("malformed Go environment entry %q", entry)
		}
		goGot[name] = value
	}
	for name := range contamination {
		if value, exists := goGot[name]; exists {
			t.Errorf("Go build environment inherited forbidden %s=%q", name, value)
		}
	}
	goAllowed := map[string]struct{}{
		"CGO_ENABLED": {}, "GOCACHE": {}, "GOENV": {}, "GOMODCACHE": {}, "GOPATH": {}, "GOPROXY": {}, "GOSUMDB": {}, "GOWORK": {},
		"HOME": {}, "LANG": {}, "LC_ALL": {}, "PATH": {}, "TMPDIR": {}, "XDG_CACHE_HOME": {}, "XDG_CONFIG_HOME": {}, "XDG_DATA_HOME": {}, "XDG_STATE_HOME": {},
	}
	for name := range goGot {
		if _, exists := goAllowed[name]; !exists {
			t.Errorf("Go build environment contains non-allowlisted variable %q", name)
		}
	}
	for name, want := range map[string]string{
		"CGO_ENABLED": "0", "GOCACHE": owned.goCache, "GOENV": "off", "GOMODCACHE": owned.goModuleCache,
		"GOPATH": owned.goPath, "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off",
	} {
		if goGot[name] != want {
			t.Errorf("Go build environment %s = %q, want %q", name, goGot[name], want)
		}
	}
}

func TestFlinkWorkspaceCapacityAllocationOpenTofu110RuntimeRecovery(t *testing.T) {
	if os.Getenv("RUN_FLINK_TOFU_RUNTIME_GATE") != "1" {
		t.Skip("set RUN_FLINK_TOFU_RUNTIME_GATE=1 to run the OpenTofu runtime recovery gate")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	root := t.TempDir()
	owned := flinkTofuRuntimeCreateOwnedEnvironment(t, root)
	flinkTofuRuntimeRequireVersion(t, ctx, flinkTofuRuntimeEnvironment(owned, owned.emptyCLI, "", "", flinkTofuRuntimeFixtureControl{}))
	assertFlinkTofuRuntimeProductionContract(t)

	providerArtifact := flinkTofuRuntimeBuildProvider(t, ctx, root, owned)
	providerSHA := flinkTofuRuntimeSHA256(t, providerArtifact)
	mirrorRoot := filepath.Join(root, "provider-mirror")
	mirrorBinary := flinkTofuRuntimeCreateMirror(t, mirrorRoot, providerArtifact)
	if mirrorSHA := flinkTofuRuntimeSHA256(t, mirrorBinary); mirrorSHA != providerSHA {
		t.Fatalf("mirror provider SHA-256 = %s, want freshly built artifact SHA-256 %s", mirrorSHA, providerSHA)
	}

	configDir := filepath.Join(root, "configuration")
	tfDataDir := owned.tfData
	pluginCacheDir := owned.pluginCache
	for _, dir := range []string{configDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create invocation-owned directory %q: %v", dir, err)
		}
	}
	ledgerPath := filepath.Join(root, "ledger.json")
	flinkTofuRuntimeWriteLedger(t, ledgerPath, flinkTofuRuntimeLedger{
		Capacity:                 1,
		FailNextAllocationCreate: true,
	})
	flinkTofuRuntimeWriteConfig(t, configDir, ledgerPath, true)

	cliConfig := filepath.Join(root, "terraformrc")
	cliConfigBody := fmt.Sprintf(`provider_installation {
  filesystem_mirror {
    path    = %s
    include = [%q]
  }
}
`, strconv.Quote(mirrorRoot), flinkTofuRuntimeProviderSource)
	if err := os.WriteFile(cliConfig, []byte(cliConfigBody), 0o600); err != nil {
		t.Fatalf("write invocation-owned CLI config: %v", err)
	}

	gate := &flinkTofuRuntimeGate{
		t:      t,
		ctx:    ctx,
		dir:    configDir,
		ledger: ledgerPath,
		env:    flinkTofuRuntimeEnvironment(owned, cliConfig, tfDataDir, pluginCacheDir, flinkTofuRuntimeFixtureControl{}),
	}

	gate.mustTofu("init", "-input=false", "-no-color")
	installedProvider := flinkTofuRuntimeInstalledProvider(t, tfDataDir)
	if !flinkTofuRuntimePathWithin(root, installedProvider) {
		t.Fatalf("installed provider resolved outside invocation-owned root: root=%q provider=%q", root, installedProvider)
	}
	if installedSHA := flinkTofuRuntimeSHA256(t, installedProvider); installedSHA != providerSHA {
		t.Fatalf("installed provider %q SHA-256 = %s, want freshly built artifact SHA-256 %s", installedProvider, installedSHA, providerSHA)
	}

	failedApplyOutput, err := gate.tofu("apply", "-input=false", "-auto-approve", "-no-color")
	if err == nil {
		t.Fatalf("initial apply succeeded; want injected allocation failure\n%s", failedApplyOutput)
	}
	if !strings.Contains(failedApplyOutput, "injected failure after one successful capacity write") {
		t.Fatalf("initial apply error = %v, output missing injected allocation failure:\n%s", err, failedApplyOutput)
	}
	failed := gate.readLedger()
	assertFlinkTofuRuntimeLedger(t, failed, flinkTofuRuntimeLedger{
		ParentExists:       true,
		ParentID:           "f-tofu-runtime-parent",
		Capacity:           2,
		ParentCreates:      1,
		AllocationAttached: true,
		AllocationCreates:  1,
		CapacityWrites:     1,
	}, "failed apply")

	cancelStarted := filepath.Join(root, "cancel-read-started")
	cancelFinished := filepath.Join(root, "cancel-read-finished")
	cancelGate := *gate
	cancelGate.env = flinkTofuRuntimeEnvironment(owned, cliConfig, tfDataDir, pluginCacheDir, flinkTofuRuntimeFixtureControl{
		readDelay: "750ms", readStarted: cancelStarted, readFinished: cancelFinished,
	})
	cancelCtx, cancelPlan := context.WithCancel(ctx)
	cancelResult := make(chan flinkTofuRuntimeCommandResult, 1)
	go func() {
		output, err := cancelGate.tofuWithContext(cancelCtx, "plan", "-input=false", "-no-color", "-out="+filepath.Join(root, "canceled.tfplan"))
		cancelResult <- flinkTofuRuntimeCommandResult{output: output, err: err}
	}()
	if err := flinkTofuRuntimeWaitForPath(cancelStarted, 5*time.Second); err != nil {
		cancelPlan()
		result := <-cancelResult
		t.Fatalf("cancellation probe never reached external provider Read: %v; command error=%v\n%s", err, result.err, result.output)
	}
	cancelPlan()
	result := <-cancelResult
	if result.err == nil || !strings.Contains(result.err.Error(), context.Canceled.Error()) {
		t.Fatalf("provider-active plan cancellation error = %v, want context canceled\n%s", result.err, result.output)
	}
	if err := flinkTofuRuntimeWaitForPath(cancelFinished, 5*time.Second); err != nil {
		t.Fatalf("canceled provider Read did not finish its bounded cleanup: %v", err)
	}
	assertFlinkTofuRuntimeLedger(t, gate.readLedger(), failed, "canceled recovery plan")

	recoveryPlanPath := filepath.Join(root, "recovery.tfplan")
	gate.mustTofu("plan", "-input=false", "-no-color", "-out="+recoveryPlanPath)
	recoveryPlan := gate.readPlan(recoveryPlanPath)
	assertFlinkTofuRuntimePlan(t, recoveryPlan, map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"delete", "create"},
	})
	for _, change := range recoveryPlan.ResourceChanges {
		if change.Address == "alicloud_flink_workspace_capacity_allocation.this" && change.ActionReason != "replace_because_tainted" {
			t.Fatalf("allocation recovery action_reason = %q, want replace_because_tainted", change.ActionReason)
		}
	}

	gate.mustTofu("apply", "-input=false", "-no-color", recoveryPlanPath)
	recovered := gate.readLedger()
	assertFlinkTofuRuntimeLedger(t, recovered, flinkTofuRuntimeLedger{
		ParentExists:       true,
		ParentID:           "f-tofu-runtime-parent",
		Capacity:           3,
		ParentCreates:      1,
		AllocationAttached: true,
		AllocationCreates:  2,
		AllocationDeletes:  1,
		CapacityWrites:     2,
	}, "saved recovery apply")

	finalPlanPath := filepath.Join(root, "final.tfplan")
	gate.mustTofu("plan", "-input=false", "-no-color", "-out="+finalPlanPath)
	assertFlinkTofuRuntimePlan(t, gate.readPlan(finalPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"no-op"},
	})
	assertFlinkTofuRuntimeLedger(t, gate.readLedger(), recovered, "final no-op plan")

	flinkTofuRuntimeWriteConfig(t, configDir, ledgerPath, false)
	detachPlanPath := filepath.Join(root, "detach.tfplan")
	gate.mustTofu("plan", "-input=false", "-no-color", "-out="+detachPlanPath)
	assertFlinkTofuRuntimePlan(t, gate.readPlan(detachPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"delete"},
	})
	assertFlinkTofuRuntimeLedger(t, gate.readLedger(), recovered, "detach plan")

	gate.mustTofu("apply", "-input=false", "-no-color", detachPlanPath)
	detached := recovered
	detached.AllocationAttached = false
	detached.AllocationDeletes++
	assertFlinkTofuRuntimeLedger(t, gate.readLedger(), detached, "detach apply")

	parentOnlyPlanPath := filepath.Join(root, "parent-only.tfplan")
	gate.mustTofu("plan", "-input=false", "-no-color", "-out="+parentOnlyPlanPath)
	assertFlinkTofuRuntimePlan(t, gate.readPlan(parentOnlyPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent": {"no-op"},
	})
	assertFlinkTofuRuntimeLedger(t, gate.readLedger(), detached, "parent-only no-op plan")
}

func flinkTofuRuntimeRequireVersion(t *testing.T, ctx context.Context, env []string) {
	t.Helper()
	binary := flinkTofuRuntimeBinary()
	cmd := exec.CommandContext(ctx, binary, "version")
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("tofu version: %v", err)
	}
	firstLine := strings.SplitN(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n", 2)[0]
	expected := flinkTofuRuntimeExpectedVersion()
	if firstLine != expected {
		t.Fatalf("tofu version first line = %q, want %q (binary %q)", firstLine, expected, binary)
	}
}

func flinkTofuRuntimeBinary() string {
	if value := os.Getenv("FLINK_TOFU_RUNTIME_BIN"); value != "" {
		return value
	}
	return "tofu"
}

func flinkTofuRuntimeExpectedVersion() string {
	if value := os.Getenv("FLINK_TOFU_RUNTIME_EXPECT_VERSION"); value != "" {
		return value
	}
	return "OpenTofu v1.10.10"
}

func assertFlinkTofuRuntimeProductionContract(t *testing.T) {
	t.Helper()
	provider, ok := Provider().(*schema.Provider)
	if !ok || provider == nil {
		t.Fatalf("Provider() = %T, want *schema.Provider", Provider())
	}
	parent := provider.ResourcesMap["alicloud_flink_workspace"]
	allocation := provider.ResourcesMap["alicloud_flink_workspace_capacity_allocation"]
	if parent == nil || allocation == nil {
		t.Fatalf("production provider boundary missing: parent=%p allocation=%p", parent, allocation)
	}
	for name, callbacks := range map[string]struct {
		got  interface{}
		want interface{}
	}{
		"parent Create":     {parent.Create, resourceAliCloudFlinkWorkspaceCreate},
		"parent Delete":     {parent.Delete, resourceAliCloudFlinkWorkspaceDelete},
		"allocation Create": {allocation.Create, resourceAliCloudFlinkWorkspaceCapacityAllocationCreate},
		"allocation Read":   {allocation.Read, resourceAliCloudFlinkWorkspaceCapacityAllocationRead},
		"allocation Update": {allocation.Update, resourceAliCloudFlinkWorkspaceCapacityAllocationUpdate},
		"allocation Delete": {allocation.Delete, resourceAliCloudFlinkWorkspaceCapacityAllocationDelete},
	} {
		if callbacks.got == nil || callbacks.want == nil || reflect.ValueOf(callbacks.got).Pointer() != reflect.ValueOf(callbacks.want).Pointer() {
			t.Fatalf("production %s callback is not anchored to its named implementation", name)
		}
	}
	initialCapacity := parent.Schema["initial_capacity"]
	if initialCapacity == nil || initialCapacity.Type != schema.TypeList || !initialCapacity.Optional || initialCapacity.Required || initialCapacity.MaxItems != 1 {
		t.Fatalf("parent initial_capacity schema = %#v, want optional create-boundary block", initialCapacity)
	}
	for _, field := range []string{"workspace_instance_id", "fixed_cu", "cross_zone_fixed_cu", "max_cu_limit", "namespace", "implicit_topology_hash", "observed_capacity_tree"} {
		if _, exists := parent.Schema[field]; exists {
			t.Fatalf("parent production schema unexpectedly owns allocation field %q", field)
		}
	}
	workspaceID := allocation.Schema["workspace_instance_id"]
	if workspaceID == nil || workspaceID.Type != schema.TypeString || !workspaceID.Required || !workspaceID.ForceNew {
		t.Fatalf("allocation workspace_instance_id schema = %#v", workspaceID)
	}
	for _, field := range []string{"fixed_cu", "cross_zone_fixed_cu", "max_cu_limit"} {
		contract := allocation.Schema[field]
		if contract == nil || contract.Type != schema.TypeInt || !contract.Required {
			t.Fatalf("allocation %s schema = %#v", field, contract)
		}
	}
	namespace := allocation.Schema["namespace"]
	if namespace == nil || namespace.Type != schema.TypeSet || !namespace.Required || namespace.MinItems != 1 || namespace.Set == nil {
		t.Fatalf("allocation namespace schema = %#v, want authoritative schema.HashResource set", namespace)
	}
	for _, field := range []string{"resource_group_id", "vpc_id", "vswitch_ids", "storage", "charge_type", "initial_capacity"} {
		if _, exists := allocation.Schema[field]; exists {
			t.Fatalf("allocation production schema unexpectedly owns parent field %q", field)
		}
	}
	detachData := schema.TestResourceDataRaw(t, allocation.Schema, map[string]interface{}{
		"workspace_instance_id": "f-contract-anchor",
		"fixed_cu":              1,
		"cross_zone_fixed_cu":   0,
		"max_cu_limit":          1,
		"namespace": []interface{}{map[string]interface{}{
			"name": "default", "fixed_cu": 1, "max_cu_limit": 1,
		}},
	})
	detachData.SetId("f-contract-anchor")
	if err := allocation.Delete(detachData, struct{}{}); err != nil {
		t.Fatalf("production allocation Delete anchor: %v", err)
	}
	if detachData.Id() != "" {
		t.Fatalf("production allocation Delete retained ID %q, want state-only detach", detachData.Id())
	}
}

func flinkTofuRuntimeBuildProvider(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment) string {
	t.Helper()
	sdkDir := filepath.Join(owned.goModuleCache, "github.com/hashicorp/terraform-plugin-sdk@v1.17.2")
	if info, err := os.Stat(sdkDir); err != nil || !info.IsDir() {
		t.Fatalf("Plugin SDK v1.17.2 module directory %q is unavailable: %v", sdkDir, err)
	}

	moduleDir := filepath.Join(root, "provider-module")
	artifactDir := filepath.Join(root, "provider-artifact")
	for _, dir := range []string{moduleDir, artifactDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create provider build directory %q: %v", dir, err)
		}
	}
	goMod := fmt.Sprintf("module example.com/flink-tofu-runtime-provider\n\ngo 1.20\n\nrequire github.com/hashicorp/terraform-plugin-sdk v1.17.2\n\nreplace github.com/hashicorp/terraform-plugin-sdk => %s\n", filepath.ToSlash(sdkDir))
	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatalf("write disposable provider go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.go"), []byte(flinkTofuRuntimeFixtureProviderProgram), 0o600); err != nil {
		t.Fatalf("write disposable provider source: %v", err)
	}

	artifact := filepath.Join(artifactDir, "terraform-provider-alicloud_v"+flinkTofuRuntimeProviderVersion)
	build := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-mod=mod", "-o", artifact, ".")
	build.Dir = moduleDir
	build.Env = flinkTofuRuntimeGoEnvironment(owned)
	buildOutput, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("build disposable Plugin SDK v1.17.2 provider: %v\n%s", err, buildOutput)
	}
	if info, err := os.Stat(artifact); err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("built provider artifact %q is not executable: info=%v err=%v", artifact, info, err)
	}
	return artifact
}

func flinkTofuRuntimeCreateMirror(t *testing.T, mirrorRoot, artifact string) string {
	t.Helper()
	platform := runtime.GOOS + "_" + runtime.GOARCH
	platformDir := filepath.Join(mirrorRoot, "registry.opentofu.org", "aliyun", "alicloud", flinkTofuRuntimeProviderVersion, platform)
	if err := os.MkdirAll(platformDir, 0o700); err != nil {
		t.Fatalf("create invocation-owned provider mirror: %v", err)
	}
	body, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatalf("read freshly built provider artifact: %v", err)
	}
	mirrorBinary := filepath.Join(platformDir, filepath.Base(artifact))
	if err := os.WriteFile(mirrorBinary, body, 0o555); err != nil {
		t.Fatalf("write invocation-owned mirror provider: %v", err)
	}
	return mirrorBinary
}

func flinkTofuRuntimeInstalledProvider(t *testing.T, tfDataDir string) string {
	t.Helper()
	platformDir := filepath.Join(tfDataDir, "providers", "registry.opentofu.org", "aliyun", "alicloud", flinkTofuRuntimeProviderVersion, runtime.GOOS+"_"+runtime.GOARCH)
	entries, err := os.ReadDir(platformDir)
	if err != nil {
		t.Fatalf("read installed provider platform %q: %v", platformDir, err)
	}
	var installed []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "terraform-provider-alicloud_v"+flinkTofuRuntimeProviderVersion) {
			installed = append(installed, filepath.Join(platformDir, entry.Name()))
		}
	}
	if len(installed) != 1 {
		t.Fatalf("installed provider candidates under %q = %v, want exactly one", platformDir, installed)
	}
	resolved, err := filepath.EvalSymlinks(installed[0])
	if err != nil {
		t.Fatalf("resolve installed provider %q: %v", installed[0], err)
	}
	return resolved
}

func flinkTofuRuntimeWriteConfig(t *testing.T, dir, ledgerPath string, includeAllocation bool) {
	t.Helper()
	content := fmt.Sprintf(`terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "= %s"
    }
  }
}

provider "alicloud" {
  ledger_path = %s
}

resource "alicloud_flink_workspace" "parent" {
  name = "tofu-runtime-parent"
}
`, flinkTofuRuntimeProviderVersion, strconv.Quote(ledgerPath))
	if includeAllocation {
		content += `
resource "alicloud_flink_workspace_capacity_allocation" "this" {
  workspace_instance_id = alicloud_flink_workspace.parent.id
  desired_capacity      = 3
}
`
	}
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o600); err != nil {
		t.Fatalf("write OpenTofu fixture configuration: %v", err)
	}
}

func (g *flinkTofuRuntimeGate) tofu(args ...string) (string, error) {
	g.t.Helper()
	commandCtx, cancel := context.WithTimeout(g.ctx, 5*time.Minute)
	defer cancel()
	return g.tofuWithContext(commandCtx, args...)
}

func (g *flinkTofuRuntimeGate) tofuWithContext(ctx context.Context, args ...string) (string, error) {
	g.t.Helper()
	binary := g.binary
	if binary == "" {
		binary = flinkTofuRuntimeBinary()
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = g.dir
	cmd.Env = g.env
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(output), fmt.Errorf("tofu %s canceled: %w", strings.Join(args, " "), ctx.Err())
	}
	return string(output), err
}

func (g *flinkTofuRuntimeGate) mustTofu(args ...string) string {
	g.t.Helper()
	output, err := g.tofu(args...)
	if err != nil {
		g.t.Fatalf("tofu %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return output
}

func (g *flinkTofuRuntimeGate) readPlan(path string) flinkTofuRuntimePlan {
	g.t.Helper()
	output := g.mustTofu("show", "-json", path)
	var plan flinkTofuRuntimePlan
	if err := json.Unmarshal([]byte(output), &plan); err != nil {
		g.t.Fatalf("decode saved plan %q: %v", path, err)
	}
	return plan
}

func (g *flinkTofuRuntimeGate) readLedger() flinkTofuRuntimeLedger {
	g.t.Helper()
	return flinkTofuRuntimeReadLedger(g.t, g.ledger)
}

func assertFlinkTofuRuntimePlan(t *testing.T, plan flinkTofuRuntimePlan, want map[string][]string) {
	t.Helper()
	got := make(map[string][]string, len(plan.ResourceChanges))
	for _, change := range plan.ResourceChanges {
		got[change.Address] = change.Change.Actions
	}
	if !reflect.DeepEqual(got, want) {
		changed := make(map[string][]string)
		shape := make(map[string]string)
		for _, change := range plan.ResourceChanges {
			for key := range change.Change.Before {
				if _, exists := change.Change.After[key]; !exists {
					changed[change.Address] = append(changed[change.Address], key+"(removed)")
				}
			}
			for key, after := range change.Change.After {
				before, exists := change.Change.Before[key]
				if !exists || !reflect.DeepEqual(before, after) {
					changed[change.Address] = append(changed[change.Address], key)
				}
			}
			for key, unknown := range change.Change.AfterUnknown {
				if value, ok := unknown.(bool); ok && value {
					changed[change.Address] = append(changed[change.Address], key+"(unknown)")
				}
			}
			beforeContext, beforePresent := change.Change.Before["workspace_bootstrap_context"]
			afterContext, afterPresent := change.Change.After["workspace_bootstrap_context"]
			unknownContext, unknownPresent := change.Change.AfterUnknown["workspace_bootstrap_context"]
			shape[change.Address] = fmt.Sprintf("context before[present=%t nil=%t type=%T] after[present=%t nil=%t type=%T] unknown[present=%t type=%T] replace_paths=%d", beforePresent, beforeContext == nil, beforeContext, afterPresent, afterContext == nil, afterContext, unknownPresent, unknownContext, len(change.Change.ReplacePaths))
		}
		t.Fatalf("saved plan actions = %#v, want exactly %#v; changed top-level fields = %#v; safe plan shape = %#v", got, want, changed, shape)
	}
}

func flinkTofuRuntimeWriteLedger(t *testing.T, path string, ledger flinkTofuRuntimeLedger) {
	t.Helper()
	body, err := json.Marshal(ledger)
	if err != nil {
		t.Fatalf("encode fixture ledger: %v", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		t.Fatalf("write fixture ledger: %v", err)
	}
}

func flinkTofuRuntimeReadLedger(t *testing.T, path string) flinkTofuRuntimeLedger {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture ledger: %v", err)
	}
	var ledger flinkTofuRuntimeLedger
	if err := json.Unmarshal(body, &ledger); err != nil {
		t.Fatalf("decode fixture ledger: %v", err)
	}
	return ledger
}

func assertFlinkTofuRuntimeLedger(t *testing.T, got, want flinkTofuRuntimeLedger, phase string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s ledger = %+v, want %+v", phase, got, want)
	}
}

func flinkTofuRuntimeSHA256(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q for SHA-256: %v", path, err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func flinkTofuRuntimeWaitForPath(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %q", path)
}

func flinkTofuRuntimePathWithin(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func flinkTofuRuntimeCreateOwnedEnvironment(t *testing.T, root string) flinkTofuRuntimeOwnedEnvironment {
	t.Helper()
	path := os.Getenv("PATH")
	if path == "" {
		t.Fatal("PATH must not be empty for the OpenTofu runtime gate")
	}
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		currentUser, err := user.Current()
		if err != nil || currentUser.HomeDir == "" {
			t.Fatalf("resolve default GOPATH home: user=%#v err=%v", currentUser, err)
		}
		goPath = filepath.Join(currentUser.HomeDir, "go")
	}
	goPathEntries := filepath.SplitList(goPath)
	if len(goPathEntries) == 0 || goPathEntries[0] == "" {
		t.Fatalf("GOPATH %q has no usable entry", goPath)
	}
	for _, entry := range goPathEntries {
		if !filepath.IsAbs(entry) {
			t.Fatalf("GOPATH entry %q must be absolute", entry)
		}
	}
	goModuleCache := os.Getenv("GOMODCACHE")
	if goModuleCache == "" {
		goModuleCache = filepath.Join(goPathEntries[0], "pkg", "mod")
	}
	if !filepath.IsAbs(goModuleCache) {
		t.Fatalf("GOMODCACHE %q must be absolute", goModuleCache)
	}

	owned := flinkTofuRuntimeOwnedEnvironment{
		path: path, home: filepath.Join(root, "home"), xdgCache: filepath.Join(root, "xdg-cache"),
		xdgConfig: filepath.Join(root, "xdg-config"), xdgData: filepath.Join(root, "xdg-data"),
		xdgState: filepath.Join(root, "xdg-state"), temp: filepath.Join(root, "tmp"),
		emptyCLI: filepath.Join(root, "empty-terraformrc"), tfData: filepath.Join(root, "tf-data"),
		pluginCache: filepath.Join(root, "plugin-cache"), goCache: filepath.Join(root, "go-cache"),
		goPath: goPath, goModuleCache: goModuleCache,
	}
	for _, dir := range []string{owned.home, owned.xdgCache, owned.xdgConfig, owned.xdgData, owned.xdgState, owned.temp, owned.tfData, owned.pluginCache, owned.goCache} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create invocation-owned environment directory %q: %v", dir, err)
		}
	}
	if err := os.WriteFile(owned.emptyCLI, nil, 0o600); err != nil {
		t.Fatalf("write invocation-owned empty OpenTofu CLI config: %v", err)
	}
	return owned
}

func flinkTofuRuntimeEnvironment(owned flinkTofuRuntimeOwnedEnvironment, cliConfig, tfDataDir, pluginCacheDir string, control flinkTofuRuntimeFixtureControl) []string {
	env := []string{
		"PATH=" + owned.path,
		"HOME=" + owned.home,
		"XDG_CACHE_HOME=" + owned.xdgCache,
		"XDG_CONFIG_HOME=" + owned.xdgConfig,
		"XDG_DATA_HOME=" + owned.xdgData,
		"XDG_STATE_HOME=" + owned.xdgState,
		"TMPDIR=" + owned.temp,
		"TF_CLI_CONFIG_FILE=" + cliConfig,
		"TF_IN_AUTOMATION=1",
		"NO_COLOR=1",
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"CHECKPOINT_DISABLE=1",
	}
	if tfDataDir != "" {
		env = append(env, "TF_DATA_DIR="+tfDataDir)
	}
	if pluginCacheDir != "" {
		env = append(env, "TF_PLUGIN_CACHE_DIR="+pluginCacheDir)
	}
	if control.readDelay != "" {
		env = append(env,
			"TOFU_RUNTIME_FIXTURE_READ_DELAY="+control.readDelay,
			"TOFU_RUNTIME_FIXTURE_READ_STARTED="+control.readStarted,
			"TOFU_RUNTIME_FIXTURE_READ_FINISHED="+control.readFinished,
		)
	}
	return env
}

func flinkTofuRuntimeGoEnvironment(owned flinkTofuRuntimeOwnedEnvironment) []string {
	return []string{
		"PATH=" + owned.path,
		"HOME=" + owned.home,
		"XDG_CACHE_HOME=" + owned.xdgCache,
		"XDG_CONFIG_HOME=" + owned.xdgConfig,
		"XDG_DATA_HOME=" + owned.xdgData,
		"XDG_STATE_HOME=" + owned.xdgState,
		"TMPDIR=" + owned.temp,
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"GOENV=off",
		"GOCACHE=" + owned.goCache,
		"GOPATH=" + owned.goPath,
		"GOMODCACHE=" + owned.goModuleCache,
		"GOPROXY=off",
		"GOSUMDB=off",
		"GOWORK=off",
		"CGO_ENABLED=0",
	}
}

const flinkTofuRuntimeFixtureProviderProgram = `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type ledger struct {
	ParentExists             bool   ` + "`" + `json:"parent_exists"` + "`" + `
	ParentID                 string ` + "`" + `json:"parent_id"` + "`" + `
	Capacity                 int    ` + "`" + `json:"capacity"` + "`" + `
	ParentCreates            int    ` + "`" + `json:"parent_creates"` + "`" + `
	ParentDeletes            int    ` + "`" + `json:"parent_deletes"` + "`" + `
	AllocationAttached       bool   ` + "`" + `json:"allocation_attached"` + "`" + `
	AllocationCreates        int    ` + "`" + `json:"allocation_creates"` + "`" + `
	AllocationDeletes        int    ` + "`" + `json:"allocation_deletes"` + "`" + `
	CapacityWrites           int    ` + "`" + `json:"capacity_writes"` + "`" + `
	FailNextAllocationCreate bool   ` + "`" + `json:"fail_next_allocation_create"` + "`" + `
}

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: func() terraform.ResourceProvider { return provider() }})
}

func provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ledger_path": {Type: schema.TypeString, Required: true},
		},
		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			if err := assertCleanRuntimeEnvironment(); err != nil { return nil, err }
			path := d.Get("ledger_path").(string)
			if path == "" { return nil, fmt.Errorf("ledger_path must not be empty") }
			return path, nil
		},
		ResourcesMap: map[string]*schema.Resource{
			"alicloud_flink_workspace": parentResource(),
			"alicloud_flink_workspace_capacity_allocation": allocationResource(),
		},
	}
}

func parentResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{"name": {Type: schema.TypeString, Required: true, ForceNew: true}},
		Create: func(d *schema.ResourceData, meta interface{}) error {
			state, err := load(meta.(string)); if err != nil { return err }
			state.ParentCreates++
			state.ParentExists = true
			state.ParentID = "f-tofu-runtime-parent"
			d.SetId(state.ParentID)
			return save(meta.(string), state)
		},
		Read: func(d *schema.ResourceData, meta interface{}) error {
			if err := delayRuntimeRead(); err != nil { return err }
			state, err := load(meta.(string)); if err != nil { return err }
			if !state.ParentExists { d.SetId("") }
			return nil
		},
		Delete: func(d *schema.ResourceData, meta interface{}) error {
			state, err := load(meta.(string)); if err != nil { return err }
			state.ParentDeletes++
			state.ParentExists = false
			d.SetId("")
			return save(meta.(string), state)
		},
	}
}

func allocationResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"workspace_instance_id": {Type: schema.TypeString, Required: true, ForceNew: true},
			"desired_capacity": {Type: schema.TypeInt, Required: true},
			"observed_capacity": {Type: schema.TypeInt, Computed: true},
		},
		Create: func(d *schema.ResourceData, meta interface{}) error {
			state, err := load(meta.(string)); if err != nil { return err }
			state.AllocationCreates++
			state.AllocationAttached = true
			d.SetId(d.Get("workspace_instance_id").(string))
			desired := d.Get("desired_capacity").(int)
			if state.Capacity < desired {
				state.Capacity++
				state.CapacityWrites++
			}
			if err := d.Set("observed_capacity", state.Capacity); err != nil { return err }
			if state.FailNextAllocationCreate {
				state.FailNextAllocationCreate = false
				if err := save(meta.(string), state); err != nil { return err }
				return fmt.Errorf("injected failure after one successful capacity write")
			}
			return save(meta.(string), state)
		},
		Read: func(d *schema.ResourceData, meta interface{}) error {
			if err := delayRuntimeRead(); err != nil { return err }
			state, err := load(meta.(string)); if err != nil { return err }
			if !state.AllocationAttached { d.SetId(""); return nil }
			return d.Set("observed_capacity", state.Capacity)
		},
		Update: func(d *schema.ResourceData, meta interface{}) error {
			state, err := load(meta.(string)); if err != nil { return err }
			desired := d.Get("desired_capacity").(int)
			if state.Capacity != desired { state.Capacity = desired; state.CapacityWrites++ }
			if err := d.Set("observed_capacity", state.Capacity); err != nil { return err }
			return save(meta.(string), state)
		},
		Delete: func(d *schema.ResourceData, meta interface{}) error {
			state, err := load(meta.(string)); if err != nil { return err }
			state.AllocationDeletes++
			state.AllocationAttached = false
			d.SetId("")
			return save(meta.(string), state)
		},
	}
}

func assertCleanRuntimeEnvironment() error {
	for _, name := range []string{
		"FLINK_TOFU_RUNTIME_READ_DELAY", "FLINK_TOFU_RUNTIME_READ_STARTED", "FLINK_TOFU_RUNTIME_READ_FINISHED",
		"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy", "NO_PROXY", "no_proxy",
		"SSH_AUTH_SOCK", "ALICLOUD_ACCESS_KEY", "ALIBABA_CLOUD_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID",
		"GITHUB_TOKEN", "RUNTIME_GATE_TEST_CREDENTIAL",
	} {
		if value, exists := os.LookupEnv(name); exists {
			return fmt.Errorf("fixture inherited forbidden environment variable %s (length %d)", name, len(value))
		}
	}
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, "FLINK_TOFU_RUNTIME_") {
			return fmt.Errorf("fixture inherited forbidden environment variable %s", name)
		}
	}
	return nil
}

func delayRuntimeRead() error {
	delayText := os.Getenv("TOFU_RUNTIME_FIXTURE_READ_DELAY")
	if delayText == "" { return nil }
	delay, err := time.ParseDuration(delayText)
	if err != nil { return fmt.Errorf("parse runtime read delay: %w", err) }
	started := os.Getenv("TOFU_RUNTIME_FIXTURE_READ_STARTED")
	finished := os.Getenv("TOFU_RUNTIME_FIXTURE_READ_FINISHED")
	if started == "" || finished == "" { return fmt.Errorf("runtime read delay markers must not be empty") }
	if err := os.WriteFile(started, []byte("started\n"), 0600); err != nil { return fmt.Errorf("write runtime read start marker: %w", err) }
	time.Sleep(delay)
	if err := os.WriteFile(finished, []byte("finished\n"), 0600); err != nil { return fmt.Errorf("write runtime read finish marker: %w", err) }
	return nil
}

func load(path string) (ledger, error) {
	body, err := os.ReadFile(path)
	if err != nil { return ledger{}, fmt.Errorf("read ledger: %w", err) }
	var state ledger
	if err := json.Unmarshal(body, &state); err != nil { return ledger{}, fmt.Errorf("decode ledger: %w", err) }
	return state, nil
}

func save(path string, state ledger) error {
	body, err := json.Marshal(state)
	if err != nil { return fmt.Errorf("encode ledger: %w", err) }
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".ledger-*")
	if err != nil { return fmt.Errorf("create ledger temp: %w", err) }
	tempPath := temp.Name()
	committed := false
	defer func() { if !committed { os.Remove(tempPath) } }()
	if err := temp.Chmod(0600); err != nil { temp.Close(); return fmt.Errorf("chmod ledger temp: %w", err) }
	if _, err := temp.Write(append(body, '\n')); err != nil { temp.Close(); return fmt.Errorf("write ledger temp: %w", err) }
	if err := temp.Sync(); err != nil { temp.Close(); return fmt.Errorf("sync ledger temp: %w", err) }
	if err := temp.Close(); err != nil { return fmt.Errorf("close ledger temp: %w", err) }
	if err := os.Rename(tempPath, path); err != nil { return fmt.Errorf("commit ledger: %w", err) }
	committed = true
	directory, err := os.Open(dir)
	if err != nil { return fmt.Errorf("open ledger directory: %w", err) }
	defer directory.Close()
	if err := directory.Sync(); err != nil { return fmt.Errorf("sync ledger directory: %w", err) }
	return nil
}
`
