package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsDispatchRule describes ARMS dispatch rule
func (s *ArmsService) DescribeArmsDispatchRule(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetDispatchRule"
	request := map[string]interface{}{
		"Id": id,
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
	v, err := jsonpath.Get("$.DispatchRule", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.DispatchRule", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

// ArmsDispatchRuleStateRefreshFunc returns state refresh function for ARMS dispatch rule
func (s *ArmsService) ArmsDispatchRuleStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsDispatchRule(id)
		if err != nil {
			if NotFoundError(err) {
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

// WaitForArmsDispatchRuleCreated waits for ARMS dispatch rule to be created
func (s *ArmsService) WaitForArmsDispatchRuleCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{}, []string{"true"}, timeout, 5*time.Second, s.ArmsDispatchRuleStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsDispatchRuleDeleted waits for ARMS dispatch rule to be deleted
func (s *ArmsService) WaitForArmsDispatchRuleDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"true", "false"}, []string{}, timeout, 5*time.Second, s.ArmsDispatchRuleStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
