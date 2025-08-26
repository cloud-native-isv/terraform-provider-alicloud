package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketPolicy related functions

func (s *OssService) DescribeOssBucketPolicy(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	policy, err := ossAPI.GetBucketPolicy(id)
	if err != nil {
		if IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("BucketPolicy", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketPolicy", AlibabaCloudSdkGoERROR)
	}

	if policy == nil || policy.Policy == "" {
		return object, WrapErrorf(NotFoundErr("BucketPolicy", id), NotFoundMsg, "policy is empty")
	}

	result := make(map[string]interface{})
	result["Policy"] = policy.Policy
	return result, nil
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
