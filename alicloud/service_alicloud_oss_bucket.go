package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	ossapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
)

// Bucket related functions

// DescribeOssBucket gets bucket information using cws-lib-go API
func (s *OssService) DescribeOssBucket(id string) (response oss.GetBucketInfoResult, err error) {
	// Try to use new cws-lib-go API first
	ossAPI, apiErr := s.GetOssAPI()
	if apiErr == nil {
		bucketInfo, err := ossAPI.GetBucketInfo(id)
		if err != nil {
			// If bucket not found with new API, check specific error
			if ossNotFoundError(err) {
				return response, WrapErrorf(err, NotFoundMsg, "cws-lib-go OSS API")
			}
			// Fall back to old implementation on API error
		} else {
			// Convert cws-lib-go response to legacy format
			response = convertBucketInfoToLegacy(bucketInfo)
			return response, nil
		}
	}

	// Fallback to original implementation
	request := map[string]string{"bucketName": id}
	var requestInfo *oss.Client
	raw, err := s.client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.GetBucketInfo(request["bucketName"])
	})
	if err != nil {
		if ossNotFoundError(err) {
			return response, WrapErrorf(err, NotFoundMsg, AliyunOssGoSdk)
		}
		return response, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketInfo", AliyunOssGoSdk)
	}

	addDebug("GetBucketInfo", raw, requestInfo, request)
	response, _ = raw.(oss.GetBucketInfoResult)
	return
}

// DescribeOssBucketNew gets bucket information using only cws-lib-go API (new method)
func (s *OssService) DescribeOssBucketNew(id string) (*ossapi.OssBucketInfo, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	bucketInfo, err := ossAPI.GetBucketInfo(id)
	if err != nil {
		if ossNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, "OSS API")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetBucketInfo", "OSS API")
	}

	return bucketInfo, nil
}

// CreateOssBucket creates a bucket using cws-lib-go API
func (s *OssService) CreateOssBucket(bucketName string, config *ossapi.OssBucket) (*ossapi.OssBucket, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	bucket, err := ossAPI.CreateBucket(bucketName, config)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, bucketName, "CreateBucket", "OSS API")
	}

	return bucket, nil
}

// DeleteOssBucket deletes a bucket using cws-lib-go API
func (s *OssService) DeleteOssBucket(bucketName string) error {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return WrapError(err)
	}

	err = ossAPI.DeleteBucket(bucketName)
	if err != nil {
		if ossNotFoundError(err) {
			return nil // Bucket already deleted
		}
		return WrapErrorf(err, DefaultErrorMsg, bucketName, "DeleteBucket", "OSS API")
	}

	return nil
}

// ListOssBuckets lists buckets using cws-lib-go API
func (s *OssService) ListOssBuckets(prefix string, maxKeys int32) ([]*ossapi.OssBucket, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	buckets, err := ossAPI.ListBuckets(prefix, maxKeys)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "ListBuckets", "ListBuckets", "OSS API")
	}

	return buckets, nil
}

// BucketExists checks if bucket exists using cws-lib-go API
func (s *OssService) BucketExists(bucketName string) (bool, error) {
	ossAPI, err := s.GetOssAPI()
	if err != nil {
		return false, WrapError(err)
	}

	exists, err := ossAPI.BucketExists(bucketName)
	if err != nil {
		return false, WrapErrorf(err, DefaultErrorMsg, bucketName, "BucketExists", "OSS API")
	}

	return exists, nil
}

func (s *OssService) WaitForOssBucket(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeOssBucket(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
				// for delete bucket replication
			} else if status == Deleted && IsExpectedErrors(err, []string{"AccessDenied"}) {
				return nil
			} else {
				return WrapError(err)
			}
		}

		if object.BucketInfo.Name != "" && status != Deleted {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object.BucketInfo.Name, status, ProviderERROR)
		}
	}
}

// convertBucketInfoToLegacy converts cws-lib-go bucket info to legacy format
func convertBucketInfoToLegacy(bucketInfo *ossapi.OssBucketInfo) oss.GetBucketInfoResult {
	// Create a mock legacy response for compatibility
	// In a real implementation, this would properly convert the structures
	var result oss.GetBucketInfoResult
	if bucketInfo != nil && bucketInfo.Bucket != nil {
		result.BucketInfo.Name = getStringValue(bucketInfo.Bucket.Name)
		result.BucketInfo.Location = getStringValue(bucketInfo.Bucket.Location)
		if bucketInfo.Bucket.CreationDate != nil {
			result.BucketInfo.CreationDate, _ = time.Parse(time.RFC3339, *bucketInfo.Bucket.CreationDate)
		}
		result.BucketInfo.StorageClass = getStringValue(bucketInfo.Bucket.StorageClass)
		result.BucketInfo.RedundancyType = getStringValue(bucketInfo.Bucket.DataRedundancyType)
		result.BucketInfo.ExtranetEndpoint = getStringValue(bucketInfo.Bucket.ExtranetEndpoint)
		result.BucketInfo.IntranetEndpoint = getStringValue(bucketInfo.Bucket.IntranetEndpoint)
		// Convert ACL from nested structure
		if bucketInfo.Bucket.AccessControlList != nil {
			result.BucketInfo.ACL = getStringValue(bucketInfo.Bucket.AccessControlList.Grant)
		}
		// Convert Owner from nested structure
		if bucketInfo.Bucket.Owner != nil {
			result.BucketInfo.Owner.ID = getStringValue(bucketInfo.Bucket.Owner.Id)
			result.BucketInfo.Owner.DisplayName = getStringValue(bucketInfo.Bucket.Owner.DisplayName)
		}
	}

	return result
}

// getStringValue safely gets string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// EmptyBucket checks for objects, versions, delete markers and multipart uploads,
// and deletes them all.
func (s *OssService) EmptyBucket(bucketName string) error {
	action := "EmptyBucket"
	_, err := s.client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		bucket, err := ossClient.Bucket(bucketName)
		if err != nil {
			return nil, err
		}

		// 1. Abort all multipart uploads
		// Pagination loop for ListMultipartUploads
		keyMarker := ""
		uploadIdMarker := ""
		for {
			lmur, err := bucket.ListMultipartUploads(
				oss.KeyMarker(keyMarker),
				oss.UploadIDMarker(uploadIdMarker),
				oss.MaxUploads(1000),
			)
			if err != nil {
				return nil, fmt.Errorf("listing multipart uploads failed: %w", err)
			}

			for _, upload := range lmur.Uploads {
				err = bucket.AbortMultipartUpload(oss.InitiateMultipartUploadResult{
					Bucket:   bucketName,
					Key:      upload.Key,
					UploadID: upload.UploadID,
				})
				if err != nil {
					return nil, fmt.Errorf("aborting multipart upload %s/%s failed: %w", upload.Key, upload.UploadID, err)
				}
			}

			if !lmur.IsTruncated {
				break
			}
			keyMarker = lmur.NextKeyMarker
			uploadIdMarker = lmur.NextUploadIDMarker
		}

		// 2. Delete all object versions and delete markers
		// Pagination loop for ListObjectVersions
		keyMarker = ""
		versionIdMarker := ""
		for {
			lovr, err := bucket.ListObjectVersions(
				oss.KeyMarker(keyMarker),
				oss.VersionIdMarker(versionIdMarker),
				oss.MaxKeys(1000),
			)
			if err != nil {
				return nil, fmt.Errorf("listing object versions failed: %w", err)
			}

			// Collect all objects and delete markers to delete
			var objectsToDelete []oss.DeleteObject
			for _, obj := range lovr.ObjectVersions {
				objectsToDelete = append(objectsToDelete, oss.DeleteObject{
					Key:       obj.Key,
					VersionId: obj.VersionId,
				})
			}
			for _, dm := range lovr.ObjectDeleteMarkers {
				objectsToDelete = append(objectsToDelete, oss.DeleteObject{
					Key:       dm.Key,
					VersionId: dm.VersionId,
				})
			}

			// Batch delete if there are any
			if len(objectsToDelete) > 0 {
				_, err = bucket.DeleteObjectVersions(objectsToDelete, oss.DeleteObjectsQuiet(true))
				if err != nil {
					return nil, fmt.Errorf("deleting object versions failed: %w", err)
				}
				log.Printf("[DEBUG] Deleted %d objects/versions from bucket %s", len(objectsToDelete), bucketName)
			}

			if !lovr.IsTruncated {
				break
			}
			keyMarker = lovr.NextKeyMarker
			versionIdMarker = lovr.NextVersionIdMarker
		}

		return nil, nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, bucketName, action, AliyunOssGoSdk)
	}

	return nil
}
