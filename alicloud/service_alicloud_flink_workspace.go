package alicloud

import (
	"fmt"
	"time"

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
	if workspace.HighAvailability != nil && workspace.HighAvailability.Enabled {
		request, err := flinkworkspace.BuildCreateInstanceBody(workspace, options)
		if err != nil {
			return nil, err
		}
		// CreateInstance has no idempotency token. Retrying a response-lost 5xx
		// could purchase a second workspace, so disable transport-level retries.
		response, err := s.client.RpcPost("foasconsole", "2021-10-28", "CreateInstance", nil, request, false)
		if err != nil {
			return nil, err
		}
		instanceID, err := flinkworkspace.InstanceIDFromCreateResponse(response)
		if err != nil {
			return nil, err
		}
		workspace.Id = instanceID
		return workspace, nil
	}
	return s.GetAPI().CreateWorkspace(workspace)
}

func (s *FlinkService) DeleteInstance(id string) error {
	return s.GetAPI().DeleteWorkspace(id)
}

type flinkRefundClient interface {
	IsInternationalAccount() bool
	RpcPostWithEndpoint(string, string, string, map[string]interface{}, map[string]interface{}, bool, string) (map[string]interface{}, error)
}

func (s *FlinkService) RefundInstance(id string) error {
	return refundFlinkWorkspace(s.client, s.client.RegionId, id, buildClientToken("RefundInstance"))
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
