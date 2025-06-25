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

// AccessPoint related functions

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

