package alicloud

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketReplication related functions

func (s *OssService) DescribeOssBucketReplication(id string) (response string, err error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return response, WrapError(err)
	}
	bucket := parts[0]
	ruleId := parts[1]

	request := map[string]string{"bucketName": bucket, "ruleId": ruleId}
	var requestInfo *oss.Client
	raw, err := s.client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.GetBucketReplication(bucket)
	})
	if err != nil {
		if ossNotFoundError(err) {
			return response, WrapErrorf(err, NotFoundMsg, AliyunOssGoSdk)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketReplication", AliyunOssGoSdk)
	}

	addDebug("GetBucketReplication", raw, requestInfo, request)
	response, _ = raw.(string)
	return
}
