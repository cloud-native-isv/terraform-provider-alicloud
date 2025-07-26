package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Security IP List Management Operations

// DescribeSelectDBSecurityIPList retrieves security IP list for a SelectDB instance
func (s *SelectDBService) DescribeSelectDBSecurityIPList(query *selectdb.SecurityIPListQuery) (*selectdb.SecurityIPListResult, error) {
	if query == nil {
		return nil, WrapError(fmt.Errorf("security IP list query cannot be nil"))
	}

	result, err := s.selectdbAPI.GetSecurityIPList(query)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return result, nil
}

// ModifySelectDBSecurityIPList modifies security IP list for a SelectDB instance
func (s *SelectDBService) ModifySelectDBSecurityIPList(modification *selectdb.SecurityIPListModification) (*selectdb.SecurityIPListModificationResult, error) {
	if modification == nil {
		return nil, WrapError(fmt.Errorf("security IP list modification cannot be nil"))
	}

	result, err := s.selectdbAPI.ModifySecurityIPList(modification)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBSecurityIPGroup retrieves a specific security IP group
func (s *SelectDBService) DescribeSelectDBSecurityIPGroup(instanceId, groupName string) (*selectdb.SecurityIPGroup, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}
	if groupName == "" {
		return nil, WrapError(fmt.Errorf("group name cannot be empty"))
	}

	query := &selectdb.SecurityIPListQuery{
		DBInstanceId: instanceId,
	}

	result, err := s.selectdbAPI.GetSecurityIPList(query)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Find the specific group by name
	for _, group := range result.GroupItems {
		if group.GroupName == groupName {
			return &group, nil
		}
	}

	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Security IP Group", groupName)), NotFoundMsg, ProviderERROR)
}

// State Management and Refresh Functions

// SelectDBSecurityIPListStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB security IP list
func (s *SelectDBService) SelectDBSecurityIPListStateRefreshFunc(instanceId, groupName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		group, err := s.DescribeSelectDBSecurityIPGroup(instanceId, groupName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", WrapErrorf(Error(GetNotFoundMessage("SelectDB Security IP Group", groupName)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		// Security IP groups don't have explicit status, so we use a simple Available/NotAvailable
		status := "Available"
		if len(group.SecurityIPList) == 0 {
			status = "NotAvailable"
		}

		for _, failState := range failStates {
			if status == failState {
				return group, status, WrapError(Error(FailedToReachTargetStatus, status))
			}
		}

		return group, status, nil
	}
}

// WaitForSelectDBSecurityIPList waits for SelectDB security IP list to reach expected status
func (s *SelectDBService) WaitForSelectDBSecurityIPList(instanceId, groupName string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		group, err := s.DescribeSelectDBSecurityIPGroup(instanceId, groupName)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		currentStatus := "Available"
		if group != nil && len(group.SecurityIPList) == 0 {
			currentStatus = "NotAvailable"
		}

		if currentStatus == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, groupName, GetFunc(1), timeout, currentStatus, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper functions for converting between Terraform schema and API types

// ConvertToSecurityIPListQuery converts schema data to API security IP list query
func ConvertToSecurityIPListQuery(d *schema.ResourceData) *selectdb.SecurityIPListQuery {
	query := &selectdb.SecurityIPListQuery{}

	if v, ok := d.GetOk("db_instance_id"); ok {
		query.DBInstanceId = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		query.RegionId = v.(string)
	}

	return query
}

// ConvertToSecurityIPListModification converts schema data to API security IP list modification
func ConvertToSecurityIPListModification(d *schema.ResourceData) *selectdb.SecurityIPListModification {
	modification := &selectdb.SecurityIPListModification{}

	if v, ok := d.GetOk("db_instance_id"); ok {
		modification.DBInstanceId = v.(string)
	}
	if v, ok := d.GetOk("group_name"); ok {
		modification.GroupName = v.(string)
	}
	if v, ok := d.GetOk("security_ip_list"); ok {
		// Convert slice to comma-separated string if needed
		switch ips := v.(type) {
		case []interface{}:
			var ipStrings []string
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					ipStrings = append(ipStrings, ipStr)
				}
			}
			modification.SecurityIPList = strings.Join(ipStrings, ",")
		case string:
			modification.SecurityIPList = ips
		}
	}
	if v, ok := d.GetOk("modify_mode"); ok {
		modification.ModifyMode = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		modification.RegionId = v.(string)
	}

	return modification
}

// ConvertSecurityIPListToMap converts API security IP list to Terraform map
func ConvertSecurityIPListToMap(result *selectdb.SecurityIPListResult) map[string]interface{} {
	if result == nil {
		return nil
	}

	resultMap := map[string]interface{}{
		"db_instance_name": result.DBInstanceName,
	}

	// Convert group items
	if len(result.GroupItems) > 0 {
		groups := make([]map[string]interface{}, 0, len(result.GroupItems))
		for _, group := range result.GroupItems {
			g := map[string]interface{}{
				"group_name":         group.GroupName,
				"group_tag":          group.GroupTag,
				"security_ip_list":   group.SecurityIPList,
				"whitelist_net_type": group.WhitelistNetType,
			}
			groups = append(groups, g)
		}
		resultMap["group_items"] = groups
	}

	return resultMap
}

// ConvertSecurityIPGroupToMap converts API security IP group to Terraform map
func ConvertSecurityIPGroupToMap(group *selectdb.SecurityIPGroup) map[string]interface{} {
	if group == nil {
		return nil
	}

	return map[string]interface{}{
		"group_name":         group.GroupName,
		"group_tag":          group.GroupTag,
		"security_ip_list":   group.SecurityIPList,
		"whitelist_net_type": group.WhitelistNetType,
	}
}

// ConvertSecurityIPListModificationResultToMap converts API modification result to Terraform map
func ConvertSecurityIPListModificationResultToMap(result *selectdb.SecurityIPListModificationResult) map[string]interface{} {
	if result == nil {
		return nil
	}

	return map[string]interface{}{
		"db_instance_name":   result.DBInstanceName,
		"group_name":         result.GroupName,
		"group_tag":          result.GroupTag,
		"security_ip_list":   result.SecurityIPList,
		"security_ip_type":   result.SecurityIPType,
		"whitelist_net_type": result.WhitelistNetType,
		"task_id":            result.TaskId,
	}
}

// Utility functions for security IP list management

// ValidateSecurityIPList validates security IP list format
func ValidateSecurityIPList(ipList []string) error {
	if len(ipList) == 0 {
		return fmt.Errorf("security IP list cannot be empty")
	}

	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// Basic validation for IP address or CIDR format
		// This is a simple check - more comprehensive validation could be added
		if !strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
			return fmt.Errorf("invalid IP address format: %s", ip)
		}
	}

	return nil
}

// NormalizeSecurityIPList normalizes security IP list by removing duplicates and empty entries
func NormalizeSecurityIPList(ipList []string) []string {
	seen := make(map[string]bool)
	var normalized []string

	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if ip != "" && !seen[ip] {
			seen[ip] = true
			normalized = append(normalized, ip)
		}
	}

	return normalized
}
