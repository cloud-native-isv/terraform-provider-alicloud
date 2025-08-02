package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsSyntheticTask describes ARMS synthetic task
func (s *ArmsService) DescribeArmsSyntheticTask(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetSyntheticTask"
	request := map[string]interface{}{
		"TaskId": id,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.TaskDetail", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.TaskDetail", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

// ArmsSyntheticTaskStateRefreshFunc returns state refresh function for ARMS synthetic task
func (s *ArmsService) ArmsSyntheticTaskStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsSyntheticTask(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fmt.Sprint(object["TaskStatus"]) == failState {
				return object, fmt.Sprint(object["TaskStatus"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["TaskStatus"])))
			}
		}

		return object, fmt.Sprint(object["TaskStatus"]), nil
	}
}

// WaitForArmsSyntheticTaskCreated waits for ARMS synthetic task to be created
func (s *ArmsService) WaitForArmsSyntheticTaskCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, timeout, 5*time.Second, s.ArmsSyntheticTaskStateRefreshFunc(id, []string{"CREATE_FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsSyntheticTaskDeleted waits for ARMS synthetic task to be deleted
func (s *ArmsService) WaitForArmsSyntheticTaskDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"RUNNING", "STOPPING"}, []string{}, timeout, 5*time.Second, s.ArmsSyntheticTaskStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// DescribeArmsAlertSilencePolicy describes ARMS alert silence policy
func (s *ArmsService) DescribeArmsAlertSilencePolicy(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert string ID to int64
		silenceIdInt, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			silencePolicy, err := s.GetAPI().GetAlertSilencePolicy(silenceIdInt)
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"SilenceId":         silencePolicy.SilenceId,
					"SilenceName":       silencePolicy.SilenceName,
					"State":             silencePolicy.State,
					"EffectiveTimeType": silencePolicy.EffectiveTimeType,
					"TimePeriod":        silencePolicy.TimePeriod,
					"TimeSlots":         silencePolicy.TimeSlots,
					"StartTime":         silencePolicy.StartTime,
					"EndTime":           silencePolicy.EndTime,
					"Comment":           silencePolicy.Comment,
					"MatchingRules":     silencePolicy.MatchingRules,
					"CreatedBy":         silencePolicy.CreatedBy,
					"CreateTime":        silencePolicy.CreateTime,
					"UpdateTime":        silencePolicy.UpdateTime,
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListSilencePolicies"
	client := s.client

	request := map[string]interface{}{
		"Page":     1,
		"Size":     PageSizeXLarge,
		"RegionId": s.client.RegionId,
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.PageBean.SilencePolicies", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.SilencePolicies", response)
	}

	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}

	// Find the silence policy with matching ID
	for _, policy := range v.([]interface{}) {
		policyMap := policy.(map[string]interface{})
		if fmt.Sprint(policyMap["SilenceId"]) == id {
			return policyMap, nil
		}
	}

	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}
