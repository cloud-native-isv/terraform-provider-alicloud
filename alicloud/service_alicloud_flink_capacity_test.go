package alicloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type fakeFlinkCapacityAPI struct {
	workspace                *flink.Workspace
	namespaces               []flink.Namespace
	targets                  map[string][]flink.DeploymentTarget
	createdNamespaceTemplate flink.Namespace
	createPreWriteErr        error
	createPostReadErr        error
	deleteErr                error
	deleteDisappearAfterRead int
	pendingDelete            string
	deleteReads              int
	deleteReadCalls          int
	created                  []flink.Namespace
	deleted                  []string
	namespaceWrites          int
	queueWrites              int
	listNamespacesCalls      int
	listTargetCalls          map[string]int
}

func (a *fakeFlinkCapacityAPI) GetWorkspace(string) (*flink.Workspace, error) {
	return a.workspace, nil
}

func (a *fakeFlinkCapacityAPI) ListNamespaces(string) ([]flink.Namespace, error) {
	a.listNamespacesCalls++
	if a.pendingDelete != "" {
		a.deleteReadCalls++
		if a.deleteReads >= a.deleteDisappearAfterRead {
			a.removeNamespace(a.pendingDelete)
			a.pendingDelete = ""
		} else {
			a.deleteReads++
		}
	}
	return append([]flink.Namespace(nil), a.namespaces...), nil
}

func (a *fakeFlinkCapacityAPI) ListDeploymentTargets(_ string, namespace string) ([]flink.DeploymentTarget, error) {
	if a.listTargetCalls == nil {
		a.listTargetCalls = make(map[string]int)
	}
	a.listTargetCalls[namespace]++
	return append([]flink.DeploymentTarget(nil), a.targets[namespace]...), nil
}

func (a *fakeFlinkCapacityAPI) CreateNamespace(_ string, namespace *flink.Namespace) (*flink.Namespace, error) {
	if a.createPreWriteErr != nil {
		return nil, a.createPreWriteErr
	}
	request := *namespace
	a.created = append(a.created, request)
	created := a.createdNamespaceTemplate
	created.Name = request.Name
	created.Ha = request.Ha
	a.namespaces = append(a.namespaces, created)
	return nil, a.createPostReadErr
}

func (a *fakeFlinkCapacityAPI) DeleteNamespace(_ string, namespace string) error {
	a.deleted = append(a.deleted, namespace)
	if a.deleteErr != nil {
		a.pendingDelete = namespace
		return a.deleteErr
	}
	a.removeNamespace(namespace)
	return nil
}

func (a *fakeFlinkCapacityAPI) removeNamespace(namespace string) {
	for i := range a.namespaces {
		if a.namespaces[i].Name == namespace {
			a.namespaces = append(a.namespaces[:i], a.namespaces[i+1:]...)
			break
		}
	}
	delete(a.targets, namespace)
}

func (a *fakeFlinkCapacityAPI) GetNamespace(_ string, namespace string) (*flink.Namespace, error) {
	for i := range a.namespaces {
		if a.namespaces[i].Name == namespace {
			result := a.namespaces[i]
			return &result, nil
		}
	}
	return nil, flink.NewFlinkServiceErrorWithCode("", "", "404", "not found", "")
}

func (a *fakeFlinkCapacityAPI) UpdateNamespaceCapacity(_ string, namespace string, ha bool, fixed, elastic *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.namespaceWrites++
	for i := range a.namespaces {
		if a.namespaces[i].Name == namespace {
			a.namespaces[i].Ha = ha
			a.namespaces[i].GuaranteedResourceSpec = fixed
			a.namespaces[i].ElasticResourceSpec = elastic
			return flink.CapacityOperation{RequestID: "namespace-write"}, nil
		}
	}
	return flink.CapacityOperation{}, fmt.Errorf("namespace %q not found", namespace)
}

func (a *fakeFlinkCapacityAPI) UpdateDeploymentTargetV2(_ string, namespace string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	a.queueWrites++
	for i := range a.targets[namespace] {
		if a.targets[namespace][i].Name == target.Name {
			a.targets[namespace][i] = *target
			return target, nil
		}
	}
	return nil, fmt.Errorf("queue %q/%q not found", namespace, target.Name)
}

func (a *fakeFlinkCapacityAPI) ModifyPrepayWorkspaceCapacity(_ string, fixed, crossZone *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.workspace.ResourceSpec = fixed
	a.workspace.HaResourceSpec = crossZone
	return flink.CapacityOperation{RequestID: "workspace-write"}, nil
}

func (a *fakeFlinkCapacityAPI) ModifyPostpayWorkspaceCapacity(string, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected workspace write")
}

func (a *fakeFlinkCapacityAPI) EnableWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected workspace write")
}

func (a *fakeFlinkCapacityAPI) ModifyWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected workspace write")
}

func TestFlinkCapacityServiceApplyNamespaceTopologySteps(t *testing.T) {
	api := &fakeFlinkCapacityAPI{}
	service := &FlinkCapacityService{api: api}

	step := flinkcapacity.Step{
		Action:             flinkcapacity.CreateNamespace,
		Ref:                flinkcapacity.Ref{Namespace: "new-ha"},
		NamespaceCrossZone: true,
	}
	wantStep := step
	_, err := service.ApplyStep(context.Background(), "f-workspace", step)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.created) != 1 {
		t.Fatalf("CreateNamespace writes = %d, want 1", len(api.created))
	}
	request := api.created[0]
	if request.Id != "f-workspace" || request.Name != "new-ha" || !request.Ha || request.ResourceSpec == nil || request.ResourceSpec.Cpu != 1 || request.ResourceSpec.MemoryGB != 4 {
		t.Fatalf("CreateNamespace request = %#v, want workspace compatibility ID, name, Ha, and 1 CU / 4 GiB", request)
	}
	if step != wantStep {
		t.Fatalf("CreateNamespace mutated step: got %#v want %#v", step, wantStep)
	}

	_, err = service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
		Action: flinkcapacity.DeleteNamespace,
		Ref:    flinkcapacity.Ref{Namespace: "new-ha"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(api.deleted, ",") != "new-ha" {
		t.Fatalf("deleted namespaces = %#v", api.deleted)
	}
}

func TestFlinkCapacityServiceTypedPostCreateReadFailuresAreAmbiguous(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cause error
	}{
		{name: "404 cause", cause: flink.NewFlinkServiceErrorWithCode("", "", "404", "namespace not found", "")},
		{name: "non-404 cause", cause: flink.NewFlinkServiceErrorWithCode("read-request", "", "InternalError", "read unavailable", "")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			postReadErr := flink.NewFlinkPostCreateReadError("f-workspace", "new", tc.cause)
			api := &fakeFlinkCapacityAPI{createPostReadErr: postReadErr}
			service := &FlinkCapacityService{api: api}

			_, err := service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
				Action: flinkcapacity.CreateNamespace,
				Ref:    flinkcapacity.Ref{Namespace: "new"},
			})
			var ambiguous interface{ Ambiguous() bool }
			if !errors.As(err, &ambiguous) || !ambiguous.Ambiguous() {
				t.Fatalf("error = %T %v, want ambiguous typed post-create read failure", err, err)
			}
			var gotPostRead *flink.FlinkPostCreateReadError
			if !errors.As(err, &gotPostRead) || gotPostRead != postReadErr {
				t.Fatalf("error chain lost typed post-create read error: %T %v", err, err)
			}
			if !errors.Is(err, tc.cause) || errors.Unwrap(postReadErr) != tc.cause {
				t.Fatalf("error chain lost post-create read cause: %T %v", err, err)
			}
			if len(api.created) != 1 {
				t.Fatalf("CreateNamespace writes = %d, want exactly one", len(api.created))
			}
		})
	}
}

func TestFlinkCapacityServiceCreateNamespaceRPCFailuresAreNotAmbiguous(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{
			name: "RPC 404 with request ID",
			err:  flink.NewFlinkServiceErrorWithCode("rpc-request", "", "404", "namespace 'f-workspace/new' not found", ""),
		},
		{
			name: "untyped wrapper-shaped 404",
			err:  flink.NewFlinkServiceErrorWithCode("", "", "404", "namespace 'f-workspace/new' not found", ""),
		},
		{
			name: "RPC product rejection",
			err:  flink.NewFlinkServiceErrorWithCode("rpc-request", "", "903021", "namespace not found", ""),
		},
		{
			name: "non-wrapper 404",
			err:  flink.NewFlinkServiceErrorWithCode("", "", "404", "create namespace rejected", ""),
		},
		{
			name: "wrapper-shaped 404 for another namespace",
			err:  flink.NewFlinkServiceErrorWithCode("", "", "404", "namespace 'f-workspace/other' not found", ""),
		},
		{
			name: "Success false service error",
			err:  testFlinkCapacitySuccessFalseError(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeFlinkCapacityAPI{createPreWriteErr: tc.err}
			service := &FlinkCapacityService{api: api}

			_, err := service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
				Action: flinkcapacity.CreateNamespace,
				Ref:    flinkcapacity.Ref{Namespace: "new"},
			})
			var ambiguous interface{ Ambiguous() bool }
			if err != tc.err {
				t.Fatalf("error = %T %v, want original %T %v", err, err, tc.err, tc.err)
			}
			if errors.As(err, &ambiguous) && ambiguous.Ambiguous() {
				t.Fatalf("pre-write rejection was marked ambiguous: %T %v", err, err)
			}
			if len(api.created) != 0 {
				t.Fatalf("CreateNamespace writes = %d, want zero", len(api.created))
			}
		})
	}
}

func TestFlinkCapacityServiceDeleteNamespaceNotFoundUsesAuthoritativeReadConfirmation(t *testing.T) {
	for _, tc := range []struct {
		name               string
		disappearAfterRead int
	}{
		{name: "already disappeared", disappearAfterRead: 0},
		{name: "eventually disappears", disappearAfterRead: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api, desired := testFlinkCapacityDeleteReconcileFixture()
			api.deleteErr = flink.NewFlinkServiceErrorWithCode("delete-request", "", "404", "namespace not found", "")
			api.deleteDisappearAfterRead = tc.disappearAfterRead
			service := &FlinkCapacityService{api: api}
			reconciler := flinkcapacity.Reconciler{
				API:          service,
				PollInterval: time.Nanosecond,
				Sleep: func(context.Context, time.Duration) error {
					return nil
				},
			}

			got, err := reconciler.ReconcileAuthoritative(context.Background(), "f-workspace", desired)
			if err != nil {
				t.Fatal(err)
			}
			if steps, err := flinkcapacity.Plan(got, desired); err != nil || len(steps) != 0 {
				t.Fatalf("final tree did not converge: steps=%#v err=%v tree=%#v", steps, err, got)
			}
			if len(api.deleted) != 1 {
				t.Fatalf("DeleteNamespace calls = %d, want exactly one", len(api.deleted))
			}
			if api.deleteReadCalls != tc.disappearAfterRead+1 {
				t.Fatalf("post-delete reads = %d, want %d", api.deleteReadCalls, tc.disappearAfterRead+1)
			}
		})
	}
}

func TestFlinkCapacityServiceDeleteNamespaceOrdinaryFailureIsNotAmbiguous(t *testing.T) {
	wantErr := flink.NewFlinkServiceErrorWithCode("delete-request", "", "InvalidParameter", "invalid", "")
	api := &fakeFlinkCapacityAPI{deleteErr: wantErr}
	service := &FlinkCapacityService{api: api}

	_, err := service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
		Action: flinkcapacity.DeleteNamespace,
		Ref:    flinkcapacity.Ref{Namespace: "drop"},
	})
	var ambiguous interface{ Ambiguous() bool }
	if err != wantErr {
		t.Fatalf("error = %T %v, want original %T %v", err, err, wantErr, wantErr)
	}
	if errors.As(err, &ambiguous) && ambiguous.Ambiguous() {
		t.Fatalf("ordinary delete rejection was marked ambiguous: %T %v", err, err)
	}
}

func TestFlinkCapacityServiceReplansAfterInitialOneCUCreateReadback(t *testing.T) {
	for _, initialCPU := range []float64{0, 1, 2} {
		t.Run(fmt.Sprintf("initial CPU %v", initialCPU), func(t *testing.T) {
			api, desired := testFlinkCapacityNamespaceReconcileFixture(initialCPU)
			service := &FlinkCapacityService{api: api}
			reconciler := flinkcapacity.Reconciler{
				API:          service,
				PollInterval: time.Nanosecond,
				Sleep: func(context.Context, time.Duration) error {
					return nil
				},
			}

			got, err := reconciler.ReconcileAuthoritative(context.Background(), "f-workspace", desired)
			if err != nil {
				t.Fatal(err)
			}
			if steps, err := flinkcapacity.Plan(got, desired); err != nil || len(steps) != 0 {
				t.Fatalf("final tree did not converge: steps=%#v err=%v tree=%#v", steps, err, got)
			}
			if len(api.created) != 1 || api.created[0].ResourceSpec == nil || api.created[0].ResourceSpec.Cpu != 1 || api.created[0].ResourceSpec.MemoryGB != 4 {
				t.Fatalf("CreateNamespace calls = %#v, want one 1 CU / 4 GiB create", api.created)
			}
		})
	}
}

func TestFlinkCapacityServiceReadTreeFindsCustomQueueAfterLaterPagesBeforeWrites(t *testing.T) {
	api, desired := testFlinkCapacityPagedCustomQueueFixture()
	service := &FlinkCapacityService{api: api}

	tree, err := service.ReadTree(context.Background(), "f-workspace")
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Namespaces) != 101 || len(tree.Namespaces[100].Queues) != 101 || tree.Namespaces[100].Queues[100].Name != "custom-later-page" {
		t.Fatalf("ReadTree lost later-page namespace or queue: %#v", tree.Namespaces[100])
	}
	if api.listNamespacesCalls != 1 || api.listTargetCalls["namespace-100"] != 1 {
		t.Fatalf("service must consume one complete-list result per ReadTree: namespace calls=%d target calls=%d", api.listNamespacesCalls, api.listTargetCalls["namespace-100"])
	}
	firstHash, err := testFlinkCapacityCanonicalTopologyHash(tree)
	if err != nil {
		t.Fatal(err)
	}

	shuffledAPI, _ := testFlinkCapacityPagedCustomQueueFixture()
	reverseFlinkCapacityFixtureOrder(shuffledAPI)
	shuffledTree, err := (&FlinkCapacityService{api: shuffledAPI}).ReadTree(context.Background(), "f-workspace")
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := testFlinkCapacityCanonicalTopologyHash(shuffledTree)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash || !strings.HasPrefix(firstHash, "v1:") {
		t.Fatalf("canonical topology hash is not stable: first=%q second=%q", firstHash, secondHash)
	}

	_, err = (flinkcapacity.Reconciler{API: service}).ReconcileAuthoritative(context.Background(), "f-workspace", desired)
	if err == nil || !strings.Contains(err.Error(), "unsupported queue") {
		t.Fatalf("reconcile error = %v, want custom queue topology rejection", err)
	}
	if len(api.created) != 0 || len(api.deleted) != 0 || api.namespaceWrites != 0 || api.queueWrites != 0 {
		t.Fatalf("writes before later-page custom queue detection: create=%d delete=%d namespace=%d queue=%d", len(api.created), len(api.deleted), api.namespaceWrites, api.queueWrites)
	}
}

func testFlinkCapacityNamespaceReconcileFixture(initialCPU float64) (*fakeFlinkCapacityAPI, flinkcapacity.Tree) {
	workspaceCPU := 2.0
	if initialCPU > 1 {
		workspaceCPU += initialCPU - 1
	}
	api := &fakeFlinkCapacityAPI{
		workspace:  testFlinkCapacityWorkspace(workspaceCPU),
		namespaces: []flink.Namespace{testFlinkCapacityNamespace("default", false, 1)},
		targets: map[string][]flink.DeploymentTarget{
			"default": {testFlinkCapacityQueue("default-queue", 1)},
			"new":     {testFlinkCapacityQueue("default-queue", initialCPU)},
		},
		createdNamespaceTemplate: testFlinkCapacityNamespace("", false, initialCPU),
		createPostReadErr:        testFlinkCapacityPostCreateNotFound("f-workspace", "new"),
	}
	return api, testFlinkCapacityDesiredTree()
}

func testFlinkCapacityPostCreateNotFound(workspace, namespace string) error {
	cause := flink.NewFlinkServiceErrorWithCode("", "", "404", fmt.Sprintf("namespace '%s/%s' not found", workspace, namespace), "")
	return flink.NewFlinkPostCreateReadError(workspace, namespace, cause)
}

func testFlinkCapacitySuccessFalseError() error {
	httpCode := int32(200)
	success := false
	requestID := "create-request"
	return flink.NewFlinkServiceErrorFromResponse(&httpCode, &success, nil, nil, &requestID)
}

func testFlinkCapacityDeleteReconcileFixture() (*fakeFlinkCapacityAPI, flinkcapacity.Tree) {
	api := &fakeFlinkCapacityAPI{
		workspace: testFlinkCapacityWorkspace(2),
		namespaces: []flink.Namespace{
			testFlinkCapacityNamespace("keep", false, 1),
			testFlinkCapacityNamespace("drop", false, 1),
		},
		targets: map[string][]flink.DeploymentTarget{
			"keep": {testFlinkCapacityQueue("default-queue", 1)},
			"drop": {testFlinkCapacityQueue("default-queue", 1)},
		},
	}
	desired := flinkcapacity.Tree{
		ChargeType: "PRE",
		Workspace:  flinkcapacity.WorkspaceCapacity{FixedCU: 2, Limit: 2},
		Namespaces: []flinkcapacity.Namespace{{
			Name:     "keep",
			Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2},
			Queues:   []flinkcapacity.Queue{{Name: "default-queue", Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2}}},
		}},
	}
	return api, desired
}

func testFlinkCapacityPagedCustomQueueFixture() (*fakeFlinkCapacityAPI, flinkcapacity.Tree) {
	api := &fakeFlinkCapacityAPI{
		workspace: testFlinkCapacityWorkspace(101),
		targets:   make(map[string][]flink.DeploymentTarget, 101),
	}
	desired := flinkcapacity.Tree{ChargeType: "PRE", Workspace: flinkcapacity.WorkspaceCapacity{FixedCU: 202, Limit: 202}}
	for i := 0; i < 101; i++ {
		name := fmt.Sprintf("namespace-%03d", i)
		api.namespaces = append(api.namespaces, testFlinkCapacityNamespace(name, false, 1))
		api.targets[name] = []flink.DeploymentTarget{testFlinkCapacityQueue("default-queue", 1)}
		desired.Namespaces = append(desired.Namespaces, flinkcapacity.Namespace{
			Name:     name,
			Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2},
			Queues:   []flinkcapacity.Queue{{Name: "default-queue", Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2}}},
		})
	}
	queues := make([]flink.DeploymentTarget, 0, 101)
	queues = append(queues, testFlinkCapacityQueue("default-queue", 1))
	for i := 1; i < 100; i++ {
		queues = append(queues, testFlinkCapacityQueue(fmt.Sprintf("first-page-custom-%03d", i), 0))
	}
	queues = append(queues, testFlinkCapacityQueue("custom-later-page", 0))
	api.targets["namespace-100"] = queues
	return api, desired
}

type testFlinkCapacityHashNamespace struct {
	Name      string                       `json:"name"`
	CrossZone bool                         `json:"cross_zone"`
	Queues    []testFlinkCapacityHashQueue `json:"queues"`
}

type testFlinkCapacityHashQueue struct {
	Name    string           `json:"name"`
	Request flinkcapacity.CU `json:"request_half_cu"`
	Limit   flinkcapacity.CU `json:"limit_half_cu"`
}

// testFlinkCapacityCanonicalTopologyHash is an independent test oracle for
// the allocation resource's frozen v1 implicit_topology_hash contract.
func testFlinkCapacityCanonicalTopologyHash(tree flinkcapacity.Tree) (string, error) {
	namespaces := append([]flinkcapacity.Namespace(nil), tree.Namespaces...)
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Name < namespaces[j].Name })
	canonical := make([]testFlinkCapacityHashNamespace, 0, len(namespaces))
	for _, namespace := range namespaces {
		entry := testFlinkCapacityHashNamespace{Name: namespace.Name, CrossZone: namespace.CrossZone}
		queues := append([]flinkcapacity.Queue(nil), namespace.Queues...)
		sort.Slice(queues, func(i, j int) bool { return queues[i].Name < queues[j].Name })
		for _, queue := range queues {
			if queue.Capacity == nil {
				return "", fmt.Errorf("queue %q/%q capacity is missing", namespace.Name, queue.Name)
			}
			entry.Queues = append(entry.Queues, testFlinkCapacityHashQueue{Name: queue.Name, Request: queue.Capacity.Fixed, Limit: queue.Capacity.Limit})
		}
		canonical = append(canonical, entry)
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "v1:" + hex.EncodeToString(digest[:]), nil
}

func reverseFlinkCapacityFixtureOrder(api *fakeFlinkCapacityAPI) {
	for left, right := 0, len(api.namespaces)-1; left < right; left, right = left+1, right-1 {
		api.namespaces[left], api.namespaces[right] = api.namespaces[right], api.namespaces[left]
	}
	for namespace := range api.targets {
		queues := api.targets[namespace]
		for left, right := 0, len(queues)-1; left < right; left, right = left+1, right-1 {
			queues[left], queues[right] = queues[right], queues[left]
		}
		api.targets[namespace] = queues
	}
}

func testFlinkCapacityWorkspace(cpu float64) *flink.Workspace {
	return &flink.Workspace{
		Id:           "f-workspace",
		Status:       "RUNNING",
		OrderState:   "NORMAL",
		ResourceId:   "resource-workspace",
		ChargeType:   "PRE",
		ResourceSpec: &flink.ResourceSpec{Cpu: cpu, MemoryGB: cpu * 4},
	}
}

func testFlinkCapacityNamespace(name string, ha bool, cpu float64) flink.Namespace {
	return flink.Namespace{
		Name:                   name,
		Ha:                     ha,
		Status:                 "SUCCESS",
		GuaranteedResourceSpec: &flink.ResourceSpec{Cpu: cpu, MemoryGB: cpu * 4},
		ElasticResourceSpec:    &flink.ResourceSpec{},
	}
}

func testFlinkCapacityQueue(name string, cpu float64) flink.DeploymentTarget {
	return flink.DeploymentTarget{
		Name: name,
		Quota: &flink.ResourceQuota{
			Request: &flink.ResourceSpec{Cpu: cpu, MemoryGB: cpu * 4},
			Limit:   &flink.ResourceSpec{Cpu: cpu, MemoryGB: cpu * 4},
		},
	}
}

func testFlinkCapacityDesiredTree() flinkcapacity.Tree {
	desired := flinkcapacity.Tree{ChargeType: "PRE", Workspace: flinkcapacity.WorkspaceCapacity{FixedCU: 4, Limit: 4}}
	for _, name := range []string{"default", "new"} {
		desired.Namespaces = append(desired.Namespaces, flinkcapacity.Namespace{
			Name:     name,
			Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2},
			Queues:   []flinkcapacity.Queue{{Name: "default-queue", Capacity: &flinkcapacity.Capacity{Fixed: 2, Limit: 2}}},
		})
	}
	return desired
}

func TestValidateFlinkCapacityWorkspaceReady(t *testing.T) {
	ready := &flink.Workspace{
		Id:         "f-test",
		Status:     "RUNNING",
		OrderState: "NORMAL",
		ResourceId: "sc-test",
	}
	if err := validateFlinkCapacityWorkspaceReady(ready); err != nil {
		t.Fatal(err)
	}
	readyElastic := *ready
	readyElastic.Elastic = true
	readyElastic.ElasticInstanceId = "f-elastic"
	readyElastic.ElasticOrderState = "NORMAL"
	readyElastic.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
	if err := validateFlinkCapacityWorkspaceReady(&readyElastic); err != nil {
		t.Fatalf("ready elastic workspace: %v", err)
	}

	for name, mutate := range map[string]func(*flink.Workspace){
		"creating":            func(workspace *flink.Workspace) { workspace.Status = "CREATING" },
		"order pending":       func(workspace *flink.Workspace) { workspace.OrderState = "PROCESSING" },
		"resource id missing": func(workspace *flink.Workspace) { workspace.ResourceId = "" },
		"elastic id missing": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
		},
		"elastic spec missing": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticInstanceId = "f-elastic"
		},
		"elastic order pending": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticInstanceId = "f-elastic"
			workspace.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
			workspace.ElasticOrderState = "PROCESSING"
		},
		"elastic order missing": func(workspace *flink.Workspace) {
			workspace.Elastic = true
			workspace.ElasticInstanceId = "f-elastic"
			workspace.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
		},
	} {
		t.Run(name, func(t *testing.T) {
			workspace := *ready
			mutate(&workspace)
			err := validateFlinkCapacityWorkspaceReady(&workspace)
			var retryable interface{ Retryable() bool }
			if err == nil || !errors.As(err, &retryable) || !retryable.Retryable() {
				t.Fatalf("error = %v, want retryable", err)
			}
		})
	}
}

func TestValidateFlinkCapacityWorkspaceRejectsTerminalElasticOrder(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		Status:              "RUNNING",
		OrderState:          "NORMAL",
		ResourceId:          "sc-test",
		Elastic:             true,
		ElasticInstanceId:   "f-elastic",
		ElasticOrderState:   "FAILED",
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
	}
	err := validateFlinkCapacityWorkspaceReady(workspace)
	if err == nil || !strings.Contains(err.Error(), "elastic order") {
		t.Fatalf("error = %v", err)
	}
	var retryable interface{ Retryable() bool }
	if errors.As(err, &retryable) && retryable.Retryable() {
		t.Fatalf("terminal elastic order must not be retryable: %v", err)
	}
}

func TestValidateFlinkCapacityWorkspaceRejectsContradictoryElasticState(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		Status:              "RUNNING",
		OrderState:          "NORMAL",
		ResourceId:          "sc-test",
		Elastic:             false,
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
	}
	err := validateFlinkCapacityWorkspaceReady(workspace)
	if err == nil || !strings.Contains(err.Error(), "Elastic=false") {
		t.Fatalf("error = %v", err)
	}
	var retryable interface{ Retryable() bool }
	if errors.As(err, &retryable) && retryable.Retryable() {
		t.Fatalf("contradictory state must be terminal: %v", err)
	}
}

func TestValidateFlinkCapacityNamespaceReady(t *testing.T) {
	for _, status := range []string{"SUCCESS", "Available"} {
		if err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: status}); err != nil {
			t.Fatalf("status %q: %v", status, err)
		}
	}
	for _, status := range []string{"CREATING", "MODIFYING"} {
		err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: status})
		var retryable interface{ Retryable() bool }
		if err == nil || !errors.As(err, &retryable) || !retryable.Retryable() {
			t.Fatalf("status %q error = %v, want retryable", status, err)
		}
	}
	if err := validateFlinkCapacityNamespaceReady(flink.Namespace{Name: "default", Status: "FAILED"}); err == nil {
		t.Fatal("FAILED namespace status was accepted")
	}
}

func TestBuildFlinkCapacityTreeFromObjects(t *testing.T) {
	workspace := &flink.Workspace{
		Id:                  "f-test",
		ResourceId:          "sc-test",
		ChargeType:          "PRE",
		Ha:                  true,
		ResourceSpec:        &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		HaResourceSpec:      &flink.ResourceSpec{Cpu: 3, MemoryGB: 12},
		ElasticResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		ClusterUsedResources: &flink.WorkspaceUsedResources{
			UsedResource: 1.25,
		},
	}
	namespaces := []flink.Namespace{{
		Name:                   "default",
		Ha:                     true,
		GuaranteedResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		ElasticResourceSpec:    &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		ResourceUsed:           &flink.ResourceUsed{Cu: 0.1},
	}}
	targets := map[string][]flink.DeploymentTarget{
		"default": {{
			Name: "q",
			Quota: &flink.ResourceQuota{
				Request: &flink.ResourceSpec{Cpu: 1.5, MemoryGB: 6},
				Limit:   &flink.ResourceSpec{Cpu: 3, MemoryGB: 12},
				Used:    &flink.ResourceSpec{Cpu: 0.25, MemoryGB: 99},
			},
		}},
	}

	got, err := buildFlinkCapacityTree(workspace, namespaces, targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.Workspace != (flinkcapacity.WorkspaceCapacity{HA: true, FixedCU: 4, CrossZoneFixedCU: 6, Limit: 18, Used: 1.25}) {
		t.Fatalf("workspace capacity = %#v", got.Workspace)
	}
	if got.WorkspaceResourceID != "sc-test" {
		t.Fatalf("workspace ResourceId = %q", got.WorkspaceResourceID)
	}
	if got.Namespaces[0].Capacity == nil || *got.Namespaces[0].Capacity != (flinkcapacity.Capacity{Fixed: 8, Limit: 12}) || got.Namespaces[0].Used != 0.1 {
		t.Fatalf("namespace = %#v", got.Namespaces[0])
	}
	if !got.Namespaces[0].CrossZone {
		t.Fatalf("namespace CrossZone = false, want true from cws-lib Ha")
	}
	queue := got.Namespaces[0].Queues[0]
	if queue.Capacity == nil || *queue.Capacity != (flinkcapacity.Capacity{Fixed: 3, Limit: 6}) || queue.Used != 0.25 {
		t.Fatalf("queue = %#v", queue)
	}
}

func TestBuildFlinkCapacityTreePostpaid(t *testing.T) {
	workspace := &flink.Workspace{ChargeType: "POST", ResourceSpec: &flink.ResourceSpec{Cpu: 8, MemoryGB: 32}}
	namespaces := []flink.Namespace{{Name: "default", GuaranteedResourceSpec: &flink.ResourceSpec{}, ElasticResourceSpec: &flink.ResourceSpec{Cpu: 8, MemoryGB: 32}}}
	targets := map[string][]flink.DeploymentTarget{"default": {{Name: "q", Quota: &flink.ResourceQuota{Request: &flink.ResourceSpec{}, Limit: &flink.ResourceSpec{Cpu: 8, MemoryGB: 32}, Used: &flink.ResourceSpec{}}}}}

	got, err := buildFlinkCapacityTree(workspace, namespaces, targets)
	if err != nil {
		t.Fatal(err)
	}
	if got.Workspace.TotalFixed() != 0 || got.Workspace.Limit != 16 {
		t.Fatalf("workspace capacity = %#v", got.Workspace)
	}
	if got.Namespaces[0].Capacity.Fixed != 0 || got.Namespaces[0].Queues[0].Capacity.Fixed != 0 {
		t.Fatalf("POST child capacities must be pure elastic: %#v", got.Namespaces[0])
	}
}

func TestFlinkResourceSpecForCU(t *testing.T) {
	got := flinkResourceSpecForCU(3)
	if got.Cpu != 1.5 || got.MemoryGB != 6 {
		t.Fatalf("resource spec = %#v", got)
	}
}

func TestCUFromResourceSpecRejectsInconsistentMemory(t *testing.T) {
	_, err := cuFromResourceSpec(&flink.ResourceSpec{Cpu: 2, MemoryGB: 16})
	if err == nil || !strings.Contains(err.Error(), "memory") {
		t.Fatalf("error = %v", err)
	}
}

func TestClassifyFlinkCapacityWriteError(t *testing.T) {
	sdkErr := flink.NewFlinkSDKError("foasconsole", "ModifyPrepayInstanceSpec", "transport failed", errors.New("connection reset"))
	classified := classifyFlinkCapacityWriteError(sdkErr)
	var ambiguous interface{ Ambiguous() bool }
	if !errors.As(classified, &ambiguous) || !ambiguous.Ambiguous() {
		t.Fatalf("SDK write error was not marked ambiguous: %T %v", classified, classified)
	}

	serviceErr := flink.NewFlinkServiceErrorWithCode("request", "", "InvalidParameter", "invalid", "")
	classified = classifyFlinkCapacityWriteError(serviceErr)
	if errors.As(classified, &ambiguous) && ambiguous.Ambiguous() {
		t.Fatalf("service rejection was marked ambiguous: %T %v", classified, classified)
	}
}
