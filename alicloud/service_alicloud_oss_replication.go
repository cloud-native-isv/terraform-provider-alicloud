package alicloud

import (
	"encoding/json"
	"fmt"
)

// BucketReplication related functions

func (s *OssService) DescribeOssBucketReplication(id string) (response string, err error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return response, WrapError(err)
	}
	bucket := parts[0]
	ruleId := parts[1]

	ossAPI := s.GetOssAPI()
	if ossAPI == nil {
		return "", WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketReplication(bucket)
	if err != nil {
		if IsNotFoundError(err) {
			return response, WrapErrorf(err, NotFoundMsg, AliyunOssGoSdk)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketReplication", AliyunOssGoSdk)
	}

	if config == nil {
		return "", WrapErrorf(NotFoundErr("BucketReplication", id), NotFoundMsg, "config is nil")
	}

	// Convert config to string format for backward compatibility
	if configStr, ok := config.(string); ok {
		return configStr, nil
	}

	// If config is not string, convert to JSON string
	if jsonBytes, err := json.Marshal(config); err == nil {
		return string(jsonBytes), nil
	}

	return "", WrapError(fmt.Errorf("failed to convert config to string"))
}
