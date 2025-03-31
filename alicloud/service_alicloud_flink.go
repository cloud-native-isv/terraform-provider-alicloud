package alicloud

import (
	foasconsole "github.com/alibabacloud-go/foasconsole-20211028/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type FoasService struct {
	client *connectivity.AliyunClient
}

func (s *FoasService) DeployJob(request *foasconsole.DeployJobRequest) (*foasconsole.DeployJobResponse, error) {
	conn, err := s.client.NewFoasconsoleClient()
	if err != nil {
		return nil, WrapError(err)
	}
	response, err := conn.DeployJob(request)
	if err != nil {
		return nil, WrapError(err)
	}
	return response, nil
}

func (s *FoasService) GetDeployment(request *foasconsole.GetDeploymentRequest) (*foasconsole.GetDeploymentResponse, error) {
	conn, err := s.client.NewFoasconsoleClient()
	if err != nil {
		return nil, WrapError(err)
	}
	response, err := conn.GetDeployment(request)
	if err != nil {
		return nil, WrapError(err)
	}
	return response, nil
}

func (s *FoasService) UpdateDeployment(request *foasconsole.UpdateDeploymentRequest) (*foasconsole.UpdateDeploymentResponse, error) {
	conn, err := s.client.NewFoasconsoleClient()
	if err != nil {
		return nil, WrapError(err)
	}
	response, err := conn.UpdateDeployment(request)
	if err != nil {
		return nil, WrapError(err)
	}
	return response, nil
}

func (s *FoasService) DeleteDeployment(request *foasconsole.DeleteDeploymentRequest) (*foasconsole.DeleteDeploymentResponse, error) {
	conn, err := s.client.NewFoasconsoleClient()
	if err != nil {
		return nil, WrapError(err)
	}
	response, err := conn.DeleteDeployment(request)
	if err != nil {
		return nil, WrapError(err)
	}
	return response, nil
}

func (s *FoasService) FlinkDeploymentStateRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parts, err := ParseResourceId(id, 3)
		if err != nil {
			return nil, "", WrapError(err)
		}
		workspaceId := parts[0]
		namespace := parts[1]
		deploymentId := parts[2]

		request := foasconsole.CreateGetDeploymentRequest()
		request.WorkspaceId = workspaceId
		request.Namespace = namespace
		request.DeploymentId = deploymentId

		raw, err := s.GetDeployment(request)
		if err != nil {
			if NotFoundError(err) {
				// deployment not found
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		if raw.Deployment == nil {
			return nil, "", nil
		}

		return raw.Deployment, raw.Deployment.Status, nil
	}
}
