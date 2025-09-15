package alicloud

import (
	"fmt"
	"log"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsAlertContact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsAlertContactCreate,
		Read:   resourceAliCloudArmsAlertContactRead,
		Update: resourceAliCloudArmsAlertContactUpdate,
		Delete: resourceAliCloudArmsAlertContactDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"contacts": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"contact_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Contact ID",
						},
						"arms_contact_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ARMS contact ID",
						},
						"contact_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Contact name",
						},
						"phone": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Phone number",
						},
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Email address",
						},
						"is_phone_verify": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Phone verified",
						},
						"is_email_verify": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Email verified",
						},
						"corp_user_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Corp user ID",
						},
						"ding_robot_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "DingTalk robot URL",
						},
						"reissue_send_notice": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Reissue send notice: 0=None, 1=Call again, 2=SMS, 3=Global default",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Region ID",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Resource group ID",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time",
						},
						"update_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Update time",
						},
					},
				},
				Description: "List of alert contacts to create",
			},
		},
	}
}

func resourceAliCloudArmsAlertContactCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get the contacts list from schema
	contactsList := d.Get("contacts").([]interface{})
	if len(contactsList) == 0 {
		return WrapError(fmt.Errorf("at least one contact must be provided"))
	}

	var contactIds []string
	var createdContacts []*aliyunArmsAPI.AlertContact

	// Process each contact in the list
	for i, contactRaw := range contactsList {
		contactMap := contactRaw.(map[string]interface{})

		// Build the alert contact object from schema data
		contact := &aliyunArmsAPI.AlertContact{
			ContactName: contactMap["contact_name"].(string),
		}

		// Set optional fields
		if email, ok := contactMap["email"]; ok && email.(string) != "" {
			contact.Email = email.(string)
		}
		if phone, ok := contactMap["phone"]; ok && phone.(string) != "" {
			contact.Phone = phone.(string)
		}
		if corpUserId, ok := contactMap["corp_user_id"]; ok && corpUserId.(string) != "" {
			contact.CorpUserId = corpUserId.(string)
		}
		if dingRobotUrl, ok := contactMap["ding_robot_url"]; ok && dingRobotUrl.(string) != "" {
			contact.DingRobotUrl = dingRobotUrl.(string)
		}
		if reissueSendNotice, ok := contactMap["reissue_send_notice"]; ok {
			contact.ReissueSendNotice = int64(reissueSendNotice.(int))
		}
		if isEmailVerify, ok := contactMap["is_email_verify"]; ok {
			contact.IsEmailVerify = isEmailVerify.(bool)
		}
		if resourceGroupId, ok := contactMap["resource_group_id"]; ok && resourceGroupId.(string) != "" {
			contact.ResourceGroupId = resourceGroupId.(string)
		}

		// Validate that at least one contact method is provided
		if contact.Email == "" && contact.Phone == "" && contact.DingRobotUrl == "" {
			return WrapError(fmt.Errorf("contact[%d]: at least one of 'email', 'phone', or 'ding_robot_url' must be provided", i))
		}

		// Call the service layer to create the contact
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			result, err := armsService.CreateOrUpdateArmsAlertContact(contact)
			if err != nil {
				if NeedRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			// Store the created contact
			createdContacts = append(createdContacts, result)
			contactIds = append(contactIds, fmt.Sprintf("%d", result.ContactId))
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_contact", "CreateOrUpdateAlertContact", AlibabaCloudSdkGoERROR)
		}
	}

	// Set the resource ID as a combination of all contact IDs
	d.SetId(fmt.Sprintf("%s", contactIds[0])) // Use first contact ID as primary ID

	return resourceAliCloudArmsAlertContactRead(d, meta)
}

func resourceAliCloudArmsAlertContactRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all alert contacts to avoid multiple API calls
	allContacts, err := armsService.DescribeArmsAlertContacts()
	if err != nil {
		if IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_alert_contact armsService.DescribeArmsAlertContacts Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Get the current contacts configuration
	contactsList := d.Get("contacts").([]interface{})
	if len(contactsList) == 0 {
		d.SetId("")
		return nil
	}

	// Build a map for quick lookup of existing contacts by name
	contactMap := make(map[string]*aliyunArmsAPI.AlertContact)
	for _, contact := range allContacts {
		contactMap[contact.ContactName] = contact
	}

	// Update the contacts list with current state
	updatedContacts := make([]interface{}, 0)
	foundAny := false

	for _, contactRaw := range contactsList {
		contactCfg := contactRaw.(map[string]interface{})
		contactName := contactCfg["contact_name"].(string)

		if existingContact, exists := contactMap[contactName]; exists {
			foundAny = true
			contactData := map[string]interface{}{
				"contact_id":          int(existingContact.ContactId),
				"arms_contact_id":     int(existingContact.ArmsContactId),
				"contact_name":        existingContact.ContactName,
				"phone":               existingContact.Phone,
				"email":               existingContact.Email,
				"is_phone_verify":     existingContact.IsPhoneVerify,
				"is_email_verify":     existingContact.IsEmailVerify,
				"corp_user_id":        existingContact.CorpUserId,
				"ding_robot_url":      existingContact.DingRobotUrl,
				"reissue_send_notice": int(existingContact.ReissueSendNotice),
				"region":              existingContact.Region,
				"resource_group_id":   existingContact.ResourceGroupId,
			}

			// Set time fields if available
			if existingContact.CreateTime != nil {
				contactData["create_time"] = existingContact.CreateTime.Format("2006-01-02T15:04:05Z")
			}
			if existingContact.UpdateTime != nil {
				contactData["update_time"] = existingContact.UpdateTime.Format("2006-01-02T15:04:05Z")
			}

			updatedContacts = append(updatedContacts, contactData)
		}
	}

	if !foundAny {
		log.Printf("[DEBUG] No contacts found, removing from state")
		d.SetId("")
		return nil
	}

	// Update the contacts in state
	d.Set("contacts", updatedContacts)

	return nil
}

func resourceAliCloudArmsAlertContactUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Check if the contacts list has changed
	if d.HasChange("contacts") {
		old, new := d.GetChange("contacts")
		oldContacts := old.([]interface{})
		newContacts := new.([]interface{})

		// Get all existing contacts for ID lookup
		allContacts, err := armsService.DescribeArmsAlertContacts()
		if err != nil {
			return WrapError(err)
		}

		// Build a map for quick lookup of existing contacts by name
		contactMap := make(map[string]*aliyunArmsAPI.AlertContact)
		for _, contact := range allContacts {
			contactMap[contact.ContactName] = contact
		}

		// Process new contacts configuration
		for i, contactRaw := range newContacts {
			contactCfg := contactRaw.(map[string]interface{})
			contactName := contactCfg["contact_name"].(string)

			// Build the alert contact object
			contact := &aliyunArmsAPI.AlertContact{
				ContactName: contactName,
			}

			// If contact exists, set its ID for update
			if existingContact, exists := contactMap[contactName]; exists {
				contact.ContactId = existingContact.ContactId
			}

			// Set optional fields
			if email, ok := contactCfg["email"]; ok && email.(string) != "" {
				contact.Email = email.(string)
			}
			if phone, ok := contactCfg["phone"]; ok && phone.(string) != "" {
				contact.Phone = phone.(string)
			}
			if corpUserId, ok := contactCfg["corp_user_id"]; ok && corpUserId.(string) != "" {
				contact.CorpUserId = corpUserId.(string)
			}
			if dingRobotUrl, ok := contactCfg["ding_robot_url"]; ok && dingRobotUrl.(string) != "" {
				contact.DingRobotUrl = dingRobotUrl.(string)
			}
			if reissueSendNotice, ok := contactCfg["reissue_send_notice"]; ok {
				contact.ReissueSendNotice = int64(reissueSendNotice.(int))
			}
			if isEmailVerify, ok := contactCfg["is_email_verify"]; ok {
				contact.IsEmailVerify = isEmailVerify.(bool)
			}
			if resourceGroupId, ok := contactCfg["resource_group_id"]; ok && resourceGroupId.(string) != "" {
				contact.ResourceGroupId = resourceGroupId.(string)
			}

			// Validate that at least one contact method is provided
			if contact.Email == "" && contact.Phone == "" && contact.DingRobotUrl == "" {
				return WrapError(fmt.Errorf("contact[%d]: at least one of 'email', 'phone', or 'ding_robot_url' must be provided", i))
			}

			// Call the service layer to create or update the contact
			err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				_, err := armsService.CreateOrUpdateArmsAlertContact(contact)
				if err != nil {
					if NeedRetry(err) {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "CreateOrUpdateAlertContact", AlibabaCloudSdkGoERROR)
			}
		}

		// Handle deleted contacts - find contacts that were in old but not in new
		oldContactNames := make(map[string]bool)
		for _, contactRaw := range oldContacts {
			contactCfg := contactRaw.(map[string]interface{})
			oldContactNames[contactCfg["contact_name"].(string)] = true
		}

		newContactNames := make(map[string]bool)
		for _, contactRaw := range newContacts {
			contactCfg := contactRaw.(map[string]interface{})
			newContactNames[contactCfg["contact_name"].(string)] = true
		}

		// Delete contacts that are no longer in the configuration
		for contactName := range oldContactNames {
			if !newContactNames[contactName] {
				if existingContact, exists := contactMap[contactName]; exists {
					err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
						err := armsService.DeleteArmsAlertContact(existingContact.ContactId)
						if err != nil {
							if IsNotFoundError(err) {
								return nil // Already deleted
							}
							if NeedRetry(err) {
								return resource.RetryableError(err)
							}
							return resource.NonRetryableError(err)
						}
						return nil
					})

					if err != nil {
						return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAlertContact", AlibabaCloudSdkGoERROR)
					}
				}
			}
		}
	}

	return resourceAliCloudArmsAlertContactRead(d, meta)
}

func resourceAliCloudArmsAlertContactDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	armsService, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get the contacts list from schema
	contactsList := d.Get("contacts").([]interface{})

	if len(contactsList) > 0 {
		// Get all existing contacts for ID lookup
		allContacts, err := armsService.DescribeArmsAlertContacts()
		if err != nil {
			if IsNotFoundError(err) {
				// No contacts exist, already deleted
				return nil
			}
			return WrapError(err)
		}

		// Build a map for quick lookup of existing contacts by name
		contactMap := make(map[string]*aliyunArmsAPI.AlertContact)
		for _, contact := range allContacts {
			contactMap[contact.ContactName] = contact
		}

		// Delete each contact in the configuration
		for _, contactRaw := range contactsList {
			contactCfg := contactRaw.(map[string]interface{})
			contactName := contactCfg["contact_name"].(string)

			if existingContact, exists := contactMap[contactName]; exists {
				// Call the service layer to delete the contact
				err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
					err := armsService.DeleteArmsAlertContact(existingContact.ContactId)
					if err != nil {
						if IsNotFoundError(err) {
							// Contact already deleted
							return nil
						}
						if NeedRetry(err) {
							return resource.RetryableError(err)
						}
						return resource.NonRetryableError(err)
					}
					return nil
				})

				if err != nil {
					return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAlertContact", AlibabaCloudSdkGoERROR)
				}
			}
		}
	}

	return nil
}
