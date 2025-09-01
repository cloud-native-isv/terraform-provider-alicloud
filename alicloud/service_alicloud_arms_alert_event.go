package alicloud

import (
	"fmt"
	"strconv"
	"strings"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
)

// =============================================================================
// Alert Item Functions (告警项 - 告警事件类型定义)
// =============================================================================

// DescribeArmsAlertItems describes ARMS alert items with pagination
func (s *ArmsService) DescribeArmsAlertItems(page, size int64) ([]*aliyunArmsAPI.AlertItem, error) {
	return s.armsAPI.ListAlertItems(page, size)
}

// DescribeArmsAllAlertItems describes all ARMS alert items across multiple pages
func (s *ArmsService) DescribeArmsAllAlertItems() ([]*aliyunArmsAPI.AlertItem, error) {
	return s.armsAPI.ListAllAlertItems()
}

// DescribeArmsAlertItemById describes a single ARMS alert item by ID
func (s *ArmsService) DescribeArmsAlertItemById(alertId string) (*aliyunArmsAPI.AlertItem, error) {
	id, err := strconv.ParseInt(alertId, 10, 64)
	if err != nil {
		return nil, WrapErrorf(err, "Invalid alert ID: %s", alertId)
	}

	// Get all alert items and find the specific one
	alertItems, err := s.DescribeArmsAllAlertItems()
	if err != nil {
		return nil, WrapError(err)
	}

	for _, item := range alertItems {
		if item.AlertId == id {
			return item, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("ARMS Alert Item", alertId), NotFoundMsg, AlibabaCloudSdkGoERROR)
}

// =============================================================================
// Alert Event Functions (告警事件 - 具体告警实例)
// =============================================================================

// DescribeArmsAlertEvents describes ARMS alert events with pagination and filters
func (s *ArmsService) DescribeArmsAlertEvents(page, size int64, filters map[string]interface{}) ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.ListAlertEvents(page, size, filters)
}

// DescribeArmsAllAlertEvents describes all ARMS alert events with filters
func (s *ArmsService) DescribeArmsAllAlertEvents(filters map[string]interface{}) ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.ListAllAlertEvents(filters)
}

// DescribeArmsAlertEventById describes a single ARMS alert event by event ID
func (s *ArmsService) DescribeArmsAlertEventById(eventId string) (*aliyunArmsAPI.AlertEvent, error) {
	if eventId == "" {
		return nil, WrapErrorf(fmt.Errorf("eventId cannot be empty"), "Invalid event ID: %s", eventId)
	}

	// Get all alert events and find the specific one
	alertEvents, err := s.DescribeArmsAllAlertEvents(nil)
	if err != nil {
		return nil, WrapError(err)
	}

	for _, event := range alertEvents {
		if event.EventId == eventId {
			return event, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("ARMS Alert Event", eventId), NotFoundMsg, AlibabaCloudSdkGoERROR)
}

// DescribeArmsActiveAlertEvents describes all active (unresolved) ARMS alert events
func (s *ArmsService) DescribeArmsActiveAlertEvents() ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.GetActiveAlertEvents()
}

// DescribeArmsAlertEventsByAlertName describes ARMS alert events by alert name
func (s *ArmsService) DescribeArmsAlertEventsByAlertName(alertName string) ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.GetAlertEventsByAlertName(alertName)
}

// DescribeArmsAlertEventsByTimeRange describes ARMS alert events by time range
func (s *ArmsService) DescribeArmsAlertEventsByTimeRange(startTime, endTime int64) ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.GetAlertEventsByTimeRange(startTime, endTime)
}

// DescribeArmsAlertEventsBySeverity describes ARMS alert events by severity level
func (s *ArmsService) DescribeArmsAlertEventsBySeverity(severity string) ([]*aliyunArmsAPI.AlertEvent, error) {
	return s.armsAPI.GetAlertEventsBySeverity(severity)
}

// =============================================================================
// Alert Activity Functions (告警活动 - 审计记录)
// =============================================================================

// DescribeArmsAlertActivities describes ARMS alert activities from all alert events
func (s *ArmsService) DescribeArmsAlertActivities(filters map[string]interface{}) ([]*aliyunArmsAPI.AlertActivity, error) {
	// Get all alert events first
	alertEvents, err := s.DescribeArmsAllAlertEvents(filters)
	if err != nil {
		return nil, WrapError(err)
	}

	// Extract activities from all events
	var allActivities []*aliyunArmsAPI.AlertActivity
	for _, event := range alertEvents {
		if event.Activities != nil {
			allActivities = append(allActivities, event.Activities...)
		}
	}

	return allActivities, nil
}

// DescribeArmsAlertActivitiesByAlertId describes ARMS alert activities by alert ID
func (s *ArmsService) DescribeArmsAlertActivitiesByAlertId(alertId int64) ([]*aliyunArmsAPI.AlertActivity, error) {
	// Filter events by alert ID
	filters := map[string]interface{}{
		"alertId": alertId,
	}

	// Get alert events for this alert ID
	alertEvents, err := s.DescribeArmsAllAlertEvents(filters)
	if err != nil {
		return nil, WrapError(err)
	}

	// Extract activities from events
	var activities []*aliyunArmsAPI.AlertActivity
	for _, event := range alertEvents {
		if event.AlertId == alertId && event.Activities != nil {
			activities = append(activities, event.Activities...)
		}
	}

	return activities, nil
}

// =============================================================================
// Alert Alarm Functions (告警历史 - 发送历史记录)
// =============================================================================

// DescribeArmsAlertAlarms describes ARMS alert alarms from all alert events
func (s *ArmsService) DescribeArmsAlertAlarms(filters map[string]interface{}) ([]*aliyunArmsAPI.AlertAlarm, error) {
	// Get all alert events first
	alertEvents, err := s.DescribeArmsAllAlertEvents(filters)
	if err != nil {
		return nil, WrapError(err)
	}

	// Extract alarms from all events
	var allAlarms []*aliyunArmsAPI.AlertAlarm
	for _, event := range alertEvents {
		if event.Alarms != nil {
			allAlarms = append(allAlarms, event.Alarms...)
		}
	}

	return allAlarms, nil
}

// DescribeArmsAlertAlarmsByAlertId describes ARMS alert alarms by alert ID
func (s *ArmsService) DescribeArmsAlertAlarmsByAlertId(alertId int64) ([]*aliyunArmsAPI.AlertAlarm, error) {
	// Filter events by alert ID
	filters := map[string]interface{}{
		"alertId": alertId,
	}

	// Get alert events for this alert ID
	alertEvents, err := s.DescribeArmsAllAlertEvents(filters)
	if err != nil {
		return nil, WrapError(err)
	}

	// Extract alarms from events
	var alarms []*aliyunArmsAPI.AlertAlarm
	for _, event := range alertEvents {
		if event.AlertId == alertId && event.Alarms != nil {
			alarms = append(alarms, event.Alarms...)
		}
	}

	return alarms, nil
}

// =============================================================================
// ID Encoding/Decoding Functions
// =============================================================================

// EncodeArmsAlertItemId encodes alert item ID for Terraform resource identification
func EncodeArmsAlertItemId(alertId int64) string {
	return fmt.Sprintf("%d", alertId)
}

// DecodeArmsAlertItemId decodes alert item ID from Terraform resource identification
func DecodeArmsAlertItemId(id string) (int64, error) {
	alertId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, WrapErrorf(err, "Invalid alert item ID format: %s", id)
	}
	return alertId, nil
}

// EncodeArmsAlertEventId encodes alert event ID for Terraform resource identification
func EncodeArmsAlertEventId(eventId string) string {
	return eventId
}

// DecodeArmsAlertEventId decodes alert event ID from Terraform resource identification
func DecodeArmsAlertEventId(id string) (string, error) {
	if id == "" {
		return "", WrapErrorf(fmt.Errorf("eventId cannot be empty"), "Invalid alert event ID format: %s", id)
	}
	return id, nil
}

// EncodeArmsAlertActivityId encodes alert activity ID for Terraform resource identification
// Format: alertId_eventId_activityId
func EncodeArmsAlertActivityId(alertId int64, eventId string, activityId int64) string {
	return fmt.Sprintf("%d_%s_%d", alertId, eventId, activityId)
}

// DecodeArmsAlertActivityId decodes alert activity ID from Terraform resource identification
func DecodeArmsAlertActivityId(id string) (int64, string, int64, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 3 {
		return 0, "", 0, WrapErrorf(fmt.Errorf("invalid activity ID format"), "expected alertId_eventId_activityId, got %s", id)
	}

	alertId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", 0, WrapErrorf(err, "Invalid alert ID in activity ID: %s", parts[0])
	}

	eventId := parts[1]

	activityId, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, "", 0, WrapErrorf(err, "Invalid activity ID: %s", parts[2])
	}

	return alertId, eventId, activityId, nil
}

// EncodeArmsAlertAlarmId encodes alert alarm ID for Terraform resource identification
// Format: alertId_eventId_alarmId
func EncodeArmsAlertAlarmId(alertId int64, eventId string, alarmId int64) string {
	return fmt.Sprintf("%d_%s_%d", alertId, eventId, alarmId)
}

// DecodeArmsAlertAlarmId decodes alert alarm ID from Terraform resource identification
func DecodeArmsAlertAlarmId(id string) (int64, string, int64, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 3 {
		return 0, "", 0, WrapErrorf(fmt.Errorf("invalid alarm ID format"), "expected alertId_eventId_alarmId, got %s", id)
	}

	alertId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", 0, WrapErrorf(err, "Invalid alert ID in alarm ID: %s", parts[0])
	}

	eventId := parts[1]

	alarmId, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, "", 0, WrapErrorf(err, "Invalid alarm ID: %s", parts[2])
	}

	return alertId, eventId, alarmId, nil
}
