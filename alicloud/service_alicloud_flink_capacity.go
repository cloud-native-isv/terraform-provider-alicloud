package alicloud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type FlinkCapacityService struct {
	api                flinkCapacityAPI
	capacityBootstrap  *flinkCapacityBootstrapRuntime
	expectedResourceID string
	observedResourceID string
}

type flinkCapacityBootstrapRuntime struct {
	context                     flinkWorkspaceCapacityBootstrapContext
	allowInitialIdentityAbsence bool
	identitySeen                bool
	now                         func() time.Time
	pinnedResourceID            string
}

type flinkCapacityReadResult struct {
	tree                           flinkcapacity.Tree
	workspaceAuthoritativelyAbsent bool
}

type flinkCapacityReadResultAPI interface {
	readTreeResult(context.Context, string) (flinkCapacityReadResult, error)
}

const flinkNamespaceInitialCU flinkcapacity.CU = 2 // One API CU in half-CU domain units.

// FOAS SDK request builders cast CPU and MemoryGB to int32. The public schema
// accepts integral CU, so this is floor(MaxInt32/4) external CU represented in
// the domain's half-CU units. Divide before multiplying to keep the constant
// overflow-safe.
const flinkMaxSerializableComponentCU flinkcapacity.CU = flinkcapacity.CU(flinkMaxCUBeforeInt32MemoryOverflow) * 2

type flinkCapacityAPI interface {
	GetWorkspace(string) (*flink.Workspace, error)
	ListWorkspaces() ([]flink.Workspace, error)
	ListNamespaces(string) ([]flink.Namespace, error)
	ListDeploymentTargets(string, string) ([]flink.DeploymentTarget, error)
	CreateNamespace(string, *flink.Namespace) (*flink.Namespace, error)
	DeleteNamespace(string, string) error
	GetNamespace(string, string) (*flink.Namespace, error)
	UpdateNamespaceCapacity(string, string, bool, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error)
	UpdateDeploymentTargetV2(string, string, *flink.DeploymentTarget) (*flink.DeploymentTarget, error)
	ModifyPrepayWorkspaceCapacity(string, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error)
	ModifyPostpayWorkspaceCapacity(string, *flink.ResourceSpec, *flink.ResourceSpec) (flink.CapacityOperation, error)
	EnableWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error)
	ModifyWorkspaceElastic(string, *flink.ResourceSpec) (flink.CapacityOperation, error)
}

func NewFlinkCapacityService(client *connectivity.AliyunClient) (*FlinkCapacityService, error) {
	service, err := NewFlinkService(client)
	if err != nil {
		return nil, err
	}
	return &FlinkCapacityService{api: service.GetAPI()}, nil
}

func (s *FlinkCapacityService) withCapacityBootstrapContext(value flinkWorkspaceCapacityBootstrapContext, allowInitialIdentityAbsence bool, now func() time.Time) *FlinkCapacityService {
	clone := *s
	if now == nil {
		now = time.Now
	}
	clone.capacityBootstrap = &flinkCapacityBootstrapRuntime{
		context:                     value,
		allowInitialIdentityAbsence: allowInitialIdentityAbsence,
		now:                         now,
	}
	return &clone
}

func (s *FlinkCapacityService) withExpectedResourceID(resourceID string) *FlinkCapacityService {
	clone := *s
	clone.expectedResourceID = resourceID
	return &clone
}

func (s *FlinkCapacityService) ReadTree(ctx context.Context, instanceID string) (flinkcapacity.Tree, error) {
	return s.readTree(ctx, instanceID, nil)
}

func (s *FlinkCapacityService) readTreeResult(ctx context.Context, instanceID string) (flinkCapacityReadResult, error) {
	result := flinkCapacityReadResult{}
	var err error
	result.tree, err = s.readTree(ctx, instanceID, &result)
	return result, err
}

func (s *FlinkCapacityService) readTree(ctx context.Context, instanceID string, result *flinkCapacityReadResult) (flinkcapacity.Tree, error) {
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Tree{}, err
	}
	workspace, err := s.api.GetWorkspace(instanceID)
	if err != nil {
		var serviceErr *flink.FlinkServiceError
		if !errors.As(err, &serviceErr) || serviceErr.GetErrorCode() != "404" {
			return flinkcapacity.Tree{}, err
		}
		workspaces, listErr := s.api.ListWorkspaces()
		if listErr != nil {
			return flinkcapacity.Tree{}, listErr
		}
		workspace = nil
		matches := 0
		for i := range workspaces {
			if workspaces[i].Id == instanceID {
				matches++
				workspace = &workspaces[i]
			}
		}
		if matches > 1 {
			return flinkcapacity.Tree{}, fmt.Errorf("workspace identity mismatch: ListWorkspaces returned %d exact entries for InstanceId %q", matches, instanceID)
		}
		if workspace == nil {
			if s.capacityBootstrapAllowsInitialIdentityAbsence() {
				return flinkcapacity.Tree{}, &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q exact identity is not visible yet", instanceID)}
			}
			if result != nil {
				result.workspaceAuthoritativelyAbsent = true
			}
			return flinkcapacity.Tree{}, err
		}
	}
	if err := validateFlinkCapacityWorkspaceIdentity(instanceID, workspace); err != nil {
		return flinkcapacity.Tree{}, err
	}
	if err := validateFlinkCapacityWorkspaceNotTerminal(workspace); err != nil {
		return flinkcapacity.Tree{}, err
	}
	if err := s.validateAndPinFlinkCapacityWorkspaceResourceID(instanceID, workspace); err != nil {
		return flinkcapacity.Tree{}, err
	}
	if s.capacityBootstrap != nil {
		s.capacityBootstrap.identitySeen = true
		if err := validateFlinkCapacityWorkspaceBootstrapProvenance(instanceID, workspace, s.capacityBootstrap.context, true); err != nil {
			return flinkcapacity.Tree{}, err
		}
	}
	if err := validateFlinkCapacityWorkspaceReady(workspace); err != nil {
		return flinkcapacity.Tree{}, err
	}
	namespaces, err := s.api.ListNamespaces(instanceID)
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	if len(namespaces) == 0 {
		return flinkcapacity.Tree{}, &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q namespaces are not visible yet", instanceID)}
	}

	targets := make(map[string][]flink.DeploymentTarget, len(namespaces))
	for _, namespace := range namespaces {
		if err := ctx.Err(); err != nil {
			return flinkcapacity.Tree{}, err
		}
		if err := validateFlinkCapacityNamespaceReady(namespace); err != nil {
			return flinkcapacity.Tree{}, err
		}
		queues, err := s.api.ListDeploymentTargets(workspace.ResourceId, namespace.Name)
		if err != nil {
			return flinkcapacity.Tree{}, err
		}
		if len(queues) == 0 {
			return flinkcapacity.Tree{}, &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("namespace %q queues are not visible yet", namespace.Name)}
		}
		for _, queue := range queues {
			if queue.Quota == nil || queue.Quota.Request == nil || queue.Quota.Limit == nil {
				return flinkcapacity.Tree{}, &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("queue %q/%q capacity quotas are not visible yet", namespace.Name, queue.Name)}
			}
		}
		targets[namespace.Name] = queues
	}
	return buildFlinkCapacityTree(workspace, namespaces, targets)
}

func (s *FlinkCapacityService) validateAndPinFlinkCapacityWorkspaceResourceID(instanceID string, workspace *flink.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("workspace %q ResourceId cannot be verified from a nil Workspace", instanceID)
	}
	if workspace.ResourceId == "" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q bootstrap ResourceId is not visible yet", instanceID)}
	}
	if s.expectedResourceID != "" && workspace.ResourceId != s.expectedResourceID {
		return fmt.Errorf("workspace %q ResourceId mismatch: observed %q, expected pinned %q", instanceID, workspace.ResourceId, s.expectedResourceID)
	}
	if s.capacityBootstrap != nil {
		if s.capacityBootstrap.pinnedResourceID == "" {
			s.capacityBootstrap.pinnedResourceID = workspace.ResourceId
		} else if workspace.ResourceId != s.capacityBootstrap.pinnedResourceID {
			return fmt.Errorf("workspace %q ResourceId changed during capacity bootstrap: observed %q, first observed %q", instanceID, workspace.ResourceId, s.capacityBootstrap.pinnedResourceID)
		}
	}
	s.observedResourceID = workspace.ResourceId
	return nil
}

func (s *FlinkCapacityService) capacityBootstrapAllowsInitialIdentityAbsence() bool {
	if s.capacityBootstrap == nil || !s.capacityBootstrap.allowInitialIdentityAbsence || s.capacityBootstrap.identitySeen {
		return false
	}
	return s.capacityBootstrap.now().Unix() < s.capacityBootstrap.context.IdentityAbsenceRetryNotAfterUnix
}

func validateFlinkCapacityWorkspaceBootstrapProvenance(instanceID string, workspace *flink.Workspace, expected flinkWorkspaceCapacityBootstrapContext, missingRetryable bool) error {
	if expected.ExpectedInstanceID != instanceID {
		return fmt.Errorf("workspace bootstrap InstanceId mismatch: allocation requested %q but context expects %q", instanceID, expected.ExpectedInstanceID)
	}
	if workspace == nil {
		return fmt.Errorf("workspace bootstrap provenance cannot be verified from a nil Workspace")
	}
	if workspace.ResourceId == "" {
		if missingRetryable {
			return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q bootstrap ResourceId is not visible yet", instanceID)}
		}
		return fmt.Errorf("workspace %q bootstrap ResourceId is missing", instanceID)
	}
	if expected.ExpectedResourceID != "" && workspace.ResourceId != expected.ExpectedResourceID {
		return fmt.Errorf("workspace %q bootstrap ResourceId mismatch: observed %q, expected %q", instanceID, workspace.ResourceId, expected.ExpectedResourceID)
	}
	if err := validateFlinkCapacityWorkspaceBootstrapTag(instanceID, workspace.Tags, flinkworkspace.CreateTokenTagKey, expected.TerraformCreateToken, "create token", missingRetryable); err != nil {
		return err
	}
	return validateFlinkCapacityWorkspaceBootstrapTag(instanceID, workspace.Tags, flinkworkspace.CreateIntentTagKey, expected.CreateIntentFingerprint, "intent fingerprint", missingRetryable)
}

func validateFlinkCapacityWorkspaceBootstrapTag(instanceID string, tags []flink.Tag, key, expected, label string, missingRetryable bool) error {
	observed, found, err := optionalUniqueFlinkWorkspaceTagValue(tags, key)
	if err != nil {
		return err
	}
	if !found || observed == "" {
		if missingRetryable {
			return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q bootstrap tag %s is not visible yet", instanceID, key)}
		}
		return fmt.Errorf("workspace %q bootstrap tag %s is missing", instanceID, key)
	}
	if observed != expected {
		return fmt.Errorf("workspace %q bootstrap %s mismatch for tag %s", instanceID, label, key)
	}
	return nil
}

func validateFlinkCapacityWorkspaceIdentity(instanceID string, workspace *flink.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("workspace identity mismatch: requested InstanceId %q, observed Workspace <nil>", instanceID)
	}
	if workspace.Id == "" || workspace.Id != instanceID {
		return fmt.Errorf("workspace identity mismatch: requested InstanceId %q, observed Workspace Id %q", instanceID, workspace.Id)
	}
	return nil
}

func validateFlinkCapacityWorkspaceReady(workspace *flink.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("workspace response is nil")
	}
	if err := validateFlinkCapacityWorkspaceNotTerminal(workspace); err != nil {
		return err
	}
	if workspace.Status != "RUNNING" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q status is %q, waiting for RUNNING", workspace.Id, workspace.Status)}
	}
	if workspace.OrderState != "NORMAL" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q order state is %q, waiting for NORMAL", workspace.Id, workspace.OrderState)}
	}
	if workspace.ResourceId == "" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q does not expose a ResourceId yet", workspace.Id)}
	}
	if workspace.Elastic && workspace.ElasticOrderState == "" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q elastic order state is not visible yet", workspace.Id)}
	}
	if workspace.ElasticOrderState != "" && workspace.ElasticOrderState != "NORMAL" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q elastic order state is %q, waiting for NORMAL", workspace.Id, workspace.ElasticOrderState)}
	}
	elasticCPU := flinkSpecCPU(workspace.ElasticResourceSpec)
	if !workspace.Elastic && elasticCPU > 0 {
		return fmt.Errorf("workspace %q returned Elastic=false with positive elastic capacity %v CU", workspace.Id, elasticCPU)
	}
	if workspace.Elastic {
		if workspace.ElasticInstanceId == "" {
			return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q elastic instance ID is not visible yet", workspace.Id)}
		}
		if elasticCPU <= 0 {
			return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q elastic capacity is not visible yet", workspace.Id)}
		}
	}
	return nil
}

func validateFlinkCapacityWorkspaceNotTerminal(workspace *flink.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("workspace response is nil")
	}
	for _, state := range []struct {
		label string
		value string
	}{
		{label: "", value: workspace.Status},
		{label: " order", value: workspace.OrderState},
		{label: " elastic order", value: workspace.ElasticOrderState},
	} {
		if flinkTerminalState(state.value) {
			return fmt.Errorf("workspace %q%s is in terminal state %q", workspace.Id, state.label, state.value)
		}
	}
	return nil
}

func validateFlinkCapacityNamespaceReady(namespace flink.Namespace) error {
	if namespace.Status == "SUCCESS" || namespace.Status == "Available" {
		return nil
	}
	if flinkTerminalState(namespace.Status) {
		return fmt.Errorf("namespace %q is in terminal state %q", namespace.Name, namespace.Status)
	}
	return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("namespace %q status is %q, waiting for SUCCESS or Available", namespace.Name, namespace.Status)}
}

func flinkTerminalState(state string) bool {
	upper := strings.ToUpper(strings.TrimSpace(state))
	if strings.Contains(upper, "FAIL") {
		return true
	}
	switch upper {
	case "DISABLE", "DISABLED", "DELETING", "DELETED", "CEASE", "CEASED", "RELEASE", "RELEASED", "RELEASING":
		return true
	default:
		return false
	}
}

func (s *FlinkCapacityService) ApplyStep(ctx context.Context, instanceID string, step flinkcapacity.Step) (flinkcapacity.Operation, error) {
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Operation{}, err
	}
	if err := validateFlinkCapacityStepSerializable(step); err != nil {
		return flinkcapacity.Operation{}, err
	}
	workspace, err := s.verifyWorkspaceBeforeCapacityWrite(ctx, instanceID)
	if err != nil {
		return flinkcapacity.Operation{}, err
	}
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Operation{}, err
	}
	var operation flink.CapacityOperation
	err = nil
	switch step.Action {
	case flinkcapacity.ModifyWorkspaceFixed:
		operation, err = s.api.ModifyPrepayWorkspaceCapacity(
			instanceID,
			flinkResourceSpecForCU(step.To.FixedCU),
			flinkResourceSpecForOptionalCU(step.To.CrossZoneFixedCU),
		)
	case flinkcapacity.ModifyWorkspacePostpaid:
		operation, err = s.api.ModifyPostpayWorkspaceCapacity(
			instanceID,
			flinkResourceSpecForCU(step.To.Limit),
			nil,
		)
	case flinkcapacity.EnableWorkspaceElastic:
		operation, err = s.api.EnableWorkspaceElastic(instanceID, flinkResourceSpecForCU(step.To.AsCapacity().Elastic()))
	case flinkcapacity.ModifyWorkspaceElastic:
		operation, err = s.api.ModifyWorkspaceElastic(instanceID, flinkResourceSpecForCU(step.To.AsCapacity().Elastic()))
	case flinkcapacity.CreateNamespace:
		namespace := flink.Namespace{Name: step.Ref.Namespace, Ha: step.NamespaceCrossZone}
		request := namespace
		request.Id = instanceID
		request.ResourceSpec = flinkResourceSpecForCU(flinkNamespaceInitialCU)
		_, err = s.api.CreateNamespace(instanceID, &request)
		var postCreateReadErr *flink.FlinkPostCreateReadError
		if errors.As(err, &postCreateReadErr) {
			err = &ambiguousFlinkCapacityWriteError{cause: err}
		}
	case flinkcapacity.DeleteNamespace:
		err = s.api.DeleteNamespace(instanceID, step.Ref.Namespace)
		if err != nil && flinkCapacityNotFoundError(err) {
			err = &ambiguousFlinkCapacityWriteError{cause: err}
		}
	case flinkcapacity.ModifyNamespace:
		namespace, getErr := s.api.GetNamespace(instanceID, step.Ref.Namespace)
		if getErr != nil {
			err = getErr
			break
		}
		if contextErr := ctx.Err(); contextErr != nil {
			err = contextErr
			break
		}
		operation, err = s.api.UpdateNamespaceCapacity(
			instanceID,
			step.Ref.Namespace,
			namespace.Ha,
			flinkResourceSpecForCU(step.To.FixedCU),
			flinkResourceSpecForCU(step.To.Limit-step.To.FixedCU),
		)
	case flinkcapacity.ModifyQueue:
		_, err = s.api.UpdateDeploymentTargetV2(workspace.ResourceId, step.Ref.Namespace, &flink.DeploymentTarget{
			Name:      step.Ref.Queue,
			Namespace: step.Ref.Namespace,
			Quota: &flink.ResourceQuota{
				Request: flinkResourceSpecForCU(step.To.FixedCU),
				Limit:   flinkResourceSpecForCU(step.To.Limit),
			},
		})
	default:
		err = fmt.Errorf("unsupported Flink capacity action %q", step.Action)
	}
	return flinkcapacity.Operation{RequestID: operation.RequestID, OrderID: operation.OrderID}, classifyFlinkCapacityWriteError(err)
}

func (s *FlinkCapacityService) verifyWorkspaceBeforeCapacityWrite(ctx context.Context, instanceID string) (*flink.Workspace, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	workspace, err := s.api.GetWorkspace(instanceID)
	if err != nil {
		return nil, err
	}
	if err := validateFlinkCapacityWorkspaceIdentity(instanceID, workspace); err != nil {
		return nil, err
	}
	if err := s.validateAndPinFlinkCapacityWorkspaceResourceID(instanceID, workspace); err != nil {
		return nil, err
	}
	if s.capacityBootstrap != nil {
		if err := validateFlinkCapacityWorkspaceBootstrapProvenance(instanceID, workspace, s.capacityBootstrap.context, false); err != nil {
			return nil, err
		}
	}
	if err := validateFlinkCapacityWorkspaceReady(workspace); err != nil {
		return nil, err
	}
	return workspace, nil
}

func validateFlinkCapacityStepSerializable(step flinkcapacity.Step) error {
	type component struct {
		name  string
		value flinkcapacity.CU
	}
	validate := func(component string, value flinkcapacity.CU) error {
		if value < 0 || value > flinkMaxSerializableComponentCU {
			return fmt.Errorf(
				"%s %s CU %v is outside the FOAS int32 CPU/MemoryGB serialization range 0..%d",
				step.Action, component, value.Float64(), flinkMaxCUBeforeInt32MemoryOverflow,
			)
		}
		return nil
	}
	validateAll := func(components ...component) error {
		for _, component := range components {
			if err := validate(component.name, component.value); err != nil {
				return err
			}
		}
		return nil
	}
	validateFOAS := func(component string, value flinkcapacity.CU) error {
		if value%2 != 0 {
			return fmt.Errorf(
				"%s %s CU %v must be an integer CU for FOAS Workspace/Namespace serialization",
				step.Action, component, value.Float64(),
			)
		}
		return validate(component, value)
	}
	validateAllFOAS := func(components ...component) error {
		for _, component := range components {
			if err := validateFOAS(component.name, component.value); err != nil {
				return err
			}
		}
		return nil
	}

	switch step.Action {
	case flinkcapacity.ModifyWorkspaceFixed:
		return validateAllFOAS(
			component{"fixed", step.To.FixedCU},
			component{"cross-zone fixed", step.To.CrossZoneFixedCU},
		)
	case flinkcapacity.ModifyWorkspacePostpaid:
		return validateFOAS("limit", step.To.Limit)
	case flinkcapacity.EnableWorkspaceElastic, flinkcapacity.ModifyWorkspaceElastic:
		elastic, err := flinkCapacityAllocationElasticCU(step.To)
		if err != nil {
			return fmt.Errorf("%s allocation: %w", step.Action, err)
		}
		return validateFOAS("elastic", elastic)
	case flinkcapacity.CreateNamespace:
		return validateFOAS("initial fixed", flinkNamespaceInitialCU)
	case flinkcapacity.ModifyNamespace:
		if step.To.FixedCU < 0 || step.To.Limit < step.To.FixedCU {
			return fmt.Errorf("%s allocation has invalid fixed/max components", step.Action)
		}
		return validateAllFOAS(
			component{"fixed", step.To.FixedCU},
			component{"elastic", step.To.Limit - step.To.FixedCU},
		)
	case flinkcapacity.ModifyQueue:
		return validateAll(
			component{"request", step.To.FixedCU},
			component{"limit", step.To.Limit},
		)
	default:
		return nil
	}
}

func flinkCapacityAllocationElasticCU(allocation flinkcapacity.Allocation) (flinkcapacity.CU, error) {
	if allocation.FixedCU < 0 || allocation.CrossZoneFixedCU < 0 || allocation.Limit < 0 {
		return 0, fmt.Errorf("fixed, cross-zone fixed, and limit CU must be non-negative")
	}
	if allocation.FixedCU > flinkcapacity.CU(math.MaxInt64)-allocation.CrossZoneFixedCU {
		return 0, fmt.Errorf("fixed CU total overflows")
	}
	totalFixed := allocation.FixedCU + allocation.CrossZoneFixedCU
	if allocation.Limit < totalFixed {
		return 0, fmt.Errorf("limit CU %v is below fixed CU %v", allocation.Limit.Float64(), totalFixed.Float64())
	}
	return allocation.Limit - totalFixed, nil
}

type ambiguousFlinkCapacityWriteError struct {
	cause error
}

func (e *ambiguousFlinkCapacityWriteError) Error() string   { return e.cause.Error() }
func (e *ambiguousFlinkCapacityWriteError) Unwrap() error   { return e.cause }
func (e *ambiguousFlinkCapacityWriteError) Ambiguous() bool { return true }

func classifyFlinkCapacityWriteError(err error) error {
	if err == nil {
		return nil
	}
	var sdkErr *flink.FlinkSDKError
	if errors.As(err, &sdkErr) {
		return &ambiguousFlinkCapacityWriteError{cause: err}
	}
	return err
}

func flinkCapacityNotFoundError(err error) bool {
	if NotFoundError(err) {
		return true
	}
	var serviceErr *flink.FlinkServiceError
	return errors.As(err, &serviceErr) && serviceErr.GetErrorCode() == "404"
}

func buildFlinkCapacityTree(workspace *flink.Workspace, namespaces []flink.Namespace, targets map[string][]flink.DeploymentTarget) (flinkcapacity.Tree, error) {
	if workspace == nil {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace must not be nil")
	}
	tree := flinkcapacity.Tree{
		ChargeType:          workspace.ChargeType,
		WorkspaceResourceID: workspace.ResourceId,
		Workspace: flinkcapacity.WorkspaceCapacity{
			HA: workspace.Ha || (workspace.HighAvailability != nil && workspace.HighAvailability.Enabled),
		},
	}
	if workspace.ChargeType == "POST" {
		limit, err := cuFromResourceSpec(workspace.ResourceSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("POST workspace limit: %w", err)
		}
		tree.Workspace.Limit = limit
	} else {
		fixed, err := cuFromOptionalResourceSpec(workspace.ResourceSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("workspace fixed capacity: %w", err)
		}
		crossZone, err := cuFromOptionalResourceSpec(workspace.HaResourceSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("workspace cross-zone fixed capacity: %w", err)
		}
		elastic, err := cuFromOptionalResourceSpec(workspace.ElasticResourceSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("workspace elastic capacity: %w", err)
		}
		tree.Workspace = flinkcapacity.WorkspaceCapacity{
			HA:               tree.Workspace.HA,
			FixedCU:          fixed,
			CrossZoneFixedCU: crossZone,
			Limit:            fixed + crossZone + elastic,
		}
	}
	if workspace.ClusterUsedResources != nil {
		used, err := usedCUFromFloat(workspace.ClusterUsedResources.UsedResource)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("workspace used capacity: %w", err)
		}
		tree.Workspace.Used = used
	}

	tree.Namespaces = make([]flinkcapacity.Namespace, 0, len(namespaces))
	for _, namespace := range namespaces {
		fixedSpec := namespace.GuaranteedResourceSpec
		elasticSpec := namespace.ElasticResourceSpec
		if fixedSpec == nil && elasticSpec == nil && namespace.ResourceSpec != nil {
			fixedSpec = namespace.ResourceSpec
		}
		fixed, err := cuFromOptionalResourceSpec(fixedSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q fixed capacity: %w", namespace.Name, err)
		}
		elastic, err := cuFromOptionalResourceSpec(elasticSpec)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q elastic capacity: %w", namespace.Name, err)
		}
		capacity := flinkcapacity.Capacity{Fixed: fixed, Limit: fixed + elastic}
		used, err := cuFromResourceUsed(namespace.ResourceUsed)
		if err != nil {
			return flinkcapacity.Tree{}, fmt.Errorf("namespace %q used capacity: %w", namespace.Name, err)
		}
		domainNamespace := flinkcapacity.Namespace{Name: namespace.Name, CrossZone: namespace.Ha, Capacity: &capacity, Used: used}

		for _, target := range targets[namespace.Name] {
			if target.Quota == nil || target.Quota.Request == nil || target.Quota.Limit == nil {
				return flinkcapacity.Tree{}, fmt.Errorf("queue %q/%q does not expose request and limit quotas", namespace.Name, target.Name)
			}
			queueFixed, err := cuFromResourceSpec(target.Quota.Request)
			if err != nil {
				return flinkcapacity.Tree{}, fmt.Errorf("queue %q/%q fixed capacity: %w", namespace.Name, target.Name, err)
			}
			queueLimit, err := cuFromResourceSpec(target.Quota.Limit)
			if err != nil {
				return flinkcapacity.Tree{}, fmt.Errorf("queue %q/%q limit: %w", namespace.Name, target.Name, err)
			}
			queueUsed, err := usedCUFromResourceSpec(target.Quota.Used)
			if err != nil {
				return flinkcapacity.Tree{}, fmt.Errorf("queue %q/%q used capacity: %w", namespace.Name, target.Name, err)
			}
			queueCapacity := flinkcapacity.Capacity{Fixed: queueFixed, Limit: queueLimit}
			domainNamespace.Queues = append(domainNamespace.Queues, flinkcapacity.Queue{Name: target.Name, Capacity: &queueCapacity, Used: queueUsed})
		}
		tree.Namespaces = append(tree.Namespaces, domainNamespace)
	}
	return flinkcapacity.Resolve(tree)
}

func flinkResourceSpecForCU(cu flinkcapacity.CU) *flink.ResourceSpec {
	value := cu.Float64()
	return &flink.ResourceSpec{Cpu: value, MemoryGB: value * 4}
}

func flinkResourceSpecForOptionalCU(cu flinkcapacity.CU) *flink.ResourceSpec {
	if cu == 0 {
		return nil
	}
	return flinkResourceSpecForCU(cu)
}

func cuFromResourceSpec(spec *flink.ResourceSpec) (flinkcapacity.CU, error) {
	if spec == nil {
		return 0, fmt.Errorf("resource specification is missing")
	}
	if math.Abs(spec.MemoryGB-spec.Cpu*4) > 1e-9 {
		return 0, fmt.Errorf("resource specification memory %v must equal CPU %v * 4", spec.MemoryGB, spec.Cpu)
	}
	return flinkcapacity.ParseCU(spec.Cpu)
}

func cuFromOptionalResourceSpec(spec *flink.ResourceSpec) (flinkcapacity.CU, error) {
	if spec == nil {
		return 0, nil
	}
	return cuFromResourceSpec(spec)
}

func cuFromResourceUsed(used *flink.ResourceUsed) (float64, error) {
	if used == nil {
		return 0, nil
	}
	return usedCUFromFloat(used.Cu)
}

func usedCUFromResourceSpec(spec *flink.ResourceSpec) (float64, error) {
	if spec == nil {
		return 0, nil
	}
	return usedCUFromFloat(spec.Cpu)
}

func usedCUFromFloat(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, fmt.Errorf("used CU must be a finite non-negative value, got %v", value)
	}
	return value, nil
}

var _ flinkcapacity.API = (*FlinkCapacityService)(nil)
