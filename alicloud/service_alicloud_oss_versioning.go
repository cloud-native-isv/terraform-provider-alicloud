package alicloud

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketVersioning related functions

func (s *OssService) DescribeOssBucketVersioning(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketVersioning(id)
	if err != nil {
		if IsNotFoundError(err) {
			return object, WrapErrorf(NotFoundErr("BucketVersioning", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketVersioning", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketVersioning", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["VersioningConfiguration"] = config
	return result, nil
}

func (s *OssService) OssBucketVersioningStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketVersioning(id)
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
