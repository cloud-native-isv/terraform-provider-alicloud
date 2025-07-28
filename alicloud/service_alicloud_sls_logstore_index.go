package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateSlsLogStoreIndex creates a new log store index
func (s *SlsService) CreateSlsLogStoreIndex(projectName string, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.CreateLogStoreIndex(projectName, logstoreName, index)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "CreateLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsLogStoreIndex updates an existing log store index
func (s *SlsService) UpdateSlsLogStoreIndex(projectName string, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateLogStoreIndex(projectName, logstoreName, index)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "UpdateLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsLogStoreIndex deletes a log store index
func (s *SlsService) DeleteSlsLogStoreIndex(projectName string, logstoreName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteLogStoreIndex(projectName, logstoreName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "DeleteLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// LogStoreIndexStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a log store index
func (s *SlsService) LogStoreIndexStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parts := strings.Split(id, ":")
		project := parts[0]
		logstore := parts[1]
		object, err := s.GetSlsLogStoreIndex(project, logstore)
		if err != nil {
			if IsNotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// LogStoreIndexAvailabilityStateRefreshFunc returns a StateRefreshFunc for checking log store index availability
func (s *SlsService) LogStoreIndexAvailabilityStateRefreshFunc(id string, expectExists bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parts := strings.Split(id, ":")
		if len(parts) != 2 {
			return nil, "", WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		}

		projectName := parts[0]
		logstoreName := parts[1]

		// First check if log store exists
		_, err := s.DescribeLogStore(projectName, logstoreName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "LogStoreNotFound", nil
			}
			return nil, "", WrapError(err)
		}

		// Then check log store index
		object, err := s.GetSlsLogStoreIndex(projectName, logstoreName)
		if err != nil {
			if IsNotFoundError(err) || IsExpectedErrors(err, []string{"IndexConfigNotExist"}) {
				if expectExists {
					return object, "IndexNotFound", nil
				} else {
					// Index doesn't exist and we don't expect it to exist (deletion scenario)
					return object, "IndexDeleted", nil
				}
			}
			return nil, "", WrapError(err)
		}

		// Index exists
		if expectExists {
			return object, "IndexAvailable", nil
		} else {
			// Index still exists but we expect it to be deleted
			return object, "IndexExists", nil
		}
	}
}

// WaitForLogStoreIndexAvailable waits for log store index to become available
func (s *SlsService) WaitForLogStoreIndexAvailable(id string, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"IndexNotFound"},
		Target:     []string{"IndexAvailable"},
		Refresh:    s.LogStoreIndexAvailabilityStateRefreshFunc(id, true),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

// WaitForLogStoreIndexDeleted waits for log store index to be deleted
func (s *SlsService) WaitForLogStoreIndexDeleted(id string, timeout int) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"IndexExists"},
		Target:     []string{"IndexDeleted"},
		Refresh:    s.LogStoreIndexAvailabilityStateRefreshFunc(id, false),
		Timeout:    time.Duration(timeout) * time.Second,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

// GetSlsLogStoreIndex encapsulates the call to aliyunSlsAPI.GetLogStoreIndex
func (s *SlsService) GetSlsLogStoreIndex(projectName, logstoreName string) (*aliyunSlsAPI.LogStoreIndex, error) {
	return s.aliyunSlsAPI.GetLogStoreIndex(projectName, logstoreName)
}
