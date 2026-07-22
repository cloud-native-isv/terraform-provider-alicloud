package alicloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type fakeFlinkCapacityAPI struct {
	calls                    []string
	workspace                *flink.Workspace
	getWorkspaceErr          error
	getWorkspaceResponses    []testFlinkCapacityWorkspaceResponse
	listedWorkspaces         []flink.Workspace
	listWorkspacesErr        error
	namespaces               []flink.Namespace
	listNamespacesErr        error
	targets                  map[string][]flink.DeploymentTarget
	listTargetErrors         map[string]error
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
	workspaceFixedWrites     []testFlinkCapacityFixedWrite
	workspacePostpaidWrites  int
	workspaceElasticWrites   []testFlinkCapacityElasticWrite
	workspaceReads           int
	listWorkspacesCalls      int
	namespaceReads           int
	getWorkspaceHook         func()
	getNamespaceHook         func()
}

type testFlinkCapacityFixedWrite struct {
	fixed     *flink.ResourceSpec
	crossZone *flink.ResourceSpec
}

type testFlinkCapacityElasticWrite struct {
	action  flinkcapacity.Action
	elastic *flink.ResourceSpec
}

type testFlinkCapacityWorkspaceResponse struct {
	workspace *flink.Workspace
	err       error
}

func (a *fakeFlinkCapacityAPI) GetWorkspace(string) (*flink.Workspace, error) {
	a.calls = append(a.calls, "get-workspace")
	a.workspaceReads++
	if a.getWorkspaceHook != nil {
		a.getWorkspaceHook()
	}
	if len(a.getWorkspaceResponses) > 0 {
		response := a.getWorkspaceResponses[0]
		a.getWorkspaceResponses = a.getWorkspaceResponses[1:]
		return response.workspace, response.err
	}
	if a.getWorkspaceErr != nil {
		return nil, a.getWorkspaceErr
	}
	return a.workspace, nil
}

func (a *fakeFlinkCapacityAPI) ListWorkspaces() ([]flink.Workspace, error) {
	a.listWorkspacesCalls++
	if a.listWorkspacesErr != nil {
		return nil, a.listWorkspacesErr
	}
	return append([]flink.Workspace(nil), a.listedWorkspaces...), nil
}

func (a *fakeFlinkCapacityAPI) ListNamespaces(string) ([]flink.Namespace, error) {
	a.listNamespacesCalls++
	if a.listNamespacesErr != nil {
		return nil, a.listNamespacesErr
	}
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
	if err := a.listTargetErrors[namespace]; err != nil {
		return nil, err
	}
	return append([]flink.DeploymentTarget(nil), a.targets[namespace]...), nil
}

func (a *fakeFlinkCapacityAPI) CreateNamespace(_ string, namespace *flink.Namespace) (*flink.Namespace, error) {
	if a.createPreWriteErr != nil {
		return nil, a.createPreWriteErr
	}
	request := *namespace
	a.calls = append(a.calls, "create-namespace")
	a.created = append(a.created, request)
	created := a.createdNamespaceTemplate
	created.Name = request.Name
	created.Ha = request.Ha
	a.namespaces = append(a.namespaces, created)
	return nil, a.createPostReadErr
}

func (a *fakeFlinkCapacityAPI) DeleteNamespace(_ string, namespace string) error {
	a.calls = append(a.calls, "delete-namespace")
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
	a.namespaceReads++
	if a.getNamespaceHook != nil {
		a.getNamespaceHook()
	}
	for i := range a.namespaces {
		if a.namespaces[i].Name == namespace {
			result := a.namespaces[i]
			return &result, nil
		}
	}
	return nil, flink.NewFlinkServiceErrorWithCode("", "", "404", "not found", "")
}

func (a *fakeFlinkCapacityAPI) UpdateNamespaceCapacity(_ string, namespace string, ha bool, fixed, elastic *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.calls = append(a.calls, "modify-namespace")
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
	a.calls = append(a.calls, "modify-queue")
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
	a.calls = append(a.calls, "modify-workspace-fixed")
	a.workspaceFixedWrites = append(a.workspaceFixedWrites, testFlinkCapacityFixedWrite{fixed: fixed, crossZone: crossZone})
	a.workspace.ResourceSpec = fixed
	a.workspace.HaResourceSpec = crossZone
	return flink.CapacityOperation{RequestID: "workspace-write"}, nil
}

func (a *fakeFlinkCapacityAPI) ModifyPostpayWorkspaceCapacity(string, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.calls = append(a.calls, "modify-workspace-postpaid")
	a.workspacePostpaidWrites++
	return flink.CapacityOperation{}, fmt.Errorf("unexpected workspace write")
}

func (a *fakeFlinkCapacityAPI) EnableWorkspaceElastic(_ string, elastic *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.calls = append(a.calls, "enable-workspace-elastic")
	a.workspaceElasticWrites = append(a.workspaceElasticWrites, testFlinkCapacityElasticWrite{action: flinkcapacity.EnableWorkspaceElastic, elastic: elastic})
	a.workspace.ElasticResourceSpec = elastic
	return flink.CapacityOperation{RequestID: "workspace-elastic-write"}, nil
}

func (a *fakeFlinkCapacityAPI) ModifyWorkspaceElastic(_ string, elastic *flink.ResourceSpec) (flink.CapacityOperation, error) {
	a.calls = append(a.calls, "modify-workspace-elastic")
	a.workspaceElasticWrites = append(a.workspaceElasticWrites, testFlinkCapacityElasticWrite{action: flinkcapacity.ModifyWorkspaceElastic, elastic: elastic})
	a.workspace.ElasticResourceSpec = elastic
	return flink.CapacityOperation{RequestID: "workspace-elastic-write"}, nil
}

func TestFlinkCapacityServiceReadTreeRecoversTyped404FromExactWorkspaceList(t *testing.T) {
	workspace := *testFlinkCapacityWorkspace(2)
	for _, tc := range []struct {
		name   string
		getErr error
	}{
		{
			name:   "direct typed 404",
			getErr: flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", ""),
		},
		{
			name: "wrapped typed 404",
			getErr: fmt.Errorf("GetWorkspace failed: %w",
				flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = tc.getErr
			api.listedWorkspaces = []flink.Workspace{{Id: "f-decoy"}, workspace}

			tree, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), workspace.Id)
			if err != nil {
				t.Fatal(err)
			}
			if tree.WorkspaceResourceID != workspace.ResourceId {
				t.Fatalf("ReadTree() WorkspaceResourceID = %q, want %q", tree.WorkspaceResourceID, workspace.ResourceId)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 1 || api.listTargetCalls["default"] != 1 {
				t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d targets=%d, want 1/1/1/1", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, api.listTargetCalls["default"])
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("ReadTree() made %d capacity writes before Workspace identity corroboration", writes)
			}
		})
	}
}

func TestFlinkCapacityServiceReadTreeIsStatelessAcrossSuccessfulAndTyped404Reads(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	var firstServiceTree flinkcapacity.Tree

	for iteration, name := range []string{"same service", "fresh service"} {
		t.Run(name, func(t *testing.T) {
			get404 := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceResponses = []testFlinkCapacityWorkspaceResponse{
				{workspace: workspace},
				{err: get404},
			}
			api.listedWorkspaces = []flink.Workspace{*workspace}
			service := &FlinkCapacityService{api: api}

			exactTree, err := service.ReadTree(context.Background(), workspace.Id)
			if err != nil {
				t.Fatalf("exact Get ReadTree() error = %v", err)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 {
				t.Fatalf("exact Get calls: get=%d list=%d, want 1/0", api.workspaceReads, api.listWorkspacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("exact Get ReadTree() made %d capacity writes", writes)
			}

			recoveredTree, err := service.ReadTree(context.Background(), workspace.Id)
			if err != nil {
				t.Fatalf("typed-404 ReadTree() error = %v, want exact-list recovery after prior success", err)
			}
			if api.workspaceReads != 2 || api.listWorkspacesCalls != 1 {
				t.Fatalf("typed-404 calls: get=%d list=%d, want 2/1", api.workspaceReads, api.listWorkspacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("typed-404 ReadTree() made %d capacity writes before Workspace identity corroboration", writes)
			}
			if !reflect.DeepEqual(recoveredTree, exactTree) {
				t.Fatalf("typed-404 recovered tree = %#v, want prior exact-Get tree %#v", recoveredTree, exactTree)
			}
			if iteration == 0 {
				firstServiceTree = recoveredTree
			} else if !reflect.DeepEqual(recoveredTree, firstServiceTree) {
				t.Fatalf("fresh service recovered tree = %#v, want first service tree %#v", recoveredTree, firstServiceTree)
			}
		})
	}
}

func TestFlinkCapacityServiceReadTreeAppliesReadinessValidationToExactListedWorkspace(t *testing.T) {
	workspace := *testFlinkCapacityWorkspace(2)
	workspace.Status = "CREATING"
	getErr := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
	api := testFlinkCapacityReadableAPI(nil)
	api.getWorkspaceErr = getErr
	api.listedWorkspaces = []flink.Workspace{workspace}

	_, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), workspace.Id)
	var notReady *flinkcapacity.NotReadyError
	if !errors.As(err, &notReady) || err == getErr {
		t.Fatalf("ReadTree() error = %T %v, want readiness error for corroborated Workspace", err, err)
	}
	if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 {
		t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d, want 1/1/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
	}
	if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
		t.Fatalf("ReadTree() made %d capacity writes while Workspace was not ready", writes)
	}
}

func TestFlinkCapacityServiceReadTreeTyped404RequiresExactListedWorkspaceID(t *testing.T) {
	for _, tc := range []struct {
		name       string
		workspaces []flink.Workspace
		wrapGetErr bool
	}{
		{name: "empty list"},
		{
			name: "similar name and resource ID",
			workspaces: []flink.Workspace{{
				Id: "f-workspace-copy", Name: "f-workspace", ResourceId: "f-workspace",
			}},
		},
		{
			name:       "different Workspace ID",
			workspaces: []flink.Workspace{{Id: "f-other"}},
		},
		{
			name: "wrapped typed 404 with decoy",
			workspaces: []flink.Workspace{{
				Id: "f-decoy", Name: "f-workspace", ResourceId: "f-workspace",
			}},
			wrapGetErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var getErr error = flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
			if tc.wrapGetErr {
				getErr = fmt.Errorf("GetWorkspace failed: %w", getErr)
			}
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = getErr
			api.listedWorkspaces = tc.workspaces

			_, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), "f-workspace")
			if err != getErr {
				t.Fatalf("ReadTree() error = %T %v, want original %T %v", err, err, getErr, getErr)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 {
				t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d, want 1/1/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("ReadTree() made %d capacity writes without exact Workspace identity", writes)
			}
		})
	}
}

func TestFlinkCapacityServiceBootstrapIdentityAbsenceRetryIsExactAndBounded(t *testing.T) {
	getErr := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
	api := testFlinkCapacityReadableAPI(nil)
	api.getWorkspaceErr = getErr
	contextValue := testFlinkCapacityBootstrapContext("f-workspace", "resource-workspace", 1_700_000_900)
	now := time.Unix(1_700_000_000, 0)
	service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time { return now })

	_, err := service.ReadTree(context.Background(), "f-workspace")
	var notReady *flinkcapacity.NotReadyError
	if !errors.As(err, &notReady) {
		t.Fatalf("pre-deadline typed 404 + exact-list absence error = %T %v, want NotReadyError", err, err)
	}
	if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("pre-deadline absence calls get/list/namespaces/writes = %d/%d/%d/%d, want 1/1/0/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, testFlinkCapacityWriteCalls(api))
	}

	now = time.Unix(contextValue.IdentityAbsenceRetryNotAfterUnix, 0)
	_, err = service.ReadTree(context.Background(), "f-workspace")
	if err != getErr {
		t.Fatalf("deadline typed 404 + exact-list absence error = %T %v, want original authoritative %T %v", err, err, getErr, getErr)
	}
	if testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("deadline absence made %d writes", testFlinkCapacityWriteCalls(api))
	}

	strictAPI := testFlinkCapacityReadableAPI(nil)
	strictAPI.getWorkspaceErr = getErr
	strict := (&FlinkCapacityService{api: strictAPI}).withCapacityBootstrapContext(contextValue, false, func() time.Time { return time.Unix(1_700_000_000, 0) })
	_, err = strict.ReadTree(context.Background(), "f-workspace")
	if err != getErr {
		t.Fatalf("strict callback error = %T %v, want original %T %v", err, err, getErr, getErr)
	}
}

func TestFlinkCapacityServiceBootstrapIdentitySeenCannotReturnToAbsenceRetry(t *testing.T) {
	getErr := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
	visible := *testFlinkCapacityWorkspace(2)
	visible.Tags = nil
	api := testFlinkCapacityReadableAPI(nil)
	api.getWorkspaceErr = getErr
	api.listedWorkspaces = []flink.Workspace{visible}
	contextValue := testFlinkCapacityBootstrapContext(visible.Id, visible.ResourceId, 1_700_000_900)
	service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time { return time.Unix(1_700_000_000, 0) })

	_, err := service.ReadTree(context.Background(), visible.Id)
	var notReady *flinkcapacity.NotReadyError
	if !errors.As(err, &notReady) || !strings.Contains(err.Error(), flinkworkspace.CreateTokenTagKey) {
		t.Fatalf("first exact identity with unpropagated tags error = %T %v, want tag NotReadyError", err, err)
	}
	if !service.capacityBootstrap.identitySeen {
		t.Fatal("exact listed identity did not consume callback-local absence permission")
	}

	api.listedWorkspaces = nil
	_, err = service.ReadTree(context.Background(), visible.Id)
	if err != getErr {
		t.Fatalf("post-identity absence error = %T %v, want original authoritative %T %v", err, err, getErr, getErr)
	}
	if testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("identity propagation sequence made %d writes", testFlinkCapacityWriteCalls(api))
	}
}

func TestFlinkCapacityServiceBootstrapPinsFirstObservedResourceIDWithinCallback(t *testing.T) {
	first := *testFlinkCapacityWorkspace(2)
	first.Tags = nil
	second := first
	second.ResourceId = "resource-replacement"
	contextValue := testFlinkCapacityBootstrapContext(first.Id, "", 1_700_000_900)
	second.Tags = []flink.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: contextValue.TerraformCreateToken},
		{Key: flinkworkspace.CreateIntentTagKey, Value: contextValue.CreateIntentFingerprint},
	}
	api := testFlinkCapacityReadableAPI(nil)
	api.getWorkspaceResponses = []testFlinkCapacityWorkspaceResponse{{workspace: &first}, {workspace: &second}}
	service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time { return time.Unix(1_700_000_000, 0) })

	_, err := service.ReadTree(context.Background(), first.Id)
	var notReady *flinkcapacity.NotReadyError
	if !errors.As(err, &notReady) || service.capacityBootstrap.pinnedResourceID != first.ResourceId {
		t.Fatalf("first observation error/pin = %T %v/%q, want NotReady/%q", err, err, service.capacityBootstrap.pinnedResourceID, first.ResourceId)
	}
	_, err = service.ReadTree(context.Background(), first.Id)
	if err == nil || !strings.Contains(err.Error(), "ResourceId changed during capacity bootstrap") {
		t.Fatalf("second observation error = %T %v, want fatal ResourceId change", err, err)
	}
	if testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("ResourceId change made %d capacity writes", testFlinkCapacityWriteCalls(api))
	}
}

func TestFlinkCapacityServiceBootstrapProvenanceMustMatchBeforeTreeReads(t *testing.T) {
	base := testFlinkCapacityWorkspace(2)
	contextValue := testFlinkCapacityBootstrapContext(base.Id, base.ResourceId, 1_700_000_900)
	ready := *base
	ready.Tags = []flink.Tag{
		{Key: flinkworkspace.CreateTokenTagKey, Value: contextValue.TerraformCreateToken},
		{Key: flinkworkspace.CreateIntentTagKey, Value: contextValue.CreateIntentFingerprint},
	}

	for _, test := range []struct {
		name         string
		mutate       func(*flink.Workspace)
		wantNotReady bool
		wantText     string
	}{
		{name: "resource ID not propagated", mutate: func(value *flink.Workspace) { value.ResourceId = "" }, wantNotReady: true, wantText: "ResourceId"},
		{name: "create token not propagated", mutate: func(value *flink.Workspace) { value.Tags = value.Tags[1:] }, wantNotReady: true, wantText: flinkworkspace.CreateTokenTagKey},
		{name: "intent fingerprint not propagated", mutate: func(value *flink.Workspace) { value.Tags = value.Tags[:1] }, wantNotReady: true, wantText: flinkworkspace.CreateIntentTagKey},
		{name: "resource ID mismatch", mutate: func(value *flink.Workspace) { value.ResourceId = "resource-other" }, wantText: "ResourceId mismatch"},
		{name: "create token mismatch", mutate: func(value *flink.Workspace) { value.Tags[0].Value = strings.Repeat("c", 64) }, wantText: "create token mismatch"},
		{name: "intent fingerprint mismatch", mutate: func(value *flink.Workspace) { value.Tags[1].Value = strings.Repeat("c", 64) }, wantText: "intent fingerprint mismatch"},
		{name: "duplicate token", mutate: func(value *flink.Workspace) { value.Tags = append(value.Tags, value.Tags[0]) }, wantText: "duplicate"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := ready
			workspace.Tags = append([]flink.Tag(nil), ready.Tags...)
			test.mutate(&workspace)
			api := testFlinkCapacityReadableAPI(&workspace)
			service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time { return time.Unix(1_700_000_000, 0) })
			_, err := service.ReadTree(context.Background(), ready.Id)
			if err == nil || !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("ReadTree() error = %T %v, want text %q", err, err, test.wantText)
			}
			var gotNotReady *flinkcapacity.NotReadyError
			if errors.As(err, &gotNotReady) != test.wantNotReady {
				t.Fatalf("ReadTree() error retryability = %T %v, want NotReady=%t", err, err, test.wantNotReady)
			}
			if api.listNamespacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 0 {
				t.Fatalf("provenance failure made namespace reads/writes = %d/%d, want 0/0", api.listNamespacesCalls, testFlinkCapacityWriteCalls(api))
			}
		})
	}

	api := testFlinkCapacityReadableAPI(&ready)
	service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time { return time.Unix(1_700_000_000, 0) })
	if _, err := service.ReadTree(context.Background(), ready.Id); err != nil {
		t.Fatalf("matching bootstrap provenance ReadTree() error = %v", err)
	}
	if api.listNamespacesCalls != 1 || api.listTargetCalls["default"] != 1 {
		t.Fatalf("matching provenance tree reads namespaces/targets = %d/%d, want 1/1", api.listNamespacesCalls, api.listTargetCalls["default"])
	}
}

func TestFlinkCapacityServiceReadTreeDoesNotListForNonTyped404(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{
			name: "typed 403 containing not found",
			err:  flink.NewFlinkServiceErrorWithCode("get-request", "", "403", "workspace not found", ""),
		},
		{
			name: "typed non-exact 404 code containing not found",
			err:  flink.NewFlinkServiceErrorWithCode("get-request", "", "4040", "workspace not found", ""),
		},
		{
			name: "business error containing not available",
			err:  errors.New("workspace is not available for this account"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(nil)
			api.getWorkspaceErr = tc.err
			bootstrap := testFlinkCapacityBootstrapContext("f-workspace", "resource-workspace", 1_700_000_900)
			service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(bootstrap, true, func() time.Time { return time.Unix(1_700_000_000, 0) })

			_, err := service.ReadTree(context.Background(), "f-workspace")
			if err != tc.err {
				t.Fatalf("ReadTree() error = %T %v, want original %T %v", err, err, tc.err, tc.err)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 0 {
				t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d, want 1/0/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("ReadTree() made %d capacity writes after non-404 Get error", writes)
			}
		})
	}
}

func TestFlinkCapacityServiceReadTreeReturnsListErrorAfterTyped404(t *testing.T) {
	for _, listErr := range []error{
		flink.NewFlinkServiceErrorWithCode("list-request", "", "403", "permission denied", ""),
		errors.New("workspace listing is not available"),
	} {
		getErr := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
		api := testFlinkCapacityReadableAPI(nil)
		api.getWorkspaceErr = getErr
		api.listWorkspacesErr = listErr

		_, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), "f-workspace")
		if err != listErr {
			t.Fatalf("ReadTree() error = %T %v, want list error %T %v", err, err, listErr, listErr)
		}
		if api.workspaceReads != 1 || api.listWorkspacesCalls != 1 || api.listNamespacesCalls != 0 {
			t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d, want 1/1/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls)
		}
		if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
			t.Fatalf("ReadTree() made %d capacity writes after list failure", writes)
		}
	}
}

func TestFlinkCapacityServiceReadTreeExactGetDoesNotListWorkspaces(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	api := testFlinkCapacityReadableAPI(workspace)
	api.listWorkspacesErr = errors.New("ListWorkspaces must not be called")

	if _, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), workspace.Id); err != nil {
		t.Fatal(err)
	}
	if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 1 || api.listTargetCalls["default"] != 1 {
		t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d targets=%d, want 1/0/1/1", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, api.listTargetCalls["default"])
	}
	if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
		t.Fatalf("ReadTree() made %d capacity writes on exact Get path", writes)
	}
}

func TestFlinkCapacityServiceReadTreeRejectsSuccessfulGetWithoutExactWorkspaceIdentity(t *testing.T) {
	requestedID := "f-workspace"
	emptyID := testFlinkCapacityWorkspace(2)
	emptyID.Id = ""
	wrongID := testFlinkCapacityWorkspace(2)
	wrongID.Id = "f-workspace-copy"
	wrongID.Name = requestedID
	wrongID.ResourceId = requestedID

	for _, tc := range []struct {
		name             string
		workspace        *flink.Workspace
		wantObservedText string
	}{
		{name: "nil Workspace", workspace: nil, wantObservedText: "<nil>"},
		{name: "empty Workspace ID", workspace: emptyID, wantObservedText: `observed Workspace Id ""`},
		{name: "different Workspace ID despite similar name and matching ResourceId", workspace: wrongID, wantObservedText: `observed Workspace Id "f-workspace-copy"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(tc.workspace)

			_, err := (&FlinkCapacityService{api: api}).ReadTree(context.Background(), requestedID)
			if err == nil || !strings.Contains(err.Error(), `requested InstanceId "f-workspace"`) || !strings.Contains(err.Error(), tc.wantObservedText) {
				t.Fatalf("ReadTree() error = %T %v, want deterministic requested/observed identity error", err, err)
			}
			var retryable interface{ Retryable() bool }
			if errors.As(err, &retryable) && retryable.Retryable() {
				t.Fatalf("ReadTree() identity error must fail closed without readiness retry: %T %v", err, err)
			}
			if flinkCapacityNotFoundError(err) {
				t.Fatalf("ReadTree() identity error must not be classified as not found: %T %v", err, err)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 0 || testFlinkCapacityTargetReadCalls(api) != 0 {
				t.Fatalf("ReadTree() calls: get=%d list=%d namespaces=%d targets=%d, want 1/0/0/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, testFlinkCapacityTargetReadCalls(api))
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Fatalf("ReadTree() made %d capacity writes without exact Workspace identity", writes)
			}
		})
	}
}

func TestFlinkCapacityServiceModifyQueueRejectsSuccessfulGetWithoutExactWorkspaceIdentity(t *testing.T) {
	requestedID := "f-workspace"
	emptyID := testFlinkCapacityWorkspace(2)
	emptyID.Id = ""
	wrongID := testFlinkCapacityWorkspace(2)
	wrongID.Id = "f-workspace-copy"
	wrongID.Name = requestedID
	wrongID.ResourceId = requestedID
	step := flinkcapacity.Step{
		Action: flinkcapacity.ModifyQueue,
		Ref:    flinkcapacity.Ref{Namespace: "default", Queue: "default-queue"},
		To:     flinkcapacity.Allocation{FixedCU: 2, Limit: 2},
	}

	for _, tc := range []struct {
		name             string
		workspace        *flink.Workspace
		wantObservedText string
	}{
		{name: "nil Workspace", workspace: nil, wantObservedText: "<nil>"},
		{name: "empty Workspace ID", workspace: emptyID, wantObservedText: `observed Workspace Id ""`},
		{name: "different Workspace ID despite similar name and matching ResourceId", workspace: wrongID, wantObservedText: `observed Workspace Id "f-workspace-copy"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(tc.workspace)
			var err error
			var panicValue interface{}
			func() {
				defer func() { panicValue = recover() }()
				_, err = (&FlinkCapacityService{api: api}).ApplyStep(context.Background(), requestedID, step)
			}()
			if panicValue != nil {
				t.Errorf("ApplyStep(ModifyQueue) panicked instead of failing closed: %v", panicValue)
			}
			if err == nil || !strings.Contains(err.Error(), `requested InstanceId "f-workspace"`) || !strings.Contains(err.Error(), tc.wantObservedText) {
				t.Errorf("ApplyStep(ModifyQueue) error = %T %v, want deterministic requested/observed identity error", err, err)
			}
			var retryable interface{ Retryable() bool }
			if errors.As(err, &retryable) && retryable.Retryable() {
				t.Errorf("ApplyStep(ModifyQueue) identity error must fail closed without readiness retry: %T %v", err, err)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 0 || testFlinkCapacityTargetReadCalls(api) != 0 || api.namespaceReads != 0 {
				t.Errorf("ApplyStep(ModifyQueue) reads: get=%d list=%d namespaces=%d targets=%d namespace_get=%d, want 1/0/0/0/0", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, testFlinkCapacityTargetReadCalls(api), api.namespaceReads)
			}
			if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
				t.Errorf("ApplyStep(ModifyQueue) made %d capacity writes without exact Workspace identity", writes)
			}
		})
	}
}

func TestFlinkCapacityServiceReconcilePreWriteReadRejectsSuccessfulGetWithoutExactWorkspaceIdentity(t *testing.T) {
	requestedID := "f-workspace"
	desired := flinkcapacity.Tree{
		ChargeType: "PRE",
		Workspace:  flinkcapacity.WorkspaceCapacity{FixedCU: 6, Limit: 6},
		Namespaces: []flinkcapacity.Namespace{{
			Name:     "default",
			Capacity: &flinkcapacity.Capacity{Fixed: 6, Limit: 6},
			Queues:   []flinkcapacity.Queue{{Name: "default-queue", Capacity: &flinkcapacity.Capacity{Fixed: 6, Limit: 6}}},
		}},
	}

	for _, tc := range []struct {
		name             string
		invalidWorkspace func() *flink.Workspace
		wantObservedText string
	}{
		{name: "nil Workspace", invalidWorkspace: func() *flink.Workspace { return nil }, wantObservedText: "<nil>"},
		{name: "empty Workspace ID", invalidWorkspace: func() *flink.Workspace {
			workspace := testFlinkCapacityWorkspace(3)
			workspace.Id = ""
			return workspace
		}, wantObservedText: `observed Workspace Id ""`},
		{name: "different Workspace ID", invalidWorkspace: func() *flink.Workspace {
			workspace := testFlinkCapacityWorkspace(3)
			workspace.Id = "f-workspace-copy"
			workspace.Name = requestedID
			workspace.ResourceId = requestedID
			return workspace
		}, wantObservedText: `observed Workspace Id "f-workspace-copy"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initialWorkspace := testFlinkCapacityWorkspace(2)
			invalidPostWrite := tc.invalidWorkspace()
			api := testFlinkCapacityReadableAPI(initialWorkspace)
			api.getWorkspaceResponses = []testFlinkCapacityWorkspaceResponse{
				{workspace: initialWorkspace},
				{workspace: invalidPostWrite},
			}
			service := &FlinkCapacityService{api: api}

			_, err := (flinkcapacity.Reconciler{API: service}).ReconcileAuthoritative(context.Background(), requestedID, desired)
			if err == nil || !strings.Contains(err.Error(), `requested InstanceId "f-workspace"`) || !strings.Contains(err.Error(), tc.wantObservedText) {
				t.Fatalf("ReconcileAuthoritative() error = %T %v, want pre-write requested/observed identity error", err, err)
			}
			if len(api.workspaceFixedWrites) != 0 || testFlinkCapacityWriteCalls(api) != 0 {
				t.Fatalf("ReconcileAuthoritative() writes = %d (workspace fixed=%d), want zero after pre-write identity failure", testFlinkCapacityWriteCalls(api), len(api.workspaceFixedWrites))
			}
			if api.workspaceReads != 3 || api.listWorkspacesCalls != 0 || api.listNamespacesCalls != 2 || testFlinkCapacityTargetReadCalls(api) != 2 {
				t.Fatalf("ReconcileAuthoritative() reads: get=%d list=%d namespaces=%d targets=%d, want initial tree, strict pre-write get, and final evidence read", api.workspaceReads, api.listWorkspacesCalls, api.listNamespacesCalls, testFlinkCapacityTargetReadCalls(api))
			}
		})
	}
}

func TestFlinkCapacityServiceWorkspaceStepsUseAbsoluteComponentArguments(t *testing.T) {
	initialFixed := &flink.ResourceSpec{Cpu: 4, MemoryGB: 16}
	initialElastic := &flink.ResourceSpec{Cpu: 3, MemoryGB: 12}
	workspace := testFlinkCapacityWorkspace(4)
	workspace.ResourceSpec = initialFixed
	workspace.ElasticResourceSpec = initialElastic
	workspace.Elastic = true
	workspace.ElasticInstanceId = "f-elastic"
	workspace.ElasticOrderState = "NORMAL"
	api := &fakeFlinkCapacityAPI{workspace: workspace}
	service := &FlinkCapacityService{api: api}

	_, err := service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
		Action: flinkcapacity.ModifyWorkspaceFixed,
		To:     flinkcapacity.Allocation{FixedCU: 6, Limit: 12},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(api.workspaceFixedWrites) != 1 || api.workspaceFixedWrites[0].fixed == nil || api.workspaceFixedWrites[0].fixed.Cpu != 3 || api.workspaceFixedWrites[0].crossZone != nil {
		t.Fatalf("single-zone fixed arguments = %#v, want absolute fixed=3 and no cross-zone pool", api.workspaceFixedWrites)
	}
	if api.workspace.ElasticResourceSpec != initialElastic {
		t.Fatal("fixed write replaced the unmodified elastic component")
	}

	_, err = service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
		Action: flinkcapacity.ModifyWorkspaceFixed,
		To:     flinkcapacity.Allocation{CrossZoneFixedCU: 8, Limit: 14},
	})
	if err != nil {
		t.Fatal(err)
	}
	haWrite := api.workspaceFixedWrites[1]
	if haWrite.fixed == nil || haWrite.fixed.Cpu != 0 || haWrite.crossZone == nil || haWrite.crossZone.Cpu != 4 {
		t.Fatalf("HA fixed arguments = %#v, want absolute primary=0 and cross-zone=4", haWrite)
	}
	if api.workspace.ElasticResourceSpec != initialElastic {
		t.Fatal("HA fixed write replaced the unmodified elastic component")
	}

	for _, tc := range []struct {
		action flinkcapacity.Action
		fixed  flinkcapacity.CU
		limit  flinkcapacity.CU
		want   float64
	}{
		{action: flinkcapacity.EnableWorkspaceElastic, fixed: 8, limit: 14, want: 3},
		{action: flinkcapacity.ModifyWorkspaceElastic, fixed: 6, limit: 14, want: 4},
	} {
		fixedBefore := api.workspace.ResourceSpec
		crossBefore := api.workspace.HaResourceSpec
		_, err = service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
			Action: tc.action,
			To:     flinkcapacity.Allocation{FixedCU: tc.fixed, Limit: tc.limit},
		})
		if err != nil {
			t.Fatal(err)
		}
		write := api.workspaceElasticWrites[len(api.workspaceElasticWrites)-1]
		if write.action != tc.action || write.elastic == nil || write.elastic.Cpu != tc.want {
			t.Fatalf("%s arguments = %#v, want absolute elastic=%v", tc.action, write, tc.want)
		}
		if api.workspace.ResourceSpec != fixedBefore || api.workspace.HaResourceSpec != crossBefore {
			t.Fatalf("%s write replaced an unmodified fixed component", tc.action)
		}
	}
}

func TestFlinkCapacityServiceApplyStepStrictlyRevalidatesWorkspaceBeforeEveryWrite(t *testing.T) {
	steps := []flinkcapacity.Step{
		{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{FixedCU: 4, Limit: 4}},
		{Action: flinkcapacity.EnableWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 4}},
		{Action: flinkcapacity.ModifyWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 4}},
		{Action: flinkcapacity.CreateNamespace, Ref: flinkcapacity.Ref{Namespace: "new"}},
		{Action: flinkcapacity.DeleteNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}},
		{Action: flinkcapacity.ModifyNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 2}},
		{Action: flinkcapacity.ModifyQueue, Ref: flinkcapacity.Ref{Namespace: "default", Queue: "default-queue"}, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 2}},
	}
	for _, step := range steps {
		t.Run(string(step.Action), func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(testFlinkCapacityWorkspace(2))
			service := &FlinkCapacityService{api: api}
			_, err := service.ApplyStep(context.Background(), "f-workspace", step)
			if err != nil {
				t.Fatal(err)
			}
			if len(api.calls) < 2 || api.calls[0] != "get-workspace" {
				t.Fatalf("ApplyStep call order = %v, want strict GetWorkspace before any write", api.calls)
			}
			if api.listWorkspacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 1 {
				t.Fatalf("ApplyStep list fallback/writes = %d/%d, want 0/1", api.listWorkspacesCalls, testFlinkCapacityWriteCalls(api))
			}
		})
	}
}

func TestFlinkCapacityServiceApplyStepCancellationAfterPreflightReadIsZeroWrite(t *testing.T) {
	t.Run("workspace read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		api := testFlinkCapacityReadableAPI(testFlinkCapacityWorkspace(2))
		api.getWorkspaceHook = cancel
		_, err := (&FlinkCapacityService{api: api}).ApplyStep(ctx, "f-workspace", flinkcapacity.Step{
			Action: flinkcapacity.ModifyWorkspaceFixed,
			To:     flinkcapacity.Allocation{FixedCU: 4, Limit: 4},
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ApplyStep error = %T %v, want context cancellation after Workspace preflight", err, err)
		}
		if writes := testFlinkCapacityWriteCalls(api); writes != 0 {
			t.Fatalf("ApplyStep made %d write(s) after context cancellation during Workspace preflight", writes)
		}
	})

	t.Run("namespace read", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		api := testFlinkCapacityReadableAPI(testFlinkCapacityWorkspace(2))
		api.getNamespaceHook = cancel
		_, err := (&FlinkCapacityService{api: api}).ApplyStep(ctx, "f-workspace", flinkcapacity.Step{
			Action: flinkcapacity.ModifyNamespace,
			Ref:    flinkcapacity.Ref{Namespace: "default"},
			To:     flinkcapacity.Allocation{FixedCU: 4, Limit: 4},
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ApplyStep error = %T %v, want context cancellation after Namespace preflight", err, err)
		}
		if api.namespaceReads != 1 || testFlinkCapacityWriteCalls(api) != 0 {
			t.Fatalf("ApplyStep namespace reads/writes = %d/%d, want 1/0 after cancellation", api.namespaceReads, testFlinkCapacityWriteCalls(api))
		}
	})
}

func TestFlinkCapacityServiceTerminalStateWinsOverMissingBootstrapProvenance(t *testing.T) {
	workspace := testFlinkCapacityWorkspace(2)
	workspace.Status = "DELETING"
	workspace.ResourceId = ""
	workspace.Tags = nil
	api := testFlinkCapacityReadableAPI(workspace)
	contextValue := testFlinkCapacityBootstrapContext("f-workspace", "resource-workspace", 1_700_000_900)
	service := (&FlinkCapacityService{api: api}).withCapacityBootstrapContext(contextValue, true, func() time.Time {
		return time.Unix(1_700_000_000, 0)
	})

	_, err := service.ReadTree(context.Background(), "f-workspace")
	if err == nil || !strings.Contains(err.Error(), "terminal state") {
		t.Fatalf("ReadTree error = %T %v, want terminal failure before missing provenance readiness", err, err)
	}
	var notReady *flinkcapacity.NotReadyError
	if errors.As(err, &notReady) {
		t.Fatalf("terminal Workspace was misclassified as retryable: %v", err)
	}
	if api.listNamespacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 0 {
		t.Fatalf("terminal Workspace downstream reads/writes = %d/%d, want 0/0", api.listNamespacesCalls, testFlinkCapacityWriteCalls(api))
	}
}

func TestFlinkCapacityServiceApplyStepIdentityFailureIsZeroWriteAndNeverLists(t *testing.T) {
	get404 := flink.NewFlinkServiceErrorWithCode("get-request", "", "404", "workspace not found", "")
	for _, test := range []struct {
		name      string
		workspace *flink.Workspace
		getErr    error
		context   *flinkWorkspaceCapacityBootstrapContext
		want      string
	}{
		{name: "typed 404", getErr: get404, want: "workspace not found"},
		{name: "wrong exact ID", workspace: &flink.Workspace{Id: "f-other", ResourceId: "resource-workspace", Status: "RUNNING", OrderState: "NORMAL"}, want: "identity mismatch"},
		{name: "terminal", workspace: &flink.Workspace{Id: "f-workspace", ResourceId: "resource-workspace", Status: "DELETING", OrderState: "NORMAL"}, want: "terminal state"},
		{name: "missing ResourceId", workspace: &flink.Workspace{Id: "f-workspace", Status: "RUNNING", OrderState: "NORMAL"}, want: "ResourceId"},
		{name: "bootstrap tag mismatch", workspace: func() *flink.Workspace {
			value := testFlinkCapacityWorkspace(2)
			value.Tags = []flink.Tag{
				{Key: flinkworkspace.CreateTokenTagKey, Value: strings.Repeat("c", 64)},
				{Key: flinkworkspace.CreateIntentTagKey, Value: strings.Repeat("b", 64)},
			}
			return value
		}(), context: func() *flinkWorkspaceCapacityBootstrapContext {
			value := testFlinkCapacityBootstrapContext("f-workspace", "resource-workspace", 1_700_000_900)
			return &value
		}(), want: "create token mismatch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := testFlinkCapacityReadableAPI(test.workspace)
			api.getWorkspaceErr = test.getErr
			service := &FlinkCapacityService{api: api}
			if test.context != nil {
				service = service.withCapacityBootstrapContext(*test.context, true, func() time.Time { return time.Unix(1_700_000_000, 0) })
			}
			_, err := service.ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
				Action: flinkcapacity.CreateNamespace,
				Ref:    flinkcapacity.Ref{Namespace: "new"},
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ApplyStep error = %T %v, want text %q", err, err, test.want)
			}
			if api.workspaceReads != 1 || api.listWorkspacesCalls != 0 || testFlinkCapacityWriteCalls(api) != 0 {
				t.Fatalf("identity failure calls get/list/writes = %d/%d/%d, want 1/0/0", api.workspaceReads, api.listWorkspacesCalls, testFlinkCapacityWriteCalls(api))
			}
		})
	}
}

func TestFlinkCapacityServiceRejectsFOASInt32ResourceSpecOverflowBeforeAPI(t *testing.T) {
	maximum := flinkcapacity.CU(flinkMaxCUBeforeInt32MemoryOverflow * 2)
	overflow := maximum + 2
	tests := []struct {
		name string
		step flinkcapacity.Step
	}{
		{name: "workspace primary fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{FixedCU: overflow, Limit: overflow}}},
		{name: "workspace cross-zone fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{CrossZoneFixedCU: overflow, Limit: overflow}}},
		{name: "postpaid workspace", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspacePostpaid, To: flinkcapacity.Allocation{Limit: overflow}}},
		{name: "enable elastic", step: flinkcapacity.Step{Action: flinkcapacity.EnableWorkspaceElastic, To: flinkcapacity.Allocation{Limit: overflow}}},
		{name: "modify elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceElastic, To: flinkcapacity.Allocation{Limit: overflow}}},
		{name: "namespace fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}, To: flinkcapacity.Allocation{FixedCU: overflow, Limit: overflow}}},
		{name: "namespace elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}, To: flinkcapacity.Allocation{FixedCU: 2, Limit: overflow + 2}}},
		{name: "queue fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyQueue, Ref: flinkcapacity.Ref{Namespace: "default", Queue: "default-queue"}, To: flinkcapacity.Allocation{FixedCU: overflow, Limit: overflow}}},
		{name: "queue limit", step: flinkcapacity.Step{Action: flinkcapacity.ModifyQueue, Ref: flinkcapacity.Ref{Namespace: "default", Queue: "default-queue"}, To: flinkcapacity.Allocation{FixedCU: 2, Limit: overflow}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacitySerializationBoundaryAPI()
			_, err := (&FlinkCapacityService{api: api}).ApplyStep(context.Background(), "f-workspace", tc.step)
			if err == nil || !strings.Contains(err.Error(), "int32") {
				t.Fatalf("error=%v, want provider-side int32 range rejection", err)
			}
			if writes := testFlinkCapacityWriteCount(api); writes != 0 {
				t.Fatalf("API writes=%d, want rejection before API for step %#v", writes, tc.step)
			}
		})
	}

	t.Run("maximum remains serializable", func(t *testing.T) {
		api := testFlinkCapacitySerializationBoundaryAPI()
		_, err := (&FlinkCapacityService{api: api}).ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
			Action: flinkcapacity.ModifyWorkspaceFixed,
			To:     flinkcapacity.Allocation{FixedCU: maximum, Limit: maximum},
		})
		if err != nil {
			t.Fatalf("maximum legal component was rejected: %v", err)
		}
		if len(api.workspaceFixedWrites) != 1 || api.workspaceFixedWrites[0].fixed.Cpu != float64(flinkMaxCUBeforeInt32MemoryOverflow) || api.workspaceFixedWrites[0].fixed.MemoryGB != float64(flinkMaxCUBeforeInt32MemoryOverflow*4) {
			t.Fatalf("maximum fixed request=%#v, want exact non-overflowing CPU/MemoryGB", api.workspaceFixedWrites)
		}
	})
}

func TestFlinkCapacityServiceRejectsFractionalFOASComponentsBeforeReadsOrAPI(t *testing.T) {
	tests := []struct {
		name string
		step flinkcapacity.Step
	}{
		{name: "workspace primary fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{FixedCU: 1, Limit: 2}}},
		{name: "workspace cross-zone fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{CrossZoneFixedCU: 1, Limit: 2}}},
		{name: "postpaid workspace", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspacePostpaid, To: flinkcapacity.Allocation{Limit: 1}}},
		{name: "enable elastic", step: flinkcapacity.Step{Action: flinkcapacity.EnableWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 3}}},
		{name: "modify elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 3}}},
		{name: "namespace fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}, To: flinkcapacity.Allocation{FixedCU: 1, Limit: 2}}},
		{name: "namespace elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, Ref: flinkcapacity.Ref{Namespace: "default"}, To: flinkcapacity.Allocation{FixedCU: 2, Limit: 3}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := testFlinkCapacitySerializationBoundaryAPI()
			_, err := (&FlinkCapacityService{api: api}).ApplyStep(context.Background(), "f-workspace", tc.step)
			if err == nil || !strings.Contains(err.Error(), "integer CU") {
				t.Fatalf("error=%v, want provider-side integral FOAS component rejection", err)
			}
			if reads := api.workspaceReads + api.namespaceReads; reads != 0 {
				t.Fatalf("API reads=%d (workspace=%d namespace=%d), want rejection before read for step %#v", reads, api.workspaceReads, api.namespaceReads, tc.step)
			}
			if writes := testFlinkCapacityWriteCount(api); writes != 0 {
				t.Fatalf("API writes=%d, want rejection before API for step %#v", writes, tc.step)
			}
		})
	}
}

func TestFlinkCapacityServiceIntegralGatePreservesMaximumFOASAndFractionalQueueV2(t *testing.T) {
	maximum := flinkcapacity.CU(flinkMaxCUBeforeInt32MemoryOverflow * 2)
	for _, tc := range []struct {
		name string
		step flinkcapacity.Step
	}{
		{name: "workspace primary fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{FixedCU: maximum, Limit: maximum}}},
		{name: "workspace cross-zone fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceFixed, To: flinkcapacity.Allocation{CrossZoneFixedCU: maximum, Limit: maximum}}},
		{name: "postpaid workspace", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspacePostpaid, To: flinkcapacity.Allocation{Limit: maximum}}},
		{name: "enable elastic", step: flinkcapacity.Step{Action: flinkcapacity.EnableWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: maximum + 2}}},
		{name: "modify elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyWorkspaceElastic, To: flinkcapacity.Allocation{FixedCU: 2, Limit: maximum + 2}}},
		{name: "namespace fixed", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, To: flinkcapacity.Allocation{FixedCU: maximum, Limit: maximum}}},
		{name: "namespace elastic", step: flinkcapacity.Step{Action: flinkcapacity.ModifyNamespace, To: flinkcapacity.Allocation{FixedCU: 2, Limit: maximum + 2}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateFlinkCapacityStepSerializable(tc.step); err != nil {
				t.Fatalf("maximum legal FOAS component was rejected: %v", err)
			}
		})
	}

	t.Run("queue V2 retains half CU", func(t *testing.T) {
		api := testFlinkCapacitySerializationBoundaryAPI()
		_, err := (&FlinkCapacityService{api: api}).ApplyStep(context.Background(), "f-workspace", flinkcapacity.Step{
			Action: flinkcapacity.ModifyQueue,
			Ref:    flinkcapacity.Ref{Namespace: "default", Queue: "default-queue"},
			To:     flinkcapacity.Allocation{FixedCU: 1, Limit: 3},
		})
		if err != nil {
			t.Fatalf("fractional Queue V2 capacity was rejected: %v", err)
		}
		if api.workspaceReads != 1 || api.namespaceReads != 0 || api.queueWrites != 1 {
			t.Fatalf("queue V2 calls: workspaceReads=%d namespaceReads=%d queueWrites=%d", api.workspaceReads, api.namespaceReads, api.queueWrites)
		}
		quota := api.targets["default"][0].Quota
		if quota == nil || quota.Request == nil || quota.Limit == nil || quota.Request.Cpu != 0.5 || quota.Limit.Cpu != 1.5 {
			t.Fatalf("fractional Queue V2 quota=%#v, want request=0.5 CU limit=1.5 CU", quota)
		}
	})
}

func testFlinkCapacitySerializationBoundaryAPI() *fakeFlinkCapacityAPI {
	return &fakeFlinkCapacityAPI{
		workspace: &flink.Workspace{Id: "f-workspace", ResourceId: "resource-workspace", Status: "RUNNING", OrderState: "NORMAL"},
		namespaces: []flink.Namespace{{
			Name: "default",
		}},
		targets: map[string][]flink.DeploymentTarget{
			"default": {{Name: "default-queue"}},
		},
	}
}

func testFlinkCapacityWriteCount(api *fakeFlinkCapacityAPI) int {
	return len(api.created) + len(api.deleted) + api.namespaceWrites + api.queueWrites + len(api.workspaceFixedWrites) + api.workspacePostpaidWrites + len(api.workspaceElasticWrites)
}

func TestFlinkCapacityServiceApplyNamespaceTopologySteps(t *testing.T) {
	api := &fakeFlinkCapacityAPI{workspace: testFlinkCapacityWorkspace(2)}
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
			api := &fakeFlinkCapacityAPI{workspace: testFlinkCapacityWorkspace(2), createPostReadErr: postReadErr}
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
			api := &fakeFlinkCapacityAPI{workspace: testFlinkCapacityWorkspace(2), createPreWriteErr: tc.err}
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
	api := &fakeFlinkCapacityAPI{workspace: testFlinkCapacityWorkspace(2), deleteErr: wantErr}
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

func testFlinkCapacityReadableAPI(workspace *flink.Workspace) *fakeFlinkCapacityAPI {
	return &fakeFlinkCapacityAPI{
		workspace:  workspace,
		namespaces: []flink.Namespace{testFlinkCapacityNamespace("default", false, 2)},
		targets: map[string][]flink.DeploymentTarget{
			"default": {testFlinkCapacityQueue("default-queue", 2)},
		},
	}
}

func testFlinkCapacityWriteCalls(api *fakeFlinkCapacityAPI) int {
	return len(api.created) + len(api.deleted) + api.namespaceWrites + api.queueWrites +
		len(api.workspaceFixedWrites) + api.workspacePostpaidWrites + len(api.workspaceElasticWrites)
}

func testFlinkCapacityTargetReadCalls(api *fakeFlinkCapacityAPI) int {
	total := 0
	for _, calls := range api.listTargetCalls {
		total += calls
	}
	return total
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

func testFlinkCapacityBootstrapContext(instanceID, resourceID string, deadline int64) flinkWorkspaceCapacityBootstrapContext {
	return flinkWorkspaceCapacityBootstrapContext{
		Version:                          flinkWorkspaceCapacityBootstrapContextVersion,
		Origin:                           flinkWorkspaceCapacityBootstrapContextManagedInitialCreate,
		ExpectedInstanceID:               instanceID,
		ExpectedResourceID:               resourceID,
		TerraformCreateToken:             strings.Repeat("a", 64),
		CreateIntentFingerprint:          strings.Repeat("b", 64),
		IdentityAbsenceRetryNotAfterUnix: deadline,
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

func TestValidateFlinkCapacityWorkspaceRejectsKnownTerminalStates(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*flink.Workspace)
	}{
		{name: "workspace DISABLE", mutate: func(value *flink.Workspace) { value.Status = "DISABLE" }},
		{name: "workspace DELETING", mutate: func(value *flink.Workspace) { value.Status = "DELETING" }},
		{name: "workspace DELETED", mutate: func(value *flink.Workspace) { value.Status = "DELETED" }},
		{name: "order CEASE", mutate: func(value *flink.Workspace) { value.OrderState = "CEASE" }},
		{name: "order CEASED", mutate: func(value *flink.Workspace) { value.OrderState = "CEASED" }},
		{name: "order RELEASE", mutate: func(value *flink.Workspace) { value.OrderState = "RELEASE" }},
		{name: "order RELEASED", mutate: func(value *flink.Workspace) { value.OrderState = "RELEASED" }},
		{name: "order RELEASING", mutate: func(value *flink.Workspace) { value.OrderState = "RELEASING" }},
		{name: "elastic order RELEASED", mutate: func(value *flink.Workspace) {
			value.Elastic = true
			value.ElasticInstanceId = "f-elastic"
			value.ElasticResourceSpec = &flink.ResourceSpec{Cpu: 2, MemoryGB: 8}
			value.ElasticOrderState = "RELEASED"
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := &flink.Workspace{Id: "f-test", Status: "RUNNING", OrderState: "NORMAL", ResourceId: "resource-test"}
			test.mutate(workspace)
			err := validateFlinkCapacityWorkspaceReady(workspace)
			if err == nil || !strings.Contains(err.Error(), "terminal state") {
				t.Fatalf("terminal workspace error = %T %v", err, err)
			}
			var retryable interface{ Retryable() bool }
			if errors.As(err, &retryable) && retryable.Retryable() {
				t.Fatalf("terminal workspace state became retryable: %v", err)
			}
		})
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
