package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketUserDefinedLogFields related functions

func (s *OssService) DescribeOssBucketUserDefinedLogFields(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?userDefinedLogFieldsConfig")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketUserDefinedLogFieldsConfig", action), query, nil, nil, hostMap, true)
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
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchUserDefinedLogFieldsConfig"}) {
			return object, WrapErrorf(NotFoundErr("BucketUserDefinedLogFields", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketUserDefinedLogFields", id), NotFoundMsg, response)
	}

	return response, nil
}

func (s *OssService) OssBucketUserDefinedLogFieldsStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketUserDefinedLogFields(id)
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

func (s *OssService) PutUserDefinedLogFields(bucketName string, headerSet []string, paramSet []string) error {
	request := &ossv2.PutUserDefinedLogFieldsConfigRequest{
		Bucket:                            StringPointer(bucketName),
		UserDefinedLogFieldsConfiguration: &ossv2.UserDefinedLogFieldsConfiguration{},
	}

	if len(headerSet) > 0 {
		request.UserDefinedLogFieldsConfiguration.HeaderSet = &ossv2.LoggingHeaderSet{
			Headers: headerSet,
		}
	}

	if len(paramSet) > 0 {
		request.UserDefinedLogFieldsConfiguration.ParamSet = &ossv2.LoggingParamSet{
			Parameters: paramSet,
		}
	}

	_, err := s.v2Client.PutUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *OssService) DeleteUserDefinedLogFields(bucketName string) error {
	request := &ossv2.DeleteUserDefinedLogFieldsConfigRequest{
		Bucket: StringPointer(bucketName),
	}

	_, err := s.v2Client.DeleteUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *OssService) GetUserDefinedLogFieldsV2(bucketName string) (*ossv2.GetUserDefinedLogFieldsConfigResult, error) {
	request := &ossv2.GetUserDefinedLogFieldsConfigRequest{
		Bucket: StringPointer(bucketName),
	}

	result, err := s.v2Client.GetUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}
