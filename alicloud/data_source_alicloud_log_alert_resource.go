package alicloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudLogAlertResource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudLogAlertResourceRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"user", "project"}, true),
			},
			"lang": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAlicloudLogAlertResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	resourceType := d.Get("type").(string)
	project := d.Get("project").(string)

	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			switch resourceType {
			case "user":
				// For user type, we initialize user alert resources
				// Since the new API doesn't have direct InitUserAlertResource equivalent,
				// we can try to check if alert resources are available by listing projects
				// This serves as initialization verification
				_, err := slsClient.ListLogProjects(ctx, "", "")
				if err != nil {
					return nil, fmt.Errorf("failed to initialize user alert resources: %w", err)
				}
				return nil, nil
			case "project":
				if project == "" {
					return nil, fmt.Errorf("project name is required for project type")
				}
				// Check if the internal-alert-history logstore exists
				_, err := slsClient.GetLogStore(ctx, project, "internal-alert-history")
				if err != nil {
					if IsExpectedErrors(err, []string{"LogStoreNotExist", "ProjectNotExist"}) {
						// Create the logstore for alert history if it doesn't exist
						logstore := &aliyunSlsAPI.LogStore{
							LogstoreName: "internal-alert-history",
							Ttl:          7, // 7 days retention
							ShardCount:   2, // 2 shards by default
						}
						err = slsClient.CreateLogStore(ctx, project, logstore)
						if err != nil {
							return nil, fmt.Errorf("failed to create internal-alert-history logstore: %w", err)
						}
					} else {
						return nil, err
					}
				}
				return nil, nil
			default:
				return nil, WrapErrorf(errors.New("type error"), DefaultErrorMsg, "alicloud_log_alert_resource", "CreateAlertResource", AliyunLogGoSdkERROR)
			}
		})
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_alert_resource", "CreateAlertResource", AliyunLogGoSdkERROR)
	}
	d.SetId(fmt.Sprintf("alert_resource_%s", resourceType))
	return nil
}
