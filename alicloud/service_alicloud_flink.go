package alicloud

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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

// Added DescribeSupportedZones method
func (s *FlinkService) DescribeSupportedZones(request *foasconsole.DescribeSupportedZonesRequest) (*foasconsole.DescribeSupportedZonesResponse, error) {
	response, err := s.foasconsoleClient.DescribeSupportedZones(request)
	return response, WrapError(err)
}

// Instance management functions added
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

// Namespace management functions added
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

// 新增方法: CreateMember
func (s *FlinkService) CreateMember(namespace *string, request *ververica.CreateMemberRequest) (*ververica.CreateMemberResponse, error) {
    response, err := s.ververicaClient.CreateMember(namespace, request)
    return response, WrapError(err)
}

// 新增方法: DescribeMember
func (s *FlinkService) GetMember(namespace *string, member *string) (*ververica.GetMemberResponse, error) {
    return s.ververicaClient.GetMember( namespace, member)
}

// 新增方法: UpdateMember
func (s *FlinkService) UpdateMember(namespace *string, request *ververica.UpdateMemberRequest) (*ververica.UpdateMemberResponse, error) {
    response, err := s.ververicaClient.UpdateMember(namespace, request)
    return response, WrapError(err)
}

// 新增方法: DeleteMember
func (s *FlinkService) DeleteMember(namespace *string, member *string) (*ververica.DeleteMemberResponse, error) {
    return s.ververicaClient.DeleteMember(namespace, member)
}

// Deployment management functions added
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

// Job management for deployments
func (s *FlinkService) StartJobWithParams(namespace *string, request *ververica.StartJobWithParamsRequest) (*ververica.StartJobWithParamsResponse, error) {
    response, err := s.ververicaClient.StartJobWithParams(namespace, request)
    return response, WrapError(err)
}

func (s *FlinkService) StopJob(namespace *string, jobId *string, request *ververica.StopJobRequest) (*ververica.StopJobResponse, error) {
    response, err := s.ververicaClient.StopJob(namespace, jobId, request)
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
		
		if object.Body.Deployment.JobSummary != nil {
			return object, *object.Body.Deployment.JobSummary.Status, nil
		}
		
		return object, "CREATED", nil
	}
}

// FlinkJobStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Flink job
func (s *FlinkService) FlinkJobStateRefreshFunc(namespace, jobId string, failStates []string) ResourceStateRefreshFunc {
	return func() (interface{}, string, error) {
		response, err := s.ververicaClient.GetJob(&namespace, &jobId)
		if err != nil {
			if IsExpectedErrors(err, []string{"JobNotFound"}) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		
		return response, *response.Body.Job.Status, nil
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
