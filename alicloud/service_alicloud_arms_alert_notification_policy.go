package alicloud

import (
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsAlertNotificationPolicy describes a notification policy by ID
func (s *ArmsService) DescribeArmsAlertNotificationPolicy(id string) (object map[string]interface{}, err error) {
	// Try using ARMS API if available
	if s.armsAPI != nil {
		policy, err := s.armsAPI.GetAlertNotificationPolicy(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			result := map[string]interface{}{
				"Id":                 policy.Id,
				"Name":               policy.Name,
				"SendRecoverMessage": policy.SendRecoverMessage,
				"RepeatInterval":     policy.RepeatInterval,
				"EscalationPolicyId": policy.EscalationPolicyId,
				"State":              policy.State,
				"CreateTime":         policy.CreateTime,
				"UpdateTime":         policy.UpdateTime,
				"GroupRule":          policy.GroupRule,
				"NotifyRule":         policy.NotifyRule,
				"MatchingRules":      policy.MatchingRules,
			}
			return result, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetNotificationPolicy"
	client := s.client

	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"Id":       id,
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

	v, err := jsonpath.Get("$.NotificationPolicy", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.NotificationPolicy", response)
	}

	object = v.(map[string]interface{})
	return object, nil
}

// ListArmsAlertNotificationPolicies lists notification policies
func (s *ArmsService) ListArmsAlertNotificationPolicies() (objects []interface{}, err error) {
	// Try using ARMS API first if available
	if s.armsAPI != nil {
		policies, err := s.armsAPI.ListAlertNotificationPolicies(1, PageSizeXLarge)
		if err == nil {
			// Convert to the format expected by Terraform
			var result []interface{}
			for _, policy := range policies {
				policyMap := map[string]interface{}{
					"Id":                 policy.Id,
					"Name":               policy.Name,
					"SendRecoverMessage": policy.SendRecoverMessage,
					"RepeatInterval":     policy.RepeatInterval,
					"EscalationPolicyId": policy.EscalationPolicyId,
					"State":              policy.State,
					"CreateTime":         policy.CreateTime,
					"UpdateTime":         policy.UpdateTime,
					"GroupRule":          policy.GroupRule,
					"NotifyRule":         policy.NotifyRule,
					"MatchingRules":      policy.MatchingRules,
				}
				result = append(result, policyMap)
			}
			return result, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListNotificationPolicies"
	client := s.client

	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"Page":     1,
		"Size":     PageSizeXLarge,
	}

	for {
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
			return objects, WrapErrorf(err, DefaultErrorMsg, "ListNotificationPolicies", action, AlibabaCloudSdkGoERROR)
		}

		v, err := jsonpath.Get("$.PageInfo.NotificationPolicies", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListNotificationPolicies", "$.PageInfo.NotificationPolicies", response)
		}

		if v != nil {
			objects = append(objects, v.([]interface{})...)
		}

		totalCount, err := jsonpath.Get("$.PageInfo.Total", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListNotificationPolicies", "$.PageInfo.Total", response)
		}

		if len(objects) >= int(totalCount.(float64)) {
			break
		}

		if page, err := jsonpath.Get("$.PageInfo.Page", response); err != nil || page == nil {
			break
		} else {
			request["Page"] = int(page.(float64)) + 1
		}
	}

	return objects, nil
}
