package alicloud

import (
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	ossapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
)

// BucketObject related functions

// PutOssObject uploads an object using cws-lib-go API
func (s *OssService) PutOssObject(bucketName, objectKey string, content []byte) (*ossapi.OssObjectInfo, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	objectInfo, err := ossAPI.PutObject(bucketName, objectKey, content)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, objectKey, "PutObject", "OSS API")
	}

	return objectInfo, nil
}

// GetOssObject gets an object using cws-lib-go API
func (s *OssService) GetOssObject(bucketName, objectKey string) (*ossapi.OssObjectInfo, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	objectInfo, err := ossAPI.GetObject(bucketName, objectKey)
	if err != nil {
		if ossNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, "OSS API")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, objectKey, "GetObject", "OSS API")
	}

	return objectInfo, nil
}

// DeleteOssObject deletes an object using cws-lib-go API
func (s *OssService) DeleteOssObject(bucketName, objectKey string) error {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return WrapError(err)
	}

	err = ossAPI.DeleteObject(bucketName, objectKey)
	if err != nil {
		if ossNotFoundError(err) {
			return nil // Object already deleted
		}
		return WrapErrorf(err, DefaultErrorMsg, objectKey, "DeleteObject", "OSS API")
	}

	return nil
}

// ListOssObjects lists objects using cws-lib-go API
func (s *OssService) ListOssObjects(bucketName, prefix string, maxKeys int32) ([]*ossapi.OssObjectInfo, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	objects, err := ossAPI.ListObjects(bucketName, prefix, maxKeys)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, bucketName, "ListObjects", "OSS API")
	}

	return objects, nil
}

// ObjectExists checks if object exists using cws-lib-go API
func (s *OssService) ObjectExists(bucketName, objectKey string) (bool, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return false, WrapError(err)
	}

	exists, err := ossAPI.ObjectExists(bucketName, objectKey)
	if err != nil {
		return false, WrapErrorf(err, DefaultErrorMsg, objectKey, "ObjectExists", "OSS API")
	}

	return exists, nil
}

func (s *OssService) WaitForOssBucketObject(bucket *oss.Bucket, id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		exist, err := bucket.IsObjectExist(id)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, id, "IsObjectExist", AliyunOssGoSdk)
		}
		addDebug("IsObjectExist", exist)

		if !exist {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, strconv.FormatBool(exist), status, ProviderERROR)
		}
	}
}

// WaitForOssBucketObjectNew waits for object status using cws-lib-go API (new method)
func (s *OssService) WaitForOssBucketObjectNew(bucketName, objectKey string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		exists, err := s.ObjectExists(bucketName, objectKey)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, objectKey, "ObjectExists", "OSS API")
		}

		if status == Deleted && !exists {
			return nil
		}
		if status != Deleted && exists {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, objectKey, GetFunc(1), timeout, strconv.FormatBool(exists), status, ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}
