package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *ArmsService) DescribeArmsIntegration(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert string ID to int64
		integrationIdInt, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			integration, err := s.GetAPI().GetIntegrationById(integrationIdInt, true)
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				result := map[string]interface{}{
					"IntegrationId":          integration.IntegrationId,
					"IntegrationName":        integration.IntegrationName,
					"IntegrationProductType": integration.IntegrationProductType,
					"ApiEndpoint":            integration.ApiEndpoint,
					"ShortToken":             integration.ShortToken,
					"State":                  integration.State,
					"Liveness":               integration.Liveness,
					"CreateTime":             integration.CreateTime,
				}

				if integration.IntegrationDetail != nil {
					result["IntegrationDetail"] = integration.IntegrationDetail
				}

				return result, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetIntegration"
	client := s.client

	request := map[string]interface{}{
		"RegionId":      s.client.RegionId,
		"IntegrationId": id,
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

	v, err := jsonpath.Get("$.Integration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Integration", response)
	}

	object = v.(map[string]interface{})
	return object, nil
}

func (s *ArmsService) ListArmsIntegrations() (objects []interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		integrations, err := s.GetAPI().ListAllIntegrations(1, PageSizeXLarge)
		if err == nil {
			// Convert to the format expected by Terraform
			var result []interface{}
			for _, integration := range integrations {
				integrationMap := map[string]interface{}{
					"IntegrationId":          integration.IntegrationId,
					"IntegrationName":        integration.IntegrationName,
					"IntegrationProductType": integration.IntegrationProductType,
					"ApiEndpoint":            integration.ApiEndpoint,
					"ShortToken":             integration.ShortToken,
					"State":                  integration.State,
					"Liveness":               integration.Liveness,
					"CreateTime":             integration.CreateTime,
				}

				if integration.IntegrationDetail != nil {
					integrationMap["IntegrationDetail"] = integration.IntegrationDetail
				}

				result = append(result, integrationMap)
			}
			return result, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListIntegration"
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
			return objects, WrapErrorf(err, DefaultErrorMsg, "ListIntegration", action, AlibabaCloudSdkGoERROR)
		}

		v, err := jsonpath.Get("$.PageInfo.Integrations", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListIntegration", "$.PageInfo.Integrations", response)
		}

		if v != nil {
			objects = append(objects, v.([]interface{})...)
		}

		totalCount, err := jsonpath.Get("$.PageInfo.Total", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListIntegration", "$.PageInfo.Total", response)
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

// DescribeArmsIntegrationExporter describes ARMS integration exporter
func (s *ArmsService) DescribeArmsIntegrationExporter(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetPrometheusIntegration"
	request := map[string]interface{}{
		"InstanceId":      id,
		"IntegrationType": "kafka",
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
	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

// ArmsIntegrationStateRefreshFunc returns state refresh function for ARMS integration
func (s *ArmsService) ArmsIntegrationStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsIntegration(id)
		if err != nil {
			if IsNotFoundError(err) {
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

// WaitForArmsIntegrationCreated waits for ARMS integration to be created
func (s *ArmsService) WaitForArmsIntegrationCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{}, []string{"Success"}, timeout, 5*time.Second, s.ArmsIntegrationStateRefreshFunc(id, []string{"Failed"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsIntegrationDeleted waits for ARMS integration to be deleted
func (s *ArmsService) WaitForArmsIntegrationDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"Success", "Failed"}, []string{}, timeout, 5*time.Second, s.ArmsIntegrationStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
