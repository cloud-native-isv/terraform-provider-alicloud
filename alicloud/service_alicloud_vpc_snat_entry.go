package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunVpcAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// VpcSnatEntryService provides operations for SNAT Entry via CWS-Lib-Go API layer.
type VpcSnatEntryService struct {
	client *connectivity.AliyunClient
	vpcAPI *aliyunVpcAPI.VpcAPI
}

// NewVpcSnatEntryService creates a new VpcSnatEntryService bound to the given AliyunClient.
func NewVpcSnatEntryService(client *connectivity.AliyunClient) (*VpcSnatEntryService, error) {
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

	return &VpcSnatEntryService{client: client, vpcAPI: vpcAPI}, nil
}

// EncodeSnatId encodes snatTableId and snatEntryId into a single Id string.
func EncodeSnatId(snatTableId, snatEntryId string) string {
	return fmt.Sprintf("%s:%s", EscapeColons(snatTableId), EscapeColons(snatEntryId))
}

// DecodeSnatId decodes a composite SNAT Id.
func DecodeSnatId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", WrapError(Error("invalid SNAT Id format, expected snatTableId:snatEntryId, got %s", id))
	}
	return parts[0], parts[1], nil
}

// CreateSnatEntry creates a SNAT entry and returns its snatEntryId.
func (s *VpcSnatEntryService) CreateSnatEntry(request map[string]interface{}) (string, error) {
	return "", WrapError(Error("CreateSnatEntry requires timeout, use CreateSnatEntryWithTimeout"))
}

// DescribeSnatEntry returns SNAT entry attributes.
func (s *VpcSnatEntryService) DescribeSnatEntry(id string) (map[string]interface{}, error) {
	if id == "" {
		return nil, WrapError(Error("SNAT Id is empty"))
	}

	// Parse the composite ID to get snatTableId and snatEntryId
	snatTableId, snatEntryId, err := DecodeSnatId(id)
	if err != nil {
		return nil, WrapError(err)
	}

	var response *aliyunVpcAPI.SNATEntry
	var apiErr error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	apiErr = resource.Retry(5*time.Minute, func() *resource.RetryError {
		// cws-lib-go does not expose DescribeSnatEntry; list with filters instead
		req := &aliyunVpcAPI.DescribeSnatEntriesRequest{SnatTableId: snatTableId, SnatEntryId: snatEntryId, PageNumber: 1, PageSize: 50}
		entries, _, err := s.vpcAPI.ListSnatEntries(req)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		// pick first matching entry
		if len(entries) > 0 {
			response = entries[0]
		} else {
			response = nil
		}
		return nil
	})

	if apiErr != nil {
		if IsNotFoundError(apiErr) || IsExpectedErrors(apiErr, []string{"InvalidSnatTableId.NotFound", "InvalidSnatEntryId.NotFound"}) {
			return nil, WrapErrorf(NotFoundErr("SnatEntry", id), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(apiErr, DefaultErrorMsg, id, "DescribeSnatEntry", AlibabaCloudSdkGoERROR)
	}

	if response == nil {
		return nil, WrapErrorf(NotFoundErr("SnatEntry", id), NotFoundMsg, ProviderERROR)
	}

	// Convert struct to map for compatibility
	result := make(map[string]interface{})
	result["SnatEntryId"] = response.SnatEntryId
	result["SnatTableId"] = response.SnatTableId
	result["SnatIp"] = response.SnatIp
	result["SourceCIDR"] = response.SourceCIDR
	result["SourceVSwitchId"] = response.SourceVSwitchId
	result["SnatEntryName"] = response.SnatEntryName
	result["Status"] = response.Status
	// EipAffinity is not available in current API strong type

	return result, nil
}

// ModifySnatEntry updates SNAT entry mutable attributes.
func (s *VpcSnatEntryService) ModifySnatEntry(id string, attrs map[string]interface{}) error {
	return WrapError(Error("ModifySnatEntry requires timeout, use ModifySnatEntryWithTimeout"))
}

// DeleteSnatEntry deletes a SNAT entry.
func (s *VpcSnatEntryService) DeleteSnatEntry(id string) error {
	return WrapError(Error("DeleteSnatEntry requires timeout, use DeleteSnatEntryWithTimeout"))
}

// SnatEntryStateRefreshFunc polls SNAT entry status.
func (s *VpcSnatEntryService) SnatEntryStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeSnatEntry(id)
		if err != nil {
			if IsNotFoundError(err) || strings.Contains(err.Error(), "not found") {
				return nil, "", nil
			}
			// 防止WrapError接收到nil错误
			if err == nil {
				return nil, "", WrapError(Error("Unknown error occurred while describing SNAT entry %s", id))
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

// WaitForSnatEntryCreating waits for SNAT entry to reach Available.
func (s *VpcSnatEntryService) WaitForSnatEntryCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Modifying", "Pending"},
		[]string{"Available"},
		timeout,
		5*time.Second,
		s.SnatEntryStateRefreshFunc(id, []string{"Deleting", "Failed"}),
	)
	_, err := stateConf.WaitForState()
	// 防止WrapErrorf接收到nil错误
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// WaitForSnatEntryDeleting waits for SNAT entry deletion.
func (s *VpcSnatEntryService) WaitForSnatEntryDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting"},
		[]string{""},
		timeout,
		5*time.Second,
		s.SnatEntryStateRefreshFunc(id, []string{}),
	)
	_, err := stateConf.WaitForState()
	// 防止WrapErrorf接收到nil错误
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// CreateSnatEntryWithTimeout creates a SNAT entry with retry handling.
func (s *VpcSnatEntryService) CreateSnatEntryWithTimeout(request map[string]interface{}, timeout time.Duration) (string, error) {
	if request == nil {
		request = map[string]interface{}{}
	}

	createRequest := &aliyunVpcAPI.CreateSnatEntryRequest{}

	if v, ok := request["SnatTableId"]; ok {
		createRequest.SnatTableId = v.(string)
	}
	if v, ok := request["SnatIp"]; ok {
		createRequest.SnatIp = v.(string)
	}
	if v, ok := request["SourceVSwitchId"]; ok {
		createRequest.SourceVSwitchId = v.(string)
	}
	if v, ok := request["SnatEntryName"]; ok {
		createRequest.SnatEntryName = v.(string)
	}
	if v, ok := request["SourceCIDR"]; ok {
		createRequest.SourceCIDR = v.(string)
	}
	// EipAffinity and ClientToken are not supported by current API layer

	var snatEntryId string
	var err error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		entryId, err := s.vpcAPI.CreateSnatEntry(createRequest)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"EIP_NOT_IN_GATEWAY", "OperationUnsupported.EipNatBWPCheck", "OperationUnsupported.EipInBinding", "InternalError", "IncorrectStatus.NATGW", "OperationConflict", "OperationUnsupported.EipNatGWCheck"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		snatEntryId = entryId
		return nil
	})

	if err != nil {
		return "", WrapErrorf(err, DefaultErrorMsg, "alicloud_snat_entry", "CreateSnatEntry", AlibabaCloudSdkGoERROR)
	}

	if snatEntryId == "" {
		return "", WrapError(Error("CreateSnatEntry response missing SnatEntryId"))
	}

	return snatEntryId, nil
}

// ModifySnatEntryWithTimeout modifies SNAT entry attributes with retry.
func (s *VpcSnatEntryService) ModifySnatEntryWithTimeout(id string, attrs map[string]interface{}, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("SNAT Id is empty"))
	}
	// The new API layer doesn't support Modify for SNAT entry; surface as unsupported
	return WrapError(Error("ModifySnatEntry is not supported by API layer"))
}

// DeleteSnatEntryWithTimeout deletes a SNAT entry with retry.
func (s *VpcSnatEntryService) DeleteSnatEntryWithTimeout(id string, timeout time.Duration) error {
	if id == "" {
		return WrapError(Error("SNAT Id is empty"))
	}

	snatTableId, snatEntryId, err := DecodeSnatId(id)
	if err != nil {
		return WrapError(err)
	}

	var err2 error
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err2 = resource.Retry(s.client.GetRetryTimeout(timeout), func() *resource.RetryError {
		err := s.vpcAPI.DeleteSnatEntry(snatTableId, snatEntryId)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"IncorretSnatEntryStatus", "IncorrectStatus.NATGW", "OperationConflict"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err2 != nil {
		if IsNotFoundError(err2) || IsExpectedErrors(err2, []string{"InvalidSnatTableId.NotFound", "InvalidSnatEntryId.NotFound"}) {
			return nil
		}
		return WrapErrorf(err2, DefaultErrorMsg, id, "DeleteSnatEntry", AlibabaCloudSdkGoERROR)
	}

	return nil
}
