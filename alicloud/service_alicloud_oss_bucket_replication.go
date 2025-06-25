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

