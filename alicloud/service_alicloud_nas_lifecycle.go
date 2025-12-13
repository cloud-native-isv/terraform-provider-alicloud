package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeNasLifecyclePolicy gets NAS lifecycle policy information
func (s *NasService) DescribeNasLifecyclePolicy(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeLifecyclePolicies"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"FileSystemId": parts[0],
		"PageNumber":   1,
		"PageSize":     PageSizeMedium,
	}
	idExist := false
	for {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			response, err = client.RpcGet("NAS", "2017-06-26", action, request, nil)
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
			if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound"}) {
				return object, WrapErrorf(NotFoundErr("NAS:LifecyclePolicy", id), NotFoundMsg, ProviderERROR, fmt.Sprint(response["RequestId"]))
			}
			return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		}
		v, err := jsonpath.Get("$.LifecyclePolicies", response)
		if err != nil {
			return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.LifecyclePolicies", response)
		}
		if len(v.([]interface{})) < 1 {
			return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
		}
		for _, v := range v.([]interface{}) {
			if fmt.Sprint(v.(map[string]interface{})["LifecyclePolicyName"]) == parts[1] {
				idExist = true
				return v.(map[string]interface{}), nil
			}
		}
		if len(v.([]interface{})) < request["PageSize"].(int) {
			break
		}
		request["PageNumber"] = request["PageNumber"].(int) + 1
	}
	if !idExist {
		return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
	}
	return
}

// DescribeNasDataFlow gets NAS data flow information
func (s *NasService) DescribeNasDataFlow(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeDataFlows"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"FileSystemId": parts[0],
		"MaxResults":   20,
	}
	idExist := false
	for {
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			response, err = client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
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
			if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound"}) {
				return object, WrapErrorf(NotFoundErr("NAS:DataFlow", id), NotFoundMsg, ProviderERROR, fmt.Sprint(response["RequestId"]))
			}
			return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		}
		v, err := jsonpath.Get("$.DataFlowInfo.DataFlow", response)
		if err != nil {
			return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.DataFlowInfo.DataFlow", response)
		}
		if len(v.([]interface{})) < 1 {
			return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
		}
		for _, v := range v.([]interface{}) {
			if fmt.Sprint(v.(map[string]interface{})["DataFlowId"]) == parts[1] {
				idExist = true
				return v.(map[string]interface{}), nil
			}
		}

		if nextToken, ok := response["NextToken"].(string); ok && nextToken != "" {
			request["NextToken"] = nextToken
		} else {
			break
		}
	}
	if !idExist {
		return object, WrapErrorf(NotFoundErr("NAS", id), NotFoundWithResponse, response)
	}
	return
}

// NasDataFlowStateRefreshFunc returns a StateRefreshFunc for NAS data flow status
func (s *NasService) NasDataFlowStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeNasDataFlow(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
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
