package alicloud

import (
	"log"
	"strconv"
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
		// Create AlertContactGroup instance for API call
		request := &aliyunArmsAPI.AlertContactGroup{
			ContactGroupName: contactGroupName,
		}

		// Convert string slice to int64 slice
		if len(contactIds) > 0 {
			request.ContactIds = make([]int64, len(contactIds))
			for i, idStr := range contactIds {
				if id, parseErr := strconv.ParseInt(idStr, 10, 64); parseErr == nil {
					request.ContactIds[i] = id
				}
			}
		}

		contactGroup, err = armsAPI.CreateOrUpdateAlertContactGroup(request)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact_group", "CreateOrUpdateContactGroup", AlibabaCloudSdkGoERROR)
	}

	d.SetId(strconv.FormatInt(contactGroup.ContactGroupId, 10))

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
	contactGroupIdInt, parseErr := strconv.ParseInt(d.Id(), 10, 64)
	if parseErr != nil {
		return WrapErrorf(parseErr, DefaultErrorMsg, d.Id(), "ParseInt", AlibabaCloudSdkGoERROR)
	}

	contactGroups, err := armsAPI.ListAlertContactGroups(1, 100)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeContactGroups", "Failed to describe alert contact groups")
	}

	var contactGroup *aliyunArmsAPI.AlertContactGroup
	for _, group := range contactGroups {
		if group.ContactGroupId == contactGroupIdInt {
			contactGroup = group
			break
		}
	}

	if contactGroup == nil {
		log.Printf("[DEBUG] Resource alicloud_arms_alert_contact_group not found, removing from state")
		d.SetId("")
		return nil
	}
	d.Set("contact_group_name", contactGroup.ContactGroupName)

	// Convert ContactIds int64 slice to string slice for schema
	contactIdsItems := make([]string, len(contactGroup.ContactIds))
	for i, id := range contactGroup.ContactIds {
		contactIdsItems[i] = strconv.FormatInt(id, 10)
	}

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

			// Create AlertContactGroup instance for API call
			request := &aliyunArmsAPI.AlertContactGroup{
				ContactGroupId:   contactGroupIdInt,
				ContactGroupName: contactGroupName,
			}

			// Convert string slice to int64 slice
			if len(contactIds) > 0 {
				request.ContactIds = make([]int64, len(contactIds))
				for i, idStr := range contactIds {
					if id, parseErr := strconv.ParseInt(idStr, 10, 64); parseErr == nil {
						request.ContactIds[i] = id
					}
				}
			}

			_, err := armsAPI.CreateOrUpdateAlertContactGroup(request)
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
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "CreateOrUpdateContactGroup", AlibabaCloudSdkGoERROR)
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
