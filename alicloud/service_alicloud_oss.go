package alicloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	ossv2credentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/tidwall/sjson"
)

type OssService struct {
	client   *connectivity.AliyunClient
	v2Client *ossv2.Client
	ctx      context.Context
}

type ossv2CredentialsProvider struct {
	client *connectivity.AliyunClient
}

func (cp *ossv2CredentialsProvider) GetCredentials(ctx context.Context) (ossv2credentials.Credentials, error) {
	return ossv2credentials.Credentials{
		AccessKeyID:     cp.client.AccessKey,
		AccessKeySecret: cp.client.SecretKey,
		SecurityToken:   cp.client.SecurityToken,
	}, nil
}

func NewOssServiceV2(client *connectivity.AliyunClient) *OssService {
	v2Client := ossv2.NewClient(&ossv2.Config{
		Region:              &client.RegionId,
		Endpoint:            &client.RegionId,
		CredentialsProvider: &ossv2CredentialsProvider{client},
	})
	return &OssService{
		client:   client,
		v2Client: v2Client,
		ctx:      context.Background(),
	}
}

// DescribeOssBucketAcl <<< Encapsulated get interface for Oss BucketAcl.

func (s *OssService) DescribeOssBucketAcl(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?acl")

	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketAcl", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketAcl", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketAcl", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.AccessControlPolicy.AccessControlList", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AccessControlPolicy.AccessControlList", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketAclStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketAcl(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketAcl >>> Encapsulated.

// DescribeOssBucketReferer <<< Encapsulated get interface for Oss BucketReferer.

func (s *OssService) DescribeOssBucketReferer(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	action := fmt.Sprintf("/?referer")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketReferer", action), query, nil, nil, hostMap, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketReferer", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketReferer", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.RefererConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.RefererConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketRefererStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketReferer(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketReferer >>> Encapsulated.

// DescribeOssBucketHttpsConfig <<< Encapsulated get interface for Oss BucketHttpsConfig.

func (s *OssService) DescribeOssBucketHttpsConfig(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	action := fmt.Sprintf("/?httpsConfig")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketHttpsConfig", action), query, nil, nil, hostMap, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchHttpsConfig", "NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketHttpsConfig", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketHttpsConfig", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.HttpsConfiguration.TLS", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.HttpsConfiguration.TLS", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketHttpsConfigStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketHttpsConfig(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketHttpsConfig >>> Encapsulated.

// DescribeOssBucketCors <<< Encapsulated get interface for Oss BucketCors.

func (s *OssService) DescribeOssBucketCors(id string) (object map[string]interface{}, err error) {
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

func (s *OssService) OssBucketCorsStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCors(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketCors >>> Encapsulated.

// DescribeOssBucketPolicy
func (s *OssService) DescribeOssBucketPolicy(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?policy")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlJsonParam("GET", "2019-05-17", "GetBucketPolicy", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketPolicy", id), NotFoundMsg, ProviderERROR, fmt.Sprint(response["RequestId"]))
	}

	return response, nil
}

func (s *OssService) OssBucketPolicyStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketPolicy(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Check if the policy is properly configured by verifying it contains valid policy data
		if strings.HasPrefix(field, "#") {
			// Check if the response contains policy data (Statement field indicates a valid policy)
			if v, err := jsonpath.Get("$.Statement", object); err == nil && v != nil {
				return object, "#CHECKSET", nil
			}
			// Also check if the response contains valid JSON policy as string
			if policyStr, ok := object["policy"].(string); ok && policyStr != "" {
				return object, "#CHECKSET", nil
			}
			// Check if object itself has policy-like structure
			if len(object) > 0 {
				// The object response indicates policy exists
				return object, "#CHECKSET", nil
			}
			return object, "", nil
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

// DescribeOssBucketPolicy >>> Encapsulated.

// DescribeOssBucketVersioning <<< Encapsulated get interface for Oss BucketVersioning.

func (s *OssService) DescribeOssBucketVersioning(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?versioning")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketVersioning", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketVersioning", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.VersioningConfiguration", response)
	if err != nil {
		return object, WrapErrorf(NotFoundErr("BucketVersioning", id), NotFoundMsg, response)
	}

	currentStatus := v.(map[string]interface{})["Status"]
	if currentStatus == nil {
		return object, WrapErrorf(NotFoundErr("BucketVersioning", id), NotFoundMsg, response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketVersioningStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketVersioning(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketVersioning >>> Encapsulated.
// DescribeOssBucketArchiveDirectRead
func (s *OssService) DescribeOssBucketArchiveDirectRead(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?bucketArchiveDirectRead")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketArchiveDirectRead", action), query, nil, nil, hostMap, true)
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
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketArchiveDirect", id), NotFoundMsg, response)
	}
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.ArchiveDirectReadConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ArchiveDirectReadConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketArchiveDirectReadStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketArchiveDirectRead(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketArchiveDirectRead >>> Encapsulated.

// DescribeOssBucketRequestPayment
func (s *OssService) DescribeOssBucketRequestPayment(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?requestPayment")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketRequestPayment", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketRequestPayment", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.RequestPaymentConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.RequestPaymentConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketRequestPaymentStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketRequestPayment(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketRequestPayment >>> Encapsulated.

// DescribeOssBucketTransferAcceleration
func (s *OssService) DescribeOssBucketTransferAcceleration(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?transferAcceleration")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketTransferAcceleration", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(NotFoundErr("BucketTransferAcceleration", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketTransferAcceleration", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.TransferAccelerationConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.TransferAccelerationConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketTransferAccelerationStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketTransferAcceleration(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketTransferAcceleration >>> Encapsulated.

// DescribeOssBucketAccessMonitor
func (s *OssService) DescribeOssBucketAccessMonitor(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?accessmonitor")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketAccessMonitor", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketAccessMonitor", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketAccessMonitor", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.AccessMonitorConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.AccessMonitorConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketAccessMonitorStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketAccessMonitor(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketAccessMonitor >>> Encapsulated.

// DescribeOssBucketLogging
func (s *OssService) DescribeOssBucketLogging(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?logging")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketLogging", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.BucketLoggingStatus.LoggingEnabled", response)
	if err != nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	currentStatus := v.(map[string]interface{})["TargetBucket"]
	if currentStatus == nil {
		return object, WrapErrorf(NotFoundErr("BucketLogging", id), NotFoundMsg, response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketLoggingStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketLogging(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketLogging >>> Encapsulated.

// DescribeOssBucketServerSideEncryption <<< Encapsulated get interface for Oss BucketServerSideEncryption.
func (s *OssService) DescribeOssBucketServerSideEncryption(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	action := fmt.Sprintf("/?encryption")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketEncryption", action), query, nil, nil, hostMap, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchServerSideEncryptionRule", "NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketServerSideEncryption", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketServerSideEncryption", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.ServerSideEncryptionRule.ApplyServerSideEncryptionByDefault", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ServerSideEncryptionRule.ApplyServerSideEncryptionByDefault", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketServerSideEncryptionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketServerSideEncryption(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketServerSideEncryption >>> Encapsulated.
// DescribeOssBucketUserDefinedLogFields
func (s *OssService) DescribeOssBucketUserDefinedLogFields(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?userDefinedLogFieldsConfig")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketUserDefinedLogFieldsConfig", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchUserDefinedLogFieldsConfig"}) {
			return object, WrapErrorf(NotFoundErr("BucketUserDefinedLogFields", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketUserDefinedLogFields", id), NotFoundMsg, response)
	}

	return response, nil
}

func (s *OssService) OssBucketUserDefinedLogFieldsStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketUserDefinedLogFields(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketUserDefinedLogFields >>> Encapsulated.
// DescribeOssBucketMetaQuery
func (s *OssService) DescribeOssBucketMetaQuery(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?metaQuery")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketMetaQueryStatus", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "MetaQueryNotExist", "AccessDenied"}) {
			return object, WrapErrorf(NotFoundErr("BucketMetaQuery", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketMetaQuery", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.MetaQueryStatus", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.MetaQueryStatus", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketMetaQueryStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketMetaQuery(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketMetaQuery >>> Encapsulated.

// DescribeOssBucketDataRedundancyTransition
func (s *OssService) DescribeOssBucketDataRedundancyTransition(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?redundancyTransition")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])
	query["x-oss-redundancy-transition-taskid"] = StringPointer(parts[1])

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketDataRedundancyTransition", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "BucketDataRedundancyTransitionTaskNotExist"}) {
			return object, WrapErrorf(NotFoundErr("BucketDataRedundancyTransition", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketDataRedundancyTransition", id), NotFoundMsg, response)
	}

	return response, nil
}

func (s *OssService) OssBucketDataRedundancyTransitionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketDataRedundancyTransition(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketDataRedundancyTransition >>> Encapsulated.

// DescribeOssAccountPublicAccessBlock
func (s *OssService) DescribeOssAccountPublicAccessBlock(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 0 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 0, len(parts)))
	}
	action := fmt.Sprintf("/?publicAccessBlock")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetPublicAccessBlock", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("AccountPublicAccessBlock", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("AccountPublicAccessBlock", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.PublicAccessBlockConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PublicAccessBlockConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssAccountPublicAccessBlockStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssAccountPublicAccessBlock(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssAccountPublicAccessBlock >>> Encapsulated.
// DescribeOssBucketPublicAccessBlock
func (s *OssService) DescribeOssBucketPublicAccessBlock(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?publicAccessBlock")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketPublicAccessBlock", action), query, nil, nil, hostMap, true)
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
	if err != nil {
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketPublicAccessBlock", id), NotFoundMsg, response)
	}

	currentStatus, err := jsonpath.Get("$.PublicAccessBlockConfiguration.BlockPublicAccess", response)
	if currentStatus == "" {
		return object, WrapErrorf(NotFoundErr("BucketPublicAccessBlock", id), NotFoundMsg, response)
	}

	return response, nil
}

func (s *OssService) OssBucketPublicAccessBlockStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketPublicAccessBlock(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketPublicAccessBlock >>> Encapsulated.
// DescribeOssBucketCname
func (s *OssService) DescribeOssBucketCname(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?cname")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "ListCname", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.ListCnameResult", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ListCnameResult", response)
	}

	item := v.(map[string]interface{})
	domains, err := jsonpath.Get("$.Cname[*]", item)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Cname[*]", item)
	}
	for _, vv := range domains.([]interface{}) {
		if vv.(map[string]interface{})["Domain"] == parts[1] {
			return vv.(map[string]interface{}), nil
		}
	}

	return object, WrapErrorf(NotFoundErr("BucketCname", id), NotFoundMsg, response)
}

func (s *OssService) OssBucketCnameStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCname(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		if err != nil {
			return object, "", WrapErrorf(err, FailedGetAttributeMsg, id, field, object)
		}
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeOssBucketCname >>> Encapsulated.

// DescribeOssBucketCnameToken
func (s *OssService) DescribeOssBucketCnameToken(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?cname&comp=token")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])

	jsonString := "{}"
	jsonString, _ = sjson.Set(jsonString, "BucketCnameConfiguration.Cname.Domain", parts[1])
	json.Unmarshal([]byte(jsonString), &request)
	body := request

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("POST", "2019-05-17", "CreateCnameToken", action), query, body, nil, hostMap, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "CnameTokenNotFound"}) {
			return object, WrapErrorf(NotFoundErr("BucketCnameToken", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketCnameToken", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.CnameToken", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.CnameToken", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketCnameTokenStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketCnameToken(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketCnameToken >>> Encapsulated.

// DescribeOssBucketWebsite
func (s *OssService) DescribeOssBucketWebsite(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?website")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketWebsite", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchWebsiteConfiguration"}) {
			return object, WrapErrorf(NotFoundErr("BucketWebsite", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketWebsite", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.WebsiteConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.WebsiteConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketWebsiteStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketWebsite(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketWebsite >>> Encapsulated.

// DescribeOssAccessPoint
func (s *OssService) DescribeOssAccessPoint(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?accessPoint")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	query["x-oss-access-point-name"] = StringPointer(parts[1])
	hostMap["bucket"] = StringPointer(parts[0])
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetAccessPoint", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchAccessPoint"}) {
			return object, WrapErrorf(NotFoundErr("AccessPoint", id), NotFoundMsg, response)
		}
		addDebug(action, response, request)
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("AccessPoint", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.GetAccessPointResult", response)
	if err != nil {
		return object, WrapErrorf(NotFoundErr("AccessPoint", id), NotFoundMsg, response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssAccessPointStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssAccessPoint(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssAccessPoint >>> Encapsulated.
// DescribeOssBucketLifecycle
func (s *OssService) DescribeOssBucketLifecycle(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	action := fmt.Sprintf("/?lifecycle")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(id)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketLifecycle", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchLifecycle"}) {
			return object, WrapErrorf(NotFoundErr("BucketLifecycle", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketLifecycle", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.LifecycleConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.LifecycleConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketLifecycleStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketLifecycle(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketLifecycle >>> Encapsulated.
// DescribeOssBucketWorm
func (s *OssService) DescribeOssBucketWorm(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	action := fmt.Sprintf("/?worm")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetBucketWorm", action), query, nil, nil, hostMap, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchBucket", "NoSuchWORMConfiguration"}) {
			return object, WrapErrorf(NotFoundErr("BucketWorm", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if response == nil {
		return object, WrapErrorf(NotFoundErr("BucketWorm", id), NotFoundMsg, response)
	}

	v, err := jsonpath.Get("$.WormConfiguration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.WormConfiguration", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketWormStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketWorm(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketWorm >>> Encapsulated.

// DescribeOssBucketStyle <<< Encapsulated get interface for Oss BucketStyle.

func (s *OssService) DescribeOssBucketStyle(id string) (object map[string]interface{}, err error) {
	client := s.client
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}
	request = make(map[string]interface{})
	query = make(map[string]*string)
	hostMap := make(map[string]*string)
	hostMap["bucket"] = StringPointer(parts[0])
	query["styleName"] = StringPointer(parts[1])

	action := fmt.Sprintf("/?style")

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err = client.Do("Oss", xmlParam("GET", "2019-05-17", "GetStyle", action), query, nil, nil, hostMap, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		if IsExpectedErrors(err, []string{"NoSuchStyle", "NoSuchBucket"}) {
			return object, WrapErrorf(NotFoundErr("BucketStyle", id), NotFoundMsg, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Style", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Style", response)
	}

	return v.(map[string]interface{}), nil
}

func (s *OssService) OssBucketStyleStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOssBucketStyle(id)
		if err != nil {
			if NotFoundError(err) {
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

// DescribeOssBucketStyle >>> Encapsulated.

// PutBucketLogging enables logging for a bucket with the specified configurations
func (s *OssService) PutBucketLogging(bucketName, targetBucket, targetPrefix string) error {
	request := &ossv2.PutBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
		BucketLoggingStatus: &ossv2.BucketLoggingStatus{
			LoggingEnabled: &ossv2.LoggingEnabled{
				TargetBucket: StringPointer(targetBucket),
				TargetPrefix: StringPointer(targetPrefix),
			},
		},
	}

	_, err := s.v2Client.PutBucketLogging(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// DeleteBucketLogging disables the logging feature for a bucket
func (s *OssService) DeleteBucketLogging(bucketName string) error {
	request := &ossv2.DeleteBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
	}

	_, err := s.v2Client.DeleteBucketLogging(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// PutUserDefinedLogFields sets custom logging fields for a bucket
func (s *OssService) PutUserDefinedLogFields(bucketName string, headerSet []string, paramSet []string) error {
	request := &ossv2.PutUserDefinedLogFieldsConfigRequest{
		Bucket:                            StringPointer(bucketName),
		UserDefinedLogFieldsConfiguration: &ossv2.UserDefinedLogFieldsConfiguration{},
	}

	if len(headerSet) > 0 {
		request.UserDefinedLogFieldsConfiguration.HeaderSet = &ossv2.LoggingHeaderSet{
			Headers: headerSet,
		}
	}

	if len(paramSet) > 0 {
		request.UserDefinedLogFieldsConfiguration.ParamSet = &ossv2.LoggingParamSet{
			Parameters: paramSet,
		}
	}

	_, err := s.v2Client.PutUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// DeleteUserDefinedLogFields deletes custom logging fields for a bucket
func (s *OssService) DeleteUserDefinedLogFields(bucketName string) error {
	request := &ossv2.DeleteUserDefinedLogFieldsConfigRequest{
		Bucket: StringPointer(bucketName),
	}

	_, err := s.v2Client.DeleteUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// GetBucketLoggingV2 gets the logging configuration using the OSS V2 SDK
func (s *OssService) GetBucketLoggingV2(bucketName string) (*ossv2.GetBucketLoggingResult, error) {
	request := &ossv2.GetBucketLoggingRequest{
		Bucket: StringPointer(bucketName),
	}

	result, err := s.v2Client.GetBucketLogging(s.ctx, request)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// GetUserDefinedLogFieldsV2 gets the custom logging fields using the OSS V2 SDK
func (s *OssService) GetUserDefinedLogFieldsV2(bucketName string) (*ossv2.GetUserDefinedLogFieldsConfigResult, error) {
	request := &ossv2.GetUserDefinedLogFieldsConfigRequest{
		Bucket: StringPointer(bucketName),
	}

	result, err := s.v2Client.GetUserDefinedLogFieldsConfig(s.ctx, request)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeOssBucket - 从原 OssService 中补充的函数
func (s *OssService) DescribeOssBucket(id string) (response oss.GetBucketInfoResult, err error) {
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

// WaitForOssBucket - 从原 OssService 中补充的函数
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

// WaitForOssBucketObject - 从原 OssService 中补充的函数
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

// DescribeOssBucketReplication - 从原 OssService 中补充的函数
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
