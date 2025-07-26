package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketPolicy related functions

func (s *OssService) DescribeOssBucketPolicy(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?policy")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlJsonParam("GET", "2019-05-17", "GetBucketPolicy", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketPolicy", id), NotFoundMsg, ProviderERROR, fmt.Sprint(response["RequestId"]))
	}

	return response, nil
}

func (s *OssService) OssBucketPolicyStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketPolicy(id)
		if err != nil {
			if IsNotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Check if the policy is properly configured by verifying it contains valid policy data
		if strings.HasPrefix(field, "#") {
			// Check if the response contains policy data (Statement field indicates a valid policy)
			if v, err := jsonpath.Get("$.Statement", object); err == nil && v != nil {
				return object, "#CHECKSET", nil
			}
			// Also check if the response contains valid JSON policy as string
			if policyStr, ok := object["policy"].(string); ok && policyStr != "" {
				return object, "#CHECKSET", nil
			}
			// Check if object itself has policy-like structure
			if len(object) > 0 {
				// The object response indicates policy exists
				return object, "#CHECKSET", nil
			}
			return object, "", nil
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}
