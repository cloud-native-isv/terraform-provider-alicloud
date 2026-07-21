package alicloud

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFlinkCapacityAllocationCoreTaintRecovery exercises the Terraform Core
// embedded in Plugin SDK v1.17.2 with both real resource addresses in one
// graph. The disposable provider injects a capacity failure after one
// successful allocation write and proves that Core replaces only the safe
// allocation attachment.
func TestFlinkCapacityAllocationCoreTaintRecovery(t *testing.T) {
	moduleCache, err := flinkAllocationCoreGoEnv("GOMODCACHE")
	if err != nil {
		t.Fatal(err)
	}
	sdkDir := filepath.Join(moduleCache, "github.com/hashicorp/terraform-plugin-sdk@v1.17.2")
	if _, err := os.Stat(sdkDir); err != nil {
		t.Fatalf("Plugin SDK v1.17.2 module directory %q: %v", sdkDir, err)
	}

	fixtureDir := t.TempDir()
	goMod := fmt.Sprintf("module github.com/hashicorp/terraform-plugin-sdk/flink-allocation-core-gate\n\ngo 1.20\n\nrequire github.com/hashicorp/terraform-plugin-sdk v1.17.2\n\nreplace github.com/hashicorp/terraform-plugin-sdk => %s\n", sdkDir)
	if err := os.WriteFile(filepath.Join(fixtureDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "main.go"), []byte(flinkCapacityAllocationCoreProgram), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "-mod=mod", ".")
	cmd.Dir = fixtureDir
	cmd.Env = append(os.Environ(), "GOPROXY=off")
	output, err := cmd.CombinedOutput()
	t.Logf("Terraform Core allocation fixture:\n%s", output)
	if err != nil {
		t.Fatalf("Terraform Core allocation recovery contract failed: %v\n%s", err, output)
	}
}

func flinkAllocationCoreGoEnv(name string) (string, error) {
	output, err := exec.Command("go", "env", name).Output()
	if err != nil {
		return "", fmt.Errorf("go env %s: %w", name, err)
	}
	return strings.TrimSpace(string(output)), nil
}

const flinkCapacityAllocationCoreProgram = `package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/helper/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/internal/plans"
	"github.com/hashicorp/terraform-plugin-sdk/internal/providers"
	"github.com/hashicorp/terraform-plugin-sdk/internal/states"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	tfplugin "github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type cloudState struct {
	parentExists             bool
	parentID                 string
	capacity                 int
	parentCreates            int
	parentDeletes            int
	allocationCreates        int
	allocationDeletes        int
	capacityWrites           int
	failNextAllocationCreate bool
}

var cloud cloudState

func main() {
	log.SetOutput(io.Discard)
	if err := run("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dir string) error {
	cloud = cloudState{failNextAllocationCreate: true}
	provider := allocationProvider()
	withAllocation, err := loadConfig(dir, true)
	if err != nil {
		return err
	}
	initialContext, err := newContext(withAllocation, nil, provider)
	if err != nil {
		return err
	}
	if _, diags := initialContext.Plan(); diags.HasErrors() {
		return fmt.Errorf("initial plan: %s", diags.Err())
	}
	failedState, diags := initialContext.Apply()
	if !diags.HasErrors() {
		return fmt.Errorf("initial apply succeeded; expected injected allocation failure")
	}

	parentAddr := resourceAddress("alicloud_flink_workspace", "parent")
	allocationAddr := resourceAddress("alicloud_flink_workspace_capacity_allocation", "this")
	parent := failedState.ResourceInstance(parentAddr)
	allocation := failedState.ResourceInstance(allocationAddr)
	if parent == nil || parent.Current == nil || allocation == nil || allocation.Current == nil {
		return fmt.Errorf("failed apply state is missing parent or allocation: parent=%#v allocation=%#v", parent, allocation)
	}
	if parent.Current.Status != states.ObjectReady {
		return fmt.Errorf("parent status=%s, want ready", parent.Current.Status)
	}
	if allocation.Current.Status != states.ObjectTainted {
		return fmt.Errorf("allocation status=%s, want tainted", allocation.Current.Status)
	}
	if cloud.parentCreates != 1 || cloud.parentDeletes != 0 || cloud.capacityWrites != 1 || cloud.capacity != 2 {
		return fmt.Errorf("failed apply cloud counters=%+v, want parent create/delete=1/0 and one completed capacity step to 2", cloud)
	}

	recoveryContext, err := newContext(withAllocation, failedState, provider)
	if err != nil {
		return err
	}
	recoveryPlan, diags := recoveryContext.Plan()
	if diags.HasErrors() {
		return fmt.Errorf("recovery plan: %s", diags.Err())
	}
	if err := assertAction(recoveryPlan, parentAddr, plans.NoOp); err != nil {
		return fmt.Errorf("parent recovery action: %w", err)
	}
	if err := assertAction(recoveryPlan, allocationAddr, plans.DeleteThenCreate); err != nil {
		return fmt.Errorf("allocation recovery action: %w", err)
	}
	writesBeforeRecovery := cloud.capacityWrites
	recoveredState, diags := recoveryContext.Apply()
	if diags.HasErrors() {
		return fmt.Errorf("recovery apply: %s", diags.Err())
	}
	if cloud.parentCreates != 1 || cloud.parentDeletes != 0 || cloud.parentID != "f-core-workspace" {
		return fmt.Errorf("parent lifecycle changed during allocation recovery: %+v", cloud)
	}
	if cloud.allocationDeletes != 1 {
		return fmt.Errorf("allocation recovery Delete count=%d, want 1", cloud.allocationDeletes)
	}
	if cloud.capacityWrites != writesBeforeRecovery+1 || cloud.capacity != 3 {
		return fmt.Errorf("allocation recovery did not resume from partial actual tree: %+v", cloud)
	}

	finalContext, err := newContext(withAllocation, recoveredState, provider)
	if err != nil {
		return err
	}
	finalPlan, diags := finalContext.Plan()
	if diags.HasErrors() {
		return fmt.Errorf("final plan: %s", diags.Err())
	}
	if err := assertAction(finalPlan, parentAddr, plans.NoOp); err != nil {
		return fmt.Errorf("final parent action: %w", err)
	}
	if err := assertAction(finalPlan, allocationAddr, plans.NoOp); err != nil {
		return fmt.Errorf("final allocation action: %w", err)
	}

	parentOnly, err := loadConfig(dir, false)
	if err != nil {
		return err
	}
	detachContext, err := newContext(parentOnly, recoveredState, provider)
	if err != nil {
		return err
	}
	detachPlan, diags := detachContext.Plan()
	if diags.HasErrors() {
		return fmt.Errorf("detach plan: %s", diags.Err())
	}
	if err := assertAction(detachPlan, parentAddr, plans.NoOp); err != nil {
		return fmt.Errorf("detach parent action: %w", err)
	}
	if err := assertAction(detachPlan, allocationAddr, plans.Delete); err != nil {
		return fmt.Errorf("detach allocation action: %w", err)
	}
	capacityBeforeDetach := cloud.capacity
	writesBeforeDetach := cloud.capacityWrites
	detachedState, diags := detachContext.Apply()
	if diags.HasErrors() {
		return fmt.Errorf("detach apply: %s", diags.Err())
	}
	if cloud.capacity != capacityBeforeDetach || cloud.capacityWrites != writesBeforeDetach {
		return fmt.Errorf("removing allocation configuration mutated cloud capacity: %+v", cloud)
	}
	if cloud.allocationDeletes != 2 || cloud.parentCreates != 1 || cloud.parentDeletes != 0 {
		return fmt.Errorf("detach lifecycle counters=%+v", cloud)
	}
	if detached := detachedState.ResourceInstance(allocationAddr); detached != nil && detached.Current != nil {
		return fmt.Errorf("allocation remained in state after detach: %#v", detached)
	}

	verifyContext, err := newContext(parentOnly, detachedState, provider)
	if err != nil {
		return err
	}
	verifyPlan, diags := verifyContext.Plan()
	if diags.HasErrors() {
		return fmt.Errorf("post-detach plan: %s", diags.Err())
	}
	if err := assertAction(verifyPlan, parentAddr, plans.NoOp); err != nil {
		return fmt.Errorf("post-detach parent action: %w", err)
	}

	fmt.Printf("failed allocation=tainted; recovery parent=%s allocation=%s; final=no-op; detach writes=%d; counters parent=%d/%d allocation=%d/%d capacity=%d\n",
		plans.NoOp, plans.DeleteThenCreate, cloud.capacityWrites, cloud.parentCreates, cloud.parentDeletes, cloud.allocationCreates, cloud.allocationDeletes, cloud.capacity)
	return nil
}

func allocationProvider() *schema.Provider {
	return &schema.Provider{ResourcesMap: map[string]*schema.Resource{
		"alicloud_flink_workspace": {
			Schema: map[string]*schema.Schema{"name": {Type: schema.TypeString, Required: true, ForceNew: true}},
			Create: func(d *schema.ResourceData, _ interface{}) error {
				cloud.parentCreates++
				cloud.parentExists = true
				cloud.parentID = "f-core-workspace"
				cloud.capacity = 1
				d.SetId(cloud.parentID)
				return nil
			},
			Read: func(d *schema.ResourceData, _ interface{}) error {
				if !cloud.parentExists { d.SetId("") }
				return nil
			},
			Delete: func(d *schema.ResourceData, _ interface{}) error {
				cloud.parentDeletes++
				cloud.parentExists = false
				d.SetId("")
				return nil
			},
		},
		"alicloud_flink_workspace_capacity_allocation": {
			Schema: map[string]*schema.Schema{
				"workspace_instance_id": {Type: schema.TypeString, Required: true, ForceNew: true},
				"desired_capacity": {Type: schema.TypeInt, Required: true},
				"observed_capacity": {Type: schema.TypeInt, Computed: true},
			},
			Create: allocationCreate,
			Read: func(d *schema.ResourceData, _ interface{}) error { return d.Set("observed_capacity", cloud.capacity) },
			Update: func(d *schema.ResourceData, _ interface{}) error {
				desired := d.Get("desired_capacity").(int)
				if cloud.capacity != desired { cloud.capacity = desired; cloud.capacityWrites++ }
				return d.Set("observed_capacity", cloud.capacity)
			},
			Delete: func(d *schema.ResourceData, _ interface{}) error {
				cloud.allocationDeletes++
				d.SetId("")
				return nil
			},
		},
	}}
}

func allocationCreate(d *schema.ResourceData, _ interface{}) error {
	cloud.allocationCreates++
	d.SetId(d.Get("workspace_instance_id").(string))
	desired := d.Get("desired_capacity").(int)
	if cloud.capacity < desired {
		cloud.capacity++
		cloud.capacityWrites++
	}
	if err := d.Set("observed_capacity", cloud.capacity); err != nil { return err }
	if cloud.failNextAllocationCreate {
		cloud.failNextAllocationCreate = false
		return fmt.Errorf("injected failure after one successful capacity step")
	}
	for cloud.capacity < desired {
		cloud.capacity++
		cloud.capacityWrites++
	}
	return d.Set("observed_capacity", cloud.capacity)
}

func resourceAddress(resourceType, name string) addrs.AbsResourceInstance {
	return addrs.Resource{Mode: addrs.ManagedResourceMode, Type: resourceType, Name: name}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance)
}

func assertAction(plan *plans.Plan, addr addrs.AbsResourceInstance, want plans.Action) error {
	change := plan.Changes.ResourceInstance(addr)
	if change == nil {
		return fmt.Errorf("plan omitted %s; want %s", addr, want)
	}
	if change.Action != want {
		return fmt.Errorf("action=%s, want %s", change.Action, want)
	}
	return nil
}

func loadConfig(dir string, includeAllocation bool) (*configs.Config, error) {
	content := "resource \"alicloud_flink_workspace\" \"parent\" { name = \"core-parent\" }\n"
	if includeAllocation {
		content += "resource \"alicloud_flink_workspace_capacity_allocation\" \"this\" {\n  workspace_instance_id = alicloud_flink_workspace.parent.id\n  desired_capacity = 3\n}\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o600); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}
	module, diags := configs.NewParser(afero.NewOsFs()).LoadConfigDir(dir)
	if diags.HasErrors() { return nil, fmt.Errorf("load config: %s", diags.Error()) }
	config, diags := configs.BuildConfig(module, nil)
	if diags.HasErrors() { return nil, fmt.Errorf("build config: %s", diags.Error()) }
	return config, nil
}

func newContext(config *configs.Config, state *states.State, provider *schema.Provider) (*terraform.Context, error) {
	ctx, diags := terraform.NewContext(&terraform.ContextOpts{
		Config: config,
		State: state,
		ProviderResolver: providers.ResolverFixed(map[string]providers.Factory{
			"alicloud": func() (providers.Interface, error) { return grpcProvider(provider) },
		}),
	})
	if diags.HasErrors() { return nil, fmt.Errorf("new context: %s", diags.Err()) }
	return ctx, nil
}

func grpcProvider(provider *schema.Provider) (providers.Interface, error) {
	listener := bufconn.Listen(256 * 1024)
	server := grpc.NewServer()
	proto.RegisterProviderServer(server, plugin.NewGRPCProviderServerShim(provider))
	go server.Serve(listener)
	conn, err := grpc.Dial("", grpc.WithDialer(func(string, time.Duration) (net.Conn, error) { return listener.Dial() }), grpc.WithInsecure())
	if err != nil { server.Stop(); return nil, fmt.Errorf("dial provider: %w", err) }
	var grpcPlugin tfplugin.GRPCProviderPlugin
	client, err := grpcPlugin.GRPCClient(context.Background(), nil, conn)
	if err != nil { server.Stop(); return nil, fmt.Errorf("create provider client: %w", err) }
	grpcClient := client.(*tfplugin.GRPCProvider)
	grpcClient.TestServer = server
	return grpcClient, nil
}
`
