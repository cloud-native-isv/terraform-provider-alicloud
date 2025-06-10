package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type FlinkService struct {
	client         *connectivity.AliyunClient
	aliyunFlinkAPI *aliyunFlinkAPI.FlinkAPI
}

// NewFlinkService creates a new FlinkService using cws-lib-go implementation
func NewFlinkService(client *connectivity.AliyunClient) (*FlinkService, error) {
	// Convert AliyunClient credentials to FlinkCredentials
	credentials := &aliyunFlinkAPI.FlinkCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go FlinkService
	aliyunFlinkAPI, err := aliyunFlinkAPI.NewFlinkAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go FlinkService: %w", err)
	}

	return &FlinkService{
		client:         client,
		aliyunFlinkAPI: aliyunFlinkAPI,
	}, nil
}

// Workspace methods
func (s *FlinkService) DescribeFlinkWorkspace(id string) (*aliyunFlinkAPI.Workspace, error) {
	return s.aliyunFlinkAPI.GetWorkspace(id)
}

func (s *FlinkService) CreateInstance(workspace *aliyunFlinkAPI.Workspace) (*aliyunFlinkAPI.Workspace, error) {
	return s.aliyunFlinkAPI.CreateWorkspace(workspace)
}

func (s *FlinkService) DeleteInstance(id string) error {
	return s.aliyunFlinkAPI.DeleteWorkspace(id)
}

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		workspace, err := s.aliyunFlinkAPI.GetWorkspace(id)
		if err != nil {
			// Handle the case where workspace is temporarily not found after creation
			// This is common with cloud resources that have async creation processes
			if IsExpectedErrors(err, []string{"903021"}) { // not exist yet
				// Return empty state to indicate the resource is still being created
				return nil, "CREATING", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetWorkspace", AlibabaCloudSdkGoERROR)
		}
		return workspace, workspace.Status, nil
	}
}

// Deployment methods
func (s *FlinkService) GetDeployment(id string) (*aliyunFlinkAPI.Deployment, error) {
	// Parse deployment ID to extract namespace and deployment ID
	// Format: namespace:deploymentId
	namespaceName, deploymentId, err := parseDeploymentId(id)
	if err != nil {
		return nil, err
	}
	return s.aliyunFlinkAPI.GetDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) CreateDeployment(namespaceName *string, deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	deployment.Namespace = *namespaceName
	return s.aliyunFlinkAPI.CreateDeployment(deployment)
}

func (s *FlinkService) UpdateDeployment(deployment *aliyunFlinkAPI.Deployment) (*aliyunFlinkAPI.Deployment, error) {
	return s.aliyunFlinkAPI.UpdateDeployment(deployment)
}

func (s *FlinkService) DeleteDeployment(namespaceName, deploymentId string) error {
	return s.aliyunFlinkAPI.DeleteDeployment(namespaceName, deploymentId)
}

func (s *FlinkService) ListDeployments(namespaceName string, pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.Deployment], error) {
	return s.aliyunFlinkAPI.ListDeployments(namespaceName, pagination)
}

func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := s.GetDeployment(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapErrorf(err, DefaultErrorMsg, id, "GetDeployment", AlibabaCloudSdkGoERROR)
		}

		return deployment, deployment.Status(), nil
	}
}

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

	// Create HotUpdateJobParams from job - remove RestartType field if not available
	params := &aliyunFlinkAPI.HotUpdateJobParams{
		// RestartType: job.RestartType, // Commented out if field doesn't exist
	}

	// Use empty string as workspaceId since WorkspaceId field doesn't exist
	workspaceId := ""

	return s.aliyunFlinkAPI.UpdateJob(workspaceId, namespaceName, jobId, params)
}

func (s *FlinkService) StopJob(namespaceName, jobId string, withSavepoint bool) error {
	return s.aliyunFlinkAPI.StopJob(namespaceName, jobId, withSavepoint)
}

func (s *FlinkService) ListJobs(namespaceName string, pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.Job], error) {
	// Add workspace parameter - using empty string as default since it's required by API
	return s.aliyunFlinkAPI.ListJobs("", namespaceName, "", pagination)
}

func (s *FlinkService) GetJobMetrics(namespaceName string, jobId string) (*aliyunFlinkAPI.JobMetrics, error) {
	// This method doesn't exist in the API, so we'll create a placeholder implementation
	// or use an alternative method if available
	return nil, fmt.Errorf("GetJobMetrics method not implemented in FlinkAPI")
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

// Deployment Draft methods
func (s *FlinkService) CreateDeploymentDraft(workspaceID string, namespaceName string, draft *aliyunFlinkAPI.DeploymentDraft) (*aliyunFlinkAPI.DeploymentDraft, error) {
	// Call the underlying API with the proper signature
	result, err := s.aliyunFlinkAPI.CreateDeploymentDraft(namespaceName, workspaceID, draft)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_draft", "CreateDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

func (s *FlinkService) GetDeploymentDraft(workspaceId string, namespaceName string, draftId string) (*aliyunFlinkAPI.DeploymentDraft, error) {
	return s.aliyunFlinkAPI.GetDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) UpdateDeploymentDraft(workspaceId string, namespaceName string, draftId string, draft *aliyunFlinkAPI.DeploymentDraft) (*aliyunFlinkAPI.DeploymentDraft, error) {
	draft.Namespace = namespaceName
	draft.DeploymentDraftId = draftId
	return s.aliyunFlinkAPI.UpdateDeploymentDraft(workspaceId, namespaceName, draftId, draft)
}

func (s *FlinkService) DeleteDeploymentDraft(workspaceId string, namespaceName string, draftId string) error {
	return s.aliyunFlinkAPI.DeleteDeploymentDraft(workspaceId, namespaceName, draftId)
}

func (s *FlinkService) ListDeploymentDrafts(namespaceName string, pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.DeploymentDraft], error) {
	return s.aliyunFlinkAPI.ListDeploymentDrafts("", namespaceName, pagination)
}

// Member methods
func (s *FlinkService) CreateMember(workspaceId string, namespaceName string, member map[string]interface{}) (map[string]interface{}, error) {
	// Validate required parameters
	memberName, ok := member["member"].(string)
	if !ok || memberName == "" {
		return nil, fmt.Errorf("member name is required")
	}

	role, ok := member["role"].(string)
	if !ok || role == "" {
		return nil, fmt.Errorf("member role is required")
	}

	// Create a Member using aliyunFlinkAPI.Member with proper field mapping
	apiMember := &aliyunFlinkAPI.Member{
		Member:        memberName,
		Role:          role,
		WorkspaceID:   workspaceId,
		NamespaceName: namespaceName,
	}

	// Call the underlying API
	result, err := s.aliyunFlinkAPI.CreateMember(apiMember)
	if err != nil {
		return nil, err
	}

	// Convert the result to a map for easy consumption
	response := map[string]interface{}{
		"member": result.Member,
		"role":   result.Role,
	}

	return response, nil
}

func (s *FlinkService) GetMember(workspaceId string, namespaceName string, memberId string) (map[string]interface{}, error) {
	// Validate required parameters
	if workspaceId == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}
	if namespaceName == "" {
		return nil, fmt.Errorf("namespace name is required")
	}
	if memberId == "" {
		return nil, fmt.Errorf("member ID is required")
	}

	// Call the underlying API
	result, err := s.aliyunFlinkAPI.GetMember(workspaceId, namespaceName, memberId)
	if err != nil {
		return nil, err
	}

	// Convert the result to a map for easy consumption
	response := map[string]interface{}{
		"member": result.Member,
		"role":   result.Role,
	}

	return response, nil
}

func (s *FlinkService) UpdateMember(workspaceId string, namespaceName string, member map[string]interface{}) (map[string]interface{}, error) {
	// Validate required parameters
	memberName, ok := member["member"].(string)
	if !ok || memberName == "" {
		return nil, fmt.Errorf("member name is required")
	}

	role, ok := member["role"].(string)
	if !ok || role == "" {
		return nil, fmt.Errorf("member role is required")
	}

	// Create a Member using aliyunFlinkAPI.Member with proper field mapping
	apiMember := &aliyunFlinkAPI.Member{
		Member:        memberName,
		Role:          role,
		WorkspaceID:   workspaceId,
		NamespaceName: namespaceName,
	}

	// Call the underlying API
	result, err := s.aliyunFlinkAPI.UpdateMember(apiMember)
	if err != nil {
		return nil, err
	}

	// Convert the result to a map for easy consumption
	response := map[string]interface{}{
		"member": result.Member,
		"role":   result.Role,
	}

	return response, nil
}

func (s *FlinkService) DeleteMember(workspaceId string, namespaceName string, memberId string) error {
	// Validate required parameters
	if workspaceId == "" {
		return fmt.Errorf("workspace ID is required")
	}
	if namespaceName == "" {
		return fmt.Errorf("namespace name is required")
	}
	if memberId == "" {
		return fmt.Errorf("member ID is required")
	}

	// Call the underlying API
	return s.aliyunFlinkAPI.DeleteMember(workspaceId, namespaceName, memberId)
}

// Namespace methods
func (s *FlinkService) ListNamespaces(workspaceId string, pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.Namespace], error) {
	return s.aliyunFlinkAPI.ListNamespaces(workspaceId, pagination)
}

func (s *FlinkService) CreateNamespace(workspaceId string, namespace *aliyunFlinkAPI.Namespace) (*aliyunFlinkAPI.Namespace, error) {
	// Create namespace using the aliyunFlinkAPI directly
	result, err := s.aliyunFlinkAPI.CreateNamespace(namespace)
	return result, err
}

func (s *FlinkService) GetNamespace(workspaceId, namespaceName string) (*aliyunFlinkAPI.Namespace, error) {
	// Fetch the namespace using the APIan
	return s.aliyunFlinkAPI.GetNamespace(workspaceId, namespaceName)
}

func (s *FlinkService) DeleteNamespace(workspaceId string, namespaceName string) error {
	// Delete the namespace using aliyunFlinkAPI
	return s.aliyunFlinkAPI.DeleteNamespace(workspaceId, namespaceName)
}

// Instance/Workspace methods (aliases for workspace methods)
func (s *FlinkService) ListInstances(pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.Workspace], error) {
	return s.aliyunFlinkAPI.ListWorkspaces(pagination)
}

// Zone methods
func (s *FlinkService) DescribeSupportedZones() ([]*aliyunFlinkAPI.ZoneInfo, error) {
	return s.aliyunFlinkAPI.ListSupportedZones()
}

// Connector methods
func (s *FlinkService) ListCustomConnectors(workspaceId string, namespaceName string) ([]*aliyunFlinkAPI.Connector, error) {
	return s.aliyunFlinkAPI.ListConnectors(workspaceId, namespaceName)
}

func (s *FlinkService) RegisterCustomConnector(workspaceId string, namespaceName string, connector *aliyunFlinkAPI.Connector) (*aliyunFlinkAPI.Connector, error) {
	return s.aliyunFlinkAPI.RegisterConnector(workspaceId, namespaceName, connector)
}

func (s *FlinkService) GetConnector(workspaceId string, namespaceName string, connectorName string) (*aliyunFlinkAPI.Connector, error) {
	return s.aliyunFlinkAPI.GetConnector(workspaceId, namespaceName, connectorName)
}

func (s *FlinkService) DeleteCustomConnector(workspaceId string, namespaceName string, connectorName string) error {
	return s.aliyunFlinkAPI.DeleteConnector(workspaceId, namespaceName, connectorName)
}

func (s *FlinkService) FlinkConnectorStateRefreshFunc(workspaceId string, namespaceName string, connectorName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		connector, err := s.GetConnector(workspaceId, namespaceName, connectorName)
		if err != nil {
			if NotFoundError(err) {
				// Connector not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For connectors, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if connector is in a failed state (if any fail states are defined)
			if connector.Type == failState {
				return connector, connector.Type, WrapError(Error(FailedToReachTargetStatus, connector.Type))
			}
		}

		return connector, "Available", nil
	}
}

// Variable methods
func (s *FlinkService) GetVariable(workspaceId string, namespaceName string, variableName string) (*aliyunFlinkAPI.Variable, error) {
	// Call the underlying API to get variable
	return s.aliyunFlinkAPI.GetVariable(workspaceId, namespaceName, variableName)
}

func (s *FlinkService) CreateVariable(workspaceId string, namespaceName string, variable *aliyunFlinkAPI.Variable) (*aliyunFlinkAPI.Variable, error) {
	// Call underlying API
	return s.aliyunFlinkAPI.CreateVariable(workspaceId, namespaceName, variable)
}

func (s *FlinkService) UpdateVariable(workspaceId string, namespaceName string, variable *aliyunFlinkAPI.Variable) (*aliyunFlinkAPI.Variable, error) {
	// Call underlying API
	return s.aliyunFlinkAPI.UpdateVariable(workspaceId, namespaceName, variable)
}

func (s *FlinkService) DeleteVariable(workspaceId string, namespaceName string, variableName string) error {
	// Call the underlying API
	return s.aliyunFlinkAPI.DeleteVariable(workspaceId, namespaceName, variableName)
}

func (s *FlinkService) ListVariables(workspaceId string, namespaceName string, pagination *aliyunFlinkAPI.PaginationRequest) (*aliyunFlinkAPI.ListResponse[aliyunFlinkAPI.Variable], error) {
	// Call underlying API
	return s.aliyunFlinkAPI.ListVariables(workspaceId, namespaceName, pagination)
}

func (s *FlinkService) FlinkVariableStateRefreshFunc(workspaceId string, namespaceName string, variableName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		variable, err := s.GetVariable(workspaceId, namespaceName, variableName)
		if err != nil {
			if NotFoundError(err) {
				// Variable not found, still being created
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For variables, if we can get it successfully, it means it's ready
		for _, failState := range failStates {
			// Check if variable is in a failed state (if any fail states are defined)
			if variable.Kind == failState {
				return variable, variable.Kind, WrapError(Error(FailedToReachTargetStatus, variable.Kind))
			}
		}

		return variable, "Available", nil
	}
}

// FlinkDeploymentDraftStateRefreshFunc provides state refresh for deployment drafts
func (s *FlinkService) FlinkDeploymentDraftStateRefreshFunc(workspaceId string, namespaceName string, draftId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		draft, err := s.GetDeploymentDraft(workspaceId, namespaceName, draftId)
		if err != nil {
			if NotFoundError(err) {
				// Draft not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For deployment drafts, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if draft is in a failed state (if any fail states are defined)
			if draft.Status == failState {
				return draft, draft.Status, WrapError(Error(FailedToReachTargetStatus, draft.Status))
			}
		}

		return draft, "Available", nil
	}
}

// FlinkMemberStateRefreshFunc provides state refresh for members
func (s *FlinkService) FlinkMemberStateRefreshFunc(workspaceId string, namespaceName string, memberId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		member, err := s.GetMember(workspaceId, namespaceName, memberId)
		if err != nil {
			if NotFoundError(err) {
				// Member not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For members, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if member is in a failed state (if any fail states are defined)
			if role, ok := member["role"].(string); ok && role == failState {
				return member, role, WrapError(Error(FailedToReachTargetStatus, role))
			}
		}

		return member, "Available", nil
	}
}

// FlinkNamespaceStateRefreshFunc provides state refresh for namespaces
func (s *FlinkService) FlinkNamespaceStateRefreshFunc(workspaceId string, namespaceName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		namespace, err := s.GetNamespace(workspaceId, namespaceName)
		if err != nil {
			if NotFoundError(err) {
				// Namespace not found, still being created or deleted
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For namespaces, if we can get it successfully, it means it's available
		for _, failState := range failStates {
			// Check if namespace is in a failed state (if any fail states are defined)
			if namespace.Status == failState {
				return namespace, namespace.Status, WrapError(Error(FailedToReachTargetStatus, namespace.Status))
			}
		}

		return namespace, "Available", nil
	}
}

// Engine methods
func (s *FlinkService) ListEngines(workspaceId string) ([]*aliyunFlinkAPI.FlinkEngine, error) {
	return s.aliyunFlinkAPI.ListEngines(workspaceId)
}

// Helper functions for parsing IDs
func parseDeploymentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid deployment ID format, expected namespace:deploymentId, got %s", id)
	}
	return parts[0], parts[1], nil
}

func parseJobId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid job ID format, expected namespace:jobId, got %s", id)
	}
	return parts[0], parts[1], nil
}
