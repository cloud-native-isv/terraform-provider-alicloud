package alicloud

import (
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Job methods
func (s *FlinkService) DescribeFlinkJob(id string) (*aliyunFlinkAPI.Job, error) {
	// Parse job ID to extract namespace and job ID
	// Format: namespace:jobId
	namespaceName, jobId, err := parseJobId(id)
	if err != nil {
		return nil, err
	}
	return s.aliyunFlinkAPI.GetJob(namespaceName, jobId)
}

func (s *FlinkService) StartJobWithParams(namespaceName string, job *aliyunFlinkAPI.Job) (*aliyunFlinkAPI.Job, error) {
	job.Namespace = namespaceName
	return s.aliyunFlinkAPI.StartJob(job)
}

func (s *FlinkService) UpdateJob(job *aliyunFlinkAPI.Job) (*aliyunFlinkAPI.HotUpdateJobResult, error) {
	// Parse job ID to extract namespace and job ID
	namespaceName, jobId, err := parseJobId(job.JobId)
	if err != nil {
		return nil, err
	}

	// Create HotUpdateJobParams from job with proper strong typing
	params := &aliyunFlinkAPI.HotUpdateJobParams{
		JobConfig: job.FlinkConf, // Use strong typed FlinkConf field
	}

	// Get workspace ID from job context or use job's workspace field
	workspaceId := job.Workspace

	return s.aliyunFlinkAPI.UpdateJob(workspaceId, namespaceName, jobId, params)
}

func (s *FlinkService) StopJob(namespaceName, jobId string, withSavepoint bool) error {
	return s.aliyunFlinkAPI.StopJob(namespaceName, jobId, withSavepoint)
}

func (s *FlinkService) ListJobs(workspaceId, namespaceName, deploymentId string) ([]aliyunFlinkAPI.Job, error) {
	return s.aliyunFlinkAPI.ListJobs(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) FlinkJobStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		job, err := s.DescribeFlinkJob(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "DescribeFlinkJob", AlibabaCloudSdkGoERROR)
		}

		return job, "", nil
	}
}
