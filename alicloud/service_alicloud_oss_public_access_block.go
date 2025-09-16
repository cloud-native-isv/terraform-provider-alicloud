package alicloud

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketPublicAccessBlock related functions

func (s *OssService) DescribeOssBucketPublicAccessBlock(id string) (object map[string]interface{}, err error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}
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
	// Convert the response object to map[string]interface{}
	publicAccessBlockConfig := make(map[string]interface{})
	if config.PublicAccessBlockConfiguration != nil {
		// Map the cws-lib-go fields to the expected BlockPublicAccess field
		// If any of the public access blocking options are enabled, we consider BlockPublicAccess as true
		blockPublicAccess := false
		if config.PublicAccessBlockConfiguration.BlockPublicAcls != nil && *config.PublicAccessBlockConfiguration.BlockPublicAcls {
			blockPublicAccess = true
		}
		if config.PublicAccessBlockConfiguration.IgnorePublicAcls != nil && *config.PublicAccessBlockConfiguration.IgnorePublicAcls {
			blockPublicAccess = true
		}
		if config.PublicAccessBlockConfiguration.BlockPublicPolicy != nil && *config.PublicAccessBlockConfiguration.BlockPublicPolicy {
			blockPublicAccess = true
		}
		if config.PublicAccessBlockConfiguration.RestrictPublicBuckets != nil && *config.PublicAccessBlockConfiguration.RestrictPublicBuckets {
			blockPublicAccess = true
		}
		publicAccessBlockConfig["BlockPublicAccess"] = blockPublicAccess
	}
	result["PublicAccessBlockConfiguration"] = publicAccessBlockConfig
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
