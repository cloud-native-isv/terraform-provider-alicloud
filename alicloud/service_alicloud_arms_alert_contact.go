package alicloud

// DescribeArmsAlertContact describes ARMS alert contact
func (s *ArmsService) DescribeArmsAlertContact(id string) (object map[string]interface{}, err error) {
	// // Direct RPC call
	// var response map[string]interface{}
	// client := s.client
	// action := "SearchAlertContact"
	// request := map[string]interface{}{
	// 	"RegionId":   s.client.RegionId,
	// 	"ContactIds": convertListToJsonString([]interface{}{id}),
	// }
	// wait := incrementalWait(3*time.Second, 3*time.Second)
	// err = resource.Retry(5*time.Minute, func() *resource.RetryError {
	// 	var retryErr error
	// 	response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
	// 	if retryErr != nil {
	// 		if NeedRetry(retryErr) {
	// 			wait()
	// 			return resource.RetryableError(retryErr)
	// 		}
	// 		return resource.NonRetryableError(retryErr)
	// 	}
	// 	return nil
	// })
	// addDebug(action, response, request)
	// if err != nil {
	// 	return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	// }
	// v, err := jsonpath.Get("$.PageBean.Contacts", response)
	// if err != nil {
	// 	return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.Contacts", response)
	// }
	// if len(v.([]interface{})) < 1 {
	// 	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	// } else {
	// 	if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactId"]) != id {
	// 		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	// 	}
	// }
	// object = v.([]interface{})[0].(map[string]interface{})
	return nil, nil
}

// DescribeArmsAlertContactGroup describes ARMS alert contact group
func (s *ArmsService) DescribeArmsAlertContactGroup(id string) (object map[string]interface{}, err error) {
	// Direct RPC call
	// var response map[string]interface{}
	// client := s.client
	// action := "SearchAlertContactGroup"
	// request := map[string]interface{}{
	// 	"RegionId":        s.client.RegionId,
	// 	"ContactGroupIds": convertListToJsonString([]interface{}{id}),
	// 	"IsDetail":        "true",
	// }
	// wait := incrementalWait(3*time.Second, 3*time.Second)
	// err = resource.Retry(5*time.Minute, func() *resource.RetryError {
	// 	var retryErr error
	// 	response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
	// 	if retryErr != nil {
	// 		if NeedRetry(retryErr) {
	// 			wait()
	// 			return resource.RetryableError(retryErr)
	// 		}
	// 		return resource.NonRetryableError(retryErr)
	// 	}
	// 	return nil
	// })
	// addDebug(action, response, request)
	// if err != nil {
	// 	return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	// }
	// v, err := jsonpath.Get("$.ContactGroups", response)
	// if err != nil {
	// 	return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ContactGroups", response)
	// }
	// if len(v.([]interface{})) < 1 {
	// 	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	// } else {
	// 	if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactGroupId"]) != id {
	// 		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	// 	}
	// }
	// object = v.([]interface{})[0].(map[string]interface{})
	return nil, nil
}
