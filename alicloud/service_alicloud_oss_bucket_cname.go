package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)


func (s *OssService) DescribeOssBucketCname(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?cname")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "ListCname", action), query, nil, nil, hostMap, true)
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
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.ListCnameResult", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ListCnameResult", response)
	}

	item := v.(map[string]interface{})
	domains, err := jsonpath.Get("$.Cname[*]", item)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Cname[*]", item)
	}
	for _, vv := range domains.([]interface{}) {
		if vv.(map[string]interface{})["Domain"] == parts[1] {
			return vv.(map[string]interface{}), nil
		}
	}

	return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
}

func (s *OssService) OssBucketCnameStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCname(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		if err != nil {
			return object, "", WrapErrorf(err, FailedGetAttributeMsg, id, field, object)
		}
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

