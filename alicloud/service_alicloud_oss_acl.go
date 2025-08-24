package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// GetBucketAcl gets bucket ACL using cws-lib-go API
func (s *OssService) GetBucketAcl(bucketName string) (string, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return "", WrapError(err)
	}

	acl, err := ossAPI.GetBucketAcl(bucketName)
	if err != nil {
		if ossNotFoundError(err) {
			return "", WrapErrorf(err, NotFoundMsg, "OSS API")
		}
		return "", WrapErrorf(err, DefaultErrorMsg, bucketName, "GetBucketAcl", "OSS API")
	}

	return acl, nil
}

// SetBucketAcl sets bucket ACL using cws-lib-go API
func (s *OssService) SetBucketAcl(bucketName, acl string) error {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return WrapError(err)
	}

	err = ossAPI.SetBucketAcl(bucketName, acl)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, bucketName, "SetBucketAcl", "OSS API")
	}

	return nil
}

func (s *OssService) DescribeOssBucketAcl(id string) (object map[string]interface{}, err error) {
	// Try to use new cws-lib-go API first
	ossAPI, apiErr := s.GetOssAPI()
	if apiErr == nil {
		acl, err := ossAPI.GetBucketAcl(id)
		if err == nil {
			// Convert ACL string to expected format
			object = map[string]interface{}{
				"Grant": acl,
			}
			return object, nil
		}
		// Fall back to old implementation on API error
	}

	// Fallback to original implementation
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?acl")

	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketAcl", action), query, nil, nil, hostMap, true)
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
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketAcl", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketAcl", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.AccessControlPolicy.AccessControlList", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AccessControlPolicy.AccessControlList", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketAclStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketAcl(id)
		if err != nil {
			if IsNotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
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
