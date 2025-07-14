package alicloud

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ram"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeRdsServiceLinkedRole describes RDS service linked role
func (s *RdsService) DescribeRdsServiceLinkedRole(id string) (*ram.GetRoleResponse, error) {
	response := &ram.GetRoleResponse{}
	request := ram.CreateGetRoleRequest()
	request.RegionId = s.client.RegionId
	request.RoleName = id
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, err := s.client.WithRamClient(func(ramClient *ram.Client) (interface{}, error) {
			return ramClient.GetRole(request)
		})
		if err != nil {
			if IsExpectedErrors(err, []string{ThrottlingUser}) {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(request.GetActionName(), raw, request.RpcRequest, request)
		response, _ = raw.(*ram.GetRoleResponse)
		return nil
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"EntityNotExist.Role"}) {
			return response, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	return response, nil
}
