package alicloud

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertContactGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertContactGroupCreate,
		Read:   resourceAliCloudArmsAlertContactGroupRead,
		Update: resourceAliCloudArmsAlertContactGroupUpdate,
		Delete: resourceAliCloudArmsAlertContactGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"contact_group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contact_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAliCloudArmsAlertContactGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}
	armsAPI, err := aliyunArmsAPI.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact_group", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	contactGroupName := d.Get("contact_group_name").(string)
	var contactIds []string
	if v, ok := d.GetOk("contact_ids"); ok {
		for _, id := range v.(*schema.Set).List() {
			contactIds = append(contactIds, id.(string))
		}
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	var contactGroup *aliyunArmsAPI.AlertContactGroup
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		contactGroup, err = armsAPI.CreateAlertContactGroup(contactGroupName, contactIds)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact_group", "CreateAlertContactGroup", AlibabaCloudSdkGoERROR)
	}

	d.SetId(contactGroup.ContactGroupId)

	return resourceAliCloudArmsAlertContactGroupRead(d, meta)
}

func resourceAliCloudArmsAlertContactGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}
	armsAPI, err := aliyunArmsAPI.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact_group", "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	// Search for the contact group using ARMS API
	contactGroups, err := armsAPI.SearchAlertContactGroup([]string{d.Id()}, true)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SearchAlertContactGroup", "Failed to search alert contact group")
	}

	if len(contactGroups) == 0 {
		log.Printf("[DEBUG] Resource alicloud_arms_alert_contact_group not found, removing from state")
		d.SetId("")
		return nil
	}

	contactGroup := contactGroups[0]
	d.Set("contact_group_name", contactGroup.ContactGroupName)

	// Parse ContactIds string into slice
	contactIdsItems := make([]string, 0)

	d.Set("contact_ids", contactIdsItems)
	return nil
}

func resourceAliCloudArmsAlertContactGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	update := false

	if d.HasChange("contact_group_name") || d.HasChange("contact_ids") {
		update = true
	}

	if update {
		// Create ARMS API client
		armsCredentials := &common.Credentials{
			AccessKey:     client.AccessKey,
			SecretKey:     client.SecretKey,
			RegionId:      client.RegionId,
			SecurityToken: client.SecurityToken,
		}
		armsAPI, err := aliyunArmsAPI.NewArmsAPI(armsCredentials)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewArmsAPI", AlibabaCloudSdkGoERROR)
		}

		contactGroupName := d.Get("contact_group_name").(string)
		var contactIds []string
		if v, ok := d.GetOk("contact_ids"); ok {
			for _, id := range v.(*schema.Set).List() {
				contactIds = append(contactIds, id.(string))
			}
		}

		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			// Convert string ID to int64
			contactGroupIdInt, parseErr := strconv.ParseInt(d.Id(), 10, 64)
			if parseErr != nil {
				return resource.NonRetryableError(WrapErrorf(parseErr, DefaultErrorMsg, d.Id(), "ParseInt", AlibabaCloudSdkGoERROR))
			}
			err := armsAPI.UpdateAlertContactGroup(contactGroupIdInt, contactGroupName, contactIds)
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
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateAlertContactGroup", AlibabaCloudSdkGoERROR)
		}
	}
	return resourceAliCloudArmsAlertContactGroupRead(d, meta)
}

func resourceAliCloudArmsAlertContactGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create ARMS API client
	armsCredentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}
	armsAPI, err := aliyunArmsAPI.NewArmsAPI(armsCredentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewArmsAPI", AlibabaCloudSdkGoERROR)
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		// Convert string ID to int64
		contactGroupIdInt, parseErr := strconv.ParseInt(d.Id(), 10, 64)
		if parseErr != nil {
			return resource.NonRetryableError(WrapErrorf(parseErr, DefaultErrorMsg, d.Id(), "ParseInt", AlibabaCloudSdkGoERROR))
		}
		err := armsAPI.DeleteAlertContactGroup(contactGroupIdInt)
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAlertContactGroup", AlibabaCloudSdkGoERROR)
	}
	return nil
}
