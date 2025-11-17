package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func parseDeploymentId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid deployment ID format, expected workspaceId:namespace:deploymentId, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func (s *FlinkService) GetDeployment(id string) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}
	return s.GetAPI().GetDeployment(workspaceId, namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, _, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	result, err := s.GetAPI().CreateDeployment(deployment)
	if err == nil && result != nil {
		addDebugJson("CreateDeployment", result)
	}
	return result, err
}

func (s *FlinkService) UpdateDeployment(id string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	deployment.Workspace = workspaceId
	deployment.Namespace = namespaceName
	deployment.DeploymentId = deploymentId
	result, err := s.GetAPI().UpdateDeployment(deployment)
	if err == nil && result != nil {
		addDebugJson("UpdateDeployment", result)
	}
	return result, err
}

func (s *FlinkService) DeleteDeployment(id string) error {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return err
	}
	err = s.GetAPI().DeleteDeployment(workspaceId, namespaceName, deploymentId)
	if err == nil {
		addDebugJson("DeleteDeployment", fmt.Sprintf("Deployment %s deleted successfully", deploymentId))
	}
	return err
}

func (s *FlinkService) ListDeployments(id string) ([]aliyunFlinkAPI.Deployment, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid namespace ID format for listing deployments, expected workspaceId:namespace, got %s", id)
	}
	workspaceId := parts[0]
	namespaceName := parts[1]
	return s.GetAPI().ListDeployments(workspaceId, namespaceName)
}

func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := s.GetDeployment(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "NotFound", nil
			}
			return nil, "FAILED", WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
		}

		if deployment == nil {
			return nil, "NotFound", nil
		}

		for _, failState := range failStates {
			if deployment.Status == failState {
				return deployment, deployment.Status, WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
			}
		}

		return deployment, deployment.Status, nil
	}
}

func (s *FlinkService) GetDeploymentJobs(id string) ([]aliyunFlinkAPI.Job, error) {
	workspaceId, namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}

	// List jobs for the deployment
	jobs, err := s.GetAPI().ListJobs(workspaceId, namespaceName, deploymentId)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "ListJobs", AlibabaCloudSdkGoERROR)
	}

	return jobs, nil
}

// WaitForDeploymentJobsTerminal 等待部署相关的所有Job进入终端状态
func (s *FlinkService) WaitForDeploymentJobsTerminal(id string, timeout time.Duration) error {
	// 获取部署相关的所有Job
	jobs, err := s.GetDeploymentJobs(id)
	if err != nil {
		// 如果获取Job列表失败，但不是因为资源不存在，则返回错误
		if !IsNotFoundError(err) {
			return WrapErrorf(err, DefaultErrorMsg, id, "GetDeploymentJobs", AlibabaCloudSdkGoERROR)
		}
		// 如果是因为资源不存在，继续执行删除操作
		return nil
	}

	// 终端状态列表
	terminalStates := []string{
		"FINISHED",
		"FAILED",
		"CANCELLED",
		"STOPPED",
	}

	// 对于每个Job，等待其进入终端状态
	for _, job := range jobs {
		jobId := EncodeJobId(job.Workspace, job.Namespace, job.JobId)

		// 检查Job状态，如果是运行状态则先尝试停止
		jobStatus := job.GetStatus()

		// 如果Job正在运行，先尝试停止它
		if jobStatus == "RUNNING" {
			// 尝试停止Job，使用savepoint=true来安全停止
			stopErr := s.StopJob(jobId, true)
			if stopErr != nil {
				// 如果停止失败，记录日志但继续处理
				addDebugJson("StopJob", fmt.Sprintf("Failed to stop job %s: %v", jobId, stopErr))
			}
		}

		// 检查Job是否已经是终端状态
		isTerminal := false
		for _, state := range terminalStates {
			if jobStatus == state {
				isTerminal = true
				break
			}
		}

		// 如果Job不是终端状态，等待其进入终端状态
		if !isTerminal {
			nonTerminalStates := []string{
				"STARTING",
				"RUNNING",
				"STOPPING",
				"CANCELLING",
				"SUBMITTING",
				"RESTARTING",
			}

			stateConf := BuildStateConf(
				nonTerminalStates, // Pending states
				terminalStates,    // Target states
				timeout,
				5*time.Second,
				s.FlinkJobStateRefreshFunc(jobId, []string{}), // No specific fail states
			)

			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, jobId)
			}
		}
	}

	return nil
}
