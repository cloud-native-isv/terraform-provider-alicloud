package alicloud

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type EipanycastService struct {
	client *connectivity.AliyunClient
}

func (s *EipanycastService) DescribeEipanycastAnycastEipAddress(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeAnycastEipAddress"
	request := map[string]interface{}{
		"RegionId":  s.client.RegionId,
		"AnycastId": id,
	}
	response, err = client.RpcPost("Eipanycast", "2020-03-09", action, nil, request, true)
	if err != nil {
		err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		return
	}
	addDebug(action, response, request)
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *EipanycastService) EipanycastAnycastEipAddressStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeEipanycastAnycastEipAddress(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object["Status"].(string) == failState {
				return object, fmt.Sprint(object["Status"]), WrapError(Error(FailedToReachTargetStatus, object["Status"].(string)))
			}
		}
		return object, fmt.Sprint(object["Status"]), nil
	}
}

func (s *EipanycastService) DescribeEipanycastAnycastEipAddressAttachment(id string) (object map[string]interface{}, err error) {
	parts, err := ParseResourceId(id, 4)
	if err != nil {
		return nil, WrapError(err)
	}
	var response map[string]interface{}
	client := s.client
	action := "DescribeAnycastEipAddress"
	request := map[string]interface{}{
		"RegionId":  s.client.RegionId,
		"AnycastId": parts[0],
	}
	response, err = client.RpcPost("Eipanycast", "2020-03-09", action, nil, request, true)
	if err != nil {
		err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		return
	}
	addDebug(action, response, request)
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	if len(v.(map[string]interface{})["AnycastEipBindInfoList"].([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("EipanycastAnycastEipAddressAttachment", id), NotFoundWithResponse, response)
	} else {
		if vv, ok := v.(map[string]interface{})["AnycastEipBindInfoList"].([]interface{})[0].(map[string]interface{}); ok {
			if parts[1]+parts[2]+parts[3] != fmt.Sprint(vv["BindInstanceId"], vv["BindInstanceRegionId"], vv["BindInstanceType"]) {
				return object, WrapErrorf(NotFoundErr("EipanycastAnycastEipAddressAttachment", id), NotFoundWithResponse, response)
			}
		}
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *EipanycastService) EipanycastAnycastEipAddressAttachmentStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeEipanycastAnycastEipAddressAttachment(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object["Status"].(string) == failState {
				return object, object["Status"].(string), WrapError(Error(FailedToReachTargetStatus, object["Status"].(string)))
			}
		}
		return object, object["Status"].(string), nil
	}
}
