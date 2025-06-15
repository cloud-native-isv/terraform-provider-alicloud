package alicloud

import (
	"context"
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const maxWaitTime = 60

func dataSourceAlicloudLogService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudLogServiceRead,

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"On", "Off"}, false),
				Optional:     true,
				Default:      "Off",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAlicloudLogServiceRead(d *schema.ResourceData, meta interface{}) error {
	if v, ok := d.GetOk("enable"); !ok || v.(string) != "On" {
		d.SetId("LogServiceHasNotBeenOpened")
		d.Set("status", "")
		return nil
	}

	// Check if service is ready by trying to list projects
	isServiceReady, err := checkLogServiceReady(meta)
	if err == nil && isServiceReady {
		d.SetId("LogServiceHasBeenOpened")
		d.Set("status", "Opened")
		return nil
	}

	// Note: Service activation via API is not currently supported in the new SLS API
	// Users should enable the service through the Alibaba Cloud console
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_service", "CheckLogService", AlibabaCloudSdkGoERROR)
	}

	// If service is not ready, return appropriate status
	d.SetId("LogServiceHasNotBeenOpened")
	d.Set("status", "NotOpened")
	return WrapError(fmt.Errorf("Log Service is not enabled. Please enable it through the Alibaba Cloud console first"))
}

func checkLogServiceReady(meta interface{}) (bool, error) {
	client := meta.(*connectivity.AliyunClient)

	_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
		ctx := context.Background()
		// Try to list projects to check if service is available
		projects, err := slsClient.ListLogProjects(ctx, "", "")
		return projects, err
	})

	if err != nil {
		// If we get specific service not enabled errors, return false
		if IsExpectedErrors(err, []string{"ServiceNotEnabled", "ServiceNotOpen", "Forbidden"}) {
			return false, nil
		}
		// Other errors should be returned
		return false, err
	}

	// If we can successfully call the API, service is ready
	return true, nil
}
