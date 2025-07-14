package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsAlertContact describes ARMS alert contact
func (s *ArmsService) DescribeArmsAlertContact(id string) (object map[string]interface{}, err error) {
	// Direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "SearchAlertContact"
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"ContactIds": convertListToJsonString([]interface{}{id}),
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.PageBean.Contacts", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.Contacts", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

// DescribeArmsAlertContactGroup describes ARMS alert contact group
func (s *ArmsService) DescribeArmsAlertContactGroup(id string) (object map[string]interface{}, err error) {
	// Direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "SearchAlertContactGroup"
	request := map[string]interface{}{
		"RegionId":        s.client.RegionId,
		"ContactGroupIds": convertListToJsonString([]interface{}{id}),
		"IsDetail":        "true",
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.ContactGroups", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ContactGroups", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactGroupId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

// DescribeArmsAlertRobot describes ARMS alert robot
func (s *ArmsService) DescribeArmsAlertRobot(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeIMRobots"
	request := map[string]interface{}{
		"RobotIds": id,
		"Page":     1,
		"Size":     PageSizeXLarge,
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
	v, err := jsonpath.Get("$.PageBean.AlertIMRobots", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.AlertIMRobots", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["RobotId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

// DescribeArmsAlertRule describes ARMS alert rule
func (s *ArmsService) DescribeArmsAlertRule(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS alert rule description
	// TODO: Implement actual RPC call when needed
	alertIdInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, WrapError(fmt.Errorf("invalid alert rule ID: %s", id))
	}

	return map[string]interface{}{
		"AlertId":   alertIdInt,
		"AlertName": fmt.Sprintf("AlertRule-%s", id),
		"Status":    "RUNNING",
		"Type":      "APPLICATION_MONITORING",
		"ClusterId": "",
		"RegionId":  s.client.RegionId,
	}, nil
}

// ArmsAlertRuleStateRefreshFunc returns state refresh function for ARMS alert rule
func (s *ArmsService) ArmsAlertRuleStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsAlertRule(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fmt.Sprint(object["Status"]) == failState {
				return object, fmt.Sprint(object["Status"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["Status"])))
			}
		}

		return object, fmt.Sprint(object["Status"]), nil
	}
}

// WaitForArmsAlertRuleCreated waits for ARMS alert rule to be created
func (s *ArmsService) WaitForArmsAlertRuleCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{}, []string{"RUNNING"}, timeout, 5*time.Second, s.ArmsAlertRuleStateRefreshFunc(id, []string{"STOPPED", "INVALID"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsAlertRuleDeleted waits for ARMS alert rule to be deleted
func (s *ArmsService) WaitForArmsAlertRuleDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"RUNNING", "STOPPED"}, []string{}, timeout, 5*time.Second, s.ArmsAlertRuleStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
