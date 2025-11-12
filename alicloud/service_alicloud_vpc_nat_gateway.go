package alicloud

import (
	"encoding/json"
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

	// Build domain object for CWS-Lib-Go API (strong-typed)
	nat := &aliyunVpcAPI.NATGateway{}
	// Required fields
	nat.RegionId = s.client.RegionId
	if v, ok := request["VpcId"]; ok {
		nat.VpcId = v.(string)
	}
	if v, ok := request["VSwitchId"]; ok {
		nat.VSwitchId = v.(string)
	}
	// Optional fields
	if v, ok := request["Name"]; ok {
		nat.Name = v.(string)
	}
	if v, ok := request["Description"]; ok {
		nat.Description = v.(string)
	}
	if v, ok := request["NatType"]; ok {
		nat.NatType = v.(string)
	}
	if v, ok := request["NetworkType"]; ok {
		nat.NetworkType = v.(string)
	}
	if v, ok := request["Spec"]; ok {
		nat.Spec = v.(string)
	}
	if v, ok := request["EipBindMode"]; ok {
		nat.EipBindMode = v.(string)
	}
	if v, ok := request["AccessMode"]; ok {
		accessModeJson := v.(string)
		if accessModeJson != "" {
			var accessModeMap map[string]interface{}
			err := json.Unmarshal([]byte(accessModeJson), &accessModeMap)
			if err != nil {
				return "", WrapErrorf(Error("Failed to unmarshal AccessMode JSON: %v", err), DefaultErrorMsg, "alicloud_nat_gateway", "CreateNatGateway", AlibabaCloudSdkGoERROR)
			}
			accessMode := &aliyunVpcAPI.AccessMode{}
			if modeValue, ok := accessModeMap["ModeValue"]; ok && modeValue != nil && modeValue.(string) != "" {
				accessMode.ModeValue = modeValue.(string)
			}
			if tunnelType, ok := accessModeMap["TunnelType"]; ok && tunnelType != nil && tunnelType.(string) != "" {
				accessMode.TunnelType = tunnelType.(string)
			}
			nat.AccessMode = accessMode
		}
	}
	if v, ok := request["PrivateLinkEnabled"]; ok {
		nat.PrivateLinkEnabled = v.(bool)
	}
	if v, ok := request["IcmpReplyEnabled"]; ok {
		nat.IcmpReplyEnabled = v.(bool)
	}
	if v, ok := request["DeletionProtection"]; ok {
		nat.DeletionProtection = v.(bool)
	}

	var natGatewayId string
	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		response, err := s.vpcAPI.CreateNatGateway(nat)
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

// DescribeNatGateway returns NAT Gateway attributes using CWS-Lib-Go strong typing.
func (s *VpcNatGatewayService) DescribeNatGateway(id string) (*aliyunVpcAPI.NATGateway, error) {
	if id == "" {
		return nil, WrapError(Error("NatGatewayId is empty"))
	}

	var response *aliyunVpcAPI.NATGateway
	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := s.vpcAPI.GetNatGateway(s.client.RegionId, id)
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

	return response, nil
}

// ModifyNatGatewayAttribute updates attributes such as name/description.
func (s *VpcNatGatewayService) ModifyNatGatewayAttribute(id string, attrs map[string]interface{}, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("NatGatewayId is empty"))
	}

	// Extract fields for Modify API
	var name, description, eipBindMode string
	var icmpPtr *bool
	if v, ok := attrs["Name"]; ok {
		name = v.(string)
	}
	if v, ok := attrs["Description"]; ok {
		description = v.(string)
	}
	if v, ok := attrs["EipBindMode"]; ok {
		eipBindMode = v.(string)
	}
	if v, ok := attrs["IcmpReplyEnabled"]; ok {
		b := v.(bool)
		icmpPtr = &b
	}

	var err error
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		err := s.vpcAPI.ModifyNatGateway(s.client.RegionId, id, name, description, eipBindMode, *icmpPtr)
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

	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		err := s.vpcAPI.ModifyNatGatewaySpec(s.client.RegionId, id, spec, "")
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

	// Force deletion flag (default false)
	force := false

	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		err := s.vpcAPI.DeleteNatGateway(s.client.RegionId, id, force)
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
			status = obj.Status
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
