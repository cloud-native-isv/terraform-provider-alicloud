package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
)

func NewCmsService(client *connectivity.AliyunClient) (*CmsService, error) {
	return &CmsService{client: client}, nil
}

func cmsServiceCapabilityBlocked(operation string) error {
	return fmt.Errorf("cms service capability %s is blocked by unresolved cws-lib-go cms compile issues", operation)
}
