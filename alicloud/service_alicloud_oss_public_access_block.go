package alicloud

import (
	"fmt"
	"log"
	"time"

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

// PutOssBucketPublicAccessBlock creates or updates the bucket PublicAccessBlock configuration with retry/backoff.
// It centralizes error gating and retry strategy for write operations.
func (s *OssService) PutOssBucketPublicAccessBlock(bucket string, blockPublicAccess bool, timeout time.Duration) error {
	// Build request payload
	action := fmt.Sprintf("/?publicAccessBlock")
	request := make(map[string]interface{})
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(bucket)
	objectDataLocalMap := map[string]interface{}{
		"BlockPublicAccess": blockPublicAccess,
	}
	request["PublicAccessBlockConfiguration"] = objectDataLocalMap
	body := request
	query := make(map[string]*string)

	// Retry with exponential backoff + jitter
	wait := ExpBackoffWait(2*time.Second, 1.8, 0.25, 30*time.Second)
	attempts := 0
	var lastErr error
	var response map[string]interface{}

	err := resource.Retry(timeout, func() *resource.RetryError {
		attempts++
		// Use legacy Do call for now; future work: switch to cws-lib-go write API if available
		resp, err := s.client.Do("Oss", xmlParam("PUT", "2019-05-17", "PutBucketPublicAccessBlock", action), query, body, nil, hostMap, false)
		if err != nil {
			if IsOssConcurrentUpdateError(err) {
				lastErr = err
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		if lastErr != nil {
			// WARN level log for retry exhaustion
			log.Printf("[WARN] PutOssBucketPublicAccessBlock retries exhausted after %d attempts for bucket %s, last error: %v", attempts, bucket, lastErr)
		}
		return WrapErrorf(err, DefaultErrorMsg+": hint: OSS reported potential concurrency conflict (ConcurrentUpdateBucketFailed). Please retry after a short cooldown.", bucket, "PutBucketPublicAccessBlock", AlibabaCloudSdkGoERROR)
	}
	return nil
}
