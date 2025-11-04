package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunVpcAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// VpcNatGatewayService provides operations for NAT Gateway via CWS-Lib-Go API layer.
type VpcNatGatewayService struct {
	client *connectivity.AliyunClient
	vpcAPI *aliyunVpcAPI.VpcAPI
}

// NewVpcNatGatewayService creates a new VpcNatGatewayService bound to the given AliyunClient.
func NewVpcNatGatewayService(client *connectivity.AliyunClient) (*VpcNatGatewayService, error) {
	// Initialize cws-lib-go VPC API client
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	vpcAPI, err := aliyunVpcAPI.NewVpcAPI(credentials)
	if err != nil {
		return nil, WrapError(err)
	}

	return &VpcNatGatewayService{client: client, vpcAPI: vpcAPI}, nil
}

// GetAPI returns the underlying cws-lib-go VPC API client.
func (s *VpcNatGatewayService) GetAPI() *aliyunVpcAPI.VpcAPI {
	return s.vpcAPI
}

// CreateNatGateway creates a NAT Gateway and returns its Id.
func (s *VpcNatGatewayService) CreateNatGateway(request map[string]interface{}, timeout time.Duration) (string, error) {
	if request == nil {
		request = make(map[string]interface{})
	}

	// Convert map to struct for CWS-Lib-Go API
	createRequest := &aliyunVpcAPI.CreateNatGatewayRequest{
		RegionId: s.client.RegionId,
	}

	if v, ok := request["Description"]; ok {
		createRequest.Description = v.(string)
	}
	if v, ok := request["InternetChargeType"]; ok {
		createRequest.InternetChargeType = v.(string)
	}
	if v, ok := request["Name"]; ok {
		createRequest.Name = v.(string)
	}
	if v, ok := request["NatType"]; ok {
		createRequest.NatType = v.(string)
	}
	if v, ok := request["InstanceChargeType"]; ok {
		createRequest.InstanceChargeType = v.(string)
	}
	if v, ok := request["Duration"]; ok {
		createRequest.Duration = v.(string)
	}
	if v, ok := request["PricingCycle"]; ok {
		createRequest.PricingCycle = v.(string)
	}
	if v, ok := request["AutoPay"]; ok {
		createRequest.AutoPay = v.(bool)
	}
	if v, ok := request["Spec"]; ok {
		createRequest.Spec = v.(string)
	}
	if v, ok := request["VSwitchId"]; ok {
		createRequest.VSwitchId = v.(string)
	}
	if v, ok := request["NetworkType"]; ok {
		createRequest.NetworkType = v.(string)
	}
	if v, ok := request["VpcId"]; ok {
		createRequest.VpcId = v.(string)
	}
	if v, ok := request["EipBindMode"]; ok {
		createRequest.EipBindMode = v.(string)
	}
	if v, ok := request["IcmpReplyEnabled"]; ok {
		createRequest.IcmpReplyEnabled = v.(bool)
	}
	if v, ok := request["PrivateLinkEnabled"]; ok {
		createRequest.PrivateLinkEnabled = v.(bool)
	}
	if v, ok := request["AccessMode"]; ok {
		createRequest.AccessMode = v.(string)
	}

	// Set ClientToken for idempotency
	createRequest.ClientToken = buildClientToken("CreateNatGateway")

	var natGatewayId string
	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		response, err := s.vpcAPI.CreateNatGateway(createRequest)
		if err != nil {
			// Retry on common retryable errors
			if NeedRetry(err) || IsExpectedErrors(err, []string{"InternalError", "ServiceUnavailable", "SystemBusy", "Throttling", "OperationConflict", "TaskConflict", "VswitchStatusError", "IncorrectStatus.VSWITCH"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if response != nil && response.NatGatewayId != "" {
			natGatewayId = response.NatGatewayId
		}
		return nil
	})

	if err != nil {
		return "", WrapErrorf(err, DefaultErrorMsg, "alicloud_nat_gateway", "CreateNatGateway", AlibabaCloudSdkGoERROR)
	}

	if natGatewayId == "" {
		return "", WrapError(Error("CreateNatGateway response missing NatGatewayId"))
	}

	return natGatewayId, nil
}

// DescribeNatGateway returns NAT Gateway attributes.
func (s *VpcNatGatewayService) DescribeNatGateway(id string) (map[string]interface{}, error) {
	if id == "" {
		return nil, WrapError(Error("NatGatewayId is empty"))
	}

	var response *aliyunVpcAPI.NatGateway
	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := s.vpcAPI.DescribeNatGateway(id)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		return nil
	})

	if err != nil {
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"InvalidNatGatewayId.NotFound", "InvalidRegionId.NotFound"}) {
			return nil, WrapErrorf(NotFoundErr("NatGateway", id), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeNatGateway", AlibabaCloudSdkGoERROR)
	}

	if response == nil {
		return nil, WrapErrorf(NotFoundErr("NatGateway", id), NotFoundMsg, ProviderERROR)
	}

	// Convert struct to map for compatibility
	result := make(map[string]interface{})
	result["NatGatewayId"] = response.NatGatewayId
	result["Description"] = response.Description
	result["InternetChargeType"] = response.InternetChargeType
	result["Name"] = response.Name
	result["NatType"] = response.NatType
	result["InstanceChargeType"] = response.InstanceChargeType
	result["Spec"] = response.Spec
	result["Status"] = response.Status
	result["VpcId"] = response.VpcId
	result["EipBindMode"] = response.EipBindMode
	result["DeletionProtection"] = response.DeletionProtection
	result["IcmpReplyEnabled"] = response.IcmpReplyEnabled
	result["PrivateLinkEnabled"] = response.PrivateLinkEnabled
	result["AccessMode"] = response.AccessMode

	// Handle ForwardTableIds and SnatTableIds if they exist
	if response.ForwardTableIds != nil {
		result["ForwardTableIds"] = map[string]interface{}{
			"ForwardTableId": response.ForwardTableIds,
		}
	}
	if response.SnatTableIds != nil {
		result["SnatTableIds"] = map[string]interface{}{
			"SnatTableId": response.SnatTableIds,
		}
	}

	// Handle NatGatewayPrivateInfo
	if response.VSwitchId != "" {
		result["NatGatewayPrivateInfo"] = map[string]interface{}{
			"VswitchId": response.VSwitchId,
		}
	}

	return result, nil
}

// ModifyNatGatewayAttribute updates attributes such as name/description.
func (s *VpcNatGatewayService) ModifyNatGatewayAttribute(id string, attrs map[string]interface{}, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("NatGatewayId is empty"))
	}

	modifyRequest := &aliyunVpcAPI.ModifyNatGatewayAttributeRequest{
		NatGatewayId: id,
		RegionId:     s.client.RegionId,
	}

	if v, ok := attrs["Description"]; ok {
		modifyRequest.Description = v.(string)
	}
	if v, ok := attrs["Name"]; ok {
		modifyRequest.Name = v.(string)
	}
	if v, ok := attrs["EipBindMode"]; ok {
		modifyRequest.EipBindMode = v.(string)
	}
	if v, ok := attrs["IcmpReplyEnabled"]; ok {
		modifyRequest.IcmpReplyEnabled = v.(bool)
	}

	var err error
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		_, err := s.vpcAPI.ModifyNatGatewayAttribute(modifyRequest)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"IncorrectStatus.NATGW", "OperationConflict"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "ModifyNatGatewayAttribute", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ModifyNatGatewaySpec updates the NAT Gateway specification.
func (s *VpcNatGatewayService) ModifyNatGatewaySpec(id string, spec string, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("NatGatewayId is empty"))
	}

	modifyRequest := &aliyunVpcAPI.ModifyNatGatewaySpecRequest{
		NatGatewayId: id,
		RegionId:     s.client.RegionId,
		Spec:         spec,
	}

	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		_, err := s.vpcAPI.ModifyNatGatewaySpec(modifyRequest)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"IncorrectStatus.NATGW", "OperationConflict"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "ModifyNatGatewaySpec", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNatGateway deletes a NAT Gateway.
func (s *VpcNatGatewayService) DeleteNatGateway(id string, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("NatGatewayId is empty"))
	}

	deleteRequest := &aliyunVpcAPI.DeleteNatGatewayRequest{
		NatGatewayId: id,
		RegionId:     s.client.RegionId,
	}

	// Note: Force parameter is not included in CWS-Lib-Go API yet
	// If needed, we can add it later when the API supports it

	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		_, err := s.vpcAPI.DeleteNatGateway(deleteRequest)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"IncorrectStatus.NATGW", "OperationConflict", "DependencyViolation.BandwidthPackages", "DependencyViolation.EIPS"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		// tolerate not found
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"INSTANCE_NOT_EXISTS", "IncorrectStatus.NatGateway", "InvalidNatGatewayId.NotFound", "InvalidRegionId.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, id, "DeleteNatGateway", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// NatGatewayStateRefreshFunc polls NAT Gateway status.
func (s *VpcNatGatewayService) NatGatewayStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeNatGateway(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		status := ""
		if obj != nil {
			if v, ok := obj["Status"]; ok {
				status = Interface2String(v)
			}
		}
		for _, fs := range failStates {
			if status == fs {
				return obj, status, WrapError(Error(FailedToReachTargetStatus, status))
			}
		}
		return obj, status, nil
	}
}

// WaitForNatGatewayCreating waits for NAT GW to reach Available.
func (s *VpcNatGatewayService) WaitForNatGatewayCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Modifying"},
		[]string{"Available"},
		timeout,
		5*time.Second,
		s.NatGatewayStateRefreshFunc(id, []string{"Deleting", "Failed"}),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// WaitForNatGatewayDeleting waits for NAT GW to be removed (Describe returns not found).
func (s *VpcNatGatewayService) WaitForNatGatewayDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting"},
		[]string{""},
		timeout,
		5*time.Second,
		s.NatGatewayStateRefreshFunc(id, []string{}),
	)
	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
