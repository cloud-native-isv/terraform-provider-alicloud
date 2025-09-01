package alicloud

import (
	"fmt"
	"strconv"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
)

// =============================================================================
// Alert Contact Functions
// =============================================================================

// DescribeArmsAlertContact describes ARMS alert contact by contact ID
func (s *ArmsService) DescribeArmsAlertContact(contactId string) (*aliyunArmsAPI.AlertContact, error) {
	id, err := strconv.ParseInt(contactId, 10, 64)
	if err != nil {
		return nil, WrapErrorf(err, "Invalid contact ID: %s", contactId)
	}

	// Use API to get all contacts and find the specific one
	contacts, err := s.armsAPI.ListAlertContacts(1, 100)
	if err != nil {
		return nil, WrapError(err)
	}

	for _, contact := range contacts {
		if contact.ContactId == id {
			return contact, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("ARMS Alert Contact", contactId), NotFoundMsg, AlibabaCloudSdkGoERROR)
}

// DescribeArmsAlertContacts describes ARMS alert contacts with filters
func (s *ArmsService) DescribeArmsAlertContacts() ([]*aliyunArmsAPI.AlertContact, error) {
	return s.armsAPI.ListAllAlertContacts()
}

// =============================================================================
// ID Encoding/Decoding Functions
// =============================================================================

// EncodeArmsAlertContactId encodes contact ID for Terraform resource identification
func EncodeArmsAlertContactId(contactId int64) string {
	return fmt.Sprintf("%d", contactId)
}

// DecodeArmsAlertContactId decodes contact ID from Terraform resource identification
func DecodeArmsAlertContactId(id string) (int64, error) {
	contactId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, WrapErrorf(err, "Invalid contact ID format: %s", id)
	}
	return contactId, nil
}

// EncodeArmsAlertContactGroupId encodes contact group ID for Terraform resource identification
func EncodeArmsAlertContactGroupId(contactGroupId int64) string {
	return fmt.Sprintf("%d", contactGroupId)
}

// DecodeArmsAlertContactGroupId decodes contact group ID from Terraform resource identification
func DecodeArmsAlertContactGroupId(id string) (int64, error) {
	contactGroupId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, WrapErrorf(err, "Invalid contact group ID format: %s", id)
	}
	return contactGroupId, nil
}
