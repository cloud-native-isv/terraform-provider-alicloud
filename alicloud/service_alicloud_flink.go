package alicloud

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// ResourceStateRefreshFunc is a function type used for resource state refresh operations
type ResourceStateRefreshFunc func() (interface{}, string, error)

type FlinkService struct {
	client            *connectivity.AliyunClient
	foasconsoleClient *foasconsole.Client
	ververicaClient   *ververica.Client
}

func NewFlinkService(client *connectivity.AliyunClient) (*FlinkService, error) {
	config := &openapi.Config{
		AccessKeyId:     &client.AccessKey,
		AccessKeySecret: &client.SecretKey,
		RegionId:        &client.RegionId,
		SecurityToken:   &client.SecurityToken,
	}
	foasconsoleClient, err := foasconsole.NewClient(config)
	if err != nil {
		return nil, err
	}
	ververicaClient, err := ververica.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &FlinkService{
		client:            client,
		foasconsoleClient: foasconsoleClient,
		ververicaClient:   ververicaClient,
	}, nil
}

// DescribeSupportedZones retrieves zones that support Flink instances
func (s *FlinkService) DescribeSupportedZones(request *foasconsole.DescribeSupportedZonesRequest) (*foasconsole.DescribeSupportedZonesResponse, error) {
	response, err := s.foasconsoleClient.DescribeSupportedZones(request)
	return response, WrapError(err)
}

// Instance management functions 
func (s *FlinkService) CreateInstance(request *foasconsole.CreateInstanceRequest) (*foasconsole.CreateInstanceResponse, error) {
	response, err := s.foasconsoleClient.CreateInstance(request)
	return response, WrapError(err)
}

func (s *FlinkService) DeleteInstance(request *foasconsole.DeleteInstanceRequest) (*foasconsole.DeleteInstanceResponse, error) {
	response, err := s.foasconsoleClient.DeleteInstance(request)
	return response, WrapError(err)
}

func (s *FlinkService) DescribeInstances(request *foasconsole.DescribeInstancesRequest) (*foasconsole.DescribeInstancesResponse, error) {
	response, err := s.foasconsoleClient.DescribeInstances(request)
	return response, WrapError(err)
}

// FlinkInstanceStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Flink instance
func (s *FlinkService) FlinkInstanceStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &foasconsole.DescribeInstancesRequest{}
		request.InstanceId = &id
		
		response, err := s.DescribeInstances(request)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		
		// Check if the instance was found
		if response == nil || response.Body == nil || len(response.Body.Instances) == 0 {
			return nil, "", nil
		}
		
		// Get the first instance (should be the only one since we're querying by ID)
		instance := response.Body.Instances[0]
		if instance.ClusterStatus == nil {
			return instance, "", nil
		}
		
		return instance, *instance.ClusterStatus, nil
	}
}

// Namespace management functions
func (s *FlinkService) CreateNamespace(request *foasconsole.CreateNamespaceRequest) (*foasconsole.CreateNamespaceResponse, error) {
	response, err := s.foasconsoleClient.CreateNamespace(request)
	return response, WrapError(err)
}

func (s *FlinkService) DeleteNamespace(request *foasconsole.DeleteNamespaceRequest) (*foasconsole.DeleteNamespaceResponse, error) {
	response, err := s.foasconsoleClient.DeleteNamespace(request)
	return response, WrapError(err)
}

func (s *FlinkService) DescribeNamespaces(request *foasconsole.DescribeNamespacesRequest) (*foasconsole.DescribeNamespacesResponse, error) {
	response, err := s.foasconsoleClient.DescribeNamespaces(request)
	return response, WrapError(err)
}

// Member management functions
func (s *FlinkService) CreateMember(namespace *string, request *ververica.CreateMemberRequest) (*ververica.CreateMemberResponse, error) {
    response, err := s.ververicaClient.CreateMember(namespace, request)
    return response, WrapError(err)
}

func (s *FlinkService) GetMember(namespace *string, member *string) (*ververica.GetMemberResponse, error) {
    response, err := s.ververicaClient.GetMember(namespace, member)
    return response, WrapError(err)
}

func (s *FlinkService) UpdateMember(namespace *string, request *ververica.UpdateMemberRequest) (*ververica.UpdateMemberResponse, error) {
    response, err := s.ververicaClient.UpdateMember(namespace, request)
    return response, WrapError(err)
}

func (s *FlinkService) DeleteMember(namespace *string, member *string) (*ververica.DeleteMemberResponse, error) {
    response, err := s.ververicaClient.DeleteMember(namespace, member)
    return response, WrapError(err)
}

// Deployment management functions
func (s *FlinkService) CreateDeployment(namespace *string, request *ververica.CreateDeploymentRequest) (*ververica.CreateDeploymentResponse, error) {
    response, err := s.ververicaClient.CreateDeployment(namespace, request)
    return response, WrapError(err)
}

func (s *FlinkService) GetDeployment(namespace *string, deploymentId *string) (*ververica.GetDeploymentResponse, error) {
    response, err := s.ververicaClient.GetDeployment(namespace, deploymentId)
    return response, WrapError(err)
}

func (s *FlinkService) UpdateDeployment(namespace *string, deploymentId *string, request *ververica.UpdateDeploymentRequest) (*ververica.UpdateDeploymentResponse, error) {
    response, err := s.ververicaClient.UpdateDeployment(namespace, deploymentId, request)
    return response, WrapError(err)
}

func (s *FlinkService) DeleteDeployment(namespace *string, deploymentId *string) (*ververica.DeleteDeploymentResponse, error) {
    response, err := s.ververicaClient.DeleteDeployment(namespace, deploymentId)
    return response, WrapError(err)
}

func (s *FlinkService) ListDeployments(namespace *string, request *ververica.ListDeploymentsRequest) (*ververica.ListDeploymentsResponse, error) {
    response, err := s.ververicaClient.ListDeployments(namespace, request)
    return response, WrapError(err)
}

// Job management functions
func (s *FlinkService) StartJobWithParams(namespace *string, request *ververica.StartJobWithParamsRequest) (*ververica.StartJobWithParamsResponse, error) {
    response, err := s.ververicaClient.StartJobWithParams(namespace, request)
    return response, WrapError(err)
}

func (s *FlinkService) StopJob(namespace *string, jobId *string, request *ververica.StopJobRequest) (*ververica.StopJobResponse, error) {
    response, err := s.ververicaClient.StopJob(namespace, jobId, request)
    return response, WrapError(err)
}

func (s *FlinkService) GetJob(namespace *string, jobId *string) (*ververica.GetJobResponse, error) {
    response, err := s.ververicaClient.GetJob(namespace, jobId)
    return response, WrapError(err)
}

// FlinkDeploymentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Flink deployment
func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string, failStates []string) ResourceStateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFlinkDeployment(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		
		for _, failState := range failStates {
			if failState == "EntityNotExist.Deployment" {
				return object, "", nil
			}
		}
				
		return object, "CREATED", nil
	}
}

// FlinkJobStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Flink job
func (s *FlinkService) FlinkJobStateRefreshFunc(namespace, jobId string, failStates []string) ResourceStateRefreshFunc {
	return func() (interface{}, string, error) {
		response, err := s.GetJob(&namespace, &jobId)
		if err != nil {
			if IsExpectedErrors(err, []string{"JobNotFound"}) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		
		if response != nil && response.Body != nil && response.Body.Data != nil && response.Body.Data.Status != nil {
			// Convert status pointer to string value
			return response, tea.StringValue(response.Body.Data.Status.CurrentJobStatus), nil
		}
		
		return response, "", nil
	}
}

// DescribeFlinkDeployment describes a flink deployment's details
func (s *FlinkService) DescribeFlinkDeployment(id string) (*ververica.GetDeploymentResponse, error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	
	namespace := parts[0]
	deploymentId := parts[1]
	
	return s.GetDeployment(&namespace, &deploymentId)
}

func (s *FlinkService) CreateVariable(request map[string]interface{}) (map[string]interface{}, error) {
	namespace := request["flink_instance_id"].(string)
	req := &ververica.CreateVariableRequest{
		Name:        tea.String(request["name"].(string)),
		Value:       tea.String(request["value"].(string)),
		Description: tea.String(request["description"].(string)),
	}
	resp, err := s.ververicaClient.CreateVariable(tea.String(namespace), req)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.ToMap(), nil
}

func (s *FlinkService) DescribeVariable(variableId string) (map[string]interface{}, error) {
	parts := strings.Split(variableId, "/")
	if len(parts) != 2 {
		return nil, WrapError(errors.New("invalid variable ID format"))
	}
	namespace, varId := parts[0], parts[1]
	resp, err := s.ververicaClient.DescribeVariable(tea.String(namespace), tea.String(varId))
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.ToMap(), nil
}

func (s *FlinkService) UpdateVariable(request map[string]interface{}) (map[string]interface{}, error) {
	variableId := request["id"].(string)
	parts := strings.Split(variableId, "/")
	if len(parts) != 2 {
		return nil, WrapError(errors.New("invalid variable ID format"))
	}
	namespace, varId := parts[0], parts[1]
	req := &ververica.UpdateVariableRequest{
		Value:       tea.String(request["value"].(string)),
		Description: tea.String(request["description"].(string)),
	}
	resp, err := s.ververicaClient.UpdateVariable(tea.String(namespace), tea.String(varId), req)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.ToMap(), nil
}

func (s *FlinkService) DeleteVariable(variableId string) error {
	parts := strings.Split(variableId, "/")
	if len(parts) != 2 {
		return WrapError(errors.New("invalid variable ID format"))
	}
	namespace, varId := parts[0], parts[1]
	_, err := s.ververicaClient.DeleteVariable(tea.String(namespace), tea.String(varId))
	return WrapError(err)
}

// 新增ListVariables方法
func (s *FlinkService) ListVariables(namespace string, opts map[string]interface{}) ([]map[string]interface{}, error) {
	request := &ververica.ListVariablesRequest{}
	// 根据opts填充分页参数等
	resp, err := s.ververicaClient.ListVariables(tea.String(namespace), request)
	if err != nil {
		return nil, WrapError(err)
	}
	vars := make([]map[string]interface{}, 0)
	for _, v := range resp.Body.Variables {
		vars = append(vars, v.ToMap())
	}
	return vars, nil
}