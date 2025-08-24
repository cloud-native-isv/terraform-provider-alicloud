package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func (s *OssService) DescribeOssAccountPublicAccessBlock(id string) (object map[string]interface{}, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 0 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 0, len(parts)))
	}

	ossAPI := s.GetOssAPI()
	if ossAPI == nil {
		return nil, WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetAccountPublicAccessBlock()
	if err != nil {
		if IsNotFoundError(err) || IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("AccountPublicAccessBlock", id), NotFoundMsg, err)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetAccountPublicAccessBlock", AlibabaCloudSdkGoERROR)
	}

	if config == nil {
		return object, WrapErrorf(NotFoundErr("AccountPublicAccessBlock", id), NotFoundMsg, "config is nil")
	}

	result := make(map[string]interface{})
	result["PublicAccessBlockConfiguration"] = config
	return result, nil
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PublicAccessBlockConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssAccountPublicAccessBlockStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssAccountPublicAccessBlock(id)
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
