package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketServerSideEncryption related functions

func (s *OssService) DescribeOssBucketServerSideEncryption(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketEncryption(&oss.GetBucketEncryptionRequest{
		Bucket: id,
	})
	if err != nil {
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"NoSuchServerSideEncryptionRule", "NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketServerSideEncryption", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketEncryption", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketServerSideEncryption", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["ServerSideEncryptionRule"] = config
	return result, nil
}

func (s *OssService) OssBucketServerSideEncryptionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketServerSideEncryption(id)
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
