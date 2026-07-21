package alicloud

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestFlinkWorkspacePostIDObservationFailureCoreRecovery exercises the real
// Terraform Core embedded in Plugin SDK v1.17.2. The disposable resource uses
// the provider contract enforced by completeFlinkWorkspaceCreate: once a paid
// ID exists, Create succeeds and every fallible observer runs in Refresh.
func TestFlinkWorkspacePostIDObservationFailureCoreRecovery(t *testing.T) {
	moduleCache, err := flinkAllocationCoreGoEnv("GOMODCACHE")
	if err != nil {
		t.Fatal(err)
	}
	sdkDir := filepath.Join(moduleCache, "github.com/hashicorp/terraform-plugin-sdk@v1.17.2")
	if _, err := os.Stat(sdkDir); err != nil {
		t.Fatalf("Plugin SDK v1.17.2 module directory %q: %v", sdkDir, err)
	}

	fixtureDir := t.TempDir()
	goMod := fmt.Sprintf("module github.com/hashicorp/terraform-plugin-sdk/flink-workspace-post-id-core-gate\n\ngo 1.20\n\nrequire github.com/hashicorp/terraform-plugin-sdk v1.17.2\n\nreplace github.com/hashicorp/terraform-plugin-sdk => %s\n", sdkDir)
	if err := os.WriteFile(filepath.Join(fixtureDir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "main.go"), []byte(flinkWorkspacePostIDCoreProgram), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "-mod=mod", ".")
	cmd.Dir = fixtureDir
	cmd.Env = append(os.Environ(), "GOPROXY=off")
	output, err := cmd.CombinedOutput()
	t.Logf("Terraform Core paid-identity fixture:\n%s", output)
	if err != nil {
		t.Fatalf("Terraform Core paid-identity contract failed: %v\n%s", err, output)
	}
}

const flinkWorkspacePostIDCoreProgram = `package main

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
	exists       bool
	id           string
	observerErr  string
	creates      int
	deletes      int
	refunds      int
	observerRuns int
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
	provider := workspaceProvider()
	config, err := loadConfig(dir)
	if err != nil { return err }
	initialContext, err := newContext(config, nil, provider)
	if err != nil { return err }
	if _, diags := initialContext.Plan(); diags.HasErrors() {
		return fmt.Errorf("initial plan: %s", diags.Err())
	}
	state, diags := initialContext.Apply()
	if diags.HasErrors() { return fmt.Errorf("initial apply: %s", diags.Err()) }
	addr := resourceAddress()
	if err := assertReady(state, addr); err != nil { return err }
	if cloud.creates != 1 || cloud.deletes != 0 || cloud.refunds != 0 {
		return fmt.Errorf("initial counters=%+v, want create/delete/refund=1/0/0", cloud)
	}

	for _, observerErr := range []string{
		"WaitForWorkspaceStarting failed",
		"DescribeFlinkWorkspace returned NotFound",
		"terraform-create-token tag missing",
		"context canceled",
	} {
		cloud.observerErr = observerErr
		failedRefresh, err := newContext(config, state, provider)
		if err != nil { return err }
		if _, refreshDiags := failedRefresh.Refresh(); !refreshDiags.HasErrors() {
			return fmt.Errorf("observer %q refresh succeeded; want fail-closed error", observerErr)
		}
		if err := assertReady(state, addr); err != nil {
			return fmt.Errorf("observer %q changed persisted object status: %w", observerErr, err)
		}
		if cloud.creates != 1 || cloud.deletes != 0 || cloud.refunds != 0 {
			return fmt.Errorf("observer %q replayed paid lifecycle: %+v", observerErr, cloud)
		}

		cloud.observerErr = ""
		retryContext, err := newContext(config, state, provider)
		if err != nil { return err }
		retryPlan, planDiags := retryContext.Plan()
		if planDiags.HasErrors() { return fmt.Errorf("retry plan after %q: %s", observerErr, planDiags.Err()) }
		if change := retryPlan.Changes.ResourceInstance(addr); change == nil || change.Action != plans.NoOp {
			return fmt.Errorf("retry action after %q=%#v, want NoOp", observerErr, change)
		}
		state, diags = retryContext.Apply()
		if diags.HasErrors() { return fmt.Errorf("retry apply after %q: %s", observerErr, diags.Err()) }
		if err := assertReady(state, addr); err != nil { return err }
		if cloud.creates != 1 || cloud.deletes != 0 || cloud.refunds != 0 {
			return fmt.Errorf("retry after %q replayed paid lifecycle: %+v", observerErr, cloud)
		}
	}

	fmt.Printf("post-ID observer failures preserve ready state; next apply=no-op; create/delete/refund=%d/%d/%d observer-runs=%d\n",
		cloud.creates, cloud.deletes, cloud.refunds, cloud.observerRuns)
	return nil
}

func workspaceProvider() *schema.Provider {
	return &schema.Provider{ResourcesMap: map[string]*schema.Resource{
		"alicloud_flink_workspace": {
			Schema: map[string]*schema.Schema{
				"name": {Type: schema.TypeString, Required: true, ForceNew: true},
				"charge_type": {Type: schema.TypeString, Required: true, ForceNew: true},
			},
			Create: func(d *schema.ResourceData, _ interface{}) error {
				cloud.creates++
				cloud.exists = true
				cloud.id = "f-paid"
				d.SetId(cloud.id)
				// Once the paid ID is known, the provider Create contract succeeds.
				return nil
			},
			Read: func(d *schema.ResourceData, _ interface{}) error {
				cloud.observerRuns++
				if cloud.observerErr != "" { return fmt.Errorf("%s", cloud.observerErr) }
				if !cloud.exists { d.SetId("") }
				return nil
			},
			Delete: func(d *schema.ResourceData, _ interface{}) error {
				if d.Get("charge_type").(string) == "PRE" { cloud.refunds++ } else { cloud.deletes++ }
				cloud.exists = false
				d.SetId("")
				return nil
			},
		},
	}}
}

func resourceAddress() addrs.AbsResourceInstance {
	return addrs.Resource{Mode: addrs.ManagedResourceMode, Type: "alicloud_flink_workspace", Name: "this"}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance)
}

func assertReady(state *states.State, addr addrs.AbsResourceInstance) error {
	instance := state.ResourceInstance(addr)
	if instance == nil || instance.Current == nil {
		return fmt.Errorf("workspace missing from state: %#v", instance)
	}
	if instance.Current.Status != states.ObjectReady {
		return fmt.Errorf("workspace status=%s, want ready", instance.Current.Status)
	}
	return nil
}

func loadConfig(dir string) (*configs.Config, error) {
	content := "resource \"alicloud_flink_workspace\" \"this\" {\n  name = \"core-workspace\"\n  charge_type = \"PRE\"\n}\n"
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
