//go:build flink_legacy_namespace_runtime_fixture
// +build flink_legacy_namespace_runtime_fixture

package alicloud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const flinkLegacyRuntimeFixtureLedgerEnv = "FLINK_LEGACY_NAMESPACE_RUNTIME_LEDGER"

const flinkLegacyRuntimeFixtureLedgerLockSuffix = ".lock"

type flinkLegacyRuntimeFixtureLedger struct {
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

var flinkLegacyRuntimeFixtureLedgerMu sync.Mutex

type flinkLegacyRuntimeFixtureService struct {
	ledgerPath string
	role       string
}

// FlinkLegacyNamespaceRuntimeFixtureProvider serves the real provider registry
// under an explicit build tag. It preserves the affected production callbacks
// and replaces only their cloud-facing service/topology boundary plus unrelated
// refresh/delete callbacks used by the disposable OpenTofu fixture.
func FlinkLegacyNamespaceRuntimeFixtureProvider() terraform.ResourceProvider {
	provider, ok := Provider().(*schema.Provider)
	if !ok || provider == nil {
		panic(fmt.Sprintf("Provider() returned %T, want *schema.Provider", Provider()))
	}
	provider.ConfigureFunc = func(*schema.ResourceData) (interface{}, error) {
		if _, err := flinkLegacyRuntimeFixtureValidateEnvironment(); err != nil {
			return nil, err
		}
		return &connectivity.AliyunClient{RegionId: "cn-test"}, nil
	}

	newFlinkWorkspaceCreateCallbackService = func(*connectivity.AliyunClient) (flinkWorkspaceCreateCallbackService, error) {
		ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
		if err != nil {
			return nil, err
		}
		return &flinkLegacyRuntimeFixtureService{ledgerPath: ledgerPath, role: "parent"}, nil
	}
	newFlinkNamespacesReadinessService = func(*connectivity.AliyunClient) (flinkLegacyNamespaceReadinessService, error) {
		ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
		if err != nil {
			return nil, err
		}
		if _, err := flinkLegacyRuntimeFixtureMutate(ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
			state.DataServiceFactories++
		}); err != nil {
			return nil, err
		}
		return &flinkLegacyRuntimeFixtureService{ledgerPath: ledgerPath, role: "data"}, nil
	}
	newFlinkNamespaceCreateService = func(*connectivity.AliyunClient) (flinkNamespaceCreateService, error) {
		ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
		if err != nil {
			return nil, err
		}
		if _, err := flinkLegacyRuntimeFixtureMutate(ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
			state.ChildServiceFactories++
		}); err != nil {
			return nil, err
		}
		return &flinkLegacyRuntimeFixtureService{ledgerPath: ledgerPath, role: "child"}, nil
	}
	resolveFlinkWorkspaceCreateTopology = func(_ *connectivity.AliyunClient, vpcID, _, _ string, primaryIDs, _ []string) (flinkworkspace.VSwitchTopology, error) {
		if vpcID != "vpc-fixture" || len(primaryIDs) != 1 || primaryIDs[0] != "vsw-fixture" {
			return flinkworkspace.VSwitchTopology{}, fmt.Errorf("unexpected fixture topology VPC=%q primary=%v", vpcID, primaryIDs)
		}
		return flinkworkspace.VSwitchTopology{PrimaryZoneID: "cn-test-a"}, nil
	}
	flinkLegacyNamespaceReadinessPollInterval = 5 * time.Millisecond
	flinkLegacyNamespaceReadinessTimeout = func(*schema.ResourceData, string) time.Duration {
		ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
		if err == nil {
			if state, readErr := flinkLegacyRuntimeFixtureRead(ledgerPath); readErr == nil && state.ErrorMode == "cancel" {
				return 5 * time.Second
			}
		}
		return 180 * time.Millisecond
	}

	parent := provider.ResourcesMap["alicloud_flink_workspace"]
	child := provider.ResourcesMap["alicloud_flink_namespace"]
	if parent == nil || child == nil {
		panic("legacy runtime fixture provider is missing workspace or namespace resources")
	}
	parent.Read = flinkLegacyRuntimeFixtureParentRead
	parent.Delete = flinkLegacyRuntimeFixtureParentDelete
	child.Read = flinkLegacyRuntimeFixtureChildRead
	child.Delete = flinkLegacyRuntimeFixtureChildDelete
	return provider
}

func (s *flinkLegacyRuntimeFixtureService) CreateInstance(request *flink.Workspace, _ flinkworkspace.CreateOptions) (*flink.Workspace, error) {
	if request == nil {
		return nil, fmt.Errorf("fixture CreateInstance request is nil")
	}
	_, err := flinkLegacyRuntimeFixtureMutate(s.ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		state.ParentCreates++
		state.ParentExists = true
		state.ParentID = "f-legacy-runtime-parent"
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
	result := *request
	result.Id = "f-legacy-runtime-parent"
	result.Status = flink.FlinkWorkspaceStatusRunning.String()
	result.ResourceId = "workspace-resource"
	return &result, nil
}

func (s *flinkLegacyRuntimeFixtureService) ListInstances() ([]flink.Workspace, error) {
	state, err := flinkLegacyRuntimeFixtureRead(s.ledgerPath)
	if err != nil || !state.ParentExists {
		return nil, err
	}
	return []flink.Workspace{{
		Id: state.ParentID, Name: "legacy-runtime-parent", Region: "cn-test",
		Status: flink.FlinkWorkspaceStatusRunning.String(), ResourceId: "workspace-resource",
		VpcId: "vpc-fixture", ChargeType: "PRE", ArchitectureType: "X86", ResourceGroupId: "rg-fixture",
		VSwitchIds: []string{"vsw-fixture"}, Storage: &flink.Storage{Oss: &flink.OSSStorage{Bucket: "fixture-bucket"}},
		ResourceSpec: &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		Tags: []flink.Tag{
			{Key: flinkworkspace.CreateTokenTagKey, Value: state.ParentCreateToken},
			{Key: flinkworkspace.CreateIntentTagKey, Value: state.ParentCreateFingerprint},
		},
	}}, nil
}

func (*flinkLegacyRuntimeFixtureService) WaitForWorkspaceStarting(string, time.Duration) error {
	return nil
}

func (s *flinkLegacyRuntimeFixtureService) DescribeFlinkWorkspace(workspaceID string) (*flink.Workspace, error) {
	state, err := flinkLegacyRuntimeFixtureMutate(s.ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		state.WorkspacePolls++
		if state.ErrorMode != "" {
			state.ErrorObservations++
		}
	})
	if err != nil {
		return nil, err
	}
	if workspaceID != state.ParentID || !state.ParentExists {
		return nil, flink.NewFlinkServiceErrorWithCode("fixture", "", "404", "fixture workspace is not visible", "")
	}
	switch state.ErrorMode {
	case "permission":
		return nil, flink.NewFlinkServiceErrorWithCode("fixture", "", "Forbidden", "injected dependent permission error", "")
	case "business":
		return nil, flink.NewFlinkServiceErrorWithCode("fixture", "", "InvalidParameter", "injected dependent business error", "")
	case "terminal":
		return &flink.Workspace{Id: workspaceID, Name: "legacy-runtime-parent", Status: "FAILED"}, nil
	case "timeout", "cancel":
		return &flink.Workspace{Id: workspaceID, Name: "legacy-runtime-parent", Status: "CREATING"}, nil
	}

	workspace := &flink.Workspace{Id: workspaceID, Name: "legacy-runtime-parent"}
	switch state.WorkspacePolls {
	case 1:
		return nil, flink.NewFlinkServiceErrorWithCode("fixture", "", "404", "injected dependent delayed visibility", "")
	case 2:
		workspace.Status = "CREATING"
	case 3:
		workspace.Status = flink.FlinkWorkspaceStatusRunning.String()
	default:
		workspace.Status = flink.FlinkWorkspaceStatusRunning.String()
		workspace.ResourceId = "workspace-resource"
	}
	return workspace, nil
}

func (s *flinkLegacyRuntimeFixtureService) ListNamespaces(workspaceID string) ([]flink.Namespace, error) {
	state, err := flinkLegacyRuntimeFixtureMutate(s.ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		state.DefaultNamespacePolls++
		if state.ChildCreated {
			state.ChildNamespacePolls++
		}
	})
	if err != nil {
		return nil, err
	}
	if workspaceID != state.ParentID || !state.ParentExists {
		return nil, flink.NewFlinkServiceErrorWithCode("fixture", "", "404", "fixture namespaces are not visible", "")
	}
	if state.DefaultNamespacePolls < 2 {
		return nil, nil
	}
	namespaces := []flink.Namespace{
		flinkLegacyRuntimeFixtureReadyNamespace("legacy-runtime-parent-default"),
		{Name: "failed-unrelated", Status: "FAILED"},
		{Name: "creating-sibling", Status: "CREATING"},
	}
	if state.ChildCreated {
		child := flinkLegacyRuntimeFixtureReadyNamespace("analytics")
		if state.ChildNamespacePolls < 2 {
			child.Status = "CREATING"
			child.ResourceSpec = nil
		}
		namespaces = append(namespaces, child)
	}
	return namespaces, nil
}

func (s *flinkLegacyRuntimeFixtureService) ListFlinkDeploymentTargets(resourceID, namespace string) ([]flink.DeploymentTarget, error) {
	if resourceID != "workspace-resource" {
		return nil, fmt.Errorf("unexpected fixture namespace resource ID %q", resourceID)
	}
	state, err := flinkLegacyRuntimeFixtureMutate(s.ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		switch namespace {
		case "legacy-runtime-parent-default":
			state.DefaultQueuePolls++
		case "analytics":
			state.ChildQueuePolls++
		default:
			state.UnrelatedQueuePolls++
		}
	})
	if err != nil {
		return nil, err
	}
	switch namespace {
	case "legacy-runtime-parent-default":
		if state.DefaultQueuePolls < 2 {
			return nil, nil
		}
	case "analytics":
		if state.ChildQueuePolls < 2 {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("readiness queried unrelated fixture namespace %q", namespace)
	}
	return []flink.DeploymentTarget{flinkLegacyRuntimeFixtureReadyQueue(namespace)}, nil
}

func (s *flinkLegacyRuntimeFixtureService) CreateNamespace(workspaceID string, namespace *flink.Namespace) (*flink.Namespace, error) {
	if namespace == nil || namespace.Name != "analytics" {
		return nil, fmt.Errorf("unexpected fixture namespace create request %#v", namespace)
	}
	state, err := flinkLegacyRuntimeFixtureMutate(s.ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		state.ChildCreates++
		state.ChildCreated = true
		state.ChildPostCreateReadErrors++
	})
	if err != nil {
		return nil, err
	}
	if workspaceID != state.ParentID || !state.ParentExists {
		return nil, fmt.Errorf("create namespace against missing fixture workspace %q", workspaceID)
	}
	return nil, flink.NewFlinkPostCreateReadError(workspaceID, namespace.Name, flink.NewFlinkServiceErrorWithCode("fixture", "", "404", "injected dependent post-create visibility", ""))
}

func flinkLegacyRuntimeFixtureReadyNamespace(name string) flink.Namespace {
	return flink.Namespace{Name: name, Status: "SUCCESS", ResourceSpec: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16}}
}

func flinkLegacyRuntimeFixtureReadyQueue(namespace string) flink.DeploymentTarget {
	return flink.DeploymentTarget{
		Name: flink.DefaultDeploymentTarget, Namespace: namespace,
		Quota: &flink.ResourceQuota{
			Request: &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
			Limit:   &flink.ResourceSpec{Cpu: 4, MemoryGB: 16},
		},
	}
}

func flinkLegacyRuntimeFixtureParentRead(d *schema.ResourceData, _ interface{}) error {
	ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
	if err != nil {
		return err
	}
	state, err := flinkLegacyRuntimeFixtureRead(ledgerPath)
	if err != nil {
		return err
	}
	if !state.ParentExists || state.ParentID != d.Id() {
		d.SetId("")
		return nil
	}
	workspace := &flink.Workspace{
		Id:               state.ParentID,
		Name:             "legacy-runtime-parent",
		Status:           flink.FlinkWorkspaceStatusRunning.String(),
		Region:           "cn-test",
		ZoneId:           "cn-test-a",
		ArchitectureType: "X86",
		ChargeType:       "PRE",
		MonitorType:      "ARMS",
		ResourceId:       "workspace-resource",
		ResourceGroupId:  "rg-fixture",
		VpcId:            "vpc-fixture",
		VSwitchIds:       []string{"vsw-fixture"},
		ResourceSpec:     &flink.ResourceSpec{Cpu: 2, MemoryGB: 8},
		Storage:          &flink.Storage{Oss: &flink.OSSStorage{Bucket: "fixture-bucket"}},
		Tags: []flink.Tag{
			{Key: flinkworkspace.CreateTokenTagKey, Value: d.Get("terraform_create_token").(string)},
			{Key: flinkworkspace.CreateIntentTagKey, Value: d.Get("create_intent_fingerprint").(string)},
		},
	}
	return applyFlinkWorkspaceReadState(d, workspace)
}

func flinkLegacyRuntimeFixtureParentDelete(d *schema.ResourceData, _ interface{}) error {
	ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
	if err != nil {
		return err
	}
	_, err = flinkLegacyRuntimeFixtureMutate(ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		if d.Get("charge_type").(string) == "PRE" {
			state.ParentRefunds++
		} else {
			state.ParentDeletes++
		}
		state.ParentExists = false
	})
	if err == nil {
		d.SetId("")
	}
	return err
}

func flinkLegacyRuntimeFixtureChildRead(d *schema.ResourceData, _ interface{}) error {
	ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
	if err != nil {
		return err
	}
	state, err := flinkLegacyRuntimeFixtureRead(ledgerPath)
	if err != nil {
		return err
	}
	if !state.ChildCreated {
		d.SetId("")
		return nil
	}
	return d.Set("status", "SUCCESS")
}

func flinkLegacyRuntimeFixtureChildDelete(d *schema.ResourceData, _ interface{}) error {
	ledgerPath, err := flinkLegacyRuntimeFixtureLedgerPath()
	if err != nil {
		return err
	}
	_, err = flinkLegacyRuntimeFixtureMutate(ledgerPath, func(state *flinkLegacyRuntimeFixtureLedger) {
		state.ChildDeletes++
		state.ChildCreated = false
	})
	if err == nil {
		d.SetId("")
	}
	return err
}

func flinkLegacyRuntimeFixtureValidateEnvironment() (string, error) {
	for _, name := range []string{
		"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy",
		"SSH_AUTH_SOCK", "ALICLOUD_ACCESS_KEY", "ALICLOUD_SECRET_KEY", "ALICLOUD_SECURITY_TOKEN",
		"ALIBABA_CLOUD_ACCESS_KEY_ID", "ALIBABA_CLOUD_ACCESS_KEY_SECRET", "AWS_ACCESS_KEY_ID", "GITHUB_TOKEN", "TF_LOG", "TF_LOG_PATH",
	} {
		if value, exists := os.LookupEnv(name); exists {
			return "", fmt.Errorf("legacy runtime fixture inherited forbidden %s length %d", name, len(value))
		}
	}
	return flinkLegacyRuntimeFixtureLedgerPath()
}

func flinkLegacyRuntimeFixtureLedgerPath() (string, error) {
	path := os.Getenv(flinkLegacyRuntimeFixtureLedgerEnv)
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("%s must be a normalized absolute path", flinkLegacyRuntimeFixtureLedgerEnv)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("inspect legacy runtime fixture ledger: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return "", fmt.Errorf("legacy runtime fixture ledger must be a mode 0600 regular file")
	}
	if err := flinkLegacyRuntimeFixtureWithLedgerLock(path, false, func() error { return nil }); err != nil {
		return "", err
	}
	return path, nil
}

func flinkLegacyRuntimeFixtureRead(path string) (flinkLegacyRuntimeFixtureLedger, error) {
	flinkLegacyRuntimeFixtureLedgerMu.Lock()
	defer flinkLegacyRuntimeFixtureLedgerMu.Unlock()
	var state flinkLegacyRuntimeFixtureLedger
	err := flinkLegacyRuntimeFixtureWithLedgerLock(path, false, func() error {
		var err error
		state, err = flinkLegacyRuntimeFixtureLoad(path)
		return err
	})
	return state, err
}

func flinkLegacyRuntimeFixtureMutate(path string, mutate func(*flinkLegacyRuntimeFixtureLedger)) (flinkLegacyRuntimeFixtureLedger, error) {
	flinkLegacyRuntimeFixtureLedgerMu.Lock()
	defer flinkLegacyRuntimeFixtureLedgerMu.Unlock()
	var state flinkLegacyRuntimeFixtureLedger
	err := flinkLegacyRuntimeFixtureWithLedgerLock(path, true, func() error {
		var err error
		state, err = flinkLegacyRuntimeFixtureLoad(path)
		if err != nil {
			return err
		}
		mutate(&state)
		return flinkLegacyRuntimeFixtureSave(path, state)
	})
	return state, err
}

func flinkLegacyRuntimeFixtureWithLedgerLock(path string, exclusive bool, operation func() error) (err error) {
	lockPath := path + flinkLegacyRuntimeFixtureLedgerLockSuffix
	lockFile, err := os.OpenFile(lockPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open legacy runtime fixture ledger lock: %w", err)
	}
	defer func() {
		if closeErr := lockFile.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close legacy runtime fixture ledger lock: %w", closeErr)
		}
	}()
	openedInfo, err := lockFile.Stat()
	if err != nil {
		return fmt.Errorf("inspect opened legacy runtime fixture ledger lock: %w", err)
	}
	pathInfo, err := os.Lstat(lockPath)
	if err != nil {
		return fmt.Errorf("inspect legacy runtime fixture ledger lock path: %w", err)
	}
	if !openedInfo.Mode().IsRegular() || openedInfo.Mode().Perm() != 0o600 || !pathInfo.Mode().IsRegular() || !os.SameFile(openedInfo, pathInfo) {
		return fmt.Errorf("legacy runtime fixture ledger lock must be a stable mode 0600 regular file")
	}
	if err := flinkLegacyRuntimeFixtureLockFile(lockFile, exclusive); err != nil {
		return fmt.Errorf("lock legacy runtime fixture ledger: %w", err)
	}
	defer func() {
		if unlockErr := flinkLegacyRuntimeFixtureUnlockFile(lockFile); err == nil && unlockErr != nil {
			err = fmt.Errorf("unlock legacy runtime fixture ledger: %w", unlockErr)
		}
	}()
	return operation()
}

func flinkLegacyRuntimeFixtureLoad(path string) (flinkLegacyRuntimeFixtureLedger, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return flinkLegacyRuntimeFixtureLedger{}, fmt.Errorf("read legacy runtime fixture ledger: %w", err)
	}
	var state flinkLegacyRuntimeFixtureLedger
	if err := json.Unmarshal(body, &state); err != nil {
		return state, fmt.Errorf("decode legacy runtime fixture ledger: %w", err)
	}
	return state, nil
}

func flinkLegacyRuntimeFixtureSave(path string, state flinkLegacyRuntimeFixtureLedger) error {
	body, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode legacy runtime fixture ledger: %w", err)
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".flink-legacy-runtime-ledger-*")
	if err != nil {
		return err
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
		return err
	}
	if _, err := temp.Write(append(body, '\n')); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	committed = true
	return nil
}
