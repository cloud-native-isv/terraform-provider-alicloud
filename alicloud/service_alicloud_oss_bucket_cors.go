package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketCors related functions

func (s *OssService) DescribeOssBucketCors(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	action := fmt.Sprintf("/?cors")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketCors", action), query, nil, nil, hostMap, false)

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
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchCORSConfiguration"}) {
			return object, WrapErrorf(NotFoundErr("BucketCors", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketCors", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.CORSConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.CORSConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketCorsStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCors(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		if strings.HasPrefix(field, "#") {
			v, _ := jsonpath.Get(strings.TrimPrefix(field, "#"), object)
			if v != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}
