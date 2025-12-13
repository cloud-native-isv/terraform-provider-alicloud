package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketLogging related functions

func (s *OssService) DescribeOssBucketLogging(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketLogging(id)
	if err != nil {
		if NotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketLogging", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["BucketLoggingStatus"] = map[string]interface{}{
		"LoggingEnabled": config,
	}
	return result, nil
}

func (s *OssService) OssBucketLoggingStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketLogging(id)
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
