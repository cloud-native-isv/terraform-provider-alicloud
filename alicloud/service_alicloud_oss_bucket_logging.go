package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketLogging related functions

func (s *OssService) DescribeOssBucketLogging(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?logging")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketLogging", action), query, nil, nil, hostMap, true)
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
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.BucketLoggingStatus.LoggingEnabled", response)
	if err != nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	currentStatus := v.(map[string]interface{})["TargetBucket"]
	if currentStatus == nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketLoggingStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketLogging(id)
		if err != nil {
			if IsNotFoundError(err) {
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

func (s *OssService) PutBucketLogging(bucketName, targetBucket, targetPrefix string) error {
	request := &ossv2.PutBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
		BucketLoggingStatus: &ossv2.BucketLoggingStatus{
			LoggingEnabled: &ossv2.LoggingEnabled{
				TargetBucket: StringPointer(targetBucket),
				TargetPrefix: StringPointer(targetPrefix),
			},
		},
	}

	_, err := s.v2Client.PutBucketLogging(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *OssService) DeleteBucketLogging(bucketName string) error {
	request := &ossv2.DeleteBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
	}

	_, err := s.v2Client.DeleteBucketLogging(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *OssService) GetBucketLoggingV2(bucketName string) (*ossv2.GetBucketLoggingResult, error) {
	request := &ossv2.GetBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
	}

	result, err := s.v2Client.GetBucketLogging(s.ctx, request)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}
