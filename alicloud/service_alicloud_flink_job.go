package alicloud

import (
	"fmt"
	"strings"
	"time"

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
	// Validate input parameter
	if id == "" {
		return nil, fmt.Errorf("job ID cannot be empty")
	}

	// Parse job ID to extract workspace ID, namespace and job ID
	// Format: workspaceId:namespace:jobId
	workspaceId, namespaceName, jobId, err := parseJobId(id)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "parseJobId", AlibabaCloudSdkGoERROR)
	}

	job, err := s.flinkAPI.GetJob(workspaceId, namespaceName, jobId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetJob", AlibabaCloudSdkGoERROR)
	}

	return job, nil
}

func (s *FlinkService) StartJob(params *aliyunFlinkAPI.JobStartParameters) (*aliyunFlinkAPI.Job, error) {
	// Validate required parameters
	if params == nil {
		return nil, fmt.Errorf("JobStartParameters cannot be nil")
	}
	if params.WorkspaceId == "" {
		return nil, fmt.Errorf("WorkspaceId is required in JobStartParameters")
	}
	if params.Namespace == "" {
		return nil, fmt.Errorf("Namespace is required in JobStartParameters")
	}

	job, err := s.flinkAPI.StartJob(params)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", params.WorkspaceId, params.Namespace), "StartJob", AlibabaCloudSdkGoERROR)
	}

	return job, nil
}

func (s *FlinkService) UpdateJob(stateId string, updateParams *aliyunFlinkAPI.HotUpdateJobParams) (*aliyunFlinkAPI.HotUpdateJobResult, error) {
	// Validate input parameters
	if stateId == "" {
		return nil, fmt.Errorf("job state ID cannot be empty")
	}
	if updateParams == nil {
		return nil, fmt.Errorf("update parameters cannot be nil")
	}

	// Parse job ID to extract workspace ID, namespace and job ID
	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, stateId, "parseJobId", AlibabaCloudSdkGoERROR)
	}

	result, err := s.flinkAPI.UpdateJob(workspaceId, namespaceName, jobId, updateParams)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, stateId, "UpdateJob", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

func (s *FlinkService) StopJob(stateId string, withSavepoint bool) error {
	// Validate input parameter
	if stateId == "" {
		return fmt.Errorf("job state ID cannot be empty")
	}

	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, stateId, "parseJobId", AlibabaCloudSdkGoERROR)
	}

	// First get the current job status
	job, err := s.flinkAPI.GetJob(workspaceId, namespaceName, jobId)
	if err != nil {
		if NotFoundError(err) {
			// Job doesn't exist, no need to stop
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, stateId, "GetJob", AlibabaCloudSdkGoERROR)
	}

	// Check if job is running, only stop if it's running
	if job != nil {
		jobStatus := job.GetStatus()
		if jobStatus == aliyunFlinkAPI.FlinkJobStatusRunning.String() {
			err = s.flinkAPI.StopJob(workspaceId, namespaceName, jobId, withSavepoint)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, stateId, "StopJob", AlibabaCloudSdkGoERROR)
			}
		}
	}

	// Job is not running, no need to stop
	return nil
}

func (s *FlinkService) DeleteJob(stateId string) error {
	// Validate input parameter
	if stateId == "" {
		return fmt.Errorf("job state ID cannot be empty")
	}

	workspaceId, namespaceName, jobId, err := parseJobId(stateId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, stateId, "parseJobId", AlibabaCloudSdkGoERROR)
	}

	err = s.flinkAPI.DeleteJob(workspaceId, namespaceName, jobId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, stateId, "DeleteJob", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *FlinkService) ListJobs(workspaceId, namespaceName, deploymentId string) ([]aliyunFlinkAPI.Job, error) {
	// Validate required parameters
	if workspaceId == "" {
		return nil, fmt.Errorf("workspaceId cannot be empty")
	}
	if namespaceName == "" {
		return nil, fmt.Errorf("namespaceName cannot be empty")
	}

	jobs, err := s.flinkAPI.ListJobs(workspaceId, namespaceName, deploymentId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", workspaceId, namespaceName), "ListJobs", AlibabaCloudSdkGoERROR)
	}

	return jobs, nil
}

func (s *FlinkService) FlinkJobStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		job, err := s.DescribeFlinkJob(id)
		if err != nil {
			if NotFoundError(err) {
				// For deletion scenarios, return nil to indicate resource absence
				// This allows WaitForState to properly handle the "waiting for absence" case
				return nil, "", nil
			}
			return nil, aliyunFlinkAPI.FlinkJobStatusFailed.String(), WrapErrorf(err, DefaultErrorMsg, id, "DescribeFlinkJob", AlibabaCloudSdkGoERROR)
		}

		// If job is nil, it means the resource doesn't exist
		if job == nil {
			// For deletion scenarios, return nil to indicate resource absence
			return nil, "", nil
		}

		return job, job.GetStatus(), nil
	}
}

func (s *FlinkService) WaitForFlinkJobCreating(id string, timeout time.Duration) error {
	createPendingStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusStarting,
		aliyunFlinkAPI.FlinkJobStatusStopped,
	})
	createExpectStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusRunning,
	})
	createFailedStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusFailed,
	})

	stateConf := BuildStateConf(
		createPendingStatus,
		createExpectStatus,
		timeout,
		5*time.Second,
		s.FlinkJobStateRefreshFunc(id, createFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

func (s *FlinkService) WaitForFlinkJobStopping(id string, timeout time.Duration) error {
	stopPendingStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusRunning,
		aliyunFlinkAPI.FlinkJobStatusStopping,
		aliyunFlinkAPI.FlinkJobStatusCancelling,
	})
	stopExpectStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusFailed,
		aliyunFlinkAPI.FlinkJobStatusFinished,
		aliyunFlinkAPI.FlinkJobStatusCancelled,
		aliyunFlinkAPI.FlinkJobStatusStopped,
		aliyunFlinkAPI.FlinkJobStatusNotFound,
	})
	stopFailedStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusFailed,
	})

	stateConf := BuildStateConf(
		stopPendingStatus,
		stopExpectStatus,
		timeout,
		5*time.Second,
		s.FlinkJobStateRefreshFunc(id, stopFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

func (s *FlinkService) WaitForFlinkJobDeleting(id string, timeout time.Duration) error {
	deletePendingStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusFailed,
		aliyunFlinkAPI.FlinkJobStatusFinished,
		aliyunFlinkAPI.FlinkJobStatusCancelled,
		aliyunFlinkAPI.FlinkJobStatusStopped,
	})
	deleteExpectStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{})
	deleteFailedStatus := aliyunFlinkAPI.FlinkJobStatusesToStrings([]aliyunFlinkAPI.FlinkJobStatus{
		aliyunFlinkAPI.FlinkJobStatusFailed,
	})

	stateConf := BuildStateConf(
		deletePendingStatus,
		deleteExpectStatus,
		timeout,
		5*time.Second,
		s.FlinkJobStateRefreshFunc(id, deleteFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}
