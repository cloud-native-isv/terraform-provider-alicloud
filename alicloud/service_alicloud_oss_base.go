package alicloud

import (
	"context"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	ossv2credentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	ossapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/oss"
)

type OssService struct {
	client   *connectivity.AliyunClient
	v2Client *ossv2.Client
	ctx      context.Context
	ossAPI   *ossapi.OssAPI // New: cws-lib-go API client
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

func NewOssService(client *connectivity.AliyunClient) *OssService {
	v2Client := ossv2.NewClient(&ossv2.Config{
		Region:              &client.RegionId,
		Endpoint:            &client.RegionId,
		CredentialsProvider: &ossv2CredentialsProvider{client},
	})

	// Initialize cws-lib-go API client
	credentials := &common.Credentials{
		AccessKey: client.AccessKey,
		SecretKey: client.SecretKey,
		RegionId:  client.RegionId,
	}

	ossAPI, err := ossapi.NewOssAPI(credentials)
	if err != nil {
		// Log error but continue with legacy functionality
		// TODO: Handle this error better in production
		ossAPI = nil
	}

	return &OssService{
		client:   client,
		v2Client: v2Client,
		ctx:      context.Background(),
		ossAPI:   ossAPI,
	}
}

// GetAPI returns the OSS v2 Client instance for direct API access (legacy)
func (service *OssService) GetAPI() *ossv2.Client {
	// add some customize logic for this API object
	return service.v2Client
}

// GetOssAPI returns the cws-lib-go OSS API client
func (service *OssService) GetOssAPI() (*ossapi.OssAPI, error) {
	if service.ossAPI == nil {
		// Lazy initialization if not created during service creation
		credentials := &common.Credentials{
			AccessKey: service.client.AccessKey,
			SecretKey: service.client.SecretKey,
			RegionId:  service.client.RegionId,
		}

		ossAPI, err := ossapi.NewOssAPI(credentials)
		if err != nil {
			return nil, WrapError(err)
		}
		service.ossAPI = ossAPI
	}
	return service.ossAPI, nil
}

// Main OSS service file
// Individual resource functions have been moved to separate files:
// - service_alicloud_oss_bucketacl.go
// - service_alicloud_oss_bucketreferer.go
// - service_alicloud_oss_buckethttpsconfig.go
// - service_alicloud_oss_bucketcors.go
// - service_alicloud_oss_bucketpolicy.go
// - service_alicloud_oss_bucketversioning.go
// - service_alicloud_oss_bucketarchivedirectread.go
// - service_alicloud_oss_bucketrequestpayment.go
// - service_alicloud_oss_buckettransferacceleration.go
// - service_alicloud_oss_bucketaccessmonitor.go
// - service_alicloud_oss_bucketlogging.go
// - service_alicloud_oss_bucketserversideencryption.go
// - service_alicloud_oss_bucketuserdefinedlogfields.go
// - service_alicloud_oss_bucketmetaquery.go
// - service_alicloud_oss_bucketdataredundancytransition.go
// - service_alicloud_oss_accountpublicaccessblock.go
// - service_alicloud_oss_bucketpublicaccessblock.go
// - service_alicloud_oss_bucketcname.go
// - service_alicloud_oss_bucketcnametoken.go
// - service_alicloud_oss_bucketwebsite.go
// - service_alicloud_oss_accesspoint.go
// - service_alicloud_oss_bucketlifecycle.go
// - service_alicloud_oss_bucketworm.go
// - service_alicloud_oss_bucketstyle.go
// - service_alicloud_oss_bucket.go
// - service_alicloud_oss_bucketobject.go
// - service_alicloud_oss_bucketreplication.go
