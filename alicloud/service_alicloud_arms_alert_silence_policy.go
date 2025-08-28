package alicloud

import (
	"fmt"
	"strconv"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateArmsAlertSilencePolicy creates an ARMS alert silence policy
func (s *ArmsService) CreateArmsAlertSilencePolicy(policy *aliyunArmsAPI.AlertSilencePolicy) (*aliyunArmsAPI.AlertSilencePolicy, error) {
	if s.armsAPI != nil {
		// Call the API to create silence policy
		result, err := s.armsAPI.CreateOrUpdateAlertSilencePolicy(
			policy.SilenceName,
			policy.EffectiveTimeType,
			policy.TimePeriod,
			policy.TimeSlots,
			policy.State,
			policy.MatchingRules,
			nil, // silenceId is nil for create
		)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, "ArmsAlertSilencePolicy", "CreateOrUpdateAlertSilencePolicy", AlibabaCloudSdkGoERROR)
		}
		return result, nil
	}

	// Fallback to placeholder
	return nil, WrapError(Error("ARMS API not initialized"))
}

// UpdateArmsAlertSilencePolicy updates an ARMS alert silence policy
func (s *ArmsService) UpdateArmsAlertSilencePolicy(silenceId int64, policy *aliyunArmsAPI.AlertSilencePolicy) (*aliyunArmsAPI.AlertSilencePolicy, error) {
	if s.armsAPI != nil {
		// Call the API to update silence policy
		result, err := s.armsAPI.UpdateAlertSilencePolicy(
			silenceId,
			policy.SilenceName,
			policy.EffectiveTimeType,
			policy.TimePeriod,
			policy.TimeSlots,
			policy.State,
			policy.MatchingRules,
		)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%d", silenceId), "UpdateAlertSilencePolicy", AlibabaCloudSdkGoERROR)
		}
		return result, nil
	}

	// Fallback to placeholder
	return nil, WrapError(Error("ARMS API not initialized"))
}

// DeleteArmsAlertSilencePolicy deletes an ARMS alert silence policy
func (s *ArmsService) DeleteArmsAlertSilencePolicy(silenceId int64) error {
	if s.armsAPI != nil {
		// Call the API to delete silence policy
		err := s.armsAPI.DeleteAlertSilencePolicy(silenceId)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%d", silenceId), "DeleteAlertSilencePolicy", AlibabaCloudSdkGoERROR)
		}
		return nil
	}

	// Fallback to placeholder
	return WrapError(Error("ARMS API not initialized"))
}

// DescribeArmsAlertSilencePolicy describes an ARMS alert silence policy
func (s *ArmsService) DescribeArmsAlertSilencePolicy(id string) (object map[string]interface{}, err error) {
	if s.armsAPI != nil {
		// Convert string ID to int64
		silenceId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, id, "ParseInt", AlibabaCloudSdkGoERROR)
		}

		// Call the API to get silence policy
		policy, err := s.armsAPI.GetAlertSilencePolicy(silenceId)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
			}
			return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetAlertSilencePolicy", AlibabaCloudSdkGoERROR)
		}

		// Convert to map[string]interface{} format expected by Terraform
		result := map[string]interface{}{
			"SilenceId":         policy.SilenceId,
			"SilenceName":       policy.SilenceName,
			"State":             policy.State,
			"EffectiveTimeType": policy.EffectiveTimeType,
			"TimePeriod":        policy.TimePeriod,
			"TimeSlots":         policy.TimeSlots,
			"StartTime":         policy.StartTime,
			"EndTime":           policy.EndTime,
			"Comment":           policy.Comment,
			"CreatedBy":         policy.CreatedBy,
		}

		// Convert matching rules to string format for Terraform compatibility
		if len(policy.MatchingRules) > 0 {
			// For now, convert to JSON string to maintain compatibility
			// TODO: Consider proper nested structure handling
			result["MatchingRules"] = fmt.Sprintf("%+v", policy.MatchingRules)
		}

		// Set time fields if available
		if policy.CreateTime != nil {
			result["CreateTime"] = policy.CreateTime.Unix()
		}
		if policy.UpdateTime != nil {
			result["UpdateTime"] = policy.UpdateTime.Unix()
		}

		return result, nil
	}

	// Fallback placeholder implementation
	return nil, WrapErrorf(Error(GetNotFoundMessage("ArmsAlertSilencePolicy", id)), NotFoundWithResponse, "DescribeArmsAlertSilencePolicy")
}

// ArmsAlertSilencePolicyStateRefreshFunc returns state refresh function for ARMS alert silence policy
func (s *ArmsService) ArmsAlertSilencePolicyStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsAlertSilencePolicy(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fmt.Sprint(object["State"]) == failState {
				return object, fmt.Sprint(object["State"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["State"])))
			}
		}

		return object, fmt.Sprint(object["State"]), nil
	}
}
