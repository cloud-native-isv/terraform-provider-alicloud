package alicloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunVpcAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// VpcEipService provides operations for EIP resources via CWS-Lib-Go API layer.
// NOTE: All cloud calls must be implemented using github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc
// Do NOT call SDK/RPC directly from this layer.
type VpcEipService struct {
	client *connectivity.AliyunClient
	vpcAPI *aliyunVpcAPI.VpcAPI
}

// NewVpcEipService creates a new VpcEipService bound to the given AliyunClient.
func NewVpcEipService(client *connectivity.AliyunClient) (*VpcEipService, error) {
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

	return &VpcEipService{client: client, vpcAPI: vpcAPI}, nil
}

// GetAPI returns the underlying cws-lib-go VPC API client.
// This allows strongly-typed operations once the migration is complete.
func (s *VpcEipService) GetAPI() *aliyunVpcAPI.VpcAPI {
	return s.vpcAPI
}

// AllocateEipAddress allocates a new EIP and returns its AllocationId.
// request should contain fields such as Bandwidth, InternetChargeType, Name, Description, Tags, etc.
func (s *VpcEipService) AllocateEipAddress(request map[string]interface{}) (string, error) {
	// TEMP: Use legacy RPC via AliyunClient, wrapped in Service layer.
	// TODO(cws-lib-go): Replace with cws-lib-go VPC API AllocateEipAddress implementation.
	client := s.client
	action := "AllocateEipAddress"

	if request == nil {
		request = make(map[string]interface{})
	}
	// RegionId is required by RPC
	request["RegionId"] = client.RegionId
	request["ClientToken"] = buildClientToken(action)

	var (
		resp  map[string]interface{}
		query map[string]interface{}
		err   error
	)
	query = make(map[string]interface{})

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err = client.RpcPost("Vpc", "2016-04-28", action, query, request, true)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "LastTokenProcessing", "IncorrectStatus", "SystemBusy", "ServiceUnavailable", "FrequentPurchase.EIP"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, resp, request)
		return nil
	})
	if err != nil {
		return "", WrapErrorf(err, DefaultErrorMsg, "alicloud_eip_address", action, AlibabaCloudSdkGoERROR)
	}

	allocationId := fmt.Sprint(resp["AllocationId"])
	if allocationId == "" || allocationId == "<nil>" {
		return "", WrapError(Error("AllocateEipAddress missing AllocationId in response"))
	}
	return allocationId, nil
}

// DescribeEipAddress returns the EIP attributes.
func (s *VpcEipService) DescribeEipAddress(allocationId string) (map[string]interface{}, error) {
	// Preferred path: try CWS-Lib-Go VPC API via common method names.
	if s.vpcAPI != nil && allocationId != "" {
		// Use the proper DescribeEipAddresses method with RegionId
		request := &aliyunVpcAPI.DescribeEipRequest{
			RegionId:     s.client.RegionId,
			AllocationId: allocationId,
		}

		eips, err := s.vpcAPI.DescribeEipAddresses(request)
		if err != nil {
			// If not found, fall back to legacy below; otherwise wrap and return
			if IsNotFoundError(err) || strings.Contains(strings.ToLower(err.Error()), "not found") {
				// Fall through to legacy implementation
			} else {
				return nil, WrapError(err)
			}
		} else if len(eips) > 0 {
			// Convert the first EIP to map[string]interface{}
			converted := convertToStringInterfaceMap(eips[0])
			if converted != nil {
				return converted, nil
			}
		}
	}

	// Fallback: legacy V2 implementation to ensure compatibility
	eipSvc := EipServiceV2{s.client}
	obj, err := eipSvc.DescribeEipAddress(allocationId)
	if err != nil {
		return nil, WrapError(err)
	}
	return obj, nil
}

// ModifyEipAddressAttribute modifies mutable attributes of the EIP.
func (s *VpcEipService) ModifyEipAddressAttribute(allocationId string, attrs map[string]interface{}) error {
	// TEMP: Use legacy RPC
	client := s.client
	action := "ModifyEipAddressAttribute"
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	// Required
	attrs["RegionId"] = client.RegionId

	var (
		resp  map[string]interface{}
		query map[string]interface{}
		err   error
	)
	query = map[string]interface{}{
		"AllocationId": allocationId,
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err = client.RpcPost("Vpc", "2016-04-28", action, query, attrs, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "LastTokenProcessing", "IncorrectStatus", "SystemBusy", "ServiceUnavailable", "IncorrectEipStatus", "IncorrectStatus.ResourceStatus", "VPC_TASK_CONFLICT"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, resp, attrs)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, allocationId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ReleaseEipAddress releases the EIP.
func (s *VpcEipService) ReleaseEipAddress(allocationId string) error {
	// TEMP: Use legacy RPC
	client := s.client
	action := "ReleaseEipAddress"
	req := map[string]interface{}{
		"RegionId": client.RegionId,
	}
	query := map[string]interface{}{
		"AllocationId": allocationId,
	}
	var (
		resp map[string]interface{}
		err  error
	)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err = client.RpcPost("Vpc", "2016-04-28", action, query, req, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "LastTokenProcessing", "IncorrectStatus", "SystemBusy", "ServiceUnavailable", "IncorrectEipStatus", "TaskConflict.AssociateGlobalAccelerationInstance"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, resp, req)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, allocationId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// AssociateEipAddress binds the EIP to a given instance.
func (s *VpcEipService) AssociateEipAddress(allocationId, instanceId, instanceType, privateIp string) error {
	// TEMP: Use legacy RPC
	client := s.client
	action := "AssociateEipAddress"
	req := map[string]interface{}{
		"RegionId":     client.RegionId,
		"InstanceType": instanceType,
		"AllocationId": allocationId,
	}
	if privateIp != "" {
		req["PrivateIpAddress"] = privateIp
	}
	query := map[string]interface{}{
		"AllocationId": allocationId,
	}
	var (
		resp map[string]interface{}
		err  error
	)

	// Depending on instance type, need to set different identifier fields
	// Commonly: InstanceId or Enni/EniId for ENI. Keep generic for now.
	if instanceId != "" {
		req["InstanceId"] = instanceId
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		resp, err = client.RpcPost("Vpc", "2016-04-28", action, query, req, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "LastTokenProcessing", "IncorrectStatus", "SystemBusy", "ServiceUnavailable", "IncorrectEipStatus", "IncorrectInstanceStatus", "InvalidBindingStatus", "IncorrectStatus.NatGateway", "InvalidStatus.EcsStatusNotSupport", "InvalidStatus.InstanceHasBandWidth", "InvalidStatus.EniStatusNotSupport", "TaskConflict"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, resp, req)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, allocationId+":"+instanceId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UnassociateEipAddress unbinds the EIP from its current association.
func (s *VpcEipService) UnassociateEipAddress(allocationId string) error {
	// TEMP: Use legacy RPC
	client := s.client
	action := "UnassociateEipAddress"
	req := map[string]interface{}{
		"RegionId": client.RegionId,
	}
	query := map[string]interface{}{
		"AllocationId": allocationId,
	}
	var (
		resp map[string]interface{}
		err  error
	)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		resp, err = client.RpcPost("Vpc", "2016-04-28", action, query, req, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"OperationConflict", "LastTokenProcessing", "IncorrectStatus", "SystemBusy", "ServiceUnavailable", "IncorrectEipStatus", "IncorrectInstanceStatus", "TaskConflict"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, resp, req)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, allocationId, action, AlibabaCloudSdkGoERROR)
	}
	return nil
}

// EipStateRefreshFunc returns a StateRefreshFunc for polling EIP status.
// Fail states should reflect provider semantics, e.g. ["Released", "Failed"].
func (s *VpcEipService) EipStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeEipAddress(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		// Best-effort: expect status field as "Status"; callers should pass correct fail states.
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

// WaitForEipCreating waits for EIP to reach a stable state (e.g., Available).
func (s *VpcEipService) WaitForEipCreating(id string, timeout time.Duration) error {
	// Default pending/target states per research/spec; callers may tune if needed.
	stateConf := BuildStateConf(
		[]string{"Allocating", "Associating", "Unassociating"},
		[]string{"Available", "InUse"},
		timeout,
		5*time.Second,
		s.EipStateRefreshFunc(id, []string{"Released", "Failed"}),
	)
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// convertToStringInterfaceMap converts an arbitrary object (struct/map) to map[string]interface{}.
// It uses JSON marshal/unmarshal to preserve json tags from cws-lib-go types.
func convertToStringInterfaceMap(obj interface{}) map[string]interface{} {
	if obj == nil {
		return nil
	}
	if m, ok := obj.(map[string]interface{}); ok {
		return m
	}
	// Try pointer to map[string]interface{}
	if rv := reflect.ValueOf(obj); rv.IsValid() {
		if rv.Kind() == reflect.Ptr {
			if rv.Elem().IsValid() {
				if mm, ok := rv.Elem().Interface().(map[string]interface{}); ok {
					return mm
				}
			}
		}
	}
	// Fallback to JSON round-trip
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil
	}
	return out
}
