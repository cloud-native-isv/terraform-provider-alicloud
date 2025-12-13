package alicloud

import (
	"fmt"
	"strings"
	"time"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// EncodeSlsConsumerGroupId encodes project, logstore and consumerGroup into a single ID string: project:logstore:consumerGroup
func EncodeSlsConsumerGroupId(project, logstore, consumerGroup string) string {
	return fmt.Sprintf("%s:%s:%s", project, logstore, consumerGroup)
}

// DecodeSlsConsumerGroupId decodes an ID string into project, logstore and consumerGroup components
func DecodeSlsConsumerGroupId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid consumer group ID format, expected project:logstore:consumer_group, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// DescribeSlsConsumerGroup retrieves a consumer group by composite ID
func (s *SlsService) DescribeSlsConsumerGroup(id string) (*aliyunSlsAPI.LogConsumerGroup, error) {
	project, logstore, group, err := DecodeSlsConsumerGroupId(id)
	if err != nil {
		return nil, WrapError(err)
	}
	cg, e := s.GetAPI().GetConsumerGroup(project, logstore, group)
	if e != nil {
		// Map plain not found error to provider's not found error for consistent handling
		if strings.Contains(strings.ToLower(e.Error()), "not found") {
			return nil, WrapErrorf(NotFoundErr("ConsumerGroup", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(e, DefaultErrorMsg, id, "GetConsumerGroup", AlibabaCloudSdkGoERROR)
	}
	return cg, nil
}

// CreateOrAdoptSlsConsumerGroup creates consumer group or adopts existing and converges behavioral params
func (s *SlsService) CreateOrAdoptSlsConsumerGroup(project, logstore string, cg *aliyunSlsAPI.LogConsumerGroup) error {
	if cg == nil {
		return fmt.Errorf("consumer group cannot be nil")
	}
	if cg.ConsumerGroup == "" {
		return fmt.Errorf("consumer group name cannot be empty")
	}

	// Ensure location fields are set
	cg.ProjectName = project
	cg.LogstoreName = logstore

	// Try to get existing
	existing, err := s.GetAPI().GetConsumerGroup(project, logstore, cg.ConsumerGroup)
	if err != nil {
		// Treat any 'not found' error as create path
		if NotFoundError(err) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return s.GetAPI().CreateConsumerGroup(cg)
		}
		return WrapErrorf(err, DefaultErrorMsg, cg.ConsumerGroup, "GetConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	// Adopt and converge behavior fields
	if existing != nil {
		// Only update behavioral fields: Timeout, Order
		patch := &aliyunSlsAPI.LogConsumerGroup{
			ProjectName:   project,
			LogstoreName:  logstore,
			ConsumerGroup: existing.ConsumerGroup,
			Timeout:       cg.Timeout,
			Order:         cg.Order,
		}
		return s.GetAPI().UpdateConsumerGroup(patch)
	}
	return nil
}

// UpdateSlsConsumerGroup updates behavioral parameters of consumer group
func (s *SlsService) UpdateSlsConsumerGroup(project, logstore, consumerGroup string, cg *aliyunSlsAPI.LogConsumerGroup) error {
	if cg == nil {
		return fmt.Errorf("consumer group cannot be nil")
	}
	cg.ProjectName = project
	cg.LogstoreName = logstore
	cg.ConsumerGroup = consumerGroup
	return s.GetAPI().UpdateConsumerGroup(cg)
}

// DeleteSlsConsumerGroup deletes a consumer group
func (s *SlsService) DeleteSlsConsumerGroup(project, logstore, consumerGroup string) error {
	return s.GetAPI().DeleteConsumerGroup(project, logstore, consumerGroup)
}

// SlsConsumerGroupStateRefreshFunc returns a StateRefreshFunc to poll consumer group existence
func (s *SlsService) SlsConsumerGroupStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeSlsConsumerGroup(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "deleted", nil
			}
			return nil, "", WrapError(err)
		}
		// consumer group existence implies available; no complex states
		return obj, "available", nil
	}
}

// WaitForSlsConsumerGroupCreating waits until the consumer group exists
func (s *SlsService) WaitForSlsConsumerGroupCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{}, []string{"available"}, timeout, 5*time.Second, s.SlsConsumerGroupStateRefreshFunc(id, []string{}))
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// WaitForSlsConsumerGroupDeleting waits until the consumer group is deleted
func (s *SlsService) WaitForSlsConsumerGroupDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"available"}, []string{"deleted"}, timeout, 5*time.Second, s.SlsConsumerGroupStateRefreshFunc(id, []string{}))
	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}
