package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	ossapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// BucketCors related functions

// GetBucketCors gets bucket CORS configuration using cws-lib-go API
func (s *OssService) GetBucketCors(bucketName string) (*ossapi.CORSConfiguration, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	corsConfig, err := ossAPI.GetBucketCors(bucketName)
	if err != nil {
		if ossNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, "OSS API")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, bucketName, "GetBucketCors", "OSS API")
	}

	return corsConfig, nil
}

// SetBucketCors sets bucket CORS configuration using cws-lib-go API
func (s *OssService) SetBucketCors(bucketName string, corsConfig *ossapi.CORSConfiguration) error {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return WrapError(err)
	}

	err = ossAPI.SetBucketCors(bucketName, corsConfig)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, bucketName, "SetBucketCors", "OSS API")
	}

	return nil
}

// DeleteBucketCors deletes bucket CORS configuration using cws-lib-go API
func (s *OssService) DeleteBucketCors(bucketName string) error {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return WrapError(err)
	}

	err = ossAPI.DeleteBucketCors(bucketName)
	if err != nil {
		if ossNotFoundError(err) {
			return nil // CORS configuration already deleted
		}
		return WrapErrorf(err, DefaultErrorMsg, bucketName, "DeleteBucketCors", "OSS API")
	}

	return nil
}

func (s *OssService) DescribeOssBucketCors(id string) (object map[string]interface{}, err error) {
	// Try to use new cws-lib-go API first
	ossAPI, apiErr := s.GetOssAPI()
	if apiErr == nil {
		corsConfig, err := ossAPI.GetBucketCors(id)
		if err == nil {
			// Convert CORS configuration to expected format
			object = make(map[string]interface{})
			object["CORSRule"] = convertCorsConfigToLegacy(corsConfig)
			return object, nil
		}
		// Fall back to old implementation on API error unless it's not found
		if !ossNotFoundError(err) {
			// Continue to fallback for other errors
		} else {
			return object, WrapErrorf(err, NotFoundMsg, "OSS API")
		}
	}

	// Fallback to original implementation
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	action := fmt.Sprintf("/?cors")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketCors", action), query, nil, nil, hostMap, false)

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
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchCORSConfiguration"}) {
			return object, WrapErrorf(NotFoundErr("BucketCors", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketCors", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.CORSConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.CORSConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

// convertCorsConfigToLegacy converts cws-lib-go CORS config to legacy format
func convertCorsConfigToLegacy(corsConfig *ossapi.CORSConfiguration) interface{} {
	// Implement conversion logic here
	// For now, return a mock structure
	rules := make([]map[string]interface{}, 0)
	if corsConfig != nil && len(corsConfig.CORSRules) > 0 {
		for _, rule := range corsConfig.CORSRules {
			legacyRule := map[string]interface{}{
				"AllowedMethod": rule.AllowedMethods,
				"AllowedOrigin": rule.AllowedOrigins,
				"AllowedHeader": rule.AllowedHeaders,
				"ExposeHeader":  rule.ExposeHeaders,
				"MaxAgeSeconds": rule.MaxAgeSeconds,
			}
			rules = append(rules, legacyRule)
		}
	}
	return rules
}

func (s *OssService) OssBucketCorsStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCors(id)
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
