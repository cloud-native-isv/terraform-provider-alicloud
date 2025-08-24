package alicloud

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketPublicAccessBlock related functions

func (s *OssService) DescribeOssBucketPublicAccessBlock(id string) (object map[string]interface{}, err error) {
	ossAPI := s.GetOssAPI()
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketPublicAccessBlock(id)
	if err != nil {
		if IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("BucketPublicAccessBlock", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketPublicAccessBlock", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketPublicAccessBlock", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["PublicAccessBlockConfiguration"] = config
	return result, nil
}

func (s *OssService) OssBucketPublicAccessBlockStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketPublicAccessBlock(id)
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
