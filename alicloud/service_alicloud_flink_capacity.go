package alicloud

import (
	"context"
	"fmt"

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
	namespaces, err := s.flinkService.GetAPI().ListNamespaces(instanceID)
	if err != nil {
		return flinkcapacity.Tree{}, err
	}
	if workspace.ResourceId == "" {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace %q does not expose a ResourceId", instanceID)
	}

	targets := make(map[string][]flink.DeploymentTarget, len(namespaces))
	for _, namespace := range namespaces {
		if err := ctx.Err(); err != nil {
			return flinkcapacity.Tree{}, err
		}
		queues, err := s.flinkService.GetAPI().ListDeploymentTargets(workspace.ResourceId, namespace.Name)
		if err != nil {
			return flinkcapacity.Tree{}, err
		}
		targets[namespace.Name] = queues
	}
	return buildFlinkCapacityTree(workspace, namespaces, targets)
}

func (s *FlinkCapacityService) ApplyStep(ctx context.Context, instanceID string, step flinkcapacity.Step) (flinkcapacity.Operation, error) {
	if err := ctx.Err(); err != nil {
		return flinkcapacity.Operation{}, err
	}
	workspace, err := s.flinkService.GetAPI().GetWorkspace(instanceID)
	if err != nil {
		return flinkcapacity.Operation{}, err
	}

	var operation flink.CapacityOperation
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
	return flinkcapacity.Operation{RequestID: operation.RequestID, OrderID: operation.OrderID}, err
}

func buildFlinkCapacityTree(workspace *flink.Workspace, namespaces []flink.Namespace, targets map[string][]flink.DeploymentTarget) (flinkcapacity.Tree, error) {
	if workspace == nil {
		return flinkcapacity.Tree{}, fmt.Errorf("workspace must not be nil")
	}
	tree := flinkcapacity.Tree{ChargeType: workspace.ChargeType}
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
			FixedCU:          fixed,
			CrossZoneFixedCU: crossZone,
			Limit:            fixed + crossZone + elastic,
		}
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
			queueUsed, err := cuFromOptionalResourceSpec(target.Quota.Used)
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
	return flinkcapacity.ParseCU(spec.Cpu)
}

func cuFromOptionalResourceSpec(spec *flink.ResourceSpec) (flinkcapacity.CU, error) {
	if spec == nil {
		return 0, nil
	}
	return cuFromResourceSpec(spec)
}

func cuFromResourceUsed(used *flink.ResourceUsed) (flinkcapacity.CU, error) {
	if used == nil {
		return 0, nil
	}
	return flinkcapacity.ParseCU(used.Cu)
}

var _ flinkcapacity.API = (*FlinkCapacityService)(nil)
