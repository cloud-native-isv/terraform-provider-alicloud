package alicloud

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkcapacity"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

type FlinkCapacityService struct {
	flinkService *FlinkService
}

func NewFlinkCapacityService(client *connectivity.AliyunClient) (*FlinkCapacityService, error) {
	service, err := NewFlinkService(client)
	if err != nil {
		return nil, err
	}
	return &FlinkCapacityService{flinkService: service}, nil
}

func (s *FlinkCapacityService) ReadTree(ctx context.Context, instanceID string) (flinkcapacity.Tree, error) {
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Tree{}, err
	}
	workspace, err := s.flinkService.GetAPI().GetWorkspace(instanceID)
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	if err := validateFlinkCapacityWorkspaceReady(workspace); err != nil {
		return flinkcapacity.Tree{}, err
	}
	namespaces, err := s.flinkService.GetAPI().ListNamespaces(instanceID)
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
		queues, err := s.flinkService.GetAPI().ListDeploymentTargets(workspace.ResourceId, namespace.Name)
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

func validateFlinkCapacityWorkspaceReady(workspace *flink.Workspace) error {
	if workspace == nil {
		return fmt.Errorf("workspace response is nil")
	}
	if workspace.Status != "RUNNING" {
		if flinkTerminalState(workspace.Status) {
			return fmt.Errorf("workspace %q is in terminal state %q", workspace.Id, workspace.Status)
		}
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q status is %q, waiting for RUNNING", workspace.Id, workspace.Status)}
	}
	if workspace.OrderState != "NORMAL" {
		if flinkTerminalState(workspace.OrderState) {
			return fmt.Errorf("workspace %q order is in terminal state %q", workspace.Id, workspace.OrderState)
		}
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q order state is %q, waiting for NORMAL", workspace.Id, workspace.OrderState)}
	}
	if workspace.ResourceId == "" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q does not expose a ResourceId yet", workspace.Id)}
	}
	if workspace.Elastic && workspace.ElasticOrderState == "" {
		return &flinkcapacity.NotReadyError{Reason: fmt.Sprintf("workspace %q elastic order state is not visible yet", workspace.Id)}
	}
	if workspace.ElasticOrderState != "" && workspace.ElasticOrderState != "NORMAL" {
		if flinkTerminalState(workspace.ElasticOrderState) {
			return fmt.Errorf("workspace %q elastic order is in terminal state %q", workspace.Id, workspace.ElasticOrderState)
		}
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
	upper := strings.ToUpper(state)
	return strings.Contains(upper, "FAIL") || upper == "DISABLE" || upper == "DELETED"
}

func (s *FlinkCapacityService) ApplyStep(ctx context.Context, instanceID string, step flinkcapacity.Step) (flinkcapacity.Operation, error) {
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Operation{}, err
	}
	var operation flink.CapacityOperation
	var err error
	switch step.Action {
	case flinkcapacity.ModifyWorkspaceFixed:
		operation, err = s.flinkService.GetAPI().ModifyPrepayWorkspaceCapacity(
			instanceID,
			flinkResourceSpecForCU(step.To.FixedCU),
			flinkResourceSpecForOptionalCU(step.To.CrossZoneFixedCU),
		)
	case flinkcapacity.ModifyWorkspacePostpaid:
		operation, err = s.flinkService.GetAPI().ModifyPostpayWorkspaceCapacity(
			instanceID,
			flinkResourceSpecForCU(step.To.Limit),
			nil,
		)
	case flinkcapacity.EnableWorkspaceElastic:
		operation, err = s.flinkService.GetAPI().EnableWorkspaceElastic(instanceID, flinkResourceSpecForCU(step.To.AsCapacity().Elastic()))
	case flinkcapacity.ModifyWorkspaceElastic:
		operation, err = s.flinkService.GetAPI().ModifyWorkspaceElastic(instanceID, flinkResourceSpecForCU(step.To.AsCapacity().Elastic()))
	case flinkcapacity.ModifyNamespace:
		namespace, getErr := s.flinkService.GetAPI().GetNamespace(instanceID, step.Ref.Namespace)
		if getErr != nil {
			err = getErr
			break
		}
		operation, err = s.flinkService.GetAPI().UpdateNamespaceCapacity(
			instanceID,
			step.Ref.Namespace,
			namespace.Ha,
			flinkResourceSpecForCU(step.To.FixedCU),
			flinkResourceSpecForCU(step.To.Limit-step.To.FixedCU),
		)
	case flinkcapacity.ModifyQueue:
		workspace, getErr := s.flinkService.GetAPI().GetWorkspace(instanceID)
		if getErr != nil {
			err = getErr
			break
		}
		if workspace.ResourceId == "" {
			err = fmt.Errorf("workspace %q does not expose a ResourceId", instanceID)
			break
		}
		_, err = s.flinkService.GetAPI().UpdateDeploymentTargetV2(workspace.ResourceId, step.Ref.Namespace, &flink.DeploymentTarget{
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
		domainNamespace := flinkcapacity.Namespace{Name: namespace.Name, Capacity: &capacity, Used: used}

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
