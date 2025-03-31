package alicloud

import (
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type FoasService struct {
	*connectivity.AliyunClient
}

// FoasService Instance related functions
func (s *FoasService) CreateInstance(request *foasconsole.CreateInstanceRequest) (*foasconsole.CreateInstanceResponse, error) {
	var response *foasconsole.CreateInstanceResponse
	err := s.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		var err error
		response, err = client.CreateInstance(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FoasService) DescribeInstance(instanceId string) (*foasconsole.DescribeInstancesResponseInstance, error) {
	var result *foasconsole.DescribeInstancesResponseInstance
	err := s.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		request := foasconsole.CreateDescribeInstancesRequest()
		request.InstanceId = instanceId
		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}
		for _, inst := range response.Instances {
			if inst.InstanceId == instanceId {
				result = &inst
				return result, nil
			}
		}
		return nil, WrapErrorf(Error("instance %s not found", instanceId))
	})
	return result, WrapError(err)
}

// FoasService Namespace related functions
func (s *FoasService) CreateNamespace(request *foasconsole.CreateNamespaceRequest) (*foasconsole.CreateNamespaceResponse, error) {
	var response *foasconsole.CreateNamespaceResponse
	err := s.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		var err error
		response, err = client.CreateNamespace(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *FoasService) DescribeNamespace(workspaceId string, namespaceId string) (*foasconsole.DescribeNamespacesResponseNamespace, error) {
	var result *foasconsole.DescribeNamespacesResponseNamespace
	err := s.WithFoasClient(func(client *foasconsole.Client) (interface{}, error) {
		request := foasconsole.CreateDescribeNamespacesRequest()
		request.WorkspaceId = workspaceId
		request.NamespaceId = namespaceId
		response, err := client.DescribeNamespaces(request)
		if err != nil {
			return nil, err
		}
		for _, ns := range response.Namespaces {
			if ns.NamespaceId == namespaceId {
				result = &ns
				return result, nil
			}
		}
		return nil, WrapErrorf(Error("namespace %s not found", namespaceId))
	})
	return result, WrapError(err)
}

// VervericaService deployment related functions
func (s *VervericaService) DeployJob(request *ververica.DeployJobRequest) (*ververica.DeployJobResponse, error) {
	var response *ververica.DeployJobResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.DeployJob(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *VervericaService) GetDeployment(request *ververica.GetDeploymentRequest) (*ververica.GetDeploymentResponse, error) {
	var response *ververica.GetDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.GetDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *VervericaService) UpdateDeployment(request *ververica.UpdateDeploymentRequest) (*ververica.UpdateDeploymentResponse, error) {
	var response *ververica.UpdateDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.UpdateDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *VervericaService) DeleteDeployment(request *ververica.DeleteDeploymentRequest) (*ververica.DeleteDeploymentResponse, error) {
	var response *ververica.DeleteDeploymentResponse
	err := s.WithVervericaClient(func(client *ververica.Client) (interface{}, error) {
		var err error
		response, err = client.DeleteDeployment(request)
		return response, err
	})
	return response, WrapError(err)
}

// VervericaService other functions
func (s *VervericaService) CreateSessionCluster(request *client.CreateSessionClusterRequest) (*client.CreateSessionClusterResponse, error) {
	var response *client.CreateSessionClusterResponse
	err := s.WithVervericaClient(func(client *client.Client) (interface{}, error) {
		var err error
		response, err = client.CreateSessionCluster(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *VervericaService) DeleteSessionCluster(request *client.DeleteSessionClusterRequest) (*client.DeleteSessionClusterResponse, error) {
	var response *client.DeleteSessionClusterResponse
	err := s.WithVervericaClient(func(client *client.Client) (interface{}, error) {
		var err error
		response, err = client.DeleteSessionCluster(request)
		return response, err
	})
	return response, WrapError(err)
}

func (s *VervericaService) DescribeNamespaces(namespace *string) ([]*client.GetDatabasesResponseDatabase, error) {
	request := client.CreateGetDatabasesRequest()
	request.Namespace = namespace

	var response *client.GetDatabasesResponse
	err := s.AliyunClient.WithVervericaClient(func(client *client.Client) (interface{}, error) {
		var err error
		response, err = client.GetDatabases(request)
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return response.Databases, nil
}