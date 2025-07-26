// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudNasAccessRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNasAccessRuleCreate,
		Read:   resourceAliCloudNasAccessRuleRead,
		Update: resourceAliCloudNasAccessRuleUpdate,
		Delete: resourceAliCloudNasAccessRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"access_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"access_rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"standard", "extreme"}, true),
			},
			"ipv6_source_cidr_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: IntBetween(0, 100),
			},
			"rw_access_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"RDWR", "RDONLY"}, true),
			},
			"source_cidr_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_access_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"no_squash", "root_squash", "all_squash"}, true),
			},
		},
	}
}

func resourceAliCloudNasAccessRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_rule", "NewNasService", AlibabaCloudSdkGoERROR)
	}

	// Prepare parameters for the service call
	accessGroupName := d.Get("access_group_name").(string)
	sourceCidrIp := ""
	if v, ok := d.GetOk("source_cidr_ip"); ok {
		sourceCidrIp = v.(string)
	}

	rwAccessType := "RDWR"
	if v, ok := d.GetOk("rw_access_type"); ok {
		rwAccessType = v.(string)
	}

	userAccessType := "no_squash"
	if v, ok := d.GetOk("user_access_type"); ok {
		userAccessType = v.(string)
	}

	priority := int32(1)
	if v, ok := d.GetOk("priority"); ok {
		priority = int32(v.(int))
	}

	fileSystemType := "standard"
	if v, ok := d.GetOk("file_system_type"); ok {
		fileSystemType = v.(string)
	}

	ipv6SourceCidrIp := ""
	if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok {
		ipv6SourceCidrIp = v.(string)
	}

	// Create access rule using service layer
	wait := incrementalWait(3*time.Second, 5*time.Second)
	var accessRule *aliyunNasAPI.AccessRule
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var err error
		accessRule, err = nasService.CreateNasAccessRule(accessGroupName, sourceCidrIp, rwAccessType, userAccessType, priority, fileSystemType, ipv6SourceCidrIp)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_rule", "CreateAccessRule", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using service helper
	resourceId := nasService.buildResourceId(accessGroupName, accessRule.AccessRuleId, fileSystemType)
	d.SetId(resourceId)

	// Set the computed attributes immediately after creation to ensure state consistency
	d.Set("access_rule_id", accessRule.AccessRuleId)
	d.Set("rw_access_type", rwAccessType)
	d.Set("user_access_type", userAccessType)

	return resourceAliCloudNasAccessRuleRead(d, meta)
}

func resourceAliCloudNasAccessRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	// Use service layer to get access rule details
	accessRule, err := nasService.DescribeNasAccessRule(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_access_rule DescribeNasAccessRule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Parse resource ID to get file system type
	accessGroupName, _, fileSystemType, err := nasService.parseResourceId(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}

	// Set attributes from the strongly-typed response
	d.Set("access_group_name", accessGroupName)
	d.Set("access_rule_id", accessRule.AccessRuleId)
	d.Set("source_cidr_ip", accessRule.SourceCidrIp)
	d.Set("priority", int(accessRule.Priority))
	d.Set("file_system_type", fileSystemType)

	// Handle optional fields with proper defaults to prevent plan drift
	// Set ipv6_source_cidr_ip - if empty from API, set as empty string (this is expected)
	d.Set("ipv6_source_cidr_ip", accessRule.Ipv6SourceCidrIp)

	// Set rw_access_type with default value if empty from API
	rwAccessType := accessRule.RWAccessType
	if rwAccessType == "" {
		rwAccessType = "RDWR"
	}
	d.Set("rw_access_type", rwAccessType)

	// Set user_access_type with default value if empty from API
	userAccessType := accessRule.UserAccessType
	if userAccessType == "" {
		userAccessType = "no_squash"
	}
	d.Set("user_access_type", userAccessType)

	return nil
}

func resourceAliCloudNasAccessRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	// Parse resource ID using service helper
	accessGroupName, accessRuleId, fileSystemType, err := nasService.parseResourceId(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}

	update := false

	// Check what fields have changed
	if d.HasChange("source_cidr_ip") || d.HasChange("priority") || d.HasChange("rw_access_type") ||
		d.HasChange("user_access_type") || d.HasChange("ipv6_source_cidr_ip") {
		update = true
	}

	if update {
		// Prepare parameters for the service call
		sourceCidrIp := ""
		if v, ok := d.GetOk("source_cidr_ip"); ok {
			sourceCidrIp = v.(string)
		}

		rwAccessType := ""
		if v, ok := d.GetOk("rw_access_type"); ok {
			rwAccessType = v.(string)
		}

		userAccessType := ""
		if v, ok := d.GetOk("user_access_type"); ok {
			userAccessType = v.(string)
		}

		priority := int32(0)
		if v, ok := d.GetOk("priority"); ok {
			priority = int32(v.(int))
		}

		ipv6SourceCidrIp := ""
		if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok {
			ipv6SourceCidrIp = v.(string)
		}

		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			err = nasService.UpdateNasAccessRule(accessGroupName, accessRuleId, sourceCidrIp, rwAccessType, userAccessType, priority, fileSystemType, ipv6SourceCidrIp)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyAccessRule", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudNasAccessRuleRead(d, meta)
}

func resourceAliCloudNasAccessRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	// Parse resource ID using service helper
	accessGroupName, accessRuleId, _, err := nasService.parseResourceId(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = nasService.DeleteNasAccessRule(accessGroupName, accessRuleId)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAccessRule", AlibabaCloudSdkGoERROR)
	}

	return nil
}
