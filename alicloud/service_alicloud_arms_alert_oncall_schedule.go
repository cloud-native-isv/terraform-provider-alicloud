package alicloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// EncodeArmsOnCallScheduleId encodes schedule ID into string format
func EncodeArmsOnCallScheduleId(scheduleId int64) string {
	return fmt.Sprintf("%d", scheduleId)
}

// DecodeArmsOnCallScheduleId decodes schedule ID string into int64
func DecodeArmsOnCallScheduleId(id string) (int64, error) {
	scheduleId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid schedule ID format, expected int64, got %s: %v", id, err)
	}
	return scheduleId, nil
}

// DescribeArmsOnCallSchedule describes a single on-call schedule by ID
func (s *ArmsService) DescribeArmsOnCallSchedule(scheduleId string, startTime, endTime string) (*arms.AlertOncallScheduleDetail, error) {
	id, err := DecodeArmsOnCallScheduleId(scheduleId)
	if err != nil {
		return nil, WrapError(err)
	}

	schedule, err := s.armsAPI.GetOnCallScheduleDetail(id, startTime, endTime)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, scheduleId, "GetOnCallScheduleDetail", AlibabaCloudSdkGoERROR)
	}

	return schedule, nil
}

// DescribeArmsOnCallSchedules describes multiple on-call schedules with pagination and filtering
func (s *ArmsService) DescribeArmsOnCallSchedules(page, size int64, name string) ([]*arms.AlertOncallSchedule, int64, error) {
	schedules, totalCount, err := s.armsAPI.ListOnCallSchedules(page, size, name)
	if err != nil {
		return nil, 0, WrapErrorf(err, DefaultErrorMsg, "arms_oncall_schedules", "ListOnCallSchedules", AlibabaCloudSdkGoERROR)
	}

	return schedules, totalCount, nil
}

// ArmsOnCallScheduleStateRefreshFunc returns a StateRefreshFunc that is used to watch on-call schedule status
func (s *ArmsService) ArmsOnCallScheduleStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsOnCallSchedule(id, "", "")
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// For on-call schedules, we can check if the schedule has valid configuration
		var currentStatus string = "Available"
		if object.Name == "" {
			currentStatus = "Unavailable"
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}

		return object, currentStatus, nil
	}
}

// WaitForArmsOnCallScheduleAvailable waits for on-call schedule to become available
func (s *ArmsService) WaitForArmsOnCallScheduleAvailable(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Unavailable"}, // pending states
		[]string{"Available"},   // target states
		timeout,
		5*time.Second,
		s.ArmsOnCallScheduleStateRefreshFunc(id, []string{"Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, id)
}

// FilterOnCallSchedulesByName filters on-call schedules by name pattern
func (s *ArmsService) FilterOnCallSchedulesByName(schedules []*arms.AlertOncallSchedule, namePattern string) []*arms.AlertOncallSchedule {
	if namePattern == "" {
		return schedules
	}

	var filteredSchedules []*arms.AlertOncallSchedule
	for _, schedule := range schedules {
		if strings.Contains(schedule.Name, namePattern) {
			filteredSchedules = append(filteredSchedules, schedule)
		}
	}

	return filteredSchedules
}
