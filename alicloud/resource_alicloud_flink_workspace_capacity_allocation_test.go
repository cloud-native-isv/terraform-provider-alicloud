package alicloud

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestFlinkWorkspaceCapacityAllocationSchema(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatal(err)
	}
	if field := resource.Schema["workspace_instance_id"]; !field.Required || !field.ForceNew || field.Type != schema.TypeString {
		t.Fatalf("workspace_instance_id schema = %#v", field)
	}
	for _, name := range []string{"fixed_cu", "cross_zone_fixed_cu", "max_cu_limit"} {
		field := resource.Schema[name]
		if field == nil || !field.Required || field.Type != schema.TypeInt {
			t.Fatalf("%s schema = %#v", name, field)
		}
	}
	namespace := resource.Schema["namespace"]
	if namespace == nil || !namespace.Required || namespace.Type != schema.TypeSet || namespace.MinItems != 1 {
		t.Fatalf("namespace schema = %#v", namespace)
	}
	if namespace.Set == nil {
		t.Fatal("namespace must use schema.HashResource")
	}
	for _, name := range []string{"implicit_topology_hash", "observed_capacity_tree"} {
		field := resource.Schema[name]
		if field == nil || !field.Computed {
			t.Fatalf("%s must be computed: %#v", name, field)
		}
	}
	if got := *resource.Timeouts.Create; got != 60*time.Minute {
		t.Fatalf("create timeout = %s", got)
	}
}

func TestExpandFlinkWorkspaceCapacityAllocationDesired(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{name: "single zone", config: flinkAllocationConfig(false), wantErr: ""},
		{name: "cross zone", config: flinkAllocationConfig(true), wantErr: ""},
		{name: "mixed pools", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) { c["cross_zone_fixed_cu"] = 2 }), wantErr: "pure"},
		{name: "fractional fixed rejected", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) { c["fixed_cu"] = 1.5 }), wantErr: "integer CU"},
		{name: "namespace fixed below one", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) {
			c["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 0, "max_cu_limit": 2}}
		}), wantErr: "at least"},
		{name: "namespace sums must match", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) {
			c["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 1, "max_cu_limit": 2}}
		}), wantErr: "sum"},
		{name: "max below fixed", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) { c["max_cu_limit"] = 1 }), wantErr: "greater than or equal"},
		{name: "duplicate namespace", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) {
			c["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 2, "max_cu_limit": 2}, map[string]interface{}{"name": "default", "fixed_cu": 2, "max_cu_limit": 2}}
		}), wantErr: "duplicate"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := expandFlinkWorkspaceCapacityAllocationDesired(tc.config)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := flinkcapacity.ValidateDesired(got); err != nil {
				t.Fatalf("expanded desired tree is invalid: %v", err)
			}
			if got.Namespaces[0].CrossZone != (tc.config["cross_zone_fixed_cu"].(int) > 0) {
				t.Fatalf("namespace type = %t", got.Namespaces[0].CrossZone)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationCustomizeDiffRejectsInvalidDesired(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	for _, tc := range []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{name: "mixed", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) { c["cross_zone_fixed_cu"] = 2 }), wantErr: "pure"},
		{name: "duplicate names", config: mutateFlinkAllocationConfig(false, func(c map[string]interface{}) {
			c["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 2, "max_cu_limit": 2}, map[string]interface{}{"name": "default", "fixed_cu": 1, "max_cu_limit": 1}}
		}), wantErr: "duplicate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resource.Diff(nil, terraform.NewResourceConfigRaw(tc.config), nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Diff() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationTopologyHash(t *testing.T) {
	actual := flinkAllocationTree(false, "PRE")
	first, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	actual.Namespaces[0].Used = 1.75
	actual.Namespaces[0].Queues[0].Used = 1.25
	second, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("used CU unexpectedly affects topology hash: %q != %q", first, second)
	}
	actual.Namespaces[0].CrossZone = true
	third, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	if third == first {
		t.Fatal("namespace CU type drift did not affect topology hash")
	}
	actual.Namespaces = append(actual.Namespaces, flinkcapacity.Namespace{
		Name: "second", CrossZone: true, Capacity: actual.Namespaces[0].Capacity,
		Queues: []flinkcapacity.Queue{{Name: "z", Capacity: actual.Namespaces[0].Capacity}, {Name: "a", Capacity: actual.Namespaces[0].Capacity}},
	})
	fourth, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	actual.Namespaces[1].Queues[0], actual.Namespaces[1].Queues[1] = actual.Namespaces[1].Queues[1], actual.Namespaces[1].Queues[0]
	actual.Namespaces[0], actual.Namespaces[1] = actual.Namespaces[1], actual.Namespaces[0]
	fifth, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	if fourth != fifth {
		t.Fatalf("topology hash depends on namespace or queue order: %q != %q", fourth, fifth)
	}
}

func TestFlinkWorkspaceCapacityAllocationFlattenObservedTree(t *testing.T) {
	tree := flinkAllocationTree(true, "PRE")
	tree.Namespaces[0].Used = 1.25
	tree.Namespaces[0].Queues[0].Used = 0.75
	flattened := flattenFlinkWorkspaceCapacityAllocationObservedTree(tree)
	workspace := firstTestBlock(flattened)
	crossZone, _ := numberAsFloat(workspace["cross_zone_fixed_cu"])
	used, _ := numberAsFloat(workspace["used_cu"])
	if crossZone != 2 || used != 0.0 {
		t.Fatalf("workspace observation = %#v", workspace)
	}
	namespace := firstTestBlock(workspace["namespace"])
	if namespace["cross_zone"] != true || namespace["used_cu"] != 1.25 {
		t.Fatalf("namespace observation = %#v", namespace)
	}
}

func TestFlinkWorkspaceCapacityAllocationDeleteIsNoop(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, flinkAllocationConfig(false))
	data.SetId("f-test")
	if err := resourceAliCloudFlinkWorkspaceCapacityAllocationDelete(data, nil); err != nil {
		t.Fatal(err)
	}
	if data.Id() != "" {
		t.Fatalf("Delete() ID = %q", data.Id())
	}
}

func TestFlinkWorkspaceCapacityAllocationImporterSetsWorkspaceID(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
	data.SetId("f-import")
	states, err := resource.Importer.State(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 || states[0].Id() != "f-import" || states[0].Get("workspace_instance_id") != "f-import" {
		t.Fatalf("import state = %#v", states)
	}
}

func TestFlinkWorkspaceCapacityAllocationCreateWritesActualStateBeforeError(t *testing.T) {
	actual := flinkAllocationTree(false, "PRE")
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{
		"workspace_instance_id": "f-test", "fixed_cu": 4, "cross_zone_fixed_cu": 0, "max_cu_limit": 4,
		"namespace": []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 4, "max_cu_limit": 4}},
	})
	api := &flinkAllocationFakeAPI{tree: actual, applyErr: errors.New("write failed")}
	withFlinkAllocationFakeService(t, api)
	err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("Create() error = %v", err)
	}
	if data.Id() != "f-test" {
		t.Fatalf("Create() did not preserve ID before reconciliation failure: %q", data.Id())
	}
	if got := data.Get("fixed_cu"); got != 2 {
		t.Fatalf("failure state did not preserve last actual fixed CU: %#v", got)
	}
	wantHash, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
	if err != nil {
		t.Fatal(err)
	}
	if got := data.Get("implicit_topology_hash"); got != wantHash {
		t.Fatalf("failure state hash = %#v, want actual %q", got, wantHash)
	}
}

func TestFlinkWorkspaceCapacityAllocationReadImportsActualTree(t *testing.T) {
	actual := flinkAllocationTree(true, "PRE")
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
	data.SetId("f-test")
	withFlinkAllocationFakeService(t, &flinkAllocationFakeAPI{tree: actual})
	if err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{}); err != nil {
		t.Fatal(err)
	}
	if got := data.Get("workspace_instance_id"); got != "f-test" {
		t.Fatalf("Read() workspace_instance_id = %#v", got)
	}
	if got := data.Get("cross_zone_fixed_cu"); got != 2 {
		t.Fatalf("Read() cross-zone CU = %#v", got)
	}
}

func TestFlinkWorkspaceCapacityAllocationReadRejectsPostBeforeStateWrites(t *testing.T) {
	actual := flinkAllocationTree(false, "POST")
	api := &flinkAllocationFakeAPI{tree: actual}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
	data.SetId("f-post")
	withFlinkAllocationFakeService(t, api)

	err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "only PRE") {
		t.Fatalf("Read() error = %v, want PRE-only rejection", err)
	}
	if api.writes != 0 {
		t.Fatalf("Read() wrote cloud capacity before rejecting POST: %d write(s)", api.writes)
	}
	if got := data.Get("workspace_instance_id"); got != "" {
		t.Fatalf("Read() wrote workspace_instance_id before rejecting POST: %#v", got)
	}
	if got := data.Get("fixed_cu"); got != 0 {
		t.Fatalf("Read() wrote fixed_cu before rejecting POST: %#v", got)
	}
	if got := data.Get("implicit_topology_hash"); got != "" {
		t.Fatalf("Read() wrote implicit_topology_hash before rejecting POST: %#v", got)
	}
}

func TestFlinkWorkspaceCapacityAllocationImportReadRejectsPostWithoutCapacityStateWrites(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
	data.SetId("f-post")
	states, err := resource.Importer.State(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	api := &flinkAllocationFakeAPI{tree: flinkAllocationTree(false, "POST")}
	withFlinkAllocationFakeService(t, api)
	err = resourceAliCloudFlinkWorkspaceCapacityAllocationRead(states[0], struct{}{})
	if err == nil || !strings.Contains(err.Error(), "only PRE") {
		t.Fatalf("import Read() error = %v, want PRE-only rejection", err)
	}
	if api.writes != 0 || states[0].Get("fixed_cu") != 0 || states[0].Get("implicit_topology_hash") != "" {
		t.Fatalf("import Read() mutated capacity state or cloud: writes=%d fixed=%#v hash=%#v", api.writes, states[0].Get("fixed_cu"), states[0].Get("implicit_topology_hash"))
	}
}

func TestFlinkWorkspaceCapacityAllocationReconcileReportsStateWriteError(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	resource.Schema["implicit_topology_hash"] = &schema.Schema{Type: schema.TypeInt, Computed: true}
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"workspace_instance_id": "f-test", "fixed_cu": 4, "cross_zone_fixed_cu": 0, "max_cu_limit": 4,
		"namespace": []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 4, "max_cu_limit": 4}},
	})
	api := &flinkAllocationFakeAPI{tree: flinkAllocationTree(false, "PRE"), applyErr: errors.New("reconciliation exploded")}
	withFlinkAllocationFakeService(t, api)

	err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "reconciliation exploded") || !strings.Contains(err.Error(), "state") {
		t.Fatalf("Create() error = %v, want reconciliation and state write contexts", err)
	}
}

func TestFlinkWorkspaceCapacityAllocationDiffDetectsImplicitTopologyDrift(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*flinkcapacity.Tree)
	}{
		{name: "queue only", mutate: func(tree *flinkcapacity.Tree) {
			capacity := flinkcapacity.Capacity{Fixed: 2, Limit: 4}
			tree.Namespaces[0].Queues[0].Capacity = &capacity
		}},
		{name: "namespace type only", mutate: func(tree *flinkcapacity.Tree) { tree.Namespaces[0].CrossZone = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
			config := flinkAllocationConfig(false)
			actual := flinkAllocationTree(false, "PRE")
			tc.mutate(&actual)
			stateData := schema.TestResourceDataRaw(t, resource.Schema, config)
			stateData.SetId("f-test")
			actualHash, err := flinkWorkspaceCapacityAllocationTopologyHash(actual)
			if err != nil {
				t.Fatal(err)
			}
			if err := stateData.Set("implicit_topology_hash", actualHash); err != nil {
				t.Fatal(err)
			}
			diff, err := resource.Diff(stateData.State(), terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			attribute := diff.Attributes["implicit_topology_hash"]
			if attribute == nil || attribute.Old == attribute.New || attribute.RequiresNew {
				t.Fatalf("implicit topology drift diff = %#v", attribute)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationDiffMarksUnknownHashComputed(t *testing.T) {
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"
	for _, tc := range []struct {
		name   string
		mutate func(map[string]interface{})
	}{
		{name: "top level", mutate: func(config map[string]interface{}) { config["fixed_cu"] = unknown }},
		{name: "nested", mutate: func(config map[string]interface{}) { firstTestBlock(config["namespace"])["fixed_cu"] = unknown }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
			config := flinkAllocationConfig(false)
			tc.mutate(config)
			diff, err := resource.Diff(nil, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			attribute := diff.Attributes["implicit_topology_hash"]
			if attribute == nil || !attribute.NewComputed {
				t.Fatalf("implicit_topology_hash diff = %#v, want NewComputed", attribute)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationDiffCanonicalizesEquivalentNumbers(t *testing.T) {
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()
	stateConfig := flinkAllocationConfig(false)
	stateConfig["fixed_cu"] = 1
	stateConfig["max_cu_limit"] = 1
	stateConfig["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 1, "max_cu_limit": 1}}
	stateData := schema.TestResourceDataRaw(t, resource.Schema, stateConfig)
	stateData.SetId("f-test")
	tree, err := expandFlinkWorkspaceCapacityAllocationDesired(stateConfig)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := flinkWorkspaceCapacityAllocationTopologyHash(tree)
	if err != nil {
		t.Fatal(err)
	}
	if err := stateData.Set("implicit_topology_hash", hash); err != nil {
		t.Fatal(err)
	}
	floatConfig := flinkAllocationConfig(false)
	floatConfig["fixed_cu"] = 1.0
	floatConfig["max_cu_limit"] = 1.0
	floatConfig["namespace"] = []interface{}{map[string]interface{}{"name": "default", "fixed_cu": 1.0, "max_cu_limit": 1.0}}
	diff, err := resource.Diff(stateData.State(), terraform.NewResourceConfigRaw(floatConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	if attribute := diff.Attributes["implicit_topology_hash"]; attribute != nil && attribute.Old != attribute.New {
		t.Fatalf("equivalent numeric spellings changed topology hash: %#v", attribute)
	}
}

func TestFlinkWorkspaceCapacityAllocationSerializesSameWorkspace(t *testing.T) {
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- withFlinkWorkspaceCapacityAllocationLock("f-test", func() error {
			close(firstEntered)
			<-releaseFirst
			return nil
		})
	}()
	<-firstEntered
	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- withFlinkWorkspaceCapacityAllocationLock("f-test", func() error { close(secondEntered); return nil })
	}()
	select {
	case <-secondEntered:
		close(releaseFirst)
		t.Fatal("second allocation callback entered before the first released its workspace lock")
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	select {
	case <-secondEntered:
	case <-time.After(time.Second):
		t.Fatal("second allocation callback did not enter after lock release")
	}
	if err := <-secondDone; err != nil {
		t.Fatal(err)
	}
}

type flinkAllocationFakeAPI struct {
	tree     flinkcapacity.Tree
	applyErr error
	writes   int
}

func (f *flinkAllocationFakeAPI) ReadTree(context.Context, string) (flinkcapacity.Tree, error) {
	return f.tree, nil
}
func (f *flinkAllocationFakeAPI) ApplyStep(context.Context, string, flinkcapacity.Step) (flinkcapacity.Operation, error) {
	f.writes++
	return flinkcapacity.Operation{}, f.applyErr
}

func withFlinkAllocationFakeService(t *testing.T, api flinkcapacity.API) {
	t.Helper()
	previous := newFlinkWorkspaceCapacityAllocationService
	newFlinkWorkspaceCapacityAllocationService = func(interface{}) (flinkcapacity.API, error) { return api, nil }
	t.Cleanup(func() { newFlinkWorkspaceCapacityAllocationService = previous })
}

func flinkAllocationConfig(crossZone bool) map[string]interface{} {
	config := map[string]interface{}{
		"workspace_instance_id": "f-test",
		"fixed_cu":              2,
		"cross_zone_fixed_cu":   0,
		"max_cu_limit":          2,
		"namespace": []interface{}{map[string]interface{}{
			"name": "default", "fixed_cu": 2, "max_cu_limit": 2,
		}},
	}
	if crossZone {
		config["fixed_cu"] = 0
		config["cross_zone_fixed_cu"] = 2
	}
	return config
}

func mutateFlinkAllocationConfig(crossZone bool, mutate func(map[string]interface{})) map[string]interface{} {
	config := flinkAllocationConfig(crossZone)
	mutate(config)
	return config
}

func flinkAllocationTree(crossZone bool, chargeType string) flinkcapacity.Tree {
	fixed, cross := flinkcapacity.CU(4), flinkcapacity.CU(0)
	if crossZone {
		fixed, cross = 0, 4
	}
	capacity := flinkcapacity.Capacity{Fixed: 4, Limit: 4}
	return flinkcapacity.Tree{
		ChargeType: chargeType,
		Workspace:  flinkcapacity.WorkspaceCapacity{HA: crossZone, FixedCU: fixed, CrossZoneFixedCU: cross, Limit: 4},
		Namespaces: []flinkcapacity.Namespace{{
			Name: "default", CrossZone: crossZone, Capacity: &capacity,
			Queues: []flinkcapacity.Queue{{Name: "default-queue", Capacity: &capacity}},
		}},
	}
}

func firstTestBlock(value interface{}) map[string]interface{} {
	items, _ := value.([]interface{})
	if len(items) == 0 {
		return nil
	}
	return items[0].(map[string]interface{})
}
