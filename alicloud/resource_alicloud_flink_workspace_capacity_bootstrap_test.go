package alicloud

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestFlinkWorkspaceCapacityBootstrapSchemaIsRequiredReadOnlyBarrier(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityBootstrap()
	instanceID := resource.Schema["workspace_instance_id"]
	context := resource.Schema["workspace_bootstrap_context"]
	observedResourceID := resource.Schema["observed_resource_id"]
	if resource.Create == nil || resource.Read == nil || resource.Update != nil || resource.Delete == nil {
		t.Fatalf("capacity bootstrap callbacks = create:%t read:%t update:%t delete:%t", resource.Create != nil, resource.Read != nil, resource.Update != nil, resource.Delete != nil)
	}
	if instanceID == nil || !instanceID.Required || !instanceID.ForceNew || instanceID.Optional || instanceID.Computed {
		t.Fatalf("workspace_instance_id schema = %#v, want required ForceNew", instanceID)
	}
	if context == nil || !context.Required || !context.Sensitive || !context.ForceNew || context.Optional || context.Computed {
		t.Fatalf("workspace_bootstrap_context schema = %#v, want required sensitive ForceNew", context)
	}
	if observedResourceID == nil || !observedResourceID.Computed || observedResourceID.Required || observedResourceID.Optional || observedResourceID.ForceNew {
		t.Fatalf("observed_resource_id schema = %#v, want computed state-only identity pin", observedResourceID)
	}
}

func TestConfigureFlinkWorkspaceCapacityBootstrapService(t *testing.T) {
	contextValue := testFlinkCapacityBootstrapContext("f-workspace", "resource-workspace", 1_700_000_900)
	raw, err := encodeFlinkWorkspaceCapacityBootstrapContext(contextValue)
	if err != nil {
		t.Fatal(err)
	}
	base := &FlinkCapacityService{api: testFlinkCapacityReadableAPI(testFlinkCapacityWorkspace(2))}
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }

	configuredAPI, err := configureFlinkWorkspaceCapacityBootstrapService(base, "f-workspace", raw, "", true, now)
	if err != nil {
		t.Fatal(err)
	}
	configured, ok := configuredAPI.(*FlinkCapacityService)
	if !ok || configured == base || configured.capacityBootstrap == nil || !configured.capacityBootstrap.allowInitialIdentityAbsence {
		t.Fatalf("configured service = %#v, want cloned bootstrap-enabled service", configuredAPI)
	}
	if base.capacityBootstrap != nil {
		t.Fatal("bootstrap configuration mutated shared base service")
	}

	strict, err := configureFlinkWorkspaceCapacityBootstrapService(base, "f-workspace", "", "", true, now)
	if err != nil || strict != base {
		t.Fatalf("empty strict context service/error = %T %v, want unchanged base", strict, err)
	}

	for _, test := range []struct {
		name       string
		service    flinkcapacity.API
		instanceID string
		raw        string
		want       string
	}{
		{name: "invalid encoding", service: base, instanceID: "f-workspace", raw: "{}", want: "invalid workspace_bootstrap_context"},
		{name: "different instance", service: base, instanceID: "f-other", raw: raw, want: "expects InstanceId"},
		{name: "non-production service", service: &flinkBootstrapSequenceAPI{}, instanceID: "f-workspace", raw: raw, want: "production Flink capacity service"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := configureFlinkWorkspaceCapacityBootstrapService(test.service, test.instanceID, test.raw, "", true, now); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("configure error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestConfigureFlinkWorkspaceCapacityBootstrapServicePinsDeferredResourceID(t *testing.T) {
	contextValue := testFlinkCapacityBootstrapContext("f-workspace", "", 1_700_000_900)
	raw, err := encodeFlinkWorkspaceCapacityBootstrapContext(contextValue)
	if err != nil {
		t.Fatal(err)
	}
	base := &FlinkCapacityService{api: testFlinkCapacityReadableAPI(testFlinkCapacityWorkspace(2))}
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }

	freshAPI, err := configureFlinkWorkspaceCapacityBootstrapService(base, "f-workspace", raw, "", true, now)
	if err != nil {
		t.Fatal(err)
	}
	fresh := freshAPI.(*FlinkCapacityService)
	if fresh.capacityBootstrap == nil || fresh.capacityBootstrap.context.ExpectedResourceID != "" || !fresh.capacityBootstrap.allowInitialIdentityAbsence {
		t.Fatalf("fresh deferred ResourceId configuration = %#v", fresh.capacityBootstrap)
	}

	readAPI, err := configureFlinkWorkspaceCapacityBootstrapService(base, "f-workspace", raw, "resource-workspace", false, now)
	if err != nil {
		t.Fatal(err)
	}
	read := readAPI.(*FlinkCapacityService)
	if read.capacityBootstrap == nil || read.capacityBootstrap.context.ExpectedResourceID != "resource-workspace" || read.capacityBootstrap.allowInitialIdentityAbsence {
		t.Fatalf("persisted ResourceId configuration = %#v", read.capacityBootstrap)
	}

	if _, err := configureFlinkWorkspaceCapacityBootstrapService(base, "f-workspace", raw, "", false, now); err == nil || !strings.Contains(err.Error(), "requires persisted observed_resource_id") {
		t.Fatalf("Read without persisted ResourceId error = %v, want fail closed", err)
	}
}

func TestFlinkWorkspaceCapacityBootstrapCreatePinsObservedResourceIDWithoutWrites(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	contextValue := testFlinkCapacityBootstrapContext(workspace.Id, "", time.Now().Add(time.Minute).Unix())
	workspace.Tags = []flink.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: contextValue.TerraformCreateToken},
		{Key: flinkworkspace.CreateIntentTagKey, Value: contextValue.CreateIntentFingerprint},
	}
	raw, err := encodeFlinkWorkspaceCapacityBootstrapContext(contextValue)
	if err != nil {
		t.Fatal(err)
	}
	api := testFlinkCapacityReadableAPI(workspace)
	withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityBootstrap().Schema, map[string]interface{}{
		"workspace_instance_id":       workspace.Id,
		"workspace_bootstrap_context": raw,
	})

	if err := resourceAliCloudFlinkWorkspaceCapacityBootstrapCreate(data, struct{}{}); err != nil {
		t.Fatal(err)
	}
	if data.Id() != workspace.Id || data.Get("observed_resource_id") != workspace.ResourceId {
		t.Fatalf("barrier state identity = %q/%q, want %q/%q", data.Id(), data.Get("observed_resource_id"), workspace.Id, workspace.ResourceId)
	}
	if testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("read-only barrier performed %d capacity write(s)", testFlinkCapacityWriteCalls(api))
	}
}

func TestWaitForFlinkWorkspaceCapacityBootstrapTreeRetriesOnlyReadiness(t *testing.T) {
	api := &flinkBootstrapSequenceAPI{readErrors: []error{
		&flinkcapacity.NotReadyError{Reason: "identity propagating"},
		&flinkcapacity.NotReadyError{Reason: "namespace propagating"},
		nil,
	}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitForFlinkWorkspaceCapacityBootstrapTree(ctx, api, "f-workspace", time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if api.reads != 3 || api.writes != 0 {
		t.Fatalf("barrier reads/writes = %d/%d, want 3/0", api.reads, api.writes)
	}
}

func TestWaitForFlinkWorkspaceCapacityBootstrapTreeFailsClosedOnTyped404(t *testing.T) {
	typed404 := flink.NewFlinkServiceErrorWithCode("request", "", "404", "workspace not found", "")
	api := &flinkBootstrapSequenceAPI{readErrors: []error{typed404}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := waitForFlinkWorkspaceCapacityBootstrapTree(ctx, api, "f-workspace", time.Millisecond)
	if !errors.Is(err, typed404) || api.reads != 1 || api.writes != 0 {
		t.Fatalf("typed 404 error/reads/writes = %T %v/%d/%d, want original/1/0", err, err, api.reads, api.writes)
	}
}

func TestWaitForFlinkWorkspaceCapacityBootstrapTreeDoesNotUseTextNotFoundFallback(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "typed 403 containing not found", err: flink.NewFlinkServiceErrorWithCode("request", "", "403", "workspace not found", "")},
		{name: "business error containing not available", err: flink.NewFlinkServiceErrorWithCode("request", "", "WorkspaceNotReady", "workspace not available", "")},
		{name: "plain text not found", err: errors.New("workspace not found")},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := &flinkBootstrapSequenceAPI{readErrors: []error{test.err}}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := waitForFlinkWorkspaceCapacityBootstrapTree(ctx, api, "f-workspace", time.Millisecond)
			if !errors.Is(err, test.err) || api.reads != 1 || api.writes != 0 {
				t.Fatalf("error/reads/writes = %T %v/%d/%d, want original/1/0", err, err, api.reads, api.writes)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityBootstrapReadValidatesPersistedProvenance(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	workspace.ResourceId = "resource-replacement"
	api := testFlinkCapacityReadableAPI(workspace)
	contextValue := testFlinkCapacityBootstrapContext("f-workspace", "resource-original", 1_700_000_900)
	raw, err := encodeFlinkWorkspaceCapacityBootstrapContext(contextValue)
	if err != nil {
		t.Fatal(err)
	}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityBootstrap().Schema, map[string]interface{}{
		"workspace_instance_id":       "f-workspace",
		"workspace_bootstrap_context": raw,
	})
	data.SetId("f-workspace")
	withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})

	err = resourceAliCloudFlinkWorkspaceCapacityBootstrapRead(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "ResourceId mismatch") {
		t.Fatalf("Read() error = %T %v, want persisted-context ResourceId mismatch", err, err)
	}
	if data.Id() != "f-workspace" || api.listNamespacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("Read() ID/namespace reads/writes = %q/%d/%d, want preserved/0/0", data.Id(), api.listNamespacesCalls, testFlinkCapacityWriteCalls(api))
	}
}

func TestFlinkWorkspaceCapacityBootstrapReadRejectsStateIdentityMismatchBeforeService(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityBootstrap().Schema, map[string]interface{}{
		"workspace_instance_id":       "f-configured",
		"workspace_bootstrap_context": flinkWorkspaceCapacityBootstrapContextStrictExisting,
	})
	data.SetId("f-state")
	err := resourceAliCloudFlinkWorkspaceCapacityBootstrapRead(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "state identity mismatch") {
		t.Fatalf("Read() error = %T %v, want state identity mismatch", err, err)
	}
	if data.Id() != "f-state" {
		t.Fatalf("Read() ID = %q, want preserved", data.Id())
	}
}

type flinkBootstrapSequenceAPI struct {
	readErrors []error
	reads      int
	writes     int
}

func (a *flinkBootstrapSequenceAPI) ReadTree(context.Context, string) (flinkcapacity.Tree, error) {
	index := a.reads
	a.reads++
	if index < len(a.readErrors) {
		return flinkcapacity.Tree{}, a.readErrors[index]
	}
	return flinkcapacity.Tree{}, nil
}

func (a *flinkBootstrapSequenceAPI) ApplyStep(context.Context, string, flinkcapacity.Step) (flinkcapacity.Operation, error) {
	a.writes++
	return flinkcapacity.Operation{}, errors.New("read-only barrier attempted a write")
}
