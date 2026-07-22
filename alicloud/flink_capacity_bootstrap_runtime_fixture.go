//go:build flink_capacity_bootstrap_runtime_fixture
// +build flink_capacity_bootstrap_runtime_fixture

package alicloud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const flinkCapacityBootstrapRuntimeLedgerEnv = "FLINK_CAPACITY_BOOTSTRAP_RUNTIME_LEDGER"

type flinkCapacityBootstrapRuntimeLedger struct {
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

type flinkCapacityBootstrapRuntimeService struct {
	ledgerPath string
}

type flinkCapacityBootstrapRuntimeAPI struct {
	ledgerPath string
}

// FlinkCapacityBootstrapRuntimeFixtureProvider retains the production
// Workspace/allocation resource registry and callbacks. Only cloud-facing
// factories, topology discovery, and paid deletion are replaced by an
// invocation-owned ledger.
func FlinkCapacityBootstrapRuntimeFixtureProvider() terraform.ResourceProvider {
	return flinkCapacityBootstrapRuntimeProvider(false)
}

// FlinkCapacityBootstrapOldSchemaRuntimeFixtureProvider is a synthetic
// old-schema state generator. It is not a historical provider artifact and is
// used only to prove that the new provider upgrades pre-field state in place.
func FlinkCapacityBootstrapOldSchemaRuntimeFixtureProvider() terraform.ResourceProvider {
	return flinkCapacityBootstrapRuntimeProvider(true)
}

func flinkCapacityBootstrapRuntimeProvider(oldSchema bool) terraform.ResourceProvider {
	provider, ok := Provider().(*schema.Provider)
	if !ok || provider == nil {
		panic(fmt.Sprintf("Provider() returned %T, want *schema.Provider", Provider()))
	}
	provider.ConfigureFunc = func(*schema.ResourceData) (interface{}, error) {
		if _, err := flinkCapacityBootstrapRuntimeLedgerPath(); err != nil {
			return nil, err
		}
		return &connectivity.AliyunClient{RegionId: "cn-test"}, nil
	}

	newFlinkWorkspaceCreateCallbackService = func(*connectivity.AliyunClient) (flinkWorkspaceCreateCallbackService, error) {
		path, err := flinkCapacityBootstrapRuntimeLedgerPath()
		return &flinkCapacityBootstrapRuntimeService{ledgerPath: path}, err
	}
	newFlinkWorkspaceReadCallbackService = func(*connectivity.AliyunClient) (flinkWorkspaceReadCallbackService, error) {
		path, err := flinkCapacityBootstrapRuntimeLedgerPath()
		return &flinkCapacityBootstrapRuntimeService{ledgerPath: path}, err
	}
	newFlinkWorkspaceCapacityAllocationService = func(interface{}) (flinkcapacity.API, error) {
		path, err := flinkCapacityBootstrapRuntimeLedgerPath()
		if err != nil {
			return nil, err
		}
		if _, err := flinkCapacityBootstrapRuntimeMutate(path, func(state *flinkCapacityBootstrapRuntimeLedger) {
			state.CapacityServiceFactories++
		}); err != nil {
			return nil, err
		}
		return &FlinkCapacityService{api: &flinkCapacityBootstrapRuntimeAPI{ledgerPath: path}}, nil
	}
	resolveFlinkWorkspaceCreateTopology = func(_ *connectivity.AliyunClient, vpcID, _, _ string, primaryIDs, _ []string) (flinkworkspace.VSwitchTopology, error) {
		if vpcID != "vpc-fixture" || len(primaryIDs) != 1 || primaryIDs[0] != "vsw-fixture" {
			return flinkworkspace.VSwitchTopology{}, fmt.Errorf("unexpected bootstrap fixture topology VPC=%q primary=%v", vpcID, primaryIDs)
		}
		return flinkworkspace.VSwitchTopology{PrimaryZoneID: "cn-test-a"}, nil
	}
	flinkWorkspaceCapacityBootstrapPollInterval = 10 * time.Millisecond

	workspace := provider.ResourcesMap["alicloud_flink_workspace"]
	allocation := provider.ResourcesMap["alicloud_flink_workspace_capacity_allocation"]
	bootstrap := provider.ResourcesMap["alicloud_flink_workspace_capacity_bootstrap"]
	if workspace == nil || allocation == nil || bootstrap == nil {
		panic("bootstrap runtime fixture provider is missing production resources")
	}
	workspace.Delete = flinkCapacityBootstrapRuntimeParentDelete
	if oldSchema {
		delete(workspace.Schema, "capacity_bootstrap_context")
		workspace.Read = flinkCapacityBootstrapOldSchemaRuntimeParentRead
		delete(provider.ResourcesMap, "alicloud_flink_workspace_capacity_bootstrap")
		delete(provider.ResourcesMap, "alicloud_flink_workspace_capacity_allocation_v2")
	}
	return provider
}

func flinkCapacityBootstrapOldSchemaRuntimeParentRead(d *schema.ResourceData, meta interface{}) error {
	err := resourceAliCloudFlinkWorkspaceRead(d, meta)
	if err != nil && strings.Contains(err.Error(), "set strict Flink workspace capacity bootstrap context: Invalid address to set") {
		return nil
	}
	return err
}

func (s *flinkCapacityBootstrapRuntimeService) CreateInstance(request *flink.Workspace, _ flinkworkspace.CreateOptions) (*flink.Workspace, error) {
	if request == nil {
		return nil, fmt.Errorf("bootstrap fixture CreateInstance request is nil")
	}
	_, err := flinkCapacityBootstrapRuntimeMutate(s.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.ParentExists = true
		state.ParentID = "f-bootstrap-runtime"
		state.ParentResourceID = "resource-bootstrap-runtime"
		state.ParentCreates++
		if state.WorkspaceFixedCU == 0 {
			state.WorkspaceFixedCU = 1
			if request.ResourceSpec != nil {
				state.WorkspaceFixedCU = int(request.ResourceSpec.Cpu)
			}
		}
		if state.NamespaceFixedCU == 0 {
			state.NamespaceFixedCU = 1
		}
		if state.QueueFixedCU == 0 {
			state.QueueFixedCU = 1
		}
		for _, tag := range request.Tags {
			switch tag.Key {
			case flinkworkspace.CreateTokenTagKey:
				state.ParentCreateToken = tag.Value
			case flinkworkspace.CreateIntentTagKey:
				state.ParentCreateFingerprint = tag.Value
			}
		}
	})
	if err != nil {
		return nil, err
	}
	// Match the production cws-lib-go adapter: paid Create returns only the
	// InstanceId. ResourceId and provider-owned tags arrive through refresh.
	return &flink.Workspace{Id: "f-bootstrap-runtime"}, nil
}

func (s *flinkCapacityBootstrapRuntimeService) ListInstances() ([]flink.Workspace, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(s.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		if state.ParentExists {
			state.ParentIdentityLists++
		}
	})
	if err != nil || !state.ParentExists {
		return nil, err
	}
	if !state.DelayedVisibility {
		return []flink.Workspace{flinkCapacityBootstrapRuntimeWorkspace(state, true, true)}, nil
	}
	switch state.ParentIdentityLists {
	case 1:
		return nil, nil
	case 2:
		return []flink.Workspace{flinkCapacityBootstrapRuntimeWorkspace(state, false, false)}, nil
	case 3:
		return []flink.Workspace{flinkCapacityBootstrapRuntimeWorkspace(state, true, false)}, nil
	default:
		return []flink.Workspace{flinkCapacityBootstrapRuntimeWorkspace(state, true, true)}, nil
	}
}

func (*flinkCapacityBootstrapRuntimeService) WaitForWorkspaceStarting(string, time.Duration) error {
	return nil
}

func (s *flinkCapacityBootstrapRuntimeService) DescribeFlinkWorkspace(instanceID string) (*flink.Workspace, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(s.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.ParentReads++
	})
	if err != nil {
		return nil, err
	}
	if !state.ParentExists || instanceID != state.ParentID {
		return nil, flink.NewFlinkServiceErrorWithCode("parent-read", "", "404", "fixture workspace not found", "")
	}
	result := flinkCapacityBootstrapRuntimeWorkspace(state, true, true)
	return &result, nil
}

func flinkCapacityBootstrapRuntimeParentDelete(d *schema.ResourceData, _ interface{}) error {
	path, err := flinkCapacityBootstrapRuntimeLedgerPath()
	if err != nil {
		return err
	}
	_, err = flinkCapacityBootstrapRuntimeMutate(path, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.ParentDeletes++
		state.ParentExists = false
	})
	if err == nil {
		d.SetId("")
	}
	return err
}

func (a *flinkCapacityBootstrapRuntimeAPI) GetWorkspace(instanceID string) (*flink.Workspace, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(a.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.CapacityWorkspaceGets++
	})
	if err != nil {
		return nil, err
	}
	if !state.ParentExists || instanceID != state.ParentID {
		return nil, flink.NewFlinkServiceErrorWithCode("capacity-get", "", "404", "fixture workspace not found", "")
	}
	if state.DelayedVisibility && state.CapacityWorkspaceGets == 1 {
		return nil, flink.NewFlinkServiceErrorWithCode("capacity-get", "", "404", "injected bootstrap identity propagation", "")
	}
	resourceVisible := !state.DelayedVisibility || state.CapacityWorkspaceGets >= 3
	tagsVisible := !state.DelayedVisibility || state.CapacityWorkspaceGets >= 4
	result := flinkCapacityBootstrapRuntimeWorkspace(state, resourceVisible, tagsVisible)
	if state.CapacityFingerprintMismatchOnGet > 0 && state.CapacityWorkspaceGets == state.CapacityFingerprintMismatchOnGet {
		for i := range result.Tags {
			if result.Tags[i].Key == flinkworkspace.CreateIntentTagKey {
				result.Tags[i].Value = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
			}
		}
	}
	if state.CapacityResourceIDMismatchOnGet > 0 && state.CapacityWorkspaceGets == state.CapacityResourceIDMismatchOnGet {
		result.ResourceId = "resource-bootstrap-replacement"
	}
	return &result, nil
}

func (a *flinkCapacityBootstrapRuntimeAPI) ListWorkspaces() ([]flink.Workspace, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(a.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.CapacityWorkspaceLists++
	})
	if err != nil {
		return nil, err
	}
	if !state.ParentExists || (state.DelayedVisibility && state.CapacityWorkspaceGets <= 1) {
		return nil, nil
	}
	return []flink.Workspace{flinkCapacityBootstrapRuntimeWorkspace(state, true, true)}, nil
}

func (a *flinkCapacityBootstrapRuntimeAPI) ListNamespaces(string) ([]flink.Namespace, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(a.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.CapacityNamespaceLists++
	})
	if err != nil {
		return nil, err
	}
	return []flink.Namespace{{
		Name:                   "default",
		Status:                 "SUCCESS",
		GuaranteedResourceSpec: flinkCapacityBootstrapRuntimeSpec(state.NamespaceFixedCU),
		ElasticResourceSpec:    &flink.ResourceSpec{},
	}}, nil
}

func (a *flinkCapacityBootstrapRuntimeAPI) ListDeploymentTargets(_ string, namespace string) ([]flink.DeploymentTarget, error) {
	state, err := flinkCapacityBootstrapRuntimeMutate(a.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.CapacityTargetLists++
	})
	if err != nil {
		return nil, err
	}
	if namespace != "default" {
		return nil, fmt.Errorf("unexpected fixture namespace %q", namespace)
	}
	return []flink.DeploymentTarget{{
		Name:      "default-queue",
		Namespace: "default",
		Quota: &flink.ResourceQuota{
			Request: flinkCapacityBootstrapRuntimeSpec(state.QueueFixedCU),
			Limit:   flinkCapacityBootstrapRuntimeSpec(state.QueueFixedCU),
		},
	}}, nil
}

func (a *flinkCapacityBootstrapRuntimeAPI) GetNamespace(_ string, namespace string) (*flink.Namespace, error) {
	state, err := flinkCapacityBootstrapRuntimeRead(a.ledgerPath)
	if err != nil {
		return nil, err
	}
	if namespace != "default" {
		return nil, flink.NewFlinkServiceErrorWithCode("namespace-get", "", "404", "fixture namespace not found", "")
	}
	return &flink.Namespace{Name: namespace, Status: "SUCCESS", GuaranteedResourceSpec: flinkCapacityBootstrapRuntimeSpec(state.NamespaceFixedCU), ElasticResourceSpec: &flink.ResourceSpec{}}, nil
}

func (a *flinkCapacityBootstrapRuntimeAPI) ModifyPrepayWorkspaceCapacity(_ string, fixed, _ *flink.ResourceSpec) (flink.CapacityOperation, error) {
	value := 0
	if fixed != nil {
		value = int(fixed.Cpu)
	}
	err := a.recordCapacityWrite(func(state *flinkCapacityBootstrapRuntimeLedger) { state.WorkspaceFixedCU = value })
	return flink.CapacityOperation{RequestID: "workspace-fixed"}, err
}

func (a *flinkCapacityBootstrapRuntimeAPI) UpdateNamespaceCapacity(_ string, namespace string, _ bool, fixed, _ *flink.ResourceSpec) (flink.CapacityOperation, error) {
	if namespace != "default" {
		return flink.CapacityOperation{}, fmt.Errorf("unexpected fixture namespace %q", namespace)
	}
	value := 0
	if fixed != nil {
		value = int(fixed.Cpu)
	}
	err := a.recordCapacityWrite(func(state *flinkCapacityBootstrapRuntimeLedger) { state.NamespaceFixedCU = value })
	return flink.CapacityOperation{RequestID: "namespace-fixed"}, err
}

func (a *flinkCapacityBootstrapRuntimeAPI) UpdateDeploymentTargetV2(_ string, namespace string, target *flink.DeploymentTarget) (*flink.DeploymentTarget, error) {
	if namespace != "default" || target == nil || target.Name != "default-queue" || target.Quota == nil || target.Quota.Request == nil {
		return nil, fmt.Errorf("unexpected fixture deployment target request")
	}
	err := a.recordCapacityWrite(func(state *flinkCapacityBootstrapRuntimeLedger) { state.QueueFixedCU = int(target.Quota.Request.Cpu) })
	return target, err
}

func (a *flinkCapacityBootstrapRuntimeAPI) recordCapacityWrite(apply func(*flinkCapacityBootstrapRuntimeLedger)) error {
	injectedFailure := false
	_, err := flinkCapacityBootstrapRuntimeMutate(a.ledgerPath, func(state *flinkCapacityBootstrapRuntimeLedger) {
		state.CapacityWrites++
		if state.DelayedVisibility && state.CapacityWorkspaceGets < 4 {
			state.PrematureCapacityWrites++
		}
		apply(state)
		if state.FailNextCapacityWrite {
			state.FailNextCapacityWrite = false
			state.InjectedCapacityWriteFailures++
			injectedFailure = true
		}
	})
	if err != nil {
		return err
	}
	if injectedFailure {
		return fmt.Errorf("injected bootstrap capacity failure after accepted write")
	}
	return nil
}

func (*flinkCapacityBootstrapRuntimeAPI) CreateNamespace(string, *flink.Namespace) (*flink.Namespace, error) {
	return nil, fmt.Errorf("unexpected bootstrap fixture CreateNamespace")
}
func (*flinkCapacityBootstrapRuntimeAPI) DeleteNamespace(string, string) error {
	return fmt.Errorf("unexpected bootstrap fixture DeleteNamespace")
}
func (*flinkCapacityBootstrapRuntimeAPI) ModifyPostpayWorkspaceCapacity(string, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected bootstrap fixture POST capacity write")
}
func (*flinkCapacityBootstrapRuntimeAPI) EnableWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected bootstrap fixture elastic enable")
}
func (*flinkCapacityBootstrapRuntimeAPI) ModifyWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error) {
	return flink.CapacityOperation{}, fmt.Errorf("unexpected bootstrap fixture elastic modify")
}

func flinkCapacityBootstrapRuntimeWorkspace(state flinkCapacityBootstrapRuntimeLedger, resourceVisible, tagsVisible bool) flink.Workspace {
	workspace := flink.Workspace{
		Id:               state.ParentID,
		Name:             "bootstrap-runtime-parent",
		Region:           "cn-test",
		ResourceGroupId:  "rg-fixture",
		VpcId:            "vpc-fixture",
		VSwitchIds:       []string{"vsw-fixture"},
		ZoneId:           "cn-test-a",
		Status:           "RUNNING",
		OrderState:       "NORMAL",
		ChargeType:       "PRE",
		ArchitectureType: "X86",
		MonitorType:      "ARMS",
		ResourceSpec:     flinkCapacityBootstrapRuntimeSpec(state.WorkspaceFixedCU),
		Storage:          &flink.Storage{Oss: &flink.OSSStorage{Bucket: "fixture-bucket"}},
	}
	if resourceVisible {
		workspace.ResourceId = state.ParentResourceID
	}
	if tagsVisible {
		workspace.Tags = []flink.Tag{
			{Key: flinkworkspace.CreateTokenTagKey, Value: state.ParentCreateToken},
			{Key: flinkworkspace.CreateIntentTagKey, Value: state.ParentCreateFingerprint},
		}
	} else if resourceVisible {
		workspace.Tags = []flink.Tag{{Key: flinkworkspace.CreateTokenTagKey, Value: state.ParentCreateToken}}
	}
	return workspace
}

func flinkCapacityBootstrapRuntimeSpec(cu int) *flink.ResourceSpec {
	return &flink.ResourceSpec{Cpu: float64(cu), MemoryGB: float64(cu * 4)}
}

func flinkCapacityBootstrapRuntimeLedgerPath() (string, error) {
	path := os.Getenv(flinkCapacityBootstrapRuntimeLedgerEnv)
	if path == "" || !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s must be an absolute path", flinkCapacityBootstrapRuntimeLedgerEnv)
	}
	return filepath.Clean(path), nil
}

func flinkCapacityBootstrapRuntimeRead(path string) (flinkCapacityBootstrapRuntimeLedger, error) {
	return flinkCapacityBootstrapRuntimeMutate(path, nil)
}

func flinkCapacityBootstrapRuntimeMutate(path string, mutate func(*flinkCapacityBootstrapRuntimeLedger)) (flinkCapacityBootstrapRuntimeLedger, error) {
	lock, err := os.OpenFile(path+".lock", os.O_RDWR, 0)
	if err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, fmt.Errorf("open bootstrap fixture ledger lock: %w", err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, fmt.Errorf("lock bootstrap fixture ledger: %w", err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	body, err := os.ReadFile(path)
	if err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, err
	}
	var state flinkCapacityBootstrapRuntimeLedger
	if err := json.Unmarshal(body, &state); err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, err
	}
	if mutate == nil {
		return state, nil
	}
	mutate(&state)
	body, err = json.Marshal(state)
	if err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, err
	}
	temp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(temp, append(body, '\n'), 0o600); err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, err
	}
	if err := os.Rename(temp, path); err != nil {
		return flinkCapacityBootstrapRuntimeLedger{}, err
	}
	return state, nil
}
