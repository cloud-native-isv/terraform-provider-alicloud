package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
)

func NewCmsService(client *connectivity.AliyunClient) (*CmsService, error) {
	return &CmsService{client: client}, nil
}

// cmsAPI constructs the cws-lib-go CMS 2.0 API handle from the Terraform
// client credentials. The handle is built lazily per call so that the
// shared CmsService struct (defined in service_alicloud_cms.go for the
// CMS 1.0 resources) keeps its single-field layout untouched.
func (s *CmsService) cmsAPI() (*cmsapi.CmsAPI, error) {
	credentials := &commonapi.Credentials{
		AccessKey:     s.client.AccessKey,
		SecretKey:     s.client.SecretKey,
		RegionId:      s.client.RegionId,
		SecurityToken: s.client.SecurityToken,
	}
	api, err := cmsapi.NewCmsAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go CmsAPI: %w", err)
	}
	return api, nil
}

// cmsServiceCapabilityBlocked is retained only for the CMS 2.0 object groups
// that are not implemented yet. Implemented groups call the cws-lib-go cms
// API instead of returning this error.
func cmsServiceCapabilityBlocked(operation string) error {
	return fmt.Errorf("cms service capability %s is blocked by unresolved cws-lib-go cms compile issues", operation)
}
