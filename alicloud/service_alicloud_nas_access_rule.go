package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeNasAccessRule gets NAS access rule information
func (s *NasService) DescribeNasAccessRule(id string) (*aliyunNasAPI.AccessRule, error) {
	accessGroupName, accessRuleId, _, err := s.parseResourceId(id)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "ParseResourceId", AlibabaCloudSdkGoERROR)
	}

	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	// Get specific access rule by ID
	accessRule, err := nasAPI.GetAccessRule(accessGroupName, accessRuleId)
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil, WrapErrorf(Error(GetNotFoundMessage("NasAccessRule", id)), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetAccessRule", AlibabaCloudSdkGoERROR)
	}

	return accessRule, nil
}

// CreateNasAccessRule creates a new NAS access rule
func (s *NasService) CreateNasAccessRule(accessGroupName, sourceCidrIp, rwAccessType, userAccessType string, priority int32, fileSystemType, ipv6SourceCidrIp string) (*aliyunNasAPI.AccessRule, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_rule", "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	accessRule, err := nasAPI.CreateAccessRule(accessGroupName, sourceCidrIp, rwAccessType, userAccessType, priority, fileSystemType, ipv6SourceCidrIp)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_rule", "CreateAccessRule", AlibabaCloudSdkGoERROR)
	}

	return accessRule, nil
}

// UpdateNasAccessRule updates an existing NAS access rule
func (s *NasService) UpdateNasAccessRule(accessGroupName, accessRuleId, sourceCidrIp, rwAccessType, userAccessType string, priority int32, fileSystemType, ipv6SourceCidrIp string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessRuleId, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.ModifyAccessRule(accessGroupName, accessRuleId, sourceCidrIp, rwAccessType, userAccessType, priority, fileSystemType, ipv6SourceCidrIp)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessRuleId, "ModifyAccessRule", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteNasAccessRule deletes a NAS access rule
func (s *NasService) DeleteNasAccessRule(accessGroupName, accessRuleId string) error {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, accessRuleId, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	err = nasAPI.DeleteAccessRule(accessGroupName, accessRuleId)
	if err != nil {
		if aliyunNasAPI.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, accessRuleId, "DeleteAccessRule", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListNasAccessRules lists all access rules for an access group
func (s *NasService) ListNasAccessRules(accessGroupName string) ([]aliyunNasAPI.AccessRule, error) {
	nasAPI, err := s.getNasAPI()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, accessGroupName, "getNasAPI", AlibabaCloudSdkGoERROR)
	}

	accessRules, err := nasAPI.ListAccessRules(accessGroupName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, accessGroupName, "ListAccessRules", AlibabaCloudSdkGoERROR)
	}

	return accessRules, nil
}

// NasAccessRuleStateRefreshFunc returns a StateRefreshFunc for NAS access rule status
func (s *NasService) NasAccessRuleStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		accessRule, err := s.DescribeNasAccessRule(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For access rules, they are usually active once created
		status := "Available"
		if accessRule.AccessRuleId != "" {
			status = "Active"
		}

		for _, failState := range failStates {
			if status == failState {
				return accessRule, status, WrapError(Error(FailedToReachTargetStatus, status))
			}
		}

		return accessRule, status, nil
	}
}

// WaitForNasAccessRule waits for NAS access rule to reach target status
func (s *NasService) WaitForNasAccessRule(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		accessRule, err := s.DescribeNasAccessRule(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if accessRule != nil && accessRule.AccessRuleId != "" && status != Deleted {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, "", status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper function to convert Terraform schema data to API create request
func (s *NasService) buildCreateAccessRuleRequest(d *schema.ResourceData) *aliyunNasAPI.CreateAccessRuleRequest {
	request := &aliyunNasAPI.CreateAccessRuleRequest{
		AccessGroupName: d.Get("access_group_name").(string),
	}

	if v, ok := d.GetOk("file_system_type"); ok {
		request.FileSystemType = v.(string)
	} else {
		request.FileSystemType = "standard"
	}

	if v, ok := d.GetOk("source_cidr_ip"); ok {
		request.SourceCidrIp = v.(string)
	}

	if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok {
		request.Ipv6SourceCidrIp = v.(string)
	}

	if v, ok := d.GetOk("rw_access_type"); ok {
		request.RWAccessType = v.(string)
	}

	if v, ok := d.GetOk("user_access_type"); ok {
		request.UserAccessType = v.(string)
	}

	if v, ok := d.GetOk("priority"); ok {
		request.Priority = int32(v.(int))
	}

	return request
}

// Helper function to convert Terraform schema data to API modify request
func (s *NasService) buildModifyAccessRuleRequest(d *schema.ResourceData) *aliyunNasAPI.CreateAccessRuleRequest {
	request := &aliyunNasAPI.CreateAccessRuleRequest{}

	if v, ok := d.GetOk("source_cidr_ip"); ok {
		request.SourceCidrIp = v.(string)
	}

	if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok {
		request.Ipv6SourceCidrIp = v.(string)
	}

	if v, ok := d.GetOk("rw_access_type"); ok {
		request.RWAccessType = v.(string)
	}

	if v, ok := d.GetOk("user_access_type"); ok {
		request.UserAccessType = v.(string)
	}

	if v, ok := d.GetOk("priority"); ok {
		request.Priority = int32(v.(int))
	}

	if v, ok := d.GetOk("file_system_type"); ok {
		request.FileSystemType = v.(string)
	}

	return request
}

// Helper function to parse resource ID
func (s *NasService) parseResourceId(id string) (accessGroupName, accessRuleId, fileSystemType string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid resource ID format: %s, expected format: access_group_name:access_rule_id[:file_system_type]", id)
	}

	accessGroupName = parts[0]
	accessRuleId = parts[1]
	if len(parts) >= 3 {
		fileSystemType = parts[2]
	} else {
		fileSystemType = "standard"
	}

	return accessGroupName, accessRuleId, fileSystemType, nil
}

// Helper function to build resource ID
func (s *NasService) buildResourceId(accessGroupName, accessRuleId, fileSystemType string) string {
	if fileSystemType == "" || fileSystemType == "standard" {
		return fmt.Sprintf("%s:%s", accessGroupName, accessRuleId)
	}
	return fmt.Sprintf("%s:%s:%s", accessGroupName, accessRuleId, fileSystemType)
}
