package alicloud

import (
	"encoding/json"
	"fmt"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
)

// BucketReplication related functions

func (s *OssService) DescribeOssBucketReplication(id string) (response string, err error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return response, WrapError(err)
	}
	bucket := parts[0]
	_ = parts[1] // ruleId not used in current implementation

	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return "", WrapError(err)
	}
	if ossAPI == nil {
		return "", WrapError(fmt.Errorf("OSS API client not available"))
	}

	config, err := ossAPI.GetBucketReplication(&oss.GetBucketReplicationRequest{
		Bucket: bucket,
	})
	if err != nil {
		if IsNotFoundError(err) {
			return response, WrapErrorf(err, NotFoundMsg, AliyunOssGoSdk)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketReplication", AliyunOssGoSdk)
	}

	if config == nil {
		return "", WrapErrorf(NotFoundErr("BucketReplication", id), NotFoundMsg, "config is nil")
	}

	// Convert config response to string format for backward compatibility
	if config.ReplicationConfiguration != nil {
		if jsonBytes, err := json.Marshal(config.ReplicationConfiguration); err == nil {
			return string(jsonBytes), nil
		}
	}

	// If config has no replication configuration
	return "", WrapErrorf(NotFoundErr("BucketReplication", id), NotFoundMsg, "no replication configuration found")
}
