package alicloud

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
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

// 新增方法: CreateMember
func (s *FlinkService) CreateMember(request *ververica.CreateMemberRequest) (*ververica.CreateMemberResponse, error) {
    return s.ververicaClient.CreateMember(request)
}

// 新增方法: DescribeMember
func (s *FlinkService) DescribeMember(request *ververica.DescribeMemberRequest) (*ververica.DescribeMemberResponse, error) {
    return s.ververicaClient.DescribeMember(request)
}

// 新增方法: UpdateMember
func (s *FlinkService) UpdateMember(request *ververica.UpdateMemberRequest) (*ververica.UpdateMemberResponse, error) {
    return s.ververicaClient.UpdateMember(request)
}

// 新增方法: DeleteMember
func (s *FlinkService) DeleteMember(request *ververica.DeleteMemberRequest) (*ververica.DeleteMemberResponse, error) {
    return s.ververicaClient.DeleteMember(request)
}