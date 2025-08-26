package alicloud

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketTransferAcceleration related functions

func (s *OssService) DescribeOssBucketTransferAcceleration(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketTransferAcceleration(&oss.GetBucketTransferAccelerationRequest{
		Bucket: id,
	})
	if err != nil {
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(NotFoundErr("BucketTransferAcceleration", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketTransferAcceleration", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("BucketTransferAcceleration", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["TransferAccelerationConfiguration"] = config
	return result, nil
}

func (s *OssService) OssBucketTransferAccelerationStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketTransferAcceleration(id)
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
