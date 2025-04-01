package alicloud

import (
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type FlinkService struct {
	client            *connectivity.AliyunClient
	foasconsoleClient *foasconsole.Client
	ververicaClient   *ververica.Client
}

func NewFlinkService(client *connectivity.AliyunClient) (*FlinkService, error) {
	accessKey, secretKey, securityToken := client.GetRefreshCredential()
	config := &openapi.Config{
		AccessKeyId:     accessKey,
		AccessKeySecret: secretKey,
		RegionId:        client.RegionId,
		SecurityToken:   securityToken,
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

func (s *FlinkService) UpdateInstance(request *foasconsole.UpdateInstanceRequest) (*foasconsole.UpdateInstanceResponse, error) {
	response, err := s.foasconsoleClient.UpdateInstance(request)
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

func (s *FlinkService) UpdateNamespace(request *foasconsole.UpdateNamespaceRequest) (*foasconsole.UpdateNamespaceResponse, error) {
	response, err := s.foasconsoleClient.UpdateNamespace(request)
	return response, WrapError(err)
}

// State refresh functions
func (s *FlinkService) FlinkDeploymentStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parts, err := ParseResourceId(id, 3)
		if err != nil {
			return nil, "", WrapError(err)
		}
		request := foasconsole.CreateGetDeploymentRequest()
		request.WorkspaceId = parts[0]
		request.Namespace = parts[1]
		request.DeploymentId = parts[2]

		raw, err := s.GetDeployment(request)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		return raw.Deployment, raw.Deployment.Status, nil
	}
}

// Job management moved from FlinkService to FlinkService
func (s *FlinkService) DeployJob(request *ververica.DeployJobRequest) (*ververica.DeployJobResponse, error) {
	var response *ververica.DeployJobResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.DeployJob(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FlinkService) GetDeployment(request *ververica.GetDeploymentRequest) (*ververica.GetDeploymentResponse, error) {
	var response *ververica.GetDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.GetDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FlinkService) UpdateDeployment(request *ververica.UpdateDeploymentRequest) (*ververica.UpdateDeploymentResponse, error) {
	var response *ververica.UpdateDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.UpdateDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FlinkService) DeleteDeployment(request *ververica.DeleteDeploymentRequest) (*ververica.DeleteDeploymentResponse, error) {
	var response *ververica.DeleteDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.DeleteDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

// FlinkService other functions
func (s *FlinkService) CreateSessionCluster(request *client.CreateSessionClusterRequest) (*client.CreateSessionClusterResponse, error) {
	var response *client.CreateSessionClusterResponse
	err := s.WithVervericaClient(func(client *client.Client) (interface{}, error) {
		var err error
		response, err = client.CreateSessionCluster(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FlinkService) DeleteSessionCluster(request *client.DeleteSessionClusterRequest) (*client.DeleteSessionClusterResponse, error) {
	var response *client.DeleteSessionClusterResponse
	err := s.WithVervericaClient(func(client *client.Client) (interface{}, error) {
		var err error
		response, err = client.DeleteSessionCluster(request)
		return response, err
	})
	return response, WrapError(err)
}
