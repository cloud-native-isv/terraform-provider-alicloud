// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsPrometheusRemoteWrite() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsRemoteWriteCreate,
		Read:   resourceAliCloudArmsRemoteWriteRead,
		Update: resourceAliCloudArmsRemoteWriteUpdate,
		Delete: resourceAliCloudArmsRemoteWriteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_write_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_write_yaml": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					yaml, _ := normalizeYamlString(v)
					return yaml
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := compareYamlTemplateAreEquivalent(old, new)
					return equal
				},
				ValidateFunc: validateYamlString,
			},
		},
	}
}

func resourceAliCloudArmsRemoteWriteCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	remoteWriteYaml := d.Get("remote_write_yaml").(string)

	// Create the remote write using service layer
	var result *aliyunArmsAPI.PrometheusRemoteWrite
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = armsService.CreateArmsPrometheusRemoteWrite(clusterId, remoteWriteYaml)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_prometheus_remote_write", "CreateArmsPrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
	}

	// Set the resource ID as clusterId:remoteWriteName
	d.SetId(fmt.Sprintf("%s:%s", SafeStringValue(result.ClusterId), SafeStringValue(result.RemoteWriteName)))

	return resourceAliCloudArmsRemoteWriteRead(d, meta)
}

func resourceAliCloudArmsRemoteWriteRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	objectRaw, err := armsService.DescribeArmsRemoteWrite(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_prometheus_remote_write DescribeArmsRemoteWrite Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("remote_write_yaml", objectRaw["RemoteWriteYaml"])
	d.Set("cluster_id", objectRaw["ClusterId"])
	d.Set("remote_write_name", objectRaw["RemoteWriteName"])

	return nil
}

func resourceAliCloudArmsRemoteWriteUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse ID parts
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(Error("invalid resource ID format"))
	}
	clusterId := parts[0]
	remoteWriteName := parts[1]

	update := false
	var remoteWriteYaml string

	if d.HasChange("remote_write_yaml") {
		update = true
		remoteWriteYaml = d.Get("remote_write_yaml").(string)
	}

	if update {
		// Update using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err = armsService.UpdateArmsPrometheusRemoteWrite(clusterId, remoteWriteName, remoteWriteYaml)
			if err != nil {
				if NeedRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateArmsPrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudArmsRemoteWriteRead(d, meta)
}

func resourceAliCloudArmsRemoteWriteDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create Arms service
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse ID parts
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(Error("invalid resource ID format"))
	}
	clusterId := parts[0]
	remoteWriteName := parts[1]

	// Delete using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = armsService.DeleteArmsPrometheusRemoteWrite(clusterId, []string{remoteWriteName})
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteArmsPrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
	}

	return nil
}
