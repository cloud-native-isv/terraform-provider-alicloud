package alicloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const flinkLegacyRuntimeLedgerEnv = "FLINK_LEGACY_NAMESPACE_RUNTIME_LEDGER"

const flinkLegacyTofuRuntimeLedgerLockSuffix = ".lock"

type flinkLegacyTofuRuntimeLedger struct {
	ParentExists              bool   `json:"parent_exists"`
	ParentID                  string `json:"parent_id"`
	ParentCreates             int    `json:"parent_creates"`
	ParentDeletes             int    `json:"parent_deletes"`
	ParentRefunds             int    `json:"parent_refunds"`
	ParentCreateToken         string `json:"parent_create_token"`
	ParentCreateFingerprint   string `json:"parent_create_fingerprint"`
	ChildCreated              bool   `json:"child_created"`
	ChildCreates              int    `json:"child_creates"`
	ChildDeletes              int    `json:"child_deletes"`
	ChildPostCreateReadErrors int    `json:"child_post_create_read_errors"`
	DataServiceFactories      int    `json:"data_service_factories"`
	ChildServiceFactories     int    `json:"child_service_factories"`
	WorkspacePolls            int    `json:"workspace_polls"`
	DefaultNamespacePolls     int    `json:"default_namespace_polls"`
	DefaultQueuePolls         int    `json:"default_queue_polls"`
	ChildNamespacePolls       int    `json:"child_namespace_polls"`
	ChildQueuePolls           int    `json:"child_queue_polls"`
	UnrelatedQueuePolls       int    `json:"unrelated_queue_polls"`
	ErrorObservations         int    `json:"error_observations"`
	ErrorMode                 string `json:"error_mode"`
}

type flinkLegacyTofuRuntimeOutput struct {
	Value interface{} `json:"value"`
}

type flinkLegacyTofuRuntimeFixture struct {
	gate   *flinkTofuRuntimeGate
	ledger string
}

func TestFlinkWorkspaceLegacyNamespacesOpenTofu110ProductionCallbacks(t *testing.T) {
	if os.Getenv("RUN_FLINK_TOFU_RUNTIME_GATE") != "1" {
		t.Skip("set RUN_FLINK_TOFU_RUNTIME_GATE=1 to run the OpenTofu legacy namespace production-callback gate")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	root := t.TempDir()
	owned := flinkTofuRuntimeCreateOwnedEnvironment(t, root)
	flinkTofuRuntimeRequireVersion(t, ctx, flinkTofuRuntimeEnvironment(owned, owned.emptyCLI, "", "", flinkTofuRuntimeFixtureControl{}))
	assertFlinkLegacyTofuRuntimeProductionContract(t)

	providerArtifact := flinkLegacyTofuRuntimeBuildProductionCallbackProvider(t, ctx, root, owned)
	providerSHA := flinkTofuRuntimeSHA256(t, providerArtifact)
	mirrorRoot := filepath.Join(root, "provider-mirror")
	mirrorBinary := flinkTofuRuntimeCreateMirror(t, mirrorRoot, providerArtifact)
	if mirrorSHA := flinkTofuRuntimeSHA256(t, mirrorBinary); mirrorSHA != providerSHA {
		t.Fatalf("legacy mirror provider SHA-256 = %s, want freshly built artifact SHA-256 %s", mirrorSHA, providerSHA)
	}
	cliConfig := filepath.Join(root, "terraformrc")
	cliConfigBody := fmt.Sprintf(`provider_installation {
  filesystem_mirror {
    path    = %s
    include = [%q]
  }
}
`, strconv.Quote(mirrorRoot), flinkTofuRuntimeProviderSource)
	if err := os.WriteFile(cliConfig, []byte(cliConfigBody), 0o600); err != nil {
		t.Fatalf("write invocation-owned legacy CLI config: %v", err)
	}

	t.Run("delayed bootstrap queue and child recovery converge once", func(t *testing.T) {
		fixture := newFlinkLegacyTofuRuntimeFixture(t, ctx, root, owned, cliConfig, "delayed", "")
		fixture.gate.mustTofu("apply", "-input=false", "-auto-approve", "-no-color")
		ledger := readFlinkLegacyTofuRuntimeLedger(t, fixture.ledger)
		assertFlinkLegacyTofuRuntimeParentInvariant(t, ledger, "delayed apply")
		if !ledger.ChildCreated || ledger.ChildCreates != 1 || ledger.ChildDeletes != 0 || ledger.ChildPostCreateReadErrors != 1 {
			t.Fatalf("delayed child lifecycle = created:%v create/delete/post-read-error:%d/%d/%d, want true 1/0/1", ledger.ChildCreated, ledger.ChildCreates, ledger.ChildDeletes, ledger.ChildPostCreateReadErrors)
		}
		if ledger.DataServiceFactories == 0 || ledger.ChildServiceFactories == 0 {
			t.Fatalf("production callback service factories data/child = %d/%d, want both invoked", ledger.DataServiceFactories, ledger.ChildServiceFactories)
		}
		if ledger.WorkspacePolls < 4 || ledger.DefaultNamespacePolls < 2 || ledger.DefaultQueuePolls < 2 || ledger.ChildNamespacePolls < 2 || ledger.ChildQueuePolls < 2 {
			t.Fatalf("delayed readiness polls workspace/default-ns/default-queue/child-ns/child-queue = %d/%d/%d/%d/%d, want every delayed layer polled", ledger.WorkspacePolls, ledger.DefaultNamespacePolls, ledger.DefaultQueuePolls, ledger.ChildNamespacePolls, ledger.ChildQueuePolls)
		}
		if ledger.UnrelatedQueuePolls != 0 {
			t.Fatalf("unrelated namespace queue polls = %d, want zero", ledger.UnrelatedQueuePolls)
		}
		assertFlinkLegacyTofuRuntimeOutputs(t, fixture.gate)

		planPath := filepath.Join(root, "delayed-final.tfplan")
		fixture.gate.mustTofu("plan", "-input=false", "-no-color", "-out="+planPath)
		assertFlinkLegacyTofuRuntimeNoReplayPlan(t, fixture.gate.readPlan(planPath))
		assertFlinkLegacyTofuRuntimeParentInvariant(t, readFlinkLegacyTofuRuntimeLedger(t, fixture.ledger), "delayed no-op plan")
	})

	t.Run("canceled dependent apply recovers without parent replay", func(t *testing.T) {
		fixture := newFlinkLegacyTofuRuntimeFixture(t, ctx, root, owned, cliConfig, "cancel", "cancel")
		cancelCtx, cancelApply := context.WithCancel(ctx)
		result := make(chan flinkTofuRuntimeCommandResult, 1)
		go func() {
			output, err := fixture.gate.tofuWithContext(cancelCtx, "apply", "-input=false", "-auto-approve", "-no-color")
			result <- flinkTofuRuntimeCommandResult{output: output, err: err}
		}()
		if err := waitForFlinkLegacyRuntimeLedgerCondition(fixture.ledger, 5*time.Second, func(state flinkLegacyTofuRuntimeLedger) bool {
			return state.ParentExists && state.WorkspacePolls > 0
		}); err != nil {
			cancelApply()
			command := <-result
			t.Fatalf("cancellation probe did not reach production dependent waiter: %v; command error=%v\n%s", err, command.err, command.output)
		}
		cancelApply()
		command := <-result
		if command.err == nil || !strings.Contains(command.err.Error(), context.Canceled.Error()) {
			t.Fatalf("canceled production dependent apply error = %v, want context canceled\n%s", command.err, command.output)
		}
		canceled, err := waitForFlinkLegacyRuntimeLedgerStable(fixture.ledger, 7*time.Second, 300*time.Millisecond)
		if err != nil {
			t.Fatalf("canceled provider tree did not stop by the callback deadline: %v", err)
		}
		assertFlinkLegacyTofuRuntimeParentInvariant(t, canceled, "canceled dependent apply")
		if canceled.ChildCreates != 0 || canceled.ChildDeletes != 0 {
			t.Fatalf("canceled dependent child create/delete = %d/%d, want 0/0", canceled.ChildCreates, canceled.ChildDeletes)
		}

		canceled = mutateFlinkLegacyTofuRuntimeLedger(t, fixture.ledger, func(state *flinkLegacyTofuRuntimeLedger) {
			state.ErrorMode = ""
		})
		fixture.gate.mustTofu("apply", "-input=false", "-auto-approve", "-no-color")
		recovered := readFlinkLegacyTofuRuntimeLedger(t, fixture.ledger)
		assertFlinkLegacyTofuRuntimeParentInvariant(t, recovered, "canceled dependent retry apply")
		if !recovered.ChildCreated || recovered.ChildCreates != 1 || recovered.ChildPostCreateReadErrors != 1 {
			t.Fatalf("canceled dependent retry child lifecycle = created:%v create/post-read-error:%d/%d, want true 1/1", recovered.ChildCreated, recovered.ChildCreates, recovered.ChildPostCreateReadErrors)
		}
	})

	for _, mode := range []string{"permission", "business", "terminal", "timeout"} {
		t.Run(mode+" dependent failure recovers without parent replay", func(t *testing.T) {
			fixture := newFlinkLegacyTofuRuntimeFixture(t, ctx, root, owned, cliConfig, mode, mode)
			output, err := fixture.gate.tofu("apply", "-input=false", "-auto-approve", "-no-color")
			wantFailure := map[string]string{
				"permission": "injected dependent permission",
				"business":   "injected dependent business",
				"terminal":   "terminal state",
				"timeout":    "timed out waiting",
			}[mode]
			if err == nil || !strings.Contains(output, wantFailure) {
				t.Fatalf("initial %s apply error = %v, output missing dependent failure:\n%s", mode, err, output)
			}
			failed := readFlinkLegacyTofuRuntimeLedger(t, fixture.ledger)
			assertFlinkLegacyTofuRuntimeParentInvariant(t, failed, mode+" failed apply")
			if failed.ChildCreates != 0 || failed.ChildDeletes != 0 || failed.ErrorObservations == 0 {
				t.Fatalf("%s failed apply child create/delete/error observations = %d/%d/%d, want 0/0/>0", mode, failed.ChildCreates, failed.ChildDeletes, failed.ErrorObservations)
			}

			failed = mutateFlinkLegacyTofuRuntimeLedger(t, fixture.ledger, func(state *flinkLegacyTofuRuntimeLedger) {
				state.ErrorMode = ""
			})
			fixture.gate.mustTofu("apply", "-input=false", "-auto-approve", "-no-color")
			recovered := readFlinkLegacyTofuRuntimeLedger(t, fixture.ledger)
			assertFlinkLegacyTofuRuntimeParentInvariant(t, recovered, mode+" retry apply")
			if !recovered.ChildCreated || recovered.ChildCreates != 1 || recovered.ChildDeletes != 0 || recovered.ChildPostCreateReadErrors != 1 {
				t.Fatalf("%s retry child lifecycle = created:%v create/delete/post-read-error:%d/%d/%d, want true 1/0/1", mode, recovered.ChildCreated, recovered.ChildCreates, recovered.ChildDeletes, recovered.ChildPostCreateReadErrors)
			}
			planPath := filepath.Join(root, mode+"-recovered.tfplan")
			fixture.gate.mustTofu("plan", "-input=false", "-no-color", "-out="+planPath)
			assertFlinkLegacyTofuRuntimeNoReplayPlan(t, fixture.gate.readPlan(planPath))
		})
	}
}

func assertFlinkLegacyTofuRuntimeProductionContract(t *testing.T) {
	t.Helper()
	provider, ok := Provider().(*schema.Provider)
	if !ok || provider == nil {
		t.Fatalf("Provider() = %T, want *schema.Provider", Provider())
	}
	parent := provider.ResourcesMap["alicloud_flink_workspace"]
	child := provider.ResourcesMap["alicloud_flink_namespace"]
	namespaces := provider.DataSourcesMap["alicloud_flink_namespaces"]
	if parent == nil || child == nil || namespaces == nil {
		t.Fatalf("production legacy graph boundary missing: parent=%p namespaces=%p child=%p", parent, namespaces, child)
	}
	for name, callbacks := range map[string]struct {
		got  interface{}
		want interface{}
	}{
		"parent Create":          {parent.Create, resourceAliCloudFlinkWorkspaceCreate},
		"namespaces data Read":   {namespaces.Read, dataSourceAliCloudFlinkNamespacesRead},
		"namespace child Create": {child.Create, resourceAliCloudFlinkNamespaceCreate},
	} {
		if callbacks.got == nil || callbacks.want == nil || reflect.ValueOf(callbacks.got).Pointer() != reflect.ValueOf(callbacks.want).Pointer() {
			t.Fatalf("production %s callback is not anchored to its named implementation", name)
		}
	}
}

func flinkLegacyTofuRuntimeBuildProductionCallbackProvider(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve legacy runtime test source path")
	}
	repoRoot := filepath.Dir(filepath.Dir(sourceFile))
	moduleDir := filepath.Join(root, "legacy-provider-main")
	artifactDir := filepath.Join(root, "legacy-provider-artifact")
	for _, dir := range []string{moduleDir, artifactDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create legacy provider build directory %q: %v", dir, err)
		}
	}
	mainSource := `package main

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: alicloud.FlinkLegacyNamespaceRuntimeFixtureProvider})
}
`
	mainPath := filepath.Join(moduleDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainSource), 0o600); err != nil {
		t.Fatalf("write legacy production-callback provider main: %v", err)
	}
	artifact := filepath.Join(artifactDir, "terraform-provider-alicloud_v"+flinkTofuRuntimeProviderVersion)
	build := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-mod=readonly", "-tags=flink_legacy_namespace_runtime_fixture", "-o", artifact, mainPath)
	build.Dir = repoRoot
	build.Env = flinkTofuRuntimeGoEnvironment(owned)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build real Provider() legacy production-callback fixture: %v\n%s", err, output)
	}
	return artifact
}

func newFlinkLegacyTofuRuntimeFixture(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment, cliConfig, name, errorMode string) flinkLegacyTofuRuntimeFixture {
	t.Helper()
	configDir := filepath.Join(root, name+"-configuration")
	tfDataDir := filepath.Join(root, name+"-tf-data")
	pluginCacheDir := filepath.Join(root, name+"-plugin-cache")
	for _, dir := range []string{configDir, tfDataDir, pluginCacheDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create legacy runtime directory %q: %v", dir, err)
		}
	}
	ledgerPath := filepath.Join(root, name+"-ledger.json")
	createFlinkLegacyTofuRuntimeLedgerLock(t, ledgerPath)
	writeFlinkLegacyTofuRuntimeLedger(t, ledgerPath, flinkLegacyTofuRuntimeLedger{ErrorMode: errorMode})
	writeFlinkLegacyTofuRuntimeConfig(t, configDir)
	env := flinkTofuRuntimeEnvironment(owned, cliConfig, tfDataDir, pluginCacheDir, flinkTofuRuntimeFixtureControl{})
	env = append(env, flinkLegacyRuntimeLedgerEnv+"="+ledgerPath)
	gate := &flinkTofuRuntimeGate{t: t, ctx: ctx, dir: configDir, ledger: ledgerPath, env: env}
	gate.mustTofu("init", "-input=false", "-no-color")
	installedProvider := flinkTofuRuntimeInstalledProvider(t, tfDataDir)
	if !flinkTofuRuntimePathWithin(root, installedProvider) {
		t.Fatalf("legacy installed provider resolved outside invocation-owned root: root=%q provider=%q", root, installedProvider)
	}
	return flinkLegacyTofuRuntimeFixture{gate: gate, ledger: ledgerPath}
}

func writeFlinkLegacyTofuRuntimeConfig(t *testing.T, dir string) {
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
  region = "cn-test"
}

resource "alicloud_flink_workspace" "parent" {
  name              = "legacy-runtime-parent"
  resource_group_id = "rg-fixture"
  vpc_id            = "vpc-fixture"
  vswitch_ids       = ["vsw-fixture"]
  charge_type       = "PRE"

  resource {
    cpu    = 2
    memory = 8
  }

  storage {
    oss_bucket = "fixture-bucket"
  }
}

data "alicloud_flink_namespaces" "all" {
  workspace_id = alicloud_flink_workspace.parent.id
}

resource "alicloud_flink_namespace" "child" {
  workspace_id   = alicloud_flink_workspace.parent.id
  namespace_name = "analytics"
}

output "all_namespaces" {
  value = data.alicloud_flink_namespaces.all.namespaces
}

output "child_id" {
  value = alicloud_flink_namespace.child.id
}

output "child_status" {
  value = alicloud_flink_namespace.child.status
}
`, flinkTofuRuntimeProviderVersion)
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o600); err != nil {
		t.Fatalf("write legacy OpenTofu fixture configuration: %v", err)
	}
}

func writeFlinkLegacyTofuRuntimeLedger(t *testing.T, path string, ledger flinkLegacyTofuRuntimeLedger) {
	t.Helper()
	if err := flinkLegacyTofuRuntimeWithLedgerLock(path, true, func() error {
		return flinkLegacyTofuRuntimeSaveLedger(path, ledger)
	}); err != nil {
		t.Fatalf("write legacy fixture ledger: %v", err)
	}
}

func mutateFlinkLegacyTofuRuntimeLedger(t *testing.T, path string, mutate func(*flinkLegacyTofuRuntimeLedger)) flinkLegacyTofuRuntimeLedger {
	t.Helper()
	var ledger flinkLegacyTofuRuntimeLedger
	if err := flinkLegacyTofuRuntimeWithLedgerLock(path, true, func() error {
		var err error
		ledger, err = flinkLegacyTofuRuntimeLoadLedger(path)
		if err != nil {
			return err
		}
		mutate(&ledger)
		return flinkLegacyTofuRuntimeSaveLedger(path, ledger)
	}); err != nil {
		t.Fatalf("mutate legacy fixture ledger: %v", err)
	}
	return ledger
}

func readFlinkLegacyTofuRuntimeLedger(t *testing.T, path string) flinkLegacyTofuRuntimeLedger {
	t.Helper()
	ledger, err := flinkLegacyTofuRuntimeReadLedger(path)
	if err != nil {
		t.Fatalf("read legacy fixture ledger: %v", err)
	}
	return ledger
}

func waitForFlinkLegacyRuntimeLedgerCondition(path string, timeout time.Duration, ready func(flinkLegacyTofuRuntimeLedger) bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ledger, err := flinkLegacyTofuRuntimeReadLedger(path)
		if err == nil && ready(ledger) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for legacy runtime ledger condition")
}

func waitForFlinkLegacyRuntimeLedgerStable(path string, timeout, quietPeriod time.Duration) (flinkLegacyTofuRuntimeLedger, error) {
	deadline := time.Now().Add(timeout)
	var previous flinkLegacyTofuRuntimeLedger
	var stableSince time.Time
	for time.Now().Before(deadline) {
		current, err := flinkLegacyTofuRuntimeReadLedger(path)
		if err == nil {
			if reflect.DeepEqual(current, previous) {
				if stableSince.IsZero() {
					stableSince = time.Now()
				}
				if time.Since(stableSince) >= quietPeriod {
					return current, nil
				}
			} else {
				previous = current
				stableSince = time.Time{}
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	return previous, fmt.Errorf("ledger did not remain stable for %s within %s", quietPeriod, timeout)
}

func createFlinkLegacyTofuRuntimeLedgerLock(t *testing.T, path string) {
	t.Helper()
	lockPath := path + flinkLegacyTofuRuntimeLedgerLockSuffix
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("create legacy fixture ledger lock: %v", err)
	}
	if err := lockFile.Close(); err != nil {
		t.Fatalf("close new legacy fixture ledger lock: %v", err)
	}
}

func flinkLegacyTofuRuntimeReadLedger(path string) (flinkLegacyTofuRuntimeLedger, error) {
	var ledger flinkLegacyTofuRuntimeLedger
	err := flinkLegacyTofuRuntimeWithLedgerLock(path, false, func() error {
		var err error
		ledger, err = flinkLegacyTofuRuntimeLoadLedger(path)
		return err
	})
	return ledger, err
}

func flinkLegacyTofuRuntimeLoadLedger(path string) (flinkLegacyTofuRuntimeLedger, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return flinkLegacyTofuRuntimeLedger{}, fmt.Errorf("read ledger: %w", err)
	}
	var ledger flinkLegacyTofuRuntimeLedger
	if err := json.Unmarshal(body, &ledger); err != nil {
		return ledger, fmt.Errorf("decode ledger: %w", err)
	}
	return ledger, nil
}

func flinkLegacyTofuRuntimeSaveLedger(path string, ledger flinkLegacyTofuRuntimeLedger) error {
	body, err := json.Marshal(ledger)
	if err != nil {
		return fmt.Errorf("encode ledger: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".flink-legacy-runtime-test-ledger-*")
	if err != nil {
		return fmt.Errorf("create temporary ledger: %w", err)
	}
	tempPath := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("secure temporary ledger: %w", err)
	}
	if _, err := temp.Write(append(body, '\n')); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary ledger: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary ledger: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary ledger: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace ledger: %w", err)
	}
	committed = true
	return nil
}

func flinkLegacyTofuRuntimeWithLedgerLock(path string, exclusive bool, operation func() error) (err error) {
	lockPath := path + flinkLegacyTofuRuntimeLedgerLockSuffix
	lockFile, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open ledger lock: %w", err)
	}
	defer func() {
		if closeErr := lockFile.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close ledger lock: %w", closeErr)
		}
	}()
	openedInfo, err := lockFile.Stat()
	if err != nil {
		return fmt.Errorf("inspect opened ledger lock: %w", err)
	}
	pathInfo, err := os.Lstat(lockPath)
	if err != nil {
		return fmt.Errorf("inspect ledger lock path: %w", err)
	}
	if !openedInfo.Mode().IsRegular() || openedInfo.Mode().Perm() != 0o600 || !pathInfo.Mode().IsRegular() || !os.SameFile(openedInfo, pathInfo) {
		return fmt.Errorf("ledger lock must be a stable mode 0600 regular file")
	}
	if err := flinkLegacyTofuRuntimeLockFile(lockFile, exclusive); err != nil {
		return fmt.Errorf("lock ledger: %w", err)
	}
	defer func() {
		if unlockErr := flinkLegacyTofuRuntimeUnlockFile(lockFile); err == nil && unlockErr != nil {
			err = fmt.Errorf("unlock ledger: %w", unlockErr)
		}
	}()
	return operation()
}

func assertFlinkLegacyTofuRuntimeParentInvariant(t *testing.T, ledger flinkLegacyTofuRuntimeLedger, phase string) {
	t.Helper()
	if !ledger.ParentExists || ledger.ParentID != "f-legacy-runtime-parent" || ledger.ParentCreates != 1 || ledger.ParentDeletes != 0 || ledger.ParentRefunds != 0 {
		t.Fatalf("%s paid parent = exists:%v id:%q create/delete/refund:%d/%d/%d, want true/f-legacy-runtime-parent/1/0/0", phase, ledger.ParentExists, ledger.ParentID, ledger.ParentCreates, ledger.ParentDeletes, ledger.ParentRefunds)
	}
}

func assertFlinkLegacyTofuRuntimeOutputs(t *testing.T, gate *flinkTofuRuntimeGate) {
	t.Helper()
	output := gate.mustTofu("output", "-json")
	var outputs map[string]flinkLegacyTofuRuntimeOutput
	if err := json.Unmarshal([]byte(output), &outputs); err != nil {
		t.Fatalf("decode legacy fixture outputs: %v", err)
	}
	namespaces, ok := outputs["all_namespaces"].Value.([]interface{})
	if !ok || len(namespaces) < 3 {
		t.Fatalf("unfiltered namespace output = %#v, want bootstrap plus two unrelated snapshots", outputs["all_namespaces"].Value)
	}
	wantStatuses := map[string]string{
		"legacy-runtime-parent-default": "SUCCESS",
		"failed-unrelated":              "FAILED",
		"creating-sibling":              "CREATING",
	}
	for _, raw := range namespaces {
		namespace, _ := raw.(map[string]interface{})
		name, _ := namespace["name"].(string)
		if want, exists := wantStatuses[name]; exists {
			if namespace["status"] != want {
				t.Fatalf("unfiltered namespace %q status = %#v, want %q", name, namespace["status"], want)
			}
			delete(wantStatuses, name)
		}
	}
	if len(wantStatuses) != 0 {
		t.Fatalf("unfiltered namespace output missing snapshots %#v: %#v", wantStatuses, namespaces)
	}
	if outputs["child_id"].Value != "f-legacy-runtime-parent:analytics" || outputs["child_status"].Value != "SUCCESS" {
		t.Fatalf("child outputs id/status = %#v/%#v, want usable child", outputs["child_id"].Value, outputs["child_status"].Value)
	}
}

func assertFlinkLegacyTofuRuntimeNoReplayPlan(t *testing.T, plan flinkTofuRuntimePlan) {
	t.Helper()
	want := map[string][]string{
		"alicloud_flink_workspace.parent": {"no-op"},
		"alicloud_flink_namespace.child":  {"no-op"},
	}
	got := make(map[string][]string)
	for _, change := range plan.ResourceChanges {
		if _, tracked := want[change.Address]; tracked {
			got[change.Address] = change.Change.Actions
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy final managed actions = %#v, want %#v", got, want)
	}
}
