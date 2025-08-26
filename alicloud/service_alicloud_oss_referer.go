package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketReferer related functions

func (s *OssService) DescribeOssBucketReferer(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketReferer(&oss.GetBucketRefererRequest{
		Bucket: id,
	})
	if err != nil {
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketReferer", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketReferer", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketReferer", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["RefererConfiguration"] = config
	return result, nil
}

func (s *OssService) OssBucketRefererStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketReferer(id)
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
