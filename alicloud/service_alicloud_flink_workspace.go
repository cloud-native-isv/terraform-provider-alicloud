package alicloud

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Workspace methods
func (s *FlinkService) DescribeFlinkWorkspace(id string) (*aliyunFlinkAPI.Workspace, error) {
	return s.GetAPI().GetWorkspace(id)
}

func (s *FlinkService) CreateInstance(workspace *aliyunFlinkAPI.Workspace, options flinkworkspace.CreateOptions) (*aliyunFlinkAPI.Workspace, error) {
	result, err := s.GetAPI().CreateWorkspaceWithOptions(workspace, options)
	if err != nil {
		return nil, classifyFlinkWorkspaceCreateError(err)
	}
	return result, nil
}

type ambiguousFlinkWorkspaceCreateError struct {
	cause error
}

func (e *ambiguousFlinkWorkspaceCreateError) Error() string   { return e.cause.Error() }
func (e *ambiguousFlinkWorkspaceCreateError) Unwrap() error   { return e.cause }
func (e *ambiguousFlinkWorkspaceCreateError) Ambiguous() bool { return true }

func isAmbiguousFlinkWorkspaceCreateError(err error) bool {
	var ambiguous interface{ Ambiguous() bool }
	return errors.As(err, &ambiguous) && ambiguous.Ambiguous()
}

func classifyFlinkWorkspaceCreateError(err error) error {
	if err == nil {
		return nil
	}
	// A structured non-retryable service response proves that the purchase was
	// rejected. Unstructured transport errors (including bare EOF) do not prove
	// that, so treat them as response-loss candidates and recover by Tag.
	var serviceErr *tea.SDKError
	if errors.As(err, &serviceErr) && !NeedRetry(err) {
		return err
	}
	var flinkServiceErr *aliyunFlinkAPI.FlinkServiceError
	if errors.As(err, &flinkServiceErr) && !flinkServiceErr.IsRetryableError() {
		return err
	}
	return &ambiguousFlinkWorkspaceCreateError{cause: err}
}

func (s *FlinkService) DeleteInstance(id string) error {
	return s.GetAPI().DeleteWorkspace(id)
}

type flinkRefundClient interface {
	IsInternationalAccount() bool
	RpcPostWithEndpoint(string, string, string, map[string]interface{}, map[string]interface{}, bool, string) (map[string]interface{}, error)
}

func (s *FlinkService) RefundInstance(id string) error {
	return refundFlinkWorkspaceForInstance(s.client, s.client.RegionId, id)
}

func refundFlinkWorkspaceForInstance(client flinkRefundClient, regionID, instanceID string) error {
	return refundFlinkWorkspace(client, regionID, instanceID, flinkworkspace.RefundClientToken(regionID, instanceID))
}

func refundFlinkWorkspace(client flinkRefundClient, regionID, instanceID, clientToken string) error {
	request := flinkworkspace.BuildRefundRequest(instanceID, client.IsInternationalAccount(), clientToken)
	_, err := client.RpcPostWithEndpoint("BssOpenApi", "2017-12-14", "RefundInstance", nil, request, true, "")
	if err == nil {
		return nil
	}
	return fmt.Errorf(
		"refund prepaid Flink workspace failed: Region=%s, WorkspaceID=%s, ProductCode=%s, ProductType=%s; the Terraform state is retained; if this product does not expose public RefundInstance, complete a non-full refund in the console and refresh state: %w",
		regionID,
		instanceID,
		flinkworkspace.RefundProductCodeDomestic,
		flinkworkspace.RefundProductType(client.IsInternationalAccount()),
		err,
	)
}

var _ flinkRefundClient = (*connectivity.AliyunClient)(nil)

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspace, err := s.GetAPI().GetWorkspace(id)
		if err != nil {
			// Handle the case where workspace is temporarily not found after creation
			// This is common with cloud resources that have async creation processes
			if NotFoundError(err) { // Use generic NotFoundError instead of specific error code
				// Return empty state to indicate the resource is still being created
				return nil, aliyunFlinkAPI.FlinkWorkspaceStatusCreating.String(), nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetWorkspace", AlibabaCloudSdkGoERROR)
		}
		return workspace, workspace.Status, nil
	}
}

// WaitForWorkspaceStarting waits for a Flink workspace to reach running state after creation
func (s *FlinkService) WaitForWorkspaceStarting(id string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending:    aliyunFlinkAPI.FlinkWorkspaceStatusesToStrings([]aliyunFlinkAPI.FlinkWorkspaceStatus{aliyunFlinkAPI.FlinkWorkspaceStatusCreating}),
		Target:     aliyunFlinkAPI.FlinkWorkspaceStatusesToStrings([]aliyunFlinkAPI.FlinkWorkspaceStatus{aliyunFlinkAPI.FlinkWorkspaceStatusRunning}),
		Refresh:    s.FlinkWorkspaceStateRefreshFunc(id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

type flinkLegacyNamespaceReadinessService interface {
	DescribeFlinkWorkspace(string) (*aliyunFlinkAPI.Workspace, error)
	ListNamespaces(string) ([]aliyunFlinkAPI.Namespace, error)
	ListFlinkDeploymentTargets(string, string) ([]aliyunFlinkAPI.DeploymentTarget, error)
}

type flinkLegacyNamespaceNotReadyError struct {
	reason string
}

func (e *flinkLegacyNamespaceNotReadyError) Error() string { return e.reason }

type flinkLegacyNamespaceReadinessSelectionMode uint8

const (
	flinkLegacyNamespaceReadinessSelectionInvalid flinkLegacyNamespaceReadinessSelectionMode = iota
	flinkLegacyNamespaceReadinessBootstrap
	flinkLegacyNamespaceReadinessExact
	flinkLegacyNamespaceReadinessBootstrapThenFullSnapshot
)

type flinkLegacyNamespaceReadinessSelection struct {
	mode          flinkLegacyNamespaceReadinessSelectionMode
	requiredNames map[string]struct{}
}

func flinkLegacyBootstrapNamespaceSelection() flinkLegacyNamespaceReadinessSelection {
	return flinkLegacyNamespaceReadinessSelection{mode: flinkLegacyNamespaceReadinessBootstrap}
}

func flinkLegacyExactNamespaceSelection(requiredNames map[string]struct{}) flinkLegacyNamespaceReadinessSelection {
	return flinkLegacyNamespaceReadinessSelection{
		mode:          flinkLegacyNamespaceReadinessExact,
		requiredNames: requiredNames,
	}
}

func flinkLegacyUnfilteredNamespaceSelection() flinkLegacyNamespaceReadinessSelection {
	return flinkLegacyNamespaceReadinessSelection{mode: flinkLegacyNamespaceReadinessBootstrapThenFullSnapshot}
}

func (s *FlinkService) ListFlinkDeploymentTargets(resourceID, namespace string) ([]aliyunFlinkAPI.DeploymentTarget, error) {
	return s.GetAPI().ListDeploymentTargets(resourceID, namespace)
}

func waitForFlinkLegacyNamespaceReadiness(
	ctx context.Context,
	service flinkLegacyNamespaceReadinessService,
	workspaceID string,
	selection flinkLegacyNamespaceReadinessSelection,
	timeout time.Duration,
	pollInterval time.Duration,
) ([]aliyunFlinkAPI.Namespace, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("Flink workspace %q legacy namespace readiness timeout must be positive", workspaceID)
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("wait for Flink workspace %q legacy namespace readiness: %w", workspaceID, err)
		}

		namespaces, err := observeFlinkLegacyNamespaceReadiness(service, workspaceID, selection)
		if err == nil {
			return namespaces, nil
		}
		var notReady *flinkLegacyNamespaceNotReadyError
		if !errors.As(err, &notReady) && !flinkLegacyNamespaceReadinessRetryableError(err) {
			return nil, err
		}
		lastErr = err

		poll := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			stopFlinkLegacyNamespaceReadinessTimer(poll)
			return nil, fmt.Errorf("wait for Flink workspace %q legacy namespace readiness: %w", workspaceID, ctx.Err())
		case <-deadline.C:
			stopFlinkLegacyNamespaceReadinessTimer(poll)
			return nil, fmt.Errorf("timed out waiting for Flink workspace %q legacy namespace readiness: %w", workspaceID, lastErr)
		case <-poll.C:
		}
	}
}

func stopFlinkLegacyNamespaceReadinessTimer(timer *time.Timer) {
	if timer.Stop() {
		return
	}
	select {
	case <-timer.C:
	default:
	}
}

func observeFlinkLegacyNamespaceReadiness(
	service flinkLegacyNamespaceReadinessService,
	workspaceID string,
	selection flinkLegacyNamespaceReadinessSelection,
) ([]aliyunFlinkAPI.Namespace, error) {
	workspace, err := service.DescribeFlinkWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("DescribeFlinkWorkspace(%q) returned nil", workspaceID)
	}
	if workspace.Id == "" || workspace.Id != workspaceID {
		return nil, fmt.Errorf("DescribeFlinkWorkspace(%q) returned identity %q while waiting for legacy namespace readiness", workspaceID, workspace.Id)
	}
	if workspace.Status != aliyunFlinkAPI.FlinkWorkspaceStatusRunning.String() {
		if flinkLegacyNamespaceTerminalState(workspace.Status) {
			return nil, fmt.Errorf("Flink workspace %q is in terminal state %q", workspaceID, workspace.Status)
		}
		return nil, &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink workspace %q status is %q, waiting for RUNNING", workspaceID, workspace.Status)}
	}
	if workspace.ResourceId == "" {
		return nil, &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink workspace %q namespace resource ID is not visible yet", workspaceID)}
	}

	namespaces, err := service.ListNamespaces(workspaceID)
	if err != nil {
		return nil, err
	}
	selected, published, err := selectFlinkLegacyReadyNamespaces(workspaceID, workspace, namespaces, selection)
	if err != nil {
		return nil, err
	}
	for _, namespace := range selected {
		if err := validateFlinkLegacyNamespaceReady(namespace); err != nil {
			return nil, err
		}
		queues, err := service.ListFlinkDeploymentTargets(workspace.ResourceId, namespace.Name)
		if err != nil {
			return nil, err
		}
		if err := validateFlinkLegacyDefaultQueueReady(namespace.Name, queues); err != nil {
			return nil, err
		}
	}
	return published, nil
}

func selectFlinkLegacyReadyNamespaces(
	workspaceID string,
	workspace *aliyunFlinkAPI.Workspace,
	namespaces []aliyunFlinkAPI.Namespace,
	selection flinkLegacyNamespaceReadinessSelection,
) ([]aliyunFlinkAPI.Namespace, []aliyunFlinkAPI.Namespace, error) {
	if len(namespaces) == 0 {
		return nil, nil, &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink workspace %q namespaces are not visible yet", workspaceID)}
	}

	requiredNames := selection.requiredNames
	publishFullSnapshot := false
	switch selection.mode {
	case flinkLegacyNamespaceReadinessBootstrap, flinkLegacyNamespaceReadinessBootstrapThenFullSnapshot:
		if workspace == nil || strings.TrimSpace(workspace.Name) == "" {
			return nil, nil, fmt.Errorf("Flink workspace %q workspace name is absent; cannot derive the service-generated bootstrap namespace", workspaceID)
		}
		requiredNames = map[string]struct{}{workspace.Name + "-default": {}}
		publishFullSnapshot = selection.mode == flinkLegacyNamespaceReadinessBootstrapThenFullSnapshot
	case flinkLegacyNamespaceReadinessExact:
		if len(requiredNames) == 0 {
			return nil, nil, fmt.Errorf("Flink workspace %q exact legacy namespace readiness selection is empty", workspaceID)
		}
	default:
		return nil, nil, fmt.Errorf("Flink workspace %q legacy namespace readiness selection mode %d is invalid", workspaceID, selection.mode)
	}

	selected := make([]aliyunFlinkAPI.Namespace, 0, len(requiredNames))
	found := make(map[string]struct{}, len(requiredNames))
	for _, namespace := range namespaces {
		if _, required := requiredNames[namespace.Name]; required {
			selected = append(selected, namespace)
			found[namespace.Name] = struct{}{}
		}
	}
	if len(found) != len(requiredNames) {
		missing := make([]string, 0, len(requiredNames)-len(found))
		for name := range requiredNames {
			if _, ok := found[name]; !ok {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		return nil, nil, &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink workspace %q requested namespaces %q are not visible yet", workspaceID, strings.Join(missing, ","))}
	}
	if publishFullSnapshot {
		return selected, namespaces, nil
	}
	return selected, selected, nil
}

func validateFlinkLegacyNamespaceReady(namespace aliyunFlinkAPI.Namespace) error {
	if namespace.Name == "" {
		return fmt.Errorf("Flink namespace readiness returned an empty namespace name")
	}
	if namespace.Status != "SUCCESS" && namespace.Status != "Available" {
		if flinkLegacyNamespaceTerminalState(namespace.Status) {
			return fmt.Errorf("Flink namespace %q is in terminal state %q", namespace.Name, namespace.Status)
		}
		return &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink namespace %q status is %q, waiting for SUCCESS or Available", namespace.Name, namespace.Status)}
	}
	if namespace.ResourceSpec == nil {
		return &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink namespace %q capacity is not visible yet", namespace.Name)}
	}
	return nil
}

func validateFlinkLegacyDefaultQueueReady(namespace string, queues []aliyunFlinkAPI.DeploymentTarget) error {
	for _, queue := range queues {
		if queue.Name != aliyunFlinkAPI.DefaultDeploymentTarget {
			continue
		}
		if queue.Quota == nil || queue.Quota.Request == nil || queue.Quota.Limit == nil {
			return &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink namespace %q default queue capacity is not visible yet", namespace)}
		}
		return nil
	}
	return &flinkLegacyNamespaceNotReadyError{reason: fmt.Sprintf("Flink namespace %q default queue is not visible yet", namespace)}
}

func flinkLegacyNamespaceReadinessRetryableError(err error) bool {
	var serviceErr *aliyunFlinkAPI.FlinkServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.GetErrorCode() == "404" || serviceErr.IsRetryableError()
	}

	var sdkErr *aliyunFlinkAPI.FlinkSDKError
	if !errors.As(err, &sdkErr) {
		return false
	}
	cause := errors.Unwrap(sdkErr)
	if cause == nil {
		return false
	}
	if errors.Is(cause, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(cause, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	var teaErr *tea.SDKError
	if !errors.As(cause, &teaErr) {
		return false
	}
	status := tea.IntValue(teaErr.StatusCode)
	if status == 408 || status == 429 || status >= 500 && status <= 599 {
		return true
	}
	switch tea.StringValue(teaErr.Code) {
	case "InternalError",
		"ServiceTemporarilyUnavailable",
		"ThrottlingUser",
		"Throttling",
		"RequestLimitExceeded",
		"SystemBusy",
		"ServiceUnavailable",
		"LastTokenProcessing",
		"SignatureNonceUsed",
		"RequestTimeout",
		"NetworkError":
		return true
	default:
		return false
	}
}

func flinkLegacyNamespaceTerminalState(state string) bool {
	upper := strings.ToUpper(state)
	return strings.Contains(upper, "FAIL") || upper == "DISABLE" || upper == "DISABLED" || upper == "DELETING" || upper == "DELETED" || upper == "CANCELLED" || upper == "CANCELED"
}

// WaitForWorkspaceDeleting waits for a Flink workspace to be completely deleted
func (s *FlinkService) WaitForWorkspaceDeleting(id string, timeout time.Duration) error {
	return resource.Retry(timeout, func() *resource.RetryError {
		workspace, err := s.DescribeFlinkWorkspace(id)
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			if NeedRetry(err) {
				return resource.RetryableError(WrapError(err))
			}
			return resource.NonRetryableError(WrapError(err))
		}
		return resource.RetryableError(fmt.Errorf("Flink workspace %q still exists with status %q", id, workspace.Status))
	})
}

// Instance/Workspace methods (aliases for workspace methods)
func (s *FlinkService) ListInstances() ([]aliyunFlinkAPI.Workspace, error) {
	return s.GetAPI().ListWorkspaces()
}
