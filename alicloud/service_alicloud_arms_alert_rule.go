package alicloud

import (
	"fmt"
	"regexp"
	"strconv"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	common "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

// =============================================================================
// AlertRule List and Filter Methods - using strong typing
// =============================================================================

// ListArmsAlertRules lists ARMS alert rules with filtering and pagination
func (s *ArmsService) ListArmsAlertRules(page, size int64, alertName, alertType, clusterId, status string, ids []string, nameRegex *regexp.Regexp) ([]*aliyunArmsAPI.AlertRule, int64, error) {
	if page <= 0 {
		page = common.DefaultStartPage
	}
	if size <= 0 {
		size = common.DefaultPageSize
	}

	// Use cws-lib-go API if available
	if s.armsAPI != nil {
		// Get rules from API (simplified API only accepts page and size)
		allRules, totalCount, err := s.GetAPI().ListAlertRules(page, size)
		if err != nil {
			return nil, 0, WrapError(err)
		}

		// Apply client-side filters since API no longer supports filtering parameters
		var filteredRules []*aliyunArmsAPI.AlertRule
		idsMap := make(map[string]bool)
		if len(ids) > 0 {
			for _, id := range ids {
				idsMap[id] = true
			}
		}

		for _, rule := range allRules {
			// Apply ID filter
			if len(idsMap) > 0 {
				ruleIdStr := strconv.FormatInt(rule.AlertId, 10)
				if !idsMap[ruleIdStr] {
					continue
				}
			}

			// Apply alert name filter
			if alertName != "" && rule.AlertName != alertName {
				continue
			}

			// Apply alert type filter
			if alertType != "" && rule.AlertType != alertType {
				continue
			}

			// Apply cluster ID filter
			if clusterId != "" && rule.ClusterId != clusterId {
				continue
			}

			// Apply status filter (if applicable to the rule structure)
			// Note: Status filtering might need adjustment based on actual AlertRule structure

			// Apply name regex filter
			if nameRegex != nil && !nameRegex.MatchString(rule.AlertName) {
				continue
			}

			filteredRules = append(filteredRules, rule)
		}

		return filteredRules, totalCount, nil
	}

	// Fallback to direct RPC call - this should not be used in practice
	return nil, 0, WrapError(fmt.Errorf("cws-lib-go ARMS API not available"))
}

// =============================================================================
// AlertRule CRUD Methods - using strong typing
// =============================================================================

// DescribeArmsAlertRule describes ARMS alert rule using strong typing
func (s *ArmsService) DescribeArmsAlertRule(id string) (*aliyunArmsAPI.AlertRule, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		alertId, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			alertRule, err := s.GetAPI().GetAlertRule(alertId)
			if err != nil {
				if IsNotFoundError(err) {
					return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundMsg, AlibabaCloudSdkGoERROR)
				}
				return nil, WrapError(err)
			}
			return alertRule, nil
		}
	}

	// Fallback implementation should not be used
	return nil, WrapError(fmt.Errorf("cws-lib-go ARMS API not available for alert rule: %s", id))
}

// CreateArmsAlertRule creates a new ARMS alert rule using strong typing
func (s *ArmsService) CreateArmsAlertRule(rule *aliyunArmsAPI.AlertRule) (*aliyunArmsAPI.AlertRule, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		return s.GetAPI().CreateAlertRule(rule)
	}

	// Fallback implementation should not be used
	return nil, WrapError(fmt.Errorf("cws-lib-go ARMS API not available"))
}

// UpdateArmsAlertRule updates an existing ARMS alert rule using strong typing
func (s *ArmsService) UpdateArmsAlertRule(rule *aliyunArmsAPI.AlertRule) (*aliyunArmsAPI.AlertRule, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		return s.GetAPI().UpdateAlertRule(rule)
	}

	// Fallback implementation should not be used
	return nil, WrapError(fmt.Errorf("cws-lib-go ARMS API not available"))
}

// DeleteArmsAlertRule deletes an ARMS alert rule
func (s *ArmsService) DeleteArmsAlertRule(alertId int64) error {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		return s.GetAPI().DeleteAlertRule(alertId)
	}

	// Fallback implementation should not be used
	return WrapError(fmt.Errorf("cws-lib-go ARMS API not available"))
}
