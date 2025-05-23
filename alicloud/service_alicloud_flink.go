package alicloud

import (
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
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

func (s *FlinkService) GetRegionId() string {
	return s.client.RegionId
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

func (s *FlinkService) ListInstances() ([]*foasconsole.DescribeInstancesResponseBodyInstances, error) {
	region := s.GetRegionId()
	pageIndex := int32(1)
	pageSize := int32(50)
	var instances []*foasconsole.DescribeInstancesResponseBodyInstances

	for {
		request := &foasconsole.DescribeInstancesRequest{
			Region:    tea.String(region),
			PageIndex: tea.Int32(pageIndex),
			PageSize:  tea.Int32(pageSize),
		}

		response, err := s.DescribeInstances(request)
		if err != nil {
			return nil, err
		}

		instances = append(instances, response.Body.Instances...)

		if *response.Body.PageIndex >= *response.Body.TotalPage {
			break
		}
		pageIndex++
	}

	return instances, nil
}

func (s *FlinkService) GetInstance(instanceId string) (*foasconsole.DescribeInstancesResponseBodyInstances, error) {
	region := s.GetRegionId()
	request := &foasconsole.DescribeInstancesRequest{
		Region:     tea.String(region),
		InstanceId: tea.String(instanceId),
	}

	response, err := s.DescribeInstances(request)
	if err != nil {
		return nil, WrapError(err)
	}

	if len(response.Body.Instances) == 0 {
		return nil, fmt.Errorf("instance %s not found", instanceId)
	}

	return response.Body.Instances[0], nil
}

func (s *FlinkService) FlinkWorkspaceStateRefreshFunc(id string) resource.StateRefreshFunc {
	region := s.GetRegionId()
	return func() (interface{}, string, error) {
		request := &foasconsole.DescribeInstancesRequest{
			Region:     tea.String(region),
			InstanceId: tea.String(id),
		}
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

func (s *FlinkService) CreateNamespace(request *foasconsole.CreateNamespaceRequest) (*foasconsole.CreateNamespaceResponse, error) {
	response, err := s.foasconsoleClient.CreateNamespace(request)
	return response, WrapError(err)
}

func (s *FlinkService) DeleteNamespace(request *foasconsole.DeleteNamespaceRequest) (*foasconsole.DeleteNamespaceResponse, error) {
	response, err := s.foasconsoleClient.DeleteNamespace(request)
	return response, WrapError(err)
}

func (s *FlinkService) ListNamespaces(workspace string) ([]*foasconsole.DescribeNamespacesResponseBodyNamespaces, error) {
	region := s.GetRegionId()
	pageIndex := int32(1)
	pageSize := int32(50)
	var namespaces []*foasconsole.DescribeNamespacesResponseBodyNamespaces
	for {
		request := &foasconsole.DescribeNamespacesRequest{
			PageIndex:  tea.Int32(pageIndex),
			PageSize:   tea.Int32(pageSize),
			Region:     tea.String(region),
			InstanceId: tea.String(workspace),
		}
		response, err := s.foasconsoleClient.DescribeNamespaces(request)
		if err != nil {
			return nil, WrapError(err)
		}

		namespaces = append(namespaces, response.Body.Namespaces...)

		if *response.Body.PageIndex >= *response.Body.TotalPage {
			break
		}
		pageIndex++
	}
	return namespaces, nil
}

func (s *FlinkService) GetNamespace(workspace string, namespace string) (*foasconsole.DescribeNamespacesResponseBodyNamespaces, error) {
	region := s.GetRegionId()
	request := &foasconsole.DescribeNamespacesRequest{
		Region:     tea.String(region),
		InstanceId: tea.String(workspace),
		Namespace:  tea.String(namespace),
	}
	response, err := s.foasconsoleClient.DescribeNamespaces(request)
	if err != nil {
		return nil, WrapError(err)
	}
	if len(response.Body.Namespaces) == 0 {
		return nil, fmt.Errorf("namespace '%s/%s' not found", workspace, namespace)
	}

	return response.Body.Namespaces[0], nil
}

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

func (s *FlinkService) CreateVariable(workspace, namespace *string, variable *ververica.Variable) (*ververica.Variable, error) {
	resp, err := s.ververicaClient.CreateVariableWithOptions(
		namespace,
		&ververica.CreateVariableRequest{
			Body: variable,
		},
		&ververica.CreateVariableHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.Body.Data, nil
}

func (s *FlinkService) UpdateVariable(workspace, namespace, varName *string, variable *ververica.Variable) (*ververica.Variable, error) {
	resp, err := s.ververicaClient.UpdateVariableWithOptions(
		namespace,
		varName,
		&ververica.UpdateVariableRequest{
			Body: variable,
		},
		&ververica.UpdateVariableHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp.Body.Data, nil
}

func (s *FlinkService) DeleteVariable(workspace, namespace, varName *string) error {
	_, err := s.ververicaClient.DeleteVariableWithOptions(
		namespace,
		varName,
		&ververica.DeleteVariableHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{})
	if err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *FlinkService) GetVariable(workspace, namespace, varName *string) (*ververica.Variable, error) {
	var pageIndex int32
	var pageSize int32

	pageIndex = 1
	pageSize = 50
	for {
		resp, err := s.ListVariables(workspace, namespace, pageIndex, pageSize)
		if err != nil {
			return nil, WrapError(err)
		}
		if resp.Body.Data == nil || len(resp.Body.Data) == 0 {
			return nil, fmt.Errorf("variable not found")
		}

		variableList := resp.Body.Data
		for _, variable := range variableList {
			if *variable.Name == *varName {
				return variable, nil
			}
		}
		pageIndex += 1
	}
}

func (s *FlinkService) ListVariables(workspace, namespace *string, PageIndex int32, PageSize int32) (*ververica.ListVariablesResponse, error) {
	resp, err := s.ververicaClient.ListVariablesWithOptions(
		namespace,
		&ververica.ListVariablesRequest{
			PageIndex: tea.Int32(PageIndex),
			PageSize:  tea.Int32(PageSize),
		},
		&ververica.ListVariablesHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	if err != nil {
		return nil, WrapError(err)
	}
	return resp, nil
}

// Connector management functions
func (s *FlinkService) ListCustomConnectors(workspace, namespace *string) (*ververica.ListCustomConnectorsResponse, error) {
	response, err := s.ververicaClient.ListCustomConnectorsWithOptions(
		namespace,
		&ververica.ListCustomConnectorsHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	return response, WrapError(err)
}

func (s *FlinkService) RegisterCustomConnector(workspace, namespace *string, request *ververica.RegisterCustomConnectorRequest) (*ververica.RegisterCustomConnectorResponse, error) {
	response, err := s.ververicaClient.RegisterCustomConnectorWithOptions(
		namespace,
		request,
		&ververica.RegisterCustomConnectorHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	return response, WrapError(err)
}

func (s *FlinkService) DeleteCustomConnector(workspace, namespace *string, connectorName *string) (*ververica.DeleteCustomConnectorResponse, error) {
	response, err := s.ververicaClient.DeleteCustomConnectorWithOptions(
		namespace,
		connectorName,
		&ververica.DeleteCustomConnectorHeaders{
			Workspace: workspace,
		},
		&util.RuntimeOptions{},
	)
	return response, WrapError(err)
}

func (s *FlinkService) GetConnector(workspace, namespace, connectorName *string) (*ververica.Connector, error) {
	response, err := s.ListCustomConnectors(workspace, namespace)
	if err != nil {
		return nil, WrapError(err)
	}

	if response == nil || response.Body == nil || response.Body.Data == nil {
		return nil, WrapErrorf(fmt.Errorf("connector not found"), "connector %s not found", *connectorName)
	}

	connectors := response.Body.Data
	for _, connector := range connectors {
		if connector.Name != nil && *connector.Name == *connectorName {
			return connector, nil
		}
	}

	return nil, WrapErrorf(fmt.Errorf("connector not found"), "connector %s not found", *connectorName)
}
