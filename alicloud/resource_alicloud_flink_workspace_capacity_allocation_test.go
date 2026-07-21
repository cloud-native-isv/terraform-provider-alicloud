package alicloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
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

type flinkAllocationCUFieldBoundaryCase struct {
	name             string
	field            string
	namespace        bool
	schemaFieldPath  string
	diffErrorParts   []string
	expandErrorParts []string
	acceptanceConfig func() map[string]interface{}
	rejectionConfig  func() map[string]interface{}
}

func flinkAllocationCUFieldBoundaryCases() []flinkAllocationCUFieldBoundaryCase {
	const maxSafeCU = flinkMaxCUBeforeInt32MemoryOverflow
	const overBoundCU = maxSafeCU + 1
	return []flinkAllocationCUFieldBoundaryCase{
		{
			name:             "workspace fixed_cu",
			field:            "fixed_cu",
			schemaFieldPath:  "fixed_cu",
			diffErrorParts:   []string{"fixed_cu", "int32"},
			expandErrorParts: []string{"fixed_cu", "int32"},
			acceptanceConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, maxSafeCU, maxSafeCU)
			},
			// The target bound is rejected before the later max/sum checks.  Those
			// checks cannot remain valid when fixed_cu alone exceeds maxSafeCU,
			// because max_cu_limit must be at least fixed_cu.
			rejectionConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(overBoundCU, 0, maxSafeCU, maxSafeCU, maxSafeCU)
			},
		},
		{
			name:             "workspace cross_zone_fixed_cu",
			field:            "cross_zone_fixed_cu",
			schemaFieldPath:  "cross_zone_fixed_cu",
			diffErrorParts:   []string{"cross_zone_fixed_cu", "int32"},
			expandErrorParts: []string{"cross_zone_fixed_cu", "int32"},
			acceptanceConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(0, maxSafeCU, maxSafeCU, maxSafeCU, maxSafeCU)
			},
			// As with fixed_cu, the target bound is rejected before the later
			// max/sum checks that cannot be valid with only this value over bound.
			rejectionConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(0, overBoundCU, maxSafeCU, maxSafeCU, maxSafeCU)
			},
		},
		{
			name:             "workspace max_cu_limit",
			field:            "max_cu_limit",
			schemaFieldPath:  "max_cu_limit",
			diffErrorParts:   []string{"max_cu_limit", "int32"},
			expandErrorParts: []string{"max_cu_limit", "int32"},
			acceptanceConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, maxSafeCU, maxSafeCU)
			},
			rejectionConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, overBoundCU, maxSafeCU, maxSafeCU)
			},
		},
		{
			name:             "namespace fixed_cu",
			field:            "fixed_cu",
			namespace:        true,
			schemaFieldPath:  "namespace.default.fixed_cu",
			diffErrorParts:   []string{"namespace", "fixed_cu", "int32"},
			expandErrorParts: []string{`namespace "default" fixed_cu`, "int32"},
			acceptanceConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, maxSafeCU, maxSafeCU)
			},
			// Workspace fixed/max remain at maxSafeCU so the namespace parser
			// reaches the isolated target before it evaluates namespace sums.
			rejectionConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, overBoundCU, maxSafeCU)
			},
		},
		{
			name:             "namespace max_cu_limit",
			field:            "max_cu_limit",
			namespace:        true,
			schemaFieldPath:  "namespace.default.max_cu_limit",
			diffErrorParts:   []string{"namespace", "max_cu_limit", "int32"},
			expandErrorParts: []string{`namespace "default" max_cu_limit`, "int32"},
			acceptanceConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, maxSafeCU, maxSafeCU)
			},
			// Workspace fixed/max remain at maxSafeCU so the namespace parser
			// reaches the isolated target before it evaluates namespace sums.
			rejectionConfig: func() map[string]interface{} {
				return flinkAllocationBoundaryConfig(maxSafeCU, 0, maxSafeCU, maxSafeCU, overBoundCU)
			},
		},
	}
}

func TestFlinkWorkspaceCapacityAllocationSchemaValidatesCUInt32MemoryBounds(t *testing.T) {
	const maxSafeCU = 536870911 // floor(math.MaxInt32 / 4)
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()

	for _, tc := range flinkAllocationCUFieldBoundaryCases() {
		t.Run(tc.name, func(t *testing.T) {
			field := resource.Schema[tc.field]
			if tc.namespace {
				field = resource.Schema["namespace"].Elem.(*schema.Resource).Schema[tc.field]
			}
			if field == nil || field.ValidateFunc == nil {
				t.Fatalf("%s schema must validate CU bounds: %#v", tc.name, field)
			}
			for _, test := range []struct {
				name    string
				value   int
				wantErr bool
			}{
				{name: "maximum safe CU", value: maxSafeCU},
				{name: "one CU above maximum", value: maxSafeCU + 1, wantErr: true},
			} {
				t.Run(test.name, func(t *testing.T) {
					_, errs := field.ValidateFunc(test.value, tc.schemaFieldPath)
					if (len(errs) > 0) != test.wantErr {
						t.Fatalf("%s ValidateFunc(%d) errors = %v, wantErr %t", tc.name, test.value, errs, test.wantErr)
					}
					if test.wantErr && (!strings.Contains(errs[0].Error(), tc.schemaFieldPath) || !strings.Contains(errs[0].Error(), "int32")) {
						t.Fatalf("%s ValidateFunc(%d) error = %v, want exact field path %q and int32 cause", tc.name, test.value, errs[0], tc.schemaFieldPath)
					}
				})
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationRejectsCUThatOverflowsInt32MemoryAtPlanTime(t *testing.T) {
	const maxSafeCU = 536870911 // floor(math.MaxInt32 / 4)
	resource := resourceAliCloudFlinkWorkspaceCapacityAllocation()

	for _, tc := range flinkAllocationCUFieldBoundaryCases() {
		t.Run(tc.name, func(t *testing.T) {
			for _, test := range []struct {
				name    string
				config  map[string]interface{}
				wantErr bool
			}{
				{name: "maximum safe CU in a valid configuration", config: tc.acceptanceConfig()},
				{name: "only target CU above maximum", config: tc.rejectionConfig(), wantErr: true},
			} {
				t.Run(test.name, func(t *testing.T) {
					if test.wantErr {
						assertFlinkAllocationCURejectionIsolated(t, test.config, tc)
					}
					_, err := resource.Diff(nil, terraform.NewResourceConfigRaw(test.config), nil)
					if test.wantErr {
						assertFlinkErrorContains(t, err, tc.diffErrorParts...)
						return
					}
					if err != nil {
						t.Fatalf("maximum safe %s Diff() error = %v", tc.name, err)
					}
				})
			}
		})
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

func TestExpandFlinkWorkspaceCapacityAllocationDesiredRejectsCUThatOverflowsInt32Memory(t *testing.T) {
	const maxSafeCU = 536870911 // floor(math.MaxInt32 / 4)
	for _, tc := range flinkAllocationCUFieldBoundaryCases() {
		t.Run(tc.name, func(t *testing.T) {
			for _, test := range []struct {
				name    string
				config  map[string]interface{}
				wantErr bool
			}{
				{name: "maximum safe CU in a valid configuration", config: tc.acceptanceConfig()},
				{name: "only target CU above maximum", config: tc.rejectionConfig(), wantErr: true},
			} {
				t.Run(test.name, func(t *testing.T) {
					if test.wantErr {
						assertFlinkAllocationCURejectionIsolated(t, test.config, tc)
					}
					desired, err := expandFlinkWorkspaceCapacityAllocationDesired(test.config)
					if test.wantErr {
						assertFlinkErrorContains(t, err, tc.expandErrorParts...)
						return
					}
					if err != nil {
						t.Fatalf("maximum safe %s expansion error = %v", tc.name, err)
					}
					if err := flinkcapacity.ValidateDesired(desired); err != nil {
						t.Fatalf("maximum safe %s expanded tree is invalid: %v", tc.name, err)
					}
				})
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationCreateRejectsOverBoundCUBeforeWrite(t *testing.T) {
	const overBoundCU = 536870912 // floor(math.MaxInt32 / 4) + 1
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{
		"workspace_instance_id": "f-test", "fixed_cu": overBoundCU, "cross_zone_fixed_cu": 0, "max_cu_limit": overBoundCU,
		"namespace": []interface{}{map[string]interface{}{
			"name": "default", "fixed_cu": overBoundCU, "max_cu_limit": overBoundCU,
		}},
	})
	api := &flinkAllocationFakeAPI{tree: flinkAllocationTree(false, "PRE"), applyErr: errors.New("write should not happen")}
	withFlinkAllocationFakeService(t, api)

	err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "int32") {
		t.Fatalf("Create() error = %v, want int32 memory validation", err)
	}
	if api.writes != 0 {
		t.Fatalf("Create() made %d writes after rejecting over-bound CU", api.writes)
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

func TestFlinkWorkspaceCapacityAllocationReadPreservesIDForUncorroboratedGetErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		getErr error
	}{
		{
			name:   "typed 403 containing not found",
			getErr: flink.NewFlinkServiceErrorWithCode("get-request", "", "403", "workspace not found", ""),
		},
		{
			name:   "business error containing not available",
			getErr: errors.New("workspace is not available for this account"),
		},
		{
			name:   "transport error containing 404",
			getErr: errors.New("workspace transport failed with status code 404"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = tc.getErr
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
			data.SetId("f-workspace")
			withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})

			err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{})
			if err == nil {
				t.Fatalf("Read() error = nil, want original Get failure")
			}
			if data.Id() != "f-workspace" {
				t.Fatalf("Read() ID = %q, want preserved after uncorroborated Get failure", data.Id())
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 0 {
				t.Fatalf("Read() calls: get=%d list=%d namespaces=%d, want 1/0/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("Read() made %d capacity writes after uncorroborated Get failure", writes)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationReadPreservesIDForListErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		listErr error
	}{
		{
			name:    "typed 403 containing not found",
			listErr: flink.NewFlinkServiceErrorWithCode("list-request", "", "403", "workspace list not found for this principal", ""),
		},
		{
			name:    "business error containing not available",
			listErr: errors.New("workspace listing is not available"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			getErr := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = getErr
			api.listWorkspacesErr = tc.listErr
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
			data.SetId("f-workspace")
			withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})

			err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{})
			if err == nil {
				t.Fatalf("Read() error = nil, want List failure %v", tc.listErr)
			}
			if data.Id() != "f-workspace" {
				t.Fatalf("Read() ID = %q, want preserved after List failure %v", data.Id(), tc.listErr)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 {
				t.Fatalf("Read() calls: get=%d list=%d namespaces=%d, want 1/1/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("Read() made %d capacity writes after List failure", writes)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationReadPreservesIDForDownstream404(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	for _, tc := range []struct {
		name       string
		fallback   bool
		downstream string
	}{
		{name: "exact Get then namespace 404", downstream: "namespace"},
		{name: "exact List then namespace 404", fallback: true, downstream: "namespace"},
		{name: "exact Get then queue 404", downstream: "queue"},
		{name: "exact List then queue 404", fallback: true, downstream: "queue"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(workspace)
			if tc.fallback {
				api.workspace = nil
				api.getWorkspaceErr = flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
				api.listedWorkspaces = []flink.Workspace{*workspace}
			}
			if tc.downstream == "namespace" {
				api.listNamespacesErr = flink.NewFlinkServiceErrorWithCode("namespace-request", "", "404", "namespace not found", "")
			} else {
				api.listTargetErrors = map[string]error{
					"default": flink.NewFlinkServiceErrorWithCode("queue-request", "", "404", "queue not found", ""),
				}
			}
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
			data.SetId(workspace.Id)
			withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})

			err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{})
			if err == nil {
				t.Fatalf("Read() error = nil, want downstream %s 404", tc.downstream)
			}
			if data.Id() != workspace.Id {
				t.Fatalf("Read() ID = %q, want %q preserved after downstream %s 404", data.Id(), workspace.Id, tc.downstream)
			}
			wantListCalls := 0
			if tc.fallback {
				wantListCalls = 1
			}
			wantTargetCalls := 0
			if tc.downstream == "queue" {
				wantTargetCalls = 1
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != wantListCalls || api.listNamespacesCalls != 1 || api.listTargetCalls["default"] != wantTargetCalls {
				t.Fatalf("Read() calls: get=%d list=%d namespaces=%d targets=%d, want 1/%d/1/%d", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, api.listTargetCalls["default"], wantListCalls, wantTargetCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("Read() made %d capacity writes after downstream %s 404", writes, tc.downstream)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationReadClearsIDOnlyForAuthoritativeWorkspaceAbsence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		getErr  error
		present bool
	}{
		{
			name:   "direct typed Get 404 and absent List",
			getErr: flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", ""),
		},
		{
			name: "wrapped typed Get 404 and absent List",
			getErr: fmt.Errorf("GetWorkspace failed: %w",
				flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")),
		},
		{
			name:    "typed Get 404 and exact listed Workspace not ready",
			getErr:  flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", ""),
			present: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = tc.getErr
			if tc.present {
				workspace := *testFlinkCapacityWorkspace(2)
				workspace.Status = "CREATING"
				api.listedWorkspaces = []flink.Workspace{workspace}
			}
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, map[string]interface{}{})
			data.SetId("f-workspace")
			withFlinkAllocationFakeService(t, &FlinkCapacityService{api: api})

			err := resourceAliCloudFlinkWorkspaceCapacityAllocationRead(data, struct{}{})
			if tc.present {
				if err == nil {
					t.Fatal("Read() error = nil, want readiness error for exact listed Workspace")
				}
				if data.Id() != "f-workspace" {
					t.Fatalf("Read() ID = %q, want preserved for exact listed Workspace", data.Id())
				}
			} else {
				if err != nil {
					t.Fatalf("Read() error = %v, want authoritative removal", err)
				}
				if data.Id() != "" {
					t.Fatalf("Read() ID = %q, want cleared after authoritative Workspace absence", data.Id())
				}
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 {
				t.Fatalf("Read() calls: get=%d list=%d namespaces=%d, want 1/1/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("Read() made %d capacity writes while resolving Workspace presence", writes)
			}
		})
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
	tree       flinkcapacity.Tree
	readErrors []error
	readFunc   func(context.Context, string) (flinkcapacity.Tree, error)
	reads      int
	applyErr   error
	writes     int
}

func (f *flinkAllocationFakeAPI) ReadTree(ctx context.Context, instanceID string) (flinkcapacity.Tree, error) {
	f.reads++
	if f.readFunc != nil {
		return f.readFunc(ctx, instanceID)
	}
	if len(f.readErrors) > 0 {
		err := f.readErrors[0]
		f.readErrors = f.readErrors[1:]
		return flinkcapacity.Tree{}, err
	}
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

func TestFlinkWorkspaceCapacityAllocationFreshCreateAndUpdateUseSameStatelessReadSemantics(t *testing.T) {
	visibleData := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, flinkAllocationConfig(false))
	visibleAPI := &flinkAllocationFakeAPI{tree: flinkAllocationTree(false, "PRE")}
	withFlinkAllocationFakeService(t, visibleAPI)
	if err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(visibleData, struct{}{}); err != nil {
		t.Fatalf("initial Create() error = %v", err)
	}
	if visibleAPI.reads != 1 || visibleAPI.writes != 0 {
		t.Fatalf("initial Create() calls: reads=%d writes=%d, want 1/0", visibleAPI.reads, visibleAPI.writes)
	}

	for _, tc := range []struct {
		name     string
		callback func(*schema.ResourceData, interface{}) error
	}{
		{name: "fresh Create callback", callback: resourceAliCloudFlinkWorkspaceCapacityAllocationCreate},
		{name: "Update callback", callback: resourceAliCloudFlinkWorkspaceCapacityAllocationUpdate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, flinkAllocationConfig(false))
			data.SetId("f-test")
			missing := GetNotFoundErrorFromString("workspace was authoritatively deleted")
			api := &flinkAllocationFakeAPI{
				tree:       flinkAllocationTree(false, "PRE"),
				readErrors: []error{missing, errors.New("unexpected callback-local visibility retry")},
			}
			withFlinkAllocationFakeService(t, api)

			err := tc.callback(data, struct{}{})
			if err == nil || !strings.Contains(err.Error(), "workspace was authoritatively deleted") {
				t.Fatalf("callback error = %v, want authoritative Workspace deletion", err)
			}
			if api.reads != 1 {
				t.Fatalf("callback ReadTree calls = %d, want one stateless authoritative read", api.reads)
			}
			if api.writes != 0 {
				t.Fatalf("callback made %d writes after Workspace deletion", api.writes)
			}
		})
	}
}

func TestFlinkWorkspaceCapacityAllocationCreateDoesNotRetryBusinessError(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, flinkAllocationConfig(false))
	api := &flinkAllocationFakeAPI{
		tree:       flinkAllocationTree(false, "PRE"),
		readErrors: []error{errors.New("forbidden to read workspace")},
	}
	withFlinkAllocationFakeService(t, api)

	err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "forbidden to read workspace") {
		t.Fatalf("Create() error = %v, want original business error", err)
	}
	if api.reads != 1 {
		t.Fatalf("Create() ReadTree calls = %d, want one non-retried business error", api.reads)
	}
	if api.writes != 0 {
		t.Fatalf("Create() made %d writes after business error", api.writes)
	}
}

func TestFlinkWorkspaceCapacityAllocationPostWriteReadUsesStatelessServiceSemantics(t *testing.T) {
	actual := flinkAllocationTree(false, "PRE")
	desiredConfig := flinkAllocationConfig(false)
	desiredConfig["fixed_cu"] = 3
	desiredConfig["max_cu_limit"] = 3
	desiredConfig["namespace"] = []interface{}{map[string]interface{}{
		"name": "default", "fixed_cu": 3, "max_cu_limit": 3,
	}}
	missing := GetNotFoundErrorFromString("workspace disappeared after capacity write")
	api := &flinkAllocationFakeAPI{tree: actual}
	api.readFunc = func(context.Context, string) (flinkcapacity.Tree, error) {
		if api.reads == 1 {
			return actual, nil
		}
		return flinkcapacity.Tree{}, missing
	}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspaceCapacityAllocation().Schema, desiredConfig)
	withFlinkAllocationFakeService(t, api)

	err := resourceAliCloudFlinkWorkspaceCapacityAllocationCreate(data, struct{}{})
	if err == nil || !strings.Contains(err.Error(), "workspace disappeared after capacity write") {
		t.Fatalf("Create() error = %v, want post-write Workspace disappearance", err)
	}
	if api.writes != 1 {
		t.Fatalf("Create() writes = %d, want exactly one non-replayed write", api.writes)
	}
	if api.reads != 3 {
		t.Fatalf("Create() ReadTree calls = %d, want initial, post-write, and final recovery reads", api.reads)
	}
}

func TestFlinkWorkspaceCapacityAllocationStatelessReadinessRetryHonorsContextCancellation(t *testing.T) {
	api := &flinkAllocationFakeAPI{}
	api.readFunc = func(context.Context, string) (flinkcapacity.Tree, error) {
		return flinkcapacity.Tree{}, &flinkcapacity.NotReadyError{Reason: "corroborated Workspace is not ready"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := (flinkcapacity.Reconciler{
		API:          api,
		PollInterval: time.Hour,
	}).ReconcileAuthoritative(ctx, "f-test", flinkAllocationTree(false, "PRE"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ReconcileAuthoritative() error = %v, want context deadline", err)
	}
	if api.writes != 0 {
		t.Fatalf("ReconcileAuthoritative() made %d writes before workspace visibility", api.writes)
	}
	if api.reads != 2 {
		t.Fatalf("ReconcileAuthoritative() ReadTree calls = %d, want initial and final recovery reads", api.reads)
	}
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

func flinkAllocationBoundaryConfig(workspaceFixed, workspaceCrossZone, workspaceLimit, namespaceFixed, namespaceLimit int) map[string]interface{} {
	return map[string]interface{}{
		"workspace_instance_id": "f-test",
		"fixed_cu":              workspaceFixed,
		"cross_zone_fixed_cu":   workspaceCrossZone,
		"max_cu_limit":          workspaceLimit,
		"namespace": []interface{}{map[string]interface{}{
			"name": "default", "fixed_cu": namespaceFixed, "max_cu_limit": namespaceLimit,
		}},
	}
}

func assertFlinkErrorContains(t *testing.T, err error, parts ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, part := range parts {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("error = %v, want substring %q", err, part)
		}
	}
}

func assertFlinkAllocationCURejectionIsolated(t *testing.T, config map[string]interface{}, tc flinkAllocationCUFieldBoundaryCase) {
	t.Helper()
	const maxSafeCU = flinkMaxCUBeforeInt32MemoryOverflow
	overBoundCU := maxSafeCU + 1
	namespace := firstTestBlock(config["namespace"])
	if namespace == nil {
		t.Fatal("rejection config must contain one namespace")
	}
	for _, field := range []struct {
		name      string
		namespace bool
		value     interface{}
	}{
		{name: "fixed_cu", value: config["fixed_cu"]},
		{name: "cross_zone_fixed_cu", value: config["cross_zone_fixed_cu"]},
		{name: "max_cu_limit", value: config["max_cu_limit"]},
		{name: "fixed_cu", namespace: true, value: namespace["fixed_cu"]},
		{name: "max_cu_limit", namespace: true, value: namespace["max_cu_limit"]},
	} {
		value, ok := field.value.(int)
		if !ok {
			t.Fatalf("%s rejection config value = %#v, want int", field.name, field.value)
		}
		isTarget := field.name == tc.field && field.namespace == tc.namespace
		if isTarget && value != overBoundCU {
			t.Fatalf("target %s rejection config value = %d, want %d", tc.name, value, overBoundCU)
		}
		if !isTarget && value > maxSafeCU {
			t.Fatalf("non-target %s rejection config value = %d, exceeds max safe %d", field.name, value, maxSafeCU)
		}
	}
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
