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

const flinkCapacityBootstrapTofuLedgerEnv = "FLINK_CAPACITY_BOOTSTRAP_RUNTIME_LEDGER"

type flinkCapacityBootstrapTofuLedger struct {
	ParentExists                     bool   `json:"parent_exists"`
	ParentID                         string `json:"parent_id"`
	ParentResourceID                 string `json:"parent_resource_id"`
	ParentCreateToken                string `json:"parent_create_token"`
	ParentCreateFingerprint          string `json:"parent_create_fingerprint"`
	ParentCreates                    int    `json:"parent_creates"`
	ParentDeletes                    int    `json:"parent_deletes"`
	ParentReads                      int    `json:"parent_reads"`
	ParentIdentityLists              int    `json:"parent_identity_lists"`
	CapacityServiceFactories         int    `json:"capacity_service_factories"`
	CapacityWorkspaceGets            int    `json:"capacity_workspace_gets"`
	CapacityWorkspaceLists           int    `json:"capacity_workspace_lists"`
	CapacityNamespaceLists           int    `json:"capacity_namespace_lists"`
	CapacityTargetLists              int    `json:"capacity_target_lists"`
	CapacityWrites                   int    `json:"capacity_writes"`
	PrematureCapacityWrites          int    `json:"premature_capacity_writes"`
	WorkspaceFixedCU                 int    `json:"workspace_fixed_cu"`
	NamespaceFixedCU                 int    `json:"namespace_fixed_cu"`
	QueueFixedCU                     int    `json:"queue_fixed_cu"`
	DelayedVisibility                bool   `json:"delayed_visibility"`
	FailNextCapacityWrite            bool   `json:"fail_next_capacity_write"`
	InjectedCapacityWriteFailures    int    `json:"injected_capacity_write_failures"`
	CapacityFingerprintMismatchOnGet int    `json:"capacity_fingerprint_mismatch_on_get"`
	CapacityResourceIDMismatchOnGet  int    `json:"capacity_resource_id_mismatch_on_get"`
}

type flinkCapacityBootstrapTofuFixture struct {
	gate   *flinkTofuRuntimeGate
	ledger string
}

func TestFlinkWorkspaceCapacityBootstrapOpenTofuProductionCallbacks(t *testing.T) {
	if os.Getenv("RUN_FLINK_TOFU_RUNTIME_GATE") != "1" {
		t.Skip("set RUN_FLINK_TOFU_RUNTIME_GATE=1 to run the OpenTofu capacity bootstrap production-callback gate")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	root := t.TempDir()
	owned := flinkTofuRuntimeCreateOwnedEnvironment(t, root)
	flinkTofuRuntimeRequireVersion(t, ctx, flinkTofuRuntimeEnvironment(owned, owned.emptyCLI, "", "", flinkTofuRuntimeFixtureControl{}))
	assertFlinkCapacityBootstrapRuntimeProductionContract(t)

	providerArtifact := flinkCapacityBootstrapRuntimeBuildProvider(t, ctx, root, owned, false)
	providerSHA := flinkTofuRuntimeSHA256(t, providerArtifact)
	mirrorRoot, cliConfig := flinkCapacityBootstrapRuntimeCreateMirrorAndCLI(t, root, "new", providerArtifact)
	_ = mirrorRoot

	t.Run("fresh graph converges in one apply after delayed identity propagation", func(t *testing.T) {
		fixture := newFlinkCapacityBootstrapTofuFixture(t, ctx, root, owned, cliConfig, "fresh", flinkCapacityBootstrapTofuLedger{DelayedVisibility: true}, true)
		assertFlinkCapacityBootstrapInstalledProvider(t, root, fixture.gate, providerSHA)
		if output, err := fixture.gate.tofu("apply", "-input=false", "-auto-approve", "-no-color"); err != nil {
			state := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
			parentContext, outputErr := fixture.gate.tofu("output", "-raw", "parent_context")
			barrierContext, barrierOutputErr := fixture.gate.tofu("output", "-raw", "barrier_context")
			t.Fatalf("fresh apply failed: %v; parent identity lists=%d parent_context_present=%t parent_output_error=%v barrier_context_present=%t barrier_output_error=%v\n%s", err, state.ParentIdentityLists, strings.TrimSpace(parentContext) != "", outputErr, strings.TrimSpace(barrierContext) != "", barrierOutputErr, output)
		}
		state := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
		assertFlinkCapacityBootstrapRuntimeConverged(t, state, "fresh apply")
		if state.ParentCreates != 1 || state.ParentDeletes != 0 || state.ParentIdentityLists != 0 || state.CapacityWorkspaceLists != 1 || state.CapacityWrites != 2 || state.PrematureCapacityWrites != 0 {
			t.Fatalf("fresh lifecycle counters = parent create/delete/lists %d/%d/%d, capacity lists/writes/premature %d/%d/%d", state.ParentCreates, state.ParentDeletes, state.ParentIdentityLists, state.CapacityWorkspaceLists, state.CapacityWrites, state.PrematureCapacityWrites)
		}
		parentContext := strings.TrimSpace(fixture.gate.mustTofu("output", "-raw", "parent_context"))
		barrierContext := strings.TrimSpace(fixture.gate.mustTofu("output", "-raw", "barrier_context"))
		if parentContext == "" || parentContext != barrierContext {
			t.Fatalf("fresh parent/barrier context differs or is empty")
		}
		parsed, err := parseFlinkWorkspaceCapacityBootstrapContext(parentContext)
		if err != nil {
			t.Fatal(err)
		}
		barrierResourceID := strings.TrimSpace(fixture.gate.mustTofu("output", "-raw", "barrier_resource_id"))
		if parsed.ExpectedInstanceID != state.ParentID || parsed.ExpectedResourceID != "" || parsed.TerraformCreateToken != state.ParentCreateToken || parsed.CreateIntentFingerprint != state.ParentCreateFingerprint || barrierResourceID != state.ParentResourceID {
			t.Fatalf("fresh provenance mismatch: instance_match=%t resource_deferred=%t token_match=%t fingerprint_match=%t barrier_resource_match=%t", parsed.ExpectedInstanceID == state.ParentID, parsed.ExpectedResourceID == "", parsed.TerraformCreateToken == state.ParentCreateToken, parsed.CreateIntentFingerprint == state.ParentCreateFingerprint, barrierResourceID == state.ParentResourceID)
		}

		planPath := filepath.Join(root, "fresh-final.tfplan")
		fixture.gate.mustTofu("plan", "-input=false", "-no-color", "-out="+planPath)
		assertFlinkTofuRuntimePlan(t, fixture.gate.readPlan(planPath), map[string][]string{
			"alicloud_flink_workspace.parent":                      {"no-op"},
			"alicloud_flink_workspace_capacity_bootstrap.ready":    {"no-op"},
			"alicloud_flink_workspace_capacity_allocation_v2.this": {"no-op"},
		})
		if got := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger); got.ParentCreates != 1 || got.CapacityWrites != state.CapacityWrites {
			t.Fatalf("fresh no-op plan replayed lifecycle: creates=%d writes_before=%d writes_after=%d", got.ParentCreates, state.CapacityWrites, got.CapacityWrites)
		}
	})

	t.Run("tainted allocation replacement reuses persisted deadline without parent replay", func(t *testing.T) {
		fixture := newFlinkCapacityBootstrapTofuFixture(t, ctx, root, owned, cliConfig, "tainted", flinkCapacityBootstrapTofuLedger{
			DelayedVisibility:     true,
			FailNextCapacityWrite: true,
		}, true)
		assertFlinkCapacityBootstrapInstalledProvider(t, root, fixture.gate, providerSHA)
		output, err := fixture.gate.tofu("apply", "-input=false", "-auto-approve", "-no-color")
		if err == nil || !strings.Contains(output, "injected bootstrap capacity failure after accepted write") {
			t.Fatalf("initial taint apply error = %v, output missing injected failure:\n%s", err, output)
		}
		failed := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
		if failed.ParentCreates != 1 || failed.ParentDeletes != 0 || failed.CapacityWrites != 1 || failed.PrematureCapacityWrites != 0 || failed.InjectedCapacityWriteFailures != 1 {
			t.Fatalf("failed taint counters = parent %d/%d, writes=%d, premature=%d, injected=%d", failed.ParentCreates, failed.ParentDeletes, failed.CapacityWrites, failed.PrematureCapacityWrites, failed.InjectedCapacityWriteFailures)
		}
		before := strings.TrimSpace(fixture.gate.mustTofu("output", "-raw", "parent_context"))
		beforeContext, err := parseFlinkWorkspaceCapacityBootstrapContext(before)
		if err != nil {
			t.Fatal(err)
		}

		recoveryPlanPath := filepath.Join(root, "tainted-recovery.tfplan")
		fixture.gate.mustTofu("plan", "-input=false", "-no-color", "-out="+recoveryPlanPath)
		recoveryPlan := fixture.gate.readPlan(recoveryPlanPath)
		assertFlinkTofuRuntimePlan(t, recoveryPlan, map[string][]string{
			"alicloud_flink_workspace.parent":                      {"no-op"},
			"alicloud_flink_workspace_capacity_bootstrap.ready":    {"no-op"},
			"alicloud_flink_workspace_capacity_allocation_v2.this": {"delete", "create"},
		})
		for _, change := range recoveryPlan.ResourceChanges {
			if change.Address == "alicloud_flink_workspace_capacity_allocation_v2.this" && change.ActionReason != "replace_because_tainted" {
				t.Fatalf("allocation recovery reason = %q, want replace_because_tainted", change.ActionReason)
			}
		}

		fixture.gate.mustTofu("apply", "-input=false", "-no-color", recoveryPlanPath)
		recovered := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
		assertFlinkCapacityBootstrapRuntimeConverged(t, recovered, "tainted recovery")
		if recovered.ParentCreates != 1 || recovered.ParentDeletes != 0 || recovered.CapacityWrites != 2 || recovered.PrematureCapacityWrites != 0 {
			t.Fatalf("recovered counters = parent %d/%d, writes=%d, premature=%d", recovered.ParentCreates, recovered.ParentDeletes, recovered.CapacityWrites, recovered.PrematureCapacityWrites)
		}
		after := strings.TrimSpace(fixture.gate.mustTofu("output", "-raw", "parent_context"))
		afterContext, err := parseFlinkWorkspaceCapacityBootstrapContext(after)
		if err != nil {
			t.Fatal(err)
		}
		if before != after || beforeContext.IdentityAbsenceRetryNotAfterUnix != afterContext.IdentityAbsenceRetryNotAfterUnix {
			t.Fatalf("taint recovery renewed or changed the persisted bootstrap context")
		}
		finalPlanPath := filepath.Join(root, "tainted-final.tfplan")
		fixture.gate.mustTofu("plan", "-input=false", "-no-color", "-out="+finalPlanPath)
		assertFlinkTofuRuntimePlan(t, fixture.gate.readPlan(finalPlanPath), map[string][]string{
			"alicloud_flink_workspace.parent":                      {"no-op"},
			"alicloud_flink_workspace_capacity_bootstrap.ready":    {"no-op"},
			"alicloud_flink_workspace_capacity_allocation_v2.this": {"no-op"},
		})
		if finalState := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger); finalState.ParentCreates != recovered.ParentCreates || finalState.ParentDeletes != recovered.ParentDeletes || finalState.CapacityWrites != recovered.CapacityWrites || finalState.WorkspaceFixedCU != recovered.WorkspaceFixedCU || finalState.NamespaceFixedCU != recovered.NamespaceFixedCU || finalState.QueueFixedCU != recovered.QueueFixedCU || finalState.ParentCreateToken != recovered.ParentCreateToken || finalState.ParentCreateFingerprint != recovered.ParentCreateFingerprint {
			t.Fatalf("taint final no-op plan changed lifecycle, capacity, or persisted provenance")
		}
	})

	t.Run("barrier provenance mismatch fails with zero capacity writes", func(t *testing.T) {
		fixture := newFlinkCapacityBootstrapTofuFixture(t, ctx, root, owned, cliConfig, "barrier-provenance-mismatch", flinkCapacityBootstrapTofuLedger{
			CapacityFingerprintMismatchOnGet: 1,
		}, true)
		assertFlinkCapacityBootstrapInstalledProvider(t, root, fixture.gate, providerSHA)
		output, err := fixture.gate.tofu("apply", "-input=false", "-auto-approve", "-no-color")
		if err == nil || !strings.Contains(output, "intent fingerprint mismatch") {
			t.Fatalf("barrier provenance mismatch apply error = %v, output missing strict provenance failure", err)
		}
		state := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
		if state.ParentCreates != 1 || state.ParentDeletes != 0 || state.CapacityWorkspaceGets < 1 || state.CapacityWrites != 0 || state.PrematureCapacityWrites != 0 {
			t.Fatalf("barrier provenance mismatch lifecycle counters = parent %d/%d, gets=%d, writes=%d, premature=%d", state.ParentCreates, state.ParentDeletes, state.CapacityWorkspaceGets, state.CapacityWrites, state.PrematureCapacityWrites)
		}
	})

	t.Run("allocation ResourceId change at write boundary fails with zero capacity writes", func(t *testing.T) {
		fixture := newFlinkCapacityBootstrapTofuFixture(t, ctx, root, owned, cliConfig, "allocation-resource-mismatch", flinkCapacityBootstrapTofuLedger{
			CapacityResourceIDMismatchOnGet: 3,
		}, true)
		assertFlinkCapacityBootstrapInstalledProvider(t, root, fixture.gate, providerSHA)
		output, err := fixture.gate.tofu("apply", "-input=false", "-auto-approve", "-no-color")
		if err == nil || !strings.Contains(output, "ResourceId mismatch") {
			t.Fatalf("allocation ResourceId mismatch apply error = %v, output missing pinned identity failure", err)
		}
		state := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
		if state.ParentCreates != 1 || state.ParentDeletes != 0 || state.CapacityWorkspaceGets < 3 || state.CapacityWrites != 0 || state.PrematureCapacityWrites != 0 {
			t.Fatalf("allocation ResourceId mismatch counters = parent %d/%d, gets=%d, writes=%d, premature=%d", state.ParentCreates, state.ParentDeletes, state.CapacityWorkspaceGets, state.CapacityWrites, state.PrematureCapacityWrites)
		}
	})

	t.Logf("bootstrap production-callback provider SHA-256: %s", providerSHA)
}

func TestFlinkWorkspaceCapacityBootstrapOpenTofuSameVersionOldStateCompatibility(t *testing.T) {
	if os.Getenv("RUN_FLINK_TOFU_RUNTIME_GATE") != "1" {
		t.Skip("set RUN_FLINK_TOFU_RUNTIME_GATE=1 to run the OpenTofu capacity bootstrap compatibility gate")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	root := t.TempDir()
	owned := flinkTofuRuntimeCreateOwnedEnvironment(t, root)
	flinkTofuRuntimeRequireVersion(t, ctx, flinkTofuRuntimeEnvironment(owned, owned.emptyCLI, "", "", flinkTofuRuntimeFixtureControl{}))

	oldArtifact := flinkCapacityBootstrapRuntimeBuildProvider(t, ctx, root, owned, true)
	newArtifact := flinkCapacityBootstrapRuntimeBuildProvider(t, ctx, root, owned, false)
	_, oldCLI := flinkCapacityBootstrapRuntimeCreateMirrorAndCLI(t, root, "compat-old", oldArtifact)
	_, newCLI := flinkCapacityBootstrapRuntimeCreateMirrorAndCLI(t, root, "compat-new", newArtifact)

	fixture := newFlinkCapacityBootstrapTofuFixture(t, ctx, root, owned, oldCLI, "compat", flinkCapacityBootstrapTofuLedger{}, false)
	assertFlinkCapacityBootstrapInstalledProvider(t, root, fixture.gate, flinkTofuRuntimeSHA256(t, oldArtifact))
	fixture.gate.mustTofu("apply", "-input=false", "-auto-approve", "-no-color")
	baseline := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
	assertFlinkCapacityBootstrapRuntimeConverged(t, baseline, "old-schema baseline")
	if baseline.ParentCreates != 1 || baseline.ParentDeletes != 0 {
		t.Fatalf("old-schema fixture baseline parent create/delete = %d/%d", baseline.ParentCreates, baseline.ParentDeletes)
	}

	// A provider-only upgrade keeps the old HCL and must remain a pure no-op.
	newGate := newFlinkCapacityBootstrapTofuGateForExistingState(t, ctx, root, owned, fixture.gate.dir, fixture.ledger, newCLI, "compat-new-data")
	assertFlinkCapacityBootstrapInstalledProvider(t, root, newGate, flinkTofuRuntimeSHA256(t, newArtifact))
	newPlanPath := filepath.Join(root, "compat-new.tfplan")
	newGate.mustTofu("plan", "-input=false", "-no-color", "-out="+newPlanPath)
	newPlan := newGate.readPlan(newPlanPath)
	assertFlinkTofuRuntimePlan(t, newPlan, map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"no-op"},
	})
	newGate.mustTofu("apply", "-input=false", "-no-color", newPlanPath)
	afterNew := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
	if afterNew.ParentCreates != 1 || afterNew.ParentDeletes != 0 || afterNew.CapacityWrites != baseline.CapacityWrites {
		t.Fatalf("provider-only upgrade changed old lifecycle or capacity")
	}
	postNormalizationPlanPath := filepath.Join(root, "compat-new-normalized.tfplan")
	newGate.mustTofu("plan", "-input=false", "-no-color", "-out="+postNormalizationPlanPath)
	assertFlinkTofuRuntimePlan(t, newGate.readPlan(postNormalizationPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"no-op"},
	})

	// Adopting the new HCL preserves the paid parent, detaches the legacy
	// allocation state, and creates the provenance-pinned v2 attachment. Both
	// allocation Deletes are state-only and equal desired capacity means zero
	// cloud writes during adoption.
	writeFlinkCapacityBootstrapRuntimeConfig(t, fixture.gate.dir, true)
	adoptGate := newFlinkCapacityBootstrapTofuGateForExistingState(t, ctx, root, owned, fixture.gate.dir, fixture.ledger, newCLI, "compat-adopt-data")
	assertFlinkCapacityBootstrapInstalledProvider(t, root, adoptGate, flinkTofuRuntimeSHA256(t, newArtifact))
	adoptPlanPath := filepath.Join(root, "compat-adopt.tfplan")
	adoptGate.mustTofu("plan", "-input=false", "-no-color", "-out="+adoptPlanPath)
	assertFlinkTofuRuntimePlan(t, adoptGate.readPlan(adoptPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                      {"no-op"},
		"alicloud_flink_workspace_capacity_bootstrap.ready":    {"create"},
		"alicloud_flink_workspace_capacity_allocation.this":    {"delete"},
		"alicloud_flink_workspace_capacity_allocation_v2.this": {"create"},
	})
	adoptGate.mustTofu("apply", "-input=false", "-no-color", adoptPlanPath)
	afterAdoption := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
	if afterAdoption.ParentCreates != 1 || afterAdoption.ParentDeletes != 0 || afterAdoption.CapacityWrites != baseline.CapacityWrites {
		t.Fatalf("barrier adoption changed paid lifecycle or capacity")
	}
	adoptFinalPlanPath := filepath.Join(root, "compat-adopt-final.tfplan")
	adoptGate.mustTofu("plan", "-input=false", "-no-color", "-out="+adoptFinalPlanPath)
	assertFlinkTofuRuntimePlan(t, adoptGate.readPlan(adoptFinalPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                      {"no-op"},
		"alicloud_flink_workspace_capacity_bootstrap.ready":    {"no-op"},
		"alicloud_flink_workspace_capacity_allocation_v2.this": {"no-op"},
	})

	// A rollback must first remove the barrier from configuration while the new
	// provider can still decode and detach it. This is a state-only delete; only
	// after this cleanup is a provider downgrade supportable. The synthetic
	// old-schema fixture is deliberately not presented as a historical binary.
	writeFlinkCapacityBootstrapRuntimeConfig(t, fixture.gate.dir, false)
	cleanupGate := newFlinkCapacityBootstrapTofuGateForExistingState(t, ctx, root, owned, fixture.gate.dir, fixture.ledger, newCLI, "compat-cleanup-data")
	assertFlinkCapacityBootstrapInstalledProvider(t, root, cleanupGate, flinkTofuRuntimeSHA256(t, newArtifact))
	cleanupPlanPath := filepath.Join(root, "compat-cleanup.tfplan")
	cleanupGate.mustTofu("plan", "-input=false", "-no-color", "-out="+cleanupPlanPath)
	assertFlinkTofuRuntimePlan(t, cleanupGate.readPlan(cleanupPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                      {"no-op"},
		"alicloud_flink_workspace_capacity_bootstrap.ready":    {"delete"},
		"alicloud_flink_workspace_capacity_allocation_v2.this": {"delete"},
		"alicloud_flink_workspace_capacity_allocation.this":    {"create"},
	})
	cleanupGate.mustTofu("apply", "-input=false", "-no-color", cleanupPlanPath)
	afterCleanup := readFlinkCapacityBootstrapRuntimeLedger(t, fixture.ledger)
	if afterCleanup.ParentCreates != 1 || afterCleanup.ParentDeletes != 0 || afterCleanup.CapacityWrites != baseline.CapacityWrites {
		t.Fatalf("barrier cleanup changed paid lifecycle or capacity")
	}
	cleanupFinalPlanPath := filepath.Join(root, "compat-cleanup-final.tfplan")
	cleanupGate.mustTofu("plan", "-input=false", "-no-color", "-out="+cleanupFinalPlanPath)
	assertFlinkTofuRuntimePlan(t, cleanupGate.readPlan(cleanupFinalPlanPath), map[string][]string{
		"alicloud_flink_workspace.parent":                   {"no-op"},
		"alicloud_flink_workspace_capacity_allocation.this": {"no-op"},
	})
}

func assertFlinkCapacityBootstrapRuntimeProductionContract(t *testing.T) {
	t.Helper()
	provider := Provider().(*schema.Provider)
	workspace := provider.ResourcesMap["alicloud_flink_workspace"]
	legacyAllocation := provider.ResourcesMap["alicloud_flink_workspace_capacity_allocation"]
	allocation := provider.ResourcesMap["alicloud_flink_workspace_capacity_allocation_v2"]
	bootstrap := provider.ResourcesMap["alicloud_flink_workspace_capacity_bootstrap"]
	if workspace == nil || legacyAllocation == nil || allocation == nil || bootstrap == nil || workspace.Create == nil || workspace.Read == nil || bootstrap.Create == nil || bootstrap.Read == nil || bootstrap.Delete == nil || allocation.Create == nil || allocation.CustomizeDiff == nil {
		t.Fatalf("production bootstrap resources/callbacks missing: workspace=%#v bootstrap=%#v pinned_allocation=%#v", workspace, bootstrap, allocation)
	}
	for name, callbacks := range map[string]struct {
		got  interface{}
		want interface{}
	}{
		"workspace Create":  {workspace.Create, resourceAliCloudFlinkWorkspaceCreate},
		"workspace Read":    {workspace.Read, resourceAliCloudFlinkWorkspaceRead},
		"bootstrap Create":  {bootstrap.Create, resourceAliCloudFlinkWorkspaceCapacityBootstrapCreate},
		"bootstrap Read":    {bootstrap.Read, resourceAliCloudFlinkWorkspaceCapacityBootstrapRead},
		"bootstrap Delete":  {bootstrap.Delete, resourceAliCloudFlinkWorkspaceCapacityBootstrapDelete},
		"allocation Create": {allocation.Create, resourceAliCloudFlinkWorkspaceCapacityAllocationV2Create},
		"allocation Read":   {allocation.Read, resourceAliCloudFlinkWorkspaceCapacityAllocationV2Read},
		"allocation Update": {allocation.Update, resourceAliCloudFlinkWorkspaceCapacityAllocationV2Update},
		"allocation Diff":   {allocation.CustomizeDiff, flinkWorkspaceCapacityAllocationCustomizeDiff},
	} {
		if reflect.ValueOf(callbacks.got).Pointer() != reflect.ValueOf(callbacks.want).Pointer() {
			t.Fatalf("runtime fixture %s callback is not the production implementation", name)
		}
	}
	parentContext := workspace.Schema["capacity_bootstrap_context"]
	childContext := bootstrap.Schema["workspace_bootstrap_context"]
	if parentContext == nil || !parentContext.Computed || !parentContext.Sensitive || childContext == nil || !childContext.Required || !childContext.Sensitive || !childContext.ForceNew {
		t.Fatalf("production bootstrap schema parent/child = %#v/%#v", parentContext, childContext)
	}
	if allocation.Schema["workspace_bootstrap_context"] == nil || allocation.Schema["workspace_resource_id"] == nil {
		t.Fatal("pinned capacity allocation is missing required bootstrap identity inputs")
	}
	if _, exists := legacyAllocation.Schema["workspace_bootstrap_context"]; exists {
		t.Fatal("legacy capacity allocation schema changed")
	}
	if workspace.SchemaVersion != 1 || legacyAllocation.SchemaVersion != 0 || allocation.SchemaVersion != 0 {
		t.Fatalf("production schema versions workspace/allocation = %d/%d, want 1/0", workspace.SchemaVersion, allocation.SchemaVersion)
	}
}

func flinkCapacityBootstrapRuntimeBuildProvider(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment, oldSchema bool) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve bootstrap runtime test source path")
	}
	repoRoot := filepath.Dir(filepath.Dir(sourceFile))
	variant := "new"
	providerFunc := "FlinkCapacityBootstrapRuntimeFixtureProvider"
	if oldSchema {
		variant = "old"
		providerFunc = "FlinkCapacityBootstrapOldSchemaRuntimeFixtureProvider"
	}
	moduleDir := filepath.Join(root, "bootstrap-provider-main-"+variant)
	artifactDir := filepath.Join(root, "bootstrap-provider-artifact-"+variant)
	for _, dir := range []string{moduleDir, artifactDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("create bootstrap provider build directory %q: %v", dir, err)
		}
	}
	mainSource := fmt.Sprintf(`package main

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: alicloud.%s})
}
`, providerFunc)
	mainPath := filepath.Join(moduleDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainSource), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(artifactDir, "terraform-provider-alicloud_v"+flinkTofuRuntimeProviderVersion)
	build := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-mod=readonly", "-tags=flink_capacity_bootstrap_runtime_fixture", "-o", artifact, mainPath)
	build.Dir = repoRoot
	build.Env = flinkTofuRuntimeGoEnvironment(owned)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build %s bootstrap production-callback provider: %v\n%s", variant, err, output)
	}
	return artifact
}

func flinkCapacityBootstrapRuntimeCreateMirrorAndCLI(t *testing.T, root, name, artifact string) (string, string) {
	t.Helper()
	mirrorRoot := filepath.Join(root, name+"-mirror")
	flinkTofuRuntimeCreateMirror(t, mirrorRoot, artifact)
	cliConfig := filepath.Join(root, name+"-terraformrc")
	body := fmt.Sprintf(`provider_installation {
  filesystem_mirror {
    path    = %s
    include = [%q]
  }
}
`, strconv.Quote(mirrorRoot), flinkTofuRuntimeProviderSource)
	if err := os.WriteFile(cliConfig, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return mirrorRoot, cliConfig
}

func newFlinkCapacityBootstrapTofuFixture(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment, cliConfig, name string, initial flinkCapacityBootstrapTofuLedger, includeContext bool) flinkCapacityBootstrapTofuFixture {
	t.Helper()
	configDir := filepath.Join(root, name+"-configuration")
	tfDataDir := filepath.Join(root, name+"-tf-data")
	pluginCache := filepath.Join(root, name+"-plugin-cache")
	for _, dir := range []string{configDir, tfDataDir, pluginCache} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	ledger := filepath.Join(root, name+"-ledger.json")
	writeFlinkCapacityBootstrapRuntimeLedger(t, ledger, initial)
	writeFlinkCapacityBootstrapRuntimeConfig(t, configDir, includeContext)
	env := flinkTofuRuntimeEnvironment(owned, cliConfig, tfDataDir, pluginCache, flinkTofuRuntimeFixtureControl{})
	env = append(env, flinkCapacityBootstrapTofuLedgerEnv+"="+ledger)
	gate := &flinkTofuRuntimeGate{t: t, ctx: ctx, dir: configDir, env: env, ledger: ledger, binary: flinkTofuRuntimeBinary(), tfData: tfDataDir}
	gate.mustTofu("init", "-input=false", "-no-color")
	return flinkCapacityBootstrapTofuFixture{gate: gate, ledger: ledger}
}

func newFlinkCapacityBootstrapTofuGateForExistingState(t *testing.T, ctx context.Context, root string, owned flinkTofuRuntimeOwnedEnvironment, configDir, ledger, cliConfig, dataName string) *flinkTofuRuntimeGate {
	t.Helper()
	lockPath := filepath.Join(configDir, ".terraform.lock.hcl")
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove invocation-owned provider lock %q: %v", lockPath, err)
	}
	tfDataDir := filepath.Join(root, dataName+"-tf-data")
	pluginCache := filepath.Join(root, dataName+"-plugin-cache")
	for _, dir := range []string{tfDataDir, pluginCache} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	env := flinkTofuRuntimeEnvironment(owned, cliConfig, tfDataDir, pluginCache, flinkTofuRuntimeFixtureControl{})
	env = append(env, flinkCapacityBootstrapTofuLedgerEnv+"="+ledger)
	gate := &flinkTofuRuntimeGate{t: t, ctx: ctx, dir: configDir, env: env, ledger: ledger, binary: flinkTofuRuntimeBinary(), tfData: tfDataDir}
	gate.mustTofu("init", "-input=false", "-no-color")
	return gate
}

func writeFlinkCapacityBootstrapRuntimeConfig(t *testing.T, dir string, includeContext bool) {
	t.Helper()
	bootstrapResource := ""
	allocationResourceType := "alicloud_flink_workspace_capacity_allocation"
	allocationInstanceID := "alicloud_flink_workspace.parent.id"
	allocationIdentity := ""
	outputs := ""
	if includeContext {
		bootstrapResource = `
resource "alicloud_flink_workspace_capacity_bootstrap" "ready" {
  workspace_instance_id       = alicloud_flink_workspace.parent.id
  workspace_bootstrap_context = alicloud_flink_workspace.parent.capacity_bootstrap_context
}
`
		allocationResourceType = "alicloud_flink_workspace_capacity_allocation_v2"
		allocationInstanceID = "alicloud_flink_workspace_capacity_bootstrap.ready.workspace_instance_id"
		allocationIdentity = `
  workspace_resource_id       = alicloud_flink_workspace_capacity_bootstrap.ready.observed_resource_id
  workspace_bootstrap_context = alicloud_flink_workspace_capacity_bootstrap.ready.workspace_bootstrap_context
`
		outputs = `
output "parent_context" {
  value     = alicloud_flink_workspace.parent.capacity_bootstrap_context
  sensitive = true
}

output "barrier_context" {
  value     = alicloud_flink_workspace_capacity_bootstrap.ready.workspace_bootstrap_context
  sensitive = true
}

output "barrier_resource_id" {
  value = alicloud_flink_workspace_capacity_bootstrap.ready.observed_resource_id
}
`
	}
	body := fmt.Sprintf(`terraform {
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
  name              = "bootstrap-runtime-parent"
  resource_group_id = "rg-fixture"
  vpc_id            = "vpc-fixture"
  vswitch_ids       = ["vsw-fixture"]
  charge_type       = "PRE"

  initial_capacity {
    fixed_cu            = 2
    cross_zone_fixed_cu = 0
  }

  storage {
    oss_bucket = "fixture-bucket"
  }
}

%s

resource "%s" "this" {
  workspace_instance_id = %s
%s
  fixed_cu              = 2
  cross_zone_fixed_cu   = 0
  max_cu_limit          = 2

  namespace {
    name         = "default"
    fixed_cu     = 2
    max_cu_limit = 2
  }
}
%s`, flinkTofuRuntimeProviderVersion, bootstrapResource, allocationResourceType, allocationInstanceID, allocationIdentity, outputs)
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeFlinkCapacityBootstrapRuntimeLedger(t *testing.T, path string, state flinkCapacityBootstrapTofuLedger) {
	t.Helper()
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readFlinkCapacityBootstrapRuntimeLedger(t *testing.T, path string) flinkCapacityBootstrapTofuLedger {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var state flinkCapacityBootstrapTofuLedger
	if err := json.Unmarshal(body, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func assertFlinkCapacityBootstrapRuntimeConverged(t *testing.T, state flinkCapacityBootstrapTofuLedger, phase string) {
	t.Helper()
	if !state.ParentExists || state.ParentID != "f-bootstrap-runtime" || state.ParentResourceID != "resource-bootstrap-runtime" || state.WorkspaceFixedCU != 2 || state.NamespaceFixedCU != 2 || state.QueueFixedCU != 2 {
		t.Fatalf("%s did not converge: parent_exists=%t instance_match=%t resource_match=%t capacity=%d/%d/%d", phase, state.ParentExists, state.ParentID == "f-bootstrap-runtime", state.ParentResourceID == "resource-bootstrap-runtime", state.WorkspaceFixedCU, state.NamespaceFixedCU, state.QueueFixedCU)
	}
	if state.ParentCreateToken == "" || state.ParentCreateFingerprint == "" {
		t.Fatalf("%s lost paid-create provenance: token_present=%t fingerprint_present=%t", phase, state.ParentCreateToken != "", state.ParentCreateFingerprint != "")
	}
}

func assertFlinkCapacityBootstrapInstalledProvider(t *testing.T, root string, gate *flinkTofuRuntimeGate, artifactSHA string) {
	t.Helper()
	installed := flinkTofuRuntimeInstalledProvider(t, gate.tfData)
	if !flinkTofuRuntimePathWithin(root, installed) {
		t.Fatalf("installed bootstrap provider resolved outside invocation-owned root: root=%q provider=%q", root, installed)
	}
	if installedSHA := flinkTofuRuntimeSHA256(t, installed); installedSHA != artifactSHA {
		t.Fatalf("installed bootstrap provider SHA-256 = %s, want freshly built artifact %s", installedSHA, artifactSHA)
	}
}
