package alicloud

import (
	"fmt"
	"strings"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func parseJobId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid job ID format, expected workspaceId:namespace:jobId, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// Job methods
func (s *FlinkService) DescribeFlinkJob(id string) (*aliyunFlinkAPI.Job, error) {
	// Parse job ID to extract workspace ID, namespace and job ID
	// Format: workspaceId:namespace:jobId
	workspaceId, namespaceName, jobId, err := parseJobId(id)
	if err != nil {
		return nil, err
	}

	return s.flinkAPI.GetJob(workspaceId, namespaceName, jobId)
}

func (s *FlinkService) StartJob(params *aliyunFlinkAPI.JobStartParameters) (*aliyunFlinkAPI.Job, error) {
	// Validate required parameters
	if params.WorkspaceId == "" {
		return nil, fmt.Errorf("WorkspaceId is required in JobStartParameters")
	}
	if params.Namespace == "" {
		return nil, fmt.Errorf("Namespace is required in JobStartParameters")
	}

	return s.flinkAPI.StartJob(params)
}

func (s *FlinkService) UpdateJob(stateId string, updateParams *aliyunFlinkAPI.HotUpdateJobParams) (*aliyunFlinkAPI.HotUpdateJobResult, error) {
	// Parse job ID to extract workspace ID, namespace and job ID
	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return nil, err
	}

	return s.flinkAPI.UpdateJob(workspaceId, namespaceName, jobId, updateParams)
}

func (s *FlinkService) StopJob(stateId string, withSavepoint bool) error {
	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return err
	}
	return s.flinkAPI.StopJob(workspaceId, namespaceName, jobId, withSavepoint)
}

func (s *FlinkService) DeleteJob(stateId string) error {
	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return err
	}
	return s.flinkAPI.DeleteJob(workspaceId, namespaceName, jobId)
}

func (s *FlinkService) ListJobs(workspaceId, namespaceName, deploymentId string) ([]aliyunFlinkAPI.Job, error) {
	return s.flinkAPI.ListJobs(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) FlinkJobStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		job, err := s.DescribeFlinkJob(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "NotFound", nil
			}
			return nil, "FAILED", WrapErrorf(err, DefaultErrorMsg, id, "DescribeFlinkJob", AlibabaCloudSdkGoERROR)
		}

		// If job is nil, it means the resource doesn't exist
		if job == nil {
			return nil, "NotFound", nil
		}

		// Check job status and generate error if job has failed
		if job.Status != nil && job.Status.CurrentJobStatus != nil {
			status := job.GetStatus()
			if status == "FINISHED" || status == "RUNNING" {
				return job, "RUNNING", nil
			}
		} else {
			return nil, "NotFound", nil
		}

		// Check for fail states
		for _, failState := range failStates {
			if job.GetStatus() == failState {
				return job, job.GetStatus(), WrapErrorf(err, DefaultErrorMsg, id, "DescribeFlinkJob", AlibabaCloudSdkGoERROR)
			}
		}

		return job, job.GetStatus(), nil
	}
}
